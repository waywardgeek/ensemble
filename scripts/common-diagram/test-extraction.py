#!/usr/bin/env python3
"""Small behavior checks for the source-only diagram pipeline."""
import copy
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import xml.etree.ElementTree as ET

root = Path(__file__).resolve().parent
repository = Path(sys.argv[2]).resolve()
output = repository / "docs/common-diagrams"
spec = importlib.util.spec_from_file_location("diagram_generator", root / "generate.py")
generator = importlib.util.module_from_spec(spec)
spec.loader.exec_module(generator)
inventory = json.loads((output / "inventory.json").read_text())
reviewed = json.loads((root / "relationships.json").read_text())
types, edges = generator.build_graph(inventory, reviewed)
assert len(types) == 156
assert len(edges) == 216
assert len(reviewed["relationships"]) == 36
assert not any(e["owner"] == "Tool" and e["field"] == "handler" for e in edges)
assert not any(e["owner"] == "Agent" and e["field"] == "memoryTask" for e in edges)
assert not any(e["owner"] == "Agent" and e["field"] == "selectedModel" for e in edges)
assert any(e["owner"] == "Model" and e["child"] == "Agent" and e["kind"] == "weak" for e in edges)
assert any(e["owner"] == "Agent" and e["child"] == "Agent" and e["kind"] == "strong" for e in edges)

# Every diagram must contain precisely its emitted node declarations. In
# particular the complete graph must not silently drop disconnected types.
namespace = {"s": "http://www.w3.org/2000/svg"}
full = ET.parse(output / "all.svg")
node_names = {n.find("s:title", namespace).text for n in full.findall('.//s:g[@class="node"]', namespace)}
assert node_names == set(types)
for svg in output.glob("*.svg"):
    ET.parse(svg)

# A typo or removed child-side back-reference must not silently yield a graph.
bad = copy.deepcopy(reviewed)
bad["relationships"][0]["back"] = "missingParent"
try:
    generator.build_graph(inventory, bad)
except ValueError as error:
    assert "Expected one field" in str(error)
else:
    raise AssertionError("Missing back-reference was accepted")

with tempfile.TemporaryDirectory(prefix="common-diagram-test-") as temporary:
    folder = Path(temporary)
    fixture = folder / "fixture"
    fixture.mkdir()
    (fixture / "A.cs").write_text('''namespace Example;
public partial class Parent {
    private Child? one;
    private List<Child> many;
    private Child left, right;
    private Func<Child> factory;
    public Child Stored { get; set; }
    public Child Computed => one!;
    private static Child Global;
}
public class Child { public Parent? parent; }
public record Snapshot(Child Item, Child[] Items);
''')
    (fixture / "B.cs").write_text('''namespace Example;
public partial class Parent { private Child? other; }
''')
    destination = folder / "fixture.json"
    dotnet = sys.argv[1] if len(sys.argv) > 1 else "dotnet"
    subprocess.run([dotnet, str(root / ".build/artifacts/bin/Extractor/debug/Extractor.dll"), str(fixture), str(destination)], check=True)
    parsed = json.loads(destination.read_text())
    found = {t["id"]: t for t in parsed["types"]}
    assert set(found) == {"Parent", "Child", "Snapshot"}
    fields = {f["name"]: f for f in found["Parent"]["fields"]}
    assert set(fields) == {"one", "many", "left", "right", "factory", "Stored", "other"}
    assert fields["many"]["many"] and not fields["one"]["many"]
    assert fields["factory"]["deferred"]
    assert len(found["Parent"]["sources"]) == 2
    assert {f["name"] for f in found["Snapshot"]["fields"]} == {"Item", "Items"}
    assert found["Snapshot"]["fields"][1]["many"]
    _, fixture_edges = generator.build_graph(parsed, {"relationships": []})
    assert not any(e["field"] == "factory" for e in fixture_edges)

    # Freshness checks refuse before rendering any artifact.
    for target in ("declaration", "lifecycle"):
        stale_inventory = copy.deepcopy(inventory)
        stale_review = copy.deepcopy(reviewed)
        if target == "declaration":
            stale_inventory["files"][0]["sha256"] = "0" * 64
        else:
            first = next(iter(stale_review["reviewed_files"]))
            stale_review["reviewed_files"][first] = "0" * 64
        try:
            generator.validate_sources(repository, stale_inventory, stale_review)
        except ValueError as error:
            expected = "Re-extract" if target == "declaration" else "Re-review"
            assert expected in str(error)
        else:
            raise AssertionError("Stale " + target + " evidence was accepted")

print("PASS: real inventory and SVG coverage; partials, record properties, multiplicity, stored/computed/static separation, callback exclusion; missing reverse and stale declaration/lifecycle evidence refuse.")
