package server

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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
)

type circleInputOriginal struct {
	Kind              string    `json:"kind"`
	Name              string    `json:"name"`
	FirstNew          [2]uint32 `json:"first_new"`
	SecondNew         [2]uint32 `json:"second_new"`
	FirstPosition     [2]uint32 `json:"first_position"`
	SecondPosition    [2]uint32 `json:"second_position"`
	FirstVelocity     [2]uint32 `json:"first_velocity"`
	SecondVelocity    [2]uint32 `json:"second_velocity"`
	FirstAccumulator  [2]uint32 `json:"first_accumulator"`
	SecondAccumulator [2]uint32 `json:"second_accumulator"`
	FirstMass         uint32    `json:"first_mass"`
	SecondMass        uint32    `json:"second_mass"`
	FirstRadius       uint32    `json:"first_radius"`
	SecondRadius      uint32    `json:"second_radius"`
	FirstClass        uint32    `json:"first_class"`
	SecondClass       uint32    `json:"second_class"`
	FirstFlags        uint32    `json:"first_flags"`
	SecondFlags       uint32    `json:"second_flags"`
	FirstType         uint16    `json:"first_type"`
	BoxSize           [2]uint32 `json:"box_size"`
	BoxOffsets        [8]uint32 `json:"box_offsets"`
	Move              int32     `json:"move"`
	Random            int       `json:"random"`
	TraceClear        bool      `json:"trace_clear"`
	Push              uint32    `json:"push"`
	Spring            uint32    `json:"spring"`
	WallType          byte      `json:"wall_type"`
	Grid              [2]int    `json:"grid"`
	AllWalls          bool      `json:"all_walls"`
	Collide           [4]uint32 `json:"collide"`
	Update            bool      `json:"update"`
	BallCache         uint32    `json:"ball_cache"`
	Origin            [2]uint32 `json:"origin"`
	Point             [2]uint32 `json:"point"`
	Lo                uint32    `json:"lo"`
	Hi                uint32    `json:"hi"`
	AllowLo           int32     `json:"allow_lo"`
	AllowHi           int32     `json:"allow_hi"`
	Mutation          int       `json:"mutation"`
}

type circleStateOriginal struct {
	New         [2]uint32 `json:"new"`
	Position    [2]uint32 `json:"position"`
	Velocity    [2]uint32 `json:"velocity"`
	Accumulator [2]uint32 `json:"accumulator"`
	Flags       uint32    `json:"flags"`
	Class       uint32    `json:"class_bits"`
	Mass        uint32    `json:"mass"`
	Radius      uint32    `json:"radius"`
	Type        uint16    `json:"type_index"`
}

type circleEventOriginal struct {
	Service                string              `json:"service"`
	Object                 string              `json:"object"`
	Other                  string              `json:"other"`
	Name                   string              `json:"name"`
	Ray                    []uint32            `json:"ray"`
	Normal                 []uint32            `json:"normal"`
	Grid                   []int               `json:"grid"`
	Flags                  uint32              `json:"flags"`
	Value                  int                 `json:"value"`
	First                  circleStateOriginal `json:"first"`
	Second                 circleStateOriginal `json:"second"`
	CollisionWall          string              `json:"collision_wall"`
	AlternateCollisionWall string              `json:"alternate_collision_wall"`
}

type circleRowOriginal struct {
	Mode                   string                `json:"mode"`
	Input                  circleInputOriginal   `json:"input"`
	First                  circleStateOriginal   `json:"first"`
	Second                 circleStateOriginal   `json:"second"`
	Events                 []circleEventOriginal `json:"events"`
	Result                 uint64                `json:"result_bits64"`
	Normal                 [2]uint32             `json:"normal"`
	Return                 int                   `json:"return_int"`
	CollisionWall          string                `json:"collision_wall"`
	AlternateCollisionWall string                `json:"alternate_collision_wall"`
	Update                 string                `json:"update"`
	Push                   uint32                `json:"push"`
	Spring                 uint32                `json:"spring"`
	Preserved              bool                  `json:"nonoutput_bytes_preserved"`
}

func TestCircleCollisionStartupOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/circle_collision_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"actual_startup_control_word"`
		Contexts  int               `json:"contexts"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Tables    []struct {
			Name string `json:"name"`
			Hex  string `json:"hex"`
		} `json:"tables"`
		Rows []circleRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || fixture.Contexts != 966 || len(fixture.Rows) != 1932 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete unchanged original circle observations")
	}
	tables := make(map[string][]byte)
	for _, table := range fixture.Tables {
		data, err := hex.DecodeString(table.Hex)
		if err != nil {
			t.Fatal(err)
		}
		tables[table.Name] = data
	}
	if len(tables["wall-definitions"]) != 264 || len(tables["wall-edge-masks"]) != 11 || len(tables["box-random-directions"]) != 32 || len(tables["circle-random-directions"]) != 32 {
		t.Fatal("incomplete original tables")
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) { testCircleCollisionOriginal(t, row, tables) })
	}
	if selected != 966 {
		t.Fatalf("startup conditions=%d", selected)
	}
}

func testCircleCollisionOriginal(t *testing.T, row circleRowOriginal, tables map[string][]byte) {
	t.Helper()
	i := row.Input
	a, fa := alloc.New(Object{})
	b, fb := alloc.New(Object{})
	u, fu := alloc.New(PlayerUpdateData{})
	alt, falt := alloc.New(PlayerUpdateData{})
	wall, fw := alloc.New(Wall{})
	for _, free := range []func(){fa, fb, fu, falt, fw} {
		t.Cleanup(free)
	}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(a), unsafe.Pointer(b), unsafe.Pointer(u), unsafe.Pointer(alt), unsafe.Pointer(wall)} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("native circle fixture below 4 GiB")
		}
	}
	*a = Object{TypeInd: i.FirstType, ObjClass: object.Class(i.FirstClass), ObjFlags: object.Flags(i.FirstFlags), NewPos: rigidPointOriginal(i.FirstNew), PosVec: rigidPointOriginal(i.FirstPosition), VelVec: rigidPointOriginal(i.FirstVelocity), Pos24: rigidPointOriginal(i.FirstAccumulator), Mass: math.Float32frombits(i.FirstMass), Field29: 0x13572468}
	*b = Object{TypeInd: 102, ObjClass: object.Class(i.SecondClass), ObjFlags: object.Flags(i.SecondFlags), NewPos: rigidPointOriginal(i.SecondNew), PosVec: rigidPointOriginal(i.SecondPosition), VelVec: rigidPointOriginal(i.SecondVelocity), Pos24: rigidPointOriginal(i.SecondAccumulator), Mass: math.Float32frombits(i.SecondMass), Field29: 0x87654321}
	a.Shape.Kind, b.Shape.Kind = ShapeKindCircle, ShapeKindBox
	a.Shape.Circle.R, b.Shape.Circle.R = math.Float32frombits(i.FirstRadius), math.Float32frombits(i.SecondRadius)
	b.Shape.Box.W, b.Shape.Box.H = math.Float32frombits(i.BoxSize[0]), math.Float32frombits(i.BoxSize[1])
	fields := []*float32{&b.Shape.Box.LeftTop, &b.Shape.Box.LeftBottom, &b.Shape.Box.LeftBottom2, &b.Shape.Box.LeftTop2, &b.Shape.Box.RightTop, &b.Shape.Box.RightBottom, &b.Shape.Box.RightBottom2, &b.Shape.Box.RightTop2}
	for index, p := range fields {
		*p = math.Float32frombits(i.BoxOffsets[index])
	}
	a.CollideP1, a.CollideP2 = rigidPointOriginal([2]uint32{i.Collide[0], i.Collide[1]}), rigidPointOriginal([2]uint32{i.Collide[2], i.Collide[3]})
	rawU := func(p *PlayerUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))
	}
	for _, p := range []*PlayerUpdateData{u, alt} {
		for index := range rawU(p) {
			rawU(p)[index] = 0xa5
		}
		p.CollisionWall = nil
	}
	if i.Update {
		a.UpdateData = unsafe.Pointer(u)
	}
	*wall = Wall{X5: 15, Y6: 15, Dir0: i.WallType}
	label := func(p *Object) string {
		switch p {
		case a:
			return "first"
		case b:
			return "second"
		case nil:
			return "nil"
		}
		t.Fatal("unknown collision object")
		return ""
	}
	wallLabel := func(p *Wall) string {
		if p == wall {
			return "wall"
		}
		if p == nil {
			return "nil"
		}
		t.Fatal("unknown collision wall")
		return ""
	}
	updateLabel := func(p unsafe.Pointer) string {
		switch p {
		case unsafe.Pointer(u):
			return "old"
		case unsafe.Pointer(alt):
			return "alternate"
		case nil:
			return "nil"
		}
		t.Fatal("unknown update pointer")
		return ""
	}
	state := func(p *Object) circleStateOriginal {
		return circleStateOriginal{rigidBitsOriginal(p.NewPos), rigidBitsOriginal(p.PosVec), rigidBitsOriginal(p.VelVec), rigidBitsOriginal(p.Pos24), uint32(p.ObjFlags), uint32(p.ObjClass), math.Float32bits(p.Mass), math.Float32bits(p.Shape.Circle.R), p.TypeInd}
	}
	rawObj := func(p *Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	writePointer := func(raw []byte, offset uintptr, p unsafe.Pointer) {
		if unsafe.Sizeof(uintptr(0)) == 8 {
			binary.LittleEndian.PutUint64(raw[offset:], uint64(uintptr(p)))
		} else {
			binary.LittleEndian.PutUint32(raw[offset:], uint32(uintptr(p)))
		}
	}
	wantObj := func(p *Object, w circleStateOriginal, update unsafe.Pointer) []byte {
		raw := bytes.Clone(rawObj(p))
		base := uintptr(unsafe.Pointer(p))
		write := func(field unsafe.Pointer, value uint32) {
			binary.LittleEndian.PutUint32(raw[uintptr(field)-base:], value)
		}
		for _, field := range []struct {
			P *types.Pointf
			W [2]uint32
		}{{&p.NewPos, w.New}, {&p.PosVec, w.Position}, {&p.VelVec, w.Velocity}, {&p.Pos24, w.Accumulator}} {
			write(unsafe.Pointer(&field.P.X), field.W[0])
			write(unsafe.Pointer(&field.P.Y), field.W[1])
		}
		write(unsafe.Pointer(&p.ObjFlags), w.Flags)
		write(unsafe.Pointer(&p.ObjClass), w.Class)
		write(unsafe.Pointer(&p.Mass), w.Mass)
		write(unsafe.Pointer(&p.Shape.Circle.R), w.Radius)
		writePointer(raw, unsafe.Offsetof(p.UpdateData), update)
		return raw
	}
	var finalUpdate unsafe.Pointer
	switch row.Update {
	case "old":
		finalUpdate = unsafe.Pointer(u)
	case "alternate":
		finalUpdate = unsafe.Pointer(alt)
	case "nil":
	default:
		t.Fatal("unknown original update")
	}
	wantA, wantB := wantObj(a, row.First, finalUpdate), wantObj(b, row.Second, nil)
	wantU, wantAlt := bytes.Clone(rawU(u)), bytes.Clone(rawU(alt))
	for _, field := range []struct {
		Raw   []byte
		Label string
	}{{wantU, row.CollisionWall}, {wantAlt, row.AlternateCollisionWall}} {
		var p unsafe.Pointer
		if field.Label == "wall" {
			p = unsafe.Pointer(wall)
		}
		writePointer(field.Raw, unsafe.Offsetof(u.CollisionWall), p)
	}
	defs := bytes.Clone(tables["wall-definitions"])
	push, spring := math.Float32frombits(i.Push), math.Float32frombits(i.Spring)
	events := make([]circleEventOriginal, 0)
	observe := func(e circleEventOriginal) {
		e.First, e.Second = state(a), state(b)
		e.CollisionWall, e.AlternateCollisionWall = wallLabel(u.CollisionWall), wallLabel(alt.CollisionWall)
		events = append(events, e)
	}
	mutate := func(where string) {
		switch {
		case i.Mutation == 1 && where == "collision":
			a.Mass = 5.5
			b.Mass = 7.25
			a.VelVec = types.Ptf(-2.5, 3.75)
			b.VelVec = types.Ptf(1.25, -.75)
			push = 23.75
			spring = 77.25
		case i.Mutation == 2 && where == "trace":
			a.NewPos = types.Ptf(299.75, 302.25)
			a.Shape.Circle.R = 14
			a.ObjFlags = 8
		case i.Mutation == 3 && where == "wall":
			a.NewPos = types.Ptf(354.25, 348.5)
			a.VelVec = types.Ptf(2.5, -3.25)
		case i.Mutation == 4 && where == "collision":
			defs[int(i.WallType)*24+12] = 0
		case i.Mutation == 5 && where == "wall-object":
			a.UpdateData = unsafe.Pointer(alt)
		case i.Mutation == 6 && where == "wake":
			b.ObjFlags |= 0x08000000
		}
	}
	random := func(box bool) types.Pointf {
		observe(circleEventOriginal{Service: "random", Value: i.Random})
		table := tables["circle-random-directions"]
		if box {
			table = tables["box-random-directions"]
		}
		off := i.Random * 8
		return rigidPointOriginal([2]uint32{binary.LittleEndian.Uint32(table[off:]), binary.LittleEndian.Uint32(table[off+4:])})
	}
	collision := func(first, second *Object, normal types.Pointf) {
		w := rigidBitsOriginal(normal)
		observe(circleEventOriginal{Service: "collision", Object: label(first), Other: label(second), Normal: []uint32{w[0], w[1]}})
		mutate("collision")
	}
	r := CircleCollisionRuntime550D00{PushScale: func() float32 { return push }, RandomDirection: random, Trace: func(first, second types.Pointf) bool {
		x, y := rigidBitsOriginal(first), rigidBitsOriginal(second)
		observe(circleEventOriginal{Service: "trace", Ray: []uint32{x[0], x[1], y[0], y[1]}})
		mutate("trace")
		return i.TraceClear
	}, AddCollision: collision, Wake: func(p *Object) { observe(circleEventOriginal{Service: "wake", Object: label(p)}); mutate("wake") }}
	ballCache := i.BallCache
	w := CircleWallRuntime54FEF0{Spring: func() float32 { return spring }, Definition: func(index byte) CircleWallDefinition54FFC0 {
		d := defs[int(index)*24:][:24]
		f := func(off int) float32 { return math.Float32frombits(binary.LittleEndian.Uint32(d[off:])) }
		return CircleWallDefinition54FFC0{d[0], d[12], f(4), f(8), f(16), f(20)}
	}, EdgeMask: func(index byte) byte { return tables["wall-edge-masks"][index] }, WallType: func(pos image.Point, flags byte) int8 {
		observe(circleEventOriginal{Service: "wall", Grid: []int{pos.X, pos.Y}, Flags: uint32(flags)})
		mutate("wall")
		if i.AllWalls || pos == image.Pt(i.Grid[0], i.Grid[1]) {
			return int8(i.WallType)
		}
		return -1
	}, WallObject: func(pos image.Point) *Wall {
		observe(circleEventOriginal{Service: "wall-object", Grid: []int{pos.X, pos.Y}})
		mutate("wall-object")
		return wall
	}, IsGameBall: func(p *Object) bool {
		if ballCache == 0 {
			observe(circleEventOriginal{Service: "type-lookup", Name: "GameBall"})
			ballCache = 777
		}
		return uint32(p.TypeInd) == ballCache
	}, AddCollision: func(p *Object, normal types.Pointf) { collision(p, nil, normal) }, OpenSecret: func(pos image.Point, p *Object) {
		if p != a {
			t.Fatal("secret object changed")
		}
		observe(circleEventOriginal{Service: "secret", Grid: []int{pos.X, pos.Y}})
	}}
	normal := types.Ptf(.5, -.5)
	result := uint64(0)
	ret := 0
	switch i.Kind {
	case "circle-circle":
		CircleCircleCollision550D00(a, b, r)
	case "circle-box":
		CircleBoxCollision54AD50(a, b, i.Move, r)
	case "box-geometry":
		result = math.Float64bits(CircleBoxGeometry54A990(a.NewPos, a.Shape.Circle.R, b, &normal, func() types.Pointf { return random(true) }))
	case "circle-wall":
		CircleWallCollision54FEF0(a, w)
	case "wall-response":
		normal = rigidPointOriginal(i.Point)
		if CircleWallResponse550380(a, normal, w) {
			ret = 1
		}
	case "first-nearest", "second-nearest":
		if CircleWallNearest550280(i.Kind == "second-nearest", rigidPointOriginal(i.Origin), math.Float32frombits(i.Lo), math.Float32frombits(i.Hi), i.AllowLo, i.AllowHi, rigidPointOriginal(i.Point), &normal) {
			ret = 1
		}
	default:
		t.Fatal("unknown original entry", i.Kind)
	}
	if !row.Preserved || state(a) != row.First || state(b) != row.Second || result != row.Result || ret != row.Return || rigidBitsOriginal(normal) != row.Normal || !reflect.DeepEqual(events, row.Events) || !bytes.Equal(rawObj(a), wantA) || !bytes.Equal(rawObj(b), wantB) || !bytes.Equal(rawU(u), wantU) || !bytes.Equal(rawU(alt), wantAlt) || math.Float32bits(push) != row.Push || math.Float32bits(spring) != row.Spring {
		t.Fatalf("native result=%016x/%d normal=%08x first=%+v second=%+v old/alternate/update=%s/%s/%s bytes=%t/%t/%t/%t; original=%016x/%d normal=%08x first=%+v second=%+v; events=%+v original=%+v", result, ret, rigidBitsOriginal(normal), state(a), state(b), wallLabel(u.CollisionWall), wallLabel(alt.CollisionWall), updateLabel(a.UpdateData), bytes.Equal(rawObj(a), wantA), bytes.Equal(rawObj(b), wantB), bytes.Equal(rawU(u), wantU), bytes.Equal(rawU(alt), wantAlt), row.Result, row.Return, row.Normal, row.First, row.Second, events, row.Events)
	}
}
