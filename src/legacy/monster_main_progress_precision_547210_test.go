package legacy

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

// Model the independent x87 instruction boundaries at 00547A19..00547A35.
// A 53-bit big.Float operation chops after each FSUB/FMUL/FADD, as selected
// by CRT 004031F2 and the gameplay frame's control word at 0043E2C1. It does
// not call a production arithmetic helper or inherit the ARM64 compiler's
// permission to contract a multiply and add into one instruction.
func monsterProgressReference547210(old, live types.Pointf) float64 {
	value := func(x float32) *big.Float {
		return new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(float64(x))
	}
	dx := value(old.X).Sub(value(old.X), value(live.X))
	dy := value(old.Y).Sub(value(old.Y), value(live.Y))
	ySquared := value(0).Mul(dy, dy)
	xSquared := value(0).Mul(dx, dx)
	result, accuracy := value(0).Add(ySquared, xSquared).Float64()
	if accuracy != big.Exact {
		panic("finite 53-bit MainAI reference outside binary64 exponent range")
	}
	return result
}

// These are branch inputs in C-owned records above 4 GiB, not mutations of a
// live map, the public food test, Quest seed or expected campaign sequence.
// Use the actual legacy MainAI wrapper, server stack and Logic RNG. At the
// fused-only boundary ARM64 used to reset progress instead of entering the
// original timeout branch, or changed progress before timeout expiry.
func TestMonsterMainProgress547210NativeArithmeticBoundaries(t *testing.T) {
	for _, vector := range []struct {
		name                        string
		oldX, oldY, liveX, liveY    uint32
		chopWord                    uint64
		moved, fusedOnly, unordered bool
	}{
		{"fused-tie-1", 0x41100000, 0x41400000, 0x32a00005, 0xb2700007, 0x406c1ffffffffffe, false, true, false},
		{"fused-tie-2", 0x41100000, 0x41400000, 0x32a00005, 0xb2700008, 0x406c200000000000, false, true, false},
		{"fused-tie-3", 0x41100000, 0x41400000, 0x32a0000d, 0xb2700013, 0x406c1ffffffffffe, false, true, false},
		{"fused-tie-4", 0x41100000, 0x41400000, 0x32a00015, 0xb2700020, 0x406c200000000000, false, true, false},
		{"rounding-tie-1", 0x41100000, 0x41400000, 0x32a00001, 0xb2700002, 0x406c1fffffffffff, false, false, false},
		{"rounding-tie-2", 0x41100000, 0x41400000, 0x32a00001, 0xb2700003, 0x406c1fffffffffff, false, false, false},
		{"rounding-tie-3", 0x41100000, 0x41400000, 0x32a00002, 0xb2700003, 0x406c1ffffffffffe, false, false, false},
		{"rounding-tie-4", 0x41100000, 0x41400000, 0x32a00002, 0xb2700004, 0x406c200000000000, false, false, false},
		{"exact-225", 0x41100000, 0x41400000, 0, 0, 0x406c200000000000, false, false, false},
		{"below-225", 0x41100000, 0x413fffff, 0, 0, 0, false, false, false},
		{"above-225", 0x41100000, 0x41400001, 0, 0, 0, true, false, false},
		{"old-qNaN", 0x7fc12345, 0x41400000, 0, 0, 0, false, false, true},
		{"live-sNaN", 0x41100000, 0x41400000, 0x7f800001, 0, 0, false, false, true},
		{"old-infinity", 0x7f800000, 0x41400000, 0, 0, 0, true, false, true},
		{"infinity-minus-infinity", 0x7f800000, 0x41400000, 0x7f800000, 0, 0, false, false, true},
	} {
		for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MOVE_TO_HOME, ai.ACTION_ROAM, ai.ACTION_FLEE} {
			for _, elapsed := range []uint32{14, 15, 16} {
				t.Run(fmt.Sprintf("%s/%s/elapsed-%d", vector.name, action, elapsed), func(t *testing.T) {
					srv, unit, update, _ := monsterMainBlockLegacyFixture547210(t)
					frame := uint32(996)
					srv.SetTickRate(30)
					srv.SetFrame(frame)
					srv.Rand.Logic, srv.Rand.Other = prand.New(1496), prand.New(3905)
					unit.ObjSubClass, unit.Buffs = 0, 0
					unit.PosVec = types.Ptf(math.Float32frombits(vector.liveX), math.Float32frombits(vector.liveY))
					unit.NewPos, unit.PrevPos, unit.SpeedCur = unit.PosVec, unit.PosVec, 10
					*update = server.MonsterUpdateData{
						Aggression: 0.5, AIStackInd: 0,
						Field124: frame - elapsed, Field125: vector.oldX, Field126: vector.oldY,
						Field363: frame + 1,
					}
					update.AIStack[0].Action = uint32(action)
					old := types.Ptf(math.Float32frombits(vector.oldX), math.Float32frombits(vector.oldY))
					if !vector.unordered {
						reference := monsterProgressReference547210(old, unit.PosVec)
						if (reference > 225) != vector.moved {
							t.Fatalf("independent original arithmetic = %.17g, moved=%t", reference, vector.moved)
						}
						if vector.chopWord != 0 && math.Float64bits(reference) != vector.chopWord {
							t.Fatalf("independent original chop bits=%016x want=%016x", math.Float64bits(reference), vector.chopWord)
						}
						if vector.fusedOnly {
							dx, dy := float64(old.X)-float64(unit.PosVec.X), float64(old.Y)-float64(unit.PosVec.Y)
							if math.Float64bits(math.FMA(dy, dy, dx*dx)) != 0x406c200000000001 {
								t.Fatal("fixture no longer distinguishes the original sum from ARM64 FMA")
							}
						}
					}
					beforeUnit := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))...)
					beforeHealth, beforeUpdate := *unit.HealthData, *update
					stalled := !vector.moved && elapsed > 15
					wantRNG := prand.New(1496)
					var duration int
					if stalled {
						if chance := wantRNG.Int(0, 100); chance < 33 {
							t.Fatalf("isolated Logic index does not select WAIT: chance=%d", chance)
						}
						duration = wantRNG.Int(15, 60)
					}
					Nox_xxx_monsterMainAIFn_547210(unit)
					if !bytes.Equal(beforeUnit, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) || *unit.HealthData != beforeHealth ||
						srv.Rand.Other.Index() != 3905 || srv.Rand.Logic.Index() != wantRNG.Index() {
						t.Fatalf("unit/health/Other RNG or Logic draw count changed: Logic=%d want=%d", srv.Rand.Logic.Index(), wantRNG.Index())
					}
					if stalled {
						beforeUpdate.StatusFlags |= object.MonStatusFrustrated
						// The real WAIT push also resets the original action timer.
						beforeUpdate.Field137 = frame
						beforeUpdate.AIStackInd = 1
						beforeUpdate.AIStack[1] = server.AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{uintptr(frame + uint32(duration))}}
						if action == ai.ACTION_FLEE {
							beforeUpdate.Field127 = frame
						}
					}
					if vector.moved || stalled {
						beforeUpdate.Field124, beforeUpdate.Field125, beforeUpdate.Field126 = frame, vector.liveX, vector.liveY
					}
					if *update != beforeUpdate || srv.AI.StackChanged != stalled {
						t.Fatalf("original progress/WAIT branch differs: moved=%t stalled=%t status=%#x progress=%d stack=%+v", vector.moved, stalled, update.StatusFlags, update.Field124, update.GetAIStack())
					}
				})
			}
		}
	}
}
