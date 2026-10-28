package server

import (
	"encoding/json"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"
)

type sgObject struct {
	Flags  uint32    `json:"flags"`
	Class  uint32    `json:"cls"`
	Pos    [2]uint32 `json:"pos"`
	End    [2]uint32 `json:"end"`
	Radius uint32    `json:"radius"`
	Update int       `json:"update_ref"`
	Next   int       `json:"next"`
	Prev   int       `json:"prev"`
	Health bool      `json:"health"`
}
type sgState struct {
	Objects       [4]sgObject  `json:"objects"`
	Data          [2][3]uint32 `json:"data"`
	Head          int          `json:"head"`
	Player        [4]uint32    `json:"player"`
	Ray           [4]uint32    `json:"ray"`
	ContextSource int          `json:"context_source"`
	ContextRay    int          `json:"context_ray"`
}
type sgInput struct {
	Name           string    `json:"name"`
	Kind           int       `json:"kind"`
	Source         int       `json:"source_ref"`
	ContextPresent bool      `json:"context_present"`
	PlayerPresent  bool      `json:"player_present"`
	Index          int32     `json:"index"`
	Mutation       string    `json:"mutation"`
	Action         string    `json:"action"`
	At             int       `json:"at"`
	WriteRay       [4]uint32 `json:"write_ray"`
	TraceResult    int32     `json:"trace_result"`
	Hit            [2]uint32 `json:"hit"`
	Candidates     []int     `json:"candidates"`
	GameFlags      uint32    `json:"game_flags"`
	RealProjection bool      `json:"real_projection"`
	ProjectResult  int32     `json:"project_result"`
	Projected      [2]uint32 `json:"projected"`
	Parent         int       `json:"parent"`
	PacketResult   int32     `json:"packet_result"`
	sgState
}
type sgEvent struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	sgState
}
type sgRow struct {
	Input     sgInput   `json:"input"`
	Output    sgState   `json:"output"`
	Events    []sgEvent `json:"events"`
	Result    int       `json:"result"`
	Fault     bool      `json:"fault"`
	Preserved bool      `json:"nonoutput_bytes_preserved"`
}

func sgRun(t *testing.T, in sgInput, want sgState) (sgState, []sgEvent, int, any) {
	t.Helper()
	var objs [4]*Object
	var data [2]*[3]uint32
	var player *Player
	var context *SentryGlobeRayContext511020
	var ray *types.Rectf
	var health *HealthData
	for n := range objs {
		objs[n] = new(Object)
	}
	for n := range data {
		data[n] = new([3]uint32)
	}
	player = new(Player)
	context = new(SentryGlobeRayContext511020)
	ray = new(types.Rectf)
	health = new(HealthData)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(objs[0])) <= math.MaxUint32 {
		t.Fatal("not high native objects")
	}

	ref := func(n int) *Object {
		if n < 0 {
			return nil
		}
		return objs[n]
	}
	id := func(p *Object) int {
		if p == nil {
			return -1
		}
		for n, o := range objs {
			if p == o {
				return n
			}
		}
		t.Fatal("unknown object")
		return -2
	}
	dataref := func(n int) unsafe.Pointer {
		if n < 0 {
			return nil
		}
		return unsafe.Pointer(data[n])
	}
	dataID := func(p unsafe.Pointer) int {
		if p == nil {
			return -1
		}
		for n, d := range data {
			if p == unsafe.Pointer(d) {
				return n
			}
		}
		t.Fatal("unknown update")
		return -2
	}
	for n, v := range in.Objects {
		o := objs[n]
		*o = Object{ObjFlags: object.Flags(v.Flags), ObjClass: object.Class(v.Class), PosVec: types.Ptf(math.Float32frombits(v.Pos[0]), math.Float32frombits(v.Pos[1])), Pos39: types.Ptf(math.Float32frombits(v.End[0]), math.Float32frombits(v.End[1])), UpdateData: dataref(v.Update), InvNextItem: ref(v.Next), Field125: ref(v.Prev)}
		o.Shape.Circle.R = math.Float32frombits(v.Radius)
		if v.Health {
			o.HealthData = health
		}
	}
	for n, d := range data {
		*d = in.Data[n]
	}
	*player = Player{Field10: uint16(in.Player[2]), Field12: uint16(in.Player[3]), Pos3632Vec: types.Ptf(math.Float32frombits(in.Player[0]), math.Float32frombits(in.Player[1]))}
	*ray = types.Rectf{Min: types.Ptf(math.Float32frombits(in.Ray[0]), math.Float32frombits(in.Ray[1])), Max: types.Ptf(math.Float32frombits(in.Ray[2]), math.Float32frombits(in.Ray[3]))}
	*context = SentryGlobeRayContext511020{Source: ref(in.ContextSource)}
	if in.ContextRay >= 0 {
		context.Ray = ray
	}
	head := ref(in.Head)

	pointWords := func(p types.Pointf) []int64 {
		return []int64{int64(math.Float32bits(p.X)), int64(math.Float32bits(p.Y))}
	}
	rayWords := func(p *types.Rectf) []int64 {
		if p == nil {
			return []int64{-1, -1, -1, -1}
		}
		return append(pointWords(p.Min), pointWords(p.Max)...)
	}
	snapshot := func() sgState {
		s := sgState{Head: id(head), Player: [4]uint32{math.Float32bits(player.Pos3632Vec.X), math.Float32bits(player.Pos3632Vec.Y), uint32(player.Field10), uint32(player.Field12)}, ContextSource: id(context.Source), ContextRay: -1, Ray: [4]uint32{math.Float32bits(ray.Min.X), math.Float32bits(ray.Min.Y), math.Float32bits(ray.Max.X), math.Float32bits(ray.Max.Y)}}
		if context.Ray != nil {
			s.ContextRay = 0
		}
		for n, o := range objs {
			s.Objects[n] = sgObject{Flags: uint32(o.ObjFlags), Class: uint32(o.ObjClass), Pos: [2]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y)}, End: [2]uint32{math.Float32bits(o.Pos39.X), math.Float32bits(o.Pos39.Y)}, Radius: math.Float32bits(o.Shape.Circle.R), Update: dataID(o.UpdateData), Next: id(o.InvNextItem), Prev: id(o.Field125), Health: o.HealthData != nil}
		}
		for n, d := range data {
			s.Data[n] = *d
		}
		return s
	}
	events := make([]sgEvent, 0)
	event := func(name string, args ...int64) {
		events = append(events, sgEvent{Service: name, Args: args, sgState: snapshot()})
	}
	applied := false
	calls := map[string]int{}
	mutate := func(service string, c *SentryGlobeRayContext511020, rp *types.Rectf) {
		number := calls[service]
		calls[service]++
		if applied || in.Mutation != service+":"+in.Action || in.At != number {
			return
		}
		applied = true
		a := objs[0]
		switch in.Action {
		case "position":
			a.PosVec = types.Ptf(100.1, -200.3)
		case "angle":
			data[0][0] = math.Float32bits(0.7)
			data[0][2] = math.Float32bits(-0.3)
		case "update":
			a.UpdateData = dataref(1)
		case "flags":
			objs[1].ObjFlags = 0x10000
		case "candidate-position":
			objs[1].PosVec = types.Ptf(111.1, 52.7)
		case "radius":
			objs[1].Shape.Circle.R = 99.1
		case "source":
			c.Source = objs[2]
		case "ray":
			*rp = types.Rectf{Min: types.Ptf(math.Float32frombits(in.WriteRay[0]), math.Float32frombits(in.WriteRay[1])), Max: types.Ptf(math.Float32frombits(in.WriteRay[2]), math.Float32frombits(in.WriteRay[3]))}
		case "next":
			a.InvNextItem = objs[2]
		case "enabled":
			a.ObjFlags ^= 0x1000000
		default:
			t.Fatal("unknown mutation")
		}
	}
	var h SentryGlobeRawRuntime510E60
	h = SentryGlobeRawRuntime510E60{
		List: SentryGlobeListRuntime510E60{Head: func() *Object { return head }, SetHead: func(o *Object) { head = o }},
		TraceRay: func(rp *types.Rectf, hp *types.Pointf, flags uint32) int32 {
			if flags != 5 {
				t.Fatal("trace flags")
			}
			event("trace", append(rayWords(rp), 0, 5)...)
			mutate("trace", context, rp)
			*hp = types.Ptf(math.Float32frombits(in.Hit[0]), math.Float32frombits(in.Hit[1]))
			return in.TraceResult
		},
		EachInRect: func(bounds *types.Rectf, c *SentryGlobeRayContext511020) {
			event("each", append(append(rayWords(bounds), int64(id(c.Source))), rayWords(c.Ray)...)...)
			mutate("each", c, c.Ray)
			for _, n := range in.Candidates {
				SentryGlobeHitRaw511020(ref(n), c, h)
			}
		},
		GameFlags: func(mask uint32) uint32 {
			if mask != 4096 {
				t.Fatal("Quest mask")
			}
			event("game", 4096)
			mutate("game", context, ray)
			return in.GameFlags
		},
		PointOnLine: func(rp *types.Rectf, pp, out *types.Pointf) int32 {
			owner := -1
			for n, o := range objs {
				if pp == &o.PosVec {
					owner = n
				}
			}
			if owner < 0 {
				t.Fatal("live candidate position pointer")
			}
			event("project", append(append(rayWords(rp), int64(owner)), pointWords(*pp)...)...)
			mutate("project", context, rp)
			if in.RealProjection {
				value, ok := PointOnTheLine(rp.Min, rp.Max, *pp)
				*out = value
				if ok {
					return 1
				}
				return 0
			}
			*out = types.Ptf(math.Float32frombits(in.Projected[0]), math.Float32frombits(in.Projected[1]))
			return in.ProjectResult
		},
		FindParent: func(o *Object) *Object {
			event("parent", int64(id(o)))
			mutate("parent", context, ray)
			return ref(in.Parent)
		},
		Damage: func(target, parent, source *Object, damage int, typ object.DamageType) {
			event("damage", int64(id(target)), int64(id(parent)), int64(id(source)), int64(damage), int64(typ))
			mutate("damage", context, ray)
		},
		Audio: func(sound uint32, o *Object) {
			event("audio", int64(sound), int64(id(o)), 0, 0)
			mutate("audio", context, ray)
		},
		Player: func(index int32) *Player {
			event("player", int64(uint32(index)))
			mutate("player", context, ray)
			if !in.PlayerPresent {
				return nil
			}
			return player
		},
		Packet: func(index, kind int32, packet []byte) int32 {
			values := []int64{int64(uint32(index)), int64(kind)}
			for _, b := range packet {
				values = append(values, int64(b))
			}
			event("packet", values...)
			mutate("packet", context, ray)
			return in.PacketResult
		},
	}
	result := 0
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		ctx := context
		if !in.ContextPresent {
			ctx = nil
		}
		switch in.Kind {
		case 0:
			SentryGlobeUpdateRaw510E60(ref(in.Source), h)
		case 1:
			result = id(SentryGlobeUnlinkRaw510FD0(ref(in.Source), h.List))
		case 2:
			SentryGlobeHitRaw511020(ref(in.Source), ctx, h)
		case 3:
			SentryGlobeSendRaw511100(in.Index, h)
		case 4:
			result = int(SentryGlobePacketRaw511250(in.Index, ref(in.Source), h.Packet))
		case 5:
			SentryGlobeResetRaw510E50(h.List)
		}
	}()

	return snapshot(), events, result, recovered
}
func TestSentryGlobeOriginalGo(t *testing.T) {
	blob, err := os.ReadFile("testdata/sentry_globe_profiled.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA      string  `json:"original_sha256"`
		CW       uint32  `json:"original_startup_FPCW"`
		Contexts int     `json:"contexts"`
		Rows     []sgRow `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.CW != 0xe3f || f.Contexts != 630 || len(f.Rows) != 630 {
		t.Fatal("original provenance")
	}
	var report *os.File
	if path := os.Getenv("OPENNOX_SENTRY_REPORT"); path != "" {
		report, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer report.Close()
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, result, fault := sgRun(t, row.Input, row.Output)
			if report != nil {
				if err := json.NewEncoder(report).Encode(map[string]any{"name": row.Input.Name, "output": out, "events": events, "result": result, "state_matches": reflect.DeepEqual(out, row.Output), "events_match": reflect.DeepEqual(events, row.Events), "fault": fault != nil}); err != nil {
					t.Fatal(err)
				}
			}
			if !row.Preserved || fault != nil {
				t.Fatal("fault/canary", fault)
			}
			if !reflect.DeepEqual(out, row.Output) || result != row.Result {
				t.Errorf("state/result differs: got %+v/%d want %+v/%d", out, result, row.Output, row.Result)
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("ordered services/arguments/state differ: got %+v want %+v", events, row.Events)
			}
		})
	}
}
func TestSentryGlobeOriginalFaultsGo(t *testing.T) {
	blob, err := os.ReadFile("testdata/sentry_globe_faults_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Rows []sgRow `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Rows) != 10 {
		t.Fatal("fault provenance")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, result, fault := sgRun(t, row.Input, row.Output)
			if (fault != nil) != row.Fault || !reflect.DeepEqual(out, row.Output) || !reflect.DeepEqual(events, row.Events) {
				t.Fatal("fault/gate/prefix/state differs", fault)
			}
			if !row.Fault && result != row.Result {
				t.Fatal("safe result")
			}
		})
	}
}
