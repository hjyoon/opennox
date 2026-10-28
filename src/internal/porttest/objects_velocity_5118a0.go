package porttest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type VelocityObject5118A0 struct {
	Typ, Cls, Flags           uint32
	Pos, New, Vel, Force, Acc [2]uint32
	Drag                      uint32
	Health                    bool
}
type VelocityState5118A0 struct {
	Cache   [10]uint32
	Objects []VelocityObject5118A0
}
type VelocityEvent5118A0 struct {
	Kind  string
	Args  []int64
	State VelocityState5118A0
}
type VelocityInput5118A0 struct {
	Name           string
	Step           uint32
	Cache, Lookup  [10]uint32
	Objects        []VelocityObject5118A0
	Order0, Order1 []int
	Actions        [3]int32
	Buffs          [3][3]int32
	Traces, Tiles  [3]int32
	Mutation       struct {
		Stage  string
		At     int
		Action string
		Object int
	}
}
type velocityResult5118A0 struct {
	State  VelocityState5118A0
	Events []VelocityEvent5118A0
	Return int32 `json:"return_value"`
}
type velocityRow5118A0 struct {
	Input     VelocityInput5118A0
	Result    velocityResult5118A0
	Preserved bool `json:"all_nonoutput_bytes_preserved"`
}

// VelocityScene5118A0 declares the same external services as the whole
// original root oracle. Objects, health and the ten cache words are C-owned;
// all output values remain integer float representations in observations.
// This does not claim the transitive behavior of collision, springs or trace.
type VelocityScene5118A0 struct {
	Input                                              VelocityInput5118A0
	Objects                                            []*server.Object
	Cache                                              *[10]uint32
	Health                                             *server.HealthData
	Events                                             []VelocityEvent5118A0
	Errors                                             []string
	Calls                                              map[string]int
	Phase, Node, Mutations                             int
	objectFrames, beforeObjects                        [][]byte
	cacheFrame, beforeCache, healthFrame, beforeHealth []byte
}

func velocityPoint5118A0(v [2]uint32) types.Pointf {
	return types.Ptf(math.Float32frombits(v[0]), math.Float32frombits(v[1]))
}
func velocityWords5118A0(v types.Pointf) [2]uint32 {
	return [2]uint32{math.Float32bits(v.X), math.Float32bits(v.Y)}
}
func (s *VelocityScene5118A0) assign(p *server.Object, v VelocityObject5118A0) {
	p.TypeInd = uint16(v.Typ)
	p.ObjClass = object.Class(v.Cls)
	p.ObjFlags = object.Flags(v.Flags)
	p.PosVec = velocityPoint5118A0(v.Pos)
	p.NewPos = velocityPoint5118A0(v.New)
	p.VelVec = velocityPoint5118A0(v.Vel)
	p.ForceVec = velocityPoint5118A0(v.Force)
	p.Pos24 = velocityPoint5118A0(v.Acc)
	p.Float28 = math.Float32frombits(v.Drag)
	p.HealthData = nil
	if v.Health {
		p.HealthData = s.Health
	}
}
func (s *VelocityScene5118A0) Snapshot() VelocityState5118A0 {
	v := VelocityState5118A0{Cache: *s.Cache, Objects: make([]VelocityObject5118A0, 0, len(s.Objects))}
	for _, p := range s.Objects {
		v.Objects = append(v.Objects, VelocityObject5118A0{Typ: uint32(p.TypeInd), Cls: uint32(p.ObjClass), Flags: uint32(p.ObjFlags), Pos: velocityWords5118A0(p.PosVec), New: velocityWords5118A0(p.NewPos), Vel: velocityWords5118A0(p.VelVec), Force: velocityWords5118A0(p.ForceVec), Acc: velocityWords5118A0(p.Pos24), Drag: math.Float32bits(p.Float28), Health: p.HealthData != nil})
	}
	return v
}
func (s *VelocityScene5118A0) event(kind string, args ...int64) {
	if args == nil {
		args = []int64{}
	}
	s.Events = append(s.Events, VelocityEvent5118A0{Kind: kind, Args: args, State: s.Snapshot()})
}
func (s *VelocityScene5118A0) Index(p *server.Object) int {
	if p == nil {
		return -1
	}
	for i, q := range s.Objects {
		if p == q {
			return i
		}
	}
	s.Errors = append(s.Errors, fmt.Sprintf("unknown native object %p", p))
	return -1
}
func (s *VelocityScene5118A0) mutate(stage string, point *types.Pointf, ray *types.Rectf) {
	n := s.Calls[stage]
	s.Calls[stage]++
	m := s.Input.Mutation
	if m.Stage != stage || m.At != n {
		return
	}
	s.Mutations++
	p := s.Objects[m.Object]
	switch m.Action {
	case "clear-acc":
		p.Pos24 = types.Pointf{}
	case "motion":
		p.ForceVec = types.Ptf(-1.1, .29)
		p.Pos24 = types.Ptf(.4, -.6)
		p.Float28 = .125
	case "frozen":
		p.ObjFlags |= 2
	case "fly":
		p.ObjFlags |= 0x4000
		p.HealthData = nil
	case "health-on":
		p.HealthData = s.Health
	case "new":
		p.NewPos = types.Ptf(7.25, 9.5)
	case "reset-new":
		p.NewPos = p.PosVec
	case "point":
		if point != &p.NewPos {
			s.Errors = append(s.Errors, "tile did not receive native NewPos pointer")
		}
		if point != nil {
			*point = types.Ptf(7.25, 9.5)
		}
	case "trace-max":
		if ray == nil {
			s.Errors = append(s.Errors, "missing live trace pointer")
		} else {
			ray.Max = types.Ptf(7.25, 9.5)
		}
	case "cache":
		for i := range s.Cache {
			s.Cache[i] = uint32(700 + i)
		}
		p.TypeInd = 706
	default:
		s.Errors = append(s.Errors, "unknown mutation "+m.Action)
	}
}
func (s *VelocityScene5118A0) Lookup(name string) int32 {
	names := []string{"SmallFlameCleanse", "SmallFlameCleanse", "MediumFlameCleanse", "FlameCleanse", "LargeFlameCleanse", "SmallBlueFlameCleanse", "SmallBlueFlameCleanse", "MediumBlueFlameCleanse", "BlueFlameCleanse", "LargeBlueFlameCleanse"}
	n := s.Calls["lookup"]
	if n >= len(names) || name != names[n] {
		s.Errors = append(s.Errors, "type lookup order/name")
		return 0
	}
	s.event("lookup", int64(n))
	s.mutate("lookup", nil, nil)
	return int32(s.Input.Lookup[n])
}
func (s *VelocityScene5118A0) order() []int {
	if s.Phase == 0 {
		return s.Input.Order0
	}
	return s.Input.Order1
}
func (s *VelocityScene5118A0) First() *server.Object {
	s.Phase++
	if s.Phase > 1 {
		s.Errors = append(s.Errors, "extra enumeration pass")
		return nil
	}
	s.event("first", int64(s.Phase))
	s.mutate("first", nil, nil)
	if order := s.order(); len(order) != 0 {
		s.Node = order[0]
		return s.Objects[s.Node]
	}
	s.Node = -1
	return nil
}
func (s *VelocityScene5118A0) Next(p *server.Object) *server.Object {
	n := s.Index(p)
	s.event("next", int64(s.Phase), int64(n))
	s.mutate("next", nil, nil)
	order := s.order()
	for i, k := range order {
		if n == k {
			if i+1 < len(order) {
				s.Node = order[i+1]
				return s.Objects[s.Node]
			}
			s.Node = -1
			return nil
		}
	}
	s.Errors = append(s.Errors, "next object outside declared enumeration")
	return nil
}
func (s *VelocityScene5118A0) Prepare(p *server.Object) {
	s.event("prepare", int64(s.Index(p)))
	s.mutate("prepare", nil, nil)
}
func (s *VelocityScene5118A0) Springs() { s.event("springs"); s.mutate("springs", nil, nil) }
func (s *VelocityScene5118A0) Action(p *server.Object, v int32) int32 {
	n := s.Index(p)
	s.event("action", int64(n), int64(v))
	s.mutate("action", nil, nil)
	if v != 67 || n < 0 {
		s.Errors = append(s.Errors, "action argument")
		return 0
	}
	return s.Input.Actions[n]
}
func (s *VelocityScene5118A0) Buff(p *server.Object, v int32) int32 {
	n := s.Index(p)
	s.event("buff", int64(n), int64(v))
	s.mutate("buff", nil, nil)
	slot := map[int32]int{5: 0, 25: 1, 28: 2}
	k, ok := slot[v]
	if !ok || n < 0 {
		s.Errors = append(s.Errors, "buff argument")
		return 0
	}
	return s.Input.Buffs[n][k]
}
func (s *VelocityScene5118A0) Trace(ray *types.Rectf, flags byte) int32 {
	s.event("trace", int64(math.Float32bits(ray.Min.X)), int64(math.Float32bits(ray.Min.Y)), int64(math.Float32bits(ray.Max.X)), int64(math.Float32bits(ray.Max.Y)), int64(flags), 0, 0)
	s.mutate("trace", nil, ray)
	return s.Input.Traces[s.Node]
}
func (s *VelocityScene5118A0) Tile(point *types.Pointf) int32 {
	if point != &s.Objects[s.Node].NewPos {
		s.Errors = append(s.Errors, "tile input copied from native object")
	}
	s.event("tile", int64(s.Node), int64(math.Float32bits(point.X)), int64(math.Float32bits(point.Y)))
	s.mutate("tile", point, nil)
	return s.Input.Tiles[s.Node]
}
func (s *VelocityScene5118A0) Collision(p *server.Object, kind uintptr, normal *types.Pointf) {
	s.event("collision", int64(s.Index(p)), int64(kind), int64(math.Float32bits(normal.X)), int64(math.Float32bits(normal.Y)))
	s.mutate("collision", nil, nil)
}
func (s *VelocityScene5118A0) Sync(p *server.Object) {
	s.event("sync", int64(s.Index(p)))
	s.mutate("sync", nil, nil)
}
func (s *VelocityScene5118A0) Coords(p *server.Object) {
	s.event("coords", int64(s.Index(p)))
	s.mutate("coords", nil, nil)
}
func (s *VelocityScene5118A0) Move(p *server.Object) {
	s.event("move", int64(s.Index(p)))
	s.mutate("move", nil, nil)
}

func (s *VelocityScene5118A0) Runtime() server.ObjectsVelocityRuntime5118A0 {
	return server.ObjectsVelocityRuntime5118A0{
		Cache: s.Cache, Lookup: s.Lookup, First: s.First, Next: s.Next,
		Prepare: s.Prepare, Springs: s.Springs, Action: s.Action, Buff: s.Buff,
		Trace: s.Trace, Tile: s.Tile, Collision: s.Collision,
		Sync: s.Sync, Coords: s.Coords, Move: s.Move,
	}
}

// RunVelocityOriginal5118A0 compares every external call and complete modeled
// state against independently executed original instructions, then checks
// every byte of every native object, health and cache frame with canaries.
func RunVelocityOriginal5118A0(t *testing.T, path string, invoke func(*VelocityScene5118A0) int32) {
	t.Helper()
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Original string `json:"original_sha256"`
		CW       uint32 `json:"control_word"`
		FLDPatch string `json:"emulator_fld_patch_sha256"`
		Rows     []velocityRow5118A0
	}
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.CW != 0xe3f || len(fixture.Rows) != 426 {
		t.Fatal("whole original velocity fixture identity/count/control")
	}
	if fixture.FLDPatch != "183d9b99af3f1eec4f6a951a8a7e6c314ea7615a4ebdc66f02697e3764d07e12" {
		t.Fatal("uncalibrated emulator FLD m32fp NaN behavior")
	}
	for _, want := range fixture.Rows {
		t.Run(want.Input.Name, func(t *testing.T) {
			if !want.Preserved {
				t.Fatal("uncertified original object frame")
			}
			s := &VelocityScene5118A0{Input: want.Input, Calls: make(map[string]int), Phase: -1, Node: -1, Events: []VelocityEvent5118A0{}}
			frame := func(size uintptr) []byte {
				p, free := alloc.Make([]byte(nil), int(size)+64)
				t.Cleanup(free)
				for i := range p {
					p[i] = 0xa5
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&p[32])) <= 0xffffffff {
					t.Fatal("native C frame below4GiB")
				}
				return p
			}
			s.healthFrame = frame(unsafe.Sizeof(server.HealthData{}))
			s.Health = (*server.HealthData)(unsafe.Pointer(&s.healthFrame[32]))
			s.beforeHealth = bytes.Clone(s.healthFrame)
			s.cacheFrame = frame(40)
			s.Cache = (*[10]uint32)(unsafe.Pointer(&s.cacheFrame[32]))
			*s.Cache = s.Input.Cache
			s.beforeCache = bytes.Clone(s.cacheFrame)
			for _, v := range s.Input.Objects {
				f := frame(unsafe.Sizeof(server.Object{}))
				p := (*server.Object)(unsafe.Pointer(&f[32]))
				s.assign(p, v)
				s.Objects = append(s.Objects, p)
				s.objectFrames = append(s.objectFrames, f)
				s.beforeObjects = append(s.beforeObjects, bytes.Clone(f))
			}
			result := invoke(s)
			if len(s.Errors) != 0 {
				t.Error(s.Errors)
			}
			if s.Phase != 1 || s.Mutations != map[bool]int{false: 0, true: 1}[s.Input.Mutation.Stage != ""] {
				t.Errorf("passes/mutations = %d/%d", s.Phase, s.Mutations)
			}
			got := velocityResult5118A0{State: s.Snapshot(), Events: s.Events, Return: result}
			if !reflect.DeepEqual(got, want.Result) {
				x, _ := json.Marshal(got)
				y, _ := json.Marshal(want.Result)
				t.Errorf("whole velocity state/events differ: got=%s want=%s", x, y)
			}
			for i, actual := range s.objectFrames {
				expected := bytes.Clone(s.beforeObjects[i])
				s.assign((*server.Object)(unsafe.Pointer(&expected[32])), got.State.Objects[i])
				if !bytes.Equal(actual, expected) {
					t.Errorf("native object%d changed nonoutput bytes/canaries", i)
				}
				s.assign((*server.Object)(unsafe.Pointer(&expected[32])), want.Result.State.Objects[i])
				if !bytes.Equal(actual, expected) {
					t.Errorf("native object%d complete frame differs from original", i)
				}
			}
			expected := bytes.Clone(s.beforeCache)
			*(*[10]uint32)(unsafe.Pointer(&expected[32])) = got.State.Cache
			if !bytes.Equal(s.cacheFrame, expected) || !bytes.Equal(s.healthFrame, s.beforeHealth) {
				t.Error("cache/health nonoutput bytes/canaries changed")
			}
			*(*[10]uint32)(unsafe.Pointer(&expected[32])) = want.Result.State.Cache
			if !bytes.Equal(s.cacheFrame, expected) {
				t.Error("complete cache frame differs from original")
			}
		})
	}
}
