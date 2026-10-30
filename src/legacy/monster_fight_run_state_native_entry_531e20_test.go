package legacy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Only the script body is a declared service here. Audio allocation, frame
// copy, public C dispatch and the real sight routine execute. This fixture
// does not claim script VM execution; the adapter's extra event13 is checked.
type fightRunNativeNoxScriptBase531E20 = NoxScript

type fightRunNativeScript531E20 struct {
	fightRunNativeNoxScriptBase531E20
	t       *testing.T
	unit    *server.Object
	caller  *server.Object
	block   *server.ScriptCallback
	calls   int
	observe func()
}

func (s *fightRunNativeScript531E20) ScriptCallback(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) unsafe.Pointer {
	if block != s.block || caller != s.caller || trigger != s.unit || event != server.NoxEventMonsterFightStart {
		s.t.Fatal("fight script cached identity or separate adapter event13")
	}
	s.calls++
	s.observe()
	return nil
}

type fightRunNativeServer531E20 struct {
	Server
	script *fightRunNativeScript531E20
}

func (s *fightRunNativeServer531E20) NoxScriptC() NoxScript { return s.script }

func TestMonsterFightRunStateNextPublicCNative531E20(t *testing.T) {
	const child = "NOX_MONSTER_FIGHT_RUN_NATIVE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		for _, entry := range []uint32{0x531e20, 0x531e90, 0x534750, 0x534780, 0x544740, 0x544750, 0x544920, 0x544930, 0x544940, 0x5456b0, 0x5456c0, 0x546410, 0x546420} {
			for slot := 0; slot < 2; slot++ {
				for policy := 0; policy < 2; policy++ {
					mode := fmt.Sprintf("%x:%d:%d", entry, slot, policy)
					t.Run(mode, func(t *testing.T) {
						bin, err := os.Executable()
						if err != nil {
							t.Fatal(err)
						}
						cmd := exec.Command(bin, "-test.run=^TestMonsterFightRunStateNextPublicCNative531E20$", "-test.v")
						cmd.Env = append(os.Environ(), child+"="+mode)
						out, err := cmd.CombinedOutput()
						t.Logf("%s", out)
						if err != nil {
							t.Fatalf("actual native fight/run entry: %v", err)
						}
					})
				}
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 3 {
		t.Fatal("fight/run mode")
	}
	entry64, err := strconv.ParseUint(parts[0], 16, 32)
	if err != nil {
		t.Fatal(err)
	}
	entry := uint32(entry64)
	slot, err := strconv.Atoi(parts[1])
	if err != nil || slot < 0 || slot > 1 {
		t.Fatal("fight/run slot")
	}
	policy, err := strconv.Atoi(parts[2])
	if err != nil || policy < 0 || policy > 1 {
		t.Fatal("fight/run policy")
	}
	s, unit, first, other := monsterMainBlockLegacyFixture547210(t)
	second, free := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(free)
	*second = *first
	updates := [2]*server.MonsterUpdateData{first, second}
	var sounds [2]*[32]uint32
	for i := 0; i < 2; i++ {
		sounds[i], free = alloc.New([32]uint32{})
		t.Cleanup(free)
		sounds[i][5] = uint32(sound.SoundHammerMissing)
		updates[i].SoundSet122 = unsafe.Pointer(sounds[i])
		updates[i].StatusFlags = object.MonsterStatus(0x80004100 | uint32(policy)*0x18000)
		updates[i].Field301, updates[i].Field303 = 1399, 1400
		updates[i].CurrentEnemy = other
		updates[i].ScriptChangeFocus = server.ScriptCallback{Flags: 0xa5, Func: 19}
	}
	unit.UpdateData = unsafe.Pointer(updates[slot])
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(first), unsafe.Pointer(second), unsafe.Pointer(sounds[0]), unsafe.Pointer(sounds[1])} {
			if uintptr(p) <= 0xffffffff {
				t.Fatal("native fight/run records must exceed4GiB")
			}
		}
	}
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	if !s.Audio.ReadAVNT(&playerAttackHammerAudioFile538960{data: []byte{7, 1, 'x', 0, 0}}) {
		t.Fatal("fight sound descriptor")
	}
	count := 0
	s.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, pos types.Pointf) {
		if id != sound.SoundHammerMissing || kind != 0 || owner != unit || pos != unit.PosVec {
			t.Fatal("real fight audio payload")
		}
		count++
	})
	copyPtr := memmap.PtrUint32(0x5D4594, 2487684)
	oldCopy := *copyPtr
	t.Cleanup(func() { *copyPtr = oldCopy })
	const copyInitial = 0x11223344
	*copyPtr = copyInitial
	beforeUnit, beforeHealth, beforeOther := *unit, *unit.HealthData, *other
	beforeUpdates := [2]server.MonsterUpdateData{*first, *second}
	beforeSounds := [2][32]uint32{*sounds[0], *sounds[1]}
	runStart := entry == 0x531e20 || entry == 0x534750 || entry == 0x544740 || entry == 0x544920 || entry == 0x5456b0
	wantCopy, wantSounds, wantScript := uint32(copyInitial), 0, 0
	script := &fightRunNativeScript531E20{t: t, unit: unit, caller: other, block: &updates[slot].ScriptChangeFocus}
	outer, oldServer := &fightRunNativeServer531E20{Server: GetServer(), script: script}, GetServer
	GetServer = func() Server { return outer }
	t.Cleanup(func() { GetServer = oldServer })
	script.observe = func() {
		if count != 1 || *copyPtr != copyInitial || updates[slot].StatusFlags != beforeUpdates[slot].StatusFlags {
			t.Fatal("script did not follow real audio and precede alert/frame copy")
		}
	}
	invokeGo := func() {
		switch entry {
		case 0x531e20:
			cgoAIAction{typ: ai.ACTION_FIGHT}.Start(unit)
		case 0x531e90:
			s.MonsterActionFightEnd531E90(unit)
		default:
			if runStart {
				s.MonsterActionRunStart534750(unit)
			} else {
				s.MonsterActionRunEnd534780(unit)
			}
		}
	}
	wantUpdates := beforeUpdates
	apply := func(update int) {
		if runStart && !wantUpdates[update].StatusFlags.Has(object.MonStatusNeverRun) {
			wantUpdates[update].StatusFlags |= object.MonStatusRunning
		} else if !runStart && !wantUpdates[update].StatusFlags.Has(object.MonStatusAlwaysRun) {
			wantUpdates[update].StatusFlags &^= object.MonStatusRunning
		}
	}
	if entry == 0x531e20 {
		wantUpdates[slot].StatusFlags |= object.MonStatusAlert
		wantCopy, wantSounds, wantScript = s.Frame(), 1, 1
	}
	if entry == 0x531e90 {
		wantUpdates[slot].StatusFlags &^= object.MonStatusAlert
	}
	apply(slot)
	check := func(wantUnit server.Object) {
		t.Helper()
		queued := 0
		s.Audio.EachEvent(func(event *server.AudioEvent) {
			if event.Sound != sound.SoundHammerMissing || event.Obj != unit || event.Kind != 0 || event.Code != 0 || event.Pos != unit.PosVec {
				t.Fatal("actual fight event queue")
			}
			queued++
		})
		if count != wantSounds || queued != wantSounds || script.calls != wantScript || *copyPtr != wantCopy {
			t.Fatalf("fight events queued/callback=%d/%d script=%d copy=%#x", queued, count, script.calls, *copyPtr)
		}
		compare := func(got, want unsafe.Pointer, n uintptr) {
			t.Helper()
			if !bytes.Equal(unsafe.Slice((*byte)(got), int(n)), unsafe.Slice((*byte)(want), int(n))) {
				t.Fatal("fight/run whole native record differs")
			}
		}
		compare(unit.CObj(), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit))
		compare(unsafe.Pointer(unit.HealthData), unsafe.Pointer(&beforeHealth), unsafe.Sizeof(beforeHealth))
		compare(other.CObj(), unsafe.Pointer(&beforeOther), unsafe.Sizeof(beforeOther))
		for i := 0; i < 2; i++ {
			compare(unsafe.Pointer(updates[i]), unsafe.Pointer(&wantUpdates[i]), unsafe.Sizeof(wantUpdates[i]))
			compare(unsafe.Pointer(sounds[i]), unsafe.Pointer(&beforeSounds[i]), unsafe.Sizeof(beforeSounds[i]))
		}
	}
	s.Audio.Reset()
	invokeGo()
	check(beforeUnit)
	t.Logf("FIGHT/RUN CONTROL COMPLETED: mode=%s unit=%p update=%p signed-status=%d real-audio=%d declared-script=%d real-sight=true", mode, unit, updates[slot], int32(updates[slot].StatusFlags), count, script.calls)
	*unit, *first, *second = beforeUnit, beforeUpdates[0], beforeUpdates[1]
	*copyPtr, count, script.calls = copyInitial, 0, 0
	s.Audio.Reset()
	wantUnit, liveSlot := beforeUnit, slot
	if entry == 0x531e20 {
		// The real script boundary changes the live update. Fight's cached
		// alert write and the later live run write must reach different records.
		wantUpdates = beforeUpdates
		wantUpdates[slot].StatusFlags |= object.MonStatusAlert
		liveSlot = 1 - slot
		wantUnit.UpdateData = unsafe.Pointer(updates[liveSlot])
		script.observe = func() {
			if count != 1 || *copyPtr != copyInitial {
				t.Fatal("actual C script order")
			}
			unit.UpdateData = unsafe.Pointer(updates[liveSlot])
		}
		apply(liveSlot)
	}
	fn := monsterFightRunStatePublicC531E20(entry)
	if fn == nil {
		t.Fatal("original escort end/cancel public C symbols are not linked")
	}
	result := ccall.CallIntPtr(fn, unit.CObj())
	if result != int(int32(wantUpdates[liveSlot].StatusFlags)) {
		t.Fatalf("actual C DWORD status return=%d want=%d", result, int32(wantUpdates[liveSlot].StatusFlags))
	}
	check(wantUnit)
	t.Logf("FIGHT/RUN NATIVE C PASS: mode=%s full-records=true real-audio=%d declared-script=%d real-sight=true signed-status=%d", mode, count, script.calls, result)
}
