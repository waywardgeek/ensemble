#!/usr/bin/env python3
"""Run every Ch4 assertion with the Chapter 5 asynchronous human prompt boundary.

The original Ch4 fixture interprets its next You> as completed model work. Ch5
must print that prompt immediately, while work is pending. This adapter follows
the acknowledged request ID through its completion before returning the display
chunk. No debugger, lifecycle, redaction, usage, or artifact assertion is skipped.
The original checker and its legacy behavior are left untouched.
"""
import re
import accept_ch04 as prior


class AsyncTerminal(prior.Terminal):
    def prompt(self, data=None):
        output=super().prompt(data)
        if data is None or data.lstrip().startswith(b'/'):
            return output
        accepted=re.search(r'Accepted\s+(\S+)\.',output)
        if accepted is None:
            raise AssertionError('ordinary chat submission lacked request acknowledgement')
        identity=accepted.group(1)
        completed=r'Request\s+'+re.escape(identity)+r'\s+\(success[^\n]*\).*?Assistant:'
        while not re.search(completed,output,re.S):
            output+=super().prompt()
        return output


if __name__=='__main__':
    prior.Terminal=AsyncTerminal
    raise SystemExit(prior.main())
