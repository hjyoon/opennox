#!/usr/bin/env python3
"""Summarize preserved Go JSON without losing failed or unfinished suites."""
import collections
import hashlib
import json
from pathlib import Path
import sys


def summarize(report):
    counts = collections.Counter()
    packages = {}
    failures = []
    active = set()
    build_failures = []
    build_warnings = []
    raw_sha = hashlib.sha256()
    raw_bytes = 0
    with (report / "test.jsonl").open("rb") as stream:
        for line in stream:
            raw_sha.update(line)
            raw_bytes += len(line)
            event = json.loads(line)
            action, name = event["Action"], event.get("Test", "")
            package = event.get("Package", event.get("ImportPath", ""))
            key = (package, name)
            if action == "run" and name:
                active.add(key)
            if action == "build-fail":
                build_failures.append(event)
            if action == "build-output" and "warning:" in event.get("Output", ""):
                build_warnings.append(event)
            if action not in ("pass", "skip", "fail"):
                continue
            if name:
                counts[action] += 1
                active.discard(key)
            else:
                if package in packages:
                    raise ValueError(f"Duplicate package terminal: {package}")
                packages[package] = action
            if action == "fail":
                failures.append(event)
    expected = set((report / "expected-packages.txt").read_text(encoding="utf-8-sig").splitlines())
    code = int((report / "test.exit").read_text(encoding="utf-8-sig"))
    result = dict(
        test_exit_code=code,
        test_actions=dict(counts),
        package_actions=packages,
        expected_packages=sorted(expected),
        missing_packages=sorted(expected - packages.keys()),
        unexpected_packages=sorted(packages.keys() - expected),
        unfinished_tests=sorted(active),
        failures=failures,
        build_failures=build_failures,
        build_warnings=build_warnings,
        raw_JSON_bytes=raw_bytes,
        raw_JSON_sha256=raw_sha.hexdigest(),
    )
    with (report / "summary.json").open("x", encoding="utf-8") as stream:
        json.dump(result, stream, indent=2)
        stream.write("\n")
    print(json.dumps({k: v for k, v in result.items() if k != "failures"}, indent=2))
    for event in failures[:20]:
        print("Failed terminal:", event["Package"], event.get("Test", ""))
    for event in build_warnings:
        print("Build warning:", event["Output"])
    complete = bool(expected) and set(packages) == expected and not active
    passed = complete and not failures and not build_failures and not build_warnings and set(packages.values()) == {"pass"}
    return code if code else (0 if passed else 1)


if __name__ == "__main__":
    sys.exit(summarize(Path(sys.argv[1])))
