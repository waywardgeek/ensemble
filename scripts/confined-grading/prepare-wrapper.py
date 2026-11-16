#!/usr/bin/env python3
"""Create a new host-owned Chapter 1 candidate wrapper; never run student code here.
Arguments: ABS_DOTNET ABS_LAUNCHER_DLL ABS_WORKSPACE ABS_CANDIDATE ABS_SDK_ROOT ABS_OUTPUT
The output directory must already exist and be outside the workspace.
"""
import os
from pathlib import Path
import shlex
import sys

if len(sys.argv) != 7 or any(not Path(value).is_absolute() for value in sys.argv[1:]):
    sys.exit(__doc__)
dotnet, launcher, workspace, candidate, sdk = [Path(value).resolve(strict=True) for value in sys.argv[1:6]]
requested = Path(sys.argv[6])
output = requested.parent.resolve(strict=True) / requested.name
if not workspace.is_dir() or not sdk.is_dir() or not candidate.is_file():
    sys.exit("Workspace/SDK must be directories and candidate must be a file.")
if not candidate.is_relative_to(workspace):
    sys.exit("Candidate must be contained in the writable workspace.")
if any(path.is_relative_to(workspace) for path in (dotnet, launcher, sdk, output)):
    sys.exit("Trusted launcher, SDK and output must be outside the candidate workspace.")
arguments = shlex.join([str(dotnet), str(launcher), str(workspace), str(candidate), str(sdk)])
text = '''#!/bin/sh
set -eu
# Preserve only the unchanged grader's public fake-provider inputs. These are
# validated by GradedCandidate before it prepares or executes the candidate.
exec /usr/bin/env -i \\
  ANTHROPIC_BASE_URL="${ANTHROPIC_BASE_URL-}" \\
  ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY-}" \\
  ANTHROPIC_MODEL="${ANTHROPIC_MODEL-}" \\
  ''' + arguments + '\n'
# Exclusive creation refuses existing files and symlinks, preserving prior work.
fd = os.open(output, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o700)
with os.fdopen(fd, 'w') as stream:
    stream.write(text)
print(output)
