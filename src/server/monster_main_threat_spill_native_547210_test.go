package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Independent integer encoding of FSTS/FSTPS under x87 ToZero. This models
// subnormal truncation and finite overflow without float32 rounding or any
// production spill/Nextafter helper. Arithmetic NaNs are quieted.
func monsterThreatSpillReference5475D5(value float64) uint32 {
	word := math.Float64bits(value)
	sign := uint32(word>>63) << 31
	exponent := int(word >> 52 & 0x7ff)
	mantissa := word & 0x000fffffffffffff
	if exponent == 0x7ff {
		if mantissa != 0 {
			return sign | 0x7f800000 | uint32(mantissa>>29) | 0x00400000
		}
		return sign | 0x7f800000
	}
	if exponent == 0 && mantissa == 0 {
		return sign
	}
	unbiased := exponent - 1023
	if exponent == 0 {
		unbiased = -1022
	} else {
		mantissa |= 1 << 52
	}
	if unbiased > 127 {
		return sign | 0x7f7fffff
	}
	if unbiased >= -126 {
		return sign | uint32(unbiased+127)<<23 | uint32(mantissa>>29)&0x007fffff
	}
	shift := -unbiased - 97
	if shift >= 64 {
		return sign
	}
	return sign | uint32(mantissa>>uint(shift))
}

// The +30 FADDS is precision 53/ToZero before the binary32 FSTPS. In
// particular, a tiny negative range still changes the stored word below 30.
func monsterThreatRangeReference547704(value float32) uint32 {
	word := math.Float32bits(value)
	if word&0x7f800000 == 0x7f800000 {
		if word&0x007fffff != 0 {
			return word | 0x00400000
		}
		return word
	}
	input := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(float64(value))
	constant := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetInt64(30)
	sum := new(big.Float).SetPrec(53).SetMode(big.ToZero).Add(input, constant)
	retained, accuracy := sum.Float64()
	if accuracy != big.Exact {
		panic("independent retained range must be binary64")
	}
	return monsterThreatSpillReference5475D5(retained)
}

func monsterThreatSpillFixture547210(t *testing.T) (*Server, *Object, *MonsterUpdateData, *Object) {
	t.Helper()
	s, unit, update := monsterHealthRetreatNativeFixture547807(t)
	enemy, freeEnemy := alloc.New(Object{})
	definition, freeDefinition := alloc.New(MonsterDef{})
	t.Cleanup(freeEnemy)
	t.Cleanup(freeDefinition)
	unit.PosVec, unit.NewPos = types.Ptf(100, 100), types.Ptf(100, 100)
	unit.Direction1, unit.Direction2 = 250, 29
	enemy.PosVec = types.Ptf(120, 100)
	*unit.HealthData = HealthData{Cur: 100, Max: 100}
	*update = MonsterUpdateData{
		AIStackInd: 0, Aggression: 0.5, FleeRange: 65,
		StatusFlags: object.MonStatusCanCastSpells, Field376: 1,
		Field371: s.Frame(), Field370_0: 5, Field370_2: 5,
		CurrentEnemy: enemy, PreferredEnemy: enemy, MonsterDef: definition,
	}
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_FIGHT), Args: [4]uintptr{11, 22, 33, 44}}
	for name, pointer := range map[string]unsafe.Pointer{
		"unit": unsafe.Pointer(unit), "update": unsafe.Pointer(update),
		"enemy": unsafe.Pointer(enemy), "definition": unsafe.Pointer(definition),
	} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("C-owned %s must exceed 4 GiB: %p", name, pointer)
		}
	}
	return s, unit, update, enemy
}

func monsterThreatSpillAssert547210(t *testing.T, s *Server, unit *Object, update *MonsterUpdateData, enemy *Object, distance float64, public, defaultDistance bool) {
	t.Helper()
	const reject, flee, blink = 0, 1, 2
	rangeValue := float64(update.FleeRange)
	want := reject
	if distance < rangeValue || math.IsNaN(distance) || math.IsNaN(rangeValue) {
		if rangeValue*0.5 > float64(math.Float32frombits(monsterThreatSpillReference5475D5(distance))) {
			want = blink
		} else if update.FleeRange != 0 && !math.IsNaN(rangeValue) {
			want = flee
		}
	}
	if public && want == reject {
		t.Fatal("a rejected threat branch does not prove the full MainAI tail")
	}
	unitBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	enemyBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(enemy), unsafe.Sizeof(*enemy))
	updateBefore := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))
	healthBefore, base := *unit.HealthData, update.AIStack[0]
	logicIndex, otherIndex := s.Rand.Logic.Index(), s.Rand.Other.Index()
	distanceCalls, casts, rolls := 0, 0, 0
	services := MonsterMainRuntime547210{
		CastSpell: func(id int32, got *Object, arg *SpellAcceptArg) {
			casts++
			if want != blink || id != 4 || got != unit || arg.Obj != unit || arg.Pos != unit.PosVec || update.AIStackInd != 0 {
				t.Fatal("immediate Blink lost native arguments or pre-stack order")
			}
		},
		RandomInt: func(min, max int) int {
			rolls++
			if want == blink {
				if casts != 1 || min != 5 || max != 5 {
					t.Fatal("Blink cooldown RNG did not follow casting")
				}
				return 5
			}
			if want != flee || casts != 0 || min != 0 || max != 1 || update.AIStackHead().Type() != ai.ACTION_FLEE {
				t.Fatal("fallback RNG did not follow the original FLEE transition")
			}
			return 0
		},
		AudioEvent: func(uint32, *Object) { t.Fatal("zero ordinary-flee sound roll emitted audio") },
	}
	if !defaultDistance {
		services.Distance = func(got, target *Object) float64 {
			distanceCalls++
			if got != unit || target != enemy {
				t.Fatal("distance callback lost native identities")
			}
			return distance
		}
	}
	var handled bool
	if public {
		handled = s.MonsterMainNativeRuntime547210(unit, services)
	} else {
		handled = s.monsterMainThreatFlee547210(unit, update, nil, services)
	}
	if handled != (want != reject) || distanceCalls != map[bool]int{false: 1, true: 0}[defaultDistance] {
		t.Errorf("threat handled=%t want=%d distance calls=%d default=%t", handled, want, distanceCalls, defaultDistance)
	}
	switch want {
	case blink:
		if casts != 1 || rolls != 1 || update.AIStackInd != 0 || update.AIStack[0] != base || update.Field371 != s.Frame()+5 {
			t.Errorf("chopped distance should Blink: distance=%016x spill=%08x stack=%+v casts/rolls=%d/%d", math.Float64bits(distance), monsterThreatSpillReference5475D5(distance), update.GetAIStack(), casts, rolls)
		}
		binary.LittleEndian.PutUint32(updateBefore[int(unsafe.Offsetof(update.Field371)):], s.Frame()+5)
		if !bytes.Equal(updateBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
			t.Error("Blink changed cached update outside its cooldown word")
		}
	case flee:
		if casts != 0 || rolls != 1 || update.AIStackInd != 4 || update.AIStack[0] != base {
			t.Error("original four-entry FLEE transition changed")
		} else {
			wantActions := []ai.ActionType{ai.ACTION_SET_ANGLE, ai.DEPENDENCY_NOT_CORNERED, ai.DEPENDENCY_ENEMY_CLOSER_THAN, ai.ACTION_FLEE}
			for i, action := range wantActions {
				if update.AIStack[i+1].Type() != action {
					t.Errorf("action %d=%v want=%v", i, update.AIStack[i+1].Type(), action)
				}
			}
			if update.AIStack[1].ArgU32(0) != uint32(int32(int16(unit.Direction1))+128) ||
				update.AIStack[3].ArgU32(0) != monsterThreatRangeReference547704(update.FleeRange) ||
				update.AIStack[4].Args != [4]uintptr{uintptr(math.Float32bits(enemy.PosVec.X)), uintptr(math.Float32bits(enemy.PosVec.Y)), 0, 0} {
				t.Error("native FLEE argument words differ from the original")
			}
		}
	case reject:
		if casts != 0 || rolls != 0 || !bytes.Equal(updateBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
			t.Error("rejected retained outer comparison changed cached state")
		}
	}
	if !bytes.Equal(unitBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
		!bytes.Equal(enemyBefore, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(enemy), unsafe.Sizeof(*enemy))) ||
		*unit.HealthData != healthBefore || s.Rand.Logic.Index() != logicIndex || s.Rand.Other.Index() != otherIndex {
		t.Fatal("distance spill changed native objects, health or either game RNG")
	}
}

func TestMonsterMainThreatSpill547210NativeReferenceCalibration(t *testing.T) {
	for _, tc := range []struct {
		value float64
		word  uint32
	}{
		{32.4999999, 0x4201ffff}, {-32.4999999, 0xc201ffff},
		{32.5, 0x42020000}, {math.Ldexp(3, -150), 1}, {math.Ldexp(-3, -150), 0x80000001},
		{math.Ldexp(1, -150), 0}, {math.Copysign(0, -1), 0x80000000},
		{float64(math.MaxFloat32) * 2, 0x7f7fffff}, {-float64(math.MaxFloat32) * 2, 0xff7fffff},
		{math.Inf(1), 0x7f800000}, {math.Inf(-1), 0xff800000},
	} {
		if got := monsterThreatSpillReference5475D5(tc.value); got != tc.word {
			t.Fatalf("independent spill=%08x want=%08x input=%016x", got, tc.word, math.Float64bits(tc.value))
		}
	}
	for _, tc := range []struct{ input, word uint32 }{
		{0x00000001, 0x41f00000}, {0x80000001, 0x41efffff},
		{0x7f7fffff, 0x7f7fffff}, {0xff7fffff, 0xff7ffffe},
		{0x42820000, 0x42be0000}, {0xc1f00000, 0},
	} {
		if got := monsterThreatRangeReference547704(math.Float32frombits(tc.input)); got != tc.word {
			t.Fatalf("independent range=%08x want=%08x input=%08x", got, tc.word, tc.input)
		}
	}
}

func TestMonsterMainThreatSpill547210NativeDistanceBoundary(t *testing.T) {
	for _, tc := range []struct {
		name      string
		distance  float64
		rangeWord uint32
	}{
		{"half-below-nearest", 32.4999999, 0x42820000},
		{"half-below-one-double-ulp", math.Nextafter(32.5, 0), 0x42820000},
		{"half-equal-control", 32.5, 0x42820000},
		{"half-above-control", math.Nextafter(32.5, math.Inf(1)), 0x42820000},
		{"half-next-float32-below-control", float64(math.Nextafter32(32.5, 0)), 0x42820000},
		{"outer-retained-below-control", 64.9999999, 0x42820000},
		{"outer-equal-control", 65, 0x42820000},
		{"outer-above-control", math.Nextafter(65, math.Inf(1)), 0x42820000},
		{"subnormal-quarter-below", math.Ldexp(3, -151), 1},
		{"subnormal-seven-quarters", math.Ldexp(7, -151), 3},
		{"subnormal-exact-control", math.Ldexp(1, -148), 3},
		{"positive-zero-control", 0, 0x42820000},
		{"negative-zero-control", math.Copysign(0, -1), 0x42820000},
		{"finite-overflow-positive", float64(math.MaxFloat32) * 2, 0x7f800000},
		{"infinite-distance-control", math.Inf(1), 0x7f800000},
		{"negative-infinity-control", math.Inf(-1), 0x42820000},
		{"unordered-distance-control", math.NaN(), 0x42820000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, update, enemy := monsterThreatSpillFixture547210(t)
			update.FleeRange = math.Float32frombits(tc.rangeWord)
			monsterThreatSpillAssert547210(t, s, unit, update, enemy, tc.distance, false, false)
		})
	}
}

func TestMonsterMainThreatSpill547210NativeDefaultDistance(t *testing.T) {
	for _, kind := range []ShapeKind{ShapeKindCenter, ShapeKindCircle, ShapeKindBox} {
		for _, offset := range []float32{float32(math.Ldexp(1, -24)), 0, -float32(math.Ldexp(1, -24))} {
			for _, public := range []bool{false, true} {
				t.Run(fmt.Sprintf("kind-%d/offset-%08x/public-%t", kind, math.Float32bits(offset), public), func(t *testing.T) {
					s, unit, update, enemy := monsterThreatSpillFixture547210(t)
					unit.PosVec, enemy.PosVec = types.Ptf(32.5, 0), types.Ptf(offset, 0)
					unit.Shape, enemy.Shape = Shape{Kind: kind}, Shape{Kind: kind}
					if kind != ShapeKindCenter {
						unit.PosVec.X = 34.5
						unit.Shape.Circle.R, enemy.Shape.Circle.R = 1, 1
						unit.Shape.Box, enemy.Shape.Box = ShapeBox{W: 2, H: 1}, ShapeBox{W: 1, H: 2}
					}
					distance := objectDistanceReferenceRetained4E6C00(unit, enemy)
					monsterThreatSpillAssert547210(t, s, unit, update, enemy, distance, public, true)
				})
			}
		}
	}
}

func TestMonsterMainThreatSpill547210NativePostCancelRange(t *testing.T) {
	for _, word := range []uint32{0, 0x80000000, 1, 0x80000001, 0x33800000, 0xb3800000, 0x3f800001, 0xbf800001, 0x41800001, 0xc1800001, 0x42820000, 0xc1f00000, 0x4e800001, 0xce800001, 0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000, 0x7fc12345, 0xffc12345} {
		t.Run(fmt.Sprintf("range-%08x", word), func(t *testing.T) {
			s, unit, update, enemy := monsterThreatSpillFixture547210(t)
			update.StatusFlags, update.Field376 = 0, 0
			update.AIStack[0].Action, update.AIStack[0].Field5 = uint32(ai.ACTION_WAIT), 1
			previous, existed := aiActions[ai.ACTION_WAIT]
			t.Cleanup(func() {
				if existed {
					aiActions[ai.ACTION_WAIT] = previous
				} else {
					delete(aiActions, ai.ACTION_WAIT)
				}
			})
			cancels, rolls := 0, 0
			aiActions[ai.ACTION_WAIT] = monsterMainFearCancel547210{cancel: func(got *Object) {
				cancels++
				if got != unit {
					t.Fatal("Cancel lost the native unit")
				}
				update.FleeRange = math.Float32frombits(word)
			}}
			healthBefore := *unit.HealthData
			logicIndex, otherIndex := s.Rand.Logic.Index(), s.Rand.Other.Index()
			if !s.monsterMainThreatFlee547210(unit, update, nil, MonsterMainRuntime547210{
				Distance: func(got, target *Object) float64 {
					if got != unit || target != enemy {
						t.Fatal("distance identities changed")
					}
					return 20
				},
				RandomInt: func(min, max int) int {
					rolls++
					if min != 0 || max != 1 || cancels != 1 || update.AIStackInd != 4 {
						t.Fatal("post-cancel FLEE/RNG order changed")
					}
					return 0
				},
			}) || cancels != 1 || rolls != 1 || update.AIStackHead().Type() != ai.ACTION_FLEE {
				t.Fatal("accepted original transition missing")
			}
			if got, want := update.AIStack[3].ArgU32(0), monsterThreatRangeReference547704(math.Float32frombits(word)); got != want {
				t.Errorf("post-push FSTPS word=%08x want=%08x cached range=%08x", got, want, word)
			}
			if *unit.HealthData != healthBefore || s.Rand.Logic.Index() != logicIndex || s.Rand.Other.Index() != otherIndex {
				t.Fatal("range store changed HP or game RNG")
			}
		})
	}
}

func TestMonsterMainThreatSpill547210NativeIndependentFiniteMatrix(t *testing.T) {
	s, unit, update, enemy := monsterThreatSpillFixture547210(t)
	random := rand.New(rand.NewSource(0x5475d5)) // test-local, not either game RNG
	for sample := 0; sample < 2048; sample++ {
		word := random.Uint32() & 0x7fffffff
		if word&0x7f800000 == 0x7f800000 {
			word ^= 0x00800000
		}
		if word == 0 {
			word = 1
		}
		update.FleeRange = math.Float32frombits(word)
		update.Field371, update.AIStackInd = s.Frame(), 0
		update.AIStack = [24]AIStackItem{}
		update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_FIGHT), Args: [4]uintptr{11, 22, 33, 44}}
		half := float64(update.FleeRange) * 0.5
		distance := half
		switch sample % 4 {
		case 0:
			distance = math.Nextafter(half, 0)
		case 1:
			distance = math.Nextafter(half, math.Inf(1))
		case 2:
			distance = -half
		}
		monsterThreatSpillAssert547210(t, s, unit, update, enemy, distance, false, false)
		if t.Failed() {
			t.Fatalf("independent finite fixture=%d range=%08x", sample, word)
		}
	}
}
