package server

import "math"

type questHealthCache4E3DD0 uint8

const (
	questHealthDamageInit4E3DD0 questHealthCache4E3DD0 = iota
	questHealthHealthInit4E3DD0
	questHealthDamageCoeff4E3DD0
	questHealthHealthCoeff4E3DD0
)

type questHealthScaleHooks4E3DD0[O, D comparable, H, T, U any] struct {
	balanceFloat      func(string) float32
	loadReady         func() uint32
	storeReady        func(uint32)
	loadCache         func(questHealthCache4E3DD0) float32
	storeCache        func(questHealthCache4E3DD0, float32)
	difficulty        func() float64
	storeDamageScale  func(float32)
	storeHealthScale  func(float32)
	healthScale       func() float64
	first             func() O
	next              func(O) O
	loadClass         func(O) uint32
	loadFlags         func(O) uint32
	loadHealth        func(O) H
	healthPointerWord func(H) uint16
	loadCurrent       func(H) uint16
	loadMaximum       func(H) uint16
	loadTypeID        func(O) uint16
	lookupType        func(uint16) T
	loadTypeHealth    func(T) H
	loadUpdate        func(O) U
	loadDefinition    func(U) D
	loadQuestHealth   func(D) uint16
	loadStatusByte    func(U) uint8
	setHP             func(O, uint16) int32
	storeMaximum      func(H, uint16)
	storeHistory      func(U, int, uint16)
	historyEndWord    func(U) uint16
}

// questHealthWord4E3DD0 models 00419AB0: spill to binary32, FABS, FISTP
// signed dword (nearest even, INT32_MIN on invalid), then retain only AX.
// The legacy C helper truncates instead, so this body must not call it.
func questHealthWord4E3DD0(value float32) uint16 {
	value = math.Float32frombits(math.Float32bits(value) & 0x7fffffff)
	return uint16(questInventoryRoundFloat32ToInt32_4F2C30(value))
}

// questHealthScale4E3DD0 preserves the whole GAME.EXE 004E3DD0 body. Pointer
// words are used only for the original short return value, never as addresses.
func questHealthScale4E3DD0[O, D comparable, H, T, U any](h questHealthScaleHooks4E3DD0[O, D, H, T, U]) int16 {
	cap := uint16(questInventoryRoundFloat32ToInt32_4F2C30(h.balanceFloat("GeneratorMaxHealth")))
	if h.loadReady() == 0 {
		h.storeCache(questHealthDamageInit4E3DD0, h.balanceFloat("PlayerDamageDiffInit"))
		h.storeCache(questHealthHealthInit4E3DD0, h.balanceFloat("SystemHealthDiffInit"))
		h.storeCache(questHealthDamageCoeff4E3DD0, h.balanceFloat("PlayerDamageDiffCoeff"))
		h.storeCache(questHealthHealthCoeff4E3DD0, h.balanceFloat("SystemHealthDiffCoeff"))
		h.storeReady(1)
	}
	// Separate 53-bit operations prohibit FMA and premature binary32 rounding.
	delta := logicRandomFloatSub64_416030(h.difficulty(), 1)
	product := logicRandomFloatMul64_416030(delta, float64(h.loadCache(questHealthDamageCoeff4E3DD0)))
	h.storeDamageScale(float32(logicRandomFloatAdd64_416030(product, float64(h.loadCache(questHealthDamageInit4E3DD0)))))
	delta = logicRandomFloatSub64_416030(h.difficulty(), 1)
	product = logicRandomFloatMul64_416030(delta, float64(h.loadCache(questHealthHealthCoeff4E3DD0)))
	h.storeHealthScale(float32(logicRandomFloatAdd64_416030(product, float64(h.loadCache(questHealthHealthInit4E3DD0)))))

	var zero O
	var zeroDefinition D
	var result uint16
	for obj := h.first(); obj != zero; {
		// A setter may change the world list; the original caches its successor.
		next := h.next(obj)
		class := h.loadClass(obj)
		result = uint16(class)
		if class&0x20000 != 0 && h.loadFlags(obj)&0x8000 == 0 {
			health := h.loadHealth(obj)
			result = h.healthPointerWord(health)
			if maximum := h.loadMaximum(health); maximum != 0 {
				result = h.loadCurrent(health)
				if result != 0 && result == maximum {
					typ := h.lookupType(h.loadTypeID(obj))
					factor := h.healthScale()
					maximum := questHealthWord4E3DD0(float32(logicRandomFloatMul64_416030(factor, float64(h.loadMaximum(h.loadTypeHealth(typ))))))
					factor = h.healthScale()
					current := questHealthWord4E3DD0(float32(logicRandomFloatMul64_416030(factor, float64(h.loadCurrent(h.loadTypeHealth(typ))))))
					if current == 0 {
						current = 1
					}
					if maximum == 0 {
						maximum = 1
					}
					if current > cap {
						current = cap
					}
					if maximum > cap {
						maximum = cap
					}
					result = uint16(h.setHP(obj, current))
					h.storeMaximum(h.loadHealth(obj), maximum)
				}
			}
		} else if class&2 != 0 {
			flags := h.loadFlags(obj)
			result = uint16(flags)
			if flags&0x8000 == 0 {
				health := h.loadHealth(obj)
				result = h.healthPointerWord(health)
				if maximum := h.loadMaximum(health); maximum != 0 {
					result = h.loadCurrent(health)
					if result != 0 && result == maximum {
						typ := h.lookupType(h.loadTypeID(obj))
						update := h.loadUpdate(obj)
						definition := h.loadDefinition(update)
						if definition != zeroDefinition {
							result = h.loadQuestHealth(definition)
						} else {
							result = h.loadMaximum(h.loadTypeHealth(typ))
						}
						if h.loadStatusByte(update)&0x80 == 0 {
							base := float64(float32(result))
							maximum := questHealthWord4E3DD0(float32(logicRandomFloatMul64_416030(h.healthScale(), base)))
							current := questHealthWord4E3DD0(float32(logicRandomFloatMul64_416030(h.healthScale(), base)))
							if current == 0 {
								current = 1
							}
							if maximum == 0 {
								maximum = 1
							}
							h.setHP(obj, current)
							h.storeMaximum(h.loadHealth(obj), maximum)
							for i := 0; i < 32; i++ {
								h.storeHistory(update, i, h.loadCurrent(h.loadHealth(obj)))
							}
							result = h.historyEndWord(update)
						}
					}
				}
			}
		}
		obj = next
	}
	return int16(result)
}
