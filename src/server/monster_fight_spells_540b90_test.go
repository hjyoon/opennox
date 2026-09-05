package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterFightSpellTestUnit540B90(t *testing.T) (*Object, *MonsterUpdateData) {
	t.Helper()
	unit := fightMonsterTestObject531E20(t)
	unit.ObjFlags = object.FlagEnabled
	update := unit.UpdateDataMonster()
	update.StatusFlags |= object.MonStatusCanCastSpells
	return unit, update
}

func monsterFightSetSpellFlag540B90(update *MonsterUpdateData, id spell.ID, mask uint32) {
	unsafe.Slice(&update.Field373, SpellsMax-1)[int(id)-1] = mask
}

func TestMonsterFightSpellFlag540B90ExactFieldSpan(t *testing.T) {
	update := new(MonsterUpdateData)
	update.Field373 = 0x11111111
	update.Field508 = 0x88888888
	if got := monsterFightSpellFlag540B90(update, 1); got != 0x11111111 {
		t.Fatalf("spell 1 flag = %#x", got)
	}
	if got := monsterFightSpellFlag540B90(update, 136); got != 0x88888888 {
		t.Fatalf("spell 136 flag = %#x", got)
	}
	for _, id := range []spell.ID{-1, 0, 137, 200} {
		if got := monsterFightSpellFlag540B90(update, id); got != 0 {
			t.Fatalf("spell %d flag = %#x, want 0", id, got)
		}
	}
}

func TestMonsterFightSpellIsSummon540D20InclusiveBounds(t *testing.T) {
	for _, tc := range []struct {
		id   spell.ID
		want bool
	}{
		{id: 74, want: false},
		{id: 75, want: true},
		{id: 114, want: true},
		{id: 115, want: false},
		{id: spell.SPELL_SUMMON_CREATURE, want: false},
	} {
		if got := monsterFightSpellIsSummon540D20(tc.id); got != tc.want {
			t.Errorf("spell %d summon = %t, want %t", tc.id, got, tc.want)
		}
	}
}

func TestMonsterFightSpellReady540B90Gate(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*Object, *MonsterUpdateData)
		frame uint32
		limit uint32
		want  bool
	}{
		{name: "ready at deadline", frame: 100, limit: 100, want: true},
		{name: "disabled", setup: func(unit *Object, _ *MonsterUpdateData) { unit.ObjFlags &^= object.FlagEnabled }, frame: 100, limit: 100},
		{name: "cannot cast", setup: func(_ *Object, update *MonsterUpdateData) { update.StatusFlags &^= object.MonStatusCanCastSpells }, frame: 100, limit: 100},
		{name: "before deadline", frame: 99, limit: 100},
		{name: "cast object", setup: func(_ *Object, update *MonsterUpdateData) {
			update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
		}, frame: 100, limit: 100},
		{name: "cast location", setup: func(_ *Object, update *MonsterUpdateData) {
			update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_LOCATION)
		}, frame: 100, limit: 100},
		{name: "cast duration", setup: func(_ *Object, update *MonsterUpdateData) {
			update.AIStack[0].Action = uint32(ai.ACTION_CAST_DURATION_SPELL)
		}, frame: 100, limit: 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, update := monsterFightSpellTestUnit540B90(t)
			if tc.setup != nil {
				tc.setup(unit, update)
			}
			if got := monsterFightSpellReady540B90(unit, update, tc.limit, func() uint32 { return tc.frame }); got != tc.want {
				t.Fatalf("ready = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestMonsterFightBuffSelf540B90ActiveEnchantSuppressesSet(t *testing.T) {
	unit, update := monsterFightSpellTestUnit540B90(t)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_FIREBALL, monsterFightSelfSpellMask540B90)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_HASTE, monsterFightSelfSpellMask540B90)
	unit.Buffs = uint32(1) << ENCHANT_HASTED
	if monsterFightBuffSelf540B90(unit, monsterFightSpellHooks540B90{
		frame:        func() uint32 { return 100 },
		spellAllowed: func(spell.ID) bool { return true },
		random:       func(int, int) int { t.Fatal("RNG consumed for active enchant"); return 0 },
		cast:         func(*Object, spell.ID, *Object) { t.Fatal("cast with active enchant") },
	}) {
		t.Fatal("active self enchant selected a spell")
	}
}

func TestMonsterFightBuffSelf540B90SelectionCooldownAndCachedUpdate(t *testing.T) {
	unit, update := monsterFightSpellTestUnit540B90(t)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_HASTE, monsterFightSelfSpellMask540B90)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_PROTECTION_FROM_FIRE, monsterFightSelfSpellMask540B90)
	update.Field365 = 90
	update.Field364_0 = 11
	update.Field364_2 = 13
	replacement := new(MonsterUpdateData)
	frames := []uint32{100, 140}
	frameInd := 0
	randomInd := 0
	hooks := monsterFightSpellHooks540B90{
		frame: func() uint32 {
			v := frames[frameInd]
			frameInd++
			return v
		},
		spellAllowed: func(id spell.ID) bool {
			return id == spell.SPELL_HASTE || id == spell.SPELL_PROTECTION_FROM_FIRE
		},
		random: func(min, max int) int {
			randomInd++
			switch randomInd {
			case 1:
				if min != 0 || max != 1 {
					t.Fatalf("selection bounds = %d..%d", min, max)
				}
				return 1
			case 2:
				if min != 11 || max != 13 {
					t.Fatalf("cooldown bounds = %d..%d", min, max)
				}
				return 12
			default:
				t.Fatalf("unexpected RNG call %d", randomInd)
				return 0
			}
		},
		cast: func(gotUnit *Object, id spell.ID, target *Object) {
			if gotUnit != unit || target != unit || id != spell.SPELL_PROTECTION_FROM_FIRE {
				t.Fatalf("cast = %p/%d/%p", gotUnit, id, target)
			}
			unit.UpdateData = unsafe.Pointer(replacement)
		},
	}
	if !monsterFightBuffSelf540B90(unit, hooks) {
		t.Fatal("self spell was not selected")
	}
	if update.Field365 != 152 || replacement.Field365 != 0 || frameInd != 2 || randomInd != 2 {
		t.Fatalf("cooldown/cache/calls = %d/%d/%d/%d", update.Field365, replacement.Field365, frameInd, randomInd)
	}
	unit.UpdateData = unsafe.Pointer(update)
}

func TestMonsterFightCastOffensive540F20TargetAndCooldown(t *testing.T) {
	unit, update := monsterFightSpellTestUnit540B90(t)
	target := new(Object)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_FIREBALL, monsterFightOffensiveSpellMask540F20)
	update.Field367 = 75
	update.Field366_0 = 7
	update.Field366_2 = 9
	frames := []uint32{100, 110}
	frameInd := 0
	randomInd := 0
	if !monsterFightCastOffensive540F20(unit, target, monsterFightSpellHooks540B90{
		frame: func() uint32 {
			v := frames[frameInd]
			frameInd++
			return v
		},
		spellAllowed: func(id spell.ID) bool { return id == spell.SPELL_FIREBALL },
		random: func(min, max int) int {
			randomInd++
			if randomInd == 1 {
				if min != 0 || max != 0 {
					t.Fatalf("selection bounds = %d..%d", min, max)
				}
				return 0
			}
			if min != 7 || max != 9 {
				t.Fatalf("cooldown bounds = %d..%d", min, max)
			}
			return 8
		},
		cast: func(gotUnit *Object, id spell.ID, gotTarget *Object) {
			if gotUnit != unit || id != spell.SPELL_FIREBALL || gotTarget != target {
				t.Fatalf("cast = %p/%d/%p", gotUnit, id, gotTarget)
			}
		},
	}) {
		t.Fatal("offensive spell was not selected")
	}
	if update.Field367 != 118 || frameInd != 2 || randomInd != 2 {
		t.Fatalf("cooldown/calls = %d/%d/%d", update.Field367, frameInd, randomInd)
	}
}

func TestMonsterFightCastRelated540D90RerollsRejectedSummon(t *testing.T) {
	unit, update := monsterFightSpellTestUnit540B90(t)
	target := new(Object)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_FIREBALL, monsterFightRelatedSpellMask540D90)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_SUMMON_BAT, monsterFightRelatedSpellMask540D90)
	update.Field368_0 = 5
	update.Field368_2 = 7
	frames := []uint32{100, 120}
	frameInd := 0
	randomResults := []int{1, 0, 6}
	randomInd := 0
	activeCalls := 0
	if !monsterFightCastRelated540D90(unit, target, monsterFightSpellHooks540B90{
		frame: func() uint32 {
			v := frames[frameInd]
			frameInd++
			return v
		},
		spellAllowed: func(id spell.ID) bool { return id == spell.SPELL_FIREBALL || id == spell.SPELL_SUMMON_BAT },
		random: func(min, max int) int {
			v := randomResults[randomInd]
			randomInd++
			if randomInd < 3 && (min != 0 || max != 1) {
				t.Fatalf("selection bounds = %d..%d", min, max)
			}
			if randomInd == 3 && (min != 5 || max != 7) {
				t.Fatalf("cooldown bounds = %d..%d", min, max)
			}
			return v
		},
		cast: func(gotUnit *Object, id spell.ID, gotTarget *Object) {
			if gotUnit != unit || id != spell.SPELL_FIREBALL || gotTarget != target {
				t.Fatalf("cast = %p/%d/%p", gotUnit, id, gotTarget)
			}
		},
		activeSummon: func(got *Object) bool {
			if got != unit {
				t.Fatalf("active summon unit = %p", got)
			}
			activeCalls++
			return true
		},
		canSummon: func(*Object, int) bool {
			t.Fatal("limit checked after active summon")
			return false
		},
	}) {
		t.Fatal("related spell was not selected")
	}
	if activeCalls != 1 || randomInd != 3 || frameInd != 2 || update.Field369 != 126 {
		t.Fatalf("active/RNG/frame/cooldown = %d/%d/%d/%d", activeCalls, randomInd, frameInd, update.Field369)
	}
}

func TestMonsterFightCastRelated540D90AllSummonRejections(t *testing.T) {
	for _, tc := range []struct {
		name       string
		active     bool
		allow      bool
		limitCalls int
	}{
		{name: "active summon", active: true},
		{name: "creature limit", allow: false, limitCalls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, update := monsterFightSpellTestUnit540B90(t)
			monsterFightSetSpellFlag540B90(update, spell.SPELL_SUMMON_BAT, monsterFightRelatedSpellMask540D90)
			randomCalls := 0
			limitCalls := 0
			if monsterFightCastRelated540D90(unit, new(Object), monsterFightSpellHooks540B90{
				frame:        func() uint32 { return 100 },
				spellAllowed: func(id spell.ID) bool { return id == spell.SPELL_SUMMON_BAT },
				random: func(min, max int) int {
					if min != 0 || max != 0 {
						t.Fatalf("selection bounds = %d..%d", min, max)
					}
					randomCalls++
					return 0
				},
				cast:         func(*Object, spell.ID, *Object) { t.Fatal("rejected summon was cast") },
				activeSummon: func(*Object) bool { return tc.active },
				canSummon: func(gotUnit *Object, size int) bool {
					if gotUnit != unit || size != 1 {
						t.Fatalf("limit args = %p/%d", gotUnit, size)
					}
					limitCalls++
					return tc.allow
				},
			}) {
				t.Fatal("rejected all-summon set reported a cast")
			}
			if randomCalls != 1 || limitCalls != tc.limitCalls || update.Field369 != 0 {
				t.Fatalf("RNG/limit/cooldown = %d/%d/%d", randomCalls, limitCalls, update.Field369)
			}
		})
	}
}

func TestMonsterFightCastRelated540D90SummonLimitIndex(t *testing.T) {
	unit, update := monsterFightSpellTestUnit540B90(t)
	target := new(Object)
	monsterFightSetSpellFlag540B90(update, spell.SPELL_SUMMON_URCHIN_SHAMAN, monsterFightRelatedSpellMask540D90)
	update.Field368_0 = 3
	update.Field368_2 = 3
	randomCalls := 0
	castCalls := 0
	frame := uint32(100)
	if !monsterFightCastRelated540D90(unit, target, monsterFightSpellHooks540B90{
		frame:        func() uint32 { frame++; return frame },
		spellAllowed: func(id spell.ID) bool { return id == spell.SPELL_SUMMON_URCHIN_SHAMAN },
		random: func(min, max int) int {
			randomCalls++
			if randomCalls == 1 {
				return 0
			}
			if min != 3 || max != 3 {
				t.Fatalf("cooldown bounds = %d..%d", min, max)
			}
			return 3
		},
		cast: func(gotUnit *Object, id spell.ID, gotTarget *Object) {
			if gotUnit != unit || id != spell.SPELL_SUMMON_URCHIN_SHAMAN || gotTarget != target {
				t.Fatalf("cast = %p/%d/%p", gotUnit, id, gotTarget)
			}
			castCalls++
		},
		activeSummon: func(*Object) bool { return false },
		canSummon: func(gotUnit *Object, size int) bool {
			if gotUnit != unit || size != 40 {
				t.Fatalf("limit args = %p/%d, want unit/40", gotUnit, size)
			}
			return true
		},
	}) {
		t.Fatal("allowed summon was not selected")
	}
	if castCalls != 1 || randomCalls != 2 || update.Field369 != 105 {
		t.Fatalf("cast/RNG/cooldown = %d/%d/%d", castCalls, randomCalls, update.Field369)
	}
}
