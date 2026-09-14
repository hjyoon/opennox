package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestForceOfNatureCreate52EF30KeepsNativeChargeAndLowDirectionWord(t *testing.T) {
	data := &WandUseData{Flags: 5}
	wand := &Object{ObjSubClass: object.SubClass(0x200000), UseData: UseDataPtr{Ptr: unsafe.Pointer(data)}}
	update := &PlayerUpdateData{EquippedWeapon: wand}
	caster := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update),
		PosVec: types.Pointf{X: 123, Y: 456}, Direction1: 0x12a5}
	charge := &Object{}
	record := &DurSpell{Caster16: caster, Field72: int32(0x5a5a00ff), Field76: uintptr(0x3fdccccc)}
	var stored *Object
	var created bool
	runtime := SpellForceOfNatureRuntime52EF30{
		NewObject: func(id string) *Object {
			if id != "ForceOfNatureCharge" {
				t.Fatalf("object ID = %q", id)
			}
			return charge
		},
		CreateAt: func(got, owner *Object, pos types.Pointf) {
			created = true
			if got != charge || owner != nil || pos != caster.PosVec {
				t.Fatalf("create = %p/%p/%v", got, owner, pos)
			}
		},
		StoreCharge: func(got *DurSpell, visual *Object) {
			if got != record {
				t.Fatalf("record = %p", got)
			}
			stored = visual
		},
	}
	if got := SpellForceOfNatureCreate52EF30(record, runtime); got != 0 {
		t.Fatalf("create result = %d", got)
	}
	if stored != charge || !created || record.Field72 != int32(0x5a5a12a5) || record.Flags88 != 2 ||
		record.Field76 != uintptr(0x3fdccccc) {
		t.Fatalf("create state = %p/%v/%#x/%#x/%#x", stored, created, record.Field72, record.Flags88, record.Field76)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(charge)) <= uintptr(^uint32(0)) {
		t.Fatal("expected charge pointer above 4 GiB")
	}
}

func TestForceOfNatureCreate52EF30AnchoredModeAndMissingCharge(t *testing.T) {
	anchor := &Object{PosVec: types.Pointf{X: 30, Y: 40}, Direction1: 0x01b4}
	record := &DurSpell{Flag20: 1, Obj24: anchor, Field72: int32(0x42420000)}
	runtime := SpellForceOfNatureRuntime52EF30{
		NewObject:   func(id string) *Object { return nil },
		CreateAt:    func(*Object, *Object, types.Pointf) { t.Fatal("created missing charge") },
		StoreCharge: func(*DurSpell, *Object) { t.Fatal("stored missing charge") },
	}
	if got := SpellForceOfNatureCreate52EF30(record, runtime); got != 0 || record.Field72 != int32(0x424201b4) {
		t.Fatalf("result/direction = %d/%#x", got, record.Field72)
	}
}

func TestForceOfNatureUpdate52EFD0ChargeAndPlayerState(t *testing.T) {
	caster := &Object{ObjClass: object.ClassPlayer}
	charge := &Object{}
	record := &DurSpell{Caster16: caster, Frame68: 20, Field76: uintptr(0x3fdccccc)}
	frame := uint32(12)
	var stored = charge
	var deleted, states int
	runtime := SpellForceOfNatureRuntime52EF30{
		CurrentFrame: func() uint32 { return frame },
		LoadCharge:   func(*DurSpell) *Object { return stored },
		StoreCharge:  func(_ *DurSpell, visual *Object) { stored = visual },
		DelayedDelete: func(got *Object) {
			deleted++
			if got != charge {
				t.Fatalf("deleted = %p", got)
			}
		},
		SetPlayerState: func(got *Object, state PlayerState) {
			states++
			if got != caster || state != 10 {
				t.Fatalf("state = %p/%d", got, state)
			}
		},
		NewObject: func(string) *Object { t.Fatal("early projectile"); return nil },
	}
	if got := SpellForceOfNatureUpdate52EFD0(record, runtime); got != 0 || deleted != 0 || states != 1 {
		t.Fatalf("before charge expiry = %d/%d/%d", got, deleted, states)
	}
	frame = 13
	if got := SpellForceOfNatureUpdate52EFD0(record, runtime); got != 0 || deleted != 1 || stored != nil || states != 2 {
		t.Fatalf("charge expiry = %d/%d/%p/%d", got, deleted, stored, states)
	}
	if record.Field76 != uintptr(0x3fdccccc) {
		t.Fatalf("PE32 slot changed: %#x", record.Field76)
	}
}

func TestForceOfNatureUpdate52EFD0LaunchNativeDeathBall(t *testing.T) {
	for _, tc := range []struct {
		name     string
		shape    Shape
		distance float32
		trace    bool
	}{
		{"center", Shape{Kind: ShapeKindCenter}, 4, true},
		{"circle", Shape{Kind: ShapeKindCircle, Circle: struct{ R, R2 float32 }{R: 9}}, 13, true},
		{"box", Shape{Kind: ShapeKindBox, Box: ShapeBox{W: 6, H: 10}}, 14, true},
		{"fallback", Shape{}, 24, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caster := &Object{PosVec: types.Pointf{X: 100, Y: 200}, Direction1: 64, Shape: tc.shape}
			ball := &Object{SpeedCur: 3}
			record := &DurSpell{Caster16: caster, Frame68: 20, Field72: int32(0x3fdccccc)}
			vector := caster.Direction1.Vec()
			want := types.Pointf{X: caster.PosVec.X + vector.X*tc.distance, Y: caster.PosVec.Y + vector.Y*tc.distance}
			events := 0
			runtime := SpellForceOfNatureRuntime52EF30{
				CurrentFrame: func() uint32 { return 19 },
				NewObject: func(id string) *Object {
					if id != "DeathBall" {
						t.Fatalf("object ID = %q", id)
					}
					return ball
				},
				TraceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
					if from != caster.PosVec || to != want || flags != MapTraceFlag1|MapTraceFlag3 {
						t.Fatalf("trace = %v/%v/%#x", from, to, flags)
					}
					return tc.trace
				},
				CreateAt: func(got, owner *Object, point types.Pointf) {
					if !tc.trace {
						want = caster.PosVec
					}
					if got != ball || owner != caster || point != want {
						t.Fatalf("create = %p/%p/%v, want %v", got, owner, point, want)
					}
				},
				EventObj: func(got *Object) {
					events++
					if got != caster {
						t.Fatalf("audio owner = %p", got)
					}
				},
			}
			if got := SpellForceOfNatureUpdate52EFD0(record, runtime); got != 1 {
				t.Fatalf("update = %d", got)
			}
			if ball.VelVec != (types.Pointf{X: vector.X * 3, Y: vector.Y * 3}) ||
				ball.Direction1 != caster.Direction1 || ball.Direction2 != caster.Direction1 || events != 1 {
				t.Fatalf("projectile = %v/%d/%d, events=%d", ball.VelVec, ball.Direction1, ball.Direction2, events)
			}
		})
	}
}

func TestForceOfNatureUpdate52EFD0AnchoredModeAndBuffCancel(t *testing.T) {
	ball := &Object{VelVec: types.Pointf{X: 50, Y: 50}, SpeedCur: 9}
	record := &DurSpell{Flag20: 1, Frame68: 10, Pos: types.Pointf{X: 300, Y: 400}, Field72: 0x1c}
	var created bool
	runtime := SpellForceOfNatureRuntime52EF30{
		CurrentFrame: func() uint32 { return 9 },
		NewObject:    func(id string) *Object { return ball },
		TraceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			if from != record.Pos || to != record.Pos {
				t.Fatalf("anchored trace = %v/%v", from, to)
			}
			return true
		},
		CreateAt: func(got, owner *Object, point types.Pointf) {
			created = true
			if got != ball || owner != nil || point != record.Pos {
				t.Fatalf("create = %p/%p/%v", got, owner, point)
			}
		},
		EventObj: func(got *Object) {
			if got != nil {
				t.Fatalf("event owner = %p", got)
			}
		},
	}
	if got := SpellForceOfNatureUpdate52EFD0(record, runtime); got != 1 || !created || ball.VelVec != (types.Pointf{}) || ball.Direction1 != 0x1c || ball.Direction2 != 0x1c {
		t.Fatalf("anchored launch = %d/%v/%v/%d/%d", got, created, ball.VelVec, ball.Direction1, ball.Direction2)
	}
	// Buff 8 stops the callback before frame and charge access.
	caster := &Object{Buffs: 1 << 8}
	record.Caster16 = caster
	runtime.CurrentFrame = func() uint32 { t.Fatal("frame read after buff"); return 0 }
	if got := SpellForceOfNatureUpdate52EFD0(record, runtime); got != 1 {
		t.Fatalf("buff cancellation = %d", got)
	}
}

func TestForceOfNatureDestroy52F1D0ClearsChargeAndWandFlag(t *testing.T) {
	data := &WandUseData{Flags: 5}
	wand := &Object{ObjSubClass: object.SubClass(0x200000), UseData: UseDataPtr{Ptr: unsafe.Pointer(data)}}
	update := &PlayerUpdateData{EquippedWeapon: wand}
	caster := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	charge := &Object{}
	record := &DurSpell{Caster16: caster, Field76: uintptr(0x3fdccccc)}
	var stored = charge
	deleted := 0
	runtime := SpellForceOfNatureRuntime52EF30{
		LoadCharge:  func(*DurSpell) *Object { return stored },
		StoreCharge: func(_ *DurSpell, got *Object) { stored = got },
		DelayedDelete: func(got *Object) {
			deleted++
			if got != charge {
				t.Fatalf("deleted = %p", got)
			}
		},
	}
	SpellForceOfNatureDestroy52F1D0(record, runtime)
	if deleted != 1 || stored != nil || data.Flags != 1 || record.Field76 != uintptr(0x3fdccccc) {
		t.Fatalf("destroy = %d/%p/%#x/%#x", deleted, stored, data.Flags, record.Field76)
	}
}
