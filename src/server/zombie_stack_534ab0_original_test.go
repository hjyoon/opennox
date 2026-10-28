package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type zombieStackState534AB0 struct {
	Cache      [3]uint32     `json:"cache"`
	Type       uint16        `json:"type"`
	Class      uint32        `json:"cls"`
	Flags      uint32        `json:"flags"`
	Done       byte          `json:"done"`
	Index      int8          `json:"stack_index"`
	Stack      [24][6]uint32 `json:"stack"`
	ResetWords [6]uint32     `json:"reset_words"`
	ResetBytes [3]byte       `json:"reset_bytes"`
	Changed    uint32        `json:"changed"`
}
type zombieStackEvent534AB0 struct {
	Service string `json:"service"`
	zombieStackState534AB0
}

func TestZombieStackOriginalNative534AB0(t *testing.T) {
	blob, err := os.ReadFile("testdata/zombie_stack_534ab0_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Contexts int `json:"contexts"`
		Rows     []struct {
			Input struct {
				Kind      string    `json:"kind"`
				Name      string    `json:"name"`
				Type      uint16    `json:"type"`
				Class     uint32    `json:"cls"`
				Flags     uint32    `json:"flags"`
				Done      byte      `json:"done"`
				Cache     [3]uint32 `json:"cache"`
				Index     int8      `json:"stack_index"`
				Previous  uint32    `json:"previous"`
				Base      uint32    `json:"base"`
				Frame     uint32    `json:"frame"`
				Predicate bool      `json:"predicate_service"`
			} `json:"input"`
			Events    []zombieStackEvent534AB0 `json:"events"`
			Preserved bool                     `json:"nonoutput_bytes_preserved"`
			zombieStackState534AB0
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 69 || len(f.Rows) != 69 {
		t.Fatal("original stack reference")
	}
	frame := func(t *testing.T, size uintptr) []byte {
		t.Helper()
		p, free := alloc.Make([]byte(nil), int(size)+64)
		t.Cleanup(free)
		for n := range p {
			p[n] = 0xa5
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&p[32])) <= 0xffffffff {
			t.Fatal("C frame below 4 GiB")
		}
		return p
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			i := row.Input
			if !row.Preserved {
				t.Fatal("original record preservation")
			}
			srv := monsterZombieTestServer544CE0(t)
			srv.SetFrame(i.Frame)
			objectFrame := frame(t, unsafe.Sizeof(Object{}))
			unit := (*Object)(unsafe.Pointer(&objectFrame[32]))
			*unit = Object{TypeInd: i.Type, ObjClass: object.Class(i.Class), ObjFlags: object.Flags(i.Flags), serverHandle: srv.handle, Field37: 0x55667788, Field38: 0x11223344}
			dataFrame := frame(t, unsafe.Sizeof(MonsterUpdateData{}))
			data := (*MonsterUpdateData)(unsafe.Pointer(&dataFrame[32]))
			*data = MonsterUpdateData{AIStackInd: i.Index, Field120_3: i.Done}
			unit.UpdateData = unsafe.Pointer(data)
			data.AIStack[0].Action = i.Base
			if i.Index > 0 {
				data.AIStack[int(i.Index)-1].Action = i.Previous
			}
			data.AIStack[i.Index].Action = 31
			cache := i.Cache
			state := func() zombieStackState534AB0 {
				out := zombieStackState534AB0{Cache: cache, Type: unit.TypeInd, Class: uint32(unit.ObjClass), Flags: uint32(unit.ObjFlags), Done: data.Field120_3, Index: data.AIStackInd, ResetWords: [6]uint32{data.Field2, data.Field67, data.Field74, data.Field91, data.Field124, data.Field137}, ResetBytes: [3]byte{data.Field120_1, data.Field120_2, data.Field120_3}}
				if srv.AI.StackChanged {
					out.Changed = 1
				}
				for n, item := range data.AIStack {
					out.Stack[n] = [6]uint32{item.Action, uint32(item.Args[0]), uint32(item.Args[1]), uint32(item.Args[2]), uint32(item.Args[3]), item.Field5}
				}
				return out
			}
			events := []zombieStackEvent534AB0{}
			expectedEvents := []zombieStackEvent534AB0{}
			for _, event := range row.Events {
				if event.Service == "audio" || event.Service == "health" || event.Service == "is-zombie" {
					expectedEvents = append(expectedEvents, event)
				}
			}
			view := TypePredicateRuntime534A10{Read: func(n int) uint32 { return cache[n] }, Write: func(n int, v uint32) { cache[n] = v }, Lookup: func(string) uint32 { t.Fatal("warm original cache lookup"); return 0 }}
			beforeObject, beforeData := bytes.Clone(objectFrame), bytes.Clone(dataFrame)
			if i.Kind == "getup" {
				srv.MonsterActionGetUp534A90(unit)
			} else {
				srv.MonsterRaiseZombie534AB0(unit, MonsterRaiseZombieRuntime534AB0{
					IsZombie: func(o *Object) bool {
						if o != unit {
							t.Fatal("predicate pointer")
						}
						if i.Predicate {
							events = append(events, zombieStackEvent534AB0{"is-zombie", state()})
							return true
						}
						return UnitIsZombie534A40(o, view) != 0
					},
					AudioEvent: func(id uint32, o *Object) {
						if id != 469 || o != unit {
							t.Fatal("audio identity")
						}
						events = append(events, zombieStackEvent534AB0{"audio", state()})
					},
					SetHealthToMax: func(o *Object) {
						if o != unit {
							t.Fatal("health identity")
						}
						events = append(events, zombieStackEvent534AB0{"health", state()})
					},
				})
			}
			if state() != row.zombieStackState534AB0 || !reflect.DeepEqual(events, expectedEvents) {
				a, _ := json.Marshal(state())
				b, _ := json.Marshal(row.zombieStackState534AB0)
				t.Fatalf("native=%s original=%s callbacks=%v expected=%v", a, b, events, expectedEvents)
			}
			// Preserve every native object/update byte outside the original write set.
			binary.LittleEndian.PutUint32(beforeObject[32+unsafe.Offsetof(unit.ObjFlags):], row.Flags)
			beforeData[32+unsafe.Offsetof(data.AIStackInd)] = byte(row.Index)
			for offset, value := range map[uintptr]uint32{unsafe.Offsetof(data.Field2): row.ResetWords[0], unsafe.Offsetof(data.Field67): row.ResetWords[1], unsafe.Offsetof(data.Field74): row.ResetWords[2], unsafe.Offsetof(data.Field91): row.ResetWords[3], unsafe.Offsetof(data.Field124): row.ResetWords[4], unsafe.Offsetof(data.Field137): row.ResetWords[5]} {
				binary.LittleEndian.PutUint32(beforeData[32+offset:], value)
			}
			for n, value := range row.ResetBytes {
				beforeData[32+unsafe.Offsetof(data.Field120_1)+uintptr(n)] = value
			}
			for n, values := range row.Stack {
				at := uintptr(32) + unsafe.Offsetof(data.AIStack) + uintptr(n)*unsafe.Sizeof(AIStackItem{})
				binary.LittleEndian.PutUint32(beforeData[at:], values[0])
				binary.LittleEndian.PutUint32(beforeData[at+unsafe.Offsetof(AIStackItem{}.Field5):], values[5])
				for k := 0; k < 4; k++ {
					offset := at + unsafe.Offsetof(AIStackItem{}.Args) + uintptr(k)*unsafe.Sizeof(uintptr(0))
					if unsafe.Sizeof(uintptr(0)) == 8 {
						binary.LittleEndian.PutUint64(beforeData[offset:], uint64(values[k+1]))
					} else {
						binary.LittleEndian.PutUint32(beforeData[offset:], values[k+1])
					}
				}
			}
			if !bytes.Equal(objectFrame, beforeObject) || !bytes.Equal(dataFrame, beforeData) {
				t.Fatal("native object/update record or canary changed")
			}
		})
	}
}
