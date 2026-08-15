package server

// RespawnSchedulerHooks4EC720 separates the pointer-width-independent
// GAME.EXE contract from object, record, clock, audio, and effect storage.
// The numerous narrow hooks are intentional: callbacks in the original code
// can mutate live object and record fields between adjacent reads.
type RespawnSchedulerHooks4EC720[O comparable, R comparable, U any, H comparable] struct {
	LoadCrown             func() uint32
	LookupTypeByName      func(string) uint32
	StoreCrown            func(uint32)
	GameFlagsCheck        func(uint32) uint32
	LoadHead              func() R
	StoreAllow            func(uint32)
	LoadPending           func(R) uint32
	StorePending          func(R, uint32)
	LoadRespawnAt         func(R) uint32
	StoreRespawnAt        func(R, uint32)
	LoadObject            func(R) O
	StoreObject           func(R, O)
	LoadNext              func(R) R
	LoadRecordTypeInd     func(R) uint32
	LoadRecordX           func(R) float32
	LoadRecordY           func(R) float32
	LoadDirection         func(R) uint16
	LoadCharge1           func(R) uint8
	LoadCharge0           func(R) uint8
	LoadObjectTypeInd     func(O) uint16
	LoadClass             func(O) uint32
	LoadFlags             func(O) uint32
	LoadInvHolder         func(O) O
	LoadField32           func(O) uint32
	LoadObjectX           func(O) float32
	LoadObjectY           func(O) float32
	LoadUseData           func(O) U
	StoreUseByte          func(U, uint32, uint8)
	LoadHealthData        func(O) H
	LoadHealthMax         func(H) uint16
	UnitDefAllowed        func(uint32) bool
	NewObjectByTypeInd    func(uint32) O
	SpecialModeCheck      func(uint32) uint32
	LoadFPS               func() uint32
	LoadFrame             func() uint32
	PointFXAtObject       func(uint32, O)
	PointFXAtRecord       func(uint32, R)
	AudioAtObjectPosition func(uint32, O)
	AudioAtRecordPosition func(uint32, R)
	AudioOnObject         func(uint32, O)
	MoveToRecord          func(O, R)
	Recharge              func(O, uint32)
	WeaponEquipFlags      func(O) uint32
	SetHP                 func(O, uint16)
	CreateAt              func(O, O, float32, float32)
	StoreDirection1       func(O, uint16)
	StoreDirection2       func(O, uint16)
	CopyModifierAttrs     func(O, R)
	DelayedDelete         func(O)
}

func respawnSchedule4EC720[O comparable, R comparable, U any, H comparable](
	rec R, hooks RespawnSchedulerHooks4EC720[O, R, U, H],
) {
	hooks.StorePending(rec, 1)
	fps := hooks.LoadFPS()
	frame := hooks.LoadFrame()
	hooks.StoreRespawnAt(rec, frame+30*fps)
}

func respawnRestoreMovedObject4EC720[O comparable, R comparable, U any, H comparable](
	rec R, hooks RespawnSchedulerHooks4EC720[O, R, U, H],
) {
	fps := hooks.LoadFPS()
	obj := hooks.LoadObject(rec)
	field32 := hooks.LoadField32(obj)
	frame := hooks.LoadFrame()
	if frame <= field32+5*fps {
		return
	}

	recX := hooks.LoadRecordX(rec)
	objX := hooks.LoadObjectX(obj)
	recY := hooks.LoadRecordY(rec)
	objY := hooks.LoadObjectY(obj)
	dx := float64(recX) - float64(objX)
	dy := float64(recY) - float64(objY)
	dx2 := dx * dx
	dy2 := dy * dy
	if !(dx2+dy2 > 2500.0) {
		return
	}

	// The first position pointer is based on the object loaded before the
	// distance test. Every later object access is a live record reload.
	hooks.PointFXAtObject(129, obj)
	obj = hooks.LoadObject(rec)
	hooks.AudioAtObjectPosition(283, obj)
	obj = hooks.LoadObject(rec)
	hooks.MoveToRecord(obj, rec)

	obj = hooks.LoadObject(rec)
	class := hooks.LoadClass(obj)
	if class&0x00001000 != 0 {
		hooks.Recharge(obj, 100)
	} else if class&0x01000000 != 0 {
		if uint8(hooks.WeaponEquipFlags(obj))&0x82 != 0 {
			obj = hooks.LoadObject(rec)
			charge1 := hooks.LoadCharge1(rec)
			useData := hooks.LoadUseData(obj)
			hooks.StoreUseByte(useData, 1, charge1)
			charge0 := hooks.LoadCharge0(rec)
			hooks.StoreUseByte(useData, 0, charge0)
		}
	}

	obj = hooks.LoadObject(rec)
	health := hooks.LoadHealthData(obj)
	var zeroH H
	if health != zeroH {
		maxHealth := hooks.LoadHealthMax(health)
		hooks.SetHP(obj, maxHealth)
	}
	hooks.PointFXAtRecord(129, rec)
	hooks.AudioAtRecordPosition(283, rec)
}

func respawnClassify4EC720[O comparable, R comparable, U any, H comparable](
	rec R, hooks RespawnSchedulerHooks4EC720[O, R, U, H],
) {
	var zeroO O
	obj := hooks.LoadObject(rec)
	if obj == zeroO {
		respawnSchedule4EC720(rec, hooks)
		return
	}

	class := hooks.LoadClass(obj)
	if uint8(class)&0x02 != 0 {
		flags := hooks.LoadFlags(obj)
		if uint8(flags)&0x20 != 0 {
			hooks.StoreObject(rec, zeroO)
			respawnSchedule4EC720(rec, hooks)
		} else if flags&0x00008000 != 0 {
			respawnSchedule4EC720(rec, hooks)
		}
		return
	}

	flags := hooks.LoadFlags(obj)
	if uint8(flags)&0x20 != 0 {
		typeInd := hooks.LoadObjectTypeInd(obj)
		allowed := hooks.UnitDefAllowed(uint32(typeInd))
		hooks.StoreObject(rec, zeroO)
		if allowed {
			respawnSchedule4EC720(rec, hooks)
		}
		return
	}

	if class&0x03001000 == 0 {
		crown := hooks.LoadCrown()
		typeInd := hooks.LoadObjectTypeInd(obj)
		if uint32(typeInd) != crown {
			if hooks.LoadInvHolder(obj) != zeroO {
				respawnSchedule4EC720(rec, hooks)
			}
			return
		}
	}

	if hooks.LoadInvHolder(obj) == zeroO {
		typeInd := hooks.LoadObjectTypeInd(obj)
		if hooks.UnitDefAllowed(uint32(typeInd)) {
			respawnRestoreMovedObject4EC720(rec, hooks)
			return
		}
	}

	// This is deliberately a second inventory-holder test on a freshly
	// loaded object. The callbacks above may have replaced rec.object.
	obj = hooks.LoadObject(rec)
	if hooks.LoadInvHolder(obj) == zeroO {
		return
	}
	typeInd := hooks.LoadObjectTypeInd(obj)
	if !hooks.UnitDefAllowed(uint32(typeInd)) {
		return
	}
	obj = hooks.LoadObject(rec)
	crown := hooks.LoadCrown()
	typeInd = hooks.LoadObjectTypeInd(obj)
	if uint32(typeInd) == crown {
		return
	}
	if hooks.SpecialModeCheck(2) != 0 {
		respawnSchedule4EC720(rec, hooks)
	}
}

// RespawnScheduler4EC720 preserves GAME.EXE 004EC720..004ECA57. Besides the
// branch results, it retains the original live reloads, callback order,
// uint32 frame arithmetic, and strict ordered x87-style distance comparison.
func RespawnScheduler4EC720[O comparable, R comparable, U any, H comparable](
	hooks RespawnSchedulerHooks4EC720[O, R, U, H],
) {
	if hooks.LoadCrown() == 0 {
		crown := hooks.LookupTypeByName("Crown")
		hooks.StoreCrown(crown)
	}
	if hooks.GameFlagsCheck(0x1200) != 0 {
		return
	}

	rec := hooks.LoadHead()
	hooks.StoreAllow(0)
	var zeroR R
	var zeroO O
	for rec != zeroR {
		if hooks.LoadPending(rec) == 0 {
			respawnClassify4EC720(rec, hooks)
			if hooks.LoadPending(rec) == 0 {
				rec = hooks.LoadNext(rec)
				continue
			}
		}

		frame := hooks.LoadFrame()
		respawnAt := hooks.LoadRespawnAt(rec)
		if frame >= respawnAt {
			typeInd := hooks.LoadRecordTypeInd(rec)
			if hooks.UnitDefAllowed(typeInd) {
				typeInd = hooks.LoadRecordTypeInd(rec)
				newObj := hooks.NewObjectByTypeInd(typeInd)
				if newObj != zeroO {
					y := hooks.LoadRecordY(rec)
					x := hooks.LoadRecordX(rec)
					hooks.CreateAt(newObj, zeroO, x, y)
					hooks.PointFXAtRecord(129, rec)
					direction := hooks.LoadDirection(rec)
					hooks.StoreDirection1(newObj, direction)
					hooks.StoreDirection2(newObj, direction)
					if hooks.LoadClass(newObj)&0x13001000 != 0 {
						hooks.CopyModifierAttrs(newObj, rec)
					}
					if hooks.LoadClass(newObj)&0x01000000 != 0 {
						if uint8(hooks.WeaponEquipFlags(newObj))&0x82 != 0 {
							useData := hooks.LoadUseData(newObj)
							charge1 := hooks.LoadCharge1(rec)
							hooks.StoreUseByte(useData, 1, charge1)
							charge0 := hooks.LoadCharge0(rec)
							hooks.StoreUseByte(useData, 0, charge0)
						}
					}
					hooks.AudioOnObject(283, newObj)
				}

				oldObj := hooks.LoadObject(rec)
				if oldObj != zeroO && uint8(hooks.LoadClass(oldObj))&0x02 != 0 {
					hooks.DelayedDelete(oldObj)
				}
				hooks.StorePending(rec, 0)
				hooks.StoreObject(rec, newObj)
			}
		}
		rec = hooks.LoadNext(rec)
	}
}
