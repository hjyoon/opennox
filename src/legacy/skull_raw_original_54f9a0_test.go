package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type skullRawContract54F9A0 interface {
	SkullRawUpdate54F9A0(*server.SkullRawState54F9A0, server.SkullRawRuntime54F9A0) uint32
	SkullRawProjectile54FA80(uint32, *server.SkullRawState54F9A0, server.SkullRawRuntime54F9A0)
	SkullRawState54FBB0(*server.SkullRawState54F9A0, server.SkullRawRuntime54F9A0)
	SkullRawScan54FBF0(*server.SkullRawState54F9A0, server.SkullRawRuntime54F9A0) uint32
	SkullRawReachable54FC50(*server.Object, *server.SkullRawState54F9A0, server.SkullRawRuntime54F9A0)
}

type skullRawInvoke54F9A0 struct {
	Entry          string
	Unit           int
	Target         int
	ProjectileType uint32 `json:"projectile_type"`
}

type skullRawPlan54F9A0 struct {
	Words, Words16, Bytes, Pointers, Globals [][]int64
	FPS                                      *uint32
	Result                                   *int64
	Targets                                  []int
	Reenter                                  json.RawMessage
}

type skullRawInput54F9A0 struct {
	Name, Entry                 string
	Unit, Target                int
	Projectile, Update, Collide int
	Flags                       uint32
	SourceType                  uint16 `json:"source_type"`
	Direction                   uint16
	Position                    [2]uint32
	RadiusBits                  uint32 `json:"radius_bits"`
	SpeedBits                   uint32 `json:"speed_bits"`
	Timers                      [2]uint32
	State, Enabled              uint8
	ProjectileType              uint32 `json:"projectile_type"`
	Targets                     []int
	TargetClass                 uint32 `json:"target_class"`
	TargetFlags                 uint32 `json:"target_flags"`
	Enemy                       uint32
	MapCheck                    uint32 `json:"map_check"`
	InFront                     uint32 `json:"in_front"`
	Caches                      [6]uint32
	FPS                         uint32
	NameResults                 map[string]uint32 `json:"name_results"`
	DamageBits                  uint32            `json:"damage_bits"`
	Script                      map[string]skullRawPlan54F9A0
}

type skullRawOutput54F9A0 struct {
	Name   string           `json:"name"`
	Result *uint32          `json:"result"`
	Fault  map[string]any   `json:"fault"`
	Events []map[string]any `json:"events"`
	Caches [6]uint32        `json:"caches"`
	FPS    uint32           `json:"fps"`
	Heap   string           `json:"heap_sha256"`
}

func TestSkullRawOriginal54F9A0(t *testing.T) {
	TestSkullRawNativeLayout54F9A0(t)
	var inputs []skullRawInput54F9A0
	trollDeathFixture54A270(t, "skull_raw_independent_inputs_54f9a0.json.gz", "2fa912b92071a20903b0596adce24a80942874efeb6e80fe5b8d8f6def6dcc36", "8a004ff539f042e37bfa144916780ab9042d5a6ed48889229bb6d6e18bce8825", &inputs)
	if len(inputs) != 753 || inputs[0].SourceType != 85 || inputs[0].DamageBits != 0x4236abcd {
		t.Fatal("complete independent input count and decoded scalar transport")
	}
	if _, ok := any((*server.Object)(nil)).(skullRawContract54F9A0); !ok {
		if os.Getenv("OPENNOX_SKULL_RAW_REQUIRED") == "1" {
			t.Fatal("five raw native Skull methods required")
		}
		t.Log("frozen complete original consumer compiled before native engine; no native leaf credit")
		return
	}
	for _, profile := range []struct {
		name, encoded, decoded string
		nearest                bool
	}{
		{"tozero", "c5033d70e53a39874dd667354eb9ead31d970e1ca63ab73b3603b6334dedde93", "f9269f877bc70d69e4defb25eecb5c200ad742df9d468d469fa152c4e06c288b", false},
		{"nearest", "08d9557293ef26581f612b6177b1df60f0e579addb9db47cf0c5b0e755d658de", "a7515ca2c9beb823f28871bf1c3256db6cd9a625c8a7078f7b4276110dccf76b", true},
	} {
		t.Run(profile.name, func(t *testing.T) {
			var expected []skullRawOutput54F9A0
			trollDeathFixture54A270(t, "skull_raw_native_original_54f9a0_pc53_"+profile.name+".json.gz", profile.encoded, profile.decoded, &expected)
			if len(expected) != len(inputs) {
				t.Fatal("complete original output domain")
			}
			for _, mode := range []string{"Go", "C"} {
				t.Run(mode, func(t *testing.T) {
					for i, in := range inputs {
						t.Run(in.Name, func(t *testing.T) {
							if expected[i].Name != in.Name {
								t.Fatal("immutable input/output identity")
							}
							got := skullRawRunOriginal54F9A0(t, in, mode, profile.nearest)
							a, e := json.Marshal(got)
							if e != nil {
								t.Fatal(e)
							}
							b, e := json.Marshal(expected[i])
							if e != nil {
								t.Fatal(e)
							}
							if !bytes.Equal(a, b) {
								t.Fatalf("complete original output differs\ngot %s\nwant%s", a, b)
							}
						})
					}
				})
			}
		})
	}
}

// Each service observes the whole declared PE heap, six live original dwords
// and FPS on entry, reentry, every candidate and return. The expected outputs
// are consulted only after this execution returns.
func skullRawRunOriginal54F9A0(t *testing.T, in skullRawInput54F9A0, mode string, nearest bool) (out skullRawOutput54F9A0) {
	m, free := newSkullRawMemory54F9A0(t, in)
	defer free()
	type guardedState struct {
		Before [16]byte
		State  server.SkullRawState54F9A0
		After  [16]byte
	}
	g, freeState := alloc.NewHigh(guardedState{})
	defer freeState()
	*g = guardedState{Before: [16]byte{0x13, 0x57, 0x9b}, After: [16]byte{0x24, 0x68, 0xac}}
	stateWords := unsafe.Slice((*uint32)(unsafe.Pointer(&g.State)), 6)
	copy(stateWords, in.Caches[:])
	guards := *g
	caches := func() (result [6]uint32) { copy(result[:], stateWords); return }
	fps := in.FPS
	events := []map[string]any{}
	occurrences := map[string]int{}
	nextIndex := 0
	type invocation struct {
		skullRawInvoke54F9A0
		phase                    string
		data, projectile, damage unsafe.Pointer
	}
	var active, failed *invocation
	var invoke func(skullRawInvoke54F9A0) *uint32
	frame := func(kind string, index int) map[string]any {
		return map[string]any{"kind": kind, "index": index, "caches": caches(), "fps": fps, "heap_sha256": m.checkpoint()}
	}
	begin := func(kind string) (map[string]any, skullRawPlan54F9A0) {
		n := occurrences[kind]
		occurrences[kind]++
		index := nextIndex
		nextIndex++
		return frame(kind, index), in.Script[fmt.Sprintf("%s:%d", kind, n)]
	}
	mutate := func(plan skullRawPlan54F9A0) {
		for _, field := range []struct {
			values [][]int64
			width  int
		}{{plan.Words, 4}, {plan.Words16, 2}, {plan.Bytes, 1}} {
			for _, v := range field.values {
				if len(v) != 3 {
					t.Fatal("mutation tuple")
				}
				m.setWord(int(v[0]), int(v[1]), field.width, uint64(v[2]))
			}
		}
		for _, v := range plan.Pointers {
			if len(v) != 3 {
				t.Fatal("pointer mutation tuple")
			}
			m.setPointer(int(v[0]), int(v[1]), int(v[2]))
		}
		for _, v := range plan.Globals {
			if len(v) != 2 || v[0] < 0 || v[0] >= 6 {
				t.Fatal("live dword mutation")
			}
			stateWords[v[0]] = uint32(v[1])
		}
		if plan.FPS != nil {
			fps = *plan.FPS
		}
		m.acceptWorldWrites()
	}
	finish := func(event map[string]any, plan skullRawPlan54F9A0, result any, actions func()) {
		events = append(events, event)
		mutate(plan)
		if len(plan.Reenter) != 0 {
			call := skullRawInvoke54F9A0{Unit: in.Unit, Target: in.Target, ProjectileType: in.ProjectileType}
			if json.Unmarshal(plan.Reenter, &call) != nil {
				t.Fatal("declared original reentry")
			}
			var arguments any
			if json.Unmarshal(plan.Reenter, &arguments) != nil {
				t.Fatal("reentry transport")
			}
			e := frame("nested", event["index"].(int))
			e["entry"], e["arguments"] = call.Entry, arguments
			events = append(events, e)
			r := invoke(call)
			e = frame("nested-return", event["index"].(int))
			e["result"] = r
			events = append(events, e)
		}
		if actions != nil {
			actions()
		}
		e := frame(event["kind"].(string)+"-return", event["index"].(int))
		e["result"] = result
		events = append(events, e)
	}
	result32 := func(plan skullRawPlan54F9A0, value uint32) uint32 {
		if plan.Result != nil {
			return uint32(*plan.Result)
		}
		return value
	}
	rt := server.SkullRawRuntime54F9A0{RoundToNearest: nearest, FPS: func() uint32 { return fps }}
	rt.Resolve = func(name string) uint32 {
		e, p := begin("name")
		e["name"] = name
		v, ok := in.NameResults[name]
		if !ok {
			t.Fatal("undeclared original type name", name)
		}
		v = result32(p, v)
		finish(e, p, v, nil)
		return v
	}
	rt.New = func(id uint32) *server.Object {
		e, p := begin("new")
		e["type_id"] = id
		role := in.Projectile
		if p.Result != nil {
			role = int(*p.Result)
		}
		projectile := m.object(role)
		finish(e, p, role, nil)
		active.projectile = unsafe.Pointer(projectile)
		return projectile
	}
	rt.Create = func(projectile, owner *server.Object, position types.Pointf, extra uint32) {
		e, p := begin("create")
		e["object"], e["owner"] = m.role(unsafe.Pointer(projectile)), m.role(unsafe.Pointer(owner))
		e["position"], e["extra"] = [2]uint32{math.Float32bits(position.X), math.Float32bits(position.Y)}, extra
		finish(e, p, nil, nil)
		active.phase = "projectile-direction-store"
	}
	rt.Balance = func() float64 {
		e, p := begin("damage")
		e["name"] = "ArrowTrapDamage"
		v := result32(p, in.DamageBits)
		if active.projectile != nil {
			active.damage = (*server.Object)(active.projectile).CollideData
		}
		finish(e, p, v, nil)
		active.phase = "projectile-damage-store"
		return float64(math.Float32frombits(v))
	}
	rt.Audio = func(id uint32, source *server.Object) {
		e, p := begin("audio")
		e["sound"], e["unit"] = id, m.role(unsafe.Pointer(source))
		finish(e, p, nil, nil)
	}
	rt.TrapFX = func(position *types.Pointf, variant uint32) {
		e, p := begin("fx")
		owner := -99
		for n := 0; n < 4; n++ {
			if position == &m.object(n).PosVec {
				owner = n
				break
			}
		}
		if owner < 0 {
			t.Fatal("undeclared live native position pointer")
		}
		e["position_owner"], e["position"], e["variant"] = owner, [2]uint32{math.Float32bits(position.X), math.Float32bits(position.Y)}, variant
		finish(e, p, nil, nil)
	}
	rt.Enemy = func(source, candidate *server.Object) uint32 {
		e, p := begin("enemy")
		e["objects"] = [2]int{m.role(unsafe.Pointer(source)), m.role(unsafe.Pointer(candidate))}
		v := result32(p, in.Enemy)
		finish(e, p, v, nil)
		return v
	}
	rt.MapCheck = func(candidate, source *server.Object) uint32 {
		e, p := begin("map")
		e["objects"] = [2]int{m.role(unsafe.Pointer(candidate)), m.role(unsafe.Pointer(source))}
		v := result32(p, in.MapCheck)
		finish(e, p, v, nil)
		active.phase = "candidate-source-direction"
		return v
	}
	rt.Front = func(source *server.Object, direction int32, candidate *server.Object) uint32 {
		e, p := begin("front")
		e["source"], e["target"], e["direction"] = m.role(unsafe.Pointer(source)), m.role(unsafe.Pointer(candidate)), uint32(direction)
		e["source_position"], e["target_position"] = [2]uint32{math.Float32bits(source.PosVec.X), math.Float32bits(source.PosVec.Y)}, [2]uint32{math.Float32bits(candidate.PosVec.X), math.Float32bits(candidate.PosVec.Y)}
		v := result32(p, in.InFront)
		finish(e, p, v, nil)
		return v
	}
	var directions struct {
		Rows []struct {
			Direction int16
			Words     [2]uint32
		}
	}
	data, err := os.ReadFile("../server/testdata/skull_raw_direction_data_54f9a0.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "3e1fe81326cb6c663fc63e12dcc35ac13e78b5c7c00a088ee78fc946e329d93d" {
		t.Fatal("immutable original direction transport")
	}
	if json.Unmarshal(data, &directions) != nil || len(directions.Rows) != 260 {
		t.Fatal("independent original direction data")
	}
	rt.DirectionComponent = func(direction int16, component uint32) float32 {
		if component > 1 {
			t.Fatal("direction component")
		}
		for _, row := range directions.Rows {
			if row.Direction == direction {
				return math.Float32frombits(row.Words[component])
			}
		}
		t.Fatal("undeclared original signed direction", direction)
		return 0
	}
	withInvocation := func(call skullRawInvoke54F9A0, body func() *uint32) (result *uint32) {
		previous := active
		context := &invocation{skullRawInvoke54F9A0: call}
		switch call.Entry {
		case "54f9a0":
			context.phase = "update-source-flags"
			if call.Unit >= 0 {
				context.data = m.object(call.Unit).UpdateData
				context.phase = "update-enabled-byte"
				if uint32(m.object(call.Unit).ObjFlags)&0x01000000 == 0 {
					context.phase = "update-disabled-byte"
				}
			}
		case "54fa80":
			context.phase = "projectile-source-radius"
		case "54fbb0":
			context.phase = "wrapper-source-data-pointer"
			if call.Unit >= 0 {
				context.data = m.object(call.Unit).UpdateData
			}
		case "54fbf0":
			context.phase = "scan-source-x"
		case "54fc50":
			context.phase = "candidate-class"
		default:
			t.Fatal("undeclared entry", call.Entry)
		}
		active = context
		returned := false
		defer func() {
			if !returned && failed == nil {
				failed = context
			}
			active = previous
		}()
		result = body()
		returned = true
		return
	}
	rt.Enumerate = func(rect types.Rectf, source *server.Object, callback func(*server.Object, *server.Object)) {
		e, p := begin("enumerate")
		unit := m.role(unsafe.Pointer(source))
		e["rectangle"], e["source"], e["callback"] = [4]uint32{math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y)}, unit, "54fc50"
		targets := in.Targets
		if p.Targets != nil {
			targets = p.Targets
		}
		finish(e, p, nil, func() {
			for _, target := range targets {
				c := frame("candidate", e["index"].(int))
				c["entry"], c["arguments"] = "54fc50", [2]int{target, unit}
				events = append(events, c)
				withInvocation(skullRawInvoke54F9A0{Entry: "54fc50", Unit: unit, Target: target}, func() *uint32 { callback(m.object(target), source); return nil })
				c = frame("candidate-return", e["index"].(int))
				c["result"] = nil
				events = append(events, c)
			}
		})
		if active.Entry == "54fbb0" {
			active.phase = "wrapper-state-byte"
			if g.State.Reachable == 0 {
				active.phase = "wrapper-state-clear"
			}
		}
	}
	oldState, oldRuntime := skullRawOriginalStateOverride54F9A0, skullRawOriginalRuntimeOverride54F9A0
	skullRawOriginalStateOverride54F9A0, skullRawOriginalRuntimeOverride54F9A0 = &g.State, &rt
	defer func() {
		skullRawOriginalStateOverride54F9A0, skullRawOriginalRuntimeOverride54F9A0 = oldState, oldRuntime
	}()
	invoke = func(call skullRawInvoke54F9A0) *uint32 {
		return withInvocation(call, func() (result *uint32) {
			source := m.object(call.Unit)
			if mode == "C" {
				index, ok := map[string]int{"54f9a0": 0, "54fa80": 1, "54fbb0": 2, "54fbf0": 3, "54fc50": 4}[call.Entry]
				if !ok || skullOriginalABI54F9A0(index) != 64 {
					t.Fatal("actual typed native public C entry required")
				}
				v := skullOriginalC54F9A0(index, source, m.object(call.Target), call.ProjectileType)
				if call.Entry == "54f9a0" || call.Entry == "54fbf0" {
					result = &v
				}
				return
			}
			body := any(source).(skullRawContract54F9A0)
			switch call.Entry {
			case "54f9a0":
				v := body.SkullRawUpdate54F9A0(&g.State, rt)
				result = &v
			case "54fa80":
				body.SkullRawProjectile54FA80(call.ProjectileType, &g.State, rt)
			case "54fbb0":
				body.SkullRawState54FBB0(&g.State, rt)
			case "54fbf0":
				v := body.SkullRawScan54FBF0(&g.State, rt)
				result = &v
			case "54fc50":
				body.SkullRawReachable54FC50(m.object(call.Target), &g.State, rt)
			}
			return
		})
	}
	oldFault := debug.SetPanicOnFault(true)
	defer debug.SetPanicOnFault(oldFault)
	var failure any
	func() {
		defer func() { failure = recover() }()
		out.Result = invoke(skullRawInvoke54F9A0{Entry: in.Entry, Unit: in.Unit, Target: in.Target, ProjectileType: in.ProjectileType})
	}()
	if failure != nil {
		text := fmt.Sprint(failure)
		if failed == nil || !strings.Contains(text, "memory") && !strings.Contains(text, "fault") {
			t.Fatal("not an actual native memory fault", failure)
		}
		site := failed.phase
		if fault, ok := failure.(interface{ Addr() uintptr }); ok && fault.Addr() != 0 {
			address := fault.Addr()
			o := server.Object{}
			source := uintptr(m.pointer(failed.Unit))
			candidate := uintptr(m.pointer(failed.Target))
			known := map[uintptr]string{}
			switch failed.Entry {
			case "54f9a0":
				known[source+unsafe.Offsetof(o.ObjFlags)] = "update-source-flags"
				known[source+unsafe.Offsetof(o.UpdateData)] = "update-source-data-pointer"
			case "54fa80":
				known[source+unsafe.Offsetof(o.Shape)+unsafe.Offsetof(o.Shape.Circle)] = "projectile-source-radius"
				known[source+unsafe.Offsetof(o.Direction1)] = "projectile-source-direction"
			case "54fbb0":
				known[source+unsafe.Offsetof(o.UpdateData)] = "wrapper-source-data-pointer"
			case "54fbf0":
				known[source+unsafe.Offsetof(o.PosVec)] = "scan-source-x"
			case "54fc50":
				known[candidate+unsafe.Offsetof(o.ObjClass)] = "candidate-class"
				known[candidate+unsafe.Offsetof(o.ObjFlags)] = "candidate-flags"
				known[source+unsafe.Offsetof(o.Direction1)] = "candidate-source-direction"
			}
			if failed.data != nil {
				if failed.Entry == "54f9a0" {
					known[uintptr(failed.data)+48] = failed.phase
				} else {
					known[uintptr(failed.data)+8] = "wrapper-state-byte"
					if g.State.Reachable == 0 {
						known[uintptr(failed.data)+8] = "wrapper-state-clear"
					}
				}
			}
			if failed.projectile != nil {
				known[uintptr(failed.projectile)+unsafe.Offsetof(o.Direction1)] = "projectile-direction-store"
			}
			if failed.damage != nil {
				known[uintptr(failed.damage)] = "projectile-damage-store"
			}
			value, found := known[address]
			if !found {
				t.Fatalf("unclassified actual native fault at%x: %v; phase%s", address, failure, site)
			}
			site = value
		}
		out.Fault = map[string]any{"kind": "memory", "site": site}
	}
	if g.Before != guards.Before || g.After != guards.After {
		t.Fatal("complete six-dword native state guards")
	}
	out.Name, out.Events, out.Caches, out.FPS, out.Heap = in.Name, events, caches(), fps, m.checkpoint()
	return
}
