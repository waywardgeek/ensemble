"""Configured stdio entry point; keep skills/shared beside this package in the sandbox.

The host/operator supplies setup.json as a literal launch argument. Model tool
arguments cannot retarget storage. This is configuration, not proof of ownership
or an authority grant: setup must name the actual initiating Agent's directory.
The existing OS sandbox and reverse dispatcher remain the authority boundaries.
"""
import argparse
import json
from pathlib import Path
import sys

# The skills/shared directory is inside the sandbox or its approved read-only tree.
# This is a fixed package relationship, not a caller-provided module search path.
sys.path.insert(0, str(Path(__file__).resolve().parent.parent / "shared"))
from ensemble_mcp import Server, check_contracts
from book_driver import _load, advance_book, bind_plan
from run_state import RunStore, StateError, _keys


class Workflow:
    """One immutable launch configuration; RunStore owns each invocation's lock."""

    def __init__(self, setup_path):
        setup = _load(Path(setup_path))
        _keys(setup, ("version", "owner_directory", "run_directory", "chapter_count",
                      "outline", "voice", "baseline_root", "source_limit", "result_limit"))
        if type(setup["version"]) is not int or setup["version"] != 1:
            raise StateError("Unsupported workflow setup version")
        self.owner = Path(setup["owner_directory"]).resolve(strict=True)
        self.run = Path(setup["run_directory"]).resolve(strict=True)
        self.baseline = Path(setup["baseline_root"]).resolve(strict=True)
        if not all(path.is_dir() for path in (self.owner, self.run, self.baseline)):
            raise StateError("Workflow setup requires existing directories")
        self.plan_arguments = {key: setup[key] for key in (
            "chapter_count", "outline", "voice", "source_limit", "result_limit")}
        self.contracts = _load(Path(__file__).with_name("contracts.json"))

    async def continue_book(self, connection, arguments):
        """Preflight before mutation; continue progress, never retry uncertainty.

        Cancellation unwinds the lock but cannot retract admitted remote work.
        Reopening reconciles the ledger. A concurrent invocation fails the lock;
        it must not queue a second independent driver for the same book.
        """
        _keys(arguments, ())
        check_contracts(self.contracts, await connection.list_tools())
        try:
            children = self.owner / "children"
            children.mkdir(exist_ok=True)
            with RunStore(self.run) as store:
                plan = bind_plan(store, baseline_root=self.baseline,
                                 children_directory=children, **self.plan_arguments)
                while True:
                    outcome = await advance_book(connection, store, plan)
                    action = outcome["action"]
                    if action in ("complete", "observe", "supervise"):
                        return {"content": [{"type": "text", "text": json.dumps(outcome)}]}
                    if action not in ("advanced", "reviewed", "approved", "revision_limit"):
                        raise StateError("Unknown workflow progress outcome")
        except StateError as error:
            # StateError is reserved for fixed, safe domain diagnostics. Do not
            # expose arbitrary exception strings from filesystem/protocol code.
            reason = str(error)
            if len(reason) > 512:
                reason = "Workflow state refused; inspect the saved run before continuing"
        except UnicodeDecodeError:
            reason = "Required workflow text is not valid UTF-8"
        except json.JSONDecodeError:
            reason = "Workflow JSON is malformed; preserve the saved run for inspection"
        except OSError:
            reason = "Workflow file access failed; inspect required files and saved state before retrying"
        return {"isError": True, "content": [{"type": "text", "text": reason}]}


def main():
    parser = argparse.ArgumentParser(description="Run the configured fiction workflow MCP server")
    parser.add_argument("setup", help="Host-prepared setup JSON; no variable expansion")
    options = parser.parse_args()
    workflow = Workflow(options.setup)
    server = Server("fiction-workflow")
    server.tool("continue_fiction_book", "Continue the configured book or report supervision needed.",
                {"type": "object", "properties": {}, "additionalProperties": False})(
                    workflow.continue_book)
    server.run()


if __name__ == "__main__":
    main()
