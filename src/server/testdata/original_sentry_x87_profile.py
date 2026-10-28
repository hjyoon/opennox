"""Reproduce the independent x87 profile used alongside the untouched oracle."""
import argparse,hashlib,json,platform,subprocess
from pathlib import Path
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--fixture',type=Path,required=True);p.add_argument('--profile',type=Path,required=True);p.add_argument('--proof',type=Path,required=True);p.add_argument('--work',type=Path,required=True)
args=p.parse_args();args.work.mkdir(exist_ok=False);fixture=json.loads(args.fixture.read_text())
angles=sorted({r['input']['data'][0][0] for r in fixture['rows'] if r['transcendental']})
inputs=''.join(f'{cw:x} {a:08x}\n' for cw in (0xe3f,0x23f) for a in angles);(args.work/'inputs.txt').write_text(inputs)
source=Path(__file__).with_suffix('.c');binary=args.work/'x87-profile';translated=platform.machine().lower() in ('arm64','aarch64')
command=['cc']+(['-arch','x86_64'] if translated else [])+['-O2',str(source),'-o',str(binary)]
build=subprocess.run(command,text=True,capture_output=True);(args.work/'build.stdout.log').write_text(build.stdout);(args.work/'build.stderr.log').write_text(build.stderr);build.check_returncode()
command=(['/usr/bin/arch','-x86_64'] if translated else [])+[str(binary)]
run=subprocess.run(command,input=inputs,text=True,capture_output=True);(args.work/'run.stdout.log').write_text(run.stdout);(args.work/'run.stderr.log').write_text(run.stderr);run.check_returncode()
rows={}
for line in run.stdout.splitlines():
    cw,angle,cosine,status,raw=line.split();v=int(raw,16);rows[f'{cw}/{angle}/{cosine}']={'significand':v&((1<<64)-1),'exponent':v>>64,'FPSW':int(status,16)}
assert len(rows)==4*len(angles)
scope='Actual x87 instruction opcodes in an isolated Rosetta x86_64 process. Independent of Unicorn and Go. This is translated x87 evidence, not physical x86 hardware. Original CW restored for each sample.' if translated else 'Actual x87 instruction opcodes in an isolated native x86 process. Original CW restored for each sample.'
with args.profile.open('x') as f:json.dump({'scope':scope,'rows':rows},f,indent=2)
sha=lambda x:hashlib.sha256(x.read_bytes()).hexdigest()
with args.proof.open('x') as f:json.dump({'fixture_sha256':sha(args.fixture),'source_sha256':sha(source),'profile_sha256':sha(args.profile),'translated':translated,'angles':len(angles),'instruction_results':len(rows),'scope':scope},f,indent=2)
print(json.dumps({'angles':len(angles),'instruction_results':len(rows),'translated':translated}))
