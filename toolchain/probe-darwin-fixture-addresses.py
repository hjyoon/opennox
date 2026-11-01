"""Compare actual original-derived C fixtures with two address allocators."""
from pathlib import Path
import collections,hashlib,json,os,platform,struct,subprocess

repo=Path.cwd().resolve()
report=Path(os.environ.get('FIXTURE_REPORT_DIR',repo/'build'/'darwin-fixture-address-probe'))
report.mkdir(parents=True)
assert platform.system()=='Darwin' and platform.machine()=='arm64'
assert subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()==os.environ['SOURCE_REVISION']
candidate=Path(os.environ['FIXTURE_CANDIDATE_HEADER'])
legacy=repo/'src'/'legacy'
data=legacy/'testdata'
original=json.loads((data/'center_messages_445450_445530_original.json').read_text())
assert original['original_sha256']=='0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
assert len(original['clear_rows'])==9 and len(original['draw_rows'])==288
packed=bytearray(struct.pack('<8iI',*original['shadow_offsets'],297))
ids={'viewport':1,'color':2,'measure':3,'draw':4,'height':5,'copy':6}
def emit(fields,row):
    expected=row['expected'];events=expected['events'];trace=[0]*160
    for i,e in enumerate(events):
        assert len(e)<=5
        values=[ids[e[0]],*e[1:]]
        trace[5*i:5*i+len(values)]=values
    result=expected['returned_slot']if fields[0]==0 else expected['returned']
    packed.extend(struct.pack('<10qI160qI',*fields,len(events),*trace,result))
for row in original['clear_rows']:
    x=row['input'];length=len(x['text']);assert length in [0,1,4]
    emit([0,x['index'],0,0,0,0,0,0,0,{0:0,1:1,4:2}[length]],row)
for row in original['draw_rows']:
    x=row['input'];emit([1,x['index'],x['mask'],x['frame'],x['width'],x['height'],x['top'],x['measured'],x['font'],0],row)
(report/'center-original-input.bin').write_bytes(packed)
def program(source,fixture,first,last,marker):
    text=(legacy/source).read_text()
    assert text.count(first)==text.count(last)==1
    body=first+text.split(first,1)[1].split(last,1)[0]
    template=(data/fixture).read_text();assert template.count(marker)==1
    return template.replace(marker,body,1)
programs={
 'ability':program('GAME2_1.c','client_ability_state_test.c','//----- (00461090)','//----- (004611A0)','// COMPLETE_SUBSYSTEM'),
 'center':program('GAME1_3.c','client_center_messages_test.c','//----- (00445450)','//----- (004456E0)','// CENTER_BODIES'),
 'control':r'''
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include "native_exact_mapping.h"
int main(void) {
 const uintptr_t original[]={0x37ffc0000ull,0x3fffc0000ull,0x440000000ull,0x47ffc0000ull,0x4fffc0000ull,0x540000000ull};
 uintptr_t delta=OPENNOX_NATIVE_TEST_ADDRESS(original[0])-original[0];
 if(delta&UINT32_MAX)return 1;
 for(unsigned i=0;i<6;i++) {
  uintptr_t address=OPENNOX_NATIVE_TEST_ADDRESS(original[i]);
  if(address-original[i]!=delta||(uint32_t)address!=(uint32_t)original[i]||address<=UINT32_MAX)return 2;
  unsigned char* memory=opennox_native_test_exact_mapping(address,0x80000);
  if((uintptr_t)memory!=address)return 3;
  memset(memory,0x5a,0x80000);
  mach_vm_address_t occupied=address;
  if(mach_vm_allocate(mach_task_self(),&occupied,0x80000,VM_FLAGS_FIXED)==KERN_SUCCESS)return 4;
  if(opennox_native_test_exact_mapping(address,0x80000)!=NULL)return 5;
  for(size_t j=0;j<0x80000;j++)if(memory[j]!=0x5a)return 6;
  opennox_native_test_release_mapping(memory,0x80000);
  memory=opennox_native_test_exact_mapping(address,0x80000);
  if((uintptr_t)memory!=address)return 7;
  for(size_t j=0;j<0x80000;j++)if(memory[j]!=0)return 8;
  opennox_native_test_release_mapping(memory,0x80000);
 }
 printf("NATIVE_LAYOUT_PASS delta=0x%llx low32/relative/boundaries=true kernel-collision-refused=true all-canaries=true zeroed-reuse=true\n",(unsigned long long)delta);
 return 0;
}
''',
}
results=[]
for variant,header in [('before',data/'native_exact_mapping.h'),('reserved',candidate)]:
    folder=report/variant;folder.mkdir();(folder/'native_exact_mapping.h').write_bytes(header.read_bytes())
    for name,code in programs.items():
        if name=='control'and variant=='before':continue
        if variant=='reserved'and name!='control':
            # Placement only: preserve all low DWORDs, relative distances,
            # boundary crossings, original inputs and every assertion.
            for address in ['37ffc0000','3fffc0000','440000000','47ffc0000','4fffc0000','540000000']:
                value='(uintptr_t)0x'+address+'ull'
                code=code.replace(value,'OPENNOX_NATIVE_TEST_ADDRESS('+value+')')
        src=folder/(name+'.c');src.write_text(code);binary=folder/name
        args=['clang','-std=c11','-O2','-fshort-wchar','-fsigned-char','-fno-strict-aliasing','-fno-strict-overflow','-Wall','-Wextra','-Werror','-I'+str(folder),'-I'+str(data),'-include',str(data/'native_stdio.h'),str(src),'-o',str(binary)]
        if name=='ability':args.insert(1,'-Wno-unused-function') # Same existing fixture flags.
        compile=subprocess.run(args,cwd=legacy,capture_output=True)
        (folder/(name+'-compile.stdout.log')).write_bytes(compile.stdout)
        (folder/(name+'-compile.stderr.log')).write_bytes(compile.stderr)
        assert compile.returncode==0,compile.stderr.decode()
        modes=['state','reset','active','lookup']if name=='ability'else['clear','draw']if name=='center'else['invariants']
        for mode in modes:
            for repeat in range(30):
                run=subprocess.run([str(binary),mode],input=bytes(packed)if name=='center'else None,cwd=legacy,capture_output=True)
                label=f'{name}-{mode}-{repeat}'
                (folder/(label+'.stdout.log')).write_bytes(run.stdout)
                (folder/(label+'.stderr.log')).write_bytes(run.stderr)
                row=dict(variant=variant,fixture=name,mode=mode,repeat=repeat,exit_code=run.returncode,stderr=run.stderr.decode(errors='replace'))
                results.append(row)
                with(report/'runs.jsonl').open('a')as out:out.write(json.dumps(row)+'\n')
    print(json.dumps(dict(variant=variant,exits=dict(collections.Counter(r['exit_code']for r in results if r['variant']==variant)))),flush=True)
proof=dict(source=os.environ['SOURCE_REVISION'],workflow=os.environ['GATE_WORKFLOW_REVISION'],native_OS=platform.platform(),native_CPU=platform.machine(),original_cases_and_assertions_unchanged=True,native_placement_translation_preserves_low32_relative_distances_and_4GiB_crossings=True,center_input_sha256=hashlib.sha256(packed).hexdigest(),original_JSON_sha256=hashlib.sha256((data/'center_messages_445450_445530_original.json').read_bytes()).hexdigest(),before_header_sha256=hashlib.sha256((data/'native_exact_mapping.h').read_bytes()).hexdigest(),candidate_header_sha256=hashlib.sha256(candidate.read_bytes()).hexdigest(),runs=results)
(report/'summary.json').write_text(json.dumps(proof,indent=2)+'\n')
assert all(r['exit_code']==0 for r in results if r['variant']=='reserved'),[r for r in results if r['variant']=='reserved'and r['exit_code']]
