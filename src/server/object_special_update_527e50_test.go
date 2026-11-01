package server

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestObjectSpecialNative527E50CachedIndexAndLiveFlags(t *testing.T) {
	p := &Player{PlayerInd: 3}
	d := &PlayerUpdateData{Player: p}
	r := &Object{UpdateData: unsafe.Pointer(d)} // Raw entry has no Player class gate.
	u := &Object{}
	u.Field140[3], u.Field140[31] = 0x100ab, 0x12345678
	var trace []string
	runtime := ObjectSpecialRuntime527E50{
		Animation: func(i int32, got *Object) int32 {
			if i != 3 || got != u {
				t.Fatal("animation arguments")
			}
			trace = append(trace, "animation")
			p.PlayerInd, r.UpdateData, d.Player = 31, nil, nil
			u.Field140[3] = 0x040300ab
			return -1
		},
		Enemy: func(gotR, gotU *Object) int32 {
			if gotR != r || gotU != u {
				t.Fatal("enemy arguments")
			}
			trace = append(trace, "enemy")
			return 0
		},
		Health: func(i int32, got *Object) int32 {
			if i != 3 || got != u {
				t.Fatal("health arguments")
			}
			trace = append(trace, "health")
			return -1
		},
		NPC: func(i int32, got *Object) int32 {
			if i != 3 || got != u {
				t.Fatal("NPC arguments")
			}
			trace = append(trace, "npc")
			return -1
		},
	}
	if got := ObjectSpecialUpdateNative527E50(r, u, runtime); got != 1 || !reflect.DeepEqual(trace, []string{"animation", "enemy", "health", "npc"}) || u.Field140[3] != 0xab || u.Field140[31] != 0x12345678 {
		t.Fatal("cached index/live flags/post-callback clears", got, trace, u.Field140)
	}
}

func TestObjectSpecialNative527E50EnemyStillClearsHealth(t *testing.T) {
	p := &Player{PlayerInd: 31}
	d := &PlayerUpdateData{Player: p}
	r := &Object{UpdateData: unsafe.Pointer(d)}
	u := &Object{}
	u.Field140[31] = 0x200ab
	got := ObjectSpecialUpdateNative527E50(r, u, ObjectSpecialRuntime527E50{Enemy: func(*Object, *Object) int32 { u.Field140[31] = 0x092200ab; return -2147483648 }, Health: func(int32, *Object) int32 { t.Fatal("enemy health must not be reported"); return 0 }})
	if got != 1 || u.Field140[31] != 0x092000ab {
		t.Fatal("enemy health clear and unhandled bits", got, u.Field140[31])
	}
}

func TestObjectSpecialNative527E50AdmissionAndNilUnit(t *testing.T) {
	p := &Player{PlayerInd: 3}
	d := &PlayerUpdateData{Player: p}
	r := &Object{UpdateData: unsafe.Pointer(d)}
	u := &Object{}
	u.Field140[3] = 0xf000abcd
	if ObjectSpecialUpdateNative527E50(r, u, ObjectSpecialRuntime527E50{}) != 0 || u.Field140[3] != 0xf000abcd {
		t.Fatal("outside-mask admission")
	}
	u.Field140[3] = 0x001000ab
	if ObjectSpecialUpdateNative527E50(r, u, ObjectSpecialRuntime527E50{}) != 1 || u.Field140[3] != 0x001000ab {
		t.Fatal("unhandled admitted bits")
	}
	if ObjectSpecialUpdateNative527E50(r, nil, ObjectSpecialRuntime527E50{}) != 1 {
		t.Fatal("nil unit return")
	}
	p.PlayerInd = 255
	if ObjectSpecialUpdateNative527E50(r, u, ObjectSpecialRuntime527E50{}) != 1 || u.Field140[3] != 0x001000ab {
		t.Fatal("unsigned byte index gate")
	}
}

func TestObjectSpecialNative527E50RawNilChainBeforeUnitGate(t *testing.T) {
	for _, nilUnit := range []bool{false, true} {
		for _, kind := range []string{"recipient", "update", "player"} {
			name := kind + "/unit"
			if nilUnit {
				name = kind + "/nil-unit"
			}
			t.Run(name, func(t *testing.T) {
				r := &Object{}
				u := &Object{}
				if nilUnit {
					u = nil
				}
				if kind == "recipient" {
					r = nil
				} else if kind == "player" {
					r.UpdateData = unsafe.Pointer(&PlayerUpdateData{})
				}
				defer func() {
					if recover() == nil {
						t.Fatal("raw nil-chain fault must precede nil-unit gate")
					}
				}()
				ObjectSpecialUpdateNative527E50(r, u, ObjectSpecialRuntime527E50{})
			})
		}
	}
}

func TestObjectSpecialNative527E50FieldWidthsAndOffsets(t *testing.T) {
	wantFlags, wantUpdate, wantPlayer := uintptr(560), uintptr(748), uintptr(276)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFlags, wantUpdate, wantPlayer = 624, 872, 336
	}
	if unsafe.Offsetof(Object{}.Field140) != wantFlags || unsafe.Sizeof(Object{}.Field140) != 128 || unsafe.Offsetof(Object{}.UpdateData) != wantUpdate || unsafe.Offsetof(PlayerUpdateData{}.Player) != wantPlayer || unsafe.Sizeof(Player{}.PlayerInd) != 1 {
		t.Fatal("native layout contract")
	}
}
