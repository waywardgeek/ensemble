"""One read-only reverse callback; no retry or durable multi-step state."""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent / "shared"))
from ensemble_mcp import Server, check_contracts

expected = json.loads(Path(__file__).with_name("contracts.json").read_text())
server = Server("brief-example")

@server.tool("inspect_brief", "Read the configured brief once", {
    "type": "object",
    "properties": {},
    "additionalProperties": False,
})
async def inspect_brief(connection, arguments):
    definitions = await connection.list_tools()
    check_contracts(expected, definitions)
    result = await connection.call_tool("read_file", {"path": "brief.md"})
    return result

if __name__ == "__main__":
    server.run()
