"""Focused comparison only: unrelated test registration is excluded explicitly."""
from pathlib import Path
import collections
import datetime
import hashlib
import json
import os
import platform
import re
import subprocess
import sys
import tempfile

repo = Path.cwd().resolve()
source = os.environ['SOURCE_REVISION']
payload = os.environ['PAYLOAD_REVISION']
profile = os.environ['PROBE_PROFILE']
assert profile in ['before', 'normal', 'strict']
assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip() == source
assert not subprocess.check_output(['git', 'status', '--porcelain'])
assert platform.machine() == 'arm64'
report = repo/'build'/('macos-interesting-'+profile)
report.mkdir(parents=True)
env = dict(os.environ)
env.pop('GOROOT', None)
env.pop('GOTOOLDIR', None)
env['TMPDIR'] = str(Path(env.get('TMPDIR', tempfile.gettempdir())).resolve())
env['GOEXPERIMENT'] = 'cgocheck2' if profile == 'strict' else ''
if profile == 'before':
    env['OPENNOX_INTERESTING_C_BEFORE'] = '1'
prefix = 'toolchain/native-baselines/interesting-sound-5466f0/'


def historical(name):
    return subprocess.check_output(['git', 'show', payload+':'+prefix+name])


seal = json.loads(historical('seal.json'))
(report/'frozen-seal.json').write_text(json.dumps(seal, indent=2)+'\n')
files = []
for row in seal['frozen_files']:
    raw = historical(row['payload'])
    assert len(raw) == row['bytes'] and hashlib.sha256(raw).hexdigest() == row['sha256']
    target = repo/row['target']
    assert target.is_relative_to(repo/'src')
    if target.exists():
        assert target.read_bytes() == raw
    else:
        with target.open('xb') as output:
            output.write(raw)
    files.append(row)
(report/'payload-files.json').write_text(json.dumps(files, indent=2)+'\n')
# All engine files remain the exact requested source. These three frozen
# comparison files are the only additions in the historical before checkout.
tracked = subprocess.check_output(['git', 'diff', '--name-only', '--', 'src'], text=True)
assert not tracked, tracked
empty = report/'unselected-legacy-tests.go'
empty.write_text('package legacy\n')
selected = repo/'src/legacy/monster_interesting_original_5466f0_test.go'
replacements = {str(path): str(empty) for path in (repo/'src/legacy').glob('*_test.go') if path != selected}
overlay = report/'focused-overlay.json'
overlay.write_text(json.dumps({'Replace': replacements}, indent=2)+'\n')
(report/'scope.json').write_text(json.dumps(dict(
    utc=datetime.datetime.now(datetime.timezone.utc).isoformat(), source=source,
    payload_revision=payload, profile=profile, frozen_files=files,
    no_existing_engine_or_assertion_edits=True,
    unrelated_test_registration_excluded=len(replacements), full_suite_credit=False,
    GOEXPERIMENT=env['GOEXPERIMENT'], TMPDIR=env['TMPDIR'],
    platform=platform.platform(),
), indent=2)+'\n')
pattern = '^TestMonsterInterestingNative(CBefore|Layout)5466F0$' if profile == 'before' else '^TestMonsterInteresting(Original|NativeLayout)5466F0$'
binary = report/'legacy.test'
command = ['go', 'test', '-work', '-o='+str(binary), '-overlay='+str(overlay),
           '-json', '-count=1', '-timeout=15m', '-run='+pattern]
if profile == 'strict':
    command += ['-tags=highres', '-gcflags=all=-d=checkptr=2']
command += ['./legacy']
(report/'command.json').write_text(json.dumps(command, indent=2)+'\n')
with (report/'test.jsonl').open('xb') as out, (report/'test.stderr.log').open('xb') as err:
    result = subprocess.run(command, cwd=repo/'src', env=env, stdout=out, stderr=err)
(report/'test.exit').write_text(str(result.returncode)+'\n')
tests = collections.Counter()
packages = {}
failures = []
build_failures = []
unfinished = set()
for line in (report/'test.jsonl').read_text().splitlines():
    event = json.loads(line)
    action, test = event['Action'], event.get('Test')
    if action == 'build-fail':
        build_failures.append(event)
    if test:
        if action == 'run':
            unfinished.add(test)
        elif action in ['pass', 'fail', 'skip']:
            unfinished.discard(test)
            tests[action] += 1
    elif action in ['pass', 'fail', 'skip']:
        packages[event['Package']] = action
    if action == 'fail':
        failures.append(event)
warning = re.compile(r'\bwarning(?:\s+[A-Z0-9_-]+)?:|DeprecationWarning|##\[warning\]', re.I)
warnings = []
for name in ['test.jsonl', 'test.stderr.log']:
    for line in (report/name).read_text(errors='replace').splitlines():
        if warning.search(line):
            warnings.append(dict(file=name, line=line))
summary = dict(source=source, profile=profile, test_exit_code=result.returncode,
               test_actions=dict(tests), package_actions=packages, failures=failures,
               build_failures=build_failures, warnings=warnings,
               unfinished_tests=sorted(unfinished), full_suite_credit=False)
(report/'summary.json').write_text(json.dumps(summary, indent=2)+'\n')
if binary.exists():
    with binary.open('rb') as stream:
        digest = hashlib.file_digest(stream, 'sha256').hexdigest()
    (report/'exact-test-executable.json').write_text(json.dumps(dict(
        path=binary.name, bytes=binary.stat().st_size, sha256=digest), indent=2)+'\n')
print(json.dumps(summary), flush=True)
assert result.returncode == 0 and not failures and not build_failures and not warnings and not unfinished
assert tests['pass'] == (2 if profile == 'before' else 5513), summary
assert tests['skip'] == 0 and packages == {'github.com/opennox/opennox/v1/legacy': 'pass'}, summary
