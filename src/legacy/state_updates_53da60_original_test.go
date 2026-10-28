package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type stateOriginal53DA60 struct {
	Objects [4][7]uint32 `json:"objects"`
	Owners  [4]int       `json:"owners"`
	Frame   uint32       `json:"frame"`
	FPS     uint32       `json:"fps"`
	Cache   uint32       `json:"cache"`
}
type stateInput53DA60 struct {
	Name     string `json:"name"`
	Kind     int    `json:"kind"`
	Lookup   uint32 `json:"lookup"`
	Spawn    int    `json:"spawn_ref"`
	Delay    int32  `json:"delay"`
	Mutation string `json:"mutation"`
	stateOriginal53DA60
}
type stateEvent53DA60 struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	stateOriginal53DA60
}
type stateRow53DA60 struct {
	Input     stateInput53DA60    `json:"input"`
	Output    stateOriginal53DA60 `json:"output"`
	Events    []stateEvent53DA60  `json:"events"`
	Preserved bool                `json:"nonoutput_bytes_preserved"`
}

func stateOriginalRun53DA60(t *testing.T, in stateInput53DA60, sourceRef int, want stateOriginal53DA60) (stateOriginal53DA60, []stateEvent53DA60, any) {
	t.Helper()
	var objs [4]*server.Object
	var frames, before [4][]byte
	for n := range objs {
		frames[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		objs[n] = (*server.Object)(unsafe.Pointer(&frames[n][32]))
		*objs[n] = server.Object{}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(objs[n])) <= math.MaxUint32 {
			t.Fatal("not a native high pointer")
		}
	}

	ref := func(n int) *server.Object {
		if n < 0 {
			return nil
		}
		return objs[n]
	}
	id := func(o *server.Object) int {
		if o == nil {
			return -1
		}
		for n, p := range objs {
			if p == o {
				return n
			}
		}
		t.Fatal("unknown object")
		return -2
	}
	for n, v := range in.Objects {
		o := objs[n]
		o.ObjFlags = object.Flags(v[0])
		o.Field5 = v[1]
		o.PosVec = types.Ptf(math.Float32frombits(v[2]), math.Float32frombits(v[3]))
		o.Field32 = v[4]
		o.Field34 = v[5]
		o.IsUpdatable = v[6]
		o.ObjOwner = ref(in.Owners[n])
	}
	for n := range objs {
		before[n] = append([]byte(nil), frames[n]...)
	}
	snapshot := func() stateOriginal53DA60 {
		s := stateOriginal53DA60{Frame: in.Frame, FPS: in.FPS, Cache: in.Cache}
		for n, o := range objs {
			s.Objects[n] = [7]uint32{uint32(o.ObjFlags), o.Field5, math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), o.Field32, o.Field34, o.IsUpdatable}
			s.Owners[n] = id(o.ObjOwner)
		}
		return s
	}
	events := make([]stateEvent53DA60, 0)
	event := func(service string, args ...int64) {
		if args == nil {
			args = []int64{}
		}
		events = append(events, stateEvent53DA60{Service: service, Args: args, stateOriginal53DA60: snapshot()})
	}
	applied := false
	mutate := func(service string) {
		prefix := service + ":"
		if applied || !strings.HasPrefix(in.Mutation, prefix) {
			return
		}
		applied = true
		switch strings.TrimPrefix(in.Mutation, prefix) {
		case "clock":
			in.Frame = 0xfffffffc
		case "fps":
			in.FPS = 0x80000001
		case "owner":
			objs[0].ObjOwner = objs[3]
		case "owner-nil":
			objs[0].ObjOwner = nil
		case "position":
			objs[0].PosVec = types.Ptf(math.Float32frombits(0x80000000), math.Float32frombits(0x7fc12345))
		case "created":
			objs[0].Field32 = 0xfffffffa
		case "deadline":
			objs[0].Field34 = 0x12345678
		case "cache":
			in.Cache = 99
		case "flags":
			objs[0].ObjFlags = 0x12348001
		case "status":
			objs[0].Field5 = 0x10000002
		default:
			t.Fatal("unknown mutation")
		}
	}
	a := server.ArachnaphobiaRawRuntime53DA60{
		Frame: func() uint32 { return in.Frame }, FPS: func() uint32 { return in.FPS },
		LoadTypeIndex: func() uint32 { return in.Cache }, StoreTypeIndex: func(v uint32) { in.Cache = v },
		TypeIndexByID: func(name string) uint32 {
			if name != "SmallSpider" {
				t.Fatal(name)
			}
			event("lookup")
			mutate("lookup")
			return in.Lookup
		},
		NewObjectByTypeInd: func(index int32) *server.Object { event("new", int64(index)); mutate("new"); return ref(in.Spawn) },
		CreateAt: func(o, owner *server.Object, p types.Pointf) {
			event("create", int64(id(o)), int64(id(owner)), int64(math.Float32bits(p.X)), int64(math.Float32bits(p.Y)), 0)
			mutate("create")
		},
		RandomInt:     func(lo, hi int) int { event("random", int64(lo), int64(hi)); mutate("random"); return int(in.Delay) },
		DelayedDelete: func(o *server.Object) { event("delete", int64(id(o))); mutate("delete") },
	}
	h := server.StateUpdateRawRuntime53DB30{
		Frame: a.Frame, FPS: a.FPS,
		UnsetXStatus: func(o *server.Object, v uint32) {
			event("unset", int64(id(o)), int64(v))
			o.Field5 &^= v
			mutate("unset")
		},
		SetXStatus:          func(o *server.Object, v uint32) { event("set", int64(id(o)), int64(v)); o.Field5 |= v; mutate("set") },
		RemoveFromUpdatable: func(o *server.Object) { event("remove", int64(id(o))); o.IsUpdatable = 0; mutate("remove") },
		DelayedDelete:       a.DelayedDelete,
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		oldA, oldH := arachnaphobiaRawRuntime53DA60, stateUpdateRawRuntime53DB30
		defer func() { arachnaphobiaRawRuntime53DA60 = oldA; stateUpdateRawRuntime53DB30 = oldH }()
		arachnaphobiaRawRuntime53DA60 = func() server.ArachnaphobiaRawRuntime53DA60 { return a }
		stateUpdateRawRuntime53DB30 = func() server.StateUpdateRawRuntime53DB30 { return h }
		stateUpdateNativeProbe53DA60(in.Kind, ref(sourceRef))
	}()
	for n, o := range objs {
		expected := append([]byte(nil), before[n]...)
		offsets := []uintptr{unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.Field5), unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.PosVec) + 4, unsafe.Offsetof(o.Field32), unsafe.Offsetof(o.Field34), unsafe.Offsetof(o.IsUpdatable)}
		for k, off := range offsets {
			binary.LittleEndian.PutUint32(expected[32+off:], want.Objects[n][k])
		}
		owner := uintptr(unsafe.Pointer(ref(want.Owners[n])))
		off := 32 + unsafe.Offsetof(o.ObjOwner)
		if unsafe.Sizeof(owner) == 8 {
			binary.LittleEndian.PutUint64(expected[off:], uint64(owner))
		} else {
			binary.LittleEndian.PutUint32(expected[off:], uint32(owner))
		}
		if !bytes.Equal(expected, frames[n]) {
			t.Fatalf("native object frame/canary %d", n)
		}
	}
	return snapshot(), events, recovered
}

func TestStateUpdatesOriginalNativeC53DA60(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/state_updates_53da60_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA      string           `json:"original_sha256"`
		CW       uint32           `json:"original_startup_FPCW"`
		Contexts int              `json:"contexts"`
		Rows     []stateRow53DA60 `json:"rows"`
	}
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.CW != 0xe3f || fixture.Contexts != 759 || len(fixture.Rows) != 759 {
		t.Fatal("original provenance")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, fault := stateOriginalRun53DA60(t, row.Input, 0, row.Output)
			if !row.Preserved || fault != nil {
				t.Fatal("original/native fault or canary", fault)
			}
			if !reflect.DeepEqual(out, row.Output) {
				t.Errorf("state differs: got %+v want %+v", out, row.Output)
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("service trace differs: got %+v want %+v", events, row.Events)
			}
		})
	}
}

func TestStateUpdatesOriginalFaultsNativeC53DA60(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/state_updates_53da60_faults_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			stateRow53DA60
			Source int `json:"source_ref"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 4 {
		t.Fatal("fault provenance")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			out, events, fault := stateOriginalRun53DA60(t, row.Input, row.Source, row.Output)
			if fault == nil || !reflect.DeepEqual(out, row.Output) || !reflect.DeepEqual(events, row.Events) {
				t.Fatal("required source fault/prefix/state differs", fault)
			}
		})
	}
}
