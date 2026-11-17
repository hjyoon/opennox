"""Replay the retained primary Windows/386 PE against native spawn C entries."""
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
out = report / "monster-spawn-registry-GOGC10"
out.mkdir()


def digest(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


source = (report / "source-revision.txt").read_text(encoding="utf-8-sig").strip()
assert subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repo, text=True).strip() == source
manifest = json.loads((report / "test-binaries.json").read_text(encoding="utf-8-sig"))
primary = [row for row in manifest if row["name"] == "legacy.test.exe"]
assert len(primary) == 1
binary = report / "test-binaries/legacy.test.exe"
before = digest(binary)
assert before == primary[0]["sha256"] and binary.stat().st_size == primary[0]["bytes"]
fixture = repo / "src/legacy/testdata/monster_spawn_raw_links_original_v1.json.gz"
assert digest(fixture) == "8c7799474ba598cbb6a574034b2e81a4a5b947d7f151bbcecb2c4d5254b7544e"
raw = gzip.decompress(fixture.read_bytes())
assert hashlib.sha256(raw).hexdigest() == "bbbca8e6171e9805e9ea7ad6997ffc5fc59d2bf955d1bb1f56d5279d37a0430f"
rows = json.loads(raw)
assert len(rows) == 129 and sum(row["entry"] == "unlink" for row in rows) == 119
repetitions = 3
prefix = "TestMonsterSpawnRegistryNative"
names = {prefix + suffix for suffix in ["Layout", "OriginalRawLinks", "EntriesAndGC", "LiveScratchQuery", "DetachedResetAndZombie"]}
for route in [0, 1]:
    for place in [0, 512, 4096]:
        case = f"/route{route}/place{place}"
        for suffix in ["OriginalRawLinks", "EntriesAndGC", "LiveScratchQuery", "DetachedResetAndZombie"]:
            names.add(prefix + suffix + case)
        for row in rows:
            names.add(prefix + "OriginalRawLinks" + case + "/" + row["name"])
expected = collections.Counter({name: repetitions for name in names})
assert len(names) == 803 and sum(expected.values()) == 2409
command = [str(binary), "-test.v", "-test.count=" + str(repetitions),
           "-test.timeout=8m", "-test.run=^TestMonsterSpawnRegistryNative"]
env = dict(os.environ, GOGC="10", GOEXPERIMENT="")
intent = dict(utc=datetime.datetime.now(datetime.timezone.utc).isoformat(), source=source,
              command=command, cwd=str(repo / "src/legacy"), GOGC=10, repetitions=repetitions,
              exact_primary_PE_sha256=before, rebuilt=False, patched=False,
              typed_public_C_entries=16, original_raw_link_cases=129,
              raw_link_routes=2, native_owner_placements=[0, 512, 4096],
              expected_terminal_events=2409, expected_getter_reads_from_passing_paths_per_repetition=13860,
              source_files={str(path.relative_to(repo)): digest(path) for path in [
                  fixture, repo / "src/legacy/monster_spawn_registry_native_fixture.go",
                  repo / "src/legacy/monster_spawn_registry_native_test.go",
                  repo / "src/legacy/monster_spawn_registry_native_entry.go",
                  repo / "src/legacy/monster_spawn_registry_50d780.h"]},
              original_GAME_EXE_executed_by_this_replay=False, new_original_algorithm_credit=0)
(out / "intent.json").write_text(json.dumps(intent, indent=2) + "\n")
timed_out = False
with (out / "native.stdout.log").open("xb") as stdout, (out / "native.stderr.log").open("xb") as stderr:
    try:
        result = subprocess.run(command, cwd=repo / "src/legacy", env=env,
                                stdout=stdout, stderr=stderr, timeout=540)
        code = result.returncode
    except subprocess.TimeoutExpired:
        code, timed_out = None, True
seen, actions = collections.Counter(), collections.Counter()
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
record = dict(intent, native_exit_code=code, timed_out=timed_out, PE_hash_after_replay=after,
              test_actions=dict(actions), exact_terminal_identities_and_multiplicities_equal=seen == expected,
              missing_terminal_events=sum((expected - seen).values()), extra_terminal_events=sum((seen - expected).values()),
              unfinished_tests=sorted(unfinished), stdout_sha256=digest(out / "native.stdout.log"),
              stderr_sha256=digest(out / "native.stderr.log"))
(out / "summary.json").write_text(json.dumps(record, indent=2) + "\n")
print(json.dumps(record), flush=True)
assert code == 0 and not timed_out and not unfinished
assert actions == {"pass": 2409} and seen == expected and after == before
