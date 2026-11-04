#!/usr/bin/env python3
"""Retain Unix test executables from a completed go test -work -json run."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil


def preserve(report, prefix="test"):
    work_paths = set()
    stderr = report / (prefix + ".stderr.log")
    for line in stderr.read_text(encoding="utf-8-sig").splitlines():
        if line.startswith("WORK="):
            work_paths.add(line.removeprefix("WORK="))
    with (report / (prefix + ".jsonl")).open("rb") as stream:
        for line in stream:
            if b"WORK=" not in line:
                continue
            event = json.loads(line)
            if event.get("Action") == "build-output":
                for output in event.get("Output", "").splitlines():
                    if output.startswith("WORK="):
                        work_paths.add(output.removeprefix("WORK="))
    if len(work_paths) != 1:
        raise ValueError(f"Expected one Go work directory, got {sorted(work_paths)}")
    work = Path(next(iter(work_paths)))
    destination = report / (prefix + "-binaries")
    destination.mkdir()
    binaries = []
    for binary in sorted(work.rglob("*.test")):
        relative = binary.relative_to(work)
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(binary, target)
        with target.open("rb") as stream:
            digest = hashlib.file_digest(stream, "sha256").hexdigest()
        binaries.append(dict(path=str(relative), name=binary.name,
                             bytes=target.stat().st_size, sha256=digest))
    with (report / (prefix + "-work-directory.json")).open("x") as stream:
        json.dump(dict(path=str(work), prefix=prefix), stream, indent=2)
        stream.write("\n")
    with (report / (prefix + "-binaries.json")).open("x") as stream:
        json.dump(binaries, stream, indent=2)
        stream.write("\n")
    required = {"v1.test", "server.test", "legacy.test"}
    missing = required - {row["name"] for row in binaries}
    if missing:
        raise ValueError(f"Missing required package executables: {sorted(missing)}")
    print(json.dumps(dict(retained=len(binaries), required_packages_complete=True,
                          manifest=str(report / (prefix + "-binaries.json")))))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("report", type=Path)
    parser.add_argument("--prefix", default="test")
    args = parser.parse_args()
    preserve(args.report, args.prefix)
