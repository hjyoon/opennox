#!/usr/bin/env python3
"""Byte-verify a dedicated same-PID Go child immediately before exec."""
from pathlib import Path
import hashlib
import json
import os
import sys

binary, expected, receipt, *arguments = sys.argv[1:]
path = Path(binary).resolve(strict=True)
with path.open('rb') as stream:
    digest = hashlib.file_digest(stream, 'sha256').hexdigest()
assert digest == expected and path.is_file()
record = dict(before_exec=True, pid=os.getpid(),
              binary=dict(path=str(path), bytes=path.stat().st_size, sha256=digest),
              argv=[str(path), *arguments], compiler_or_original_or_game_executions=0)
with Path(receipt).open('x') as stream:
    json.dump(record, stream, indent=2)
    stream.write('\n')
    stream.flush()
    os.fsync(stream.fileno())
os.execve(str(path), record['argv'], os.environ)
