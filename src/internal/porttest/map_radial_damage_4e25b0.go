package porttest

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image"
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

type radialObject4E25B0 struct {
	Pos      [2]uint32
	Flags    uint32
	Callback bool
}
type radialContext4E25B0 struct {
	Center                    int
	Outer, Inner, Damage, Typ uint32
	Source, Excluded          int
}
type radialState4E25B0 struct {
	Objects []radialObject4E25B0
	Points  [2][2]uint32
	Flag    uint32
	Context *radialContext4E25B0
}
type radialEvent4E25B0 struct {
	Service string
	Args    []int64
	State   radialState4E25B0
}
type RadialInput4E25B0 struct {
	Name                     string
	Kind                     int
	Objects                  []radialObject4E25B0
	Points                   [2][2]uint32
	Center                   int
	Outer, Inner             uint32
	Damage, Typ, Flag        int32
	Source, Excluded, Target int
	Targets                  []int
	TraceResult              int32 `json:"trace_result"`
	Stage, Action            string
	At                       int
	NilContext               bool `json:"nil_context"`
}
type radialRow4E25B0 struct {
	Input  RadialInput4E25B0
	State  radialState4E25B0
	Events []radialEvent4E25B0
	Fault  bool
}

// RadialScene4E25B0 supplies the same declared services to the Go and C-entry
// tests. Every pointer passed to C is C-owned and its whole frame is checked.
type RadialScene4E25B0 struct {
	Input   RadialInput4E25B0
	Objects []*server.Object
	Points  []*types.Pointf
	Context *server.MapRadialDamageContext4E25B0
	Runtime server.MapRadialDamageRuntime4E25B0
	active  *server.MapRadialDamageContext4E25B0
	flag    uint32
	events  []radialEvent4E25B0
	calls   map[string]int
}

func radialPoint4E25B0(bits [2]uint32) types.Pointf {
	return types.Ptf(math.Float32frombits(bits[0]), math.Float32frombits(bits[1]))
}
func radialBits4E25B0(p types.Pointf) [2]uint32 {
	return [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)}
}
func (s *RadialScene4E25B0) Object(index int) *server.Object {
	if index < 0 {
		return nil
	}
	return s.Objects[index]
}
func (s *RadialScene4E25B0) Center(index int) *types.Pointf {
	if index < 0 {
		return nil
	}
	return s.Points[index]
}
func (s *RadialScene4E25B0) objectIndex(obj *server.Object) int {
	if obj == nil {
		return -1
	}
	for n, p := range s.Objects {
		if p == obj {
			return n
		}
	}
	panic("foreign object")
}
func (s *RadialScene4E25B0) centerIndex(point *types.Pointf) int {
	if point == nil {
		return -1
	}
	for n, p := range s.Points {
		if p == point {
			return n
		}
	}
	panic("foreign center")
}
func (s *RadialScene4E25B0) state() radialState4E25B0 {
	v := radialState4E25B0{Flag: s.flag}
	for _, p := range s.Objects {
		v.Objects = append(v.Objects, radialObject4E25B0{Pos: radialBits4E25B0(p.PosVec), Flags: uint32(p.ObjFlags), Callback: p.Damage != nil})
	}
	for n, p := range s.Points {
		v.Points[n] = radialBits4E25B0(*p)
	}
	if c := s.active; c != nil {
		v.Context = &radialContext4E25B0{Center: s.centerIndex(c.Center), Outer: math.Float32bits(c.Outer), Inner: math.Float32bits(c.Inner), Damage: uint32(c.Damage), Typ: uint32(c.Type), Source: s.objectIndex(c.Source), Excluded: s.objectIndex(c.Excluded)}
	}
	return v
}
func (s *RadialScene4E25B0) event(name string, args ...int64) {
	s.events = append(s.events, radialEvent4E25B0{Service: name, Args: args, State: s.state()})
}
func (s *RadialScene4E25B0) mutate(stage string, rect *types.Rectf) {
	n := s.calls[stage]
	s.calls[stage]++
	if stage != s.Input.Stage || n != s.Input.At {
		return
	}
	switch s.Input.Action {
	case "center-values":
		*s.Points[0] = types.Ptf(.3, .7)
	case "target-position":
		s.Objects[1].PosVec = types.Ptf(-99.1, 17.3)
	case "switch-center":
		s.active.Center = s.Points[1]
	case "nil-center":
		s.active.Center = nil
	case "outer":
		s.active.Outer = 20.3
	case "inner":
		s.active.Inner = 200.7
	case "damage":
		s.active.Damage = -117
	case "typ":
		s.active.Type = 9
	case "source":
		s.active.Source = s.Objects[3]
	case "excluded":
		s.active.Excluded = s.Objects[1]
	case "enable":
		s.flag = 1
	case "disable":
		s.flag = 0
	case "nil-callback":
		s.Objects[1].Damage = nil
	case "rect":
		*rect = types.Rectf{Min: types.Ptf(47.9, -47.9), Max: types.Ptf(-23.9, 23.9)}
	default:
		panic("unknown declared mutation")
	}
}
func (s *RadialScene4E25B0) RunGo() {
	i := s.Input
	if i.Kind == 0 {
		server.MapRadialDamageRaw4E25B0(s.Center(i.Center), math.Float32frombits(i.Outer), math.Float32frombits(i.Inner), i.Damage, object.DamageType(i.Typ), s.Object(i.Source), s.Object(i.Excluded), s.Runtime)
		return
	}
	c := s.Context
	if i.NilContext {
		c = nil
	}
	server.MapRadialDamageTargetRaw4E2700(s.Object(i.Target), c, s.Runtime)
}

func RunRadialOriginal4E25B0(t *testing.T, path string, rootsOnly bool, invoke func(*RadialScene4E25B0)) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows        []radialRow4E25B0
		ControlWord uint32 `json:"control_word"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 497 || fixture.ControlWord != 0xe3f {
		t.Fatal("original fixture count/control word")
	}
	count := 0
	for _, row := range fixture.Rows {
		if rootsOnly && row.Input.Kind != 0 {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			s := &RadialScene4E25B0{Input: row.Input, flag: uint32(row.Input.Flag), calls: map[string]int{}, events: []radialEvent4E25B0{}}
			frame := func(size uintptr) []byte {
				v, free := alloc.Make([]byte(nil), int(size)+64)
				t.Cleanup(free)
				for n := range v {
					v[n] = 0xa5
				}
				return v
			}
			var of, pf, ow, pw [][]byte
			for _, o := range row.Input.Objects {
				f := frame(unsafe.Sizeof(server.Object{}))
				p := (*server.Object)(unsafe.Pointer(&f[32]))
				p.PosVec = radialPoint4E25B0(o.Pos)
				p.ObjFlags = object.Flags(o.Flags)
				p.Damage = nil
				if o.Callback {
					p.Damage = unsafe.Pointer(p)
				}
				s.Objects = append(s.Objects, p)
				of = append(of, f)
				ow = append(ow, bytes.Clone(f))
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= 0xffffffff {
					t.Fatal("C-owned object below4GiB")
				}
			}
			for _, p := range row.Input.Points {
				f := frame(8)
				point := (*types.Pointf)(unsafe.Pointer(&f[32]))
				*point = radialPoint4E25B0(p)
				s.Points = append(s.Points, point)
				pf = append(pf, f)
				pw = append(pw, bytes.Clone(f))
			}
			cf := frame(unsafe.Sizeof(server.MapRadialDamageContext4E25B0{}))
			s.Context = (*server.MapRadialDamageContext4E25B0)(unsafe.Pointer(&cf[32]))
			i := row.Input
			*s.Context = server.MapRadialDamageContext4E25B0{Center: s.Center(i.Center), Outer: math.Float32frombits(i.Outer), Inner: math.Float32frombits(i.Inner), Damage: i.Damage, Type: i.Typ, Source: s.Object(i.Source), Excluded: s.Object(i.Excluded)}
			cw := bytes.Clone(cf)
			if i.Kind != 0 && !i.NilContext {
				s.active = s.Context
			}
			s.Runtime = server.MapRadialDamageRuntime4E25B0{
				DamageSource: func() bool { return s.flag != 0 }, ResetDamageSource: func() { s.flag = 1 },
				EachInRect: func(rect *types.Rectf, c *server.MapRadialDamageContext4E25B0, apply func(*server.Object, *server.MapRadialDamageContext4E25B0)) {
					s.active = c
					b, e := radialBits4E25B0(rect.Min), radialBits4E25B0(rect.Max)
					s.event("each", int64(b[0]), int64(b[1]), int64(e[0]), int64(e[1]))
					s.mutate("each", rect)
					for _, n := range i.Targets {
						apply(s.Object(n), c)
					}
				},
				Trace: func(ray *types.Rectf) int32 {
					b, e := radialBits4E25B0(ray.Min), radialBits4E25B0(ray.Max)
					s.event("trace", int64(b[0]), int64(b[1]), int64(e[0]), int64(e[1]))
					s.mutate("trace", nil)
					return i.TraceResult
				},
				Damage: func(target, source *server.Object, d int32, typ object.DamageType) {
					s.event("damage", int64(s.objectIndex(target)), int64(s.objectIndex(source)), -1, int64(uint32(d)), int64(uint32(typ)))
					s.mutate("damage", nil)
				},
				DamageWalls: func(rect image.Rectangle, center *types.Pointf, radius float32, d int32, typ object.DamageType, source *server.Object) {
					s.active = nil
					p := radialBits4E25B0(*center)
					s.event("walls", int64(uint32(rect.Min.X)), int64(uint32(rect.Min.Y)), int64(uint32(rect.Max.X)), int64(uint32(rect.Max.Y)), int64(p[0]), int64(p[1]), int64(math.Float32bits(radius)), int64(uint32(d)), int64(uint32(typ)), int64(s.objectIndex(source)))
					s.mutate("walls", nil)
				},
			}
			fault := false
			func() {
				defer func() {
					if recover() != nil {
						fault = true
					}
				}()
				invoke(s)
			}()
			if i.Kind == 0 {
				s.active = nil
			}
			got := s.state()
			if fault != row.Fault || !reflect.DeepEqual(got, row.State) || !reflect.DeepEqual(s.events, row.Events) {
				t.Errorf("fault/state/events differ: fault=%t/%t state=%+v want=%+v events=%+v want=%+v", fault, row.Fault, got, row.State, s.events, row.Events)
			}
			point := func(b []byte, offset uintptr, p types.Pointf) {
				binary.LittleEndian.PutUint32(b[32+offset:], math.Float32bits(p.X))
				binary.LittleEndian.PutUint32(b[36+offset:], math.Float32bits(p.Y))
			}
			ptr := func(b []byte, offset uintptr, p unsafe.Pointer) {
				if unsafe.Sizeof(p) == 8 {
					binary.LittleEndian.PutUint64(b[32+offset:], uint64(uintptr(p)))
				} else {
					binary.LittleEndian.PutUint32(b[32+offset:], uint32(uintptr(p)))
				}
			}
			for n, o := range row.State.Objects {
				point(ow[n], unsafe.Offsetof(s.Objects[n].PosVec), radialPoint4E25B0(o.Pos))
				binary.LittleEndian.PutUint32(ow[n][32+unsafe.Offsetof(s.Objects[n].ObjFlags):], o.Flags)
				var p unsafe.Pointer
				if o.Callback {
					p = unsafe.Pointer(s.Objects[n])
				}
				ptr(ow[n], unsafe.Offsetof(s.Objects[n].Damage), p)
				if !bytes.Equal(of[n], ow[n]) {
					t.Error("complete object frame/canary changed")
				}
			}
			for n, p := range row.State.Points {
				point(pw[n], 0, radialPoint4E25B0(p))
				if !bytes.Equal(pf[n], pw[n]) {
					t.Error("complete center frame/canary changed")
				}
			}
			if c := row.State.Context; c != nil {
				ptr(cw, unsafe.Offsetof(s.Context.Center), unsafe.Pointer(s.Center(c.Center)))
				ptr(cw, unsafe.Offsetof(s.Context.Source), unsafe.Pointer(s.Object(c.Source)))
				ptr(cw, unsafe.Offsetof(s.Context.Excluded), unsafe.Pointer(s.Object(c.Excluded)))
				for _, f := range []struct {
					off uintptr
					v   uint32
				}{{unsafe.Offsetof(s.Context.Outer), c.Outer}, {unsafe.Offsetof(s.Context.Inner), c.Inner}, {unsafe.Offsetof(s.Context.Damage), c.Damage}, {unsafe.Offsetof(s.Context.Type), c.Typ}} {
					binary.LittleEndian.PutUint32(cw[32+f.off:], f.v)
				}
			}
			if !bytes.Equal(cf, cw) {
				t.Error("complete context frame/canary changed")
			}
		})
	}
	if rootsOnly && count != 257 {
		t.Fatal("native C root context count")
	}
}
