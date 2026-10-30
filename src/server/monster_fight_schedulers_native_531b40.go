package server

import "github.com/opennox/opennox/v1/common/unit/ai"

// MonsterFightSchedulerRuntime531B40 supplies the indirect engine services.
// RelatedSpell overrides selection; otherwise CanSummon binds its limit check.
// A missing limit service rejects summons while allowing other related spells.
// A nil CancelAction uses the ordinary registered action implementations.
// Push admission, stack writes, reset and distance always run in production code.
type MonsterFightSchedulerRuntime531B40 struct {
	RelatedSpell func(*Object, *Object) bool
	CancelAction func(ai.ActionType) func(*Object)
	CanSummon    func(*Object, int) bool
}

// MonsterFightScheduler531B40 implements the four public scheduler entries.
// Their original return may be a stack pointer, zero or the integer sentinel 1;
// uintptr keeps the latter out of Go's pointer representation.
func (s *Server) MonsterFightScheduler531B40(entry uint32, unit, target *Object, runtime MonsterFightSchedulerRuntime531B40) uintptr {
	if unit == nil || target == nil || unit.UpdateData == nil {
		return 0
	}
	related := runtime.RelatedSpell
	if related == nil {
		canSummon := runtime.CanSummon
		if canSummon == nil {
			canSummon = func(*Object, int) bool { return false }
		}
		related = func(unit, target *Object) bool {
			return monsterFightCastRelated540D90(unit, target, s.monsterFightSpellHooks540B90(canSummon))
		}
	}
	hooks := monsterActionFightHooks531EC0{
		castRelated: related,
		distance:    ObjectDistance4E6C00,
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			item := unit.monsterPushActionWithCancel50A260(action, "go", 0, runtime.CancelAction)
			item.SetArgs(args...)
			return item
		},
	}
	switch entry {
	case 0x531b40:
		return monsterFightSelectAttack531B40(unit, target, hooks)
	case 0x531c10:
		return monsterFightSchedulePursuit531C10(target, hooks)
	case 0x531c60:
		return monsterFightScheduleMelee531C60(unit, target, hooks)
	case 0x531d50:
		return monsterFightScheduleMissile531D50(unit, target, hooks)
	default:
		return 0
	}
}
