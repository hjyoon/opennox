package server

import "math"

type questHecubahSpawnHooks51A5A0[O, D comparable, U, H, T, P any] struct {
	questNecroSpawnHooks51A7A0[O, D, U, H, T, P]
	balanceFloat func(string) float32
}

// questHecubahSpawn51A5A0 is GAME.EXE 0051A5A0. Hecubah shares the exact HP
// arithmetic with 0051A7A0, but has a different ordered AI prefix, a required
// (discarded) balance lookup, and four separately evaluated stage+2 rewards.
func questHecubahSpawn51A5A0[O, D comparable, U, H, T, P any](h questHecubahSpawnHooks51A5A0[O, D, U, H, T, P]) {
	unit := h.newObject("Hecubah")
	scale := float32(h.healthScale())
	var zero O
	if unit == zero {
		return
	}
	update := h.loadUpdate(unit)
	definition := h.loadDefinition(update)
	var zeroDefinition D
	var base int32
	if definition != zeroDefinition {
		base = h.loadQuestHealth(definition)
	} else {
		base = int32(h.loadMaximum(h.loadTypeHealth(h.lookupType(h.loadType(unit)))))
	}
	if scale < 1 || math.IsNaN(float64(scale)) {
		scale = 1
	}
	current, maximum := questMinionHealth51A7A0(base, scale)
	h.setHP(unit, current)
	h.storeMaximum(h.loadHealth(unit), maximum)
	if h.loadCurrent(h.loadHealth(unit)) == 0 {
		h.setHP(unit, 1)
	}
	health := h.loadHealth(unit)
	if h.loadMaximum(health) == 0 {
		h.storeMaximum(health, 1)
	}
	h.storeAI(update, 411, 0x10000000)
	h.storeAI(update, 423, 0x10000000)
	h.storeAI(update, 340, 4)
	h.storeAI(update, 326, 0x3f547ae1)
	h.storeAI(update, 510, 3)
	h.storeAI(update, 410, 0x08000000)
	h.storeAI(update, 444, 0x20000000)
	h.storeAI(update, 388, 0x40000000)
	h.storeAI(update, 415, 0x40000000)
	h.balanceFloat("HecubahQuestSkill")
	h.storeAI(update, 330, 0x3f59999a)
	h.createAt(unit, h.loadPosition())
	marker := h.newObject("RewardMarker")
	if marker == zero {
		return
	}
	for i := 0; i < 4; i++ {
		if reward := h.activateReward(marker, h.stage()+2); reward != zero {
			h.inventoryPut(unit, reward, 0)
		}
	}
	h.freeObject(marker)
}
