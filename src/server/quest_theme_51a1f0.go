package server

const (
	questHardcoreStageKey51A1F0 = "QuestHardcoreStage"
	questMinionsAlwaysKey51A1F0 = "MinionsAlwaysStage"
	questHecubahMarker51A1F0    = "HecubahMarker"
	questNecroMarker51A1F0      = "NecromancerMarker"
)

type questThemeHooks51A1F0[O comparable, U any] struct {
	questStage                       func() uint32
	balanceFloat                     func(string) float64
	floatToInt                       func(float32) int32
	hecubahType, necroType           func() uint32
	storeHecubahType, storeNecroType func(uint32)
	lookupType                       func(string) uint32
	first                            func() O
	next                             func(O) O
	loadClass                        func(O) uint32
	loadSubclassByte                 func(O) uint8
	loadType                         func(O) uint16
	storeType                        func(O, uint16)
	loadUpdate                       func(O) U
	loadCreature                     func(U, int32) O
	loadSelector                     func(U, int32) uint8
	truncQwordLow                    func(float64) int32
	loadMaximum                      func(U) uint8
	storeMaximum                     func(U, uint8)
	generatorType                    func(O) int32
	delete                           func(O)
	random                           func(int32, int32) int32
	setMinions                       func(int32)
	spawnHecubah, spawnNecro         func(O)
}

// questTheme51A1F0 follows GAME.EXE 0051A1F0's four passes. Deletion passes
// cache the successor before callbacks; the minion-spawn pass reloads it
// afterward. Generator update pointers are cached, but the selected creature
// and spawn-rate byte are reloaded after the balance/stage callbacks.
func questTheme51A1F0[O comparable, U any](group int32, h questThemeHooks51A1F0[O, U]) {
	var zero O
	var exitCount, markerCount uint32
	stage := h.questStage()
	hardcore := uint32(h.floatToInt(float32(h.balanceFloat(questHardcoreStageKey51A1F0))))
	if h.hecubahType() == 0 {
		h.storeHecubahType(h.lookupType(questHecubahMarker51A1F0))
		h.storeNecroType(h.lookupType(questNecroMarker51A1F0))
	}
	for unit := h.first(); unit != zero; {
		next := h.next(unit)
		class := h.loadClass(unit)
		if class&0x20 != 0 && h.loadSubclassByte(unit)&1 != 0 {
			exitCount++
		} else {
			marker := h.hecubahType()
			if uint32(h.loadType(unit)) == marker {
				markerCount++
			}
		}
		if class&0x20000 != 0 {
			update := h.loadUpdate(unit)
			if h.loadCreature(update, group) != zero {
				selector := h.loadSelector(update, group)
				if selector <= 3 {
					value := h.balanceFloat(monsterGeneratorMaxActiveKeys4F0590[selector])
					h.storeMaximum(update, uint8(h.truncQwordLow(value)))
				}
				// The hardcore comparison is unsigned; the later minion gates
				// use the signed entry-time stage instead of this live value.
				if h.questStage() >= hardcore && h.loadSelector(update, group) != 3 {
					h.storeMaximum(update, h.loadMaximum(update)*2)
				}
				if id := h.generatorType(h.loadCreature(update, group)); id != 0 {
					h.storeType(unit, uint16(id))
				}
			} else {
				h.delete(unit)
			}
		}
		unit = next
	}
	if exitCount > 1 {
		selected := h.random(0, int32(exitCount-1))
		var index int32
		for unit := h.first(); unit != zero; {
			next := h.next(unit)
			if uint8(h.loadClass(unit))&0x20 != 0 && h.loadSubclassByte(unit)&1 != 0 {
				if index != selected {
					h.delete(unit)
				}
				index++
			}
			unit = next
		}
	}
	h.setMinions(0)
	if int32(stage) >= 5 {
		always := h.floatToInt(float32(h.balanceFloat(questMinionsAlwaysKey51A1F0)))
		if stage == 5 || int32(stage) >= always || stage&1 != 0 && h.random(1, 100) >= 50 {
			h.setMinions(1)
			if markerCount != 0 {
				selected := h.random(1, int32(markerCount))
				var index int32
				for unit := h.first(); unit != zero; unit = h.next(unit) {
					marker := h.hecubahType()
					if uint32(h.loadType(unit)) == marker {
						index++
						if index == selected {
							h.spawnHecubah(unit)
						}
					}
					marker = h.necroType()
					if uint32(h.loadType(unit)) == marker && h.random(1, 100) >= 50 {
						h.spawnNecro(unit)
					}
				}
			}
		}
	}
	for unit := h.first(); unit != zero; {
		next := h.next(unit)
		if uint32(h.loadType(unit)) == h.hecubahType() {
			h.delete(unit)
		}
		marker := h.necroType()
		if uint32(h.loadType(unit)) == marker {
			h.delete(unit)
		}
		unit = next
	}
}
