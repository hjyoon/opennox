package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func installPlayerAttackProjectileRuntime538960(
	t *testing.T, deps playerAttackProjectileRuntime538960,
) {
	t.Helper()
	old := playerAttackProjectileRuntimeFactory538960
	playerAttackProjectileRuntimeFactory538960 = func() playerAttackProjectileRuntime538960 {
		return deps
	}
	t.Cleanup(func() { playerAttackProjectileRuntimeFactory538960 = old })
}

func installPlayerAttackProjectileServer538960(t *testing.T) *server.Server {
	t.Helper()
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	old := GetServer
	GetServer = func() Server { return &playerAttackLegacyServer538960{srv: srv} }
	t.Cleanup(func() { GetServer = old })
	return srv
}

func pinPlayerAttackProjectilePointers538960(
	t *testing.T, pin *runtime.Pinner, pointers ...unsafe.Pointer,
) {
	t.Helper()
	for index, pointer := range pointers {
		pin.Pin(pointer)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native projectile pointer %d = %p, want address above ABI32", index, pointer)
		}
	}
}

func TestPlayerAttackProjectileApplyEffects539F40MatchesOriginalSlots(t *testing.T) {
	recoilIdentity := unsafe.Pointer(uintptr(0x1234))
	speedIdentity := unsafe.Pointer(uintptr(0x5678))
	ignored := &server.ModifierEff{Attack40: server.ModifierEffFnc{Fnc: speedIdentity}}
	recoil := &server.ModifierEff{AttackPreHit52: server.ModifierEffFnc{Fnc: recoilIdentity}}
	speed := &server.ModifierEff{Attack40: server.ModifierEffFnc{Fnc: speedIdentity}}
	weaponData := &server.ModifierInitData{Modifiers: [4]*server.ModifierEff{
		ignored, nil, recoil, speed,
	}}
	projectileData := &server.ModifierInitData{}
	owner := &server.Object{}
	weapon := &server.Object{InitData: unsafe.Pointer(weaponData)}
	projectile := &server.Object{InitData: unsafe.Pointer(projectileData)}

	var speedCalls int
	playerAttackProjectileApplyEffects539F40(owner, weapon, projectile,
		playerAttackProjectileRuntime538960{
			RecoilEffect:          recoilIdentity,
			ProjectileSpeedEffect: speedIdentity,
			ApplyProjectileSpeed: func(gotEffect *server.ModifierEff, gotProjectile *server.Object) {
				speedCalls++
				if gotEffect != speed || gotProjectile != projectile {
					t.Fatalf("speed effect = %p/%p, want %p/%p", gotEffect, gotProjectile, speed, projectile)
				}
			},
		})

	if projectileData.Modifiers[3] != recoil {
		t.Fatalf("projectile recoil modifier = %p, want %p", projectileData.Modifiers[3], recoil)
	}
	if speedCalls != 1 {
		t.Fatalf("projectile speed calls = %d, want 1", speedCalls)
	}
}

func TestPlayerAttackExport538960LaunchesNPCRoundChakramWithNativePointers(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}
	srv := installPlayerAttackProjectileServer538960(t)

	unit := &server.Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterNPC),
		PosVec:      types.Ptf(100, 200),
		NewPos:      types.Ptf(100, 200),
		Direction1:  17,
	}
	unit.Shape.Kind = server.ShapeKindCircle
	unit.Shape.Circle.R = 6
	update := &server.MonsterUpdateData{
		Field331:         25,
		Field481:         0xaabbcc01,
		WeaponEquipFlags: uint32(object.WeaponChakram),
		Field517:         0x11223344,
	}
	weaponAttrs := &server.ModifierInitData{}
	weapon := &server.Object{
		TypeInd:     0x3101,
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponChakram),
		ObjFlags:    object.FlagEquipped,
		InvHolder:   unit,
		InitData:    unsafe.Pointer(weaponAttrs),
	}
	modifier := &server.Modifier{TypeInd: uint32(weapon.TypeInd)}
	unit.UpdateData = unsafe.Pointer(update)
	unit.InvFirstItem = weapon
	srv.Modif.Dword_5d4594_251600 = modifier

	projectileData := &server.ChakramUpdateData{}
	projectileAttrs := &server.ModifierInitData{}
	projectile := &server.Object{
		SpeedCur:   7,
		UpdateData: unsafe.Pointer(projectileData),
		InitData:   unsafe.Pointer(projectileAttrs),
	}
	spawn := playerAttackProjectileSpawnPoint538960(unit)
	var traceFlags server.MapTraceFlags
	var created, detached, inserted, attributes bool
	var audio sound.ID
	installPlayerAttackProjectileRuntime538960(t, playerAttackProjectileRuntime538960{
		Frame: func() uint32 { return 2 },
		AnimFrames: func(action int) (int, int) {
			if action != 44 {
				t.Fatalf("round chakram animation = %d, want 44", action)
			}
			return 4, 0
		},
		Trace: func(from, to types.Pointf, flags server.MapTraceFlags) bool {
			if from != unit.PosVec || to != spawn {
				t.Fatalf("round chakram trace = %+v -> %+v, want %+v -> %+v", from, to, unit.PosVec, spawn)
			}
			traceFlags = flags
			return true
		},
		NewObject: func(name string) *server.Object {
			if name != "RoundChakramInMotion" {
				t.Fatalf("round chakram type = %q", name)
			}
			return projectile
		},
		CreateAt: func(gotProjectile, gotOwner *server.Object, pos types.Pointf) {
			created = gotProjectile == projectile && gotOwner == unit && pos == spawn
		},
		ApplyModifierAttrs: func(gotProjectile *server.Object, attrs *server.ModifierInitData) {
			attributes = gotProjectile == projectile && attrs == weaponAttrs
		},
		DetachInventory: func(gotOwner, gotWeapon *server.Object) {
			detached = gotOwner == unit && gotWeapon == weapon
		},
		InventoryPut: func(gotProjectile, gotWeapon *server.Object, report bool) {
			inserted = gotProjectile == projectile && gotWeapon == weapon && report
		},
		AudioEvent: func(id sound.ID, gotOwner *server.Object) {
			if gotOwner != unit {
				t.Fatalf("round chakram audio owner = %p, want %p", gotOwner, unit)
			}
			audio = id
		},
	})

	var pin runtime.Pinner
	pinPlayerAttackProjectilePointers538960(t, &pin,
		unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(weapon),
		unsafe.Pointer(weaponAttrs), unsafe.Pointer(modifier))
	defer pin.Unpin()

	if got := playerAttackNativeEntry538960(unit); got != 1 {
		t.Fatalf("round chakram attack result = %d, want active frame", got)
	}
	if traceFlags != server.MapTraceFlags(5) || !created || !detached || !inserted || !attributes {
		t.Fatalf("round chakram effects = trace:%d create:%v detach:%v insert:%v attrs:%v",
			traceFlags, created, detached, inserted, attributes)
	}
	cosine, sine := server.SinCosDir(byte(unit.Direction1))
	if projectile.VelVec != types.Ptf(cosine*projectile.SpeedCur, sine*projectile.SpeedCur) ||
		projectile.Direction1 != unit.Direction1 || projectile.Direction2 != unit.Direction1 {
		t.Fatalf("round chakram motion = velocity:%+v directions:%d/%d",
			projectile.VelVec, projectile.Direction1, projectile.Direction2)
	}
	if projectileData.Reflections != 4 || projectileData.OwnerPos != unit.PosVec || projectileData.ReturnState != 2 {
		t.Fatalf("round chakram state = reflections:%d owner:%+v return:%d",
			projectileData.Reflections, projectileData.OwnerPos, projectileData.ReturnState)
	}
	if audio != sound.ID(891) || update.Field481 != 0xaabbcc02 || update.Field517 != 0x11223344 {
		t.Fatalf("round chakram result state = audio:%d frame:%#x animation:%#x",
			audio, update.Field481, update.Field517)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(weapon)
	runtime.KeepAlive(weaponAttrs)
	runtime.KeepAlive(modifier)
}

func TestPlayerAttackExport538960LaunchesNPCBowAndCrossbowWithNativePointers(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}
	for _, tc := range []struct {
		name           string
		weaponFlag     object.WeaponClass
		frame          uint32
		previous       uint8
		readiness      int32
		wantProjectile string
		wantAudio      sound.ID
		wantResult     int
		wantStored     uint8
		wantTraceFlags server.MapTraceFlags
	}{
		{
			name: "bow", weaponFlag: object.WeaponBow, frame: 3, previous: 2,
			wantProjectile: "ArcherArrow", wantAudio: sound.ID(885), wantStored: 3,
			wantTraceFlags: server.MapTraceFlags(5),
		},
		{
			name: "crossbow", weaponFlag: object.WeaponCrossbow, frame: 1, previous: 0,
			readiness: 1, wantProjectile: "ArcherBolt", wantAudio: sound.ID(886),
			wantResult: 1, wantStored: 2, wantTraceFlags: server.MapTraceFlags(5),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := installPlayerAttackProjectileServer538960(t)
			unit := &server.Object{
				ObjClass:    object.ClassMonster,
				ObjSubClass: object.SubClass(object.MonsterNPC),
				PosVec:      types.Ptf(320, 240),
				NewPos:      types.Ptf(320, 240),
				Direction1:  43,
			}
			unit.Shape.Kind = server.ShapeKindCircle
			unit.Shape.Circle.R = 5
			update := &server.MonsterUpdateData{
				Field331:         20,
				Field481:         0x33445500 | uint32(tc.previous),
				WeaponEquipFlags: uint32(tc.weaponFlag),
			}
			weaponAmmo := &server.AmmoUseData{}
			weaponAttrs := &server.ModifierInitData{}
			weapon := &server.Object{
				TypeInd:     0x3201,
				ObjClass:    object.ClassWeapon,
				ObjSubClass: object.SubClass(tc.weaponFlag),
				ObjFlags:    object.FlagEquipped,
				InvHolder:   unit,
				UseData:     server.UseDataPtr{Ptr: unsafe.Pointer(weaponAmmo)},
				InitData:    unsafe.Pointer(weaponAttrs),
			}
			quiverAmmo := &server.AmmoUseData{Charge1: 5}
			quiverAttrs := &server.ModifierInitData{}
			quiver := &server.Object{
				TypeInd:     0x3202,
				ObjClass:    object.ClassWeapon,
				ObjSubClass: object.SubClass(object.WeaponQuiver),
				ObjFlags:    object.FlagEquipped,
				InvHolder:   unit,
				UseData:     server.UseDataPtr{Ptr: unsafe.Pointer(quiverAmmo)},
				InitData:    unsafe.Pointer(quiverAttrs),
			}
			weapon.InvNextItem = quiver
			unit.InvFirstItem = weapon
			unit.UpdateData = unsafe.Pointer(update)
			modifier := &server.Modifier{TypeInd: uint32(weapon.TypeInd)}
			srv.Modif.Dword_5d4594_251600 = modifier

			collide := &server.ArrowCollideData{}
			projectileAttrs := &server.ModifierInitData{}
			projectile := &server.Object{
				SpeedCur:    9,
				CollideData: unsafe.Pointer(collide),
				InitData:    unsafe.Pointer(projectileAttrs),
			}
			spawn := playerAttackProjectileSpawnPoint538960(unit)
			var gotType string
			var gotAudio sound.ID
			var traceFlags server.MapTraceFlags
			var created, attributes bool
			installPlayerAttackProjectileRuntime538960(t, playerAttackProjectileRuntime538960{
				Frame: func() uint32 { return tc.frame },
				AnimFrames: func(action int) (int, int) {
					wantAction := 33
					if tc.weaponFlag == object.WeaponCrossbow {
						wantAction = 34
					}
					if action != wantAction {
						t.Fatalf("projectile animation = %d, want %d", action, wantAction)
					}
					return 4, 0
				},
				Readiness: func(gotWeapon *server.Object) int32 {
					if gotWeapon != weapon {
						t.Fatalf("readiness weapon = %p, want %p", gotWeapon, weapon)
					}
					return tc.readiness
				},
				Trace: func(from, to types.Pointf, flags server.MapTraceFlags) bool {
					if from != unit.PosVec || to != spawn {
						t.Fatalf("projectile trace = %+v -> %+v, want %+v -> %+v", from, to, unit.PosVec, spawn)
					}
					traceFlags = flags
					return true
				},
				NewObject: func(name string) *server.Object {
					gotType = name
					return projectile
				},
				CreateAt: func(gotProjectile, gotOwner *server.Object, pos types.Pointf) {
					created = gotProjectile == projectile && gotOwner == unit && pos == spawn
				},
				ApplyModifierAttrs: func(gotProjectile *server.Object, attrs *server.ModifierInitData) {
					attributes = gotProjectile == projectile && attrs == quiverAttrs
				},
				AudioEvent: func(id sound.ID, gotOwner *server.Object) {
					if gotOwner != unit {
						t.Fatalf("projectile audio owner = %p, want %p", gotOwner, unit)
					}
					gotAudio = id
				},
				WeaponInventoryFlags: func(item *server.Object) uint32 {
					switch item {
					case weapon:
						return uint32(tc.weaponFlag)
					case quiver:
						return uint32(object.WeaponQuiver)
					default:
						t.Fatalf("weapon flag lookup = %p", item)
						return 0
					}
				},
				QuestMode: func() bool { return false },
			})

			var pin runtime.Pinner
			pinPlayerAttackProjectilePointers538960(t, &pin,
				unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(weapon),
				unsafe.Pointer(weaponAmmo), unsafe.Pointer(weaponAttrs), unsafe.Pointer(quiver),
				unsafe.Pointer(quiverAmmo), unsafe.Pointer(quiverAttrs), unsafe.Pointer(modifier))
			defer pin.Unpin()

			if got := playerAttackNativeEntry538960(unit); got != tc.wantResult {
				t.Fatalf("%s attack result = %d, want %d", tc.name, got, tc.wantResult)
			}
			if gotType != tc.wantProjectile || gotAudio != tc.wantAudio ||
				traceFlags != tc.wantTraceFlags || !created || !attributes {
				t.Fatalf("%s launch = type:%q audio:%d trace:%d create:%v attrs:%v",
					tc.name, gotType, gotAudio, traceFlags, created, attributes)
			}
			if collide.Owner != unit {
				t.Fatalf("%s projectile owner = %p, want %p", tc.name, collide.Owner, unit)
			}
			if quiverAmmo.Charge1 != 5 {
				t.Fatalf("%s NPC quiver charge = %d, want unchanged 5", tc.name, quiverAmmo.Charge1)
			}
			if uint8(update.Field481) != tc.wantStored {
				t.Fatalf("%s stored frame = %d, want %d", tc.name, uint8(update.Field481), tc.wantStored)
			}
			cosine, sine := server.SinCosDir(byte(unit.Direction1))
			if projectile.VelVec != types.Ptf(cosine*projectile.SpeedCur, sine*projectile.SpeedCur) ||
				projectile.Direction1 != unit.Direction1 || projectile.Direction2 != unit.Direction1 {
				t.Fatalf("%s projectile motion = %+v %d/%d",
					tc.name, projectile.VelVec, projectile.Direction1, projectile.Direction2)
			}
			runtime.KeepAlive(unit)
			runtime.KeepAlive(update)
			runtime.KeepAlive(weapon)
			runtime.KeepAlive(quiver)
			runtime.KeepAlive(modifier)
		})
	}
}

func TestPlayerAttackExport538960LaunchesNPCFanChakramAndEquipsNextWithNativePointers(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}
	srv := installPlayerAttackProjectileServer538960(t)
	unit := &server.Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterNPC),
		PosVec:      types.Ptf(700, 900),
		NewPos:      types.Ptf(700, 900),
		Direction1:  71,
	}
	unit.Shape.Kind = server.ShapeKindCircle
	unit.Shape.Circle.R = 8
	update := &server.MonsterUpdateData{
		Field331:         30,
		Field481:         0x77889901,
		WeaponEquipFlags: uint32(object.WeaponShuriken),
	}
	ammo := &server.AmmoUseData{Charge1: 1}
	weapon := &server.Object{
		TypeInd:     0x3301,
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponShuriken),
		ObjFlags:    object.FlagEquipped,
		InvHolder:   unit,
		UseData:     server.UseDataPtr{Ptr: unsafe.Pointer(ammo)},
	}
	next := &server.Object{
		TypeInd:     0x3302,
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponShuriken),
		InvHolder:   unit,
	}
	weapon.InvNextItem = next
	unit.InvFirstItem = weapon
	unit.UpdateData = unsafe.Pointer(update)
	modifier := &server.Modifier{TypeInd: uint32(weapon.TypeInd)}
	srv.Modif.Dword_5d4594_251600 = modifier

	collide := &server.ArrowCollideData{}
	projectile := &server.Object{SpeedCur: 11, CollideData: unsafe.Pointer(collide)}
	var deleted, equipped *server.Object
	var traceFlags server.MapTraceFlags
	var audio sound.ID
	installPlayerAttackProjectileRuntime538960(t, playerAttackProjectileRuntime538960{
		Frame: func() uint32 { return 2 },
		AnimFrames: func(action int) (int, int) {
			if action != 44 {
				t.Fatalf("fan chakram animation = %d, want 44", action)
			}
			return 4, 0
		},
		Trace: func(_, _ types.Pointf, flags server.MapTraceFlags) bool {
			traceFlags = flags
			return true
		},
		NewObject: func(name string) *server.Object {
			if name != "FanChakramInMotion" {
				t.Fatalf("fan chakram type = %q", name)
			}
			return projectile
		},
		CreateAt: func(gotProjectile, gotOwner *server.Object, _ types.Pointf) {
			if gotProjectile != projectile || gotOwner != unit {
				t.Fatalf("fan chakram create = %p/%p, want %p/%p", gotProjectile, gotOwner, projectile, unit)
			}
		},
		DetachInventory: func(gotOwner, gotWeapon *server.Object) {
			if gotOwner != unit || gotWeapon != weapon {
				t.Fatalf("fan chakram detach = %p/%p, want %p/%p", gotOwner, gotWeapon, unit, weapon)
			}
			unit.InvFirstItem = next
			weapon.InvNextItem = nil
		},
		AudioEvent: func(id sound.ID, gotOwner *server.Object) {
			if gotOwner != unit {
				t.Fatalf("fan chakram audio owner = %p, want %p", gotOwner, unit)
			}
			audio = id
		},
		DelayedDelete: func(obj *server.Object) { deleted = obj },
		WeaponInventoryFlags: func(item *server.Object) uint32 {
			if item != weapon && item != next {
				t.Fatalf("fan chakram flag lookup = %p", item)
			}
			return uint32(object.WeaponShuriken)
		},
		EquipWeapon: func(gotOwner, item *server.Object) int {
			if gotOwner != unit {
				t.Fatalf("fan chakram equip owner = %p, want %p", gotOwner, unit)
			}
			equipped = item
			return 1
		},
	})

	var pin runtime.Pinner
	pinPlayerAttackProjectilePointers538960(t, &pin,
		unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(weapon),
		unsafe.Pointer(ammo), unsafe.Pointer(next), unsafe.Pointer(modifier))
	defer pin.Unpin()

	if got := playerAttackNativeEntry538960(unit); got != 1 {
		t.Fatalf("fan chakram attack result = %d, want active frame", got)
	}
	if traceFlags != server.MapTraceFlags(4) || collide.Owner != unit ||
		deleted != weapon || equipped != next || ammo.Charge1 != 0 || audio != sound.ID(891) {
		t.Fatalf("fan chakram result = trace:%d owner:%p delete:%p equip:%p charge:%d audio:%d",
			traceFlags, collide.Owner, deleted, equipped, ammo.Charge1, audio)
	}
	if uint8(update.Field481) != 2 {
		t.Fatalf("fan chakram stored frame = %d, want 2", uint8(update.Field481))
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(weapon)
	runtime.KeepAlive(ammo)
	runtime.KeepAlive(next)
	runtime.KeepAlive(modifier)
}
