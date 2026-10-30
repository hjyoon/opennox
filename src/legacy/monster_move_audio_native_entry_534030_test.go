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

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterMoveAudioNextPublicCNative534030(t *testing.T) {
	const child = "NOX_MONSTER_MOVE_AUDIO_NATIVE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		for _, kind := range []string{"frame-A", "frame-B", "not-frame", "latched", "nil-sound"} {
			for slot := 0; slot < 2; slot++ {
				mode := fmt.Sprintf("%s:%d", kind, slot)
				t.Run(mode, func(t *testing.T) {
					bin, err := os.Executable()
					if err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(bin, "-test.run=^TestMonsterMoveAudioNextPublicCNative534030$", "-test.v")
					cmd.Env = append(os.Environ(), child+"="+mode)
					out, err := cmd.CombinedOutput()
					t.Logf("%s", out)
					if err != nil {
						t.Fatalf("actual native MoveAudio entry: %v", err)
					}
				})
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 2 {
		t.Fatal("MoveAudio mode")
	}
	slot, err := strconv.Atoi(parts[1])
	if err != nil || slot < 0 || slot > 1 {
		t.Fatal("MoveAudio native slot")
	}
	s, unit, update, other := monsterMainBlockLegacyFixture547210(t)
	var defs [2]*server.MonsterDef
	var sounds [2]*[32]uint32
	for i := 0; i < 2; i++ {
		var free func()
		defs[i], free = alloc.New(server.MonsterDef{})
		t.Cleanup(free)
		*defs[i] = server.MonsterDef{MoveSndFrameA100: 1, MoveSndFrameB104: 2}
		sounds[i], free = alloc.New([32]uint32{})
		t.Cleanup(free)
		sounds[i][18] = uint32(sound.SoundHammerMissing)
	}
	unit.ObjSubClass = 0
	update.MonsterDef, update.SoundSet122 = defs[slot], unsafe.Pointer(sounds[slot])
	update.Field120_1, update.Field120_2 = 1, 0
	want := 1
	switch parts[0] {
	case "frame-A":
	case "frame-B":
		update.Field120_1 = 2
	case "not-frame":
		update.Field120_1 = 3
		want = 0
	case "latched":
		update.Field120_2 = 1
		want = 0
	case "nil-sound":
		update.SoundSet122 = nil
		want = 0
	default:
		t.Fatal("MoveAudio branch")
	}
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	// Use the real native descriptor parser and audio event allocator. This
	// enables an existing named sound without substituting AudioEvent.
	if !s.Audio.ReadAVNT(&playerAttackHammerAudioFile538960{data: []byte{7, 1, 'x', 0, 0}}) {
		t.Fatal("native movement sound descriptor")
	}
	count := 0
	s.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _pos types.Pointf) {
		if id != sound.SoundHammerMissing || kind != 0 || owner != unit {
			t.Fatal("native movement sound owner/id/kind")
		}
		count++
	})
	beforeUnit, beforeUpdate, beforeHealth, beforeOther := *unit, *update, *unit.HealthData, *other
	beforeDefs := [2]server.MonsterDef{*defs[0], *defs[1]}
	beforeSounds := [2][32]uint32{*sounds[0], *sounds[1]}
	check := func() {
		t.Helper()
		queued := 0
		s.Audio.EachEvent(func(event *server.AudioEvent) {
			if event.Sound != sound.SoundHammerMissing || event.Obj != unit || event.Kind != 0 || event.Code != 0 || event.Pos != unit.PosVec {
				t.Fatal("real movement audio queue payload")
			}
			queued++
		})
		if count != want || queued != want {
			t.Fatalf("actual movement audio events=%d queued=%d want=%d", count, queued, want)
		}
		compare := func(got, want unsafe.Pointer, n uintptr) {
			t.Helper()
			if !bytes.Equal(unsafe.Slice((*byte)(got), int(n)), unsafe.Slice((*byte)(want), int(n))) {
				t.Fatal("whole native MoveAudio record changed")
			}
		}
		compare(unit.CObj(), unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit))
		compare(unsafe.Pointer(update), unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate))
		compare(unsafe.Pointer(unit.HealthData), unsafe.Pointer(&beforeHealth), unsafe.Sizeof(beforeHealth))
		compare(other.CObj(), unsafe.Pointer(&beforeOther), unsafe.Sizeof(beforeOther))
		for i := 0; i < 2; i++ {
			compare(unsafe.Pointer(defs[i]), unsafe.Pointer(&beforeDefs[i]), unsafe.Sizeof(beforeDefs[i]))
			compare(unsafe.Pointer(sounds[i]), unsafe.Pointer(&beforeSounds[i]), unsafe.Sizeof(beforeSounds[i]))
		}
	}
	count = 0
	s.Audio.Reset()
	s.MonsterMoveAudio534030(unit)
	check()
	t.Logf("MOVE AUDIO CONTROL COMPLETED: mode=%s unit=%p update=%p definition=%p sound=%p real-audio-events=%d", mode, unit, update, defs[slot], sounds[slot], count)
	count = 0
	s.Audio.Reset()
	ccall.CallVoidPtr(monsterMoveAudioPublicC534030(), unit.CObj())
	check()
	t.Logf("MOVE AUDIO NATIVE C PASS: mode=%s real-audio-events=%d full-records=true void-consumer=true", mode, count)
}
