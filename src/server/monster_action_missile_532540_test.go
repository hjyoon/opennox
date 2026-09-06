package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func missileMonsterTestObject532540(t *testing.T) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_MISSILE_ATTACK)
	update.MonsterDef = &MonsterDef{
		MissileAttackRange212:    15,
		MissileAttackFrame216:    2,
		MissileAttackDelayMin220: 20,
		MissileAttackDelayMax224: 30,
	}
	copy(update.MonsterDef.MissileName148[:], "TestMissile")
	return unit
}

func assertMissileEvents532540(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
}

func TestMonsterActionMissileStart532540AttackStateAndOracleOrder(t *testing.T) {
	unit := missileMonsterTestObject532540(t)
	update := unit.UpdateDataMonster()
	update.Field128 = 100
	var sounds [17]uint32
	sounds[10] = 1010
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	frames := []uint32{100, 101, 102}
	frameIndex := 0
	var events []string
	handled := monsterActionMissileStart532540(unit, monsterActionMissileStartHooks532540{
		frame: func() uint32 {
			events = append(events, "frame")
			if frameIndex >= len(frames) {
				t.Fatal("too many frame reads")
			}
			frame := frames[frameIndex]
			frameIndex++
			return frame
		},
		random: func(minimum, maximum int) int {
			if minimum != 20 || maximum != 30 {
				t.Fatalf("random bounds = %d..%d", minimum, maximum)
			}
			events = append(events, "random")
			return 24
		},
		buffOff: func(got *Object, enchant EnchantID) {
			if got != unit {
				t.Fatalf("buff unit = %p", got)
			}
			switch enchant {
			case 0:
				events = append(events, "buff0")
			case 23:
				events = append(events, "buff23")
			default:
				t.Fatalf("unexpected enchant = %d", enchant)
			}
		},
		audio: func(id uint32, got *Object) {
			if id != 1010 || got != unit || unit.Field34 != 101 || update.Field128 != 126 {
				t.Fatalf("audio state = %d/%p/%d/%d", id, got, unit.Field34, update.Field128)
			}
			events = append(events, "audio")
		},
		push: func(ai.ActionType, ...any) *AIStackItem {
			t.Fatal("ready attack pushed an action")
			return nil
		},
	})
	if !handled || frameIndex != 3 {
		t.Fatalf("handled/frame reads = %v/%d, want true/3", handled, frameIndex)
	}
	assertMissileEvents532540(t, events, []string{"frame", "buff0", "buff23", "frame", "random", "frame", "audio"})
}

func TestMonsterActionMissileStart532540CooldownKeepsNativeTarget(t *testing.T) {
	unit := missileMonsterTestObject532540(t)
	target := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = target
	update.Field128 = 150
	var pushed []*AIStackItem
	handled := monsterActionMissileStart532540(unit, monsterActionMissileStartHooks532540{
		frame:  func() uint32 { return 100 },
		random: func(int, int) int { t.Fatal("cooldown branch used RNG"); return 0 },
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			item := &AIStackItem{Action: uint32(action)}
			item.SetArgs(args...)
			pushed = append(pushed, item)
			return item
		},
	})
	if !handled || len(pushed) != 2 {
		t.Fatalf("handled/pushed = %v/%d, want true/2", handled, len(pushed))
	}
	if pushed[0].Type() != ai.DEPENDENCY_OBJECT_CLOSER_THAN || pushed[0].ArgF32(0) != 18 ||
		pushed[0].ArgU32(1) != 0 || pushed[0].ArgObj(2) != target {
		t.Fatalf("dependency = %#v", pushed[0])
	}
	if pushed[1].Type() != ai.ACTION_WAIT || pushed[1].ArgU32(0) != 150 {
		t.Fatalf("wait = %#v", pushed[1])
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && pushed[0].Args[2] <= uintptr(^uint32(0)) {
		t.Fatalf("target argument was truncated to %#x", pushed[0].Args[2])
	}
}

func TestMonsterProjectileLeadAndTrajectory532610(t *testing.T) {
	unit := &Object{PosVec: types.Ptf(100, 80), Direction1: 32}
	unit.Shape.Circle.R = 6
	target := &Object{PosVec: types.Ptf(148, 104), VelVec: types.Ptf(2, -1)}
	projectile := &Object{SpeedCur: 12}
	lead := monsterProjectileLead533080(unit, target, projectile.SpeedCur)
	if math.Abs(float64(lead.X)-156.94427) > 0.00001 || math.Abs(float64(lead.Y)-99.52786) > 0.00001 {
		t.Fatalf("lead = %v", lead)
	}
	spawn, velocity, traceTo := monsterMissileTrajectory532610(unit, projectile, lead)
	direction := unit.Direction1.Vec()
	wantSpawn := types.Ptf(
		float32(float64(10)*float64(direction.X)+100),
		float32(float64(10)*float64(direction.Y)+80),
	)
	if spawn != wantSpawn || traceTo != spawn.Add(velocity) {
		t.Fatalf("spawn/velocity/trace = %v/%v/%v, want spawn %v and trace sum", spawn, velocity, traceTo, wantSpawn)
	}
	if velocity.X <= 0 || velocity.Y <= 0 {
		t.Fatalf("velocity = %v, want positive X/Y", velocity)
	}
}

func TestMonsterActionMissileUpdate532610CreatesPredictedProjectile(t *testing.T) {
	unit := missileMonsterTestObject532540(t)
	unit.PosVec = types.Ptf(100, 80)
	unit.Direction1 = 32
	unit.Shape.Circle.R = 6
	target := monsterActionTestObject50A910(t)
	target.PosVec = types.Ptf(148, 104)
	target.VelVec = types.Ptf(2, -1)
	update := unit.UpdateDataMonster()
	update.AIStack[0].SetArgs(types.Ptf(-1, -2), target)
	update.Field120_1 = 2
	update.Field120_2 = 0
	update.Field120_3 = 1
	var sounds [17]uint32
	sounds[11] = 1111
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	projectile := &Object{SpeedCur: 12, Direction1: 17, Direction2: 18}
	lead := monsterProjectileLead533080(unit, target, projectile.SpeedCur)
	wantSpawn, wantVelocity, wantTraceTo := monsterMissileTrajectory532610(unit, projectile, lead)
	var events []string
	handled := monsterActionMissileUpdate532610(unit, monsterActionMissileUpdateHooks532610{
		newObject: func(id string) *Object {
			if id != "TestMissile" {
				t.Fatalf("projectile ID = %q", id)
			}
			events = append(events, "new")
			return projectile
		},
		trace: func(from, to types.Pointf, flags MapTraceFlags) bool {
			if from != unit.PosVec || to != wantTraceTo || flags != MapTraceFlags(5) {
				t.Fatalf("trace = %v -> %v flags %#x, want %v -> %v flags 5", from, to, flags, unit.PosVec, wantTraceTo)
			}
			events = append(events, "trace")
			return true
		},
		createAt: func(got, owner *Object, pos types.Pointf) {
			if got != projectile || owner != unit || pos != wantSpawn {
				t.Fatalf("create = %p/%p/%v", got, owner, pos)
			}
			if projectile.VelVec != (types.Pointf{}) || projectile.Direction1 != 17 || projectile.Direction2 != 18 {
				t.Fatalf("projectile mutated before create = %v/%d/%d", projectile.VelVec, projectile.Direction1, projectile.Direction2)
			}
			events = append(events, "create")
		},
		delayedDelete: func(*Object) { t.Fatal("clear trace deleted projectile") },
		audio: func(id uint32, got *Object) {
			if id != 1111 || got != unit || projectile.VelVec != wantVelocity ||
				projectile.Direction1 != unit.Direction1 || projectile.Direction2 != unit.Direction1 {
				t.Fatalf("audio/final state = %d/%p/%v/%d/%d", id, got, projectile.VelVec, projectile.Direction1, projectile.Direction2)
			}
			events = append(events, "audio")
		},
		pop: func() int { events = append(events, "pop"); return 0 },
	})
	if !handled {
		t.Fatal("missile update was not handled")
	}
	assertMissileEvents532540(t, events, []string{"new", "trace", "create", "audio", "pop"})
}

func TestMonsterActionMissileUpdate532610BlockedTraceDeletesButStillSounds(t *testing.T) {
	unit := missileMonsterTestObject532540(t)
	unit.PosVec = types.Ptf(40, 50)
	unit.Direction1 = 0
	unit.Shape.Circle.R = 3
	update := unit.UpdateDataMonster()
	update.AIStack[0].SetArgs(types.Ptf(90, 70))
	update.Field120_1 = 2
	var sounds [17]uint32
	sounds[11] = 1111
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	projectile := &Object{SpeedCur: 8, VelVec: types.Ptf(7, 9)}
	_, _, wantTraceTo := monsterMissileTrajectory532610(unit, projectile, types.Ptf(90, 70))
	var events []string
	handled := monsterActionMissileUpdate532610(unit, monsterActionMissileUpdateHooks532610{
		newObject: func(string) *Object { events = append(events, "new"); return projectile },
		trace: func(from, to types.Pointf, flags MapTraceFlags) bool {
			if from != unit.PosVec || to != wantTraceTo || flags != 5 {
				t.Fatalf("trace = %v -> %v flags %#x", from, to, flags)
			}
			events = append(events, "trace")
			return false
		},
		createAt: func(*Object, *Object, types.Pointf) { t.Fatal("blocked trace created projectile") },
		delayedDelete: func(got *Object) {
			if got != projectile {
				t.Fatalf("deleted = %p", got)
			}
			events = append(events, "delete")
		},
		audio: func(id uint32, got *Object) {
			if id != 1111 || got != unit {
				t.Fatalf("audio = %d/%p", id, got)
			}
			events = append(events, "audio")
		},
		pop: func() int { t.Fatal("unfinished animation popped"); return 0 },
	})
	if !handled || projectile.VelVec != types.Ptf(7, 9) {
		t.Fatalf("handled/velocity = %v/%v", handled, projectile.VelVec)
	}
	assertMissileEvents532540(t, events, []string{"new", "trace", "delete", "audio"})
}

func TestMonsterActionMissileUpdate532610AllocationFailureStillSoundsAndCompletes(t *testing.T) {
	unit := missileMonsterTestObject532540(t)
	update := unit.UpdateDataMonster()
	update.Field120_1 = 2
	update.Field120_3 = 1
	var sounds [17]uint32
	sounds[11] = 1111
	update.SoundSet122 = unsafe.Pointer(&sounds[0])
	var events []string
	handled := monsterActionMissileUpdate532610(unit, monsterActionMissileUpdateHooks532610{
		newObject: func(string) *Object { events = append(events, "new"); return nil },
		trace: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
			t.Fatal("nil projectile was traced")
			return false
		},
		audio: func(uint32, *Object) {
			events = append(events, "audio")
		},
		pop: func() int { events = append(events, "pop"); return 0 },
	})
	if !handled {
		t.Fatal("allocation failure was not handled")
	}
	assertMissileEvents532540(t, events, []string{"new", "audio", "pop"})
}

func TestMonsterActionMissileUpdate532610NPCUsesTypedPlayerAttack(t *testing.T) {
	for _, tc := range []struct {
		name    string
		attack  int
		wantPop bool
	}{
		{name: "continues", attack: 1},
		{name: "completes", attack: 0, wantPop: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := missileMonsterTestObject532540(t)
			unit.ObjSubClass = object.SubClass(object.MonsterNPC)
			called := false
			popped := false
			handled := monsterActionMissileUpdate532610(unit, monsterActionMissileUpdateHooks532610{
				playerAttack: func(got *Object) int {
					if got != unit {
						t.Fatalf("attack unit = %p", got)
					}
					called = true
					return tc.attack
				},
				newObject: func(string) *Object { t.Fatal("NPC allocated monster projectile"); return nil },
				pop:       func() int { popped = true; return 0 },
			})
			if !handled || !called || popped != tc.wantPop {
				t.Fatalf("handled/called/popped = %v/%v/%v, want true/true/%v", handled, called, popped, tc.wantPop)
			}
		})
	}
}

func TestMonsterActionMissileUpdate532610ContainsInvalidMetadata(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*MonsterUpdateData)
	}{
		{name: "missing head", setup: func(update *MonsterUpdateData) { update.AIStackInd = -1 }},
		{name: "head out of range", setup: func(update *MonsterUpdateData) { update.AIStackInd = int8(len(update.AIStack)) }},
		{name: "missing definition", setup: func(update *MonsterUpdateData) { update.MonsterDef = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := missileMonsterTestObject532540(t)
			tc.setup(unit.UpdateDataMonster())
			popped := false
			handled := monsterActionMissileUpdate532610(unit, monsterActionMissileUpdateHooks532610{
				newObject: func(string) *Object { t.Fatal("invalid metadata allocated projectile"); return nil },
				pop:       func() int { popped = true; return 0 },
			})
			if !handled || !popped {
				t.Fatalf("handled/popped = %v/%v, want true/true", handled, popped)
			}
		})
	}
}
