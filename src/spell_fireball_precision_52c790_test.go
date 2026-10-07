package opennox

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Independent arithmetic oracle for GAME.EXE 0052C790..0052C8C7. Each
// register operation has 53 significant bits, round-toward-zero; the FST(S)
// instructions spill binary32 at different points in X and Y. This does not
// call the production rounding helpers or run the original Windows program.
func fireballReference53_52C790(a, b float64, multiply bool) float64 {
	x := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(a)
	y := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(b)
	if multiply {
		x.Mul(x, y)
	} else {
		x.Add(x, y)
	}
	value, _ := x.Float64()
	return value
}

func fireballReferenceSpill32_52C790(value float64) float32 {
	if value == 0 {
		return float32(value)
	}
	abs := math.Abs(value)
	if abs < float64(math.Float32frombits(0x00800000)) {
		// Precision 24 alone does not model the subnormal exponent boundary.
		// Compute the exact integer multiple of 2^-149 instead.
		ratio := new(big.Rat).SetFloat64(abs)
		ratio.Mul(ratio, new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 149)))
		mantissa := new(big.Int).Quo(ratio.Num(), ratio.Denom()).Uint64()
		bits := uint32(mantissa)
		if math.Signbit(value) {
			bits |= 0x80000000
		}
		return math.Float32frombits(bits)
	}
	rounded := new(big.Float).SetPrec(24).SetMode(big.ToZero).SetFloat64(value)
	stored, _ := rounded.Float32()
	if math.IsInf(float64(stored), 0) {
		if math.Signbit(value) {
			return -math.MaxFloat32
		}
		return math.MaxFloat32
	}
	return stored
}

func fireballReferenceSpawn52C790(from, velocity types.Pointf, radius, cosine, sine float32) types.Pointf {
	radius2 := fireballReference53_52C790(float64(radius), float64(radius), false)
	x := fireballReference53_52C790(radius2, float64(cosine), true)
	x = fireballReference53_52C790(x, float64(from.X), false)
	x = float64(fireballReferenceSpill32_52C790(x)) // 0052C7FD
	x = fireballReference53_52C790(x, float64(velocity.X), false)
	y := fireballReference53_52C790(radius2, float64(sine), true)
	y = fireballReference53_52C790(y, float64(from.Y), false)
	y = fireballReference53_52C790(y, float64(velocity.Y), false)
	return types.Ptf(fireballReferenceSpill32_52C790(x), fireballReferenceSpill32_52C790(y))
}

func fireballReferenceFlight52C790(velocity types.Pointf, baseSpeed float32, coefficient float64, cosine, sine float32) (float32, types.Pointf) {
	speed := fireballReference53_52C790(coefficient, float64(baseSpeed), true)
	x := fireballReference53_52C790(speed, float64(cosine), true)
	x = float64(fireballReferenceSpill32_52C790(x)) // 0052C876/0052C87A
	x = fireballReference53_52C790(x, float64(velocity.X), false)
	y := fireballReference53_52C790(speed, float64(sine), true)
	// 0052C881 stores Y without popping its unspilled register value.
	y = fireballReference53_52C790(y, float64(velocity.Y), false)
	return fireballReferenceSpill32_52C790(speed), types.Ptf(fireballReferenceSpill32_52C790(x), fireballReferenceSpill32_52C790(y))
}

func fireballAssertPointBits52C790(t *testing.T, label string, got, want types.Pointf) {
	t.Helper()
	if math.Float32bits(got.X) != math.Float32bits(want.X) || math.Float32bits(got.Y) != math.Float32bits(want.Y) {
		t.Fatalf("%s = (%08x, %08x), want (%08x, %08x)", label,
			math.Float32bits(got.X), math.Float32bits(got.Y), math.Float32bits(want.X), math.Float32bits(want.Y))
	}
}

func fireballPrecisionRecords52C790(t *testing.T) (*server.Object, *server.Object) {
	t.Helper()
	caster, freeCaster := alloc.New(server.Object{})
	projectile, freeProjectile := alloc.New(server.Object{})
	t.Cleanup(freeCaster)
	t.Cleanup(freeProjectile)
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(caster), unsafe.Pointer(projectile)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("C-owned Fireball object is not above 4 GiB: %p", pointer)
		}
	}
	return caster, projectile
}

// These finite binary32 inputs exercise the actual native cast function, all
// 256 stock direction entries and both trace outcomes. World services are
// spies, so this is an arithmetic/callback contract, not a stock-map hit test.
func TestCastFireballNative52C790OriginalPrecision(t *testing.T) {
	for _, input := range []struct {
		name           string
		from, velocity types.Pointf
		radius, speed  float32
		coefficient    float64
	}{
		{"ordinary", types.Ptf(3426.8457, 2309.9163), types.Ptf(-0.137, 0.943), 12.125, 14.375, float64(float32(1.35))},
		{"cancellation", types.Ptf(1000000, -1000000), types.Ptf(-1000000, 1000000), math.Float32frombits(0x3f012345), 13.7, float64(float32(1.23))},
		{"dense", types.Ptf(math.Float32frombits(0x42a519da), -math.Float32frombits(0x44062193)), types.Ptf(0.037, -0.8), 7.714, math.Float32frombits(0x41d17463), float64(float32(0.85))},
		{"subnormal", types.Ptf(math.Float32frombits(3), -math.Float32frombits(7)), types.Ptf(-math.Float32frombits(2), math.Float32frombits(11)), math.Float32frombits(0x00812345), math.Float32frombits(0x000fffff), float64(float32(1.333))},
	} {
		for direction := 0; direction < 256; direction++ {
			for _, clear := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/dir-%03d/clear-%t", input.name, direction, clear), func(t *testing.T) {
					caster, projectile := fireballPrecisionRecords52C790(t)
					*caster = server.Object{PosVec: input.from, VelVec: input.velocity,
						Direction1: server.Dir16(direction), Direction2: 91, Field29: 0x12345678}
					caster.Shape.Circle.R = input.radius
					*projectile = server.Object{SpeedCur: input.speed, SpeedBase: 37, PosVec: types.Ptf(9, 17), Field29: 0xabcdef01}
					beforeCaster, beforeProjectile := *caster, *projectile
					cosine, sine := server.SinCosDir(byte(direction)) // Table input only.
					candidate := fireballReferenceSpawn52C790(input.from, input.velocity, input.radius, cosine, sine)
					spawn := candidate
					if !clear {
						spawn = input.from
					}
					wantSpeed, wantVelocity := fireballReferenceFlight52C790(input.velocity, input.speed, input.coefficient, cosine, sine)
					level := direction%5 + 1
					wantType := []string{"Fireball", "StrongFireball", "TitanFireball", "TitanFireball", "TitanFireball"}[level-1]
					var events []string
					got := castFireballNative52C790(spell.SPELL_FIREBALL, caster, level, fireballCastHooks52C790{
						newObject: func(kind string) *server.Object {
							events = append(events, "new")
							if kind != wantType {
								t.Fatalf("projectile type = %q, want %q", kind, wantType)
							}
							return projectile
						},
						traceRay: func(from, to types.Pointf) bool {
							events = append(events, "trace")
							fireballAssertPointBits52C790(t, "trace-from", from, input.from)
							fireballAssertPointBits52C790(t, "trace-to", to, candidate)
							return clear
						},
						createAt: func(obj, owner *server.Object, position types.Pointf) {
							events = append(events, "create")
							if obj != projectile || owner != caster {
								t.Fatal("create lost a native object identity")
							}
							fireballAssertPointBits52C790(t, "spawn", position, spawn)
						},
						speedCoeff: func(index int) float64 {
							events = append(events, "balance")
							if index != level-1 {
								t.Fatalf("balance index = %d, want %d", index, level-1)
							}
							return input.coefficient
						},
						playCastAudio: func(id spell.ID, owner *server.Object) {
							events = append(events, "audio")
							if id != spell.SPELL_FIREBALL || owner != caster {
								t.Fatal("audio lost spell/caster identity")
							}
						},
					})
					if got != 1 || !reflect.DeepEqual(events, []string{"new", "trace", "create", "balance", "audio"}) {
						t.Fatalf("result/events = %d/%v", got, events)
					}
					if math.Float32bits(projectile.SpeedCur) != math.Float32bits(wantSpeed) {
						t.Fatalf("speed = %08x, want %08x", math.Float32bits(projectile.SpeedCur), math.Float32bits(wantSpeed))
					}
					fireballAssertPointBits52C790(t, "velocity", projectile.VelVec, wantVelocity)
					if projectile.Direction1 != beforeCaster.Direction1 || projectile.Direction2 != beforeCaster.Direction1 || *caster != beforeCaster {
						t.Fatal("directions differ or the cast changed its caster")
					}
					remaining := *projectile
					remaining.SpeedCur, remaining.VelVec = beforeProjectile.SpeedCur, beforeProjectile.VelVec
					remaining.Direction1, remaining.Direction2 = beforeProjectile.Direction1, beforeProjectile.Direction2
					if remaining != beforeProjectile {
						t.Fatal("cast changed a projectile field outside speed/velocity/directions")
					}
				})
			}
		}
	}
}

// Callback changes are explicit synthetic inputs. Cache geometry after
// allocation, keep the pre-trace fallback, then read live speed/velocity and
// direction after placement/balance, as the original instruction stream does.
func TestCastFireballNative52C790OriginalCallbackReads(t *testing.T) {
	for _, stage := range []string{"none", "new", "trace", "create", "balance"} {
		for _, clear := range []bool{false, true} {
			for level := 1; level <= 5; level++ {
				t.Run(fmt.Sprintf("%s/clear-%t/level-%d", stage, clear, level), func(t *testing.T) {
					caster, projectile := fireballPrecisionRecords52C790(t)
					*caster = server.Object{PosVec: types.Ptf(107.137, 89.421), VelVec: types.Ptf(-0.7, 1.3), Direction1: 31, Field29: 17}
					caster.Shape.Circle.R = 12.125
					*projectile = server.Object{SpeedCur: 13.7, SpeedBase: 71, Field29: 29}
					mutate := func(at string) {
						if stage == at {
							caster.PosVec, caster.VelVec = types.Ptf(811.765, 903.431), types.Ptf(0.97, -1.73)
							caster.Direction1, caster.Shape.Circle.R = 197, 7.714
							projectile.SpeedCur = 27.3
						}
					}
					coefficient := float64(float32(1.35))
					var from, candidate types.Pointf
					var cosine, sine, wantSpeed float32
					var wantVelocity types.Pointf
					var wantDirection server.Dir16
					var beforeCaster, beforeProjectile server.Object
					var events []string
					got := castFireballNative52C790(spell.SPELL_FIREBALL, caster, level, fireballCastHooks52C790{
						newObject: func(kind string) *server.Object {
							events = append(events, "new")
							wantType := []string{"Fireball", "StrongFireball", "TitanFireball", "TitanFireball", "TitanFireball"}[level-1]
							if kind != wantType {
								t.Fatalf("type = %q, want %q", kind, wantType)
							}
							mutate("new")
							from = caster.PosVec
							cosine, sine = server.SinCosDir(byte(caster.Direction1))
							candidate = fireballReferenceSpawn52C790(from, caster.VelVec, caster.Shape.Circle.R, cosine, sine)
							return projectile
						},
						traceRay: func(start, end types.Pointf) bool {
							events = append(events, "trace")
							fireballAssertPointBits52C790(t, "trace-from", start, from)
							fireballAssertPointBits52C790(t, "trace-to", end, candidate)
							mutate("trace")
							return clear
						},
						createAt: func(obj, owner *server.Object, position types.Pointf) {
							events = append(events, "create")
							if obj != projectile || owner != caster {
								t.Fatal("create lost native identities")
							}
							want := candidate
							if !clear {
								want = from
							}
							fireballAssertPointBits52C790(t, "cached spawn", position, want)
							mutate("create")
						},
						speedCoeff: func(index int) float64 {
							events = append(events, "balance")
							if index != level-1 {
								t.Fatalf("balance index = %d", index)
							}
							mutate("balance")
							beforeCaster, beforeProjectile = *caster, *projectile
							wantSpeed, wantVelocity = fireballReferenceFlight52C790(caster.VelVec, projectile.SpeedCur, coefficient, cosine, sine)
							wantDirection = caster.Direction1
							return coefficient
						},
						playCastAudio: func(id spell.ID, owner *server.Object) {
							events = append(events, "audio")
							if id != spell.SPELL_FIREBALL || owner != caster {
								t.Fatal("audio lost native identities")
							}
							if math.Float32bits(projectile.SpeedCur) != math.Float32bits(wantSpeed) {
								t.Fatal("audio observed wrong speed")
							}
							fireballAssertPointBits52C790(t, "pre-audio velocity", projectile.VelVec, wantVelocity)
							if projectile.Direction1 != wantDirection || projectile.Direction2 != wantDirection {
								t.Fatal("audio did not observe the live final direction")
							}
						},
					})
					if got != 1 || !reflect.DeepEqual(events, []string{"new", "trace", "create", "balance", "audio"}) || *caster != beforeCaster {
						t.Fatalf("result/events/caster = %d/%v/%t", got, events, *caster == beforeCaster)
					}
					remaining := *projectile
					remaining.SpeedCur, remaining.VelVec = beforeProjectile.SpeedCur, beforeProjectile.VelVec
					remaining.Direction1, remaining.Direction2 = beforeProjectile.Direction1, beforeProjectile.Direction2
					if remaining != beforeProjectile {
						t.Fatal("cast changed unrelated projectile fields")
					}
				})
			}
		}
	}
}
