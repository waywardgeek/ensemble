#!/usr/bin/env python3
"""Generate local Graphviz diagrams from Roslyn declarations and reviewed ownership.

No package dependencies. No application execution, network, or credential access.
The field inventory is mechanical; cascade-delete semantics are reviewed input.
"""
import argparse
import hashlib
import json
import re
from pathlib import Path
import subprocess
import textwrap

COLORS = {"strong": "#164f86", "weak": "#a74b00", "data": "#687682"}


def quote(value):
    return json.dumps(str(value), ensure_ascii=False)


def field_of(types, owner, name):
    matches = [f for f in types[owner]["fields"] if f["name"] == name]
    if len(matches) != 1:
        raise ValueError(f"Expected one field: {owner}.{name}")
    return matches[0]


def build_graph(inventory, reviewed):
    types = {t["id"]: t for t in inventory["types"] if not t["isStatic"]}
    edges = []
    covered = set()
    for relationship in reviewed["relationships"]:
        edge = dict(relationship)
        owner, child, name = edge["owner"], edge["child"], edge["field"]
        field = field_of(types, owner, name)
        if child not in field["references"] or field["deferred"]:
            raise ValueError(f"Reviewed relationship no longer matches: {owner}.{name}")
        if edge["kind"] not in ("strong", "weak"):
            raise ValueError("Unknown ownership kind")
        edge["many"] = field["many"]
        edge["source"] = field["source"]
        covered.add((owner, name, child))
        if edge["back"]:
            reverse = field_of(types, child, edge["back"])
            if owner not in reverse["references"]:
                raise ValueError(f"Missing reverse membership: {child}.{edge['back']}")
            covered.add((child, edge["back"], owner))
        edges.append(edge)
    for owner, declaration in types.items():
        for field in declaration["fields"]:
            # A function signature is not containment. A Task's result is a
            # deferred value, not a direct relationship to that result object.
            if field["deferred"]:
                continue
            for child in field["references"]:
                if child not in types or (owner, field["name"], child) in covered:
                    continue
                edges.append(dict(owner=owner, child=child, field=field["name"],
                                  back=None, kind="data", many=field["many"],
                                  source=field["source"], evidence="", note="Stored type reference; no cascade-delete claim."))
    return types, edges


def dot_graph(title, names, edges, types, lifecycle):
    lines = ["digraph Common {", 'graph [rankdir=LR, bgcolor="white", pad=0.4, nodesep=0.30, ranksep=1.1, overlap=false, splines=spline, outputorder=edgesfirst, fontname="Helvetica", fontsize=20, labelloc=t];',
             'node [shape=box, style="rounded,filled", fontname="Helvetica", fontsize=14, margin="0.16,0.10", color="#617384", penwidth=1.2];',
             'edge [fontname="Helvetica", fontsize=10, arrowsize=1.25];',
             'label=' + quote(title + "\nHeavy blue: cascade-delete   |   Thin orange: weak   |   Gray: stored data\nOpen triangle: one slot (may be null)   |   Filled triangle: many   |   Arrows: parent / holder → child / value") + ";"]
    for name in sorted(names):
        t = types[name]
        label = name.replace(".", ".\n")
        if t["bases"]:
            label += "\n(base: " + ", ".join(t["bases"]) + ")"
        fill = "#e3effb" if name in lifecycle else "#f3f5f6"
        if t["kind"] == "interface":
            fill = "#fff0dc"
        source = t["sources"][0]
        tooltip = f"{name} — agent/Common/{source['file']}:{source['line']}"
        lines.append(f"{quote(name)} [label={quote(label)}, fillcolor={quote(fill)}, tooltip={quote(tooltip)}];")
    for index, e in enumerate(edges):
        if e["owner"] not in names or e["child"] not in names:
            continue
        kind = e["kind"]
        width = 3.0 if kind == "strong" else 1.1
        # Use an explicit arrow stroke width independent of the node color.
        arrow = "normal" if e["many"] else "onormal"
        label = "\n".join(textwrap.wrap(e["field"], 25))
        tooltip = f"{e['owner']}.{e['field']} → {e['child']}; {kind}; {'many' if e['many'] else 'one slot'}"
        if e["back"]:
            tooltip += f"; reverse: {e['child']}.{e['back']}"
        if e["note"]:
            tooltip += "; " + e["note"]
        constraint = "false" if kind == "weak" else "true"
        lines.append(f"{quote(e['owner'])} -> {quote(e['child'])} [id=\"rel{index}\", color={quote(COLORS[kind])}, fontcolor={quote(COLORS[kind])}, penwidth={width}, arrowhead={arrow}, label={quote(label)}, tooltip={quote(tooltip)}, constraint={constraint}];")
    lines.append("}")
    return "\n".join(lines) + "\n"


def print_graph(types, relationships, selected_names=None, inheritance=False, rankdir="LR"):
    """Wrap labels for print, preserving all supplied edges and optional subtypes."""
    names = set(selected_names) if selected_names is not None else {e[key] for e in relationships for key in ("owner", "child")}
    lines = ["digraph Print {",
             f'graph [rankdir={rankdir}, pad=0.05, nodesep=0.12, ranksep=0.22, splines=spline, outputorder=edgesfirst];',
             'node [shape=box, style="rounded,filled", fillcolor="white", color="black", fontname="Helvetica", fontsize=20, margin="0.07,0.05"];',
             'edge [arrowsize=1.1];']
    for name in sorted(names):
        words = re.findall(r"[A-Z]+(?=[A-Z][a-z]|$)|[A-Z]?[a-z]+", name.replace("Agent.", ""))
        rows = []
        current = ""
        for word in words:
            if current and len(current) + len(word) > 11:
                rows.append(current)
                current = word
            else:
                current += word
        rows.append(current)
        label = "\n".join(rows)
        if name == "Agent.WorkProductTransfer":
            label = "WorkProduct\nTransfer"
        lines.append(f"{quote(name)} [label={quote(label)}];")
    for e in relationships:
        field = field_of(types, e["owner"], e["field"])
        width = 3 if e["kind"] == "strong" else 1
        arrow = "normal" if field["many"] else "onormal"
        constraint = "false" if e["kind"] == "weak" else "true"
        color = COLORS[e["kind"]]
        lines.append(f'{quote(e["owner"])} -> {quote(e["child"])} [penwidth={width}, color={quote(color)}, arrowhead={arrow}, constraint={constraint}];')
    if inheritance:
        # Generalization is not containment: keep it visually distinct and
        # derive it from declarations, never from an invented ownership edge.
        for name in sorted(names):
            for base in types[name]["bases"]:
                if base in names:
                    lines.append(f'{quote(name)} -> {quote(base)} [style=dashed, color="#704090", arrowhead=empty, constraint=false];')
    lines.append("}")
    return "\n".join(lines) + "\n"


def validate_sources(repository, inventory, reviewed):
    # Reject stale inventories and changed reviewed implementation evidence.
    for source in inventory["files"]:
        path = repository / "agent/Common" / source["file"]
        if hashlib.sha256(path.read_bytes()).hexdigest() != source["sha256"]:
            raise ValueError(f"Re-extract changed declaration: {path}")
    actual = {p.name for p in (repository / "agent/Common").glob("*.cs")}
    if actual != {f["file"] for f in inventory["files"]}:
        raise ValueError("Common file set changed: re-extract declarations")
    for name, digest in reviewed["reviewed_files"].items():
        if hashlib.sha256((repository / name).read_bytes()).hexdigest() != digest:
            raise ValueError(f"Re-review ownership before regenerating: {name}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("repository", type=Path)
    parser.add_argument("--output", type=Path, default=Path(__file__).resolve().parents[2] / "docs/common-diagrams")
    args = parser.parse_args()
    out = args.output
    source = Path(__file__).resolve().parent
    out.mkdir(parents=True, exist_ok=True)
    inventory = json.loads((out / "inventory.json").read_text())
    reviewed = json.loads((source / "relationships.json").read_text())
    validate_sources(args.repository, inventory, reviewed)
    types, edges = build_graph(inventory, reviewed)
    lifecycle = {e[k] for e in edges if e["kind"] != "data" for k in ("owner", "child")}
    views = [("ownership", "Lifecycle ownership", lifecycle, [e for e in edges if e["kind"] != "data"])]
    groups = {
        "history": ("History and recorded data", ["History", "Event", "Context", "Entry", "Part", "ToolResultPart", "TextPart", "OpaquePart", "ToolCallPart", "BlobPart", "RedactedPart", "HandoffPart", "RecallPart", "SkillPart", "HistorySnapshot"]),
        "models": ("Models, configuration and authentication", ["Ensemble", "ModelVendor", "ModelEndpoint", "Model", "ModelAuthentication", "Credentials", "CodexCredentials", "Configuration", "ModelIdentity", "ModelFeatures", "ModelRequestPolicy"]),
        "memory": ("Memory, compression and recall data", ["Memory", "Recaller", "MemorySettings", "MemoryCapture", "MemoryStage", "MemoryWorkerResult", "MemoryCompletion", "Context", "CompressionOutcome", "CompressionLaunch", "PendingCompression", "RecallIndex", "RecallDocument", "RecallHit", "RecallCompletion", "AgentSettings"]),
        "tools": ("Tools, skills, jobs and MCP", ["Agent", "Jobs", "Job", "Tool", "Skills", "Mcp", "McpClient", "McpBrowser", "McpProcessTransport", "McpStreamTransport", "McpHttpTransport", "McpWebSocketTransport", "IMcpTransport"]),
    }
    for key, (title, seeds) in groups.items():
        names = set(seeds)
        # Expand outgoing data references once. Do not pull the entire graph in
        # merely because a subsystem has a back-path to the Agent/root.
        for edge in edges:
            if edge["owner"] in seeds and edge["kind"] == "data" and edge["owner"] != "Agent":
                names.add(edge["child"])
        views.append((key, title, names, edges))
    views.append(("all", "All Common structures — declaration graph", set(types), edges))
    svgs = {}
    for key, title, names, selected in views:
        dot = dot_graph(title, names, selected, types, lifecycle)
        (out / f"{key}.dot").write_text(dot)
        for extension in ("svg", "pdf"):
            subprocess.run(["dot", f"-T{extension}", str(out / f"{key}.dot"), "-o", str(out / f"{key}.{extension}")], check=True)
        svg = (out / f"{key}.svg").read_text()
        svgs[key] = svg[svg.index("<svg"):]
    print_source = print_graph(types, reviewed["relationships"])
    (out / "ownership-print.dot").write_text(print_source)
    figure = args.repository / "book/figures/common-ownership.svg"
    figure.parent.mkdir(parents=True, exist_ok=True)
    subprocess.run(["dot", "-Tsvg", str(out / "ownership-print.dot"), "-o", str(figure)], check=True)
    # These are focused printed sheets, not a second relationship authority.
    # All stored links between selected types come from the same graph above.
    context_views = json.loads((source / "context-views.json").read_text())
    for key, view in context_views.items():
        names = set(view["types"])
        if not names <= types.keys():
            raise ValueError(f"Unknown type in {key} sheet: {names - types.keys()}")
        selected = [e for e in edges if e["owner"] in names and e["child"] in names]
        diagram = print_graph(types, selected, names, view["inheritance"], view["rankdir"])
        dot_path = out / f"{key}-print.dot"
        dot_path.write_text(diagram)
        for extension in ("svg", "pdf"):
            subprocess.run(["dot", f"-T{extension}", str(dot_path), "-o", str(out / f"{key}.{extension}")], check=True)
        (args.repository / f"book/figures/common-{key}.svg").write_bytes((out / f"{key}.svg").read_bytes())
        svg = (out / f"{key}.svg").read_text()
        svgs[key] = svg[svg.index("<svg"):]
    # EPUB readers can mis-scale SVG viewports. Raster companions share exactly
    # the printed DOT layout; 240 DPI provides detail for reader zoom.
    for key in ("ownership", "context", "contents"):
        subprocess.run([
            "dot", "-Tpng", "-Gdpi=240", str(out / f"{key}-print.dot"),
            "-o", str(args.repository / f"book/figures/common-{key}.png"),
        ], check=True)

    table = ["# Ensemble Common relationships", "", "Generated from source declarations plus reviewed lifecycle semantics.", "", "One means one slot, possibly null; many means a collection. Gray data links do not assert destruction ownership. Reverse memberships are not drawn twice.", "", "| Parent / holder | Field | Child / value | Kind | Multiplicity | Reverse field | Declaration |", "|---|---|---|---|---|---|---|"]
    for e in edges:
        loc = e["source"]
        table.append(f"| {e['owner']} | {e['field']} | {e['child']} | {e['kind']} | {'many' if e['many'] else 'one'} | {e['back'] or '—'} | {loc['file']}:{loc['line']} |")
    (out / "relationships.md").write_text("\n".join(table) + "\n")
    payload = dict(types=types, edges=edges, svgs=svgs, revision=reviewed["source_revision"])
    template = (source / "viewer-template.html").read_text()
    # Prevent a source comment from ending the JSON script element.
    encoded = json.dumps(payload).replace("<", "\\u003c").replace(">", "\\u003e").replace("&", "\\u0026")
    (out / "index.html").write_text(template.replace("__GRAPH_DATA__", encoded))
    print(f"Generated {len(svgs)} SVG/PDF views; {len(types)} nonstatic types, {len(edges)} arrows, {len(reviewed['relationships'])} reviewed lifecycle relationships.")


if __name__ == "__main__":
    main()
