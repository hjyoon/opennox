"""Execute added class variants without changing the frozen base producer."""
import hashlib
from pathlib import Path
p=Path(__file__).resolve().parent
source=p/'original_monster_state_predicates_534710.py'
if not source.exists():source=p/'original_monster_state_predicates_534710_v1.py'
raw=source.read_bytes()
assert hashlib.sha256(raw).hexdigest()=='e2f62d44e1cbb6c8ada20d4d2052ff3957e88e0d7dd43fc6980590d44f8666f8'
text=raw.decode()
old='rows=[]\nfor case in cases:'
new='''extra=[]
for case in cases:
    if case['entry']in [0x534710,0x5347a0,0x534810]:
        for cls in [0,6,0xffffffff]:
            extra.append(dict(case,name=case['name']+f"-class-{cls:08x}",input=dict(case['input'],cls=cls)))
            per_entry[case['entry']]+=1
assert len(cases)==6831 and len(extra)==9609
cases+=extra
rows=[]
for case in cases:'''
assert text.count(old)==1
text=text.replace(old,new).replace('original-monster-state-predicates-534710-v1','original-monster-state-predicates-classes-534710-v1')
exec(compile(text,str(source),'exec'),{'__file__':str(__file__),'__name__':'__main__'})

