package server

import "github.com/opennox/libs/types"

// QuestThemeRuntime51A1F0 supplies the original scalar Quest globals and the
// outer-server callbacks. No native Object or update pointer occupies a PE32
// four-byte slot.
type QuestThemeRuntime51A1F0 struct {
	QuestStage                       func() uint32
	HecubahType, NecroType           func() uint32
	StoreHecubahType, StoreNecroType func(uint32)
	GeneratorType                    func(*Object) int32
	DelayedDelete                    func(*Object)
	SetMinions                       func(int32)
	SpawnHecubah, SpawnNecro         func(*types.Pointf)
}

func (s *Server) questThemeNativeHooks51A1F0(r QuestThemeRuntime51A1F0) questThemeHooks51A1F0[*Object, *MonsterGenUpdateData] {
	return questThemeHooks51A1F0[*Object, *MonsterGenUpdateData]{
		questStage: r.QuestStage,
		balanceFloat: func(key string) float64 {
			// Original 00419D40 returns an FLD-dword value in ST0.
			return float64(float32(s.Balance.Float(key)))
		},
		floatToInt:       questInventoryRoundFloat32ToInt32_4F2C30,
		hecubahType:      r.HecubahType,
		necroType:        r.NecroType,
		storeHecubahType: r.StoreHecubahType,
		storeNecroType:   r.StoreNecroType,
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		first: s.Objs.First,
		next:  (*Object).Next,
		loadClass: func(unit *Object) uint32 {
			return uint32(unit.ObjClass)
		},
		loadSubclassByte: func(unit *Object) uint8 {
			return uint8(unit.ObjSubClass)
		},
		loadType: func(unit *Object) uint16 {
			return unit.TypeInd
		},
		storeType: func(unit *Object, id uint16) {
			unit.TypeInd = id
		},
		loadUpdate: func(unit *Object) *MonsterGenUpdateData {
			// 0051A2A3 has no UpdateData nil guard.
			return (*MonsterGenUpdateData)(unit.UpdateData)
		},
		loadCreature: func(update *MonsterGenUpdateData, group int32) *Object {
			// Three groups each occupy four native template-pointer slots.
			// All original callers select 0..2; Go faults on an invalid index
			// instead of reading adjacent native memory.
			return update.Field0[4*group]
		},
		loadSelector: func(update *MonsterGenUpdateData, group int32) uint8 {
			return update.QuestSpawnRate[group]
		},
		truncQwordLow: x87TruncSignedQwordLow566DCC,
		loadMaximum: func(update *MonsterGenUpdateData) uint8 {
			return update.MaxActive
		},
		storeMaximum: func(update *MonsterGenUpdateData, maximum uint8) {
			update.MaxActive = maximum
		},
		generatorType: r.GeneratorType,
		delete:        r.DelayedDelete,
		random: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		setMinions: r.SetMinions,
		spawnHecubah: func(unit *Object) {
			r.SpawnHecubah(&unit.PosVec)
		},
		spawnNecro: func(unit *Object) {
			r.SpawnNecro(&unit.PosVec)
		},
	}
}

// QuestTheme51A1F0 applies the chosen Quest generator group, retains one exit,
// optionally spawns minions, and removes marker objects with native pointers.
func (s *Server) QuestTheme51A1F0(group int32, r QuestThemeRuntime51A1F0) {
	questTheme51A1F0(group, s.questThemeNativeHooks51A1F0(r))
}
