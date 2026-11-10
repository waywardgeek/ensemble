"""Small stdio MCP server for Ensemble skills (Python standard library only).

Server owns the duplex connection, registered tools, pending replies and handlers.
Copy this module inside the skill's sandbox. It provides no authorization: reverse
calls are executed by Ensemble under the initiating Agent's current policy.
"""

import asyncio
import inspect
import json
import sys


class ProtocolError(Exception):
    """The connection or remote RPC failed; an operation may have taken effect."""


class ContractError(Exception):
    """A required tool's advertised input schema differs from the expectation."""


def schema_difference(expected, actual, path="$"):
    """Return the first structural difference, or None. Object order is ignored.

    Arrays retain order. JSON booleans are distinct from numbers. This compares
    contracts; it does not implement JSON Schema validation or prove behavior.
    """
    if type(expected) is not type(actual):
        return path + ": type differs"
    if isinstance(expected, dict):
        if expected.keys() != actual.keys():
            return path + ": object keys differ"
        for key in sorted(expected):
            difference = schema_difference(expected[key], actual[key], path + "[" + json.dumps(key) + "]")
            if difference is not None:
                return difference
    elif isinstance(expected, list):
        if len(expected) != len(actual):
            return path + ": array length differs"
        for index, value in enumerate(expected):
            difference = schema_difference(value, actual[index], path + "[" + str(index) + "]")
            if difference is not None:
                return difference
    elif expected != actual:
        return path + ": value differs"
    return None


def check_contracts(expected, definitions):
    """Check only named expected tools against an actual MCP tools/list result.

    Missing or duplicate names refuse. A successful check grants nothing and must
    not cause automatic retries if a later call loses its reply.
    """
    actual = {}
    for definition in definitions:
        name = definition["name"]
        if name in actual:
            raise ContractError("Duplicate tool: " + name)
        actual[name] = definition["inputSchema"]
    for name, schema in expected.items():
        if name not in actual:
            raise ContractError("Missing tool: " + name)
        difference = schema_difference(schema, actual[name])
        if difference is not None:
            raise ContractError(name + ": " + difference)


class Server:
    """One connection owner. Register async handlers before calling run().

    Handlers take (server, arguments) and return an MCP result object. Ordinary
    tools and workflows use the same decorator; workflows additionally await
    server.list_tools() or server.call_tool(). Never print to protocol stdout.
    """

    def __init__(self, name):
        self.name = name
        self.tools = {}
        self.pending = {}
        self.handlers = set()
        self.next_id = 0
        self.ready = False
        self.initialized = False
        self.closed = False
        self.writer = None
        self.sending = asyncio.Lock()

    def tool(self, name, description, input_schema):
        """Register one async handler; registration confers no Ensemble grant."""
        def register(handler):
            if self.writer is not None or name in self.tools:
                raise ValueError("Tool registration is closed or name is duplicate")
            if not inspect.iscoroutinefunction(handler):
                raise TypeError("Tool handlers must be async functions")
            # Freeze caller-owned schema data, not a second mutable authority.
            definition = json.loads(json.dumps({
                "name": name, "description": description, "inputSchema": input_schema
            }, allow_nan=False))
            self.tools[name] = (definition, handler)
            return handler
        return register

    async def send(self, message):
        data = (json.dumps(message, allow_nan=False, separators=(",", ":")) + "\n").encode("utf-8")
        async with self.sending:
            if self.closed:
                raise ProtocolError("MCP connection closed; outcome may be unknown")
            self.writer.write(data)
            await self.writer.drain()

    async def request(self, method, parameters):
        """Issue a correlated reverse request once. Cancellation is not rollback."""
        if not self.ready or self.closed:
            raise ProtocolError("MCP connection is not ready")
        self.next_id += 1
        identity = "python:" + str(self.next_id)
        reply = asyncio.get_running_loop().create_future()
        self.pending[identity] = reply
        try:
            await self.send({"jsonrpc": "2.0", "id": identity,
                             "method": method, "params": parameters})
            return await reply
        finally:
            self.pending.pop(identity, None)

    async def list_tools(self):
        result = await self.request("tools/list", {})
        return result["tools"]

    async def call_tool(self, name, arguments):
        """Return the full MCP result, including isError and all content parts."""
        return await self.request("tools/call", {"name": name, "arguments": arguments})

    async def _handle(self, message):
        identity = message["id"]
        method = message["method"]
        try:
            if not self.ready:
                raise ProtocolError("MCP initialization is incomplete")
            if method == "tools/list":
                result = {"tools": [entry[0] for entry in self.tools.values()]}
            elif method == "tools/call":
                parameters = message["params"]
                handler = self.tools[parameters["name"]][1]
                result = await handler(self, parameters.get("arguments", {}))
            else:
                await self.send({"jsonrpc": "2.0", "id": identity,
                                 "error": {"code": -32601, "message": "Method not found"}})
                return
            response = {"jsonrpc": "2.0", "id": identity, "result": result}
        except ContractError as error:
            # Contract diagnostics contain the expected tool/schema location,
            # not arbitrary handler exception text or actual schema values.
            response = {"jsonrpc": "2.0", "id": identity, "result": {
                "isError": True, "content": [{"type": "text", "text": str(error)}]}}
        except Exception:
            # Arbitrary exception messages may contain private paths or contents.
            response = {"jsonrpc": "2.0", "id": identity,
                        "error": {"code": -32603, "message": "Skill request failed"}}
        await self.send(response)

    def _finished(self, task):
        # Remove both ownership directions when admitted work completes. Retrieve
        # failures immediately; a broken output stream terminates the connection.
        self.handlers.discard(task)
        if not task.cancelled() and task.exception() is not None:
            self.writer.close()

    async def serve(self, reader, writer):
        """Run one reader; own and join every admitted handler on disconnect.

        Streams are supplied by run() or a real local test peer. The caller owns
        stream closure; this server owns work admitted through them.
        """
        if self.writer is not None:
            raise RuntimeError("Server can serve only one connection")
        self.writer = writer
        try:
            while line := await reader.readline():
                message = json.loads(line)
                if not isinstance(message, dict) or message.get("jsonrpc") != "2.0":
                    raise ProtocolError("Invalid MCP envelope")
                if "method" not in message:
                    reply = self.pending.get(message.get("id"))
                    if reply is not None and not reply.done():
                        if "error" in message:
                            reply.set_exception(ProtocolError("Reverse RPC refused"))
                        elif "result" in message:
                            reply.set_result(message["result"])
                        else:
                            raise ProtocolError("Reply has no result or error")
                    continue
                method = message["method"]
                if "id" not in message:
                    if method == "notifications/initialized":
                        if not self.initialized:
                            raise ProtocolError("Initialization notification out of order")
                        self.ready = True
                    continue
                if method == "initialize":
                    if self.initialized or message.get("params", {}).get("protocolVersion") != "2024-11-05":
                        raise ProtocolError("Invalid or repeated initialization")
                    self.initialized = True
                    await self.send({"jsonrpc": "2.0", "id": message["id"], "result": {
                        "protocolVersion": "2024-11-05", "capabilities": {"tools": {}},
                        "serverInfo": {"name": self.name, "version": "1"}}})
                    continue
                task = asyncio.create_task(self._handle(message))
                self.handlers.add(task)
                task.add_done_callback(self._finished)
        finally:
            self.closed = True
            for reply in self.pending.values():
                if not reply.done():
                    reply.set_exception(ProtocolError("MCP disconnected; outcome may be unknown"))
            for task in self.handlers:
                if not task.done():
                    task.cancel()
            await asyncio.gather(*self.handlers, return_exceptions=True)
            self.handlers.clear()
            self.pending.clear()

    def run(self):
        """Serve Unix stdio with a 16 MiB frame limit; never truncate a frame.

        OS confinement belongs to Ensemble's process owner.
        """
        async def main():
            loop = asyncio.get_running_loop()
            reader = asyncio.StreamReader(limit=16 * 1024 * 1024)
            protocol = asyncio.StreamReaderProtocol(reader)
            incoming, _ = await loop.connect_read_pipe(lambda: protocol, sys.stdin.buffer)
            outgoing, output_protocol = await loop.connect_write_pipe(
                asyncio.streams.FlowControlMixin, sys.stdout.buffer)
            writer = asyncio.StreamWriter(outgoing, output_protocol, None, loop)
            try:
                await self.serve(reader, writer)
            finally:
                incoming.close()
                writer.close()
        asyncio.run(main())
