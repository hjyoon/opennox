package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type playerCircleEvent518040 struct {
	Kind, Candidate string
	Bits, Center    []uint32
	Radius          uint32
}

type playerCircleMutation518040 struct {
	Center []uint32
}

type playerCircleCandidate518040 struct {
	Name          string
	Position      [2]uint32
	Radius, Shape uint32
	BoxBits       uint64                     `json:"box_bits"`
	OnBox         playerCircleMutation518040 `json:"on_box"`
	OnVisit       playerCircleMutation518040 `json:"on_visit"`
}

type playerCircleCase518040 struct {
	Name        string
	Radius      uint32
	Center      [2]uint32
	Candidates  []playerCircleCandidate518040
	OnEnumerate playerCircleMutation518040 `json:"on_enumerate"`
	NilCenter   bool                       `json:"nil_center"`
}

// Unhooked original enumeration/filter with supplied precision53/ToZero.
// Candidate enumeration, box returns and visit mutations are declared
// services; this is not original wholeworld, attack damage or caller CW proof.
func TestPlayerAttackCircle518040OriginalNative(t *testing.T) {
	b, err := os.ReadFile("testdata/player_attack_circle_518040.json")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != "c940e2a9fbf6d69b4a5b58962527cf338c5ce8af16fa4225355f071cd1795710" {
		t.Fatal("changed unchanged-PE32 bounded circle oracle")
	}
	var oracle struct {
		Rows []struct {
			Input       playerCircleCase518040 `json:"case"`
			Events      []playerCircleEvent518040
			FinalCenter [2]uint32 `json:"final_center"`
		}
	}
	if err := json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	if len(oracle.Rows) != 56 {
		t.Fatal("lost original bounded circle cases")
	}
	for _, row := range oracle.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			center, freeCenter := alloc.New(types.Pointf{})
			t.Cleanup(freeCenter)
			*center = types.Ptf(math.Float32frombits(row.Input.Center[0]), math.Float32frombits(row.Input.Center[1]))
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(center)) <= math.MaxUint32 {
				t.Fatal("center must retain its live C-owned >4 GiB pointer")
			}
			inputs := make(map[*Object]playerCircleCandidate518040)
			before := make(map[*Object][]byte)
			var candidates []*Object
			for _, input := range row.Input.Candidates {
				candidate, free := alloc.New(Object{})
				t.Cleanup(free)
				candidate.PosVec = types.Ptf(math.Float32frombits(input.Position[0]), math.Float32frombits(input.Position[1]))
				candidate.Shape.Kind, candidate.Shape.Circle.R = ShapeKind(input.Shape), math.Float32frombits(input.Radius)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(candidate)) <= math.MaxUint32 {
					t.Fatal("candidate must retain its C-owned >4 GiB pointer")
				}
				inputs[candidate], candidates = input, append(candidates, candidate)
				before[candidate] = bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(candidate)), int(unsafe.Sizeof(*candidate))))
			}
			centerBits := func() []uint32 { return []uint32{math.Float32bits(center.X), math.Float32bits(center.Y)} }
			mutate := func(change playerCircleMutation518040) {
				if len(change.Center) != 0 {
					if len(change.Center) != 2 {
						t.Fatal("invalid original center mutation")
					}
					*center = types.Ptf(math.Float32frombits(change.Center[0]), math.Float32frombits(change.Center[1]))
				}
			}
			events := []playerCircleEvent518040{}
			eachRect := func(rect types.Rectf, callback func(*Object) bool) {
				events = append(events, playerCircleEvent518040{Kind: "rect", Bits: []uint32{
					math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y)}})
				mutate(row.Input.OnEnumerate)
				for _, candidate := range candidates {
					if !callback(candidate) {
						t.Fatal("original void visit cannot terminate enumeration")
					}
				}
			}
			box := func(pos types.Pointf, radius float32, candidate *Object) (float64, types.Pointf) {
				input, ok := inputs[candidate]
				if !ok || pos != *center || math.Float32bits(radius) != row.Input.Radius {
					t.Fatal("changed original box service arguments")
				}
				events = append(events, playerCircleEvent518040{Kind: "box", Candidate: input.Name, Center: centerBits(), Radius: math.Float32bits(radius)})
				mutate(input.OnBox)
				return math.Float64frombits(input.BoxBits), types.Pointf{}
			}
			visit := func(candidate *Object) {
				input, ok := inputs[candidate]
				if !ok {
					t.Fatal("truncated or substituted candidate")
				}
				events = append(events, playerCircleEvent518040{Kind: "visit", Candidate: input.Name, Center: centerBits()})
				mutate(input.OnVisit)
			}
			argument := center
			if row.Input.NilCenter {
				argument = nil
			}
			PlayerAttackScanCircle518040(argument, math.Float32frombits(row.Input.Radius), eachRect, box, visit)
			if !reflect.DeepEqual(events, row.Events) || !reflect.DeepEqual(centerBits(), row.FinalCenter[:]) {
				t.Errorf("events=%+v center=%v; original=%+v/%v", events, centerBits(), row.Events, row.FinalCenter)
			}
			for candidate, snapshot := range before {
				if !bytes.Equal(snapshot, unsafe.Slice((*byte)(unsafe.Pointer(candidate)), int(unsafe.Sizeof(*candidate)))) {
					t.Fatal("shape filter changed candidate record/HP/flags/pointers")
				}
			}
		})
	}
}

func TestPlayerAttackCircle518040NilGuards(t *testing.T) {
	t.Run("center", func(t *testing.T) {
		PlayerAttackScanCircle518040(nil, 1, nil, nil, func(*Object) { t.Fatal("nil center visited") })
	})
	t.Run("visit", func(t *testing.T) {
		PlayerAttackScanCircle518040(&types.Pointf{}, 1, nil, nil, nil)
	})
	t.Run("candidate", func(t *testing.T) {
		PlayerAttackScanCircle518040(&types.Pointf{}, 1, func(_ types.Rectf, callback func(*Object) bool) {
			if !callback(nil) {
				t.Fatal("nil candidate must not terminate enumeration")
			}
		}, nil, func(*Object) { t.Fatal("nil candidate visited") })
	})
}
