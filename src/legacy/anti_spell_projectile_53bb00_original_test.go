package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type antiSpellState53BB00 struct {
	Objects    [6][12]uint32 `json:"objects"`
	Owners     [6]int64      `json:"owners"`
	Updates    [2][7]int64   `json:"updates"`
	UpdateRefs [2]int        `json:"update_refs"`
	Selected   int64         `json:"selected"`
	Best       uint32        `json:"best"`
	Frame      uint32        `json:"frame"`
	FPS        uint32        `json:"fps"`
	RNG        int           `json:"rng"`
}
type antiSpellEvent53BB00 struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	antiSpellState53BB00
}
type antiSpellInput53BB00 struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	antiSpellState53BB00
	Order          []int  `json:"order"`
	InteractMask   uint32 `json:"interact_mask"`
	Candidate      int    `json:"candidate"`
	NativeGrid     bool   `json:"native_grid"`
	MutationEvent  string `json:"mutation_event"`
	MutationAction string `json:"mutation_action"`
}

func TestAntiSpellOriginalNative53BB00(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/anti_spell_53bb00_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		OriginalSHA string `json:"original_sha256"`
		Contexts    int    `json:"contexts"`
		Executions  int    `json:"root_executions"`
		Rows        []struct {
			Mode      string                 `json:"mode"`
			Input     antiSpellInput53BB00   `json:"input"`
			Events    []antiSpellEvent53BB00 `json:"events"`
			Mutated   bool                   `json:"mutated"`
			Preserved bool                   `json:"nonoutput_bytes_preserved"`
			antiSpellState53BB00
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Contexts != 776 || f.Executions != 1552 || len(f.Rows) != 1552 {
		t.Fatal("original AntiSpell provenance")
	}
	for _, entry := range []string{"go-root", "actual-c"} {
		t.Run(entry, func(t *testing.T) {
			seen := 0
			for _, row := range f.Rows {
				if row.Mode != "actual-startup" || entry == "go-root" && row.Input.Kind != "root" {
					continue
				}
				seen++
				t.Run(row.Input.Name, func(t *testing.T) {
					i := row.Input
					if !row.Preserved {
						t.Fatal("original frame canaries")
					}
					var objects [6]*server.Object
					var frames, before [6][]byte
					for n, w := range i.Objects {
						frames[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
						objects[n] = (*server.Object)(unsafe.Pointer(&frames[n][32]))
						o := objects[n]
						*o = server.Object{PosVec: types.Ptf(math.Float32frombits(w[0]), math.Float32frombits(w[1])), VelVec: types.Ptf(math.Float32frombits(w[2]), math.Float32frombits(w[3])), ForceVec: types.Ptf(math.Float32frombits(w[4]), math.Float32frombits(w[5])), SpeedCur: math.Float32frombits(w[6]), ObjFlags: object.Flags(w[7]), ObjClass: object.Class(w[8]), ObjSubClass: object.SubClass(w[9]), Field32: w[10], Field34: w[11]}
						o.NewPos = o.PosVec
						o.PrevPos = o.PosVec
						o.Shape.Kind = server.ShapeKindCircle
						o.Shape.Circle.R = 1
						o.Shape.Circle.R2 = 1
					}
					pointer := func(index int64) *server.Object {
						if index < 0 {
							return nil
						}
						if index >= 6 {
							t.Fatal("foreign original object reference")
						}
						return objects[index]
					}
					index := func(o *server.Object) int64 {
						if o == nil {
							return -1
						}
						for n, p := range objects {
							if p == o {
								return int64(n)
							}
						}
						t.Fatal("foreign native object pointer")
						return -2
					}
					for n, owner := range i.Owners {
						objects[n].ObjOwner = pointer(owner)
					}
					var updates [2]*server.MissileUpdateData
					var uframes, ubefore [2][]byte
					for n, w := range i.Updates {
						uframes[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.MissileUpdateData{}))
						updates[n] = (*server.MissileUpdateData)(unsafe.Pointer(&uframes[n][32]))
						*updates[n] = server.MissileUpdateData{Owner: pointer(w[0]), Target: pointer(w[1]), Field_8: uint32(w[2]), SpellID: int32(w[3]), Field_16: uint32(w[4]), Field_20: uint32(w[5]), Field_24: uint32(w[6])}
						ubefore[n] = bytes.Clone(uframes[n])
					}
					for n, k := range []int{0, 4} {
						objects[k].UpdateData = unsafe.Pointer(updates[i.UpdateRefs[n]])
					}
					sframe := radialCFrameOriginal(t, unsafe.Sizeof(server.AntiSpellTargetState53BD10{}))
					selection := (*server.AntiSpellTargetState53BD10)(unsafe.Pointer(&sframe[32]))
					*selection = server.AntiSpellTargetState53BD10{Target: pointer(i.Selected), BestDistance: math.Float32frombits(i.Best)}
					sbefore := bytes.Clone(sframe)
					world := new(server.Server)
					if i.NativeGrid {
						world = server.New(nil, nil, strman.New())
						t.Cleanup(world.Close)
						world.Map.Init()
						t.Cleanup(world.Map.Free)
						// The PE32 oracle prepares bucket links directly, including
						// inactive members. Seed the same list through the native
						// insertion API, then restore the exact original flag words.
						// Native insertion prepends, so reverse the declared visit order.
						for k := len(i.Order) - 1; k >= 0; k-- {
							o := objects[i.Order[k]]
							flags := o.ObjFlags
							o.ObjFlags = (flags | object.FlagActive) &^ object.FlagPartitioned
							world.Map.AddObjectToIndex(o)
							o.ObjFlags = flags
						}
						defer func() {
							for _, n := range i.Order {
								flags := objects[n].ObjFlags
								objects[n].ObjFlags |= object.FlagPartitioned
								world.Map.RemoveObjectFromIndex(objects[n])
								objects[n].ObjFlags = flags
							}
						}()
						old := GetServer
						GetServer = func() Server { return &powderCService53C9A0{srv: world} }
						defer func() { GetServer = old }()
					}
					for n := range frames {
						before[n] = bytes.Clone(frames[n])
					}
					rng := prand.New(i.RNG)
					world.Rand.Logic = rng
					frame, fps := i.Frame, i.FPS
					state := func() antiSpellState53BB00 {
						s := antiSpellState53BB00{Selected: index(selection.Target), Best: math.Float32bits(selection.BestDistance), Frame: frame, FPS: fps, RNG: rng.Index(), UpdateRefs: [2]int{-1, -1}}
						for n, o := range objects {
							s.Objects[n] = [12]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.VelVec.X), math.Float32bits(o.VelVec.Y), math.Float32bits(o.ForceVec.X), math.Float32bits(o.ForceVec.Y), math.Float32bits(o.SpeedCur), uint32(o.ObjFlags), uint32(o.ObjClass), uint32(o.ObjSubClass), o.Field32, o.Field34}
							s.Owners[n] = index(o.ObjOwner)
						}
						for n, u := range updates {
							s.Updates[n] = [7]int64{index(u.Owner), index(u.Target), int64(u.Field_8), int64(uint32(u.SpellID)), int64(u.Field_16), int64(u.Field_20), int64(u.Field_24)}
							for k, m := range []int{0, 4} {
								if objects[m].UpdateData == unsafe.Pointer(u) {
									s.UpdateRefs[k] = n
								}
							}
						}
						return s
					}
					events := []antiSpellEvent53BB00{}
					mutated := false
					floatVisit := 0
					visits := [6]int{}
					var runtime server.AntiSpellProjectileRuntime53BB00
					callRoot := func(o *server.Object) {
						if entry == "go-root" {
							server.AntiSpellProjectileUpdate53BB00(o, runtime)
						} else {
							antiSpellCProbe53BB00(0, o, nil)
						}
					}
					mutate := func(stage string) bool {
						if mutated || stage != i.MutationEvent {
							return false
						}
						mutated = true
						m := objects[0]
						switch i.MutationAction {
						case "position":
							m.PosVec = types.Ptf(-20.3, 300.1)
						case "far-position":
							m.PosVec = types.Ptf(3200.1, -300.2)
						case "target-position":
							objects[2].PosVec = types.Ptf(321.25, -200.3)
						case "velocity":
							m.VelVec = types.Ptf(5.1, -7.3)
						case "speed":
							m.SpeedCur = 1.75
						case "target-ref":
							updates[0].Target = objects[3]
						case "target-nil":
							updates[0].Target = nil
						case "replace-update":
							m.UpdateData = unsafe.Pointer(updates[1])
						case "owner":
							m.ObjOwner = objects[2]
						case "class":
							objects[2].ObjClass = 1
							objects[2].ObjSubClass = 0
						case "flags":
							objects[2].ObjFlags = 0x8020
						case "frame":
							frame, fps = 32, 0xffffffff
						case "best":
							selection.BestDistance = -1
						case "selected":
							selection.Target = objects[3]
						case "rng":
							rng.Reset(4095)
						case "nested-query":
							return true
						default:
							t.Fatal("unknown declared mutation")
						}
						return false
					}
					emit := func(stage string, args ...int64) bool {
						if args == nil {
							args = []int64{}
						}
						events = append(events, antiSpellEvent53BB00{stage, args, state()})
						return mutate(stage)
					}
					currentMissile := objects[0]
					enumerate := func(origin *types.Pointf, radius float32, visit func(*server.Object) bool) {
						m := currentMissile
						if origin != &m.PosVec || radius != 600 {
							t.Fatal("borrowed search origin or radius")
						}
						n := index(m)
						emit("enumerate:"+antiSpellSmallInt53BB00(n), n, int64(math.Float32bits(origin.X)), int64(math.Float32bits(origin.Y)), int64(math.Float32bits(radius)))
						if i.NativeGrid {
							if entry == "actual-c" {
								antiSpellCircleCProbe518170(origin, m)
							} else {
								world.Map.EachMissileInCirclePtr518170(origin, radius, visit)
							}
							return
						}
						order := i.Order
						if n == 4 {
							order = []int{5}
						}
						for _, k := range order {
							if !visit(objects[k]) {
								t.Fatal("original void callback must continue")
							}
						}
					}
					runtime = server.AntiSpellProjectileRuntime53BB00{TargetState: selection, Frame: func() uint32 { return frame }, FPS: func() uint32 { return fps }, EachMissilePtr: enumerate,
						EachMissile: func(pos types.Pointf, radius float32, visit func(*server.Object) bool) {
							if pos != currentMissile.PosVec && !(math.IsNaN(float64(pos.X)) || math.IsNaN(float64(pos.Y))) {
								t.Fatal("search origin value")
							}
							if i.NativeGrid {
								n := index(currentMissile)
								emit("enumerate:"+antiSpellSmallInt53BB00(n), n, int64(math.Float32bits(pos.X)), int64(math.Float32bits(pos.Y)), int64(math.Float32bits(radius)))
								world.Map.EachMissileInCircle(pos, radius, visit)
							} else {
								enumerate(&currentMissile.PosVec, radius, visit)
							}
						},
						CanInteract: func(m, c *server.Object) bool {
							n, k := index(m), index(c)
							v := visits[n]
							visits[n]++
							nested := emit("interact:"+antiSpellSmallInt53BB00(n)+":"+antiSpellSmallInt53BB00(int64(v)), n, k)
							if nested {
								old := currentMissile
								currentMissile = objects[4]
								callRoot(objects[4])
								currentMissile = old
								return true
							}
							return i.InteractMask&(1<<uint(k)) != 0
						},
						Audio: func(m *server.Object) { emit("audio", index(m)) }, DelayedDelete: func(m *server.Object) { n := index(m); emit("delete:"+antiSpellSmallInt53BB00(n), n) },
						RandomInt: func(min, max int) int { emit("int", int64(min), int64(max)); return rng.IntClamp(min, max) },
						CreateSpark: func(pos types.Pointf, kind, lifetime int, velocity types.Pointf, z float32, m *server.Object) {
							emit("spark", int64(math.Float32bits(pos.X)), int64(math.Float32bits(pos.Y)), int64(uint32(kind)), int64(uint32(lifetime)), int64(math.Float32bits(velocity.X)), int64(math.Float32bits(velocity.Y)), int64(math.Float32bits(z)), index(m))
						},
					}
					randomFloat := func(min, max float32) float64 {
						stage := "float:" + antiSpellSmallInt53BB00(int64(floatVisit))
						floatVisit++
						emit(stage, int64(math.Float32bits(min)), int64(math.Float32bits(max)))
						return world.RandomFloat416030(min, max)
					}
					runtime.RandomFloatWide = randomFloat
					runtime.RandomFloat = func(min, max float32) float32 { return float32(randomFloat(min, max)) }
					oldRuntime := antiSpellProjectileRuntime53BB00
					antiSpellProjectileRuntime53BB00 = func() server.AntiSpellProjectileRuntime53BB00 { return runtime }
					defer func() { antiSpellProjectileRuntime53BB00 = oldRuntime }()
					if i.Kind == "root" {
						callRoot(objects[0])
					} else {
						antiSpellCProbe53BB00(1, objects[0], objects[i.Candidate])
					}
					if got := state(); got != row.antiSpellState53BB00 || !reflect.DeepEqual(events, row.Events) || mutated != row.Mutated {
						a, _ := json.Marshal(got)
						b, _ := json.Marshal(row.antiSpellState53BB00)
						t.Fatalf("native=%s original=%s events_equal=%t mutation=%t/%t", a, b, reflect.DeepEqual(events, row.Events), mutated, row.Mutated)
					}
					putPointer := func(w []byte, off uintptr, p unsafe.Pointer) {
						if unsafe.Sizeof(uintptr(0)) == 8 {
							binary.LittleEndian.PutUint64(w[off:], uint64(uintptr(p)))
						} else {
							binary.LittleEndian.PutUint32(w[off:], uint32(uintptr(p)))
						}
					}
					for n, o := range objects {
						want := bytes.Clone(before[n])
						offsets := []uintptr{unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.PosVec) + 4, unsafe.Offsetof(o.VelVec), unsafe.Offsetof(o.VelVec) + 4, unsafe.Offsetof(o.ForceVec), unsafe.Offsetof(o.ForceVec) + 4, unsafe.Offsetof(o.SpeedCur), unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.ObjSubClass), unsafe.Offsetof(o.Field32), unsafe.Offsetof(o.Field34)}
						for k, off := range offsets {
							binary.LittleEndian.PutUint32(want[32+off:], row.Objects[n][k])
						}
						putPointer(want, 32+unsafe.Offsetof(o.ObjOwner), unsafe.Pointer(pointer(row.Owners[n])))
						for k, m := range []int{0, 4} {
							if n == m {
								putPointer(want, 32+unsafe.Offsetof(o.UpdateData), unsafe.Pointer(updates[row.UpdateRefs[k]]))
							}
						}
						if !bytes.Equal(frames[n], want) {
							t.Fatalf("object%d full frame/canary mismatch", n)
						}
					}
					for n, u := range updates {
						want := bytes.Clone(ubefore[n])
						w := row.Updates[n]
						putPointer(want, 32+unsafe.Offsetof(u.Owner), unsafe.Pointer(pointer(w[0])))
						putPointer(want, 32+unsafe.Offsetof(u.Target), unsafe.Pointer(pointer(w[1])))
						for k, off := range []uintptr{unsafe.Offsetof(u.Field_8), unsafe.Offsetof(u.SpellID), unsafe.Offsetof(u.Field_16), unsafe.Offsetof(u.Field_20), unsafe.Offsetof(u.Field_24)} {
							binary.LittleEndian.PutUint32(want[32+off:], uint32(w[k+2]))
						}
						if !bytes.Equal(uframes[n], want) {
							t.Fatalf("update%d full frame/canary mismatch", n)
						}
					}
					want := bytes.Clone(sbefore)
					putPointer(want, 32+unsafe.Offsetof(selection.Target), unsafe.Pointer(pointer(row.Selected)))
					binary.LittleEndian.PutUint32(want[32+unsafe.Offsetof(selection.BestDistance):], row.Best)
					if !bytes.Equal(sframe, want) {
						t.Fatal("native shared result frame/canary mismatch")
					}
				})
			}
			want := 776
			if entry == "go-root" {
				want = 634
			}
			if seen != want {
				t.Fatal("incomplete original condition selection")
			}
		})
	}
}

func antiSpellSmallInt53BB00(n int64) string {
	if n < 0 {
		return "-1"
	}
	return string(rune('0' + n))
}
