package server

import "unsafe"

// The original signed index addresses a word at 24*(index+23). Preserve
// every read inside its 2200-byte monster record when native pointer fields
// and AI stack items expand. Pointer-valued slots must compare the full word.
// Indices outside that original record are invalid memory accesses; indexing
// this table faults instead of manufacturing a valid player state.
var playerBotActionWords4FABC0 = [92]struct {
	offset  uintptr
	pointer bool
}{
	0:  {unsafe.Offsetof(MonsterUpdateData{}.Field0) + 0, false},
	1:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 12, false},
	2:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 36, false},
	3:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 60, false},
	4:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 84, false},
	5:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 108, false},
	6:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 132, false},
	7:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 156, false},
	8:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 180, false},
	9:  {unsafe.Offsetof(MonsterUpdateData{}.Path) + 204, false},
	10: {unsafe.Offsetof(MonsterUpdateData{}.Path) + 228, false},
	11: {unsafe.Offsetof(MonsterUpdateData{}.Path) + 252, false},
	12: {unsafe.Offsetof(MonsterUpdateData{}.Field72) + 0, false},
	13: {unsafe.Offsetof(MonsterUpdateData{}.Waypoints) + 3*unsafe.Sizeof(uintptr(0)), true},
	14: {unsafe.Offsetof(MonsterUpdateData{}.Waypoints) + 9*unsafe.Sizeof(uintptr(0)), true},
	15: {unsafe.Offsetof(MonsterUpdateData{}.Waypoints) + 15*unsafe.Sizeof(uintptr(0)), true},
	16: {unsafe.Offsetof(MonsterUpdateData{}.Pos95) + 4, false},
	17: {unsafe.Offsetof(MonsterUpdateData{}.Field102) + 0, false},
	18: {unsafe.Offsetof(MonsterUpdateData{}.HealthGraph103) + 20, false},
	19: {unsafe.Offsetof(MonsterUpdateData{}.HealthGraph103) + 44, false},
	20: {unsafe.Offsetof(MonsterUpdateData{}.Field120_0) + 0, false},
	21: {unsafe.Offsetof(MonsterUpdateData{}.Field126) + 0, false},
	47: {unsafe.Offsetof(MonsterUpdateData{}.Stamina) + 0, false},
	48: {unsafe.Offsetof(MonsterUpdateData{}.SeenEnemies) + 5*unsafe.Sizeof(uintptr(0)), true},
	49: {unsafe.Offsetof(MonsterUpdateData{}.SeenEnemies) + 11*unsafe.Sizeof(uintptr(0)), true},
	50: {unsafe.Offsetof(MonsterUpdateData{}.Field300) + 0, false},
	51: {unsafe.Offsetof(MonsterUpdateData{}.ScriptLookingForEnemy) + 0, false},
	52: {unsafe.Offsetof(MonsterUpdateData{}.ScriptIsHit) + 0, false},
	53: {unsafe.Offsetof(MonsterUpdateData{}.ScriptCollision) + 0, false},
	54: {unsafe.Offsetof(MonsterUpdateData{}.ScriptLostEnemy) + 0, false},
	55: {unsafe.Offsetof(MonsterUpdateData{}.Field330) + 0, false},
	56: {unsafe.Offsetof(MonsterUpdateData{}.ResumeLevel) + 0, false},
	57: {unsafe.Offsetof(MonsterUpdateData{}.Field342) + 0, false},
	58: {unsafe.Offsetof(MonsterUpdateData{}.Field348) + 0, false},
	59: {unsafe.Offsetof(MonsterUpdateData{}.Field354) + 0, false},
	60: {unsafe.Offsetof(MonsterUpdateData{}.StatusFlags) + 0, false},
	61: {unsafe.Offsetof(MonsterUpdateData{}.Field366_0) + 0, false},
	62: {unsafe.Offsetof(MonsterUpdateData{}.Field372) + 0, false},
	63: {unsafe.Offsetof(MonsterUpdateData{}.Field378) + 0, false},
	64: {unsafe.Offsetof(MonsterUpdateData{}.Field384) + 0, false},
	65: {unsafe.Offsetof(MonsterUpdateData{}.Field390) + 0, false},
	66: {unsafe.Offsetof(MonsterUpdateData{}.Field396) + 0, false},
	67: {unsafe.Offsetof(MonsterUpdateData{}.Field402) + 0, false},
	68: {unsafe.Offsetof(MonsterUpdateData{}.Field408) + 0, false},
	69: {unsafe.Offsetof(MonsterUpdateData{}.Field414) + 0, false},
	70: {unsafe.Offsetof(MonsterUpdateData{}.Field420) + 0, false},
	71: {unsafe.Offsetof(MonsterUpdateData{}.Field426) + 0, false},
	72: {unsafe.Offsetof(MonsterUpdateData{}.Field432) + 0, false},
	73: {unsafe.Offsetof(MonsterUpdateData{}.Field438) + 0, false},
	74: {unsafe.Offsetof(MonsterUpdateData{}.Field444) + 0, false},
	75: {unsafe.Offsetof(MonsterUpdateData{}.Field450) + 0, false},
	76: {unsafe.Offsetof(MonsterUpdateData{}.Field456) + 0, false},
	77: {unsafe.Offsetof(MonsterUpdateData{}.Field462) + 0, false},
	78: {unsafe.Offsetof(MonsterUpdateData{}.Field468) + 0, false},
	79: {unsafe.Offsetof(MonsterUpdateData{}.Field474) + 0, false},
	80: {unsafe.Offsetof(MonsterUpdateData{}.Field480) + 0, false},
	81: {unsafe.Offsetof(MonsterUpdateData{}.Field486) + 0, false},
	82: {unsafe.Offsetof(MonsterUpdateData{}.Field492) + 0, false},
	83: {unsafe.Offsetof(MonsterUpdateData{}.Field498) + 0, false},
	84: {unsafe.Offsetof(MonsterUpdateData{}.Field504) + 0, false},
	85: {unsafe.Offsetof(MonsterUpdateData{}.Field510) + 0, false},
	86: {unsafe.Offsetof(MonsterUpdateData{}.Field516) + 0, false},
	87: {unsafe.Offsetof(MonsterUpdateData{}.Color) + 12, false},
	88: {unsafe.Offsetof(MonsterUpdateData{}.DialogPortraitBuf) + 6, false},
	89: {unsafe.Offsetof(MonsterUpdateData{}.Field534) + 0, false},
	90: {unsafe.Offsetof(MonsterUpdateData{}.Field535) + 20, false},
	91: {unsafe.Offsetof(MonsterUpdateData{}.Field546) + 0, false},
}

// PlayerBotActionToState4FABC0 implements the raw public entry. Selected nil
// records fault before any stores, just as they do in the original function.
func (unit *Object) PlayerBotActionToState4FABC0() byte {
	monster := (*PlayerUpdateData)(unit.UpdateData).Field73
	index := monster.AIStackInd
	if index == -1 {
		return 13
	}
	var action uintptr
	if index >= 0 && index < 24 {
		action = uintptr(monster.AIStack[index].Action)
	} else {
		slot := playerBotActionWords4FABC0[int(index)+23]
		address := unsafe.Add(unsafe.Pointer(monster), slot.offset)
		if slot.pointer {
			action = *(*uintptr)(address)
		} else {
			action = uintptr(*(*uint32)(address))
		}
	}
	switch action {
	case 7, 8, 10, 13, 29:
		if monster.StatusFlags&0x4000 != 0 {
			return 5
		}
		return 0
	case 9:
		return 0
	case 16, 17:
		return 1
	case 18, 19, 20:
		return 2
	case 21, 23:
		return 16
	case 22:
		return 17
	case 24:
		return 5
	case 30:
		return 3
	case 31:
		return 4
	default:
		return 13
	}
}
