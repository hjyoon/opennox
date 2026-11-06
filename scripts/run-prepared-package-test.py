#!/usr/bin/env python3
"""Preserve a real Go package main under its admitted unique package identity."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import signal
import sys

sys.dont_write_bytecode = True


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--report', type=Path, required=True)
    parser.add_argument('--admission', type=Path, required=True)
    parser.add_argument('--timeout-seconds', type=float, default=1320)
    parser.add_argument('--heartbeat-seconds', type=float, default=30)
    parser.add_argument('command', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    command = args.command[1:] if args.command[:1] == ['--'] else args.command
    admission = json.loads(args.admission.read_bytes())
    directory = str(Path.cwd().resolve())
    package = admission['expected_packages_by_directory'][directory]
    key = hashlib.sha256(package.encode()).hexdigest()
    report = args.report/'package-mains'/key
    report.mkdir(parents=True, exist_ok=False)
    with (report/'package-context-before-exec.json').open('x') as f:
        json.dump(dict(package=package, directory=directory, package_key=key,
                       actual_command=command, preservation_phase='before-actual-main-execution'), f, indent=2)
        f.write('\n')
    path = Path(__file__).resolve().with_name('run-preserved-go-test.py')
    spec = importlib.util.spec_from_file_location('unchanged_package_supervisor', path)
    supervisor = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(supervisor)
    return supervisor.run(report, command, args.timeout_seconds, args.heartbeat_seconds)


if __name__ == '__main__':
    code = main()
    if code < 0:
        signal.signal(-code, signal.SIG_DFL)
        os.kill(os.getpid(), -code)
    raise SystemExit(code)
