package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterEscortNoopHooks546430() monsterActionEscortHooks546430 {
	return monsterActionEscortHooks546430{
		resolveTarget:    func() *Object { return nil },
		canAttackAtWill:  func() bool { return false },
		mediumAggression: func() bool { return false },
		noticeThreat:     func() int { return 0 },
		lookAtDamager:    func() bool { return false },
		interestingSound: func() int { return 0 },
		hasAntiMagic:     func() bool { return true },
		healSomeone:      func() int { return 0 },
		frame:            func() uint32 { return 0 },
		push: func(action ai.ActionType) *AIStackItem {
			return &AIStackItem{Action: uint32(action)}
		},
		pop: func() int { return 0 },
	}
}

func monsterEscortTestUnit546430(t *testing.T, target *Object) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	unit.PosVec = types.Ptf(0, 0)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_ESCORT)
	update.Field329 = 10
	monsterEscortSetTarget546430(&update.AIStack[0], target)
	return unit
}

func monsterEscortSetNameTest546600(unit *Object, name string) []byte {
	buf := monsterEscortNameBuffer546600(unit.UpdateDataMonster())
	clear(buf)
	copy(buf, name)
	return buf
}

func TestMonsterEscortScriptNameMatches546600(t *testing.T) {
	for _, tc := range []struct {
		id, name string
		want     bool
	}{
		{id: "Shopkeeper", name: "Shopkeeper", want: true},
		{id: "War01A:Shopkeeper", name: "Shopkeeper", want: true},
		{id: "War01A:Room:Shopkeeper", name: "Room:Shopkeeper", want: false},
		{id: "War01A:Room:Shopkeeper", name: "War01A:Room:Shopkeeper", want: true},
		{id: "War01A:Room:Shopkeeper", name: "Shopkeeper", want: false},
		{id: "", name: "", want: false},
	} {
		if got := monsterEscortScriptNameMatches546600(tc.id, tc.name); got != tc.want {
			t.Fatalf("match(%q, %q) = %v, want %v", tc.id, tc.name, got, tc.want)
		}
	}
}

func TestMonsterGetObjEscortName546600ExplicitAndOwner(t *testing.T) {
	unit := monsterEscortTestUnit546430(t, nil)
	wanted := new(Object)
	buf := monsterEscortSetNameTest546600(unit, "Shopkeeper")
	var lookup string
	got := monsterGetObjEscortName546600(unit, monsterEscortNameHooks546600{
		objectByID: func(name string) *Object {
			lookup = name
			return wanted
		},
	})
	if got != wanted || lookup != "Shopkeeper" {
		t.Fatalf("explicit lookup = (%p, %q), want (%p, Shopkeeper)", got, lookup, wanted)
	}
	if buf[0] != 0 || buf[1] != 'h' {
		t.Fatalf("name prefix after lookup = {%#x, %#x}, want {0, 'h'}", buf[0], buf[1])
	}

	owner := new(Object)
	buf = monsterEscortSetNameTest546600(unit, "**OWNER**")
	got = monsterGetObjEscortName546600(unit, monsterEscortNameHooks546600{
		owner: func() *Object { return owner },
	})
	if got != owner || buf[0] != 0 || buf[1] != '*' {
		t.Fatalf("owner lookup = %p, name prefix {%#x, %#x}", got, buf[0], buf[1])
	}
}

func TestMonsterGetObjEscortName546600PlayerCallOrder(t *testing.T) {
	unit := monsterEscortTestUnit546430(t, nil)
	buf := monsterEscortSetNameTest546600(unit, "**PLAYER**")
	p1, p2, p3 := new(Object), new(Object), new(Object)
	next := map[*Object]*Object{p1: p2, p2: p3, p3: nil}
	labels := map[*Object]string{p1: "p1", p2: "p2", p3: "p3"}
	var events []string
	got := monsterGetObjEscortName546600(unit, monsterEscortNameHooks546600{
		firstPlayer: func() *Object {
			events = append(events, "first")
			return p1
		},
		nextPlayer: func(player *Object) *Object {
			events = append(events, "next:"+labels[player])
			return next[player]
		},
		random: func(minimum, maximum int) int {
			events = append(events, "random")
			if minimum != 0 || maximum != 2 {
				t.Fatalf("random bounds = [%d,%d], want [0,2]", minimum, maximum)
			}
			return 1
		},
	})
	wantEvents := []string{"first", "next:p1", "next:p2", "next:p3", "random", "first", "next:p1"}
	if got != p2 || !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("player lookup = %p events %v, want %p %v", got, events, p2, wantEvents)
	}
	if buf[0] != 0 || buf[1] != '*' {
		t.Fatalf("name prefix after player lookup = {%#x, %#x}", buf[0], buf[1])
	}
}

func TestMonsterGetObjEscortName546600EmptyPlayerListStillDraws(t *testing.T) {
	unit := monsterEscortTestUnit546430(t, nil)
	monsterEscortSetNameTest546600(unit, "**PLAYER**")
	draws := 0
	got := monsterGetObjEscortName546600(unit, monsterEscortNameHooks546600{
		firstPlayer: func() *Object { return nil },
		nextPlayer:  func(*Object) *Object { t.Fatal("next player called"); return nil },
		random: func(minimum, maximum int) int {
			draws++
			if minimum != 0 || maximum != -1 {
				t.Fatalf("random bounds = [%d,%d], want [0,-1]", minimum, maximum)
			}
			return 0
		},
	})
	if got != nil || draws != 1 {
		t.Fatalf("empty player lookup = %p draws %d, want nil/1", got, draws)
	}
}

func TestMonsterActionEscort546430ResolvesTargetOrPops(t *testing.T) {
	t.Run("resolved", func(t *testing.T) {
		target := &Object{PosVec: types.Ptf(12.5, -8.25)}
		unit := monsterEscortTestUnit546430(t, nil)
		hooks := monsterEscortNoopHooks546430()
		hooks.resolveTarget = func() *Object { return target }
		if !monsterActionEscort546430(unit, hooks) {
			t.Fatal("valid escort action was rejected")
		}
		head := unit.UpdateDataMonster().AIStackHead()
		if head.ArgObj(2) != target || head.ArgPos(0) != target.PosVec {
			t.Fatalf("resolved head = target %p position %v", head.ArgObj(2), head.ArgPos(0))
		}
		if head.Args[2] <= uintptr(^uint32(0)) {
			t.Fatalf("target pointer = %#x, want native high address", head.Args[2])
		}
	})

	t.Run("missing", func(t *testing.T) {
		unit := monsterEscortTestUnit546430(t, nil)
		hooks := monsterEscortNoopHooks546430()
		pops := 0
		hooks.pop = func() int { pops++; return 0 }
		if !monsterActionEscort546430(unit, hooks) || pops != 1 {
			t.Fatalf("missing target result = pops %d, want 1", pops)
		}
	})
}

func TestMonsterActionEscort546430AggressiveEnemyReloadsAfterPush(t *testing.T) {
	target := &Object{PosVec: types.Ptf(10, 0)}
	enemyBefore := &Object{PosVec: types.Ptf(1, 2)}
	enemyAfter := &Object{PosVec: types.Ptf(30, 40)}
	unit := monsterEscortTestUnit546430(t, target)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = enemyBefore
	hooks := monsterEscortNoopHooks546430()
	hooks.canAttackAtWill = func() bool { return true }
	hooks.frame = func() uint32 { return 0x89abcdef }
	var fight *AIStackItem
	hooks.push = func(action ai.ActionType) *AIStackItem {
		if action != ai.ACTION_FIGHT {
			t.Fatalf("pushed %s, want ACTION_FIGHT", action)
		}
		update.CurrentEnemy = enemyAfter
		fight = &AIStackItem{Action: uint32(action)}
		return fight
	}
	monsterActionEscort546430(unit, hooks)
	if fight == nil || fight.ArgPos(0) != enemyAfter.PosVec || fight.ArgU32(2) != 0x89abcdef {
		t.Fatalf("fight args = position %v frame %#x, want %v/0x89abcdef", fight.ArgPos(0), fight.ArgU32(2), enemyAfter.PosVec)
	}
}

func TestMonsterActionEscort546430MediumThreatCallOrder(t *testing.T) {
	unit := monsterEscortTestUnit546430(t, &Object{PosVec: types.Ptf(10, 0)})
	hooks := monsterEscortNoopHooks546430()
	var events []string
	hooks.canAttackAtWill = func() bool { events = append(events, "aggressive"); return false }
	hooks.mediumAggression = func() bool { events = append(events, "medium"); return true }
	hooks.noticeThreat = func() int { events = append(events, "notice"); return 1 }
	hooks.lookAtDamager = func() bool { events = append(events, "damager"); return true }
	monsterActionEscort546430(unit, hooks)
	want := []string{"aggressive", "medium", "notice"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestMonsterActionEscort546430NearBranches(t *testing.T) {
	t.Run("passive-heals-at-inclusive-radius", func(t *testing.T) {
		unit := monsterEscortTestUnit546430(t, &Object{PosVec: types.Ptf(40, 0)})
		hooks := monsterEscortNoopHooks546430()
		var events []string
		hooks.canAttackAtWill = func() bool { events = append(events, "aggressive"); return false }
		hooks.mediumAggression = func() bool { events = append(events, "medium"); return false }
		hooks.interestingSound = func() int { events = append(events, "sound"); return 1 }
		hooks.hasAntiMagic = func() bool { events = append(events, "anti-magic"); return false }
		hooks.healSomeone = func() int { events = append(events, "heal"); return 0 }
		monsterActionEscort546430(unit, hooks)
		want := []string{"aggressive", "medium", "aggressive", "anti-magic", "heal"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})

	t.Run("aggressive-sound-short-circuits-heal", func(t *testing.T) {
		unit := monsterEscortTestUnit546430(t, &Object{PosVec: types.Ptf(40, 0)})
		hooks := monsterEscortNoopHooks546430()
		var events []string
		hooks.canAttackAtWill = func() bool { events = append(events, "aggressive"); return true }
		hooks.lookAtDamager = func() bool { events = append(events, "damager"); return false }
		hooks.interestingSound = func() int { events = append(events, "sound"); return 1 }
		hooks.hasAntiMagic = func() bool { events = append(events, "anti-magic"); return false }
		monsterActionEscort546430(unit, hooks)
		want := []string{"aggressive", "damager", "aggressive", "sound"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})
}

func TestMonsterActionEscort546430FarPushOrderAndLiveReloads(t *testing.T) {
	targetBefore := &Object{PosVec: types.Ptf(100, 0)}
	targetForDependency := &Object{PosVec: types.Ptf(200, 1)}
	targetForMove := &Object{PosVec: types.Ptf(300, 2)}
	unit := monsterEscortTestUnit546430(t, targetBefore)
	update := unit.UpdateDataMonster()
	escort := update.AIStackHead()
	hooks := monsterEscortNoopHooks546430()
	var events []string
	hooks.canAttackAtWill = func() bool { events = append(events, "aggressive"); return true }
	hooks.mediumAggression = func() bool { events = append(events, "medium"); return false }
	hooks.lookAtDamager = func() bool { events = append(events, "damager"); return false }
	items := make(map[ai.ActionType]*AIStackItem)
	hooks.push = func(action ai.ActionType) *AIStackItem {
		events = append(events, "push:"+action.String())
		item := &AIStackItem{Action: uint32(action)}
		items[action] = item
		switch action {
		case ai.DEPENDENCY_OBJECT_FARTHER_THAN:
			update.Field329 = 22
			monsterEscortSetTarget546430(escort, targetForDependency)
		case ai.ACTION_MOVE_TO:
			monsterEscortSetTarget546430(escort, targetForMove)
		}
		return item
	}
	monsterActionEscort546430(unit, hooks)
	wantEvents := []string{
		"aggressive", "damager", "medium", "aggressive",
		"push:" + ai.DEPENDENCY_NOT_UNDER_ATTACK.String(),
		"aggressive", "push:" + ai.DEPENDENCY_NO_VISIBLE_ENEMY.String(),
		"push:" + ai.DEPENDENCY_OBJECT_FARTHER_THAN.String(),
		"push:" + ai.ACTION_MOVE_TO.String(),
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	farther := items[ai.DEPENDENCY_OBJECT_FARTHER_THAN]
	if farther.ArgF32(0) != 22 || farther.ArgObj(2) != targetForDependency {
		t.Fatalf("farther args = radius %v target %p", farther.ArgF32(0), farther.ArgObj(2))
	}
	move := items[ai.ACTION_MOVE_TO]
	if move.ArgPos(0) != targetForMove.PosVec || move.ArgObj(2) != targetForMove {
		t.Fatalf("move args = position %v target %p", move.ArgPos(0), move.ArgObj(2))
	}
	if move.Args[2] <= uintptr(^uint32(0)) {
		t.Fatalf("move target pointer = %#x, want native high address", move.Args[2])
	}
}

func TestMonsterActionEscort546430NaNDistanceTakesFarBranch(t *testing.T) {
	target := &Object{PosVec: types.Ptf(float32(math.NaN()), 0)}
	unit := monsterEscortTestUnit546430(t, target)
	hooks := monsterEscortNoopHooks546430()
	var actions []ai.ActionType
	hooks.push = func(action ai.ActionType) *AIStackItem {
		actions = append(actions, action)
		return &AIStackItem{Action: uint32(action)}
	}
	monsterActionEscort546430(unit, hooks)
	want := []ai.ActionType{ai.DEPENDENCY_OBJECT_FARTHER_THAN, ai.ACTION_MOVE_TO}
	if !reflect.DeepEqual(actions, want) {
		t.Fatalf("actions = %v, want %v", actions, want)
	}
}

func TestMonsterActionEscort546430RejectsMalformedAdmission(t *testing.T) {
	unit := monsterEscortTestUnit546430(t, nil)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_GUARD)
	if monsterActionEscort546430(unit, monsterActionEscortHooks546430{}) {
		t.Fatal("wrong action head was accepted")
	}
	if uintptr(unit.UpdateData) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(unit)) <= uintptr(^uint32(0)) {
		t.Fatal("test did not exercise native high addresses")
	}
}
