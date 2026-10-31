package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// No allocation or trace function is supplied: the ordinary public action
// must use its real type registry, object allocator and empty wall map.
func TestMonsterMissileNativeDefaultBindings532540(t *testing.T) {
	for _, mode := range []string{"clear-map", "outside-map", "missing-type"} {
		t.Run(mode, func(t *testing.T) {
			s := New(nil, nil, strman.New())
			t.Cleanup(s.Close)
			d, freeD := alloc.New(MonsterUpdateData{})
			t.Cleanup(freeD)
			definition, freeDef := alloc.New(MonsterDef{})
			t.Cleanup(freeDef)
			sounds, freeSounds := alloc.New([17]uint32{})
			t.Cleanup(freeSounds)
			definition.MissileAttackFrame216 = 2
			sounds[11] = 1111
			if !s.Objs.Init(3) {
				t.Fatal("ordinary object allocator initialization")
			}
			t.Cleanup(s.Objs.FreeObjects)
			s.Map.Init()
			t.Cleanup(s.Map.Free)
			s.Walls.byPos = make([]*Wall, wallsPerBucket*WallGridSize)
			monsterType := &ObjectType{s: &s.Types, ind: 1, id: "NativeMissileOwner", class: object.ClassMonster, flags: object.FlagNoCollide}
			missileType := &ObjectType{s: &s.Types, ind: 2, id: "NativeMissileProjectile", class: object.ClassMissile, flags: object.FlagNoCollide, Speed: 12}
			s.Types.byID = map[string]*ObjectType{"nativemissileowner": monsterType, "nativemissileprojectile": missileType}
			s.Types.byInd = []*ObjectType{nil, monsterType, missileType}
			u := s.NewObjectByTypeID("NativeMissileOwner")
			if u == nil {
				t.Fatal("ordinary owner factory")
			}
			u.UpdateData = unsafe.Pointer(d)
			u.PosVec = types.Ptf(300, 300)
			u.Shape.Circle.R = 10
			u.Direction1 = 0
			if mode == "outside-map" {
				u.PosVec.X = -50
			}
			d.MonsterDef = definition
			d.SoundSet122 = unsafe.Pointer(&sounds[0])
			d.Field120_1 = 2
			d.AIStackInd = 0
			d.AIStack[0].Action = uint32(ai.ACTION_MISSILE_ATTACK)
			d.AIStack[0].SetArgs(types.Ptf(u.PosVec.X+30, 340))
			copy(definition.MissileName148[:], "NativeMissileProjectile")
			if mode == "missing-type" {
				clear(definition.MissileName148[:])
				copy(definition.MissileName148[:], "MissingNativeMissile")
			}
			// Confirm the real map's clear/blocked outcomes independently, before
			// observing what the native action chooses to create or delete.
			if got := s.MapTraceRay(u.PosVec, types.Ptf(u.PosVec.X+20, 304), 5); got != (mode != "outside-map") {
				t.Fatal("empty-map trace control")
			}
			before := s.Objs.Alive
			var created, deleted *Object
			audio := 0
			handled := s.MonsterActionMissileUpdate532610(u, MonsterActionMissileRuntime532540{
				AudioEvent: func(id uint32, owner *Object) {
					if id != 1111 || owner != u {
						t.Fatal("ordinary action audio binding")
					}
					audio++
				},
				CreateObjectAt: func(o, owner *Object, pos types.Pointf) {
					if created != nil || owner != u || pos != types.Ptf(u.PosVec.X+14, 300) {
						t.Fatal("ordinary creation binding")
					}
					created = o
				},
				DelayedDelete: func(o *Object) {
					if deleted != nil {
						t.Fatal("duplicate deletion")
					}
					deleted = o
				},
			})
			if !handled || audio != 1 {
				t.Fatalf("ordinary action completion/audio: handled=%v audio=%d class=%#x subclass=%#x UD=%p d=%p sound=%p", handled, audio, u.ObjClass, u.ObjSubClass, u.UpdateData, d, d.SoundSet122)
			}
			if mode == "missing-type" {
				if created != nil || deleted != nil || s.Objs.Alive != before {
					t.Fatal("missing registered type fabricated an object")
				}
				return
			}
			o := created
			if mode == "outside-map" {
				o = deleted
				if created != nil {
					t.Fatal("blocked ordinary trace created a projectile")
				}
			} else if deleted != nil {
				t.Fatal("clear ordinary trace deleted a projectile")
			}
			if o == nil || o.Server() != s || o.TypeInd != 2 || o.Class() != object.ClassMissile || o.SpeedCur != 12 || s.Objs.Alive != before+1 {
				t.Fatal("ordinary registered factory result differs")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(o)) <= math.MaxUint32 {
				t.Fatal("ordinary projectile pointer below4GiB")
			}
			if mode == "clear-map" && (o.VelVec.X <= 0 || o.VelVec.Y <= 0 || o.Direction1 != u.Direction1 || o.Direction2 != u.Direction1) {
				t.Fatal("ordinary created projectile final trajectory/directions")
			}
		})
	}
}
