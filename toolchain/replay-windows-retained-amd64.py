"""Replay the exact failed Windows binaries; never build or patch runtime/tests."""

from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
import collections
import hashlib
import json
import os
import platform
import shutil
import subprocess
import urllib.parse
import urllib.request
import zipfile


report = Path(os.environ["REPORT_DIR"]).resolve()
report.mkdir(parents=True, exist_ok=True)
source = os.environ["SOURCE_REVISION"]
artifact_id = os.environ["RETAINED_ARTIFACT_ID"]
assert artifact_id.isdecimal()
assert subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip() == source
api = "https://api.github.com/repos/" + os.environ["GITHUB_REPOSITORY"]
headers = {
    "Authorization": "Bearer " + os.environ["GH_TOKEN"],
    "Accept": "application/vnd.github+json",
}


class ArtifactRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, response_headers, url):
        redirected = super().redirect_request(req, fp, code, msg, response_headers, url)
        if urllib.parse.urlsplit(req.full_url).netloc != urllib.parse.urlsplit(url).netloc:
            redirected.remove_header("Authorization")
        return redirected


opener = urllib.request.build_opener(ArtifactRedirect())


def request(path):
    return opener.open(urllib.request.Request(api + path, headers=headers), timeout=120)


with request("/actions/artifacts/" + artifact_id) as response:
    artifact = json.load(response)
assert artifact["workflow_run"]["head_sha"] == source
assert artifact["name"] == "opennox-server-windows-amd64" and not artifact["expired"]
(report / "original-artifact.json").write_text(json.dumps(artifact, indent=2))
archive = report / "retained.zip"
with request("/actions/artifacts/" + artifact_id + "/zip") as response, archive.open("xb") as output:
    shutil.copyfileobj(response, output)
with archive.open("rb") as stream:
    digest = hashlib.file_digest(stream, "sha256").hexdigest()
assert artifact["digest"] == "sha256:" + digest
assert archive.stat().st_size == artifact["size_in_bytes"]
with zipfile.ZipFile(archive) as bundle:
    manifest = json.loads(bundle.read("test-binaries.json").decode("utf-8-sig"))
    original_summary = json.loads(bundle.read("summary.json").decode("utf-8-sig"))
    binaries = {}
    for row in manifest:
        if row["name"] not in {"v1.test.exe", "server.test.exe", "legacy.test.exe"}:
            continue
        path = report / row["name"]
        with bundle.open("test-binaries/" + row["name"]) as input, path.open("xb") as output:
            shutil.copyfileobj(input, output)
        assert path.stat().st_size == row["bytes"]
        with path.open("rb") as stream:
            assert hashlib.file_digest(stream, "sha256").hexdigest() == row["sha256"]
        binaries[row["name"]] = row
assert len(binaries) == 3
(report / "original-summary.json").write_text(json.dumps(original_summary, indent=2))
environment = dict(os.environ, GOTOOLCHAIN="local", GOARCH="amd64", CGO_ENABLED="0", GOWORK="off")
environment.pop("GOEXPERIMENT", None)
(report / "replay-environment.json").write_text(json.dumps({
    "platform": platform.platform(),
    "go_version": subprocess.check_output(["go", "version"], text=True).strip(),
    "source": source,
    "artifact_zip_sha256": digest,
    "retained_binaries": binaries,
    "environment": {key: environment.get(key) for key in ["GOGC", "GODEBUG", "GOEXPERIMENT", "GOARCH"]},
    "no_engine_runtime_or_test_changes": True,
}, indent=2))


def replay(task):
    name, package_dir, pattern, count, gc, debug, label = task
    executable = report / name
    args = ["go", "tool", "test2json", "-t", "-p", "github.com/opennox/opennox/v1" +
            ("/" + package_dir if package_dir else ""), str(executable),
            "-test.v=test2json", "-test.count=" + str(count), "-test.timeout=15m", "-test.run=" + pattern]
    env = dict(environment, GOGC=gc, GODEBUG=debug)
    with (report / (label + ".jsonl")).open("xb") as out, (report / (label + ".stderr.log")).open("xb") as err:
        try:
            result = subprocess.run(args, cwd=Path("src") / package_dir, env=env,
                                    stdout=out, stderr=err, timeout=1000)
            code = result.returncode
        except subprocess.TimeoutExpired:
            code = "timeout"
    actions = collections.Counter()
    unfinished = set()
    faults = []
    with (report / (label + ".jsonl")).open(encoding="utf-8") as stream:
        for line in stream:
            event = json.loads(line)
            test = event.get("Test")
            action = event["Action"]
            if test:
                if action == "run":
                    unfinished.add(test)
                elif action in {"pass", "skip", "fail"}:
                    unfinished.discard(test)
                    actions[action] += 1
            output = event.get("Output", "")
            if any(word in output for word in ["fatal error:", "unexpected return pc", "runtime: s.allocCount="]):
                faults.append(event)
    row = dict(label=label, args=args, exit_code=code, actions=dict(actions),
               unfinished_tests=sorted(unfinished), fault_prefixes=faults, GOGC=gc, GODEBUG=debug)
    (report / (label + ".summary.json")).write_text(json.dumps(row, indent=2))
    print(json.dumps(row), flush=True)
    return row


cases = [
    ("v1.test.exe", "", "^TestAIEnemyDistanceDependency546E8EMissingInputFaultPrefix$"),
    ("server.test.exe", "server", "^TestDropEligibilityNative4EDCD0FaultOrder$"),
    ("legacy.test.exe", "legacy", "^TestSourceLessElectricPrecisionOriginal4E17B0Native$"),
]
tasks = []
for name, package_dir, pattern in cases:
    label = name.removesuffix(".test.exe")
    tasks += [
        (name, package_dir, pattern, 10, "100", "", label + "-focused-ordinary"),
        (name, package_dir, pattern, 2, "10", "gccheckmark=1", label + "-focused-gc-stress"),
        (name, package_dir, ".", 1, "100", "", label + "-full-ordinary"),
    ]
# Packages execute independently, as in the original ordinary gate.
with ThreadPoolExecutor(max_workers=3) as pool:
    results = list(pool.map(replay, tasks))
for name, row in binaries.items():
    with (report / name).open("rb") as stream:
        assert hashlib.file_digest(stream, "sha256").hexdigest() == row["sha256"]
(report / "results.json").write_text(json.dumps(results, indent=2))
(report / "scope.json").write_text(json.dumps(dict(
    source=source, exact_retained_executables=binaries, artifact_zip_sha256=digest,
    diagnostic_only=True, no_engine_or_runtime_changes=True,
    original_failed_ordinary_gate_remains_failed=True,
), indent=2))
