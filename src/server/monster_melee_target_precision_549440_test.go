package server

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/big"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type meleeTargetEvent549440 struct {
	Kind, Candidate, Target, Source, Weapon string
	Bits                                    []uint32
	Flags, Damage, Type                     uint32
}

type meleeTargetCandidate549440 struct {
	Name          string
	Position      [2]uint32
	Radius, Class uint32
	Flags         uint32
	Enemy         bool
	HPMax         uint16 `json:"hp_max"`
	Mutation      map[string]uint32
}

type meleeTargetOracle549440 struct {
	SourceSHA256         string `json:"source_sha256"`
	DirectionTableSHA256 string `json:"direction_table_sha256"`
	Results              []struct {
		Input struct {
			Name          string
			UnitPosition  [2]uint32 `json:"unit_position"`
			Radius, Range uint32
			Direction     uint16
			AllowFriendly bool `json:"allow_friendly"`
			Indexed       bool
			Strike        bool
			Trace         *int32
			Candidates    []meleeTargetCandidate549440
		}
		Selected      *string
		Result        int32
		RectangleBits [4]uint32 `json:"rectangle_bits"`
		Events        []meleeTargetEvent549440
	}
}

func readMeleeTargetOracle549440(t *testing.T) meleeTargetOracle549440 {
	t.Helper()
	b, err := os.ReadFile("testdata/monster_melee_target_549440.json")
	if err != nil {
		t.Fatal(err)
	}
	var out meleeTargetOracle549440
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.SourceSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(out.Results) != 23 {
		t.Fatal("changed bounded original oracle")
	}
	return out
}

// Candidate/allegiance/trace/damage services are boundaries, not a battle.
// The two indexed cases run real native partition enumeration and compare
// its complete callback/strike order against an unchanged PE32 partition.
// Recorded position cases have supplied, not captured, radii of 12.
func TestMonsterPickMeleeTarget549440OriginalNative(t *testing.T) {
	for _, row := range readMeleeTargetOracle549440(t).Results {
		t.Run(row.Input.Name, func(t *testing.T) {
			s, unit, update := monsterHealthRetreatNativeFixture547807(t)
			def, freeDef := alloc.New(MonsterDef{})
			t.Cleanup(freeDef)
			update.MonsterDef = def
			def.MeleeAttackRange112 = math.Float32frombits(row.Input.Range)
			def.MeleeAttackDamage116, def.MeleeAttackDamageType124 = 8, 0
			unit.PosVec = types.Ptf(math.Float32frombits(row.Input.UnitPosition[0]), math.Float32frombits(row.Input.UnitPosition[1]))
			unit.Direction1, unit.Shape.Circle.R = Dir16(row.Input.Direction), math.Float32frombits(row.Input.Radius)
			var candidates []*Object
			byObject := make(map[*Object]meleeTargetCandidate549440)
			for _, input := range row.Input.Candidates {
				candidate, free := alloc.New(Object{})
				t.Cleanup(free)
				candidate.ObjClass, candidate.ObjFlags = object.Class(input.Class), object.Flags(input.Flags)
				candidate.PosVec = types.Ptf(math.Float32frombits(input.Position[0]), math.Float32frombits(input.Position[1]))
				candidate.Shape.Circle.R = math.Float32frombits(input.Radius)
				if input.HPMax != 0 {
					hp, freeHP := alloc.New(HealthData{})
					t.Cleanup(freeHP)
					hp.Max, candidate.HealthData = input.HPMax, hp
				}
				candidates = append(candidates, candidate)
				byObject[candidate] = input
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(candidate)) <= math.MaxUint32 {
					t.Fatal("candidate must retain its C-owned native high pointer")
				}
			}
			if row.Input.Indexed {
				s.Map.Init()
				t.Cleanup(s.Map.Free)
				// Insert backwards because the production bucket prepends.
				for i := len(candidates) - 1; i >= 0; i-- {
					candidate := candidates[i]
					candidate.ObjFlags |= object.FlagActive
					candidate.NewPos = candidate.PosVec
					candidate.Shape.Kind = ShapeKindCircle
					s.Map.AddObjectToIndex(candidate)
				}
			}
			logic, other, frame, hp := s.Rand.Logic.Index(), s.Rand.Other.Index(), s.Frame(), *unit.HealthData
			var events []meleeTargetEvent549440
			var selected *Object
			pick := func(got *Object, friendly bool) *Object {
				if got != unit {
					t.Fatal("lost native unit")
				}
				selected = monsterPickMeleeTarget549440(got, friendly, monsterPickMeleeTargetHooks549440{
					eachInRect: func(rect types.Rectf, fn func(*Object) bool) {
						bits := [4]uint32{math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y)}
						if bits != row.RectangleBits {
							t.Errorf("search rectangle=%08x original=%08x", bits, row.RectangleBits)
						}
						events = append(events, meleeTargetEvent549440{Kind: "rect", Bits: bits[:]})
						if row.Input.Indexed {
							s.Map.EachObjInRect(rect, fn)
						} else {
							for _, candidate := range candidates {
								if !fn(candidate) {
									t.Fatal("original candidate callback must not stop the scan")
								}
							}
						}
					},
					isEnemy: func(source, candidate *Object) bool {
						input := byObject[candidate]
						if source != unit {
							t.Fatal("changed original allegiance source")
						}
						events = append(events, meleeTargetEvent549440{Kind: "enemy", Candidate: input.Name})
						for field, value := range input.Mutation {
							if field != "56" {
								t.Fatal("unexpected oracle mutation")
							}
							candidate.PosVec.X = math.Float32frombits(value)
						}
						return input.Enemy
					},
				})
				return selected
			}
			if row.Input.Strike {
				result := monsterStrikeDefault549380(unit, monsterStrikeDefaultHooks549380{
					loadUpdate: func(*Object) *MonsterUpdateData { return update }, pickTarget: pick,
					loadUnitY: func(o *Object) float32 { return o.PosVec.Y }, loadUnitX: func(o *Object) float32 { return o.PosVec.X },
					loadTargetX: func(o *Object) float32 { return o.PosVec.X }, loadTargetY: func(o *Object) float32 { return o.PosVec.Y },
					trace: func(from, to types.Pointf, flags MapTraceFlags) int32 {
						events = append(events, meleeTargetEvent549440{Kind: "trace", Flags: uint32(flags), Bits: []uint32{
							math.Float32bits(from.X), math.Float32bits(from.Y), math.Float32bits(to.X), math.Float32bits(to.Y)}})
						if row.Input.Trace != nil {
							return *row.Input.Trace
						}
						return 1
					},
					loadMonsterDef: func(u *MonsterUpdateData) *MonsterDef { return u.MonsterDef },
					loadDamageType: func(d *MonsterDef) uint32 { return d.MeleeAttackDamageType124 },
					loadDamage:     func(d *MonsterDef) uint32 { return d.MeleeAttackDamage116 },
					damage: func(target, source, weapon *Object, damage, typ uint32) {
						if source != unit || weapon != unit {
							t.Fatal("changed native-width source/weapon attribution")
						}
						events = append(events, meleeTargetEvent549440{Kind: "damage", Target: byObject[target].Name,
							Source: "unit", Weapon: "unit", Damage: damage, Type: typ})
					},
					loadImpact: func(d *MonsterDef) float32 { return d.MeleeAttackImpact120 },
					applyForce: func(*Object, *Object, float32) { t.Fatal("zero impact dispatched force") },
				})
				if result != row.Result {
					t.Errorf("strike result=%d original=%d", result, row.Result)
				}
			} else {
				pick(unit, row.Input.AllowFriendly)
			}
			got := ""
			if selected != nil {
				got = byObject[selected].Name
			}
			want := ""
			if row.Selected != nil {
				want = *row.Selected
			}
			if got != want || !reflect.DeepEqual(events, row.Events) {
				t.Errorf("selected=%q original=%q; events=%+v original=%+v", got, want, events, row.Events)
			}
			if *unit.HealthData != hp || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || s.Frame() != frame {
				t.Fatal("selection/dispatch probe changed HP, clocks or a game RNG")
			}
		})
	}
}

func TestMonsterPickMeleeTarget549440OriginalDirectionTable(t *testing.T) {
	b := make([]byte, 256*8)
	for direction := 0; direction < 256; direction++ {
		x, y := SinCosDir(byte(direction))
		binary.LittleEndian.PutUint32(b[8*direction:], math.Float32bits(x))
		binary.LittleEndian.PutUint32(b[8*direction+4:], math.Float32bits(y))
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != readMeleeTargetOracle549440(t).DirectionTableSHA256 {
		t.Fatal("native direction table differs from unchanged original PE32 table")
	}
}

// Independent big.Float reference models each retained arithmetic operation
// at precision 53 / ToZero, then only the original binary32 FSTP stores.
// It shares no production FMA, Nextafter or arithmetic helper.
func TestMonsterPickMeleeTarget549440AllDirectionsPrecision(t *testing.T) {
	op := func(a, b float64, code byte) float64 {
		x := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(a)
		y := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(b)
		z := new(big.Float).SetPrec(53).SetMode(big.ToZero)
		switch code {
		case '+':
			z.Add(x, y)
		case '*':
			z.Mul(x, y)
		case '/':
			z.Quo(x, y)
		case 's':
			z.Sqrt(x)
		}
		v, _ := z.Float64()
		return v
	}
	spill := func(v float64) float64 {
		x := new(big.Float).SetPrec(24).SetMode(big.ToZero).SetFloat64(v)
		f, _ := x.Float32()
		return float64(f)
	}
	unit := meleeMonsterTestObject532130(t)
	unit.PosVec = types.Ptf(100.00001, 200.00002)
	unit.Shape.Circle.R = 4
	unit.UpdateDataMonster().MonsterDef.MeleeAttackRange112 = 30
	for direction := 0; direction < 256; direction++ {
		unit.Direction1 = Dir16(direction)
		for _, position := range []types.Pointf{types.Ptf(110, 217.31935), types.Ptf(90, 182.68065), types.Ptf(125.001, 200)} {
			candidate := &Object{ObjClass: object.ClassPlayer, PosVec: position}
			candidate.Shape.Circle.R = 1
			dx := op(float64(position.X), -float64(unit.PosVec.X), '+')
			dy := spill(op(float64(position.Y), -float64(unit.PosVec.Y), '+'))
			distance := spill(op(op(op(op(dy, dy, '*'), op(dx, dx, '*'), '+'), 0, 's'), float64(math.Float32frombits(0x3a83126f)), '+'))
			x, y := SinCosDir(byte(direction))
			dot := op(op(op(dy, distance, '/'), float64(y), '*'), op(op(dx, distance, '/'), float64(x), '*'), '+')
			want := dot > 0.5 && op(distance, -5, '+') < 30
			got := monsterPickMeleeTarget549440(unit, false, monsterPickMeleeTargetHooks549440{
				eachInRect: func(_ types.Rectf, fn func(*Object) bool) { fn(candidate) },
				isEnemy:    func(*Object, *Object) bool { return true },
			}) != nil
			if got != want {
				t.Errorf("direction=%d position=%v admitted=%t original-precision=%t", direction, position, got, want)
			}
		}
	}
}
