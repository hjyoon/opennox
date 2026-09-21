package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerInputAttackTestObject4F9C70 struct {
	name   string
	update *playerInputAttackTestUpdate4F9C70
	frame  uint32
}

type playerInputAttackTestUpdate4F9C70 struct {
	state  PlayerState
	player *playerInputAttackTestPlayer4F9C70
	weapon *playerInputAttackTestWeapon4F9C70
	anim   uint8
	cursor *playerInputAttackTestObject4F9C70
}

type playerInputAttackTestPlayer4F9C70 struct {
	equip uint32
}

type playerInputAttackTestWeapon4F9C70 struct {
	name string
	use  *playerInputAttackTestUse4F9C70
}

type playerInputAttackTestUse4F9C70 struct {
	charge    uint8
	maxCharge uint8
	flags     uint32
}

type playerInputAttackTestRuntime4F9C70 struct {
	calls        []string
	aim          int32
	action       int32
	frame        uint32
	subResult    int32
	setResult    bool
	staminaCost  int32
	onAction     func()
	onSetState   func()
	usedWeapon   *playerInputAttackTestWeapon4F9C70
	adjustAmount uint8
}

func (r *playerInputAttackTestRuntime4F9C70) hooks() playerInputAttackHooks4F9C70[
	*playerInputAttackTestObject4F9C70,
	*playerInputAttackTestUpdate4F9C70,
	*playerInputAttackTestPlayer4F9C70,
	*playerInputAttackTestWeapon4F9C70,
	*playerInputAttackTestUse4F9C70,
] {
	return playerInputAttackHooks4F9C70[
		*playerInputAttackTestObject4F9C70,
		*playerInputAttackTestUpdate4F9C70,
		*playerInputAttackTestPlayer4F9C70,
		*playerInputAttackTestWeapon4F9C70,
		*playerInputAttackTestUse4F9C70,
	]{
		aimsAtEnemy: func(*playerInputAttackTestObject4F9C70) int32 {
			r.calls = append(r.calls, "aim")
			return r.aim
		},
		loadUpdate: func(unit *playerInputAttackTestObject4F9C70) *playerInputAttackTestUpdate4F9C70 {
			r.calls = append(r.calls, "update")
			return unit.update
		},
		loadPlayer: func(update *playerInputAttackTestUpdate4F9C70) *playerInputAttackTestPlayer4F9C70 {
			r.calls = append(r.calls, "player")
			return update.player
		},
		loadWeaponEquip: func(player *playerInputAttackTestPlayer4F9C70) uint32 {
			r.calls = append(r.calls, "equip")
			return player.equip
		},
		actionState: func(*playerInputAttackTestObject4F9C70) int32 {
			r.calls = append(r.calls, "action")
			if r.onAction != nil {
				r.onAction()
			}
			return r.action
		},
		loadWeapon: func(update *playerInputAttackTestUpdate4F9C70) *playerInputAttackTestWeapon4F9C70 {
			r.calls = append(r.calls, "weapon:"+update.weapon.name)
			return update.weapon
		},
		loadWeaponUseData: func(weapon *playerInputAttackTestWeapon4F9C70) *playerInputAttackTestUse4F9C70 {
			r.calls = append(r.calls, "use-data:"+weapon.name)
			return weapon.use
		},
		loadCharge: func(use *playerInputAttackTestUse4F9C70) uint8 {
			r.calls = append(r.calls, "charge")
			return use.charge
		},
		loadMaxCharge: func(use *playerInputAttackTestUse4F9C70) uint8 {
			r.calls = append(r.calls, "max-charge")
			return use.maxCharge
		},
		subStamina: func(_ *playerInputAttackTestObject4F9C70, amount int32) int32 {
			r.calls = append(r.calls, fmt.Sprintf("sub:%d", amount))
			return r.subResult
		},
		loadWeaponFlags: func(use *playerInputAttackTestUse4F9C70) uint32 {
			r.calls = append(r.calls, "flags")
			return use.flags
		},
		storeWeaponFlags: func(use *playerInputAttackTestUse4F9C70, flags uint32) {
			r.calls = append(r.calls, fmt.Sprintf("store-flags:%#x", flags))
			use.flags = flags
		},
		loadFrame: func() uint32 {
			r.calls = append(r.calls, "frame")
			return r.frame
		},
		storeAttackFrame: func(unit *playerInputAttackTestObject4F9C70, frame uint32) {
			r.calls = append(r.calls, fmt.Sprintf("store-frame:%#x", frame))
			unit.frame = frame
		},
		storeAnimFrame: func(update *playerInputAttackTestUpdate4F9C70, frame uint8) {
			r.calls = append(r.calls, fmt.Sprintf("store-anim:%d", frame))
			update.anim = frame
		},
		setState: func(_ *playerInputAttackTestObject4F9C70, state PlayerState) bool {
			r.calls = append(r.calls, fmt.Sprintf("set:%d", state))
			if r.onSetState != nil {
				r.onSetState()
			}
			return r.setResult
		},
		useByNetCode: func(_ *playerInputAttackTestObject4F9C70, weapon *playerInputAttackTestWeapon4F9C70) int32 {
			r.calls = append(r.calls, "use:"+weapon.name)
			r.usedWeapon = weapon
			return -1
		},
		loadState: func(update *playerInputAttackTestUpdate4F9C70) PlayerState {
			r.calls = append(r.calls, "state")
			return update.state
		},
		weaponStamina: func(equip uint32) int32 {
			r.calls = append(r.calls, fmt.Sprintf("stamina:%#x", equip))
			return r.staminaCost
		},
		adjustStamina: func(_ *playerInputAttackTestObject4F9C70, amount uint8) {
			r.calls = append(r.calls, fmt.Sprintf("adjust:%d", amount))
			r.adjustAmount = amount
		},
		buffOff: func(_ *playerInputAttackTestObject4F9C70, buff int32) int32 {
			r.calls = append(r.calls, fmt.Sprintf("buff:%d", buff))
			return -1
		},
		cancelDurSpell: func(spell int32, _ *playerInputAttackTestObject4F9C70) int32 {
			r.calls = append(r.calls, fmt.Sprintf("cancel:%d", spell))
			return -1
		},
	}
}

func TestPlayerInputAttack4F9C70NilAndAimGate(t *testing.T) {
	runtime := &playerInputAttackTestRuntime4F9C70{aim: 1}
	playerInputAttack4F9C70((*playerInputAttackTestObject4F9C70)(nil), runtime.hooks())
	if runtime.calls != nil {
		t.Fatalf("nil calls = %v, want nil", runtime.calls)
	}

	runtime.aim = 0
	unit := &playerInputAttackTestObject4F9C70{}
	playerInputAttack4F9C70(unit, runtime.hooks())
	if want := []string{"aim"}; !reflect.DeepEqual(runtime.calls, want) {
		t.Fatalf("aim-gated calls = %v, want %v", runtime.calls, want)
	}
}

func TestPlayerInputAttack4F9C70UnarmedStateGateSkipsCleanup(t *testing.T) {
	for _, tc := range []struct {
		name      string
		state     PlayerState
		wantCalls []string
	}{
		{name: "transition", state: PlayerState0, wantCalls: []string{"aim", "update", "player", "equip", "state", "set:1"}},
		{name: "already attacking", state: PlayerState1, wantCalls: []string{"aim", "update", "player", "equip", "state"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runtime := &playerInputAttackTestRuntime4F9C70{aim: 1, setResult: true}
			update := &playerInputAttackTestUpdate4F9C70{
				state: tc.state, player: &playerInputAttackTestPlayer4F9C70{},
			}
			playerInputAttack4F9C70(&playerInputAttackTestObject4F9C70{update: update}, runtime.hooks())
			if !reflect.DeepEqual(runtime.calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", runtime.calls, tc.wantCalls)
			}
		})
	}
}

func TestPlayerInputAttack4F9C70ReadyRangedReloadsWeaponAfterSetState(t *testing.T) {
	oldWeapon := &playerInputAttackTestWeapon4F9C70{
		name: "old", use: &playerInputAttackTestUse4F9C70{charge: 1},
	}
	newWeapon := &playerInputAttackTestWeapon4F9C70{name: "new"}
	update := &playerInputAttackTestUpdate4F9C70{
		player: &playerInputAttackTestPlayer4F9C70{equip: 0x10000},
		weapon: oldWeapon,
		anim:   0xff,
	}
	unit := &playerInputAttackTestObject4F9C70{update: update}
	runtime := &playerInputAttackTestRuntime4F9C70{
		aim: 1, action: 31, frame: 0x89abcdef, setResult: true,
		onSetState: func() { update.weapon = newWeapon },
	}
	playerInputAttack4F9C70(unit, runtime.hooks())

	wantCalls := []string{
		"aim", "update", "player", "equip", "action", "weapon:old", "use-data:old", "charge",
		"frame", "store-frame:0x89abcdef", "store-anim:0", "set:1", "weapon:new", "use:new",
		"buff:0", "buff:23", "cancel:67",
	}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
	if runtime.usedWeapon != newWeapon || unit.frame != runtime.frame || update.anim != 0 {
		t.Fatalf("weapon/frame/anim = %p/%#x/%d, want %p/%#x/0", runtime.usedWeapon, unit.frame, update.anim, newWeapon, runtime.frame)
	}
}

func TestPlayerInputAttack4F9C70ChargedRangedPreservesFullFlags(t *testing.T) {
	use := &playerInputAttackTestUse4F9C70{maxCharge: 9, flags: 0xa5b6c4d0}
	weapon := &playerInputAttackTestWeapon4F9C70{name: "charged", use: use}
	update := &playerInputAttackTestUpdate4F9C70{
		player: &playerInputAttackTestPlayer4F9C70{equip: 0x4000000},
		weapon: weapon,
		anim:   0xff,
	}
	unit := &playerInputAttackTestObject4F9C70{update: update}
	runtime := &playerInputAttackTestRuntime4F9C70{
		aim: 1, action: 31, frame: 0xfedcba98, subResult: 1, setResult: true,
	}
	playerInputAttack4F9C70(unit, runtime.hooks())

	wantCalls := []string{
		"aim", "update", "player", "equip", "action", "weapon:charged", "use-data:charged", "charge", "max-charge",
		"sub:45", "flags", "store-flags:0xa5b6c4d2", "frame", "store-frame:0xfedcba98", "store-anim:0", "set:1",
		"buff:0", "buff:23", "cancel:67",
	}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
	if use.flags != 0xa5b6c4d2 {
		t.Fatalf("flags = %#x, want %#x", use.flags, uint32(0xa5b6c4d2))
	}
}

func TestPlayerInputAttack4F9C70OrdinaryReloadsEquipAndRefunds(t *testing.T) {
	initialPlayer := &playerInputAttackTestPlayer4F9C70{equip: 0x10000}
	livePlayer := &playerInputAttackTestPlayer4F9C70{equip: 0x12345678}
	update := &playerInputAttackTestUpdate4F9C70{state: PlayerState0, player: initialPlayer}
	unit := &playerInputAttackTestObject4F9C70{update: update}
	runtime := &playerInputAttackTestRuntime4F9C70{
		aim: 1, action: 29, frame: 7, subResult: 1, staminaCost: 300,
		onAction: func() { update.player = livePlayer },
	}
	playerInputAttack4F9C70(unit, runtime.hooks())

	wantCalls := []string{
		"aim", "update", "player", "equip", "action", "state", "player", "equip", "stamina:0x12345678", "sub:300",
		"frame", "store-frame:0x7", "store-anim:0", "set:1", "adjust:212",
		"buff:0", "buff:23", "cancel:67",
	}
	if !reflect.DeepEqual(runtime.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runtime.calls, wantCalls)
	}
	if runtime.adjustAmount != uint8(212) {
		t.Fatalf("refund = %d, want 212", runtime.adjustAmount)
	}
}

func TestPlayerInputAttack4F9C70OrdinaryFailureStillCleansUp(t *testing.T) {
	for _, tc := range []struct {
		name      string
		state     PlayerState
		subResult int32
		wantCalls []string
	}{
		{
			name: "state one", state: PlayerState1,
			wantCalls: []string{"aim", "update", "player", "equip", "state", "buff:0", "buff:23", "cancel:67"},
		},
		{
			name: "stamina rejected", state: PlayerState0,
			wantCalls: []string{"aim", "update", "player", "equip", "state", "player", "equip", "stamina:0x4", "sub:6", "buff:0", "buff:23", "cancel:67"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			player := &playerInputAttackTestPlayer4F9C70{equip: 4}
			update := &playerInputAttackTestUpdate4F9C70{state: tc.state, player: player}
			runtime := &playerInputAttackTestRuntime4F9C70{
				aim: 1, subResult: tc.subResult, staminaCost: 6,
			}
			playerInputAttack4F9C70(&playerInputAttackTestObject4F9C70{update: update}, runtime.hooks())
			if !reflect.DeepEqual(runtime.calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", runtime.calls, tc.wantCalls)
			}
		})
	}
}

func TestPlayerAimsAtEnemy4F9DC0ShortCircuitOrder(t *testing.T) {
	type testCase struct {
		name      string
		unitNil   bool
		cursorNil bool
		enemy     bool
		quest     bool
		want      int32
		wantCalls []string
	}
	for _, tc := range []testCase{
		{name: "nil unit", unitNil: true, want: 0},
		{name: "nil cursor", cursorNil: true, want: 1, wantCalls: []string{"update", "cursor"}},
		{name: "enemy", enemy: true, want: 1, wantCalls: []string{"update", "cursor", "enemy"}},
		{name: "quest fallback", quest: true, want: 1, wantCalls: []string{"update", "cursor", "enemy", "quest"}},
		{name: "friendly non-quest", want: 0, wantCalls: []string{"update", "cursor", "enemy", "quest"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			cursor := &playerInputAttackTestObject4F9C70{name: "cursor"}
			if tc.cursorNil {
				cursor = nil
			}
			unit := &playerInputAttackTestObject4F9C70{
				update: &playerInputAttackTestUpdate4F9C70{cursor: cursor},
			}
			if tc.unitNil {
				unit = nil
			}
			got := playerAimsAtEnemy4F9DC0(unit, playerAimsAtEnemyHooks4F9DC0[
				*playerInputAttackTestObject4F9C70, *playerInputAttackTestUpdate4F9C70,
			]{
				loadUpdate: func(unit *playerInputAttackTestObject4F9C70) *playerInputAttackTestUpdate4F9C70 {
					calls = append(calls, "update")
					return unit.update
				},
				loadCursor: func(update *playerInputAttackTestUpdate4F9C70) *playerInputAttackTestObject4F9C70 {
					calls = append(calls, "cursor")
					return update.cursor
				},
				isEnemy: func(_, _ *playerInputAttackTestObject4F9C70) bool {
					calls = append(calls, "enemy")
					return tc.enemy
				},
				questMode: func() bool {
					calls = append(calls, "quest")
					return tc.quest
				},
			})
			if got != tc.want || !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("result/calls = %d/%v, want %d/%v", got, calls, tc.want, tc.wantCalls)
			}
		})
	}
}

func TestPlayerInputAttack4F9C70DoesNotGuardUpdate(t *testing.T) {
	runtime := &playerInputAttackTestRuntime4F9C70{aim: 1}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update did not preserve the original fault contract")
		}
	}()
	playerInputAttack4F9C70(&playerInputAttackTestObject4F9C70{}, runtime.hooks())
}
