#!/usr/bin/env python3
"""Prebuild/archive the isolated Darwin/ARM64 initial-memory reader before Go.

Import does not query Go or compile. No worker, test, game, original code or
network is executed. Returns ONLY4 worker environment variables; caller owns
OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE and every original suite flag/package/selector.
"""
from pathlib import Path
import argparse
import hashlib
import json
import os
import platform
import re
import shlex
import shutil
import subprocess
import sys
import traceback
import zipfile

sys.dont_write_bytecode = True
SOURCES = [
    'src/legacy/default_damage_parent_initial_controls_4e0b30.go',
    'src/legacy/default_damage_parent_initial_arena_4e0b30.go',
    'src/legacy/default_damage_parent_full_initial_4e0b30.h',
    'src/legacy/default_damage_parent_full_initial_fixture_4e0b30_test.go',
    'src/legacy/default_damage_parent_full_initial_run_4e0b30_test.go',
    'src/legacy/testdata/default_damage_parent_full_initial_4e0b30.json.gz',
    'src/legacy/testdata/generate_default_damage_parent_full_initial_4e0b30.py',
    'src/legacy/testdata/default_damage_parent_initial_worker_4e0b30.c',
    'src/legacy/testdata/default_damage_parent_initial_worker_preexec_4e0b30.py',
]
FLAGS = ['-arch', 'arm64', '-std=c11', '-O2', '-g0', '-Wall', '-Wextra', '-Werror',
         '-fno-builtin', '-fno-stack-protector', '-fshort-wchar', '-fsigned-char',
         '-fno-strict-aliasing', '-fno-strict-overflow', '-DNOX_CGO']


def record(path):
    path = Path(path).resolve()
    with path.open('rb') as f:
        return dict(path=str(path), bytes=path.stat().st_size,
                    sha256=hashlib.file_digest(f, 'sha256').hexdigest())


def save(path, value):
    with Path(path).open('x') as f:
        json.dump(value, f, indent=2)
        f.write('\n')


def capture(argv, cwd, env, folder, label, timeout=30):
    out, err = folder/(label+'.stdout.log'), folder/(label+'.stderr.log')
    with out.open('xb') as stdout, err.open('xb') as stderr:
        try:
            process = subprocess.run(argv, cwd=cwd, env=env, stdout=stdout, stderr=stderr, timeout=timeout)
        except subprocess.TimeoutExpired:
            save(folder/(label+'-exit.json'), dict(argv=argv, timed_out=True, timeout_seconds=timeout,
                 stdout=record(out), stderr=record(err)))
            raise
    save(folder/(label+'-exit.json'), dict(argv=argv, returncode=process.returncode, timed_out=False,
                                          stdout=record(out), stderr=record(err)))
    assert process.returncode == 0, (label, process.returncode, str(err))
    return out.read_text()


def loaders_from_disassembly(text):
    lines, found = text.splitlines(), []
    for width, mnemonic, register in [(1, 'ldrb', 'w'), (2, 'ldrh', 'w'), (4, 'ldr', 'w'), (8, 'ldr', 'x')]:
        symbol = '_ddi_initial_load'+str(width)+'_4e0b30:'
        start = lines.index(symbol)+1
        instructions = []
        for line in lines[start:]:
            if re.fullmatch(r'\S+:', line):
                break
            if re.match(r'^[0-9a-fA-F]{16}\s', line):
                instructions.append(line)
        assert instructions and re.search(r'\b'+mnemonic+r'\s+'+register+r'\d+,', instructions[0]), (width, instructions)
        assert len([line for line in instructions if re.search(r'\b(?:ldr|ldrb|ldrh)\b', line)]) == 1
        assert any(re.search(r'\bret\b', line) for line in instructions)
        found.append(dict(width=width, symbol=symbol, first_instruction=instructions[0], body=instructions,
                          first_faulting_load_is_function_entry=True))
    return found


def dependency_records(depfile, cwd):
    text = depfile.read_text().replace('\\\n', ' ')
    paths = shlex.split(text.split(':', 1)[1])
    assert paths
    found, missing = {}, []
    for path in paths:
        resolved = Path(path)
        resolved = (resolved if resolved.is_absolute() else cwd/resolved).resolve()
        if resolved.is_file():
            found[str(resolved)] = record(resolved)
        else:
            missing.append(str(resolved))
    return found, missing


def archive_inputs(repo, folder, archive, dependencies, source_facts, *, allow_changed=False):
    files = {'tools/prepare-default-damage-initial-worker.py': Path(__file__).resolve()}
    files.update({'source/'+name: repo/name for name in source_facts})
    expected_inputs = {'source/'+name: expected for name, expected in source_facts.items()}
    for path in dependencies:
        p = Path(path)
        if p.is_relative_to(repo):
            member = 'repository-dependency/'+str(p.relative_to(repo))
        else:
            # Archive actual transitive SDK headers, not the entire SDK.
            # Content-addressing deduplicates identical external dependencies.
            member = 'external-dependency/'+dependencies[path]['sha256']+'/'+p.name
        files[member] = p
        expected_inputs[member] = dependencies[path]
    for path in folder.rglob('*'):
        if path.is_file():
            files['build/'+str(path.relative_to(folder))] = path
    unavailable = {name: str(path) for name, path in files.items() if not path.is_file()}
    files = {name: path for name, path in files.items() if path.is_file()}
    facts = {name: record(path) for name, path in files.items()}
    mismatches = {name: dict(expected=expected, actual=facts.get(name))
                  for name, expected in expected_inputs.items() if facts.get(name) != expected}
    binding = folder/('archive-input-binding-'+archive.stem+'.json')
    save(binding, dict(expected_source_and_transitive_dependency_records=expected_inputs,
         observed_records={name: facts.get(name) for name in expected_inputs},
         unavailable_input_paths=unavailable, mismatches=mismatches,
         all_expected_bytes_exact=not mismatches and not unavailable, failure_preservation_only=allow_changed))
    files['build/'+binding.name] = binding
    facts['build/'+binding.name] = record(binding)
    assert allow_changed or (not mismatches and not unavailable), 'Input changed before PREEXEC archive; preserve failure, never accept new baseline'
    with zipfile.ZipFile(archive, 'x', compression=zipfile.ZIP_DEFLATED, compresslevel=6, allowZip64=True) as z:
        for name, path in files.items():
            z.write(path, name)
    with zipfile.ZipFile(archive) as z:
        assert len(z.namelist()) == len(set(z.namelist())) == len(facts)
        for name, expected in facts.items():
            with z.open(name) as f:
                assert z.getinfo(name).file_size == expected['bytes']
                assert hashlib.file_digest(f, 'sha256').hexdigest() == expected['sha256']
    return record(archive), facts


def prepare_for_suite(repo, go, report, environment=None):
    """Called once per Go invocation, outside the test executable.

    Caller may supply all ordinary Go flags/packages unchanged. Artifact output
    must be unique per invocation. Repeated -count/-cpu full-fixture invocations
    inside ONE test process need a separately reviewed lifecycle contract.
    """
    repo, report = Path(repo).resolve(), Path(report).resolve()
    env = dict(os.environ if environment is None else environment)
    query_env = dict(env, GOTOOLCHAIN='local', GOENV='off')
    facts = json.loads(subprocess.check_output([str(go), 'env', '-json', 'GOVERSION', 'GOOS', 'GOARCH', 'CGO_ENABLED', 'CC'],
                                               cwd=repo/'src', env=query_env, text=True, timeout=30))
    if (facts['GOOS'], facts['GOARCH']) != ('darwin', 'arm64'):
        return dict(applicable=False, environment={}, target_facts=facts,
                    reason='Four fixture Go files select darwin && arm64; no tests weakened or skipped')
    present = [name for name in SOURCES if (repo/name).exists()]
    if not present:
        return dict(applicable=False, environment={}, target_facts=facts,
                    reason='This source revision predates the fixture; no fixture test exists')
    assert set(present) == set(SOURCES), 'Incomplete source installation; never hide fixture tests'
    assert platform.system() == 'Darwin' and platform.machine() == 'arm64', 'Native execution target/host required'
    assert facts['CGO_ENABLED'] == '1'
    assert facts['GOVERSION'] == (repo/'toolchain/go-version.txt').read_text().strip()
    for name in SOURCES:
        if name.endswith('.go'):
            assert (repo/name).read_bytes().startswith(b'//go:build darwin && arm64\n\n')
    source_facts = {name: record(repo/name) for name in SOURCES}
    report.mkdir(parents=True, exist_ok=True)
    folder, artifacts = report/'default-damage-initial-worker-build', report/'initial-artifacts'
    assert not folder.exists() and not artifacts.exists(), 'Refuse reusable worker/case output directory'
    folder.mkdir(mode=0o700)
    save(folder/'actual-Go-target-facts.json', facts)
    dependencies, archive = {}, report/'default-damage-initial-worker-inputs-PREEXEC.zip'
    binary = folder/'default-damage-initial-worker-arm64'
    try:
        requested_sdk = env.get('SDKROOT')
        if not requested_sdk:
            requested_sdk = capture(['/usr/bin/xcrun', '--sdk', 'macosx', '--show-sdk-path'], repo, env, folder, 'SDK-path').strip()
        sdk = Path(requested_sdk).resolve(strict=True)
        assert sdk.is_dir()
        cc = shlex.split(facts['CC'])
        assert cc and shutil.which(cc[0], path=env.get('PATH'))
        cc[0] = shutil.which(cc[0], path=env.get('PATH'))
        version = capture([*cc, '--version'], repo, env, folder, 'compiler-version')
        assert 'clang' in version.lower()
        depfile = folder/'actual-compiler-dependencies.d'
        command = [*cc, *FLAGS, '-I', str(repo/'src/legacy'), '-isysroot', str(sdk), '-MD', '-MF', str(depfile),
                   str(repo/SOURCES[-2]), '-lz', '-o', str(binary)]
        sdk_settings = {str(p): record(p) for p in [sdk/'SDKSettings.json', sdk/'SDKSettings.plist'] if p.is_file()}
        save(folder/'compile-intent.json', dict(command=command, compiler_version=version, compiler=record(cc[0]),
             SDK_requested=requested_sdk, SDK_resolved=str(sdk), SDK_settings=sdk_settings,
             source_files=source_facts, target_facts=facts, helper=record(__file__),
             caller_BEFORE_flag_unchanged=env.get('OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE'),
             worker_test_or_original_executions=0))
        capture(command, repo, env, folder, 'compile', timeout=45)
        dependencies, missing = dependency_records(depfile, repo)
        save(folder/'actual-dependency-discovery.json', dict(dependencies=dependencies, missing=missing))
        assert not missing, 'Compiler dependency no longer present'
        assert (folder/'compile.stderr.log').stat().st_size == 0, 'Compiler warnings retained, not silently accepted'
        with binary.open('rb') as f:
            assert f.read(8) == bytes.fromhex('cffaedfe0c000001')
        disassembly = capture(['/usr/bin/otool', '-arch', 'arm64', '-tvV', str(binary)], repo, env, folder, 'actual-disassembly')
        loaders = loaders_from_disassembly(disassembly)
        save(folder/'actual-loaders-and-dependencies.json', dict(worker=record(binary), loaders=loaders,
             compiler_dependencies=dependencies, all_transitive_headers_byte_preserved_in_PREEXEC_archive=True))
        for name, expected in source_facts.items():
            assert record(repo/name) == expected
        for name, expected in dependencies.items():
            assert record(name) == expected
        proof = dict(schema='opennox-portable-default-damage-initial-worker-setup-v1', applicable=True,
                     worker=record(binary), loaders=loaders, compiler_dependencies=dependencies,
                     source_files=source_facts, compiler_or_SDK_source_closure_preserved=True,
                     worker_executions=0, Go_test_executions=0, parent_engine_leaves=0)
        save(folder/'setup-before-archive.json', proof)
        archived, members = archive_inputs(repo, folder, archive, dependencies, source_facts)
        for name, expected in source_facts.items():
            assert record(repo/name) == expected
        for name, expected in dependencies.items():
            assert record(name) == expected
        python = str(Path(sys.executable).resolve())
        variables = dict(OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_WORKER=str(binary),
                         OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_WORKER_SHA256=proof['worker']['sha256'],
                         OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_ARTIFACTS=str(artifacts),
                         OPENNOX_RAW_DEFAULT_DAMAGE_INITIAL_PYTHON=python)
        # BEFORE is deliberately NOT returned/forced; ordinary and future engine
        # phases select it independently without changing frozen test assertions.
        result = dict(**proof, environment=variables, preexecution_archive=archived, archive_members=members,
                      all_archive_member_bytes_verified=True, target_facts=facts,
                      original_caller_BEFORE_flag=env.get('OPENNOX_RAW_DEFAULT_DAMAGE_BEFORE'),
                      worker_and_case_outputs_owned_by_this_report=True)
        save(report/'default-damage-initial-worker-setup.json', result)
        return result
    except BaseException as exc:
        # A compiler may write a valid .d even when compilation, disassembly or
        # a loader assertion fails. Preserve every known available header.
        depfile = folder/'actual-compiler-dependencies.d'
        if depfile.is_file():
            try:
                available, missing = dependency_records(depfile, repo)
                for name, observed in available.items():
                    dependencies.setdefault(name, observed)
                save(folder/'failure-dependency-discovery.json', dict(available=available, missing=missing))
            except (ValueError, OSError, AssertionError, IndexError) as dep_error:
                save(folder/'failure-dependency-discovery.json', dict(error=repr(dep_error),
                     unparsed_dependency_file=record(depfile), source_closure_not_qualified=True))
        save(folder/'setup-failure.json', dict(error=repr(exc), traceback=traceback.format_exc(),
             worker_or_Go_test_or_original_execution=False, partial_binary_retained=binary.exists()))
        failure_archive = report/'default-damage-initial-worker-failure-inputs.zip'
        archived, members = archive_inputs(repo, folder, failure_archive, dependencies, source_facts, allow_changed=True)
        save(report/'default-damage-initial-worker-setup-failure.json', dict(archive=archived, members=members,
             all_available_failure_bytes_retained=True, tests_not_started_by_helper=True))
        raise


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--repo', type=Path, required=True)
    parser.add_argument('--go', required=True)
    parser.add_argument('--report', type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(prepare_for_suite(args.repo, args.go, args.report), indent=2))


if __name__ == '__main__':
    main()
