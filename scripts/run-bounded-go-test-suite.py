#!/usr/bin/env python3
"""Run an unchanged Go suite with independent driver and test deadlines."""

import argparse
import importlib.util
import json
from pathlib import Path
import shutil
import sys


def load_script(name):
    path = Path(__file__).resolve().with_name(name)
    spec = importlib.util.spec_from_file_location(name.replace("-", "_"), path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def go_exec_argument(arguments):
    # cmd/internal/quoted.Split accepts one unescaped quoted field at a time.
    # This is Go's -exec syntax; no shell or Windows command-string parsing.
    fields = []
    for argument in arguments:
        if argument and not any(c.isspace() or c in "\"'" for c in argument):
            fields.append(argument)
        elif '"' not in argument:
            fields.append('"' + argument + '"')
        elif "'" not in argument:
            fields.append("'" + argument + "'")
        else:
            raise ValueError("Go -exec cannot quote an argument containing both quote characters")
    return " ".join(fields)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--report", type=Path, required=True)
    parser.add_argument("--test-output", type=Path, required=True)
    parser.add_argument("--go", default="go")
    parser.add_argument("--suite-timeout-seconds", type=float, default=2400)
    parser.add_argument("--test-timeout-seconds", type=float, default=1320)
    parser.add_argument("--heartbeat-seconds", type=float, default=30)
    parser.add_argument("arguments", nargs=argparse.REMAINDER)
    options = parser.parse_args()
    arguments = options.arguments
    if arguments[:1] == ["--"]:
        arguments = arguments[1:]
    if not arguments:
        arguments = ["-tags=server", "-count=1", "-timeout=20m", ".", "./server", "./legacy"]
    report = options.report.resolve()
    report.mkdir(parents=True, exist_ok=True)
    test_output = options.test_output.resolve()
    test_output.mkdir(parents=True, exist_ok=True)
    wrapper = Path(__file__).resolve().with_name("run-preserved-go-test.py")
    exec_argument = go_exec_argument([
        sys.executable, str(wrapper), "--report", str(report),
        "--timeout-seconds", str(options.test_timeout_seconds),
        "--heartbeat-seconds", str(options.heartbeat_seconds),
    ])
    go = shutil.which(options.go)
    if go is None:
        raise FileNotFoundError(options.go)
    command = [go, "test", "-work", "-o=" + str(test_output), "-json", "-exec=" + exec_argument] + arguments
    with (report / "test-command.json").open("x", encoding="utf-8") as stream:
        json.dump(dict(command=command, cwd=str(Path.cwd()),
                       suite_timeout_seconds=options.suite_timeout_seconds,
                       test_timeout_seconds=options.test_timeout_seconds), stream, indent=2)
        stream.write("\n")
    supervisor = load_script("run-preserved-go-test.py")
    with (report / "test.jsonl").open("xb") as out, (report / "test.stderr.log").open("xb") as err:
        code = supervisor.run(
            report, command, options.suite_timeout_seconds, options.heartbeat_seconds,
            retain_binary=False, stdout=out, stderr=err,
        )
    with (report / "test.exit").open("x", encoding="utf-8") as stream:
        stream.write(str(code) + "\n")
    # Aggregate the pre-execution records even on failed or timed-out suites.
    # The ordinary gate still requires all three completed package executables.
    collector = load_script("preserve-go-test-binaries.py")
    collector.preserve(report, test_output)
    return code


if __name__ == "__main__":
    raise SystemExit(main())
