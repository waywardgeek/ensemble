"""Print-only sidebar grouping: python3 -m unittest discover -s book -p 'test_*.py'."""
import json
from pathlib import Path
import subprocess
import unittest

FILTER = Path(__file__).resolve().parents[2] / 'scripts' / 'print-sidebars.lua'


def parse(text, filtered=True):
    args = ['pandoc', '--from=markdown', '--to=json']
    if filtered:
        args += ['--lua-filter', str(FILTER)]
    return json.loads(subprocess.check_output(args, input=text, text=True))['blocks']


class SidebarTests(unittest.TestCase):
    def test_sidebar_contains_prose_and_code_but_not_next_section(self):
        blocks = parse('''### Sidebar: Example

First paragraph.

```python
print("keep code")
```

### Regular section

Outside the sidebar.
''')
        self.assertEqual([b['t'] for b in blocks], ['Div', 'Header', 'Para'])
        attr, contents = blocks[0]['c']
        self.assertEqual(attr[1], ['print-sidebar'])
        self.assertEqual(dict(attr[2])['role'], 'note')
        self.assertEqual(dict(attr[2])['aria-labelledby'], contents[0]['c'][1][0])
        self.assertIn('print-sidebar-title', contents[0]['c'][1][1])
        self.assertEqual([b['t'] for b in contents], ['Header', 'Para', 'CodeBlock'])
        self.assertEqual(contents[-1]['c'][1], 'print("keep code")')

    def test_nested_heading_stays_inside_and_higher_heading_ends_sidebar(self):
        blocks = parse('### Sidebar: Example\n\nText.\n\n#### Detail\n\nMore.\n\n## Main\n\nOutside.')
        self.assertEqual([b['t'] for b in blocks], ['Div', 'Header', 'Para'])
        self.assertEqual([b['t'] for b in blocks[0]['c'][1]], ['Header', 'Para', 'Header', 'Para'])

    def test_sidebar_tabs_is_not_a_callout_and_code_is_not_a_heading(self):
        text = '## Sidebar tabs\n\nRegular text.\n\n```markdown\n### Sidebar: not a heading\n```'
        self.assertEqual(parse(text), parse(text, filtered=False))

    def test_adjacent_sidebars_and_end_of_document(self):
        blocks = parse('### Sidebar: One\n\nFirst.\n\n### Sidebar: Two\n\nLast.')
        self.assertEqual([b['t'] for b in blocks], ['Div', 'Div'])
        self.assertTrue(all(len(b['c'][1]) == 2 for b in blocks))


if __name__ == '__main__':
    unittest.main()
