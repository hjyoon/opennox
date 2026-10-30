package server

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

// These expected sound IDs come from unchanged GAME.EXE selector blocks,
// not from a second Go copy of the selection logic.
func TestDoorSoundOriginal53AC50(t *testing.T) {
	blob, err := os.ReadFile("testdata/door_sound_original_53ac50.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference struct {
		OriginalSHA256 string `json:"original_sha256"`
		Rows           []struct {
			Input struct {
				Cls, Subclass uint32
				Material      uint16
				Reached       bool
			}
			Sound uint32
		}
	}
	if err := json.Unmarshal(blob, &reference); err != nil {
		t.Fatal(err)
	}
	if reference.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(reference.Rows) != 384 {
		t.Fatal("unsealed original door sound reference")
	}
	for i, row := range reference.Rows {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			c := row.Input
			door, objectCheck := retreatOriginalAlloc545440[Object](t)
			update, updateCheck := retreatOriginalAlloc545440[DoorUpdateData](t)
			door.ObjClass, door.ObjSubClass = object.Class(c.Cls), object.SubClass(c.Subclass)
			door.Material, door.UpdateData = c.Material, unsafe.Pointer(update)
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(door.CObj()) <= 0xffffffff || uintptr(door.UpdateData) <= 0xffffffff) {
				t.Fatal("door and update must use native addresses above4GiB")
			}
			if got := doorUpdateSound53AC50(door, c.Reached); got != row.Sound {
				t.Fatalf("original selector: class=%#x subclass=%#x material=%#x reached=%t got=%d original=%d", c.Cls, c.Subclass, c.Material, c.Reached, got, row.Sound)
			}
			*update = DoorUpdateData{TargetDirection: 3, SyncedDirection: 3, CurrentDirection: 4}
			var s Server
			if c.Reached {
				update.SyncedDirection, update.CurrentDirection = 2, 3
				s.Objs.AddToUpdatable(door)
			}
			calls := 0
			s.DoorUpdate53AC50(door, DoorUpdateRuntime53AC50{AudioEvent: func(id uint32, owner *Object) {
				if id != row.Sound || owner != door {
					t.Fatalf("full door update submitted sound=%d owner=%p, want=%d/%p", id, owner, row.Sound, door)
				}
				calls++
			}})
			if calls != 1 || update.SyncedDirection != update.CurrentDirection || door.Field38 != ^uint32(0) || door.IsUpdatable != 0 {
				t.Fatal("original sound phase did not retain door synchronization/removal")
			}
			if door.ObjClass != object.Class(c.Cls) || door.ObjSubClass != object.SubClass(c.Subclass) || door.Material != c.Material || door.UpdateData != unsafe.Pointer(update) {
				t.Fatal("door sound changed input type/material/pointer fields")
			}
			objectCheck()
			updateCheck()
		})
	}
}
