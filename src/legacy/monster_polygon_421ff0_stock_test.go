package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type monsterPolygonInput421FF0 struct {
	Name, Mutation       string
	Current, Destination uint32
	Clazz                uint32
	Position, Previous   [2]uint32
	Leave, Enter, Active bool
	UnitNull             bool `json:"unit_null"`
	UpdateNull           bool `json:"update_null"`
}

type monsterPolygonOutcome421FF0 struct {
	Events             json.RawMessage
	Current, Alternate uint32
	CachedUnitUpdate   bool      `json:"cached_unit_update"`
	PolygonIDs         [3]uint32 `json:"polygon_ids"`
}

type monsterPolygonFixtureServer421FF0 struct {
	Server
	runtime MonsterPolygonRuntime421FF0
}

func (s monsterPolygonFixtureServer421FF0) MonsterPolygonEnterNative421FF0(unit *server.Object) {
	MonsterPolygonTransition421FF0(unit, s.runtime)
}

func TestMonsterPolygonFrozenOriginalC421FF0(t *testing.T) {
	content, err := os.ReadFile("testdata/monster_polygon_original_v1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if sum := fmt.Sprintf("%x", sha256.Sum256(content)); sum != "f6d451df9f094902020036f5a5d6aa01af265f24c6737f24870a5a0f8c63ab04" {
		t.Fatalf("stock monster polygon hash = %s", sum)
	}
	reader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var reference struct {
		Startup   string `json:"startup_control_word"`
		Parent    int    `json:"original_parent_bytes"`
		Converter int    `json:"original_converter_bytes"`
		Guards    bool   `json:"guards_verified"`
		Entries   map[string]int
		Rows      []struct {
			Input    monsterPolygonInput421FF0
			Observed monsterPolygonOutcome421FF0
		}
		Faults []struct {
			Input    monsterPolygonInput421FF0
			Observed struct {
				Address, At uint32
			}
		}
	}
	if err := json.NewDecoder(reader).Decode(&reference); err != nil {
		t.Fatal(err)
	}
	if reference.Startup != "0xe3f" || reference.Parent != 327 || reference.Converter != 17 || !reference.Guards ||
		reference.Entries["parent"] != 37 || reference.Entries["converter"] != 58 ||
		len(reference.Rows) != 35 || len(reference.Faults) != 2 {
		t.Fatal("unexpected original monster polygon proof")
	}
	for _, row := range reference.Rows {
		t.Run(row.Input.Name, func(t *testing.T) { compareMonsterPolygon421FF0(t, row.Input, row.Observed) })
	}
	for _, fault := range reference.Faults {
		t.Run(fault.Input.Name, func(t *testing.T) {
			old := GetServer
			GetServer = func() Server { return monsterPolygonFixtureServer421FF0{} }
			t.Cleanup(func() { GetServer = old })
			switch {
			case fault.Input.UnitNull && fault.Observed.Address == 0x2ec && fault.Observed.At == 0x421ffc:
				monsterPolygonFixture421FF0(nil)
			case fault.Input.UpdateNull && fault.Observed.Address == 0 && fault.Observed.At == 0x422012:
				unit := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster})
				monsterPolygonFixture421FF0(unit)
			default:
				t.Fatalf("unexpected original fault %+v", fault)
			}
		})
	}
}

func compareMonsterPolygon421FF0(t *testing.T, in monsterPolygonInput421FF0, want monsterPolygonOutcome421FF0) {
	t.Helper()
	f32 := math.Float32frombits
	update := game4ConsumerValue(t, server.MonsterUpdateData{Field0: in.Current})
	alternate := game4ConsumerValue(t, server.MonsterUpdateData{Field0: 0x137fff0})
	unit := game4ConsumerValue(t, server.Object{
		ObjClass: object.Class(in.Clazz), UpdateData: unsafe.Pointer(update),
		PosVec: types.Ptf(f32(in.Position[0]), f32(in.Position[1])), PrevPos: types.Ptf(f32(in.Previous[0]), f32(in.Previous[1])),
	})
	var records [5]*Nox_player_polygon_check_data
	recordID := make(map[*Nox_player_polygon_check_data]uint32, 3)
	for i := uint32(2); i <= 4; i++ {
		data := Nox_player_polygon_check_data{}
		data.Field_0[20] = i
		data.Field_0[21] = uint32(bool2int(in.Active))
		data.Field_0[29], data.Field_0[31] = ^uint32(0), ^uint32(0)
		if in.Enter {
			data.Field_0[29] = 200 + i
		}
		if in.Leave {
			data.Field_0[31] = 300 + i
		}
		records[i] = game4ConsumerValue(t, data)
		recordID[records[i]] = i
	}
	events := make([][]any, 0)
	event := func(args ...any) { events = append(events, args) }
	runtime := MonsterPolygonRuntime421FF0{
		Find: func(point [2]int32, previous uint32) *Nox_player_polygon_check_data {
			kind := "transition"
			if previous == monsterPolygonUninitialized421FF0 {
				kind, previous = "initial", 0
			}
			event(kind, point[0], point[1], previous)
			switch in.Mutation {
			case "lookup-current":
				update.Field0 = 4
			case "lookup-update":
				unit.UpdateData = unsafe.Pointer(alternate)
			}
			return records[in.Destination]
		},
		ByID: func(id uint32) *Nox_player_polygon_check_data { return records[id] },
		Callback: func(p *Nox_player_polygon_check_data, word int, value *server.Object, typ server.ScriptEventType) {
			if value != unit || recordID[p] == 0 {
				t.Fatalf("callback pointer %p / %p", value, p)
			}
			switch typ {
			case server.NoxEventPolygonEnterZZZ:
				if word != 30 {
					t.Fatalf("leave callback word = %d", word)
				}
				event("transition-leave", recordID[p], p.Field_0[word+1], update.Field0)
				switch in.Mutation {
				case "leave-disable-enter":
					records[in.Destination].Field_0[29] = ^uint32(0)
				case "leave-replace-id":
					records[in.Destination].Field_0[20] = 37
				}
			case server.NoxEventPolygonEnterYYY:
				if word != 28 {
					t.Fatalf("enter callback word = %d", word)
				}
				event("transition-enter", recordID[p], p.Field_0[word+1], update.Field0)
				if in.Mutation == "enter-replace-id" {
					records[in.Destination].Field_0[20] = 37
				}
			case server.NoxEventPolygonEnterXXX:
				if word != 30 {
					t.Fatalf("exit callback word = %d", word)
				}
				event("exit", recordID[p], p.Field_0[word+1], update.Field0)
				if in.Mutation == "exit-replace-update" {
					unit.UpdateData = unsafe.Pointer(alternate)
				}
			default:
				t.Fatalf("unexpected callback type %v", typ)
			}
		},
	}
	old := GetServer
	GetServer = func() Server { return monsterPolygonFixtureServer421FF0{runtime: runtime} }
	t.Cleanup(func() { GetServer = old })
	monsterPolygonFixture421FF0(unit)
	gotEvents, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	ids := [3]uint32{records[2].Field_0[20], records[3].Field_0[20], records[4].Field_0[20]}
	if update.Field0 != want.Current || alternate.Field0 != want.Alternate || ids != want.PolygonIDs ||
		(unit.UpdateData == unsafe.Pointer(update)) != want.CachedUnitUpdate || !bytes.Equal(gotEvents, want.Events) {
		t.Fatalf("current/alternate/ids = %d/%d/%v, want %d/%d/%v\ntrace: %s\nstock: %s",
			update.Field0, alternate.Field0, ids, want.Current, want.Alternate, want.PolygonIDs, gotEvents, want.Events)
	}
}
