#!/usr/bin/env python3
"""Prepare one unique preserved local test invocation; never compile in Go tests."""
import argparse
import datetime
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import uuid

sys.dont_write_bytecode = True

# cmd/go Go1.27.3 work.AddBuildFlags/AddCoverFlags and test.testFlags.
# These sets classify metadata flags only. Actual test arguments are unchanged.
BUILD_BOOL = set('a asan buildvcs cover linkshared modcacherw msan race trimpath work n x'.split())
BUILD_VALUE = set(('p asmflags compiler buildmode gcflags gccgoflags mod modfile overlay installsuffix '
                   'ldflags pgo pkgdir tags toolexec covermode coverpkg '
                   'debug-actiongraph debug-runtime-trace debug-trace').split())
TEST_BOOL = set('artifacts benchmem failfast fullpath short v c json'.split())
TEST_VALUE = set(('bench benchtime blockprofile blockprofilerate count cpu cpuprofile fuzz '
                  'fuzztime fuzzminimizetime list memprofile memprofilerate mutexprofile '
                  'mutexprofilefraction outputdir parallel run skip timeout trace shuffle '
                  'coverprofile o exec vet').split())
PASS_FLAGS = TEST_BOOL - {'c', 'json'} | TEST_VALUE - {'o', 'exec', 'vet'}


def fact(path):
    path = Path(path).resolve()
    with path.open('rb') as f:
        return dict(path=str(path), bytes=path.stat().st_size,
                    sha256=hashlib.file_digest(f, 'sha256').hexdigest())


def save(path, value):
    with path.open('x') as f:
        json.dump(value, f, indent=2)
        f.write('\n')


def metadata_arguments(arguments):
    """Mirror Go1.27.3 testFlags package-list/unknown-flag boundaries.

    Test selectors and binary arguments never become go-list build flags.
    Build flags from GOFLAGS remain inherited by both commands unchanged.
    """
    build, packages, in_packages, after_unknown = [], None, False, False
    index = 0
    while index < len(arguments):
        argument = arguments[index]
        was_after_unknown, after_unknown = after_unknown, False
        if argument == '--':
            break
        if not argument.startswith('-') or argument == '-':
            if not in_packages and packages is not None:
                if was_after_unknown:
                    index += 1
                    continue
                break
            in_packages = True
            if packages is None:
                packages = []
            packages.append(argument)
            index += 1
            continue
        in_packages = False
        name, separator, value = argument.lstrip('-').partition('=')
        if name == 'C':
            raise ValueError('-C belongs before the Go subcommand; this entry uses REPO/src')
        if name in ('args', ''):
            break
        short = name[5:] if name.startswith('test.') and name[5:] in PASS_FLAGS else name
        known = short in BUILD_BOOL | BUILD_VALUE | TEST_BOOL | TEST_VALUE
        if not known:
            if packages is None:
                packages = []
            after_unknown = not separator
            index += 1
            continue
        consumed = [argument]
        if not separator and short in BUILD_VALUE | TEST_VALUE:
            index += 1
            if index == len(arguments):
                raise ValueError('Missing flag value: '+argument)
            consumed.append(arguments[index])
        if name in BUILD_BOOL | BUILD_VALUE:
            build.extend(consumed)
        index += 1
    return build, packages or ['.']


def query(command, cwd, environment, report, name):
    stdout, stderr = report/(name+'.stdout.json'), report/(name+'.stderr.log')
    with stdout.open('xb') as out, stderr.open('xb') as err:
        try:
            completed = subprocess.run(command, cwd=cwd, env=environment,
                                       stdout=out, stderr=err, timeout=30)
        except subprocess.TimeoutExpired:
            save(report/(name+'-exit.json'), dict(command=command, timed_out=True,
                 stdout=fact(stdout), stderr=fact(stderr)))
            raise
    save(report/(name+'-exit.json'), dict(command=command, returncode=completed.returncode,
         timed_out=False, stdout=fact(stdout), stderr=fact(stderr)))
    if completed.returncode != 0:
        raise RuntimeError('Metadata failed; complete original arguments retained: '+str(stderr))
    return stdout.read_text()


def json_values(text):
    decoder = json.JSONDecoder()
    at = 0
    while at < len(text):
        if text[at].isspace():
            at += 1
            continue
        value, at = decoder.raw_decode(text, at)
        yield value


def load_script(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def preserve_package_mains(report, expected):
    actual, errors, work_paths = {}, [], set()
    for line in (report/'test.stderr.log').read_text(errors='replace').splitlines():
        if line.startswith('WORK='):
            work_paths.add(line[5:])
    with (report/'test.jsonl').open('rb') as events:
        for index, line in enumerate(events):
            try:
                event = json.loads(line)
            except (ValueError, UnicodeError) as error:
                errors.append(dict(problem='Incomplete/invalid actual Go JSON line retained', line=index+1, error=repr(error)))
                continue
            if event.get('Action') == 'build-output':
                for output in event.get('Output', '').splitlines():
                    if output.startswith('WORK='):
                        work_paths.add(output[5:])
    if len(work_paths) != 1:
        errors.append(dict(problem='Expected exactly one actual retained WORK', observed=sorted(work_paths)))
    save(report/'test-work-directory.json', dict(paths=sorted(work_paths),
         sources=['test.stderr.log', 'test.jsonl build-output'], cleanup_performed=False))
    for context in sorted((report/'package-mains').glob('*/package-context-before-exec.json')):
        admitted = json.loads(context.read_bytes())
        package = admitted['package']
        folder = context.parent
        rows = list((folder/'test-binary-records').glob('*.json'))
        if package not in expected or package in actual or len(rows) != 1:
            errors.append(dict(package=package, problem='Unexpected/duplicate/missing actual main record'))
            continue
        row = json.loads(rows[0].read_bytes())
        binary = folder/'test-binaries'/row['name']
        observed = fact(binary)
        if (observed['bytes'], observed['sha256']) != (row['bytes'], row['sha256']):
            errors.append(dict(package=package, problem='Actual retained main bytes changed'))
        if not row['preservation_before_execution'] or not row.get('execution_started') or row.get('phase') not in ('exited', 'timed-out', 'supervisor-error'):
            errors.append(dict(package=package, problem='Actual main lacks PREEXEC or terminal lifecycle'))
        if row.get('phase') != 'exited' or row.get('forced_termination') is not False or row.get('native_exit_code') != 0 or row.get('wrapper_exit_code') != 0:
            errors.append(dict(package=package, problem='Actual main did not exit naturally with native/wrapper zero'))
        actual[package] = dict(record=row, binary=observed, context=admitted)
    if set(actual) != set(expected):
        errors.append(dict(problem='Pre-admitted package main set differs',
                           missing=sorted(set(expected)-set(actual)), extra=sorted(set(actual)-set(expected))))
    result = dict(expected_packages=expected, actual=actual, problems=errors,
                  all_available_mains_preserved_before_execution=not errors,
                  no_report_or_WORK_cleanup=True)
    save(report/'package-main-preservation-status.json', result)
    return errors


def run_with_isolated_main_names(repo, go, report, arguments, environment, admission):
    helper = load_script(repo/'scripts/prepare-default-damage-initial-worker.py', 'initial_worker_setup')
    setup = helper.prepare_for_suite(repo, go, report, environment)
    environment.update(setup['environment'])
    central = load_script(repo/'scripts/run-bounded-go-test-suite.py', 'unchanged_bounded_quoting')
    exec_argument = central.go_exec_argument([
        sys.executable, str(repo/'scripts/run-prepared-package-test.py'),
        '--report', str(report), '--admission', str(report/'prepared-entry-package-admission.json'),
    ])
    # Its test-result cache key includes the complete exec-command arguments,
    # including this unique report path; no cached main can replace this run.
    # cmd/go/internal/test/test.go expressly permits os.DevNull with multiple
    # packages (including duplicate binary basenames). -work keeps originals;
    # the package-specific unchanged supervisor retains each BEFORE execution.
    command = [str(go), 'test', '-work', '-o='+os.devnull, '-json', '-exec='+exec_argument, *arguments]
    save(report/'test-command.json', dict(command=command, cwd=str(repo/'src'),
         initial_worker_setup=setup, admission=admission, expected_package_executables=admission['expected_mains'],
         scope_remains_one_original_Go_invocation=True, binary_basename_collisions_isolated_by_package=True))
    supervisor = load_script(repo/'scripts/run-preserved-go-test.py', 'unchanged_Go_driver_supervisor')
    os.environ.clear(); os.environ.update(environment)
    os.chdir(repo/'src')
    with (report/'test.jsonl').open('xb') as out, (report/'test.stderr.log').open('xb') as err:
        code = supervisor.run(report, command, 2400, 30, retain_binary=False, stdout=out, stderr=err)
    with (report/'test.exit').open('x') as f:
        f.write(str(code)+'\n')
    errors = preserve_package_mains(report, admission['expected_test_package_identities'])
    if code == 0 and errors:
        raise RuntimeError('Preservation incomplete; all report/WORK/worker bytes retained')
    return code


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--go', type=Path)
    parser.add_argument('--label', default='local-tests')
    parser.add_argument('--expected-mains', type=int)
    parser.add_argument('--isolate-main-names', action='store_true',
                        help='Retain duplicate package basenames separately; one unchanged Go invocation')
    parser.add_argument('arguments', nargs=argparse.REMAINDER)
    args = parser.parse_args()
    arguments = args.arguments[1:] if args.arguments[:1] == ['--'] else args.arguments
    assert arguments, 'Entrypoint must pass its original explicit/default package scope'
    assert re.fullmatch('[a-zA-Z0-9_-]+', args.label)
    repo = Path(__file__).resolve().parents[1]
    environment = dict(os.environ)
    # This affects Python cache litter only, not Go experiments or test flags.
    environment['PYTHONDONTWRITEBYTECODE'] = '1'
    root = Path(environment.get('OPENNOX_LOCAL_TEST_ARTIFACTS', repo/'build/local-preserved-tests')).resolve()
    root.mkdir(parents=True, exist_ok=True)
    run = root/(args.label+'-'+datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ')+'-'+uuid.uuid4().hex)
    run.mkdir(mode=0o700)
    report, output = run/'report', run/'actual-executables'
    report.mkdir(); output.mkdir()
    save(report/'prepared-entry-original-arguments.json', dict(arguments=arguments, caller_cwd=str(Path.cwd()),
         source_root=str(repo), entry=fact(__file__), explicit_expected_mains=args.expected_mains,
         unique_run=str(run), BEFORE_caller_owned=environment.get('OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE')))
    if args.go is None:
        # Make previously used go.sh, which chooses the pinned compiler and
        # clears GOEXPERIMENT. Preserve that behavior without changing go.sh.
        selected = query([str(repo/'scripts/go.sh'), 'env', 'GOROOT'], repo,
                         environment, report, 'pinned-toolchain-root').strip()
        go = Path(selected)/'bin/go'
        environment.pop('GOROOT', None)
        environment['GOEXPERIMENT'] = ''
        environment['GOTOOLCHAIN'] = 'local'
        # These are the same empty-value defaults go.sh applies in its child.
        # GOROOT-query subprocess changes do not propagate back to this parent.
        if not environment.get('CGO_CFLAGS_ALLOW'):
            environment['CGO_CFLAGS_ALLOW'] = '-f.*'
        openal_pkgconfig = Path('/opt/homebrew/opt/openal-soft/lib/pkgconfig')
        if not environment.get('PKG_CONFIG_PATH') and openal_pkgconfig.is_dir():
            environment['PKG_CONFIG_PATH'] = str(openal_pkgconfig)
    else:
        # Strict shell already verified/configured this exact executable.
        # In particular, do not route it back through go.sh's experiment reset.
        go = args.go.resolve(strict=True)
    facts = json.loads(query([str(go), 'env', '-json', 'GOVERSION', 'GOROOT', 'GOOS', 'GOARCH', 'GOEXPERIMENT', 'GOFLAGS'],
                             repo/'src', environment, report, 'pinned-Go-entry-facts'))
    assert facts['GOVERSION'] == (repo/'toolchain/go-version.txt').read_text().strip() == 'go1.27.3'
    build_flags, packages = metadata_arguments(arguments)
    package_rows = list(json_values(query([str(go), 'list', '-json', *build_flags, *packages],
                                         repo/'src', environment, report, 'selected-package-metadata')))
    assert package_rows and all(not p.get('Error') and not p.get('Incomplete') for p in package_rows)
    imports = [p['ImportPath'] for p in package_rows]
    assert len(imports) == len(set(imports))
    expected = sorted(p['ImportPath'] for p in package_rows if p.get('TestGoFiles') or p.get('XTestGoFiles'))
    assert expected, 'No package executable selected; metadata retained, no fake success'
    if args.expected_mains is not None:
        assert len(expected) == args.expected_mains, (len(expected), args.expected_mains)
    central = repo/'scripts/run-bounded-go-test-suite.py'
    command = [sys.executable, str(central), '--go', str(go), '--report', str(report),
               '--test-output', str(output), '--expected-mains', str(len(expected)), '--', *arguments]
    admitted_packages = {str(Path(p['Dir']).resolve()): p['ImportPath'] for p in package_rows if p['ImportPath'] in expected}
    assert len(admitted_packages) == len(expected), 'Distinct package directories required for pre-exec admission'
    admission = dict(target_facts=facts,
         original_test_arguments=arguments, metadata_build_flags=build_flags, metadata_packages=packages,
         expected_test_package_identities=expected, expected_mains=len(expected),
         expected_packages_by_directory=admitted_packages,
         no_package_or_selector_or_test_arguments_rewritten=True,
         central=fact(central), standard_bounded_command=command if not args.isolate_main_names else None,
         isolate_main_names=args.isolate_main_names, report_owned_raw_outputs_preserved=True,
         cleanup_performed=False, BEFORE_caller_owned=environment.get('OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE'))
    save(report/'prepared-entry-package-admission.json', admission)
    print(json.dumps(dict(prepared_local_test_report=str(report), expected_mains=len(expected))), flush=True)
    if args.isolate_main_names:
        return run_with_isolated_main_names(repo, str(go), report, arguments, environment, admission)
    os.chdir(repo/'src')
    os.execve(sys.executable, command, environment)


if __name__ == '__main__':
    raise SystemExit(main())
