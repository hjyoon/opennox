package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterAnimationIndex533790(t *testing.T) {
	tests := []struct {
		name                 string
		action               ai.ActionType
		status               object.MonsterStatus
		mimic, plant, zombie bool
		enemy                bool
		want                 int
	}{
		{name: "idle", action: ai.ACTION_IDLE, want: 8},
		{name: "move", action: ai.ACTION_MOVE_TO, want: 12},
		{name: "running move", action: ai.ACTION_MOVE_TO, status: object.MonStatusRunning, want: 13},
		{name: "melee", action: ai.ACTION_MELEE_ATTACK, want: 1},
		{name: "missile", action: ai.ACTION_MISSILE_ATTACK, want: 3},
		{name: "spell", action: ai.ACTION_CAST_DURATION_SPELL, want: 7},
		{name: "block", action: ai.ACTION_WEAPON_BLOCK, want: 5},
		{name: "block finish", action: ai.ACTION_BLOCK_FINISH, want: 6},
		{name: "flee", action: ai.ACTION_FLEE, want: 13},
		{name: "dying", action: ai.ACTION_DYING, want: 9},
		{name: "dead", action: ai.ACTION_DEAD, want: 10},
		{name: "morph in", action: ai.ACTION_MORPH_INTO_CHEST, want: 14},
		{name: "morph back", action: ai.ACTION_MORPH_BACK_TO_SELF, want: 15},
		{name: "morphed mimic idle", action: ai.ACTION_IDLE, status: object.MonStatusMorphed, mimic: true, want: 0},
		{name: "plant idle", action: ai.ACTION_IDLE, plant: true, want: 14},
		{name: "plant with enemy", action: ai.ACTION_IDLE, plant: true, enemy: true, want: 8},
		{name: "burning zombie death", action: ai.ACTION_DYING, status: object.MonStatusOnFire, zombie: true, want: 15},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			update := &MonsterUpdateData{AIStackInd: 0, StatusFlags: tc.status}
			update.AIStack[0].Action = uint32(tc.action)
			if tc.enemy {
				update.CurrentEnemy = &Object{}
			}
			if got := monsterAnimationIndex533790(update, tc.mimic, tc.plant, tc.zombie); got != tc.want {
				t.Fatalf("animation = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestMonsterUpdateNonNPCAnim50A850(t *testing.T) {
	animations := new([16]MonsterAnim)
	animations[8] = MonsterAnim{frames: 2, field10: 1}
	update := &MonsterUpdateData{Field119: animations}

	monsterUpdateNonNPCAnim50A850(update, 8)
	if update.Field120_0 != 2 || update.Field120_1 != 0 || update.Field120_2 != 1 || update.Field120_3 != 0 {
		t.Fatalf("first tick = %d/%d/%d/%d, want 2/0/1/0", update.Field120_0, update.Field120_1, update.Field120_2, update.Field120_3)
	}
	monsterUpdateNonNPCAnim50A850(update, 8)
	monsterUpdateNonNPCAnim50A850(update, 8)
	monsterUpdateNonNPCAnim50A850(update, 8)
	if update.Field120_1 != 1 || update.Field120_2 != 0 || update.Field120_3 != 1 {
		t.Fatalf("terminal tick = frame:%d sub:%d done:%d, want 1/0/1", update.Field120_1, update.Field120_2, update.Field120_3)
	}
}

func TestMonsterUpdateNonNPCAnim50A850Loops(t *testing.T) {
	animations := new([16]MonsterAnim)
	animations[8] = MonsterAnim{frames: 1, loop: 1}
	update := &MonsterUpdateData{Field119: animations}
	monsterUpdateNonNPCAnim50A850(update, 8)
	if update.Field120_1 != 0 || update.Field120_2 != 0 || update.Field120_3 != 0 {
		t.Fatalf("loop state = %d/%d/%d, want 0/0/0", update.Field120_1, update.Field120_2, update.Field120_3)
	}
}

func TestMonsterUpdateNonNPCAnim50A850ZeroFrames(t *testing.T) {
	animations := new([16]MonsterAnim)
	update := &MonsterUpdateData{Field119: animations}
	monsterUpdateNonNPCAnim50A850(update, 8)
	if update.Field120_0 != 0 || update.Field120_3 != 1 {
		t.Fatalf("zero-frame state = frames:%d done:%d, want 0/1", update.Field120_0, update.Field120_3)
	}
}

func TestMonsterUpdateNPCAnim50A850GuardCompletesEmptyAnimation(t *testing.T) {
	s := &Server{}
	update := &MonsterUpdateData{AIStackInd: 0}
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	unit := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterNPC),
		UpdateData:  unsafe.Pointer(update),
	}

	if !s.MonsterUpdateNPCAnim50A850(unit) {
		t.Fatal("NPC animation was not handled natively")
	}
	if update.Field120_0 != 0 || update.Field120_3 != 1 {
		t.Fatalf("guard animation = frames:%d done:%d, want 0/1", update.Field120_0, update.Field120_3)
	}
}

func TestMonsterUpdateNPCAnim50A850UsesPlayerFrames(t *testing.T) {
	s := &Server{}
	s.Types.playerAnimFrames = make([][2]int, len(playerAnimTypes))
	s.Types.playerAnimFrames[21] = [2]int{3, 1}
	update := &MonsterUpdateData{AIStackInd: 0}
	update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
	unit := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterNPC),
		UpdateData:  unsafe.Pointer(update),
	}

	for range 6 {
		if !s.MonsterUpdateNPCAnim50A850(unit) {
			t.Fatal("NPC animation was not handled natively")
		}
	}
	if update.Field120_0 != 3 || update.Field120_1 != 2 || update.Field120_2 != 0 || update.Field120_3 != 1 {
		t.Fatalf("spell animation = %d/%d/%d/%d, want 3/2/0/1", update.Field120_0, update.Field120_1, update.Field120_2, update.Field120_3)
	}
}

func TestMonsterUpdateNPCAnim50A850AttackCompletesImmediately(t *testing.T) {
	s := &Server{}
	update := &MonsterUpdateData{AIStackInd: 0, Field120_0: 7, Field120_1: 3, Field120_2: 2}
	update.AIStack[0].Action = uint32(ai.ACTION_MELEE_ATTACK)
	unit := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterNPC),
		UpdateData:  unsafe.Pointer(update),
	}

	if !s.MonsterUpdateNPCAnim50A850(unit) {
		t.Fatal("NPC animation was not handled natively")
	}
	if update.Field120_0 != 7 || update.Field120_1 != 3 || update.Field120_2 != 2 || update.Field120_3 != 1 {
		t.Fatalf("attack animation = %d/%d/%d/%d, want 7/3/2/1", update.Field120_0, update.Field120_1, update.Field120_2, update.Field120_3)
	}
}
