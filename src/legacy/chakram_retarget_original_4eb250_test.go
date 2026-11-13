package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"runtime/debug"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type chakramRetargetRawContract4EB250 interface {
	ChakramRetargetNative4EB250(*server.ChakramRetargetState4EB250, server.ChakramRetargetRuntime4EB250) *server.Object
	ChakramRetargetCandidateNative4EB340(*types.Pointf, *server.ChakramRetargetState4EB250, server.ChakramRetargetRuntime4EB250)
}
type chakramOriginalObject4EB250 struct {
	Class  uint32    `json:"cls"`
	Flags  uint32    `json:"flags"`
	Pos    [2]uint32 `json:"pos"`
	Speed  uint32    `json:"speed"`
	Vel    [2]uint32 `json:"vel"`
	Owner  int       `json:"owner"`
	Update int       `json:"update"`
}
type chakramOriginalUpdate4EB250 struct {
	Last  int   `json:"last"`
	State uint8 `json:"state"`
}
type chakramOriginalScratch4EB250 struct {
	Source   int    `json:"source"`
	Owner    int    `json:"owner"`
	Last     int    `json:"last"`
	Selected int    `json:"selected"`
	Best     uint32 `json:"best"`
}
type chakramOriginalState4EB250 struct {
	Objects []chakramOriginalObject4EB250 `json:"objects"`
	Updates []chakramOriginalUpdate4EB250 `json:"updates"`
	Scratch chakramOriginalScratch4EB250  `json:"scratch"`
}
type chakramOriginalMutation4EB250 struct {
	Objects []struct {
		ID            int
		Class         *uint32 `json:"cls"`
		Flags, Speed  *uint32
		Pos, Vel      *[2]uint32
		Owner, Update *int
	}
	Updates []struct {
		ID    int
		Last  *int
		State *uint8
	}
	Scratch map[string]int64
}
type chakramOriginalNested4EB250 struct{ Source int }
type chakramOriginalStep4EB250 struct {
	Value       *uint32
	Mutation    chakramOriginalMutation4EB250
	AfterNested chakramOriginalMutation4EB250 `json:"after_nested"`
	Reenter     *chakramOriginalNested4EB250
}
type chakramOriginalEnum4EB250 struct {
	Candidates    []int
	Before, After chakramOriginalMutation4EB250
	AfterEach     []chakramOriginalMutation4EB250 `json:"after_each"`
	Reenter       *chakramOriginalNested4EB250
}
type chakramOriginalInput4EB250 struct {
	Name, Entry              string
	Source, Candidate, Point int
	PointOffset              int `json:"point_offset"`
	Objects                  []chakramOriginalObject4EB250
	Updates                  []chakramOriginalUpdate4EB250
	Scratch                  chakramOriginalScratch4EB250
	Enums                    []chakramOriginalEnum4EB250
	Buffs, Maps              []chakramOriginalStep4EB250
}
type chakramOriginalRow4EB250 struct {
	Input  chakramOriginalInput4EB250
	Output struct {
		Result *int
		Events []json.RawMessage
		State  chakramOriginalState4EB250
		Fault  []json.RawMessage
		Guard  bool `json:"whole_heap_scratch_guarded_caller_valid_nonvolatile_and_FPU_preserved"`
	}
}

func TestChakramRetargetOriginal4EB250(t *testing.T) {
	TestChakramRetargetNativeLayout4EB250(t)
	profiles := []struct {
		name, gzipSHA, decodedSHA string
		nearest                   bool
	}{
		{"pc53_tozero", "e03c702b1cbe7aaaec0c71746bc48376bcb8fdfb30ce50c8ffb8d579cbd332eb", "ee3ad3ef672f9c805230c5b81940e8d27c6eddaa50b81d30de2a5ed8025e0a31", false},
		{"pc53_nearest", "b44f412c643178ab6d9b1b30d4df37d9b145d70079683001bdfdcdccbb3b7aa3", "4b71bece36f8c49c3523ff7006ba1688eb0985150daacbe94837360386b08315", true},
	}
	for _, profile := range profiles {
		raw, err := os.ReadFile("../server/testdata/chakram_retarget_native_original_4eb250_" + profile.name + ".json.gz")
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != profile.gzipSHA {
			t.Fatal("frozen original gzip changed", profile.name)
		}
		z, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := io.ReadAll(z)
		if err != nil {
			t.Fatal(err)
		}
		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(decoded)) != profile.decodedSHA {
			t.Fatal("frozen decoded original changed", profile.name)
		}
		var data struct{ Rows []chakramOriginalRow4EB250 }
		if err := json.Unmarshal(decoded, &data); err != nil || len(data.Rows) != 272 {
			t.Fatal("original input corpus", err, len(data.Rows))
		}
		for _, mode := range []string{"C", "Go"} {
			t.Run(profile.name+"/"+mode, func(t *testing.T) {
				for _, row := range data.Rows {
					t.Run(row.Input.Name, func(t *testing.T) {
						in := row.Input
						if !row.Output.Guard {
							t.Fatal("original guards not certified")
						}
						invalidNative := originalInvalidNativePointer32(t)
						unmapped := func() unsafe.Pointer {
							if invalidNative != nil {
								return invalidNative
							}
							return secretWallOriginalUnmapped4DF120()
						}
						buf, free := alloc.MakeHigh([]byte{}, 0x10000)
						defer free()
						for i := range buf {
							buf[i] = byte(i*37 + 0xa5)
						}
						var objects [6]*server.Object
						var updates [3]*server.ChakramUpdateData
						for i := range objects {
							objects[i] = (*server.Object)(unsafe.Pointer(&buf[0x1000+i*0x1000]))
						}
						for i := range updates {
							updates[i] = (*server.ChakramUpdateData)(unsafe.Pointer(&buf[0x8000+i*0x1000]))
						}
						at := func(index int) *server.Object {
							if index == -1 {
								return nil
							}
							if index == -2 {
								return (*server.Object)(unmapped())
							}
							return objects[index]
						}
						updateAt := func(index int) unsafe.Pointer {
							if index == -1 {
								return nil
							}
							if index == -2 {
								return unmapped()
							}
							return unsafe.Pointer(updates[index])
						}
						role := func(obj *server.Object) int {
							if obj == nil {
								return -1
							}
							if obj == at(-2) {
								return -2
							}
							for i, value := range objects {
								if obj == value {
									return i
								}
							}
							t.Fatalf("unknown native object pointer %p", obj)
							return -3
						}
						updateRole := func(ptr unsafe.Pointer) int {
							for i := -2; i < len(updates); i++ {
								if updateAt(i) == ptr {
									return i
								}
							}
							t.Fatalf("unknown native update pointer %p", ptr)
							return -3
						}
						for i, value := range in.Objects {
							obj := objects[i]
							obj.ObjClass, obj.ObjFlags = object.Class(value.Class), object.Flags(value.Flags)
							obj.PosVec = types.Ptf(math.Float32frombits(value.Pos[0]), math.Float32frombits(value.Pos[1]))
							obj.VelVec = types.Ptf(math.Float32frombits(value.Vel[0]), math.Float32frombits(value.Vel[1]))
							if unsafe.Sizeof(uintptr(0)) == 4 {
								// These initial pointer cells still contain raw canary words,
								// including 0x300be6c1 at UpdateData. Clear only bytes
								// this store replaces before Go's barrier reads them.
								alloc.Memset(unsafe.Pointer(&obj.ObjOwner), 0, unsafe.Sizeof(obj.ObjOwner))
								alloc.Memset(unsafe.Pointer(&obj.UpdateData), 0, unsafe.Sizeof(obj.UpdateData))
							}
							obj.SpeedCur, obj.ObjOwner, obj.UpdateData = math.Float32frombits(value.Speed), at(value.Owner), updateAt(value.Update)
						}
						for i, value := range in.Updates {
							if unsafe.Sizeof(uintptr(0)) == 4 {
								alloc.Memset(unsafe.Pointer(&updates[i].LastHit), 0, unsafe.Sizeof(updates[i].LastHit))
							}
							updates[i].LastHit, updates[i].ReturnState = at(value.Last), value.State
						}
						initial := bytes.Clone(buf)
						oldScratch, oldRuntime := chakramRetargetNativeScratch4EB250, chakramRetargetNativeRuntimeOverride4EB250
						defer func() {
							chakramRetargetNativeScratch4EB250, chakramRetargetNativeRuntimeOverride4EB250 = oldScratch, oldRuntime
						}()
						state := &chakramRetargetNativeScratch4EB250
						*state = server.ChakramRetargetState4EB250{Source: at(in.Scratch.Source), Owner: at(in.Scratch.Owner), LastHit: at(in.Scratch.Last), Selected: at(in.Scratch.Selected), BestDistance: math.Float32frombits(in.Scratch.Best)}
						observe := func() chakramOriginalState4EB250 {
							out := chakramOriginalState4EB250{Objects: make([]chakramOriginalObject4EB250, len(objects)), Updates: make([]chakramOriginalUpdate4EB250, len(updates))}
							for i, obj := range objects {
								out.Objects[i] = chakramOriginalObject4EB250{Class: uint32(obj.ObjClass), Flags: uint32(obj.ObjFlags), Pos: [2]uint32{math.Float32bits(obj.PosVec.X), math.Float32bits(obj.PosVec.Y)}, Speed: math.Float32bits(obj.SpeedCur), Vel: [2]uint32{math.Float32bits(obj.VelVec.X), math.Float32bits(obj.VelVec.Y)}, Owner: role(obj.ObjOwner), Update: updateRole(obj.UpdateData)}
							}
							for i, update := range updates {
								out.Updates[i] = chakramOriginalUpdate4EB250{Last: role(update.LastHit), State: update.ReturnState}
							}
							out.Scratch = chakramOriginalScratch4EB250{Source: role(state.Source), Owner: role(state.Owner), Last: role(state.LastHit), Selected: role(state.Selected), Best: math.Float32bits(state.BestDistance)}
							return out
						}
						mutate := func(m chakramOriginalMutation4EB250) {
							for _, change := range m.Objects {
								obj := objects[change.ID]
								if change.Class != nil {
									obj.ObjClass = object.Class(*change.Class)
								}
								if change.Flags != nil {
									obj.ObjFlags = object.Flags(*change.Flags)
								}
								if change.Speed != nil {
									obj.SpeedCur = math.Float32frombits(*change.Speed)
								}
								if change.Pos != nil {
									obj.PosVec = types.Ptf(math.Float32frombits(change.Pos[0]), math.Float32frombits(change.Pos[1]))
								}
								if change.Vel != nil {
									obj.VelVec = types.Ptf(math.Float32frombits(change.Vel[0]), math.Float32frombits(change.Vel[1]))
								}
								if change.Owner != nil {
									obj.ObjOwner = at(*change.Owner)
								}
								if change.Update != nil {
									obj.UpdateData = updateAt(*change.Update)
								}
							}
							for _, change := range m.Updates {
								if change.Last != nil {
									updates[change.ID].LastHit = at(*change.Last)
								}
								if change.State != nil {
									updates[change.ID].ReturnState = *change.State
								}
							}
							for key, value := range m.Scratch {
								switch key {
								case "source":
									state.Source = at(int(value))
								case "owner":
									state.Owner = at(int(value))
								case "last":
									state.LastHit = at(int(value))
								case "selected":
									state.Selected = at(int(value))
								case "best":
									state.BestDistance = math.Float32frombits(uint32(value))
								default:
									t.Fatal("unknown original mutation", key)
								}
							}
						}
						events := make([]map[string]any, 0)
						var runtime server.ChakramRetargetRuntime4EB250
						var driver func(*server.Object) *server.Object
						candidate := func(obj *server.Object, point *types.Pointf) {
							if mode == "C" {
								if chakramRetargetOriginalABI4EB250(1) != 64 {
									t.Fatal("raw C callback not implemented")
								}
								chakramRetargetOriginalC4EB250(1, obj, point)
							} else {
								native, ok := any(obj).(chakramRetargetRawContract4EB250)
								if !ok {
									t.Fatal("raw Go callback not implemented")
								}
								native.ChakramRetargetCandidateNative4EB340(point, state, runtime)
							}
						}
						reenter := func(nested *chakramOriginalNested4EB250, kind string, number int) {
							if nested == nil {
								return
							}
							events = append(events, map[string]any{"kind": "nested", "service": kind, "index": number, "source": nested.Source, "before": observe()})
							result := driver(at(nested.Source))
							events = append(events, map[string]any{"kind": "nested-return", "service": kind, "index": number, "result": role(result), "after": observe()})
						}
						enumIndex, buffIndex, mapIndex := 0, 0, 0
						service := func(kind string, number int, obj, source *server.Object, enchant uint32, step chakramOriginalStep4EB250) bool {
							e := map[string]any{"kind": kind, "index": number, "candidate": role(obj), "before": observe()}
							if kind == "buff" {
								e["enchant"] = enchant
							} else {
								e["source"] = role(source)
							}
							events = append(events, e)
							mutate(step.Mutation)
							reenter(step.Reenter, kind, number)
							mutate(step.AfterNested)
							value := uint32(1)
							if kind == "buff" {
								value = 0
							}
							if step.Value != nil {
								value = *step.Value
							}
							events = append(events, map[string]any{"kind": kind + "-return", "index": number, "value": value, "after": observe()})
							return value != 0
						}
						runtime = server.ChakramRetargetRuntime4EB250{
							RoundToNearest: profile.nearest,
							EachRect: func(rect types.Rectf, callback func(*server.Object, *types.Pointf), point *types.Pointf) {
								number := enumIndex
								enumIndex++
								plan := chakramOriginalEnum4EB250{}
								if number < len(in.Enums) {
									plan = in.Enums[number]
								}
								pointObject := -1
								for i, obj := range objects {
									if point == &obj.PosVec {
										pointObject = i
									}
								}
								if pointObject < 0 {
									t.Fatal("enumeration point must retain actual source position pointer")
								}
								events = append(events, map[string]any{"kind": "enum", "index": number, "rectangle": [4]uint32{math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y)}, "point_object": pointObject, "before": observe()})
								mutate(plan.Before)
								reenter(plan.Reenter, "enum", number)
								for i, index := range plan.Candidates {
									if mode == "C" {
										candidate(at(index), point)
									} else {
										callback(at(index), point)
									}
									// The runtime must invoke the actual raw callback even in
									// C mode; its argument identity and shared state are live.
									events = append(events, map[string]any{"kind": "callback-return", "enum": number, "index": i, "candidate": index, "after": observe()})
									if i < len(plan.AfterEach) {
										mutate(plan.AfterEach[i])
									}
								}
								mutate(plan.After)
								events = append(events, map[string]any{"kind": "enum-return", "index": number, "after": observe()})
							},
							HasEnchant: func(obj *server.Object, enchant uint32) bool {
								number := buffIndex
								buffIndex++
								step := chakramOriginalStep4EB250{}
								if number < len(in.Buffs) {
									step = in.Buffs[number]
								}
								return service("buff", number, obj, nil, enchant, step)
							},
							CanSee: func(obj, source *server.Object) bool {
								number := mapIndex
								mapIndex++
								step := chakramOriginalStep4EB250{}
								if number < len(in.Maps) {
									step = in.Maps[number]
								}
								return service("map", number, obj, source, 0, step)
							},
						}
						driver = func(obj *server.Object) *server.Object {
							if mode == "C" {
								if chakramRetargetOriginalABI4EB250(0) != 64 {
									t.Fatal("raw C driver not implemented")
								}
								return chakramRetargetOriginalC4EB250(0, obj, nil)
							}
							native, ok := any(obj).(chakramRetargetRawContract4EB250)
							if !ok {
								t.Fatal("raw Go driver not implemented")
							}
							return native.ChakramRetargetNative4EB250(state, runtime)
						}
						chakramRetargetNativeRuntimeOverride4EB250 = &runtime
						var result *server.Object
						var fault any
						oldFault := debug.SetPanicOnFault(true)
						defer debug.SetPanicOnFault(oldFault)
						func() {
							defer func() { fault = recover() }()
							if in.Entry == "driver" {
								result = driver(at(in.Source))
								return
							}
							var point *types.Pointf
							if in.Point < 0 {
								if invalidNative != nil && in.Point == -2 {
									point = (*types.Pointf)(unsafe.Add(invalidNative, in.PointOffset))
								} else {
									point = chakramRetargetOriginalBadPoint4EB250(in.Point, in.PointOffset)
								}
							} else {
								switch in.PointOffset {
								case 56:
									point = &at(in.Point).PosVec
								case 80:
									point = &at(in.Point).VelVec
								default:
									t.Fatal("unknown original point offset")
								}
							}
							candidate(at(in.Candidate), point)
						}()
						if (fault != nil) != (len(row.Output.Fault) != 0) {
							t.Fatal("original fault boundary", fault, row.Output.Fault)
						}
						if row.Output.Result != nil && role(result) != *row.Output.Result {
							t.Fatal("original return", role(result), *row.Output.Result)
						}
						actual, _ := json.Marshal(events)
						wantEvents, _ := json.Marshal(row.Output.Events)
						var actualJSON, wantJSON any
						if err := json.Unmarshal(actual, &actualJSON); err != nil {
							t.Fatal(err)
						}
						if err := json.Unmarshal(wantEvents, &wantJSON); err != nil {
							t.Fatal(err)
						}
						if !reflect.DeepEqual(actualJSON, wantJSON) {
							t.Fatalf("original complete ordered service/callback/state prefix differs\ngot %s\nwant %s", actual, wantEvents)
						}
						if got := observe(); !reflect.DeepEqual(got, row.Output.State) {
							t.Fatalf("complete original final state differs\ngot %+v\nwant %+v", got, row.Output.State)
						}
						layout := chakramRetargetOriginalLayout4EB250()
						put := func(offset, n int, value uint64) {
							for i := 0; i < n; i++ {
								initial[offset+i] = byte(value >> (8 * i))
							}
						}
						for i, v := range row.Output.State.Objects {
							base := 0x1000 + i*0x1000
							put(base+int(layout[0]), 4, uint64(v.Class))
							put(base+int(layout[1]), 4, uint64(v.Flags))
							for j := 0; j < 2; j++ {
								put(base+int(layout[2])+j*4, 4, uint64(v.Pos[j]))
								put(base+int(layout[3])+j*4, 4, uint64(v.Vel[j]))
							}
							put(base+int(layout[4]), int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(unsafe.Pointer(at(v.Owner)))))
							put(base+int(layout[5]), 4, uint64(v.Speed))
							put(base+int(layout[6]), int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(updateAt(v.Update))))
						}
						for i, v := range row.Output.State.Updates {
							base := 0x8000 + i*0x1000
							put(base+int(layout[8]), int(unsafe.Sizeof(uintptr(0))), uint64(uintptr(unsafe.Pointer(at(v.Last)))))
							put(base+int(layout[9]), 1, uint64(v.State))
						}
						if !bytes.Equal(buf, initial) {
							for i := range buf {
								if buf[i] != initial[i] {
									t.Fatalf("whole native64KiB records/canaries differ at%x got%02x want%02x", i, buf[i], initial[i])
								}
							}
						}
					})
				}
			})
		}
	}
}
