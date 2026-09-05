package server

import (
	"image"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
)

func TestCreateSpellProjectileNative4FDDA0PreservesPointersAndFixedWidths(t *testing.T) {
	const spellID = int32(math.MinInt32)
	source := &Object{
		PosVec:     types.Pointf{X: 10, Y: 20},
		VelVec:     types.Pointf{X: 1, Y: 2},
		Direction1: 5,
	}
	source.Shape.Circle.R = 6
	source.Buffs = uint32(1) << createSpellProjectileEnchant4FDDA0
	source.BuffsDur[createSpellProjectileEnchant4FDDA0] = math.MaxUint16
	source.BuffsPower[createSpellProjectileEnchant4FDDA0] = 0xfe
	target := new(Object)
	projectileUpdate := new(SpellProjectileUpdateData)
	projectile := &Object{
		SpeedCur:   7,
		UpdateData: unsafe.Pointer(projectileUpdate),
	}

	var (
		createdAt types.Pointf
		audioSeen bool
		applySeen bool
	)
	got := createSpellProjectileNative4FDDA0(source, target, spellID, createSpellProjectileNativeDeps4FDDA0{
		spellFlags: func(int32) uint32 {
			t.Fatal("explicit target loaded spell flags")
			return 0
		},
		searchTarget: func(*types.Pointf, *Object, uint32, float32, int32, *Object) *Object {
			t.Fatal("explicit target searched for a replacement")
			return nil
		},
		mapTrace: func(origin, destination types.Pointf, outPoint *types.Pointf, outGrid *image.Point, flags int32) int32 {
			if origin != (types.Pointf{X: 10, Y: 20}) {
				t.Fatalf("trace origin = %v", origin)
			}
			if destination != (types.Pointf{X: 13.5, Y: 27}) {
				t.Fatalf("trace destination = %v", destination)
			}
			if outPoint != nil || outGrid != nil || flags != createSpellProjectileTraceFlag4FDDA0 {
				t.Fatalf("trace outputs/flags = %p/%p/%d", outPoint, outGrid, flags)
			}
			return math.MinInt32
		},
		newObject: func(name string) *Object {
			if name != createSpellProjectileType4FDDA0 {
				t.Fatalf("object type = %q", name)
			}
			return projectile
		},
		directionX: func(direction int16) float32 {
			switch direction {
			case 5:
				return 0.25
			case 11:
				return 2
			default:
				t.Fatalf("direction X index = %d", direction)
				return 0
			}
		},
		directionY: func(direction int16) float32 {
			switch direction {
			case 5:
				return 0.5
			case 11:
				return -3
			default:
				t.Fatalf("direction Y index = %d", direction)
				return 0
			}
		},
		indexedDirection: func(direction int16, scratch *types.Pointf) {
			if direction != 11 || scratch == nil {
				t.Fatalf("indexed direction = %d/%p", direction, scratch)
			}
		},
		spellAudio: func(id, field int32) int32 {
			if id != spellID || field != 0 {
				t.Fatalf("spell audio args = %d/%d", id, field)
			}
			return math.MinInt32
		},
		audio: func(id int32, object *Object, kind int32, code uint32) {
			audioSeen = true
			if id != math.MinInt32 || object != source || kind != 0 || code != 0 {
				t.Fatalf("audio args = %d/%p/%d/%d", id, object, kind, code)
			}
		},
		runtime: CreateSpellProjectileRuntime4FDDA0{
			SpellGetPower: func(id spell.ID, object *Object) int32 {
				if int32(id) != spellID || object != source {
					t.Fatalf("power args = %d/%p", id, object)
				}
				return math.MinInt32
			},
			CreateAt: func(object, owner *Object, position types.Pointf, reserved int32) {
				if object != projectile || owner != source || reserved != 0 {
					t.Fatalf("create args = %p/%p/%d", object, owner, reserved)
				}
				createdAt = position
				// Every later source direction/velocity read is live in GAME.EXE.
				source.Direction1 = 11
				source.VelVec = types.Pointf{X: 100, Y: 200}
			},
			ApplyEnchant: func(object *Object, enchant EnchantID, duration int16, power uint8) {
				applySeen = true
				if object != projectile || int32(enchant) != createSpellProjectileEnchant4FDDA0 || duration != -1 || power != 0xfe {
					t.Fatalf("enchant args = %p/%d/%d/%d", object, enchant, duration, power)
				}
			},
		},
	})

	if got != projectile {
		t.Fatalf("result = %p, want %p", got, projectile)
	}
	if createdAt != (types.Pointf{X: 13.5, Y: 27}) {
		t.Fatalf("created at = %v", createdAt)
	}
	if projectileUpdate.Level16 != 0x80000000 || projectileUpdate.Spell12 != 0x80000000 {
		t.Fatalf("level/spell bits = %#x/%#x", projectileUpdate.Level16, projectileUpdate.Spell12)
	}
	if projectileUpdate.Field0 != source || projectileUpdate.Target != target || projectileUpdate.Field8 != source {
		t.Fatalf("projectile pointers = %p/%p/%p", projectileUpdate.Field0, projectileUpdate.Target, projectileUpdate.Field8)
	}
	if projectile.Direction1 != 11 || projectile.Direction2 != 11 {
		t.Fatalf("projectile directions = %d/%d", projectile.Direction1, projectile.Direction2)
	}
	if projectile.VelVec != (types.Pointf{X: 114, Y: 179}) {
		t.Fatalf("projectile velocity = %v", projectile.VelVec)
	}
	if !applySeen || !audioSeen {
		t.Fatalf("apply/audio seen = %t/%t", applySeen, audioSeen)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"source":     uintptr(unsafe.Pointer(source)),
			"target":     uintptr(unsafe.Pointer(target)),
			"projectile": uintptr(unsafe.Pointer(projectile)),
			"update":     uintptr(unsafe.Pointer(projectileUpdate)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}
	runtime.KeepAlive(source)
	runtime.KeepAlive(target)
	runtime.KeepAlive(projectile)
	runtime.KeepAlive(projectileUpdate)
}

func TestCreateSpellProjectileNative4FDDA0LoadsPlayerCursorAsSignedDwords(t *testing.T) {
	player := new(Player)
	player.CursorVec.X = int((int64(1) << 32) | 1)
	player.CursorVec.Y = int(int64(math.MinInt32) - 1)
	update := &PlayerUpdateData{Player: player}
	source := &Object{
		ObjClass:   4,
		Direction1: 0,
		UpdateData: unsafe.Pointer(update),
	}
	source.Shape.Circle.R = 1
	target := new(Object)

	got := createSpellProjectileNative4FDDA0(source, nil, math.MaxInt32, createSpellProjectileNativeDeps4FDDA0{
		spellFlags: func(id int32) uint32 {
			if id != math.MaxInt32 {
				t.Fatalf("flags spell = %d", id)
			}
			return 0xf1234567
		},
		searchTarget: func(aim *types.Pointf, object *Object, flags uint32, distance float32, mode int32, self *Object) *Object {
			if aim == nil || aim.X != 1 || aim.Y != float32(math.MaxInt32) {
				t.Fatalf("aim = %v", aim)
			}
			if object != source || self != source || flags != 0xf1234567 || distance != 600 || mode != 0 {
				t.Fatalf("search args = %p/%#x/%v/%d/%p", object, flags, distance, mode, self)
			}
			return target
		},
		mapTrace: func(types.Pointf, types.Pointf, *types.Pointf, *image.Point, int32) int32 {
			return 0
		},
		newObject: func(string) *Object {
			t.Fatal("zero trace created an object")
			return nil
		},
		directionX: func(direction int16) float32 {
			if direction != 0 {
				t.Fatalf("direction X = %d", direction)
			}
			return 1
		},
		directionY: func(direction int16) float32 {
			if direction != 0 {
				t.Fatalf("direction Y = %d", direction)
			}
			return 0
		},
	})
	if got != nil {
		t.Fatalf("zero trace result = %p", got)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(source)
	runtime.KeepAlive(target)
}

func TestCreateSpellProjectileDirection4FDDA0DoesNotWrapSignedIndex(t *testing.T) {
	cosine, sine := SinCosDir(255)
	if got := createSpellProjectileDirection4FDDA0(255); got != [2]float32{cosine, sine} {
		t.Fatalf("direction 255 = %v, want %v/%v", got, cosine, sine)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("negative int16 direction wrapped instead of faulting")
		}
	}()
	_ = createSpellProjectileDirection4FDDA0(-1)
}

func TestCreateSpellProjectileClassifyDirection4FDDA0(t *testing.T) {
	for _, tc := range []struct {
		value int32
		want  int32
	}{
		{value: 7, want: 1},
		{value: 6, want: 0},
		{value: -6, want: 0},
		{value: -7, want: -1},
	} {
		if got := createSpellProjectileClassifyDirection4FDDA0(tc.value, 6); got != tc.want {
			t.Fatalf("classify(%d) = %d, want %d", tc.value, got, tc.want)
		}
	}
}
