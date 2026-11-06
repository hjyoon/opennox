"""Same-PID exec only; no graph/algorithm interpretation or compilation."""
from pathlib import Path
import hashlib,json,os,sys
binary,binary_sha,plan,plan_sha,record=sys.argv[1:]
facts=[]
for name,expected in [(binary,binary_sha),(plan,plan_sha)]:
    data=Path(name).read_bytes();assert hashlib.sha256(data).hexdigest()==expected
    facts.append(dict(path=name,bytes=len(data),sha256=expected))
payload=dict(pid=os.getpid(),before_exec=True,binary=facts[0],input_gzip=facts[1],
             argv=[binary,'--input-gzip',plan],python=str(Path(sys.executable).resolve()),
             original_or_Go_parent_calls=0)
with Path(record).open('x')as f:
    json.dump(payload,f,indent=2);f.write('\n');f.flush();os.fsync(f.fileno())
os.execv(binary,[binary,'--input-gzip',plan])
