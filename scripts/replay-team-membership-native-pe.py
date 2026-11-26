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
out = report / "team-membership-original-GOGC10"
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

fixture = repo / "src/legacy/testdata/team_membership_queries_original_v1.json.gz"
assert digest(fixture) == "ca0158f2c2b279e17312902fd55c008be1e4178c667bcd169099b420352c801d"
rows = json.loads(gzip.decompress(fixture.read_bytes()))
assert len(rows) == len({row["input"]["name"] for row in rows}) == 2088
owners = repo / "src/legacy/testdata/team_membership_owner_native_v1.json.gz"
assert digest(owners) == "d62bd0b52667a1cbc0280608eaa0b2c0ea0cffb8d9c1d4041d77f61b73f38a71"
owner_rows = json.loads(gzip.decompress(owners.read_bytes()))
assert len(owner_rows) == 1044 and sum(len(r["events"]) for r in owner_rows) == 14877
repetitions = 3
expected = collections.Counter({name: repetitions for name in [
    "TestTeamMembershipNativeLayout",
    "TestTeamMembershipNativePinnedGoNodes",
    "TestTeamMembershipNativeOriginalQueries",
    "TestTeamMembershipNativeOwnerCompatibility"]})
for route in range(3):
    for place in [0, 512, 4096]:
        expected[f"TestTeamMembershipNativeOriginalQueries/route{route}/place{place}"] = repetitions
for parent in range(2):
    expected[f"TestTeamMembershipNativeOwnerCompatibility/parents{parent}"] = repetitions
    for program in ["link-commit-detach", "move-deferred-commit", "detach-ID-fallback", "reset-reuse"]:
        expected[f"TestTeamMembershipNativeOwnerCompatibility/parents{parent}/{program}"] = repetitions
expected_terminals = sum(expected.values())
command = [str(binary), "-test.v", "-test.count=" + str(repetitions),
           "-test.timeout=8m", "-test.run=^TestTeamMembershipNative"]
env = dict(os.environ, GOGC="10", GOEXPERIMENT="")
intent = dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(), source=source,
              command=command, cwd=str(repo / "src/legacy"), GOGC=10,
              repetitions=repetitions, exact_primary_PE_sha256=before, rebuilt=False,
              original_query_contexts=2088, incoming_owner_cases=1044,
              incoming_owner_observations=14877, native_calls_per_repetition=554364,
              actual_C_calls_per_repetition=369576, executed_owner_cases_per_repetition=2088, owner_observations_per_repetition=29754,
              expected_terminal_events=expected_terminals)
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
assert actions == {"pass": expected_terminals} and seen == expected and after == before
