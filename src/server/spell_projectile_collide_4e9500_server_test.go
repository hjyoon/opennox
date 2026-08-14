package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
)

func TestSpellProjectileReflect4E0A70OriginalSpillsAndCopies(t *testing.T) {
	projectile := &Object{
		PosVec:     types.Ptf(10.25, -3.5),
		PrevPos:    types.Ptf(math.Float32frombits(0x7fc12345), math.Float32frombits(0x80000000)),
		NewPos:     types.Ptf(99, 88),
		VelVec:     types.Ptf(6.5, -2.25),
		Direction1: 200,
	}
	other := &Object{PosVec: types.Ptf(-2.75, 4.25)}
	spellProjectileReflect4E0A70(projectile, other)
	if got, want := math.Float32bits(projectile.VelVec.X), uint32(0xc0a0224a); got != want {
		t.Fatalf("velocity X bits = %#08x, want %#08x", got, want)
	}
	if got, want := math.Float32bits(projectile.VelVec.Y), uint32(0x4092c8b8); got != want {
		t.Fatalf("velocity Y bits = %#08x, want %#08x", got, want)
	}
	if projectile.Direction1 != 72 {
		t.Fatalf("direction = %d, want 72", projectile.Direction1)
	}
	if math.Float32bits(projectile.NewPos.X) != 0x7fc12345 || math.Float32bits(projectile.NewPos.Y) != 0x80000000 {
		t.Fatalf("NewPos bits = (%#08x,%#08x)", math.Float32bits(projectile.NewPos.X), math.Float32bits(projectile.NewPos.Y))
	}
}

func TestSpellProjectileReflect4E0A70SignedDirectionBranch(t *testing.T) {
	projectile := &Object{Direction1: 0x7f80}
	spellProjectileReflect4E0A70(projectile, &Object{})
	if projectile.Direction1 != 0x8000 {
		t.Fatalf("signed-negative direction = %#04x, want 0x8000", projectile.Direction1)
	}

	projectile = &Object{Direction1: 0xff80}
	spellProjectileReflect4E0A70(projectile, &Object{})
	if projectile.Direction1 != 0 {
		t.Fatalf("wrapped direction = %#04x, want 0", projectile.Direction1)
	}
}

func TestSpellProjectileWallReflect57B810NoBinary32ProductSpill(t *testing.T) {
	tiny := math.Float32frombits(1)
	projectile := &Object{VelVec: types.Ptf(2, -3)}
	spellProjectileWallReflect57B810(&types.Pointf{X: tiny, Y: tiny}, projectile)
	if projectile.VelVec != (types.Pointf{X: 3, Y: -2}) {
		t.Fatalf("positive tiny normal reflection = %v, want {3 -2}", projectile.VelVec)
	}

	projectile.VelVec = types.Ptf(math.Float32frombits(0x7fc12345), math.Float32frombits(0x80000000))
	spellProjectileWallReflect57B810(&types.Pointf{X: 1, Y: -1}, projectile)
	if math.Float32bits(projectile.VelVec.X) != 0x80000000 || math.Float32bits(projectile.VelVec.Y) != 0x7fc12345 {
		t.Fatalf("swap bits = (%#08x,%#08x)", math.Float32bits(projectile.VelVec.X), math.Float32bits(projectile.VelVec.Y))
	}

	projectile.VelVec = types.Ptf(7, -11)
	spellProjectileWallReflect57B810(&types.Pointf{X: math.Float32frombits(0x7fc00001), Y: 1}, projectile)
	if projectile.VelVec != (types.Pointf{X: -11, Y: 7}) {
		t.Fatalf("unordered normal reflection = %v, want {-11 7}", projectile.VelVec)
	}
}

func TestSpellProjectileInversionNative4FA4F0(t *testing.T) {
	var inversionMarker, wrongMarker byte
	inversion := unsafe.Pointer(&inversionMarker)
	weak := &ModifierEff{DefendCollide88: ModifierEffFnc{Fnc: inversion, Val: 0}}
	wrong := &ModifierEff{DefendCollide88: ModifierEffFnc{Fnc: unsafe.Pointer(&wrongMarker), Val: 9}}
	strong := &ModifierEff{DefendCollide88: ModifierEffFnc{Fnc: inversion, Val: 2}}
	firstData := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, weak, wrong}}
	secondData := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, nil, strong}}
	second := &Object{ObjFlags: 0x100, ObjClass: 0x1000, InitData: unsafe.Pointer(secondData)}
	first := &Object{ObjFlags: 0x100, ObjClass: 0x10000000, InitData: unsafe.Pointer(firstData), InvNextItem: second}
	target := &Object{InvFirstItem: first}
	projectile := &Object{}
	if got := spellProjectileInversionNative4FA4F0(target, projectile, inversion); got != 1 {
		t.Fatalf("native inversion = %d, want 1", got)
	}
	strong.DefendCollide88.Val = math.MinInt32
	if got := spellProjectileInversionNative4FA4F0(target, projectile, inversion); got != 0 {
		t.Fatalf("negative inversion = %d, want 0", got)
	}
}

func TestSpellProjectileCollideNative4E9500GreatSwordAndAccept(t *testing.T) {
	owner := &Object{}
	source := &Object{}
	player := &Player{WeaponEquip: 0x400}
	playerUpdate := &PlayerUpdateData{State: PlayerState13, Player: player}
	other := &Object{ObjClass: 4, Direction1: 17, UpdateData: unsafe.Pointer(playerUpdate)}
	update := &SpellProjectileUpdateData{Field0: owner, Target: other, Field8: source, Spell12: 31, Level16: 7}
	projectile := &Object{
		PosVec:     types.Ptf(3, 4),
		PrevPos:    types.Ptf(8, 9),
		VelVec:     types.Ptf(1, 2),
		Direction1: 250,
		UpdateData: unsafe.Pointer(update),
	}
	events := make([]string, 0, 16)
	spellProjectileCollideNative4E9500(projectile, other, nil, spellProjectileNativeDeps4E9500{
		runtime: SpellProjectileCollideRuntime4E9500{
			CheckDirection: func(first types.Pointf, direction int16, second types.Pointf) int32 {
				events = append(events, "direction")
				if first != other.PosVec || direction != int16(other.Direction1) || second != projectile.PrevPos {
					t.Fatalf("direction args = (%v,%d,%v)", first, direction, second)
				}
				return 1
			},
			ChangeOwner: func(gotProjectile, gotOther *Object) {
				events = append(events, "owner")
				if gotProjectile != projectile || gotOther != other {
					t.Fatal("wrong owner arguments")
				}
			},
			SetPlayerState: func(obj *Object, state PlayerState) bool {
				events = append(events, "state")
				playerUpdate.State = state
				return true
			},
			SpellAccept: func(spellID spell.ID, gotSource, gotOwner, gotProjectile *Object, arg *SpellAcceptArg, level int) bool {
				events = append(events, "accept")
				if spellID != 31 || gotSource != source || gotOwner != owner || gotProjectile != projectile || arg.Obj != other || arg.Pos != (types.Pointf{}) || level != 7 {
					t.Fatalf("spell args = (%d,%p,%p,%p,%v,%d)", spellID, gotSource, gotOwner, gotProjectile, *arg, level)
				}
				return false
			},
			DelayedDelete: func(obj *Object) {
				events = append(events, "delete")
				if obj != projectile {
					t.Fatal("wrong deleted object")
				}
			},
		},
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, "random")
			if minimum != 18 || maximum != 20 {
				t.Fatalf("random bounds = %d..%d", minimum, maximum)
			}
			return 18
		},
		playerAnimFrames: func(action int32) (int32, int32) {
			events = append(events, "frames")
			if action != 48 {
				t.Fatalf("action = %d, want 48", action)
			}
			return 11, 15
		},
		audio: func(id uint32, obj *Object) {
			events = append(events, "audio")
			if id != 890 || obj != other {
				t.Fatalf("audio = (%d,%p)", id, obj)
			}
		},
	})
	want := []string{"direction", "random", "audio", "owner", "state", "frames", "accept", "delete"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if playerUpdate.State != PlayerState18 || playerUpdate.Field59_0 != 10 {
		t.Fatalf("state/frame = %d/%d, want 18/10", playerUpdate.State, playerUpdate.Field59_0)
	}
	if projectile.Direction1 != 122 || projectile.NewPos != projectile.PrevPos {
		t.Fatalf("projectile direction/NewPos = %d/%v", projectile.Direction1, projectile.NewPos)
	}
}
