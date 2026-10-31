#!/usr/bin/env python3
"""Preserve the executables from an exact go test -work -json invocation."""
import hashlib
import json
from pathlib import Path
import shutil
import sys


def preserve(report):
    work_paths = set()
    sources = []
    for line in (report / "test.stderr.log").read_text(encoding="utf-8-sig").splitlines():
        if line.startswith("WORK="):
            work_paths.add(line.removeprefix("WORK="))
            sources.append("test.stderr.log")
    # Go 1.26 emits -work through build-output when -json is enabled.
    with (report / "test.jsonl").open("rb") as stream:
        for line in stream:
            if b"WORK=" not in line:
                continue
            event = json.loads(line)
            if event.get("Action") != "build-output":
                continue
            for output in event.get("Output", "").splitlines():
                if output.startswith("WORK="):
                    work_paths.add(output.removeprefix("WORK="))
                    sources.append("test.jsonl build-output")
    if len(work_paths) != 1:
        raise ValueError(f"Expected one retained Go work directory, got {sorted(work_paths)}")
    work = Path(next(iter(work_paths)))
    destination = report / "test-binaries"
    destination.mkdir()
    binaries = []
    for binary in sorted(work.rglob("*.test.exe")):
        target = destination / binary.name
        if target.exists():
            raise ValueError(f"Duplicate test executable name: {binary.name}")
        shutil.copyfile(binary, target)
        with target.open("rb") as stream:
            digest = hashlib.file_digest(stream, "sha256").hexdigest()
        binaries.append(dict(name=binary.name, bytes=target.stat().st_size, sha256=digest))
    with (report / "test-work-directory.json").open("x", encoding="utf-8") as stream:
        json.dump(dict(path=str(work), sources=sources), stream, indent=2)
        stream.write("\n")
    with (report / "test-binaries.json").open("x", encoding="utf-8") as stream:
        json.dump(binaries, stream, indent=2)
        stream.write("\n")
    if len(binaries) != 3:
        raise ValueError(f"Expected all three package executables, retained {len(binaries)}")
    print(json.dumps(binaries, indent=2))


if __name__ == "__main__":
    preserve(Path(sys.argv[1]))
