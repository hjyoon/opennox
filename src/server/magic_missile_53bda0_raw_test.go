package server

import (
	"testing"
	"unsafe"
)

func TestMagicMissileRawBoundaries53BDA0(t *testing.T) {
	for _, name := range []string{"root-nil", "update-nil", "owner-nil", "destroyed-owner-nil-collision", "expired-nil-collision", "destroyed-owner-valid-collision", "dead-owner-nil-collision"} {
		t.Run(name, func(t *testing.T) {
			owner := new(Object)
			update := &MissileUpdateData{Owner: owner}
			missile := &Object{UpdateData: unsafe.Pointer(update)}
			rt := MagicMissileUpdateRuntime53BDA0{Frame: func() uint32 { return 10 }, FPS: func() uint32 { return 30 }}
			mustFault := true
			switch name {
			case "root-nil":
				missile = nil
				MagicMissileUpdate53BDA0(nil, MagicMissileUpdateRuntime53BDA0{})
			case "update-nil":
				missile.UpdateData = nil
				MagicMissileUpdate53BDA0(missile, MagicMissileUpdateRuntime53BDA0{})
			case "owner-nil":
				update.Owner = nil
			case "destroyed-owner-nil-collision":
				owner.ObjFlags = 0x20
			case "expired-nil-collision":
				rt.Frame = func() uint32 { return 1000 }
			case "destroyed-owner-valid-collision":
				owner.ObjFlags = 0x20
				mustFault = false
				rt.Frame = func() uint32 { t.Fatal("original destroyed-owner branch must precede frame/FPS reads"); return 0 }
				rt.CollideResult = func(m *Object) int32 {
					if m != missile {
						t.Fatal("collision object pointer")
					}
					return -1985229329
				}
			case "dead-owner-nil-collision":
				owner.ObjFlags = 0x8000
				mustFault = false
			}
			if mustFault {
				defer func() {
					if recover() == nil {
						t.Fatal("original required pointer did not fault")
					}
				}()
			}
			result := MagicMissileUpdateRaw53BDA0(missile, rt)
			if name == "destroyed-owner-valid-collision" && uint32(result) != 0x89abcdef {
				t.Fatalf("collision result %#x", uint32(result))
			}
			if name == "dead-owner-nil-collision" && result != 0 {
				t.Fatalf("original force index%d", result)
			}
		})
	}
}
