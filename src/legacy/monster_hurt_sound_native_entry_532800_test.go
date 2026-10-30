package legacy

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterHurtSoundNextPublicCNative532800(t *testing.T) {
	const child = "NOX_MONSTER_HURT_SOUND_NATIVE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		for _, kind := range []string{"due", "before-deadline", "nil-sound", "non-monster", "frame-wrap"} {
			for slot := 0; slot < 2; slot++ {
				mode := fmt.Sprintf("%s:%d", kind, slot)
				t.Run(mode, func(t *testing.T) {
					bin, err := os.Executable()
					if err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(bin, "-test.run=^TestMonsterHurtSoundNextPublicCNative532800$", "-test.v")
					cmd.Env = append(os.Environ(), child+"="+mode)
					out, err := cmd.CombinedOutput()
					t.Logf("%s", out)
					if err != nil {
						t.Fatalf("actual native HurtSound entry: %v", err)
					}
				})
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 2 {
		t.Fatal("HurtSound mode")
	}
	slot, err := strconv.Atoi(parts[1])
	if err != nil || slot < 0 || slot > 1 {
		t.Fatal("HurtSound native slot")
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
		sounds[i][2] = uint32(sound.SoundHammerMissing)
		updates[i].SoundSet122 = unsafe.Pointer(sounds[i])
		updates[i].Field133 = 0xcafe1313
	}
	update := updates[slot]
	unit.UpdateData = unsafe.Pointer(update)
	s.SetTickRate(30)
	frame, deadline, wantSounds, due := uint32(1400), uint32(1400), 1, true
	switch parts[0] {
	case "due":
	case "before-deadline":
		deadline, wantSounds, due = 1401, 0, false
	case "nil-sound":
		update.SoundSet122, wantSounds = nil, 0
	case "non-monster":
		unit.ObjClass, wantSounds, due = object.ClassPlayer, 0, false
	case "frame-wrap":
		frame, deadline = 0xfffffff0, 0xffffffe0
	default:
		t.Fatal("HurtSound branch")
	}
	s.SetFrame(frame)
	update.Field133 = deadline
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(first), unsafe.Pointer(second), unsafe.Pointer(sounds[0]), unsafe.Pointer(sounds[1])} {
			if uintptr(p) <= 0xffffffff {
				t.Fatal("native HurtSound records must exceed4GiB")
			}
		}
	}
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	if !s.Audio.ReadAVNT(&playerAttackHammerAudioFile538960{data: []byte{7, 1, 'x', 0, 0}}) {
		t.Fatal("native hurt sound descriptor")
	}
	count := 0
	s.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, pos types.Pointf) {
		if id != sound.SoundHammerMissing || kind != 0 || owner != unit || pos != unit.PosVec {
			t.Fatal("native hurt sound owner/id/kind/position")
		}
		count++
	})
	const seed = 731
	rng := prand.New(seed)
	wantDeadline := deadline
	if due {
		wantDeadline = frame + uint32(rng.IntClamp(60, 120))
	}
	wantNext := rng.IntClamp(0, math.MaxInt32)
	beforeUnit, beforeHealth, beforeOther := *unit, *unit.HealthData, *other
	beforeUpdates := [2]server.MonsterUpdateData{*first, *second}
	beforeSounds := [2][32]uint32{*sounds[0], *sounds[1]}
	beforeUpdates[slot].Field133 = wantDeadline
	check := func() {
		t.Helper()
		queued := 0
		s.Audio.EachEvent(func(event *server.AudioEvent) {
			if event.Sound != sound.SoundHammerMissing || event.Obj != unit || event.Kind != 0 || event.Code != 0 || event.Pos != unit.PosVec {
				t.Fatal("real hurt audio queue payload")
			}
			queued++
		})
		if count != wantSounds || queued != wantSounds || update.Field133 != wantDeadline {
			t.Fatalf("actual hurt audio events=%d queued=%d deadline=%#x want=%d/%#x", count, queued, update.Field133, wantSounds, wantDeadline)
		}
		if next := s.Rand.Logic.IntClamp(0, math.MaxInt32); next != wantNext {
			t.Fatalf("real hurt RNG next=%d want=%d", next, wantNext)
		}
		compare := func(got, want unsafe.Pointer, n uintptr) {
			t.Helper()
			if !bytes.Equal(unsafe.Slice((*byte)(got), int(n)), unsafe.Slice((*byte)(want), int(n))) {
				t.Fatal("whole native HurtSound record differs")
			}
		}
		compare(unit.CObj(), unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit))
		compare(unsafe.Pointer(unit.HealthData), unsafe.Pointer(&beforeHealth), unsafe.Sizeof(beforeHealth))
		compare(other.CObj(), unsafe.Pointer(&beforeOther), unsafe.Sizeof(beforeOther))
		for i := 0; i < 2; i++ {
			compare(unsafe.Pointer(updates[i]), unsafe.Pointer(&beforeUpdates[i]), unsafe.Sizeof(beforeUpdates[i]))
			compare(unsafe.Pointer(sounds[i]), unsafe.Pointer(&beforeSounds[i]), unsafe.Sizeof(beforeSounds[i]))
		}
	}
	count = 0
	s.Audio.Reset()
	s.Rand.Logic = prand.New(seed)
	s.MonsterHurtSound532800(unit)
	check()
	t.Logf("HURT SOUND CONTROL COMPLETED: mode=%s unit=%p update=%p sound=%p deadline=%#x real-audio-events=%d", mode, unit, update, update.SoundSet122, update.Field133, count)
	count = 0
	update.Field133 = deadline
	s.Audio.Reset()
	s.Rand.Logic = prand.New(seed)
	ccall.CallVoidPtr(monsterHurtSoundPublicC532800(), unit.CObj())
	check()
	t.Logf("HURT SOUND NATIVE C PASS: mode=%s real-audio-events=%d full-records=true RNG-next=true void-consumer=true", mode, count)
}
