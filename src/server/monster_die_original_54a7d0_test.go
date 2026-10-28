package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"
)

type monsterDieOriginalEvent54A7D0 struct {
	Service   string
	Flags     uint32
	Minimum   int
	Maximum   int
	Unit      bool
	Type      string
	Modifiers [4]string
	Charge    uint8
}

type monsterDieOriginalRow54A7D0 struct {
	Input struct {
		Name           string
		Kind           MonsterDieCallbackKind54A7D0
		Roll           int32
		Unit           bool
		RandomMutation bool `json:"random_mutation"`
		DropMutation   bool `json:"drop_mutation"`
	}
	Result int32
	Flags  uint32
	Events []monsterDieOriginalEvent54A7D0
}

func TestMonsterDieOriginalGo54A7D0(t *testing.T) {
	b, err := os.ReadFile("testdata/monster_die_original_54a7d0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []monsterDieOriginalRow54A7D0
	}
	if err := json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 550 {
		t.Fatal("original fixture context count")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			// Aligned Object storage, with complete nonoutput bytes and canaries
			// checked after the random/drop services finish their declared writes.
			words := make([]uint64, (unsafe.Sizeof(Object{})+64+7)/8)
			frame := unsafe.Slice((*byte)(unsafe.Pointer(&words[0])), len(words)*8)
			for n := range frame {
				frame[n] = byte(n*37 + 11)
			}
			unit := (*Object)(unsafe.Pointer(&frame[32]))
			unit.ObjFlags = 0x12345678
			want := append([]byte(nil), frame...)
			argument := unit
			if !row.Input.Unit {
				argument = nil
			}
			var events []monsterDieOriginalEvent54A7D0
			result := MonsterDieCallbackRaw54A7D0(argument, row.Input.Kind, MonsterDieCallbackRawRuntime54A7D0{
				RandomInt: func(lo, hi int) int {
					events = append(events, monsterDieOriginalEvent54A7D0{Service: "random", Flags: uint32(unit.ObjFlags), Minimum: lo, Maximum: hi})
					if row.Input.RandomMutation {
						unit.ObjFlags = 0x89abcdef
					}
					return int(row.Input.Roll)
				},
				DropItem: func(got *Object, drop MonsterDieDrop54A390) {
					if got != argument {
						t.Fatal("drop object pointer identity")
					}
					events = append(events, monsterDieOriginalEvent54A7D0{Service: "drop", Flags: uint32(unit.ObjFlags), Unit: got != nil, Type: drop.TypeID, Modifiers: drop.Modifiers, Charge: drop.Charge})
					if row.Input.DropMutation {
						unit.ObjFlags = 0x76543210
					}
				},
			})
			if result != row.Result || uint32(unit.ObjFlags) != row.Flags || !reflect.DeepEqual(events, row.Events) {
				t.Fatalf("result/flags/events = %d/%08x/%+v, want %d/%08x/%+v", result, uint32(unit.ObjFlags), events, row.Result, row.Flags, row.Events)
			}
			binary.LittleEndian.PutUint32(want[32+unsafe.Offsetof(unit.ObjFlags):], row.Flags)
			if !bytes.Equal(frame, want) {
				t.Fatal("complete object frame or canary changed")
			}
		})
	}
}
