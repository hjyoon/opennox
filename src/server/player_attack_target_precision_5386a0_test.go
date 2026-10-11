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
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type playerTargetEvent5386A0 struct {
	Kind, Candidate string
	Flags, Radius   uint32
	Center          []uint32
}

type playerTargetCandidate5386A0 struct {
	Name                 string
	Position             [2]uint32
	Radius, Shape, Flags uint32
	Class                uint32 `json:"cls"`
	Enemy, Interact      bool
	BoxBits              uint64            `json:"box_bits"`
	OnEnemy              map[string]uint32 `json:"on_enemy"`
	OnInteract           map[string]uint32 `json:"on_interact"`
}

type playerTargetCase5386A0 struct {
	Name       string
	Position   [2]uint32
	Range      uint32
	Direction  uint16
	Candidates []playerTargetCandidate5386A0
}

type playerTargetOracle5386A0 struct {
	DirectionTableSHA256 string   `json:"direction_table_sha256"`
	DirectionTable       []uint32 `json:"direction_table"`
	Rows                 []struct {
		Input    playerTargetCase5386A0 `json:"case"`
		Selected string
		Distance uint32
		Events   []playerTargetEvent5386A0
	}
}

func readPlayerTargetOracle5386A0(t *testing.T) playerTargetOracle5386A0 {
	t.Helper()
	b, err := os.ReadFile("testdata/player_attack_target_5386a0.json")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != "91cd2f0e85e335d7b3f427331f75719f2febfb0b68f4b17ba572f4c248007b55" {
		t.Fatal("changed unchanged-PE32 bounded candidate oracle")
	}
	var oracle playerTargetOracle5386A0
	if err := json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	if len(oracle.Rows) != 58 || len(oracle.DirectionTable) != 512 {
		t.Fatal("lost original bounded cases/directions")
	}
	return oracle
}

// This compares the unhooked original candidate function. Allegiance,
// visibility and box return are declared service boundaries, not a battle.
func TestPlayerAttackTarget5386A0OriginalNative(t *testing.T) {
	oracle := readPlayerTargetOracle5386A0(t)
	for _, row := range oracle.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			s := New(nil, nil, strman.New())
			t.Cleanup(s.Close)
			attacker, free := alloc.New(Object{})
			t.Cleanup(free)
			attacker.ObjClass = object.ClassPlayer
			attacker.PosVec = types.Ptf(math.Float32frombits(row.Input.Position[0]), math.Float32frombits(row.Input.Position[1]))
			attacker.Direction1 = Dir16(row.Input.Direction)
			var closest *Object
			distance := math.Float32frombits(row.Input.Range)
			var events []playerTargetEvent5386A0
			inputs := make(map[*Object]playerTargetCandidate5386A0)
			var candidates []*Object
			for _, input := range row.Input.Candidates {
				candidate, free := alloc.New(Object{})
				t.Cleanup(free)
				candidate.ObjClass, candidate.ObjFlags = object.Class(input.Class), object.Flags(input.Flags)
				candidate.PosVec = types.Ptf(math.Float32frombits(input.Position[0]), math.Float32frombits(input.Position[1]))
				candidate.Shape.Kind, candidate.Shape.Circle.R = ShapeKind(input.Shape), math.Float32frombits(input.Radius)
				inputs[candidate], candidates = input, append(candidates, candidate)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(candidate)) <= math.MaxUint32 {
					t.Fatal("candidate must retain its C-owned >4 GiB pointer")
				}
			}
			mutate := func(candidate *Object, changes map[string]uint32) {
				for field, raw := range changes {
					switch field {
					case "x":
						candidate.PosVec.X = math.Float32frombits(raw)
					case "y":
						candidate.PosVec.Y = math.Float32frombits(raw)
					case "flags":
						candidate.ObjFlags = object.Flags(raw)
					case "class":
						candidate.ObjClass = object.Class(raw)
					case "radius":
						candidate.Shape.Circle.R = math.Float32frombits(raw)
					case "range":
						distance = math.Float32frombits(raw)
					case "owner_x":
						attacker.PosVec.X = math.Float32frombits(raw)
					default:
						t.Fatal("unknown oracle mutation", field)
					}
				}
			}
			hooks := PlayerAttackTargetRuntime5386A0{
				IsEnemy: func(owner, candidate *Object) bool {
					if owner != attacker {
						t.Fatal("changed allegiance owner")
					}
					input := inputs[candidate]
					events = append(events, playerTargetEvent5386A0{Kind: "enemy", Candidate: input.Name})
					mutate(candidate, input.OnEnemy)
					return input.Enemy
				},
				CanInteract: func(owner, candidate *Object, flags int) bool {
					if owner != attacker || flags != 1 {
						t.Fatal("changed visibility arguments")
					}
					input := inputs[candidate]
					events = append(events, playerTargetEvent5386A0{Kind: "interact", Candidate: input.Name, Flags: uint32(flags)})
					mutate(candidate, input.OnInteract)
					return input.Interact
				},
				Direction: func(direction Dir16) types.Pointf {
					if direction != attacker.Direction1 {
						t.Fatal("truncated native direction WORD")
					}
					i := int(direction) * 2
					return types.Ptf(math.Float32frombits(oracle.DirectionTable[i]), math.Float32frombits(oracle.DirectionTable[i+1]))
				},
				Box: func(center types.Pointf, radius float32, candidate *Object) (float64, types.Pointf) {
					input := inputs[candidate]
					events = append(events, playerTargetEvent5386A0{Kind: "box", Candidate: input.Name, Radius: math.Float32bits(radius),
						Center: []uint32{math.Float32bits(center.X), math.Float32bits(center.Y)}})
					return math.Float64frombits(input.BoxBits), types.Pointf{}
				},
			}
			logic, other, frame := s.Rand.Logic.Index(), s.Rand.Other.Index(), s.Frame()
			for _, candidate := range candidates {
				PlayerAttackSelectCandidate5386A0(candidate, attacker, &distance, &closest, hooks)
			}
			selected := ""
			if closest != nil {
				selected = inputs[closest].Name
			}
			if selected != row.Selected || math.Float32bits(distance) != row.Distance || !reflect.DeepEqual(events, row.Events) {
				t.Errorf("selected=%q range=%08x events=%+v; original=%q/%08x/%+v", selected, math.Float32bits(distance), events, row.Selected, row.Distance, row.Events)
			}
			if s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || s.Frame() != frame {
				t.Fatal("selection changed a game RNG or clock")
			}
		})
	}
}

func TestPlayerAttackTarget5386A0OriginalDirectionTable(t *testing.T) {
	b := make([]byte, 2048)
	for direction := 0; direction < 256; direction++ {
		x, y := SinCosDir(byte(direction))
		binary.LittleEndian.PutUint32(b[direction*8:], math.Float32bits(x))
		binary.LittleEndian.PutUint32(b[direction*8+4:], math.Float32bits(y))
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != readPlayerTargetOracle5386A0(t).DirectionTableSHA256 {
		t.Fatal("native direction table differs from unchanged PE32")
	}
}

// Independent reference shares no production arithmetic/FMA/Nextafter.
func TestPlayerAttackTarget5386A0AllDirectionsPrecision(t *testing.T) {
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
	attacker := &Object{PosVec: types.Ptf(100.00001, 200.00002)}
	for direction := 0; direction < 256; direction++ {
		attacker.Direction1 = Dir16(direction)
		for _, position := range []types.Pointf{types.Ptf(110, 217.31935), types.Ptf(90, 182.68065), types.Ptf(125.001, 200)} {
			candidate := &Object{ObjClass: object.ClassMonster, PosVec: position}
			candidate.Shape.Kind, candidate.Shape.Circle.R = ShapeKindCircle, 1
			dx := spill(op(float64(position.X), -float64(attacker.PosVec.X), '+'))
			dy := op(float64(position.Y), -float64(attacker.PosVec.Y), '+')
			sy := spill(dy)
			length := op(op(op(dy, sy, '*'), op(dx, dx, '*'), '+'), 0, 's')
			x, y := SinCosDir(byte(direction))
			dot := op(op(op(sy, length, '/'), float64(y), '*'), op(op(dx, length, '/'), float64(x), '*'), '+')
			edge := max(op(length, -1, '+'), 0)
			want := dot > 0.5 && edge < 30
			var chosen *Object
			distance := float32(30)
			PlayerAttackSelectCandidate5386A0(candidate, attacker, &distance, &chosen, PlayerAttackTargetRuntime5386A0{
				IsEnemy: func(*Object, *Object) bool { return true }, CanInteract: func(*Object, *Object, int) bool { return true },
				Direction: func(Dir16) types.Pointf { return types.Ptf(x, y) },
				Box: func(types.Pointf, float32, *Object) (float64, types.Pointf) {
					t.Fatal("circle dispatched box")
					return 0, types.Pointf{}
				},
			})
			wantRange := float32(30)
			if want {
				wantRange = float32(spill(edge))
			}
			if (chosen != nil) != want || math.Float32bits(distance) != math.Float32bits(wantRange) {
				t.Errorf("direction=%d position=%v selected=%t range=%08x reference=%t/%08x", direction, position, chosen != nil, math.Float32bits(distance), want, math.Float32bits(wantRange))
			}
		}
	}
}

func TestPlayerAttackTarget5386A0NullSelfGuards(t *testing.T) {
	a, b := new(Object), new(Object)
	distance := float32(30)
	var closest *Object
	h := PlayerAttackTargetRuntime5386A0{}
	PlayerAttackSelectCandidate5386A0(nil, a, &distance, &closest, h)
	PlayerAttackSelectCandidate5386A0(b, nil, &distance, &closest, h)
	PlayerAttackSelectCandidate5386A0(a, a, &distance, &closest, h)
	PlayerAttackSelectCandidate5386A0(b, a, nil, &closest, h)
	PlayerAttackSelectCandidate5386A0(b, a, &distance, nil, h)
	if closest != nil || distance != 30 {
		t.Fatal("rejected candidate changed search state")
	}
}
