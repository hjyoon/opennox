"""Preserve full native macOS test streams, products and executed test binaries."""
from pathlib import Path
import collections
import hashlib
import json
import os
import platform
import re
import subprocess
import time

repo = Path.cwd().resolve()
variant = os.environ['GATE_VARIANT']
source = os.environ['SOURCE_REVISION']
assert re.fullmatch(r'[0-9a-f]{40}', source)
assert variant in ['normal', 'hd', 'server', 'server-hd', 'strict']
report = repo / 'build' / ('macos-arm64-' + variant)
report.mkdir(parents=True)
env = dict(os.environ, GOTOOLCHAIN='local')
# The native filesystem returns physical paths. Match the local gate's
# canonical temporary-directory setup before tests create private Save roots.
env['TMPDIR'] = str(Path(env.get('TMPDIR', '/private/tmp')).resolve(strict=True))
for key in ['GOROOT', 'GOTOOLDIR']:
    env.pop(key, None)
assert platform.system() == 'Darwin' and platform.machine() == 'arm64'
assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip() == source
assert not subprocess.check_output(['git', 'status', '--porcelain'], text=True)


def capture(label, args, required=True):
    start = time.monotonic()
    with (report / (label + '.stdout.log')).open('xb') as out, (report / (label + '.stderr.log')).open('xb') as err:
        result = subprocess.run(args, cwd=repo / 'src', env=env, stdout=out, stderr=err)
    row = dict(label=label, args=args, exit_code=result.returncode,
               seconds=round(time.monotonic() - start, 3))
    with (report / 'commands.jsonl').open('a') as output:
        output.write(json.dumps(row) + '\n')
    print(json.dumps(row), flush=True)
    if required:
        assert result.returncode == 0, row
    return result.returncode


capture('go-version', ['go', 'version'])
assert 'go1.26.9 darwin/arm64' in (report / 'go-version.stdout.log').read_text()
capture('compiler', ['clang', '--version'])
capture('libraries', ['pkg-config', '--modversion', 'sdl2', 'openal'])
capture('SDK', ['xcrun', '--show-sdk-path'])
(report / 'source-revision.txt').write_text(source + '\n')
tags = {'normal': '', 'hd': 'highres', 'server': 'server',
        'server-hd': 'server,highres', 'strict': 'highres'}[variant]
flags = ['-tags=' + tags] if tags else []
profile = dict(source=source, workflow_revision=os.environ['GATE_WORKFLOW_REVISION'],
               variant=variant, native_OS=platform.platform(), CPU=platform.machine(),
               tags=tags, CGO_ENABLED=env['CGO_ENABLED'], CGO_CFLAGS=env['CGO_CFLAGS'],
               SDKROOT=env['SDKROOT'], GOFLAGS=env['GOFLAGS'], GOMAXPROCS=env['GOMAXPROCS'])
profile['TMPDIR'] = env['TMPDIR']
if variant == 'strict':
    env.update(GOEXPERIMENT='cgocheck2', NOX_TEST_OPENAL='true',
               NOX_TEST_NATIVE_FX='true', ALSOFT_DRIVERS='null')
    pattern = Path(env['GATE_STRICT_PATTERN']).read_text().strip()
    assert pattern.startswith('^Test(')
    flags += ['-gcflags=all=-d=checkptr=2', '-run', pattern]
    profile.update(GOEXPERIMENT='cgocheck2', NOX_TEST_OPENAL='true',
                   NOX_TEST_NATIVE_FX='true', ALSOFT_DRIVERS='null', pattern=pattern)
(report / 'profile.json').write_text(json.dumps(profile, indent=2) + '\n')
if variant != 'strict':
    product = {'normal': 'opennox', 'hd': 'opennox-hd', 'server': 'opennox-server',
               'server-hd': 'opennox-server-hd'}[variant]
    capture('product-build', ['go', 'run', './internal/noxbuild', '-o=' + str(report),
                             '-os=darwin', '-arch=arm64', '-cc=clang', '-cxx=clang++', product])
    binary = report / product
    assert binary.is_file()
    capture('product-verify', ['go', 'run', './internal/noxbuild', '-verify', str(binary)])
    capture('product-build-info', ['go', 'version', '-m', str(binary)])
    info = (report / 'product-build-info.stdout.log').read_text()
    for token in ['go1.26.9', 'GOOS=darwin', 'GOARCH=arm64', 'CGO_ENABLED=1',
                  'vcs.modified=false', 'vcs.revision=' + source]:
        assert token in info, token
    capture('product-help', [str(binary), '-h'])
    with binary.open('rb') as stream:
        digest = hashlib.file_digest(stream, 'sha256').hexdigest()
    (report / 'product.json').write_text(json.dumps(dict(
        name=product, bytes=binary.stat().st_size, sha256=digest, source=source), indent=2) + '\n')
packages = ['.', './server', './legacy', './legacy/client/audio/ail']
capture('expected-packages', ['go', 'list', *(['-tags=' + tags] if tags else []), *packages])
expected = set((report / 'expected-packages.stdout.log').read_text().splitlines())
assert len(expected) == 4
executables = report / 'test-binaries'
executables.mkdir()
status = capture('test', ['go', 'test', *flags, '-json', '-count=1', '-timeout=20m',
                          '-o=' + str(executables), *packages], required=False)
tests, terminal, failures, build_failures = collections.Counter(), {}, [], []
digest, size = hashlib.sha256(), 0
with (report / 'test.stdout.log').open('rb') as stream:
    for raw in stream:
        digest.update(raw)
        size += len(raw)
        event = json.loads(raw)
        action = event['Action']
        if action in ['pass', 'skip', 'fail']:
            if event.get('Test'):
                tests[action] += 1
            else:
                assert event['Package'] not in terminal
                terminal[event['Package']] = action
            if action == 'fail':
                failures.append(event)
        if action == 'build-fail':
            build_failures.append(event)
summary = dict(source=source, test_exit_code=status, test_actions=dict(tests),
               package_actions=terminal, expected_packages=sorted(expected),
               missing_packages=sorted(expected - terminal.keys()),
               unexpected_packages=sorted(terminal.keys() - expected),
               failures=failures, build_failures=build_failures,
               raw_JSON_bytes=size, raw_JSON_sha256=digest.hexdigest())
(report / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
binaries = []
for index, binary in enumerate(sorted(executables.iterdir())):
    if not binary.is_file():
        continue
    with binary.open('rb') as stream:
        digest = hashlib.file_digest(stream, 'sha256').hexdigest()
    code = capture('test-binary-symbols-' + str(index),
                   ['go', 'tool', 'nm', '-size', str(binary)], required=False)
    symbols = (report / ('test-binary-symbols-' + str(index) + '.stdout.log')).read_text()
    binaries.append(dict(name=binary.name, bytes=binary.stat().st_size,
                         sha256=digest, nm_exit_code=code,
                         GC_symbols='runtime.gcdata' in symbols and 'runtime.gcbss' in symbols))
(report / 'test-binaries.json').write_text(json.dumps(binaries, indent=2) + '\n')
warning = re.compile(r'\bwarning(?:\s+[A-Z0-9_-]+)?:|DeprecationWarning|##\[warning\]', re.I)
warnings = []
for path in report.glob('*.log'):
    with path.open(errors='replace') as stream:
        for line in stream:
            text = line
            if path.name == 'test.stdout.log':
                text = json.loads(line).get('Output', '')
            if warning.search(text):
                warnings.append(dict(file=path.name, text=text.rstrip()))
(report / 'warnings.json').write_text(json.dumps(warnings, indent=2) + '\n')
print(json.dumps(dict(tests=tests, packages=terminal, warnings=len(warnings))), flush=True)
assert status == 0 and not failures and not build_failures, summary
assert set(terminal) == expected
for package in ['github.com/opennox/opennox/v1' + suffix for suffix in ['', '/server', '/legacy']]:
    assert terminal[package] == 'pass', (package, terminal)
assert binaries and all(row['nm_exit_code'] == 0 and row['GC_symbols'] for row in binaries)
assert not warnings, warnings[:10]
(report / 'gate-passed.json').write_text(json.dumps(dict(
    source=source, variant=variant, passed=True, native_CPU='arm64',
    all_emitted_compiler_and_linker_warnings=0, all_test_binaries_preserved=True,
    no_suppression_or_expectation_changes=True, overall_porting_100_percent=False), indent=2) + '\n')
