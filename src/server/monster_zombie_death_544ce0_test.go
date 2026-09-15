package server

import (
	"math"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterZombieTestServer544CE0(t *testing.T) *Server {
	t.Helper()
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	return s
}

func TestMonsterActionDeadStart544D80ZombieDuration(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	deadFunc := unsafe.Pointer(new(byte))
	update.MonsterDef = &MonsterDef{DeadFunc232: deadFunc}
	unit.ObjFlags = object.FlagDead
	unit.VelVec = types.Ptf(1, 2)
	unit.ForceVec = types.Ptf(3, 4)
	unit.Pos24 = types.Ptf(5, 6)
	var events []string
	runtime := MonsterActionDeadRuntime544D80{
		IsZombie:    func(got *Object) bool { return got == unit },
		CanDeadFunc: func(got unsafe.Pointer) bool { return got == deadFunc },
		DeadFunc: func(got unsafe.Pointer, obj *Object) {
			if got != deadFunc || obj != unit || unit.VelVec != (types.Pointf{}) ||
				unit.ForceVec != (types.Pointf{}) || unit.Pos24 != (types.Pointf{}) {
				t.Fatalf("dead callback state = %p/%p/%v/%v/%v", got, obj, unit.VelVec, unit.ForceVec, unit.Pos24)
			}
			events = append(events, "dead")
		},
		ZombieDeadDuration: func(index int) float32 {
			events = append(events, []string{"duration-min", "duration-max"}[index])
			return []float32{7.9, 12.9}[index]
		},
		RandomInt: func(minimum, maximum int) int {
			if minimum != 7 || maximum != 12 {
				t.Fatalf("duration bounds = %d/%d, want 7/12", minimum, maximum)
			}
			events = append(events, "random")
			return 9
		},
	}
	if !new(Server).MonsterActionDeadStart544D80(unit, runtime) {
		t.Fatal("zombie dead start was not handled")
	}
	wantEvents := []string{"dead", "duration-min", "duration-max", "random"}
	if len(events) != len(wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	for i := range wantEvents {
		if events[i] != wantEvents[i] {
			t.Fatalf("events = %v, want %v", events, wantEvents)
		}
	}
	if update.Field123 != 9 {
		t.Fatalf("zombie dead duration = %d, want 9", update.Field123)
	}
	if !unit.ObjFlags.Has(object.FlagShort) || unit.ObjFlags.Has(object.FlagAllowOverlap) {
		t.Fatalf("zombie dead flags = %#x, want Short without AllowOverlap", unit.ObjFlags)
	}
}

func TestMonsterActionDeadStart544D80ZombiePreflightsDuration(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	unit.UpdateDataMonster().MonsterDef = &MonsterDef{}
	unit.VelVec = types.Ptf(1, 2)
	before := *unit
	var reason string
	if new(Server).MonsterActionDeadStart544D80(unit, MonsterActionDeadRuntime544D80{
		IsZombie: func(*Object) bool { return true },
		Unsupported: func(got string, _ *Object) {
			reason = got
		},
	}) {
		t.Fatal("missing zombie duration callbacks were accepted")
	}
	if reason != "zombie dead duration" || *unit != before {
		t.Fatalf("preflight result = %q, mutated=%v", reason, *unit != before)
	}
}

func TestMonsterActionDeadUpdate544EC0BurningZombiePriority(t *testing.T) {
	s := new(Server)
	s.SetFrame(100)
	unit := monsterActionTestObject50A910(t)
	unit.PosVec = types.Ptf(321, 654)
	update := unit.UpdateDataMonster()
	update.Field137 = 1
	update.Field123 = 2
	update.CurrentEnemy = &Object{}
	update.StatusFlags = object.MonStatusOnFire | object.MonStatusStayDead
	var events []string
	if !s.MonsterActionDeadUpdate544EC0(unit, MonsterActionDeadRuntime544D80{
		IsZombie: func(*Object) bool { return true },
		SparkExplosion: func(position types.Pointf, strength byte) {
			if position != unit.PosVec || strength != 100 {
				t.Fatalf("spark explosion = %v/%d", position, strength)
			}
			events = append(events, "spark")
		},
		ZombieBurnDelete: func(got *Object) {
			if got != unit {
				t.Fatalf("burn-delete unit = %p, want %p", got, unit)
			}
			events = append(events, "delete")
		},
		RaiseZombie: func(*Object) { t.Fatal("burning zombie was raised") },
	}) {
		t.Fatal("burning zombie update was not handled")
	}
	if len(events) != 2 || events[0] != "spark" || events[1] != "delete" {
		t.Fatalf("events = %v, want [spark delete]", events)
	}
}

func TestMonsterActionDeadUpdate544EC0ZombieRaiseGates(t *testing.T) {
	tests := []struct {
		name      string
		frame     uint32
		deadFrame uint32
		duration  uint32
		status    object.MonsterStatus
		enemy     bool
		wantRaise bool
	}{
		{name: "expired", frame: 13, deadFrame: 10, duration: 2, enemy: true, wantRaise: true},
		{name: "strict boundary", frame: 12, deadFrame: 10, duration: 2, enemy: true},
		{name: "unsigned wrap", frame: 1, deadFrame: math.MaxUint32 - 1, duration: 2, enemy: true, wantRaise: true},
		{name: "stay dead", frame: 13, deadFrame: 10, duration: 2, status: object.MonStatusStayDead, enemy: true},
		{name: "no enemy", frame: 13, deadFrame: 10, duration: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := new(Server)
			s.SetFrame(test.frame)
			unit := monsterActionTestObject50A910(t)
			update := unit.UpdateDataMonster()
			update.Field137 = test.deadFrame
			update.Field123 = test.duration
			update.StatusFlags = test.status
			if test.enemy {
				update.CurrentEnemy = &Object{}
			}
			var raises int
			if !s.MonsterActionDeadUpdate544EC0(unit, MonsterActionDeadRuntime544D80{
				IsZombie: func(*Object) bool { return true },
				RaiseZombie: func(got *Object) {
					if got != unit {
						t.Fatalf("raise unit = %p, want %p", got, unit)
					}
					raises++
				},
			}) {
				t.Fatal("zombie dead update was not handled")
			}
			if got := raises != 0; got != test.wantRaise {
				t.Fatalf("raised = %v, want %v", got, test.wantRaise)
			}
		})
	}
}

func TestMonsterRaiseZombie534AB0RestoresActionAndFlags(t *testing.T) {
	s := monsterZombieTestServer544CE0(t)
	s.SetFrame(77)
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_DEAD)}
	unit.ObjFlags = object.FlagEnabled | object.FlagDestroyed | object.FlagAllowOverlap |
		object.FlagShort | object.FlagNoCollide | object.FlagDead
	blocked := object.FlagAllowOverlap | object.FlagShort | object.FlagNoCollide | object.FlagDead
	var events []string
	if !s.MonsterRaiseZombie534AB0(unit, MonsterRaiseZombieRuntime534AB0{
		IsZombie: func(got *Object) bool { return got == unit },
		AudioEvent: func(id uint32, got *Object) {
			if id != 469 || got != unit || !unit.ObjFlags.HasAny(blocked) {
				t.Fatalf("raise audio = %d/%p flags=%#x", id, got, unit.ObjFlags)
			}
			events = append(events, "audio")
		},
		SetHealthToMax: func(got *Object) {
			if got != unit || !unit.ObjFlags.HasAny(blocked) {
				t.Fatalf("raise health = %p flags=%#x", got, unit.ObjFlags)
			}
			events = append(events, "health")
		},
	}) {
		t.Fatal("zombie raise was not handled")
	}
	if len(events) != 2 || events[0] != "audio" || events[1] != "health" {
		t.Fatalf("events = %v, want [audio health]", events)
	}
	if update.AIStackInd != 1 || update.AIStack[0].Type() != ai.DEPENDENCY_UNINTERRUPTABLE ||
		update.AIStack[1].Type() != ai.ACTION_GET_UP || update.Field137 != 77 {
		t.Fatalf("raise action stack = %#v, frame=%d", update.GetAIStack(), update.Field137)
	}
	if unit.ObjFlags.HasAny(blocked) || !unit.ObjFlags.HasAny(object.FlagEnabled|object.FlagDestroyed) {
		t.Fatalf("raised zombie flags = %#x", unit.ObjFlags)
	}
}

func TestMonsterRaiseZombie534AB0PreflightsStackCapacity(t *testing.T) {
	s := monsterZombieTestServer544CE0(t)
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	update := unit.UpdateDataMonster()
	update.AIStackInd = int8(len(update.AIStack) - 1)
	update.AIStack[len(update.AIStack)-2].Action = uint32(ai.ACTION_IDLE)
	update.AIStack[len(update.AIStack)-1].Action = uint32(ai.ACTION_DEAD)
	before := *update
	var reason string
	if s.MonsterRaiseZombie534AB0(unit, MonsterRaiseZombieRuntime534AB0{
		IsZombie:       func(*Object) bool { return true },
		AudioEvent:     func(uint32, *Object) {},
		SetHealthToMax: func(*Object) {},
		Unsupported: func(got string, _ *Object) {
			reason = got
		},
	}) {
		t.Fatal("full zombie action stack was accepted")
	}
	if reason != "zombie raise action capacity" || *update != before {
		t.Fatalf("capacity result = %q, mutated=%v", reason, *update != before)
	}
}

func TestScriptRaiseZombie516D00ClearsStayDeadBeforeRaise(t *testing.T) {
	s := monsterZombieTestServer544CE0(t)
	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	unit.ObjFlags = object.FlagDead
	update := unit.UpdateDataMonster()
	update.StatusFlags = object.MonStatusStayDead | object.MonStatusAlert
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_DEAD)
	if !s.ScriptRaiseZombie516D00(unit, MonsterRaiseZombieRuntime534AB0{
		IsZombie: func(*Object) bool {
			if update.StatusFlags.Has(object.MonStatusStayDead) {
				t.Fatal("StayDead was not cleared before zombie predicate")
			}
			return true
		},
		AudioEvent:     func(uint32, *Object) {},
		SetHealthToMax: func(*Object) {},
	}) {
		t.Fatal("script zombie raise was not handled")
	}
	if update.StatusFlags != object.MonStatusAlert || update.AIStackHead().Type() != ai.ACTION_GET_UP {
		t.Fatalf("script raise state = status:%#x action:%s", update.StatusFlags, update.AIStackHead().Type())
	}
}

func TestScriptRaiseZombie516D00NonDeadIsNoOp(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.StatusFlags = object.MonStatusStayDead
	if !new(Server).ScriptRaiseZombie516D00(unit, MonsterRaiseZombieRuntime534AB0{
		IsZombie: func(*Object) bool {
			t.Fatal("non-dead monster reached zombie predicate")
			return false
		},
	}) {
		t.Fatal("non-dead script call was not handled")
	}
	if update.StatusFlags != object.MonStatusStayDead {
		t.Fatalf("non-dead status = %#x", update.StatusFlags)
	}
}

func TestZombieBurnDelete544CE0PlayerOwnerReloadsIndex(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	unit.ObjSubClass = 0x1ff
	unit.PosVec = types.Ptf(10, 20)
	firstPlayer := &Player{PlayerInd: 4}
	secondPlayer := &Player{PlayerInd: 9}
	ownerUpdate := &PlayerUpdateData{Player: firstPlayer}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ownerUpdate)}
	unit.ObjOwner = owner
	var events []string
	if !new(Server).ZombieBurnDelete544CE0(unit, ZombieBurnDeleteRuntime544CE0{
		NetFxShield: func(playerIndex int, got *Object) {
			if playerIndex != 4 || got != unit || uint32(unit.ObjSubClass) != 0x17f {
				t.Fatalf("shield = %d/%p subclass=%#x", playerIndex, got, unit.ObjSubClass)
			}
			events = append(events, "shield")
			ownerUpdate.Player = secondPlayer
		},
		UnmarkMinimap: func(playerIndex int, got *Object, flags uint32) {
			if playerIndex != 9 || got != unit || flags != 1 {
				t.Fatalf("unmark = %d/%p/%d", playerIndex, got, flags)
			}
			events = append(events, "unmark")
		},
		SoloMonsterKillReward: func(got *Object) {
			if got != unit {
				t.Fatalf("reward unit = %p, want %p", got, unit)
			}
			events = append(events, "reward")
			unit.PosVec = types.Ptf(30, 40)
		},
		MakeScorch: func(position types.Pointf, kind int) {
			if position != (types.Ptf(30, 40)) || kind != 1 {
				t.Fatalf("scorch = %v/%d", position, kind)
			}
			events = append(events, "scorch")
		},
		DelayedDelete: func(got *Object) {
			if got != unit {
				t.Fatalf("delete unit = %p, want %p", got, unit)
			}
			events = append(events, "delete")
		},
	}) {
		t.Fatal("zombie burn-delete was not handled")
	}
	want := []string{"shield", "unmark", "reward", "scorch", "delete"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestZombieBurnDelete544CE0PreflightsPlayerEffects(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	unit.ObjSubClass = 0x80
	ownerUpdate := &PlayerUpdateData{Player: &Player{PlayerInd: 3}}
	unit.ObjOwner = &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ownerUpdate)}
	var reason string
	if new(Server).ZombieBurnDelete544CE0(unit, ZombieBurnDeleteRuntime544CE0{
		SoloMonsterKillReward: func(*Object) { t.Fatal("reward called") },
		MakeScorch:            func(types.Pointf, int) { t.Fatal("scorch called") },
		DelayedDelete:         func(*Object) { t.Fatal("delete called") },
		Unsupported: func(got string, _ *Object) {
			reason = got
		},
	}) {
		t.Fatal("missing player-owner effects were accepted")
	}
	if reason != "zombie player owner effects" || uint32(unit.ObjSubClass) != 0x80 {
		t.Fatalf("preflight result = %q, subclass=%#x", reason, unit.ObjSubClass)
	}
}
