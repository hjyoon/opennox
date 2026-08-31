package server

import "strconv"

type ExecAbilityClass struct {
	Abil   Ability           // 0, 0
	Unit   *Object           // 1, 4
	Frame  uint32            // 2, 8
	Active uint32            // 3, 12
	Next   *ExecAbilityClass // 4, 16
	Prev   *ExecAbilityClass // 5, 20
}

type serverAbilities struct {
	s         *Server
	cooldowns [abilityRuntimePlayerSlots4FB990][AbilityMax]int32
	execList  *ExecAbilityClass
}

func (a *serverAbilities) init(s *Server) {
	a.s = s
}

func (a *serverAbilities) PlayerAbilityCooldownAt(index uint8, abil Ability) int32 {
	return a.cooldowns[index][abil]
}

func (a *serverAbilities) SetPlayerAbilityCooldownAt(index uint8, abil Ability, cooldown int32) {
	a.cooldowns[index][abil] = cooldown
}

func (a *serverAbilities) ExecHead() *ExecAbilityClass {
	return a.execList
}

func (a *serverAbilities) SetExecHead(head *ExecAbilityClass) {
	a.execList = head
}

func (a *serverAbilities) GetCooldownForUnit(unit *Object, abil Ability) int {
	return int(a.PlayerAbilityCooldownGet4FBE60(unit, abil))
}

func (a *serverAbilities) SetCooldownForUnit(unit *Object, abil Ability, cd int) {
	a.PlayerAbilityCooldownSet4FBEA0(unit, abil, int32(cd))
}

func (a *serverAbilities) DisableAbilityAaa(u *Object, abil Ability) {
	if u == nil {
		return
	}
	if !abil.Valid() {
		return
	}
	var next *ExecAbilityClass
	for it := a.execList; it != nil; it = next {
		next = it.Next
		if it.Unit == u && it.Abil == abil {
			if next != nil {
				next.Prev = it.Prev
			}
			if prev := it.Prev; prev != nil {
				prev.Next = it.Next
			} else {
				a.execList = it.Next
			}
			*it = ExecAbilityClass{}
		}
	}
}

type Ability int32

func (a Ability) String() string {
	if a >= 0 && int(a) < len(AbilityNames) {
		return AbilityNames[a]
	}
	return "Ability(" + strconv.Itoa(int(a)) + ")"
}

func (a Ability) Valid() bool {
	return a > AbilityInvalid && a < AbilityMax
}

const (
	AbilityInvalid = Ability(iota)
	AbilityBerserk
	AbilityWarcry
	AbilityHarpoon
	AbilityTreadLightly
	AbilityInfravis
	AbilityMax
)

var AbilityNames = []string{
	"ABILITY_INVALID", "ABILITY_BERSERKER_CHARGE", "ABILITY_WARCRY",
	"ABILITY_HARPOON", "ABILITY_TREAD_LIGHTLY", "ABILITY_EYE_OF_THE_WOLF",
}
