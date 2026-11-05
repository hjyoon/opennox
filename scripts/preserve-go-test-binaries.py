#!/usr/bin/env python3
"""Preserve the executables from an exact go test -work -json invocation."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil


def write_new_or_verify(path, value):
    if path.exists():
        if json.loads(path.read_text(encoding="utf-8-sig")) != value:
            raise ValueError(f"Refusing to replace different preservation metadata: {path}")
        return
    with path.open("x", encoding="utf-8") as stream:
        json.dump(value, stream, indent=2)
        stream.write("\n")


def preserve(report, executed_directory=None, *, allow_incomplete=False, expected_count=3):
    report = Path(report)
    work_paths = set()
    sources = []
    stderr = report / "test.stderr.log"
    for line in (stderr.read_text(encoding="utf-8-sig") if stderr.exists() else "").splitlines():
        if line.startswith("WORK="):
            work_paths.add(line.removeprefix("WORK="))
            sources.append("test.stderr.log")
    # Go 1.26 emits -work through build-output when -json is enabled.
    stdout = report / "test.jsonl"
    with stdout.open("rb") if stdout.exists() else open(os.devnull, "rb") as stream:
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
    problems = []
    if len(work_paths) != 1:
        problems.append(f"Expected one retained Go work directory, got {sorted(work_paths)}")
    work = Path(next(iter(work_paths))) if len(work_paths) == 1 else None
    destination = report / "test-binaries"
    destination.mkdir(exist_ok=True)
    binaries = {}
    records = []
    for path in sorted((report / "test-binary-records").glob("*.json")):
        row = json.loads(path.read_text(encoding="utf-8-sig"))
        if Path(row["name"]).name != row["name"] or not row["preservation_before_execution"]:
            raise ValueError(f"Invalid pre-execution binary record: {path}")
        records.append(row)
        target = destination / row["name"]
        with target.open("rb") as stream:
            digest = hashlib.file_digest(stream, "sha256").hexdigest()
        actual = dict(name=row["name"], bytes=target.stat().st_size, sha256=digest)
        if actual["bytes"] != row["bytes"] or digest != row["sha256"]:
            raise ValueError(f"Pre-execution binary changed: {target}")
        if row["name"] in binaries and binaries[row["name"]] != actual:
            raise ValueError(f"Different actual executables share a name: {row['name']}")
        binaries[row["name"]] = actual
    search_directories = []
    if not records:
        # -o is installed after the test run. A cancelled suite may have only
        # its original WORK binaries, even when an output directory was given.
        for directory in (executed_directory, work):
            if directory is not None and Path(directory).is_dir():
                search_directories.append(Path(directory))
        for directory in search_directories:
            for binary in sorted(directory.rglob("*.test.exe")):
                with binary.open("rb") as stream:
                    digest = hashlib.file_digest(stream, "sha256").hexdigest()
                actual = dict(name=binary.name, bytes=binary.stat().st_size, sha256=digest)
                if binary.name in binaries and binaries[binary.name] != actual:
                    raise ValueError(f"Different test executables share a name: {binary.name}")
                target = destination / binary.name
                if target.exists():
                    with target.open("rb") as stream:
                        if hashlib.file_digest(stream, "sha256").hexdigest() != digest:
                            raise ValueError(f"Retained executable changed: {target}")
                else:
                    shutil.copyfile(binary, target)
                binaries[binary.name] = actual
    binaries = [binaries[name] for name in sorted(binaries)]
    if len(binaries) != expected_count:
        problems.append(f"Expected {expected_count} package executables, retained {len(binaries)}")
    pending_records = [row["name"] for row in records if not row.get("execution_started") or
                       row.get("phase") not in ("exited", "timed-out", "supervisor-error")]
    if pending_records:
        problems.append(f"Executions did not record completion: {pending_records}")
    write_new_or_verify(report / "test-work-directory.json", dict(
        path=str(work) if work is not None else None, sources=sources,
        executable_source="pre-execution -exec records" if records else [str(p) for p in search_directories],
    ))
    write_new_or_verify(report / "test-binaries.json", binaries)
    write_new_or_verify(report / "test-preservation-status.json", dict(
        expected_count=expected_count, retained_count=len(binaries),
        pre_execution_records=len(records), actual_executions_started=sum(bool(r.get("execution_started")) for r in records),
        pending_executions=pending_records, problems=problems,
        all_recorded_executables_hash_verified=True,
        complete_package_executable_set=not problems,
    ))
    if problems and not allow_incomplete:
        raise ValueError("; ".join(problems))
    print(json.dumps(binaries, indent=2))
    return binaries


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("report", type=Path)
    parser.add_argument("executed_directory", type=Path, nargs="?")
    parser.add_argument("--allow-incomplete", action="store_true")
    parser.add_argument("--expected-count", type=int, default=3)
    options = parser.parse_args()
    preserve(options.report, options.executed_directory,
             allow_incomplete=options.allow_incomplete, expected_count=options.expected_count)
