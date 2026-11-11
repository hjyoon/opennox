"""Retain and execute Windows foreign-allocation controls before full suites."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--report', type=Path, required=True)
    parser.add_argument('--profile', choices=['normal', 'report', 'checkptr', 'gccheckmark', 'race'], default='normal')
    args = parser.parse_args()
    report = args.report.resolve() / 'allocator-controls'
    report.mkdir(parents=True)
    repo = Path(__file__).resolve().parents[1]
    facts = json.loads(subprocess.check_output(
        ['go', 'env', '-json', 'GOVERSION', 'GOOS', 'GOARCH', 'CGO_ENABLED', 'GOEXPERIMENT'], text=True))
    assert os.name == 'nt' and facts['GOOS'] == 'windows'
    assert facts['GOVERSION'] == (repo / 'toolchain/go-version.txt').read_text().strip()
    assert facts['GOARCH'] in ['386', 'amd64'] and facts['CGO_ENABLED'] == '1'
    requirements = {
        'alloc': ['TestHighAllocationOwnershipAndRealloc',
                  'TestHighFixtureTypedFramesAndClassOwnership',
                  'TestHighAllocationRejectsInvalidExtents'],
        'handles': ['TestHighOpaqueFixtureBank'],
        'memguard': ['TestNewPage', 'TestNewHighPageProtection'],
    }
    for package, required in requirements.items():
        folder = report / package
        folder.mkdir()
        binary = folder / (package + '.test.exe')
        relative = './legacy/common/alloc' + ('/' + package if package != 'alloc' else '')
        command = ['go', 'test', '-c', '-o=' + str(binary)]
        if args.profile == 'checkptr':
            assert facts['GOEXPERIMENT'] == 'cgocheck2'
            command += ['-gcflags=all=-d=checkptr=2']
        if args.profile == 'race':
            command += ['-race']
        command += [relative]
        (folder / 'intent.json').write_text(json.dumps(dict(
            profile=args.profile, target=facts, command=command), indent=2) + '\n')
        with (folder / 'build.stdout.log').open('xb') as out, (folder / 'build.stderr.log').open('xb') as err:
            built = subprocess.run(command, cwd=repo / 'src', stdout=out, stderr=err, timeout=300)
        assert built.returncode == 0, 'Allocator control build failed; original logs retained'
        with binary.open('rb') as stream:
            before = hashlib.file_digest(stream, 'sha256').hexdigest()
        (folder / 'binary-before-execution.json').write_text(json.dumps(dict(
            name=binary.name, bytes=binary.stat().st_size, sha256=before), indent=2) + '\n')
        with (folder / 'test.stdout.log').open('xb') as out, (folder / 'test.stderr.log').open('xb') as err:
            tested = subprocess.run([str(binary), '-test.v', '-test.count=1', '-test.timeout=5m'],
                                    cwd=repo / 'src' / relative, stdout=out, stderr=err, timeout=330)
        with binary.open('rb') as stream:
            after = hashlib.file_digest(stream, 'sha256').hexdigest()
        summary = dict(profile=args.profile, target=facts, exit_code=tested.returncode,
                       binary_sha256_before=before, binary_sha256_after=after)
        text = (folder / 'test.stdout.log').read_text()
        summary['required_controls_passed'] = {name: ('--- PASS: ' + name + ' (') in text for name in required}
        (folder / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
        print(json.dumps(summary), flush=True)
        assert tested.returncode == 0 and before == after, 'Allocator controls failed; full evidence retained'
        assert all(summary['required_controls_passed'].values()), 'Missing explicit high-allocation control execution'



if __name__ == '__main__':
    main()
