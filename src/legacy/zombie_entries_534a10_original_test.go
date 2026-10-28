package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type zombieEntryState534A10 struct {
	Cache [3]uint32 `json:"cache"`
	Type  uint16    `json:"type"`
	Class uint32    `json:"cls"`
	Flags uint32    `json:"flags"`
	Done  byte      `json:"done"`
}
type zombieEntryEvent534A10 struct {
	Service string `json:"service"`
	zombieEntryState534A10
}
type zombieEntryInput534A10 struct {
	Kind           string    `json:"kind"`
	Name           string    `json:"name"`
	Cache          [3]uint32 `json:"cache"`
	Lookup         [3]uint32 `json:"lookup"`
	Type           uint16    `json:"type"`
	Class          uint32    `json:"cls"`
	Flags          uint32    `json:"flags"`
	Done           byte      `json:"done"`
	Action         uint32    `json:"action"`
	Pop            uint32    `json:"pop"`
	Push           [2]uint32 `json:"push"`
	MutationEvent  string    `json:"mutation_event"`
	MutationAction string    `json:"mutation_action"`
}

func TestZombieEntriesOriginalNativeC534A10(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/zombie_entries_534a10_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Contexts    int    `json:"contexts"`
		OriginalSHA string `json:"original_sha256"`
		Rows        []struct {
			Input     zombieEntryInput534A10   `json:"input"`
			Result    uint32                   `json:"result"`
			Events    []zombieEntryEvent534A10 `json:"events"`
			Preserved bool                     `json:"nonoutput_bytes_preserved"`
			zombieEntryState534A10
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 375 || len(f.Rows) != 375 || f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" {
		t.Fatal("original entry provenance")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			i := row.Input
			if !row.Preserved {
				t.Fatal("original canaries")
			}
			frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{})+256)
			base := unsafe.Pointer(&frame[32])
			delta := (uintptr(0x80) - uintptr(base)) & 255
			unit := (*server.Object)(unsafe.Add(base, delta))
			start := uintptr(32) + delta
			*unit = server.Object{TypeInd: i.Type, ObjClass: object.Class(i.Class), ObjFlags: object.Flags(i.Flags), Field37: 0x55667788, Field38: 0x11223344}
			udFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.MonsterUpdateData{}))
			ud := (*server.MonsterUpdateData)(unsafe.Pointer(&udFrame[32]))
			*ud = server.MonsterUpdateData{}
			ud.Field120_3 = i.Done
			unit.UpdateData = unsafe.Pointer(ud)
			cacheFrame := radialCFrameOriginal(t, 12)
			cache := (*[3]uint32)(unsafe.Pointer(&cacheFrame[32]))
			*cache = i.Cache
			before, udBefore, cacheBefore := bytes.Clone(frame), bytes.Clone(udFrame), bytes.Clone(cacheFrame)
			state := func() zombieEntryState534A10 {
				return zombieEntryState534A10{*cache, unit.TypeInd, uint32(unit.ObjClass), uint32(unit.ObjFlags), ud.Field120_3}
			}
			events := []zombieEntryEvent534A10{}
			mutated := false
			pushes := 0
			mutate := func(stage string) {
				if mutated || stage != i.MutationEvent {
					return
				}
				mutated = true
				switch i.MutationAction {
				case "type":
					unit.TypeInd = 201
				case "flags":
					unit.ObjFlags = object.Flags(0xfedcba98)
				case "class":
					unit.ObjClass = object.Class(0x400000)
				case "done":
					ud.Field120_3 = 0
				case "plant-cache":
					cache[0] = 0x10064
				case "zombie-cache":
					cache[1] = 0x10064
				case "vile-cache":
					cache[2] = 201
				case "all-cache":
					*cache = [3]uint32{0, 201, 100}
				default:
					t.Fatal("unknown declared mutation")
				}
			}
			emit := func(stage string) { events = append(events, zombieEntryEvent534A10{stage, state()}); mutate(stage) }
			view := server.TypePredicateRuntime534A10{Read: func(n int) uint32 { return cache[n] }, Write: func(n int, v uint32) { cache[n] = v }, Lookup: func(name string) uint32 {
				n, stage := 0, "plant"
				switch name {
				case "CarnivorousPlant":
				case "Zombie":
					n, stage = 1, "zombie"
				case "VileZombie":
					n, stage = 2, "vile"
				default:
					t.Fatal("unknown original lookup")
				}
				emit(stage)
				return i.Lookup[n]
			}}
			oldTypes, oldPop, oldRaise := zombieEntryTypes534A10, zombieEntryPop534A90, zombieEntryRaise534AB0
			defer func() {
				zombieEntryTypes534A10, zombieEntryPop534A90, zombieEntryRaise534AB0 = oldTypes, oldPop, oldRaise
			}()
			zombieEntryTypes534A10 = func() server.TypePredicateRuntime534A10 { return view }
			zombieEntryPop534A90 = func(o *server.Object) uint32 {
				if o != unit {
					t.Fatal("pop pointer")
				}
				emit("pop")
				return i.Pop
			}
			zombieEntryRaise534AB0 = func() server.MonsterRaiseZombieRawRuntime534AB0 {
				return server.MonsterRaiseZombieRawRuntime534AB0{
					IsZombie: func(o *server.Object) uint32 { return server.UnitIsZombie534A40(o, view) },
					CurrentAction: func(o *server.Object) uint32 {
						if o != unit {
							t.Fatal("action pointer")
						}
						emit("action")
						return i.Action
					},
					Pop: zombieEntryPop534A90,
					Push: func(o *server.Object, action uint32) uintptr {
						if o != unit || pushes >= 2 {
							t.Fatal("push pointer/order")
						}
						stage := "push-61"
						if action == 35 {
							stage = "push-35"
						} else if action != 61 {
							t.Fatal("push action")
						}
						emit(stage)
						value := i.Push[pushes]
						pushes++
						return uintptr(value)
					},
					Audio: func(id uint32, o *server.Object) {
						if id != 469 || o != unit {
							t.Fatal("audio arguments")
						}
						emit("audio")
					},
					Health: func(o *server.Object) {
						if o != unit {
							t.Fatal("health pointer")
						}
						emit("health")
					},
				}
			}
			kind := map[string]int{"plant": 0, "zombie": 1, "getup": 2, "raise": 3}[i.Kind]
			result := zombieEntryCProbe534A10(kind, unit)
			if result != row.Result || !reflect.DeepEqual(events, row.Events) || state() != row.zombieEntryState534A10 {
				a, _ := json.Marshal(struct {
					Result uint32
					Events []zombieEntryEvent534A10
					State  zombieEntryState534A10
				}{result, events, state()})
				b, _ := json.Marshal(row)
				t.Fatalf("native=%s original=%s", a, b)
			}
			binary.LittleEndian.PutUint16(before[start+unsafe.Offsetof(unit.TypeInd):], row.Type)
			binary.LittleEndian.PutUint32(before[start+unsafe.Offsetof(unit.ObjClass):], row.Class)
			binary.LittleEndian.PutUint32(before[start+unsafe.Offsetof(unit.ObjFlags):], row.Flags)
			udBefore[32+unsafe.Offsetof(ud.Field120_3)] = row.Done
			for n, v := range row.Cache {
				binary.LittleEndian.PutUint32(cacheBefore[32+4*n:], v)
			}
			if !bytes.Equal(frame, before) || !bytes.Equal(udFrame, udBefore) || !bytes.Equal(cacheFrame, cacheBefore) {
				t.Fatal("complete native object/update/cache record or canary changed")
			}
		})
	}
}
