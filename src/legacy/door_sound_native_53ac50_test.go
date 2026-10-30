package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type doorSoundFile53AC50 struct {
	playerAttackHammerAudioFile538960
	name string
}

func (f *doorSoundFile53AC50) ReadString8() (string, error) { return f.name, nil }

// Audio allocation, owner/position and the actual public C door body run;
// playback is covered separately with the product's real audio backend.
func TestDoorSoundActualNativeC53AC50(t *testing.T) {
	for _, tc := range []struct {
		subclass    uint32
		material    uint16
		open, close sound.ID
	}{
		{0, 0, sound.SoundOpenWoodenDoor, sound.SoundCloseWoodenDoor},
		{4, 0, sound.SoundOpenGate, sound.SoundCloseGate},
		{4, 8, sound.SoundOpenWoodenGate, sound.SoundCloseWoodenGate},
		{1, 0, sound.SoundOpenHeavyWoodenDoor, sound.SoundCloseHeavyWoodenDoor},
		{0x1000, 0, sound.SoundStoneDoorOpen, sound.SoundStoneDoorClose},
	} {
		for _, reached := range []bool{false, true} {
			t.Run(fmt.Sprintf("%x:%x:%t", tc.subclass, tc.material, reached), func(t *testing.T) {
				s, door, _, _ := monsterMainBlockLegacyFixture547210(t)
				update, free := alloc.New(server.DoorUpdateData{})
				t.Cleanup(free)
				*update = server.DoorUpdateData{TargetDirection: 3, SyncedDirection: 3, CurrentDirection: 4}
				if reached {
					update.SyncedDirection, update.CurrentDirection = 2, 3
				}
				door.ObjClass, door.ObjSubClass, door.Material = object.ClassDoor, object.SubClass(tc.subclass), tc.material
				door.ObjFlags &^= object.Flags(0x1000000)
				door.UpdateData = unsafe.Pointer(update)
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(door.CObj()) <= 0xffffffff || uintptr(door.UpdateData) <= 0xffffffff) {
					t.Fatal("actual C door records must exceed4GiB")
				}
				s.Audio.Init(s)
				t.Cleanup(s.Audio.Free)
				for _, id := range []sound.ID{tc.open, tc.close} {
					if !s.Audio.ReadAVNT(&doorSoundFile53AC50{name: id.String(), playerAttackHammerAudioFile538960: playerAttackHammerAudioFile538960{data: []byte{7, 1, 'x', 0, 0}}}) {
						t.Fatal("native door sound descriptor")
					}
				}
				want := tc.open
				if reached {
					want = tc.close
				}
				calls := 0
				s.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, pos types.Pointf) {
					if id != want || kind != 0 || owner != door || pos != door.PosVec {
						t.Errorf("actual C door sound=%s kind=%d owner=%p, want=%s/%p", id, kind, owner, want, door)
					}
					calls++
				})
				s.Audio.Reset()
				ccall.CallVoidPtr(doorSoundPublicC53AC50(), door.CObj())
				queued := 0
				s.Audio.EachEvent(func(event *server.AudioEvent) {
					if event.Sound != want || event.Obj != door || event.Kind != 0 || event.Code != 0 || event.Pos != door.PosVec {
						t.Errorf("native door queue sound=%s owner=%p want=%s/%p", event.Sound, event.Obj, want, door)
					}
					queued++
				})
				if calls != 1 || queued != 1 || update.SyncedDirection != update.CurrentDirection {
					t.Fatal("native door queue/synchronization count")
				}
			})
		}
	}
}
