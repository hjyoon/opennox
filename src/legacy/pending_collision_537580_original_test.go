package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type pendingState537580 struct {
	Objects [][6]uint32 `json:"objects"`
	Links   []int       `json:"links"`
	Head    int         `json:"head"`
	Tail    int         `json:"tail"`
	Path    uint32      `json:"path"`
}
type pendingEvent537580 struct {
	Service string `json:"service"`
	pendingState537580
}
type pendingStep537580 struct {
	Result int64                `json:"result"`
	Events []pendingEvent537580 `json:"events"`
	pendingState537580
}
type pendingInput537580 struct {
	Name           string               `json:"name"`
	Objects        [][6]uint32          `json:"objects"`
	Lookup         [2]uint32            `json:"lookup"`
	Ops            [][2]json.RawMessage `json:"ops"`
	MutationEvent  string               `json:"mutation_event"`
	MutationAction string               `json:"mutation_action"`
	MutationObject int                  `json:"mutation_object"`
}

func TestPendingCollisionOriginalNativeC(t *testing.T) {
	blob, err := os.ReadFile("testdata/pending_collision_537580_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Contexts       int    `json:"contexts"`
		RootExecutions int    `json:"root_executions"`
		OriginalSHA    string `json:"original_sha256"`
		Rows           []struct {
			Input     pendingInput537580  `json:"input"`
			Steps     []pendingStep537580 `json:"steps"`
			Preserved bool                `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 887 || len(f.Rows) != 887 || f.RootExecutions != 5353 || f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" {
		t.Fatal("original pending-collision metadata")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			if !row.Preserved || len(row.Steps) != len(row.Input.Ops) {
				t.Fatal("incomplete original record")
			}
			pendingCompare537580(t, row.Input, row.Steps)
		})
	}
}

func pendingCompare537580(t *testing.T, in pendingInput537580, wantSteps []pendingStep537580) {
	t.Helper()
	// Every pointer is backed by C-owned memory. Distinct callback identities
	// retain their original low byte, but are never called as executable code.
	callbacks := map[uint32]unsafe.Pointer{0: nil}
	identityFrame := radialCFrameOriginal(t, 512)
	for _, word := range []uint32{0x1000012a, 0x1000027b, 0x4eadf0, 0x4ebd40} {
		base := unsafe.Pointer(&identityFrame[32])
		offset := (uintptr(word) - uintptr(base)) & 255
		callbacks[word] = unsafe.Add(base, offset)
	}
	var objects [5]*server.Object
	var frames [5][]byte
	for n := range objects {
		frames[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		o := (*server.Object)(unsafe.Pointer(&frames[n][32]))
		objects[n] = o
		data := in.Objects[n]
		*o = server.Object{TypeInd: uint16(data[0]), ObjClass: object.Class(data[1]), ObjFlags: object.Flags(data[2]), Field116: data[3], Update: callbacks[data[4]], Collide: callbacks[data[5]], Field115: 0xffffffff, Field37: 0x55667788, Field38: 0x11223344}
	}
	machine := pendingProbeMachine537580(callbacks[0x4eadf0], callbacks[0x4ebd40])
	t.Cleanup(machine.close)
	ref := func(word uintptr) int {
		if word == 0 {
			return -1
		}
		if word == 0xffffffff {
			return -2
		}
		for n, o := range objects {
			if word == uintptr(unsafe.Pointer(o)) {
				return n
			}
		}
		t.Fatalf("foreign pending pointer %#x", word)
		return -3
	}
	callbackWord := func(p unsafe.Pointer) uint32 {
		for word, value := range callbacks {
			if p == value {
				return word
			}
		}
		t.Fatalf("foreign callback pointer %p", p)
		return 0
	}
	path := uint32(1)
	state := func() pendingState537580 {
		out := pendingState537580{Objects: make([][6]uint32, 5), Links: make([]int, 5), Head: ref(uintptr(unsafe.Pointer(machine.head()))), Tail: ref(machine.tail()), Path: path}
		for n, o := range objects {
			out.Objects[n] = [6]uint32{uint32(o.TypeInd), uint32(o.ObjClass), uint32(o.ObjFlags), o.Field116, callbackWord(o.Update), callbackWord(o.Collide)}
			out.Links[n] = ref(machine.link(o))
		}
		return out
	}
	events := []pendingEvent537580{}
	mutated := false
	mutate := func(stage string) {
		if mutated || stage != in.MutationEvent {
			return
		}
		mutated = true
		o := objects[in.MutationObject]
		switch in.MutationAction {
		case "class-allowed":
			o.ObjClass = object.Class(0x400080)
		case "class-blocked":
			o.ObjClass = object.Class(0x400000)
		case "class-actor":
			o.ObjClass = object.Class(8)
		case "class-plain":
			o.ObjClass = 0
		case "active":
			o.ObjFlags |= 4
		case "inactive":
			o.ObjFlags &^= 4
		case "nocollide":
			o.ObjFlags |= 0x40
		case "marked":
			o.Field116 |= 1
		case "status-high":
			o.Field116 = 0xfedcba80
		case "type-spike":
			o.TypeInd = uint16(in.Lookup[0])
		case "type-periodic":
			o.TypeInd = uint16(in.Lookup[1])
		case "collide-fist":
			o.Collide = callbacks[0x4eadf0]
		case "collide-undead":
			o.Collide = callbacks[0x4ebd40]
		default:
			t.Fatal("unknown original mutation")
		}
	}
	oldLookup, oldPrepare := pendingProbeLookup537610, pendingProbePrepare537610
	defer func() { pendingProbeLookup537610, pendingProbePrepare537610 = oldLookup, oldPrepare }()
	pendingProbeLookup537610 = func(name string) uint32 {
		index, stage := 0, "spike"
		if name == "PeriodicSpike" {
			index, stage = 1, "periodic"
		} else if name != "Spike" {
			t.Fatal("unexpected type lookup " + name)
		}
		events = append(events, pendingEvent537580{stage, state()})
		mutate(stage)
		return in.Lookup[index]
	}
	pendingProbePrepare537610 = func() { events = append(events, pendingEvent537580{"prepare", state()}); path = 0 }
	before := make([][]byte, 5)
	for n := range frames {
		before[n] = bytes.Clone(frames[n])
	}
	for step, raw := range in.Ops {
		var name string
		var index int
		if err := json.Unmarshal(raw[0], &name); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw[1], &index); err != nil {
			t.Fatal(err)
		}
		var o *server.Object
		if index >= 0 {
			o = objects[index]
		}
		var result int64
		switch name {
		case "enqueue":
			result = int64(machine.enqueue(o))
		case "contains":
			result = int64(machine.contains(o))
		case "empty":
			result = int64(machine.empty())
		case "remove":
			machine.remove(o)
		case "pop":
			result = int64(ref(uintptr(unsafe.Pointer(machine.pop()))))
		case "head":
			result = int64(ref(uintptr(unsafe.Pointer(machine.head()))))
		case "next":
			result = int64(ref(machine.next(o)))
		default:
			t.Fatal("unknown operation " + name)
		}
		got := pendingStep537580{result, append([]pendingEvent537580{}, events...), state()}
		events = events[:0]
		want := wantSteps[step]
		if !reflect.DeepEqual(got, want) {
			a, _ := json.Marshal(got)
			b, _ := json.Marshal(want)
			t.Fatalf("step %d %s: actual=%s original=%s", step, name, a, b)
		}
		compat := machine.compat()
		head, tail := uint32(0), uint32(0)
		if want.Head >= 0 {
			head = 1
		}
		if want.Tail >= 0 {
			tail = 1
		}
		if compat != [2]uint32{head, tail} {
			t.Fatalf("native compatibility globals %v", compat)
		}
		for n, frame := range frames {
			expected := bytes.Clone(before[n])
			obj := (*server.Object)(unsafe.Pointer(&expected[32]))
			data := want.Objects[n]
			// Write through byte slices, never install pointers in the Go-owned copy.
			binary.LittleEndian.PutUint16(expected[32+unsafe.Offsetof(obj.TypeInd):], uint16(data[0]))
			for offset, value := range map[uintptr]uint32{unsafe.Offsetof(obj.ObjClass): data[1], unsafe.Offsetof(obj.ObjFlags): data[2], unsafe.Offsetof(obj.Field116): data[3]} {
				binary.LittleEndian.PutUint32(expected[32+offset:], value)
			}
			link := uint32(0)
			if want.Links[n] == -2 {
				link = 0xffffffff
			}
			binary.LittleEndian.PutUint32(expected[32+unsafe.Offsetof(obj.Field115):], link)
			for offset, word := range map[uintptr]uint32{unsafe.Offsetof(obj.Update): data[4], unsafe.Offsetof(obj.Collide): data[5]} {
				value := uintptr(callbacks[word])
				if unsafe.Sizeof(value) == 8 {
					binary.LittleEndian.PutUint64(expected[32+offset:], uint64(value))
				} else {
					binary.LittleEndian.PutUint32(expected[32+offset:], uint32(value))
				}
			}
			if !bytes.Equal(frame, expected) {
				t.Fatal(fmt.Sprintf("step %d %s object %d record or canary changed", step, name, n))
			}
		}
	}
}

func TestPendingCollisionOriginalNilFaults(t *testing.T) {
	blob, err := os.ReadFile("testdata/pending_collision_nil_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var proof struct {
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Name  string `json:"name"`
			Fault struct {
				Address uint32 `json:"address"`
			} `json:"fault"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &proof); err != nil {
		t.Fatal(err)
	}
	if len(proof.Rows) != 4 || proof.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" {
		t.Fatal("original fault provenance")
	}
	for _, row := range proof.Rows {
		t.Run(row.Name, func(t *testing.T) {
			var queue server.PendingCollisionQueue537580
			defer func() {
				if recover() == nil {
					t.Fatal("original dependent nil access must fault")
				}
			}()
			switch row.Name {
			case "contains":
				if row.Fault.Address != 464 {
					t.Fatal("fault address")
				}
				queue.Contains(nil)
			case "remove":
				if row.Fault.Address != 464 {
					t.Fatal("fault address")
				}
				queue.Remove(nil)
			case "enqueue":
				if row.Fault.Address != 744 {
					t.Fatal("fault address")
				}
				queue.Enqueue(nil, server.PendingCollisionRuntime537610{})
			case "pop":
				if row.Fault.Address != 460 {
					t.Fatal("fault address")
				}
				queue.Pop()
			default:
				t.Fatal("unknown original fault")
			}
		})
	}
}

func TestPendingCollisionActualNativeCallbackIdentities(t *testing.T) {
	runtime := pendingCollisionRuntime537610
	for name, callback := range map[string]unsafe.Pointer{"fist": runtime.Fist, "undead-killer": runtime.UndeadKiller} {
		t.Run(name, func(t *testing.T) {
			if callback == nil || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(callback) <= 0xffffffff {
				t.Fatal("native static C callback identity")
			}
			frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
			o := (*server.Object)(unsafe.Pointer(&frame[32]))
			identityFrame := radialCFrameOriginal(t, 256)
			*o = server.Object{TypeInd: 100, ObjClass: object.Class(0x400000), ObjFlags: 4, Field115: 0xffffffff, Field116: 0x112233e0, Update: unsafe.Pointer(&identityFrame[32]), Collide: callback}
			before := bytes.Clone(frame)
			oldLookup, oldPrepare := pendingProbeLookup537610, pendingProbePrepare537610
			defer func() { pendingProbeLookup537610, pendingProbePrepare537610 = oldLookup, oldPrepare }()
			pendingProbeLookup537610 = func(name string) uint32 {
				if name == "Spike" {
					return 1001
				}
				if name == "PeriodicSpike" {
					return 1002
				}
				t.Fatal("unexpected lookup")
				return 0
			}
			pendingProbePrepare537610 = func() { t.Fatal("unexpected path reset") }
			machine := pendingProbeMachine537580(runtime.Fist, runtime.UndeadKiller)
			t.Cleanup(machine.close)
			if got := machine.enqueue(o); got != 0xe1 {
				t.Fatalf("activation byte=%x", got)
			}
			if machine.head() != o || machine.tail() != uintptr(unsafe.Pointer(o)) || machine.next(o) != 0 {
				t.Fatal("native callback queue identity")
			}
			expected := bytes.Clone(before)
			binary.LittleEndian.PutUint32(expected[32+unsafe.Offsetof(o.Field115):], 0)
			binary.LittleEndian.PutUint32(expected[32+unsafe.Offsetof(o.Field116):], 0x112233e1)
			if !bytes.Equal(frame, expected) {
				t.Fatal("callback object record or canary changed")
			}
			if machine.pop() != o || machine.empty() != 1 || !bytes.Equal(frame, before) {
				t.Fatal("native callback drain or full record restoration")
			}
		})
	}
}
