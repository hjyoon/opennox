"""Diagnose existing 386 failures using exact retained executables, without rebuilding."""
from pathlib import Path
import collections
import hashlib
import json
import os
import shutil
import subprocess
import urllib.parse
import urllib.request
import zipfile

repo = Path.cwd().resolve()
report = Path(os.environ['REPORT_DIR']).resolve() / 'faults386'
report.mkdir(parents=True, exist_ok=False)
source = os.environ['SOURCE_REVISION']
artifact_id = os.environ['RETAINED_ARTIFACT_ID']
assert artifact_id.isdecimal()
assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip() == source
api = 'https://api.github.com/repos/' + os.environ['GITHUB_REPOSITORY']
headers = {'Authorization': 'Bearer ' + os.environ['GH_TOKEN'], 'Accept': 'application/vnd.github+json'}


class ArtifactRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, response_headers, url):
        redirected = super().redirect_request(req, fp, code, msg, response_headers, url)
        if urllib.parse.urlsplit(req.full_url).netloc != urllib.parse.urlsplit(url).netloc:
            redirected.remove_header('Authorization')
        return redirected


opener = urllib.request.build_opener(ArtifactRedirect())


def request(path):
    return opener.open(urllib.request.Request(api + path, headers=headers), timeout=120)


def sha(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


with request('/actions/artifacts/' + artifact_id) as response:
    artifact = json.load(response)
assert artifact['workflow_run']['head_sha'] == source and not artifact['expired']
assert artifact['name'] == 'opennox-server-windows-386'
(report / 'original-artifact.json').write_text(json.dumps(artifact, indent=2))
archive = report / 'retained.zip'
with request('/actions/artifacts/' + artifact_id + '/zip') as stream, archive.open('xb') as output:
    shutil.copyfileobj(stream, output)
assert archive.stat().st_size == artifact['size_in_bytes']
assert artifact['digest'] == 'sha256:' + sha(archive)
binaries = {}
with zipfile.ZipFile(archive) as bundle:
    assert bundle.read('source-revision.txt').decode('utf-8-sig').strip() == source
    manifest = json.loads(bundle.read('test-binaries.json').decode('utf-8-sig'))
    (report / 'original-summary.json').write_bytes(bundle.read('summary.json'))
    for package in ['server', 'legacy']:
        record = next(row for row in manifest if row['name'] == package + '.test.exe')
        executable = report / record['name']
        with bundle.open('test-binaries/' + record['name']) as stream, executable.open('xb') as output:
            shutil.copyfileobj(stream, output)
        assert executable.stat().st_size == record['bytes'] and sha(executable) == record['sha256']
        binaries[package] = dict(record=record, path=str(executable))

environment = dict(os.environ, GOTOOLCHAIN='local', GOARCH='386', GOGC='100', GODEBUG='')
environment.pop('GOEXPERIMENT', None)
results = []


def replay(label, package, pattern, count=1, gc='100'):
    executable = binaries[package]['path']
    args = [executable, '-test.v=test2json', '-test.count=' + str(count),
            '-test.timeout=15m', '-test.run=' + pattern]
    env = dict(environment, GOGC=gc)
    raw, errors = report / (label + '.stdout.log'), report / (label + '.stderr.log')
    with raw.open('xb') as output, errors.open('xb') as error:
        try:
            process = subprocess.run(args, cwd=repo / 'src' / package, env=env,
                                     stdout=output, stderr=error, timeout=1000)
            code = process.returncode
        except subprocess.TimeoutExpired:
            code = None
    events = report / (label + '.jsonl')
    with raw.open('rb') as stream, events.open('xb') as output, (report / (label + '.converter.stderr.log')).open('xb') as error:
        conversion = subprocess.run(['go', 'tool', 'test2json', '-t', '-p',
                                     'github.com/opennox/opennox/v1/' + package],
                                    stdin=stream, stdout=output, stderr=error, env=env, timeout=180)
    actions, unfinished = collections.Counter(), set()
    with events.open(encoding='utf-8') as stream:
        for line in stream:
            event = json.loads(line)
            name, action = event.get('Test'), event['Action']
            if name and action == 'run':
                unfinished.add(name)
            elif name and action in ['pass', 'skip', 'fail']:
                unfinished.discard(name)
                actions[action] += 1
    row = dict(label=label, package=package, source=source, args=args, GOGC=gc,
               binary_exit_code=code, binary_exit_hex=None if code is None else hex(code & 0xffffffff),
               converter_exit_code=conversion.returncode, actions=dict(actions),
               unfinished_tests=sorted(unfinished), stdout_sha256=sha(raw), stderr_sha256=sha(errors),
               executable_sha256=sha(executable), original_binary=binaries[package]['record'])
    assert row['executable_sha256'] == row['original_binary']['sha256']
    results.append(row)
    (report / 'results.json').write_text(json.dumps(results, indent=2))
    print(json.dumps(row), flush=True)


for index in range(10):
    replay('server-required-nil-cold-' + str(index), 'server', '^TestCurePoisonCast52CDB0NilTargetAndRequiredArgument$')
replay('server-cure-poison-controls', 'server', '^TestCurePoisonCast52CDB0', count=10)
replay('legacy-troll-original', 'legacy', '^TestTrollDeathOriginal54A270$', count=3)
replay('legacy-troll-GC-stress', 'legacy', '^TestTrollDeathOriginal54A270$', count=3, gc='10')
replay('server-full-original', 'server', '.')
replay('legacy-full-original', 'legacy', '.')
(report / 'scope.json').write_text(json.dumps(dict(
    diagnostic_only=True, no_rebuild_or_engine_or_runtime_or_test_changes=True,
    exact_retained_executables=binaries, artifact_zip_sha256=sha(archive),
    direct_binary_exit_separate_from_converter=True,
    original_failed_native_gate_still_failed=True), indent=2))
