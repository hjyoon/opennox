package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type dyingStartNoxScriptBase544C40 = NoxScript

type dyingStartNativeScript544C40 struct {
	dyingStartNoxScriptBase544C40
	t      *testing.T
	unit   *server.Object
	block  *server.ScriptCallback
	calls  int
	mutate func()
}

func (s *dyingStartNativeScript544C40) ScriptCallback(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) unsafe.Pointer {
	if block != s.block || caller != nil || trigger != s.unit || event != server.NoxEventMonsterDead {
		s.t.Fatal("actual native cached script block/caller/trigger and existing event7 adapter")
	}
	s.calls++
	if s.mutate != nil {
		s.mutate()
	}
	return nil
}

type dyingStartNativeServer544C40 struct {
	Server
	script *dyingStartNativeScript544C40
}

func (s *dyingStartNativeServer544C40) NoxScriptC() NoxScript { return s.script }

type dyingStartNativeExtra544C40 struct {
	Before      [32]byte
	Alternate   server.MonsterUpdateData
	Definitions [3]server.MonsterDef
	Sounds      [2][16]uint32
	After       [32]byte
}

type dyingStartNativeFrame544C40 struct {
	s             *server.Server
	unit, missile *server.Object
	updates       [2]*server.MonsterUpdateData
	extra         *dyingStartNativeExtra544C40
	script        *dyingStartNativeScript544C40
}

func dyingStartNativeFixture544C40(t *testing.T, entry, soundIndex, callback int) dyingStartNativeFrame544C40 {
	t.Helper()
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	p, free := alloc.New(dyingStartNativeExtra544C40{})
	t.Cleanup(free)
	*p = dyingStartNativeExtra544C40{}
	for i := range p.Before {
		p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
	}
	updates := [2]*server.MonsterUpdateData{update, &p.Alternate}
	for n, ud := range updates {
		ud.MonsterDef = &p.Definitions[2]
		ud.SoundSet122 = unsafe.Pointer(&p.Sounds[1][0])
		ud.ScriptDeath = server.ScriptCallback{Flags: 0x5a, Func: 19}
		if n == entry {
			ud.MonsterDef = &p.Definitions[0]
			ud.SoundSet122 = nil
			if soundIndex >= 0 {
				ud.SoundSet122 = unsafe.Pointer(&p.Sounds[soundIndex][0])
			}
			ud.ScriptDeath = server.ScriptCallback{Flags: 0xa5, Func: 17}
		}
	}
	functions := dyingStartStockFunctions544C40()
	for n := range p.Definitions {
		id := n + 1
		if n == 0 {
			id = callback
		}
		p.Definitions[n].DieFunc228 = functions[id]
	}
	for n := range p.Sounds {
		for i := range p.Sounds[n] {
			p.Sounds[n][i] = 0xa5000000 + uint32(n)*0x100 + uint32(i)
		}
		p.Sounds[n][15] = 715 + uint32(n)
	}
	unit.ObjClass, unit.ObjSubClass, unit.ObjFlags = object.ClassMonster, 0xa5, 0x12000004
	unit.PosVec = types.Ptf(321, 654)
	unit.UpdateData = unsafe.Pointer(updates[entry])
	script := &dyingStartNativeScript544C40{t: t, unit: unit, block: &updates[entry].ScriptDeath}
	outer := &dyingStartNativeServer544C40{Server: GetServer(), script: script}
	old := GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = old })
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(updates[0]), unsafe.Pointer(updates[1]), unsafe.Pointer(&p.Definitions[0]), unsafe.Pointer(&p.Sounds[0][0])} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatal("native C-owned unit/update/definition/sound links must exceed4GiB")
			}
		}
	}
	return dyingStartNativeFrame544C40{s: s, unit: unit, missile: missile, updates: updates, extra: p, script: script}
}

func dyingStartStockFunctions544C40() [6]unsafe.Pointer {
	return [6]unsafe.Pointer{nil, monsterDieFunctions["SWORDSMANDIE"], monsterDieFunctions["URCHINSHAMANDIE"], monsterDieFunctions["ARCHERDIE"], monsterDieFunctions["OGREDIE"], monsterDieFunctions["OGREWARLORDDIE"]}
}

func TestMonsterDyingStartNextPublicCNative544C40(t *testing.T) {
	const child = "NOX_MONSTER_DYING_START_NATIVE_CHILD"
	if os.Getenv(child) == "" {
		for _, mode := range []string{"dying", "getter"} {
			t.Run(mode, func(t *testing.T) {
				bin, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(bin, "-test.run=^TestMonsterDyingStartNextPublicCNative544C40$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual native dying-start/getter C path: %v", err)
				}
			})
		}
		return
	}
	f := dyingStartNativeFixture544C40(t, 0, 0, 0)
	beforeUnit, beforeUpdate, beforeExtra, beforeMissile := *f.unit, *f.updates[0], *f.extra, *f.missile
	a, ok := server.GetAIAction(ai.ACTION_DYING).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary dying dispatcher")
	}
	a.Start(f.unit)
	if f.script.calls != 1 {
		t.Fatal("ordinary cached death-script call")
	}
	raw := func(p unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(p), int(n)) }
	check := func() {
		for _, v := range []struct {
			got, want unsafe.Pointer
			n         uintptr
		}{{f.unit.CObj(), unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)}, {unsafe.Pointer(f.updates[0]), unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)}, {unsafe.Pointer(f.extra), unsafe.Pointer(&beforeExtra), unsafe.Sizeof(beforeExtra)}, {f.missile.CObj(), unsafe.Pointer(&beforeMissile), unsafe.Sizeof(beforeMissile)}} {
			if !bytes.Equal(raw(v.got, v.n), raw(v.want, v.n)) {
				t.Fatal("full native dying/getter canaries changed")
			}
		}
	}
	check()
	t.Logf("DYING START ORDINARY COMPLETED: mode=%s unit=%p update=%p sound=%p script=1 callback=nil", os.Getenv(child), f.unit, f.updates[0], f.updates[0].SoundSet122)
	f.script.calls = 0
	if os.Getenv(child) == "getter" {
		got := ccall.CallUPtrPtr(monsterSoundSetPublicC424300(), f.unit.CObj())
		if got != uintptr(f.updates[0].SoundSet122) || f.script.calls != 0 {
			t.Fatalf("native sound pointer=%x want=%p script=%d", got, f.updates[0].SoundSet122, f.script.calls)
		}
	} else {
		if got := ccall.CallIntPtr(monsterDyingStartPublicC544C40(), f.unit.CObj()); got != 0 || f.script.calls != 1 {
			t.Fatal("native public int result/script count", got, f.script.calls)
		}
	}
	check()
	t.Logf("DYING START NATIVE C PASS: mode=%s original-public-result-and-sound-pointer=true full-records=true", os.Getenv(child))
}
