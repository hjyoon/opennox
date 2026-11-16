"""Replay the exact primary Windows/386 PE against original vectors under GC."""
from pathlib import Path
import argparse
import collections
import datetime
import gzip
import hashlib
import json
import os
import platform
import re
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument("--report", type=Path, required=True)
args = parser.parse_args()
report = args.report.resolve()
repo = Path(__file__).resolve().parents[1]
assert platform.system() == "Windows" and os.environ["GOARCH"] == "386"
assert os.environ.get("GOEXPERIMENT", "") == ""
out = report / "global-record-original-GOGC10"
out.mkdir()


def digest(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


source = (report / "source-revision.txt").read_text(encoding="utf-8-sig").strip()
assert subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repo, text=True).strip() == source
manifest = json.loads((report / "test-binaries.json").read_text(encoding="utf-8-sig"))
original = [row for row in manifest if row["name"] == "legacy.test.exe"]
assert len(original) == 1
binary = report / "test-binaries/legacy.test.exe"
before = digest(binary)
assert before == original[0]["sha256"] and binary.stat().st_size == original[0]["bytes"]

fixture = repo / "src/legacy/testdata/global_record_original_v1.json.gz"
assert digest(fixture) == "d5e527a06bf8e0126b4258d0563010675120a8f6235c21e3ade6bac60cfc8fef"
rows = json.loads(gzip.decompress(fixture.read_bytes()))
names = {row["input"]["name"] for row in rows}
assert len(names) == len(rows) == 1088
repetitions = 3
expected = collections.Counter()
for root in ["TestGlobalRecordNativeLayout", "TestGlobalRecordNativeOriginal",
             "TestGlobalRecordNativeWrapper", "TestGlobalRecordNativeBoundedCParent"]:
    expected[root] = repetitions
for route in range(3):
    for placement in [0, 64, 512]:
        for name in names:
            expected[f"TestGlobalRecordNativeOriginal/route{route}/place{placement}/{name}"] = repetitions
assert sum(expected.values()) == 29388
command = [str(binary), "-test.v", "-test.count=" + str(repetitions),
           "-test.timeout=8m", "-test.run=^TestGlobalRecordNative"]
env = dict(os.environ, GOGC="10", GOEXPERIMENT="")
intent = dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(), source=source,
              command=command, cwd=str(repo / "src/legacy"), GOGC=10,
              repetitions=repetitions, exact_primary_PE_sha256=before, rebuilt=False,
              original_contexts=1088, expected_terminal_events=29388)
(out / "intent.json").write_text(json.dumps(intent, indent=2) + "\n")
timed_out = False
with (out / "native.stdout.log").open("xb") as stdout, (out / "native.stderr.log").open("xb") as stderr:
    try:
        result = subprocess.run(command, cwd=repo / "src/legacy", env=env,
                                stdout=stdout, stderr=stderr, timeout=540)
        code = result.returncode
    except subprocess.TimeoutExpired:
        code, timed_out = None, True
seen = collections.Counter()
actions = collections.Counter()
unfinished = set()
terminal = re.compile(r"^\s*--- (PASS|FAIL|SKIP): (\S+) \(")
run = re.compile(r"^=== RUN\s+(\S+)\s*$")
for line in (out / "native.stdout.log").open(encoding="utf-8", errors="replace"):
    if match := run.match(line):
        unfinished.add(match[1])
    if match := terminal.match(line):
        action, name = match.groups()
        actions[action.lower()] += 1
        unfinished.discard(name)
        if action == "PASS":
            seen[name] += 1
after = digest(binary)
record = dict(intent, native_exit_code=code, timed_out=timed_out,
              PE_hash_after_replay=after, test_actions=dict(actions),
              exact_original_terminal_identities_and_multiplicities_equal=seen == expected,
              missing_terminal_events=sum((expected - seen).values()),
              extra_terminal_events=sum((seen - expected).values()),
              unfinished_tests=sorted(unfinished), stdout_sha256=digest(out / "native.stdout.log"),
              stderr_sha256=digest(out / "native.stderr.log"))
(out / "summary.json").write_text(json.dumps(record, indent=2) + "\n")
print(json.dumps(record), flush=True)
assert code == 0 and not timed_out and not unfinished
assert actions == {"pass": 29388} and seen == expected and after == before
