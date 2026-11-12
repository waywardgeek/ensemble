"""External stdio exercise check. No imports from candidate implementation/tests."""
import argparse
import asyncio
import json
from pathlib import Path
import re
import sys
import tempfile


def result(value):
    text = value if isinstance(value, str) else json.dumps(value)
    return {"content": [{"type": "text", "text": text}], "isError": False}


class Peer:
    """Working child boundary: persistent submissions, files and joined lifetime."""
    def __init__(self, root):
        self.root = root
        self.live = set()
        self.saved = {}
        self.tasks = []
        self.lost = False
        # Public task destinations and actual bytes identify story artifacts;
        # candidate-private directory names are not part of this interface.
        self.chapter_outputs = {}
        self.last_writing_round = 0
        self.review_positions = {}
        self.last_global_source = None

    def call(self, name, args):
        child = args.get("agent_id")
        if name == "agents_status":
            return result({"agents": [{"agent_id": item, "activity": "idle",
                "paused": False, "failure": None} for item in self.live]})
        if name == "get_submission":
            return result({"agent_id": child, "historical": True,
                           "payload": self.saved[child]})
        if name == "shutdown_agent":
            assert child in self.live, "Duplicate or foreign shutdown"
            self.live.remove(child)
            return result({"agent_id": child, "stopped": True})
        assert name in ("spawn_sub_agent", "ask_agent"), name
        if name == "spawn_sub_agent":
            assert child not in self.live, "Duplicate live child"
            self.live.add(child)
        else:
            assert child in self.live, "Turn on absent child"
        task = args.get("task", args.get("content"))
        identity = re.search(r"^Workflow assignment_id: (.+)$", task, re.M).group(1)
        assert all(identity != old[0] for old in self.tasks), "Redispatched assignment"
        self.tasks.append((identity, task))
        payload = {"assignment_id": identity}
        outputs = re.findall(r"^(?:Write chapter|Write continuity carryover|Write feedback): (.+)$",
                             task, re.M)
        if "Load fiction-reviewer" not in task:
            position = re.search(r"Chapter (\d+) of (\d+); review round (\d+)\.", task)
            assert position, "Missing public chapter/round task label"
            self.last_writing_round = int(position.group(3))
            if "Load fiction-author." in task:
                destination = re.search(r"^Write chapter: (.+)$", task, re.M)
                assert destination, "Missing public chapter destination"
                self.chapter_outputs[int(position.group(1))] = Path(destination.group(1))
        if "Load fiction-editor." in task:
            payload["decision"] = "revise"
        if "Load fiction-reviewer" in task:
            match = re.search(r"passage references to (.+?)\. Finish the write", task)
            assert match, "Missing review destination"
            outputs = [match.group(1)]
            read = re.search(r"with these arguments: (\{[^\n]+\})\.\n", task)
            assert read, "Missing full-read instruction"
            arguments = json.loads(read.group(1))
            source = Path(arguments["path"]).read_bytes()
            assert arguments["max_bytes"] >= len(source)
            assert arguments["start_line"] == 1 and arguments["end_line"] == 0
            assert arguments["max_output_bytes"] > len(source)
            matching_chapters = [number for number, path in self.chapter_outputs.items()
                                 if path.read_bytes() == source]
            if len(matching_chapters) == 1:
                chapter = matching_chapters[0]
            else:
                assert not matching_chapters, "Ambiguous chapter review source"
                assert self.chapter_outputs and all(
                    path.read_bytes() in source for path in self.chapter_outputs.values()
                ), "Global review source omitted a current chapter"
                chapter = 0
                self.last_global_source = source
            self.review_positions[identity] = (self.last_writing_round + 1, chapter)
        assert outputs, "Task supplied no output destinations"
        for value in outputs:
            path = Path(value).resolve()
            assert self.root in path.parents, "Output escaped disposable run"
            path.write_text("Synthetic text for " + identity + "\n", encoding="utf-8")
        self.saved[child] = payload
        # A real remote operation may complete even though its caller loses the
        # completion report. Preserve that result across the server restart.
        if not self.lost:
            self.lost = True
            return result("Ambiguous completion; consult saved submission")
        if name == "spawn_sub_agent":
            return result(json.dumps({"agent_id": child, "task_admitted": True}) + "\n" +
                json.dumps({"agent_id": child, "task_completed": True,
                            "submitted": True, "payload": payload}))
        return result({"agent_id": child, "response": "Submitted"})


class CreativePeer(Peer):
    """Public task labels and result fields only; no candidate ledger access."""
    def __init__(self, root):
        super().__init__(root)
        self.turns = {}
        self.carryovers = {}
        self.repairs = 0
        self.reviewed_repairs = 0
        self.lost_repair = False

    def call(self, name, args):
        response = super().call(name, args)
        if name not in ("spawn_sub_agent", "ask_agent"):
            return response
        identity, task = self.tasks[-1]
        if "Load fiction-reviewer" in task:
            return response
        child = args["agent_id"]
        self.turns[child] = self.turns.get(child, 0) + 1
        attempt = self.turns[child]
        chapter = int(re.search(r"Chapter (\d+) of", task).group(1))
        payload = self.saved[child]
        if "Load fiction-author." in task:
            outline = re.search(r"^Working outline: (.+)$", task, re.M)
            assert outline, "Missing public working-outline destination"
            path = Path(outline.group(1)).resolve()
            assert self.root in path.parents and path != self.root / "outline.md"
            path.write_text("Working outline revision " + identity)
            self.carryovers[chapter] = Path(re.search(
                r"^Write continuity carryover: (.+)$", task, re.M).group(1))
            if chapter == 2 and attempt == 1:
                payload["prior_chapter_requests"] = [1]
            elif chapter == 2 and attempt == 2:
                permissions = re.search(
                    r"^Editor-authorized earlier chapter numbers for THIS turn: (.+)$", task, re.M)
                assert permissions and json.loads(permissions.group(1)) == [1], "Missing granted repair"
                self.chapter_outputs[1].write_text("Authorized earlier repair " + identity)
                self.carryovers[1].write_text("Updated earlier continuity " + identity)
                self.repairs += 1
                if not self.lost_repair:
                    self.lost_repair = True
                    return result("Lost reply after authorized earlier repair")
        elif chapter == 2:
            if attempt == 1:
                payload["prior_chapter_permissions"] = [1]
            elif attempt == 2:
                reads = [json.loads(value) for value in re.findall(
                    r"^Read the complete input \(\d+ source bytes\): (.+)$", task, re.M)]
                for path in (self.chapter_outputs[1], self.carryovers[1]):
                    matches = [r for r in reads if Path(r["path"]) == path]
                    assert matches and matches[0]["max_bytes"] >= path.stat().st_size, "Editor omitted repaired text"
                self.reviewed_repairs += 1
        if name == "spawn_sub_agent" and response["content"][0]["text"].startswith("{"):
            return result(json.dumps({"agent_id": child, "task_admitted": True}) + "\n" +
                          json.dumps({"agent_id": child, "task_completed": True,
                                      "submitted": True, "payload": payload}))
        return response


async def invoke(server, setup, peer, catalog):
    process = await asyncio.create_subprocess_exec(sys.executable, "-u", str(server), str(setup),
        stdin=asyncio.subprocess.PIPE, stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.PIPE, limit=2 * 1024 * 1024)
    errors = asyncio.create_task(process.stderr.read())

    async def send(value):
        process.stdin.write((json.dumps(value) + "\n").encode())
        await process.stdin.drain()

    async def receive():
        line = await asyncio.wait_for(process.stdout.readline(), 15)
        assert line, "Server closed before reply"
        return json.loads(line)

    try:
        await send({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {
            "protocolVersion": "2024-11-05", "capabilities": {},
            "clientInfo": {"name": "external-workflow-grader", "version": "1"}}})
        assert (await receive())["id"] == 1
        await send({"jsonrpc": "2.0", "method": "notifications/initialized"})
        await send({"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {
            "name": "continue_fiction_book", "arguments": {}}})
        while True:
            message = await receive()
            if message.get("id") == 2 and "method" not in message:
                return message["result"]
            method = message.get("method")
            if method == "tools/list":
                response = {"tools": catalog}
            else:
                assert method == "tools/call", method
                response = peer.call(message["params"]["name"], message["params"]["arguments"])
            await send({"jsonrpc": "2.0", "id": message["id"], "result": response})
    finally:
        process.stdin.close()
        forced = False
        try:
            await asyncio.wait_for(process.wait(), 5)
        except asyncio.TimeoutError:
            forced = True
            process.kill()
            await process.wait()
        diagnostics = (await errors).decode(errors="replace")
        assert not forced, "Server failed to join on EOF"
        assert process.returncode == 0, diagnostics


async def grade(server, creative=False):
    contracts = json.loads(Path(__file__).with_name("contracts.json").read_text())
    catalog = [{"name": name, "inputSchema": schema} for name, schema in contracts.items()]
    with tempfile.TemporaryDirectory(prefix="workflow-public-") as directory:
        root = Path(directory).resolve()
        for role in ("author", "editor"):
            home = root / "baseline" / role
            home.mkdir(parents=True)
            (home / "SOUL.md").write_text("Synthetic independent " + role)
            (home / "MEMORY.md").write_bytes(b"")
        (root / "outline.md").write_text("Two short chapters with continuity.")
        (root / "voice.md").write_text("Plain and specific.")
        setup = root / "setup.json"
        setup.write_text(json.dumps({"version": 1, "owner_directory": str(root),
            "run_directory": str(root), "baseline_root": str(root / "baseline"),
            "chapter_count": 2, "outline": "outline.md", "voice": "voice.md",
            "source_limit": 1048576, "result_limit": 2097152}))
        peer = CreativePeer(root) if creative else Peer(root)
        refused = await invoke(server, setup, peer, [])
        assert refused.get("isError"), "Missing contracts were accepted"
        assert not peer.tasks and not (root / "state.json").exists()
        print("PASS missing-contract refusal before work")
        interrupted = await invoke(server, setup, peer, catalog)
        # Surface a stale public contract as a preflight error, not a JSON parse failure.
        assert not interrupted.get("isError"), interrupted["content"]
        assert json.loads(interrupted["content"][0]["text"])["action"] == "supervise"
        assert len(peer.tasks) == 1
        completed = await invoke(server, setup, peer, catalog)
        if creative:
            assert json.loads(completed["content"][0]["text"])["action"] == "supervise"
            completed = await invoke(server, setup, peer, catalog)
        assert json.loads(completed["content"][0]["text"])["action"] == "complete"
        # Two chapters, four writing passes, three author/editor pairs per pass;
        # three rounds each add two detail reviews and one global review.
        assert len(peer.tasks) == 57, "Wrong revision or review count: %d" % len(peer.tasks)
        expected = []
        for round_number in range(4):
            if round_number:
                expected.extend([("review", round_number, number) for number in (1, 2, 0)])
            for number in (1, 2):
                expected.extend([(role, round_number, number)
                                 for _ in range(3) for role in ("author", "editor")])
        actual = []
        for identity, task in peer.tasks:
            if "Load fiction-reviewer" in task:
                round_number, chapter = peer.review_positions[identity]
                actual.append(("review", round_number, chapter))
            else:
                position = re.search(r"Chapter (\d+) of 2; review round (\d+)", task)
                role = "author" if "Load fiction-author." in task else "editor"
                actual.append((role, int(position.group(2)), int(position.group(1))))
        assert actual == expected, "Review/revision ordering differs from the required schedule"
        assert not peer.live, "Completion left live children"
        final = (root / "final-book/manuscript.md").read_bytes()
        for number in (1, 2):
            assert peer.chapter_outputs[number].read_bytes() in final
        assert peer.last_global_source is not None
        assert final != peer.last_global_source
        print("PASS uncertain-result reconciliation, bounded revisions, final assembly and cleanup")
        await invoke(server, setup, peer, catalog)
        assert len(peer.tasks) == 57, "Completed restart dispatched more work"
        print("PASS completed restart without redispatch")
        if creative:
            assert peer.repairs == peer.reviewed_repairs == 4
            assert (root / "outline.md").read_text() == "Two short chapters with continuity."
            assert b"Authorized earlier repair" in final
            print("PASS evolving outline, editor-authorized prior repairs and lost-reply recovery")


async def grade_all(server):
    await grade(server)
    await grade(server, creative=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--server", type=Path, required=True)
    args = parser.parse_args()
    asyncio.run(grade_all(args.server.resolve(strict=True)))
