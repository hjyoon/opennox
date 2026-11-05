#!/usr/bin/env python3
"""A go test -exec wrapper that retains the actual binary before startup."""

import argparse
import ctypes
import datetime
import hashlib
import json
import math
import os
from pathlib import Path
import shutil
import signal
import subprocess
import time
import uuid


def digest(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def write_json(path, value):
    temporary = path.with_name(path.name + "." + uuid.uuid4().hex + ".tmp")
    with temporary.open("x", encoding="utf-8") as stream:
        json.dump(value, stream, indent=2)
        stream.write("\n")
        stream.flush()
        os.fsync(stream.fileno())
    os.replace(temporary, path)


def stop_process_tree(process):
    if process.poll() is not None:
        return None
    if os.name == "nt":
        result = subprocess.run(
            ["taskkill", "/PID", str(process.pid), "/T", "/F"],
            stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=15,
        )
        return dict(command="taskkill /T /F", exit_code=result.returncode,
                    stdout=result.stdout.decode(errors="replace"),
                    stderr=result.stderr.decode(errors="replace"))
    # A test may create its own process group. Include descendants when the
    # bounded outer Go driver is stopped, rather than leaving those tests alive.
    descendants = []
    try:
        parents = {}
        for line in subprocess.check_output(["ps", "-axo", "pid=,ppid="], text=True).splitlines():
            pid, parent = map(int, line.split())
            parents.setdefault(parent, []).append(pid)
        pending = [process.pid]
        while pending:
            pid = pending.pop()
            descendants.append(pid)
            pending.extend(parents.get(pid, []))
        for pid in reversed(descendants):
            try:
                os.kill(pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
    except (OSError, subprocess.SubprocessError):
        os.killpg(process.pid, signal.SIGKILL)
    return dict(command="SIGKILL owned process tree", observed_pids=descendants)


def run(report, command, timeout_seconds, heartbeat_seconds, *, retain_binary=True,
        stdout=None, stderr=None):
    if not command:
        raise ValueError("Missing actual Go test executable")
    if not math.isfinite(timeout_seconds) or timeout_seconds <= 0:
        raise ValueError("Test deadline must be positive and finite")
    if not math.isfinite(heartbeat_seconds) or heartbeat_seconds <= 0:
        raise ValueError("Heartbeat interval must be positive and finite")
    report = report.resolve()
    binary = Path(os.path.abspath(command[0]))
    if not binary.is_file():
        raise FileNotFoundError(binary)
    binary_directory = report / "test-binaries"
    record_directory = report / ("test-binary-records" if retain_binary else "go-driver-records")
    event_directory = report / "test-supervisor-events"
    for directory in (record_directory, event_directory):
        directory.mkdir(parents=True, exist_ok=True)
    sha256 = digest(binary)
    retained = binary_directory / binary.name
    if retain_binary:
        binary_directory.mkdir(parents=True, exist_ok=True)
    if retain_binary and retained.exists():
        if retained.stat().st_size != binary.stat().st_size or digest(retained) != sha256:
            raise ValueError(f"Refusing to replace a different retained executable: {retained}")
    elif retain_binary:
        # Go rejects duplicate test basenames for -o directory package lists.
        # Exclusive creation also refuses an unexpected concurrent replacement.
        with binary.open("rb") as source, retained.open("xb") as target:
            shutil.copyfileobj(source, target)
            target.flush()
            os.fsync(target.fileno())
        if digest(retained) != sha256:
            raise ValueError(f"Executable copy changed before startup: {binary}")
    invocation = binary.name + "." + uuid.uuid4().hex
    record_path = record_directory / (invocation + ".json")
    events_path = event_directory / (invocation + ".jsonl")
    record = dict(
        name=binary.name, path=binary.name, bytes=binary.stat().st_size,
        sha256=sha256, executable_source=str(binary), resolved_source=str(binary.resolve()),
        retained_path=str(retained) if retain_binary else None,
        command=list(command), cwd=str(Path.cwd()), supervisor_pid=os.getpid(),
        timeout_seconds=timeout_seconds, heartbeat_seconds=heartbeat_seconds,
        preservation_before_execution=retain_binary, execution_started=False,
        process_kind="test" if retain_binary else "go-driver",
        forced_termination=False,
        environment={key: os.environ.get(key) for key in
                     ("GOARCH", "GOOS", "GOGC", "GODEBUG", "GOEXPERIMENT", "GOMAXPROCS")},
    )
    started = time.monotonic()

    def event(phase):
        record.update(phase=phase, utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
                      elapsed_seconds=round(time.monotonic() - started, 3))
        with events_path.open("a", encoding="utf-8") as stream:
            stream.write(json.dumps(record) + "\n")
            stream.flush()
            os.fsync(stream.fileno())
        write_json(record_path, record)

    event("preserved-before-start")
    process = None
    try:
        # Inherit Go's streams, environment, flags and package working directory.
        # Run the supplied original path; the retained copy is evidence only.
        process = subprocess.Popen(command, stdout=stdout, stderr=stderr,
                                   start_new_session=os.name != "nt")
        record.update(pid=process.pid, execution_started=True)
        event("running")
        deadline = time.monotonic() + timeout_seconds
        while True:
            if process.poll() is not None:
                code = process.returncode
                record.update(native_exit_code=code, wrapper_exit_code=code)
                event("exited")
                return code
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                record["forced_termination"] = True
                event("deadline-exceeded")
                record["termination"] = stop_process_tree(process)
                record["native_exit_code"] = process.wait(timeout=15)
                record["wrapper_exit_code"] = 124
                event("timed-out")
                return 124
            try:
                code = process.wait(timeout=min(heartbeat_seconds, remaining))
                record.update(native_exit_code=code, wrapper_exit_code=code)
                event("exited")
                return code
            except subprocess.TimeoutExpired:
                event("heartbeat")
    except BaseException as error:
        record["supervisor_error"] = repr(error)
        if process is not None and process.poll() is None:
            record["forced_termination"] = True
            record["termination"] = stop_process_tree(process)
            record["native_exit_code"] = process.wait(timeout=15)
        event("supervisor-error")
        raise


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--report", type=Path, required=True)
    parser.add_argument("--timeout-seconds", type=float, default=1320)
    parser.add_argument("--heartbeat-seconds", type=float, default=30)
    parser.add_argument("command", nargs=argparse.REMAINDER)
    options = parser.parse_args()
    command = options.command
    if command[:1] == ["--"]:
        command = command[1:]
    code = run(options.report, command, options.timeout_seconds, options.heartbeat_seconds)
    if os.name == "nt":
        # Preserve the original Windows DWORD exit status, including exceptions.
        os._exit(ctypes.c_int(code).value)
    if code < 0:
        signal.signal(-code, signal.SIG_DFL)
        os.kill(os.getpid(), -code)
    raise SystemExit(code)


if __name__ == "__main__":
    main()
