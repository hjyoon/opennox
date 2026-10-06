package opennox

import (
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Independent arbitrary-precision reference: each original subtraction,
// multiplication and addition rounds toward zero at 53 bits. It never
// calls the production TwoSum/FMA/Nextafter operations.
func aiNoNewEnemyReferenceSquare546D34(unit, target types.Pointf) float64 {
	chop := func(value float32) *big.Float {
		return new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(float64(value))
	}
	dx := chop(0).Sub(chop(target.X), chop(unit.X))
	dy := chop(0).Sub(chop(target.Y), chop(unit.Y))
	ySquare := chop(0).Mul(dy, dy)
	xSquare := chop(0).Mul(dx, dx)
	sum := chop(0).Add(ySquare, xSquare)
	value, accuracy := sum.Float64()
	if accuracy != big.Exact {
		panic("53-bit finite centre-distance reference outside binary64 exponent range")
	}
	return value
}

func TestAINoNewEnemyDependency546D1BIndependentChopModel(t *testing.T) {
	_, unit, old, _ := aiAliveDependencyNative546A70(t)
	enemy, freeEnemy := alloc.New(server.Object{})
	t.Cleanup(freeEnemy)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(enemy)) <= math.MaxUint32 {
		t.Fatalf("native model enemy below 4 GiB: %p", enemy)
	}
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = enemy
	slot := server.AIStackItem{Args: [4]uintptr{uintptr(unsafe.Pointer(old)), 0x12345678, uintptr(unsafe.Pointer(unit)), 0xff800001}, Field5: 0xfedcba98}
	for _, domain := range []string{"finite-binary32", "map-coordinates", "subnormal-coordinates", "adjacent-targets"} {
		t.Run(domain, func(t *testing.T) {
			word := uint32(0x546d1b01)
			next := func() float32 {
				word = 1664525*word + 1013904223
				bits := word
				switch domain {
				case "map-coordinates":
					bits = word&0x807fffff | (uint32(130)+(word>>23)%10)<<23
				case "subnormal-coordinates":
					bits = word & 0x807fffff
				default:
					if bits&0x7f800000 == 0x7f800000 {
						bits ^= 0x00800000
					}
				}
				return math.Float32frombits(bits)
			}
			for i := 0; i < 8192; i++ {
				unit.PosVec = types.Pointf{X: next(), Y: next()}
				old.PosVec = types.Pointf{X: next(), Y: next()}
				enemy.PosVec = types.Pointf{X: next(), Y: next()}
				if domain == "adjacent-targets" {
					enemy.PosVec = old.PosVec
					// Adjacent finite raw words probe comparisons where early
					// narrowing can make unequal centre distances appear equal.
					bits := math.Float32bits(enemy.PosVec.Y) ^ 1
					enemy.PosVec.Y = math.Float32frombits(bits)
				}
				wantOld := aiNoNewEnemyReferenceSquare546D34(unit.PosVec, old.PosVec)
				wantNew := aiNoNewEnemyReferenceSquare546D34(unit.PosVec, enemy.PosVec)
				gotOld := aiDependencyCenterDistanceSquared546D34(unit, old)
				gotNew := aiDependencyCenterDistanceSquared546D34(unit, enemy)
				if math.Float64bits(gotOld) != math.Float64bits(wantOld) || math.Float64bits(gotNew) != math.Float64bits(wantNew) {
					t.Fatalf("directed squared distances at %d differ: old=%016x/%016x new=%016x/%016x inputs=%v,%v,%v", i, math.Float64bits(gotOld), math.Float64bits(wantOld), math.Float64bits(gotNew), math.Float64bits(wantNew), unit.PosVec, old.PosVec, enemy.PosVec)
				}
				beforeSlot, beforeUpdate := slot, *update
				if got := aiDependencyNoNewEnemy546D1B(unit, update, &slot); got != (wantNew >= wantOld) {
					t.Fatalf("original C0 comparison at %d differs: got=%t old=%016x new=%016x", i, got, math.Float64bits(wantOld), math.Float64bits(wantNew))
				}
				if slot != beforeSlot || *update != beforeUpdate {
					t.Fatal("distance predicate changed the native cached update or action slot")
				}
			}
		})
	}
}

func TestAINoNewEnemyDependency546D1BRetainedSquareBits(t *testing.T) {
	for _, tc := range []struct {
		name               string
		unitX, unitY, x, y uint32
		want               uint64
	}{
		{"zero", 0, 0, 0, 0, 0},
		{"one", 0, 0, 0x3f800000, 0, 0x3ff0000000000000},
		{"above-one", 0, 0, 0x3f800000, 0x39000000, 0x3ff0000004000000},
		{"three-four-five-squared", 0, 0, 0x40400000, 0x40800000, 0x4039000000000000},
		{"unspilled-difference-squared", 0x4b800000, 0, 0xbf800000, 0, 0x42f0000020000010},
		{"positive-difference-chop", 1, 0, 0x3f800000, 0, 0x3feffffffffffffe},
		{"negative-difference-chop", 0x80000001, 0, 0xbf800000, 0, 0x3feffffffffffffe},
		{"smallest-subnormal-squared", 0, 0, 1, 0, 0x2d50000000000000},
		{"subnormal-diagonal-squared", 0, 0, 1, 1, 0x2d60000000000000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := server.Object{PosVec: types.Pointf{X: math.Float32frombits(tc.unitX), Y: math.Float32frombits(tc.unitY)}}
			target := server.Object{PosVec: types.Pointf{X: math.Float32frombits(tc.x), Y: math.Float32frombits(tc.y)}}
			if got := math.Float64bits(aiDependencyCenterDistanceSquared546D34(&unit, &target)); got != tc.want {
				t.Fatalf("original unspilled squared-distance bits=%016x want=%016x", got, tc.want)
			}
		})
	}
}

func TestAINoNewEnemyDependency546D1BMissingInputPrefix(t *testing.T) {
	_, unit, old, _ := aiAliveDependencyNative546A70(t)
	update := unit.UpdateDataMonster()
	for _, tc := range []struct {
		name                                            string
		noUnit, noUpdate, noSlot, noOld, noEnemy, valid bool
		fault                                           bool
	}{
		{name: "nil-slot", noSlot: true, fault: true},
		{name: "nil-old-before-nil-update", noOld: true, noUpdate: true},
		{name: "nil-old-before-nil-unit", noOld: true, noUnit: true},
		{name: "nil-old-before-nil-unit-and-update", noOld: true, noUnit: true, noUpdate: true},
		{name: "nil-enemy-before-nil-unit", noEnemy: true, noUnit: true, valid: true},
		{name: "nil-update-before-positions", noUpdate: true, fault: true},
		{name: "nil-unit-after-enemy", noUnit: true, fault: true},
		{name: "nil-unit-and-update", noUnit: true, noUpdate: true, fault: true},
		{name: "nil-slot-and-update", noSlot: true, noUpdate: true, fault: true},
		{name: "nil-all", noUnit: true, noUpdate: true, noSlot: true, noOld: true, noEnemy: true, fault: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			update.CurrentEnemy = old
			if tc.noEnemy {
				update.CurrentEnemy = nil
			}
			slot := server.AIStackItem{Args: [4]uintptr{uintptr(unsafe.Pointer(old)), 0x12345678, uintptr(unsafe.Pointer(unit)), 0xff800001}, Field5: 0xfedcba98}
			if tc.noOld {
				slot.Args[0] = 0
			}
			selectedUnit, selectedUpdate, selectedSlot := unit, update, &slot
			if tc.noUnit {
				selectedUnit = nil
			}
			if tc.noUpdate {
				selectedUpdate = nil
			}
			if tc.noSlot {
				selectedSlot = nil
			}
			beforeSlot, beforeUpdate := slot, *update
			var fault any
			var got bool
			func() {
				defer func() { fault = recover() }()
				got = aiDependencyNoNewEnemy546D1B(selectedUnit, selectedUpdate, selectedSlot)
			}()
			if (fault != nil) != tc.fault || got != tc.valid || slot != beforeSlot || *update != beforeUpdate {
				t.Fatalf("original input prefix differs: result=%t/%t fault=%v/%t or mutated cached data", got, tc.valid, fault, tc.fault)
			}
		})
	}
}

func TestAINoNewEnemyDependency546D1BCachedUpdateAndLivePositions(t *testing.T) {
	_, unit, old, _ := aiAliveDependencyNative546A70(t)
	enemy, freeEnemy := alloc.New(server.Object{})
	t.Cleanup(freeEnemy)
	other, freeOther := alloc.New(server.MonsterUpdateData{CurrentEnemy: old, Field97: 91, Field101: 92})
	t.Cleanup(freeOther)
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(enemy), unsafe.Pointer(other)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("replacement native enemy/update below 4 GiB: %p", pointer)
		}
	}
	cached := unit.UpdateDataMonster()
	cached.CurrentEnemy = enemy
	slot := server.AIStackItem{Args: [4]uintptr{uintptr(unsafe.Pointer(old)), 0x12345678, uintptr(unsafe.Pointer(unit)), 0xff800001}, Field5: 0xfedcba98}
	beforeUpdatePointer, beforeClass := unit.UpdateData, unit.ObjClass
	unit.UpdateData, unit.ObjClass = unsafe.Pointer(other), object.ClassFood
	defer func() { unit.UpdateData, unit.ObjClass = beforeUpdatePointer, beforeClass }()
	for _, tc := range []struct {
		name                string
		unitX, oldX, enemyX float32
		valid               bool
	}{
		{"cached-enemy-closer", 0, 2, 1, false},
		{"live-unit-position", 4, 2, 1, true},
		{"live-old-position", 4, 8, 1, false},
		{"live-enemy-position", 4, 8, 10, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit.PosVec, old.PosVec, enemy.PosVec = types.Pointf{X: tc.unitX}, types.Pointf{X: tc.oldX}, types.Pointf{X: tc.enemyX}
			beforeSlot, beforeCached, beforeOther := slot, *cached, *other
			if got := aiDependencyNoNewEnemy546D1B(unit, cached, &slot); got != tc.valid {
				t.Fatalf("cached enemy/live positions result=%t want=%t", got, tc.valid)
			}
			if slot != beforeSlot || *cached != beforeCached || *other != beforeOther || unit.UpdateData != unsafe.Pointer(other) || unit.ObjClass != object.ClassFood {
				t.Fatal("predicate changed cached/replacement native identities or class")
			}
		})
	}
}
