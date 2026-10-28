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
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type spellFlightState53B940 struct {
	Objects [4][11]uint32 `json:"objects"`
	Updates [2][7]int64   `json:"updates"`
	Update  int           `json:"update"`
	Frame   uint32        `json:"frame"`
	FPS     uint32        `json:"fps"`
}
type spellFlightEvent53B940 struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	spellFlightState53B940
}
type spellFlightInput53B940 struct {
	Name string `json:"name"`
	spellFlightState53B940
	Lifetime       uint64 `json:"lifetime"`
	SpellFlags     uint32 `json:"spell_flags"`
	SearchResult   int    `json:"search_result"`
	MutationEvent  string `json:"mutation_event"`
	MutationAction string `json:"mutation_action"`
}

func TestSpellFlightOriginalNative53B940(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/spell_flight_53b940_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		OriginalSHA string `json:"original_sha256"`
		Contexts    int    `json:"contexts"`
		Executions  int    `json:"root_executions"`
		Rows        []struct {
			Mode      string                   `json:"mode"`
			Input     spellFlightInput53B940   `json:"input"`
			Events    []spellFlightEvent53B940 `json:"events"`
			Mutated   bool                     `json:"mutated"`
			Preserved bool                     `json:"nonoutput_bytes_preserved"`
			spellFlightState53B940
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Contexts != 681 || f.Executions != 1362 || len(f.Rows) != 1362 {
		t.Fatal("original spell-flight provenance")
	}
	for _, entry := range []string{"native-go", "actual-c"} {
		t.Run(entry, func(t *testing.T) {
			seen := 0
			for _, row := range f.Rows {
				if row.Mode != "actual-startup" {
					continue
				}
				seen++
				t.Run(row.Input.Name, func(t *testing.T) {
					i := row.Input
					if !row.Preserved {
						t.Fatal("original frame canaries")
					}
					var objects [4]*server.Object
					var frames, before [4][]byte
					for n, words := range i.Objects {
						frames[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
						objects[n] = (*server.Object)(unsafe.Pointer(&frames[n][32]))
						*objects[n] = server.Object{
							PosVec:   types.Ptf(math.Float32frombits(words[0]), math.Float32frombits(words[1])),
							VelVec:   types.Ptf(math.Float32frombits(words[2]), math.Float32frombits(words[3])),
							ForceVec: types.Ptf(math.Float32frombits(words[4]), math.Float32frombits(words[5])),
							Float28:  math.Float32frombits(words[6]), Field32: words[7], Field34: words[8],
							SpeedCur: math.Float32frombits(words[9]), ObjFlags: object.Flags(words[10]),
						}
					}
					pointer := func(index int64) *server.Object {
						if index < 0 {
							return nil
						}
						if index >= int64(len(objects)) {
							t.Fatal("foreign original reference")
						}
						return objects[index]
					}
					index := func(obj *server.Object) int64 {
						if obj == nil {
							return -1
						}
						for n, candidate := range objects {
							if candidate == obj {
								return int64(n)
							}
						}
						t.Fatal("foreign native pointer")
						return -2
					}
					var updates [2]*server.SpellProjectileUpdateData
					var updateFrames, updateBefore [2][]byte
					for n, words := range i.Updates {
						updateFrames[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.SpellProjectileUpdateData{}))
						updates[n] = (*server.SpellProjectileUpdateData)(unsafe.Pointer(&updateFrames[n][32]))
						*updates[n] = server.SpellProjectileUpdateData{Field0: pointer(words[0]), Target: pointer(words[1]), Field8: pointer(words[2]), Spell12: uint32(words[3]), Level16: uint32(words[4]), Field20: uint32(words[5]), Field24: uint32(words[6])}
						updateBefore[n] = bytes.Clone(updateFrames[n])
					}
					missile := objects[0]
					missile.UpdateData = unsafe.Pointer(updates[0])
					for n := range frames {
						before[n] = bytes.Clone(frames[n])
					}
					frame, fps := i.Frame, i.FPS
					state := func() spellFlightState53B940 {
						out := spellFlightState53B940{Frame: frame, FPS: fps, Update: -1}
						for n, obj := range objects {
							out.Objects[n] = [11]uint32{math.Float32bits(obj.PosVec.X), math.Float32bits(obj.PosVec.Y), math.Float32bits(obj.VelVec.X), math.Float32bits(obj.VelVec.Y), math.Float32bits(obj.ForceVec.X), math.Float32bits(obj.ForceVec.Y), math.Float32bits(obj.Float28), obj.Field32, obj.Field34, math.Float32bits(obj.SpeedCur), uint32(obj.ObjFlags)}
						}
						for n, update := range updates {
							out.Updates[n] = [7]int64{index(update.Field0), index(update.Target), index(update.Field8), int64(update.Spell12), int64(update.Level16), int64(update.Field20), int64(update.Field24)}
							if missile.UpdateData == unsafe.Pointer(update) {
								out.Update = n
							}
						}
						return out
					}
					events := []spellFlightEvent53B940{}
					mutated := false
					mutate := func(stage string) {
						if mutated || stage != i.MutationEvent {
							return
						}
						mutated = true
						switch i.MutationAction {
						case "replace-update":
							missile.UpdateData = unsafe.Pointer(updates[1])
						case "owner":
							updates[0].Field0 = objects[3]
						case "target":
							updates[0].Target = objects[3]
						case "source":
							updates[0].Field8 = objects[3]
						case "spell":
							updates[0].Spell12 = 0xfedcba98
						case "position":
							missile.PosVec = types.Ptf(-20.3, 300.1)
						case "target-position":
							objects[3].PosVec = types.Ptf(321.25, -200.3)
						case "speed":
							missile.SpeedCur = 1.75
						case "velocity":
							missile.VelVec = types.Ptf(-5.1, 7.3)
						case "time":
							frame, fps = 32, 0xffffffff
						case "created":
							missile.Field32 = 31
						case "searched":
							missile.Field34 = 31
						case "flags":
							for _, obj := range objects {
								obj.ObjFlags = 0x8020
							}
						default:
							t.Fatal("unknown declared mutation")
						}
					}
					emit := func(service string, args ...int64) {
						if args == nil {
							args = []int64{}
						}
						events = append(events, spellFlightEvent53B940{service, args, state()})
						mutate(service)
					}
					runtime := server.SpellProjectileUpdateRuntime53B940{
						Frame: func() uint32 { return frame }, TickRate: func() uint32 { return fps },
						Lifetime: func(key string) float64 {
							var targeted int64
							if key == "TargetedSpellLifetime" {
								targeted = 1
							} else if key != "UnTargetedSpellLifetime" {
								t.Fatal("foreign balance key")
							}
							emit("lifetime", targeted)
							return math.Float64frombits(i.Lifetime)
						},
						SearchTarget: func(obj, owner *server.Object, spell uint32) *server.Object {
							if obj != missile {
								t.Fatal("search missile pointer")
							}
							emit("flags", int64(spell))
							emit("search", int64(i.SpellFlags), index(owner))
							return pointer(int64(i.SearchResult))
						},
						SendPointFX: func(pos types.Pointf) {
							emit("point-fx", int64(math.Float32bits(pos.X)), int64(math.Float32bits(pos.Y)))
						},
						Expire: func(obj *server.Object) {
							if obj != missile {
								t.Fatal("expiry pointer")
							}
							emit("expire")
						},
					}
					if entry == "native-go" {
						server.SpellProjectileUpdate53B940(missile, runtime)
					} else {
						old := spellProjectileUpdateRuntime53B940
						defer func() { spellProjectileUpdateRuntime53B940 = old }()
						spellProjectileUpdateRuntime53B940 = func() server.SpellProjectileUpdateRuntime53B940 { return runtime }
						spellFlightCProbe53B940(missile)
					}
					if got := state(); got != row.spellFlightState53B940 || !reflect.DeepEqual(events, row.Events) || mutated != row.Mutated {
						a, _ := json.Marshal(got)
						b, _ := json.Marshal(row.spellFlightState53B940)
						t.Fatalf("native=%s original=%s events_equal=%t mutation=%t/%t", a, b, reflect.DeepEqual(events, row.Events), mutated, row.Mutated)
					}
					// Compare whole C-owned frames, allowing only original stores and
					// the explicitly declared service mutations at native field offsets.
					for n, obj := range objects {
						want := bytes.Clone(before[n])
						words := row.Objects[n]
						offsets := []uintptr{unsafe.Offsetof(obj.PosVec), unsafe.Offsetof(obj.PosVec) + 4, unsafe.Offsetof(obj.VelVec), unsafe.Offsetof(obj.VelVec) + 4, unsafe.Offsetof(obj.ForceVec), unsafe.Offsetof(obj.ForceVec) + 4, unsafe.Offsetof(obj.Float28), unsafe.Offsetof(obj.Field32), unsafe.Offsetof(obj.Field34), unsafe.Offsetof(obj.SpeedCur), unsafe.Offsetof(obj.ObjFlags)}
						for k, off := range offsets {
							binary.LittleEndian.PutUint32(want[32+off:], words[k])
						}
						if n == 0 {
							off := 32 + unsafe.Offsetof(obj.UpdateData)
							p := uintptr(unsafe.Pointer(updates[row.Update]))
							if unsafe.Sizeof(p) == 8 {
								binary.LittleEndian.PutUint64(want[off:], uint64(p))
							} else {
								binary.LittleEndian.PutUint32(want[off:], uint32(p))
							}
						}
						if !bytes.Equal(frames[n], want) {
							t.Fatalf("object %d frame/canary changed outside original stores", n)
						}
					}
					for n, update := range updates {
						want := bytes.Clone(updateBefore[n])
						words := row.Updates[n]
						for k, off := range []uintptr{unsafe.Offsetof(update.Field0), unsafe.Offsetof(update.Target), unsafe.Offsetof(update.Field8)} {
							p := uintptr(unsafe.Pointer(pointer(words[k])))
							if unsafe.Sizeof(p) == 8 {
								binary.LittleEndian.PutUint64(want[32+off:], uint64(p))
							} else {
								binary.LittleEndian.PutUint32(want[32+off:], uint32(p))
							}
						}
						for k, off := range []uintptr{unsafe.Offsetof(update.Spell12), unsafe.Offsetof(update.Level16), unsafe.Offsetof(update.Field20), unsafe.Offsetof(update.Field24)} {
							binary.LittleEndian.PutUint32(want[32+off:], uint32(words[k+3]))
						}
						if !bytes.Equal(updateFrames[n], want) {
							t.Fatalf("update %d frame/canary changed outside original stores", n)
						}
					}
				})
			}
			if seen != 681 {
				t.Fatal("missing original startup conditions")
			}
		})
	}
}
