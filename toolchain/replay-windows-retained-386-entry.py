"""Compare exact Windows/386 legacy executables without rebuilding or patching."""
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

repo = Path.cwd().resolve()
report = Path(os.environ["REPORT_DIR"]).resolve() / "entry386"
report.mkdir(parents=True, exist_ok=False)
source = os.environ["SOURCE_REVISION"]
assert subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip() == source
api = "https://api.github.com/repos/" + os.environ["GITHUB_REPOSITORY"]
headers = {"Authorization": "Bearer " + os.environ["GH_TOKEN"],
           "Accept": "application/vnd.github+json"}


class ArtifactRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, response_headers, url):
        redirected = super().redirect_request(req, fp, code, msg, response_headers, url)
        if urllib.parse.urlsplit(req.full_url).netloc != urllib.parse.urlsplit(url).netloc:
            redirected.remove_header("Authorization")
        return redirected


opener = urllib.request.build_opener(ArtifactRedirect())


def request(path):
    return opener.open(urllib.request.Request(api + path, headers=headers), timeout=120)


def sha(path):
    with Path(path).open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def retain(label, artifact_id, revision, checkout):
    assert artifact_id.isdecimal() and len(revision) == 40
    folder = report / label
    folder.mkdir()
    with request("/actions/artifacts/" + artifact_id) as response:
        artifact = json.load(response)
    assert artifact["workflow_run"]["head_sha"] == revision
    assert artifact["name"] == "opennox-server-windows-386" and not artifact["expired"]
    (folder / "artifact.json").write_text(json.dumps(artifact, indent=2))
    archive = folder / "retained.zip"
    with request("/actions/artifacts/" + artifact_id + "/zip") as response, archive.open("xb") as out:
        shutil.copyfileobj(response, out)
    assert archive.stat().st_size == artifact["size_in_bytes"]
    assert artifact["digest"] == "sha256:" + sha(archive)
    with zipfile.ZipFile(archive) as bundle:
        assert bundle.read("source-revision.txt").decode("utf-8-sig").strip() == revision
        manifest = json.loads(bundle.read("test-binaries.json").decode("utf-8-sig"))
        binary = next(row for row in manifest if row["name"] == "legacy.test.exe")
        executable = folder / "legacy.test.exe"
        with bundle.open("test-binaries/legacy.test.exe") as stream, executable.open("xb") as out:
            shutil.copyfileobj(stream, out)
        assert executable.stat().st_size == binary["bytes"] and sha(executable) == binary["sha256"]
        (folder / "original-summary.json").write_bytes(bundle.read("summary.json"))
    if checkout != repo:
        subprocess.run(["git", "-c", "core.longpaths=true", "worktree", "add", "--detach", str(checkout), revision], check=True)
    assert subprocess.check_output(["git", "-C", str(checkout), "rev-parse", "HEAD"], text=True).strip() == revision
    metadata = dict(label=label, source=revision, artifact_id=artifact_id, archive_sha256=sha(archive),
                    binary=binary, checkout=str(checkout), no_rebuild_or_runtime_patch=True)
    (folder / "binding.json").write_text(json.dumps(metadata, indent=2))
    return folder, executable, metadata


environment = dict(os.environ, GOTOOLCHAIN="local", GOARCH="386", CGO_ENABLED="1",
                   GOWORK=str(repo / "go.work"), GOGC="100", GODEBUG="")
environment.pop("GOEXPERIMENT", None)
compiler = Path(environment["CC"])
assert compiler.name == "i686-w64-mingw32-clang.exe" and compiler.is_file()
assert environment["CGO_CFLAGS"] == "-O2 -g"
(report / "environment.json").write_text(json.dumps(dict(
    platform=platform.platform(), go_version=subprocess.check_output(["go", "version"], text=True).strip(),
    compiler_version=subprocess.check_output([str(compiler), "--version"], text=True),
    environment={key: environment.get(key) for key in ["GOARCH", "GOGC", "GODEBUG", "GOEXPERIMENT", "CC", "CXX", "CGO_CFLAGS", "CGO_ENABLED", "GOWORK"]},
    original_independent_C_probes_keep_native_386_compiler=True,
    no_engine_test_or_runtime_changes=True), indent=2))
current = retain("current", os.environ["RETAINED_ARTIFACT_ID"], source, repo)
baseline_id = os.environ.get("BASELINE_RETAINED_ARTIFACT_ID", "")
baseline_source = os.environ.get("BASELINE_SOURCE_REVISION", "")
assert bool(baseline_id) == bool(baseline_source)
groups = [current]
if baseline_id:
    baseline_checkout = Path(os.environ["RUNNER_TEMP"]) / "opennox-entry386-baseline"
    groups.append(retain("baseline", baseline_id, baseline_source, baseline_checkout))

results = []


def replay(group, label, pattern, count):
    folder, executable, metadata = group
    args = [str(executable), "-test.v=test2json", "-test.count=" + str(count),
            "-test.timeout=15m", "-test.run=" + pattern]
    raw, errors = folder / (label + ".stdout.log"), folder / (label + ".stderr.log")
    replay_env = dict(environment, GOWORK=str(Path(metadata["checkout"]) / "go.work"))
    with raw.open("xb") as out, errors.open("xb") as err:
        try:
            run = subprocess.run(args, cwd=Path(metadata["checkout"]) / "src/legacy",
                                 env=replay_env, stdout=out, stderr=err, timeout=1000)
            code = run.returncode
        except subprocess.TimeoutExpired:
            code = None
    events = folder / (label + ".jsonl")
    with raw.open("rb") as stream, events.open("xb") as out, (folder / (label + ".converter.stderr.log")).open("xb") as err:
        conversion = subprocess.run(["go", "tool", "test2json", "-t", "-p",
                                     "github.com/opennox/opennox/v1/legacy"],
                                    stdin=stream, stdout=out, stderr=err, env=environment, timeout=180)
    actions, unfinished = collections.Counter(), set()
    with events.open(encoding="utf-8") as stream:
        for line in stream:
            event = json.loads(line)
            name, action = event.get("Test"), event["Action"]
            if not name:
                continue
            if action == "run":
                unfinished.add(name)
            elif action in ["pass", "skip", "fail"]:
                unfinished.discard(name)
                actions[action] += 1
    row = dict(group=metadata["label"], source=metadata["source"], label=label, args=args,
               binary_exit_code=code, binary_exit_hex=None if code is None else hex(code & 0xffffffff),
               converter_exit_code=conversion.returncode, actions=dict(actions),
               unfinished_tests=sorted(unfinished), stdout_sha256=sha(raw), stderr_sha256=sha(errors),
               executable_sha256=sha(executable), original_binary_sha256=metadata["binary"]["sha256"])
    assert row["executable_sha256"] == row["original_binary_sha256"]
    results.append(row)
    (folder / (label + ".summary.json")).write_text(json.dumps(row, indent=2))
    (report / "results.json").write_text(json.dumps(results, indent=2))
    print(json.dumps(row), flush=True)


for group in groups:
    for index in range(10):
        replay(group, "cold-single-" + str(index),
               "^TestToggleTriggerSwitchOriginalFaultsNativeC$/^trigger$/^nil-update-cold$", 1)
    replay(group, "all-original-faults-repeated", "^TestToggleTriggerSwitchOriginalFaultsNativeC$", 20)
    replay(group, "valid-and-fault-original-controls", "^TestToggleTriggerSwitchOriginal(NativeC|FaultsNativeC)$", 1)
    replay(group, "full-legacy-ordinary", ".", 1)

(report / "scope.json").write_text(json.dumps(dict(
    diagnostic_only=True, no_engine_runtime_or_test_changes=True,
    exact_retained_386_executables=[group[2] for group in groups],
    raw_direct_binary_status_preserved_separately_from_test2json_converter=True,
    cold_repeated_valid_and_full_ordinary_replays=True,
    original_failed_native_gate_still_failed=True), indent=2))
