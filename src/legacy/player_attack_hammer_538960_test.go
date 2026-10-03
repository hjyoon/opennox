package legacy

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"runtime"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

type playerAttackHammerAudioFile538960 struct {
	server.File
	data []byte
}

func (f *playerAttackHammerAudioFile538960) ReadString8() (string, error) {
	return "HammerMissing", nil
}

func (f *playerAttackHammerAudioFile538960) ReadU8() uint8 {
	v := f.data[0]
	f.data = f.data[1:]
	return v
}

func (f *playerAttackHammerAudioFile538960) Skip(n int) { f.data = f.data[n:] }

type playerAttackHammerServer538960 struct {
	*playerAttackLegacyServer538960
	wallPos    types.Pointf
	wallRadius float32
	wallDamage int
	wallType   object.DamageType
	order      []string
}

func (s *playerAttackHammerServer538960) Nox_xxx_mapDamageToWalls_534FC0(
	_ image.Rectangle, pos types.Pointf, radius float32, damage int, typ object.DamageType, who *server.Object,
) bool {
	s.wallDamageCalls++
	s.wallDamageAttacker = who
	s.wallPos, s.wallRadius, s.wallDamage, s.wallType = pos, radius, damage, typ
	s.order = append(s.order, "wall")
	return false
}

type playerAttackHammerFixture538960 struct {
	srv       *server.Server
	bridge    *playerAttackHammerServer538960
	unit      *server.Object
	weapon    *server.Object
	update    *server.PlayerUpdateData
	player    *server.Player
	animation int
	sounds    int
	quake     []byte
	pin       runtime.Pinner
}

func newPlayerAttackHammerFixture538960(t *testing.T) *playerAttackHammerFixture538960 {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width attack routing applies to 64-bit builds")
	}
	f := &playerAttackHammerFixture538960{srv: server.New(nil, nil, strman.New())}
	t.Cleanup(f.srv.Close)
	f.srv.Map.Init()
	t.Cleanup(f.srv.Map.Free)
	if f.srv.Walls.Init() == 0 {
		t.Fatal("cannot initialize hammer wall grid")
	}
	t.Cleanup(f.srv.Walls.Free)
	f.srv.NetList.Init()
	t.Cleanup(f.srv.NetList.Free)
	f.bridge = &playerAttackHammerServer538960{playerAttackLegacyServer538960: &playerAttackLegacyServer538960{srv: f.srv}}
	oldGetServer := GetServer
	GetServer = func() Server { return f.bridge }
	t.Cleanup(func() { GetServer = oldGetServer })
	oldAnimation := playerAnimFrames4F9F90
	playerAnimFrames4F9F90 = func(action int) (int, int) {
		f.animation = action
		return 4, 1
	}
	t.Cleanup(func() { playerAnimFrames4F9F90 = oldAnimation })
	x, y := memmap.PtrFloat32(0x587000, 194136), memmap.PtrFloat32(0x587000, 194140)
	oldX, oldY := *x, *y
	*x, *y = 1, 0
	t.Cleanup(func() { *x, *y = oldX, oldY })
	f.unit = &server.Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(321, 654), Field34: 100}
	f.unit.Shape.Kind, f.unit.Shape.Circle.R, f.unit.Shape.Circle.R2 = server.ShapeKindCircle, 5, 25
	f.update = &server.PlayerUpdateData{Field59_0: 0}
	f.weapon = &server.Object{
		TypeInd: 0x3214, ObjClass: object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponHammer), ObjFlags: object.FlagEquipped, InvHolder: f.unit,
	}
	modifier := &server.Modifier{
		TypeInd: uint32(f.weapon.TypeInd), ReqStrength60: 20,
		DamageCoeffOrArmor64: 1.5, Range68: 40, DamageMin72: 10,
	}
	f.srv.Modif.Dword_5d4594_251600 = modifier
	t.Cleanup(f.pin.Unpin)
	pinPlayerAttackProjectilePointers538960(t, &f.pin,
		unsafe.Pointer(f.unit), unsafe.Pointer(f.update), unsafe.Pointer(f.weapon), unsafe.Pointer(modifier))
	f.player = f.srv.Players.NewRaw(2201)
	if f.player == nil {
		t.Fatal("cannot allocate hammer player")
	}
	f.player.WeaponEquip, f.player.Field8 = uint32(object.WeaponHammer), 0xaa17
	f.player.Info().SetField2239(37)
	f.player.PlayerUnit, f.player.Pos3632Vec = f.unit, f.unit.PosVec
	t.Cleanup(func() { f.player.PlayerUnit = nil })
	f.unit.UpdateData = unsafe.Pointer(f.update)
	f.update.Player, f.update.EquippedWeapon = f.player, f.weapon
	if !f.srv.Audio.ReadAVNT(&playerAttackHammerAudioFile538960{data: []byte{7, 1, 'x', 0, 0}}) {
		t.Fatal("cannot initialize the hammer sound descriptor")
	}
	f.srv.Audio.Reset()
	f.srv.Audio.OnSound(func(id sound.ID, kind int, unit *server.Object, pos types.Pointf) {
		if id != sound.SoundHammerMissing || kind != 0 || unit != f.unit || pos != f.unit.PosVec {
			t.Fatalf("hammer sound = %d/%d/%p/%v", id, kind, unit, pos)
		}
		f.sounds++
		f.bridge.order = append(f.bridge.order, "sound")
		// Audio follows the wall hit and the actual quake send, never before.
		f.quake = f.srv.NetList.CopyPacketsA(ntype.PlayerInd(f.player.PlayerInd), netlist.Kind1)
	})
	return f
}

func TestPlayerAttackExport538960RestoresHammerAOEAndLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name     string
		targets  bool
		mods     bool
		deadline uint32
	}{
		{"area hit", true, false, 0},
		{"four native effects", true, true, 777},
		{"miss still sounds and quakes", false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPlayerAttackHammerFixture538960(t)
			f.update.Field0 = tc.deadline
			wantDamage := 36
			if tc.mods {
				mod := &server.ModifierEff{Attack40: server.ModifierEffFnc{Fnc: modifierDamageMultiplierPointer4E04C0(), Valf: 1.5}}
				data := &server.ModifierInitData{Modifiers: [4]*server.ModifierEff{mod, mod, mod, mod}}
				pinPlayerAttackProjectilePointers538960(t, &f.pin, mod.C(), unsafe.Pointer(data))
				f.weapon.InitData = unsafe.Pointer(data)
				wantDamage = 180 // binary32 35.5 * 1.5^4, then +0.5 truncation
			}
			var hits [6]int
			var targets []*server.Object
			if tc.targets {
				damagePtr := objectDamageNativeProbePtr()
				server.RegisterObjectDamageGo(fmt.Sprintf("PlayerHammer%d", objectDamageNativeTestSequence.Add(1)), damagePtr,
					func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
						index := slices.Index(targets, target)
						if index < 0 || source != f.unit || weapon != f.weapon || damage != int32(wantDamage) || typ != object.DamageCrush {
							t.Fatalf("hammer damage = index:%d source:%p weapon:%p value:%d/%d", index, source, weapon, damage, typ)
						}
						hits[index]++
						f.bridge.order = append(f.bridge.order, "damage")
						return true
					})
				// Center is +35, radius is modifier 40 ONLY (not + unit R=5).
				// Two forward targets hit; exact radius+shape edge, outside,
				// behind the facing plane, and destroyed target must not hit.
				for i, x := range []float32{346, 390, 401, 402, 320, 350} {
					data := &server.MonsterUpdateData{}
					obj := &server.Object{TypeInd: math.MaxUint16, ObjClass: object.ClassMonster, ObjFlags: object.FlagActive,
						PosVec: types.Ptf(x, 654), NewPos: types.Ptf(x, 654), UpdateData: unsafe.Pointer(data), Damage: damagePtr}
					obj.Shape.Kind, obj.Shape.Circle.R, obj.Shape.Circle.R2 = server.ShapeKindCircle, 5, 25
					if i == 5 {
						obj.ObjFlags |= object.FlagDestroyed
					}
					pinPlayerAttackProjectilePointers538960(t, &f.pin, unsafe.Pointer(obj), unsafe.Pointer(data))
					f.srv.Map.AddObjectToIndex(obj)
					targets = append(targets, obj)
				}
			}
			for _, frame := range []uint32{100, 102, 104, 104, 108, 110} {
				f.srv.SetFrame(frame)
				wantActive := 1
				wantFrame := byte((frame - 100) / 2)
				if frame >= 108 {
					wantActive, wantFrame = 0, 3
				}
				if got := playerAttackNativeEntry538960(f.unit); got != wantActive {
					t.Fatalf("frame %d attack = %d, want %d", frame, got, wantActive)
				}
				wantDeadline := tc.deadline
				if wantDeadline == 0 {
					wantDeadline = 108
				}
				if f.animation != 39 || f.update.Field59_0 != wantFrame || f.update.Field0 != wantDeadline || f.player.Field8 != 0xaa17 {
					t.Fatalf("frame %d hammer state = animation:%d frame:%d deadline:%d variant:%#x", frame, f.animation, f.update.Field59_0, f.update.Field0, f.player.Field8)
				}
				wantHits := 0
				if frame >= 104 {
					wantHits = 1
				}
				if f.sounds != wantHits || f.bridge.wallDamageCalls != wantHits {
					t.Fatalf("frame %d hammer calls = sounds:%d walls:%d, want %d", frame, f.sounds, f.bridge.wallDamageCalls, wantHits)
				}
			}
			wantHits := [6]int{}
			if tc.targets {
				wantHits = [6]int{1, 1}
			}
			if hits != wantHits || !bytes.Equal(f.quake, []byte{0x97, 4}) ||
				f.bridge.wallPos != types.Ptf(356, 654) || f.bridge.wallRadius != 65 || f.bridge.wallDamage != wantDamage ||
				f.bridge.wallType != object.DamageCrush || f.bridge.wallDamageAttacker != f.weapon {
				t.Fatalf("hammer outcome = hits:%v quake:% x wall:%v/%g/%d/%d/%p", hits, f.quake, f.bridge.wallPos, f.bridge.wallRadius, f.bridge.wallDamage, f.bridge.wallType, f.bridge.wallDamageAttacker)
			}
			wantOrder := []string{"wall", "sound"}
			if tc.targets {
				wantOrder = []string{"damage", "damage", "wall", "sound"}
			}
			if !slices.Equal(f.bridge.order, wantOrder) || f.weapon.PosVec != f.unit.PosVec || f.weapon.PrevPos != f.unit.PosVec {
				t.Fatalf("hammer order/weapon = %v/%v/%v", f.bridge.order, f.weapon.PosVec, f.weapon.PrevPos)
			}
		})
	}
}

func TestPlayerAttackExport538960HammerDoesNotHitAfterSkippedMiddleFrame(t *testing.T) {
	f := newPlayerAttackHammerFixture538960(t)
	for _, frame := range []uint32{100, 102, 106, 108} {
		f.srv.SetFrame(frame)
		playerAttackNativeEntry538960(f.unit)
	}
	if f.sounds != 0 || f.bridge.wallDamageCalls != 0 || len(f.quake) != 0 || f.update.Field59_0 != 3 || f.update.Field0 != 108 {
		t.Fatalf("skipped hammer frame = sounds:%d walls:%d quake:% x frame:%d deadline:%d", f.sounds, f.bridge.wallDamageCalls, f.quake, f.update.Field59_0, f.update.Field0)
	}
}

func TestPlayerAttackExport538960HammerKeepsMeleePriority(t *testing.T) {
	for _, tc := range []struct {
		mask      uint32
		animation int
	}{
		{0, 39}, {0x200, 28}, {0x100, 27}, {0x400, 37},
		{0x800, 39}, {0x1000, 39}, {0x2000, 39}, {0x3000, 39},
	} {
		t.Run(fmt.Sprintf("%#x", tc.mask), func(t *testing.T) {
			f := newPlayerAttackHammerFixture538960(t)
			f.player.WeaponEquip |= tc.mask
			f.srv.SetFrame(100)
			if got := playerAttackNativeEntry538960(f.unit); got != 1 || f.animation != tc.animation || f.sounds != 0 || f.bridge.wallDamageCalls != 0 {
				t.Fatalf("mask %#x = result:%d animation:%d sounds:%d walls:%d", f.player.WeaponEquip, got, f.animation, f.sounds, f.bridge.wallDamageCalls)
			}
		})
	}
}
