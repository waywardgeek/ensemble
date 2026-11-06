"""Grader-owned inputs and expectations; no candidate parser supplies the oracle."""
import uuid


def buffered_fixture():
    token = uuid.uuid4().hex
    marker, human = 'answer-' + token, 'human-' + token
    first = ('  界 😀 ' + token + '\n\n\tend;  \n'
             '<script>window.copyFixtureExecuted = true</script>\n')
    long_code = ('  long code 界 ' + token + '\n') * 200
    markdown = (
        marker + '\n\n| Key | Value |\n| --- | --- |\n| fixture | safe |\n\n'
        '[safe](https://example.invalid/safe) [bad](javascript:alert(1))\n\n'
        f'![remote](https://example.invalid/{token})\n\n'
        f'<img src="https://example.invalid/{token}" '
        'onerror="window.copyFixtureExecuted = true">\n\n'
        'Inline `code`.\n\n    indented code\n\n'
        '```csharp\r\n' + first.replace('\n', '\r\n') + '```\r\n\r\n'
        '> ```text\r\n>   quoted\r\n>\r\n> ```\r\n\r\n'
        '```text\n' + long_code + '```\n\n```\n```\n\n```js\nunfinished ' + token)
    return dict(marker=marker, human=human,
                prompt=human + '\n\n```text\nhuman code\n```',
                markdown=markdown, payloads=[first, '  quoted\n\n', long_code, ''])
