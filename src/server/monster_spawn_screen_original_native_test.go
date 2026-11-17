package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

//go:embed testdata/monster_spawn_screen_original_v1.json.gz
var monsterSpawnScreenCPU50D960 []byte

//go:embed testdata/monster_spawn_screen_record_expectations_v1.json.gz
var monsterSpawnScreenRecordExpected50D960 []byte

type monsterSpawnScreenInput50D960 struct {
	Name    string
	Order   []int
	Players []struct {
		Index         *uint8
		Active        *uint32
		Width, Height *uint16
		Pos           [2]float32
		QuestExit     bool `json:"quest_exit"`
	}
	Limit          *float64
	Positions      *[4][2]float32
	Classes, Flags *[4]uint32
	Point          [2]float32
	Distances      [][]float64
	TraceResults   []int `json:"trace_results"`
	ScratchFailAt  int   `json:"scratch_fail_at"`
	Enumeration    []int
	Zombie         int
	VisibleCount   uint32 `json:"visible_count"`
	Frame          uint32
	Rate           *uint32
}

type monsterSpawnScreenRecordExpected struct {
	Role         int
	AverageBits  uint32 `json:"average_bits"`
	Mask         uint32
	DistanceBits [32]uint32 `json:"distance_bits"`
}

func monsterSpawnDecodeFixture50D960(t *testing.T, packed []byte, packedSHA, rawSHA string, out any) {
	t.Helper()
	if fmt.Sprintf("%x", sha256.Sum256(packed)) != packedSHA {
		t.Fatal("CPU fixture compressed digest changed")
	}
	reader, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err = reader.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != rawSHA {
		t.Fatal("CPU fixture raw digest changed")
	}
	if err = json.Unmarshal(raw, out); err != nil {
		t.Fatal(err)
	}
}

func TestMonsterSpawnScreenOriginalNative50D960(t *testing.T) {
	var cases []struct {
		Input        monsterSpawnScreenInput50D960
		Observations []struct {
			Entry       string
			Arguments   []uint32
			ReturnValue *uint32 `json:"return_value"`
			GlobalsHex  string  `json:"globals_hex"`
			Services    []any
		}
	}
	monsterSpawnDecodeFixture50D960(t, monsterSpawnScreenCPU50D960, "4fd4f276432f487bb2fa0747aef50ac174fb8432fb618b0d4d7317c15cbcbc09", "d082eb37cb5b78bd89e84bbdf0fc1d381369ad9ddb9576fcf2a52c11bd652443", &cases)
	var recordCases []struct {
		Name    string
		Records []monsterSpawnScreenRecordExpected `json:"released_record_values"`
	}
	monsterSpawnDecodeFixture50D960(t, monsterSpawnScreenRecordExpected50D960, "8d9aff431cd1a8c60a14f7f8745ad074635c96fdd39fa133ffae03cade928b4b", "d204eb53f92260948e0957952524204bd6fe3ac4428a3906311c64723f1b4685", &recordCases)
	if len(cases) != 169 || len(recordCases) != 169 {
		t.Fatal("original input count changed")
	}
	for row, fixture := range cases {
		t.Run(fixture.Input.Name, func(t *testing.T) {
			c := fixture.Input
			if recordCases[row].Name != c.Name {
				t.Fatal("derived record identity changed")
			}
			s := new(Server)
			defer s.MonsterSpawnFree50D820()
			s.MonsterSpawnInit50D780()
			rate := uint32(30)
			if c.Rate != nil {
				rate = *c.Rate
			}
			s.SetTickRate(rate)
			s.SetFrame(c.Frame)
			generator, generatorData := monsterSpawnTestGenerator50D780(254)
			var objects [4]*Object
			var updates [4]*MonsterUpdateData
			addresses := map[*Object]uint32{generator: 0x24001000}
			positions := [4][2]float32{{0, 0}, {10, 0}, {20, 0}, {30, 0}}
			if c.Positions != nil {
				positions = *c.Positions
			}
			for i := range objects {
				objects[i], updates[i] = monsterSpawnTestMonster50D780(0)
				objects[i].PosVec = types.Ptf(positions[i][0], positions[i][1])
				if c.Classes != nil {
					objects[i].ObjClass = object.Class((*c.Classes)[i])
				}
				if c.Flags != nil {
					objects[i].ObjFlags = object.Flags((*c.Flags)[i])
				}
				addresses[objects[i]] = 0x24002000 + uint32(i)*0x400
			}
			for i := len(c.Order) - 1; i >= 0; i-- {
				if !s.MonsterSpawnRegister50E030(generator, objects[c.Order[i]], MonsterSpawnRegisterRuntime50E030{}) {
					t.Fatal("fixture registration failed")
				}
			}
			generatorData.ActiveCount = 254
			players := make([]*Object, len(c.Players))
			playerIndices := make(map[*Object]int)
			for i, p := range c.Players {
				info := &Player{PlayerInd: uint8(i), Field4792: 1, Field10: 300, Field12: 200}
				if p.Index != nil {
					info.PlayerInd = *p.Index
				}
				if p.Active != nil {
					info.Field4792 = *p.Active
				}
				if p.Width != nil {
					info.Field10 = *p.Width
				}
				if p.Height != nil {
					info.Field12 = *p.Height
				}
				update := &PlayerUpdateData{Player: info}
				if p.QuestExit {
					update.QuestExit = new(Object)
				}
				unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), PosVec: types.Ptf(p.Pos[0], p.Pos[1])}
				players[i] = unit
				playerIndices[unit] = i
				addresses[unit] = 0x24004000 + uint32(i)*0x400
			}
			// These are complete native Go owners; their pointer-bearing bytes must
			// remain unchanged during the read-only query/cull boundary providers.
			owners := make(map[unsafe.Pointer][]byte)
			keep := func(pointer unsafe.Pointer, size uintptr) {
				owners[pointer] = bytes.Clone(unsafe.Slice((*byte)(pointer), int(size)))
			}
			keep(unsafe.Pointer(generator), unsafe.Sizeof(*generator))
			keep(unsafe.Pointer(generatorData), unsafe.Sizeof(*generatorData))
			for i, obj := range objects {
				keep(unsafe.Pointer(obj), unsafe.Sizeof(*obj))
				keep(unsafe.Pointer(updates[i]), unsafe.Sizeof(*updates[i]))
			}
			for _, unit := range players {
				keep(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
				keep(unit.UpdateData, unsafe.Sizeof(PlayerUpdateData{}))
				keep(unsafe.Pointer(unit.UpdateDataPlayer().Player), unsafe.Sizeof(Player{}))
			}
			var events []any
			log := func(name string, args ...any) { events = append(events, append([]any{name}, args...)) }
			var currentPlayer *Object
			traceAt, attempts := 0, 0
			recordRoles := make(map[*MonsterSpawnScreenRecord50D960]int)
			freed := make(map[int]monsterSpawnScreenRecordExpected)
			vector := func(p types.Pointf) []float32 { return []float32{p.X, p.Y} }
			rectangle := func(r types.Rectf) []float32 { return []float32{r.Min.X, r.Min.Y, r.Max.X, r.Max.Y} }
			rt := MonsterSpawnDeleteRuntime50E210{
				DelayedDelete: func(obj *Object) { log("delayed_delete", addresses[obj]) },
				FirstPlayerUnit: func() *Object {
					log("first_player")
					currentPlayer = nil
					if len(players) > 0 {
						currentPlayer = players[0]
					}
					return currentPlayer
				},
				NextPlayerUnit: func(obj *Object) *Object {
					log("next_player", addresses[obj])
					currentPlayer = nil
					if i := playerIndices[obj] + 1; i < len(players) {
						currentPlayer = players[i]
					}
					return currentPlayer
				},
				MaxOnscreen: func() float64 {
					log("balance", "MaxOnscreenMonsterCount")
					if c.Limit != nil {
						return *c.Limit
					}
					return 2
				},
				PointInRect: func(point types.Pointf, rect types.Rectf) bool {
					log("point_in_rect", vector(point), rectangle(rect))
					return point.X >= rect.Min.X && point.X <= rect.Max.X && point.Y >= rect.Min.Y && point.Y <= rect.Max.Y
				},
				Trace: func(from, to types.Pointf, flags MapTraceFlags) int {
					log("trace", append(vector(from), vector(to)...), 0, 0, int(flags))
					value := 1
					if len(c.TraceResults) > 0 {
						value = c.TraceResults[traceAt%len(c.TraceResults)]
					}
					traceAt++
					return value
				},
				Distance: func(a, b *Object) float64 {
					log("distance", addresses[a], addresses[b])
					unit, spawn := a, b
					if _, ok := playerIndices[unit]; !ok {
						unit, spawn = b, a
					}
					if len(c.Distances) > 0 {
						role := (addresses[spawn] - 0x24002000) / 0x400
						return c.Distances[playerIndices[unit]][role]
					}
					return math.Hypot(float64(unit.PosVec.X)-float64(spawn.PosVec.X), float64(unit.PosVec.Y)-float64(spawn.PosVec.Y))
				},
				EachUnitInRect: func(rect types.Rectf, visit func(*Object)) {
					log("units_in_rect", rectangle(rect), 0x50dfb0, addresses[currentPlayer])
					order := c.Enumeration
					if order == nil {
						order = []int{0, 1, 2, 3}
					}
					for _, role := range order {
						visit(objects[role])
					}
				},
				IsZombie: func(obj *Object) int { log("zombie", addresses[obj]); return c.Zombie },
				NewScreenRecord: func() *MonsterSpawnScreenRecord50D960 {
					log("alloc_zero", 0x24000200)
					attempts++
					if c.ScratchFailAt == attempts {
						return nil
					}
					record := new(MonsterSpawnScreenRecord50D960)
					recordRoles[record] = len(recordRoles)
					return record
				},
				FreeScreenRecord: func(record *MonsterSpawnScreenRecord50D960) {
					role := recordRoles[record]
					log("free_object", 0x24000200, 0x24000600+role*0xa0)
					if _, ok := freed[role]; ok {
						t.Fatal("record freed twice")
					}
					value := monsterSpawnScreenRecordExpected{Role: role, AverageBits: math.Float32bits(record.Average), Mask: record.Mask}
					for i, distance := range record.Distances {
						value.DistanceBits[i] = math.Float32bits(distance)
					}
					freed[role] = value
				},
			}
			s.monsterSpawns50D780.visibleCount = c.VisibleCount
			for _, observation := range fixture.Observations {
				events = []any{}
				var result uint32
				switch observation.Entry {
				case "screen":
					result = uint32(s.MonsterSpawnCullVisible50D960(rt))
				case "distant":
					s.MonsterSpawnCullDistant50D8D0(rt)
				case "visible":
					s.MonsterSpawnVisible50DFB0(objects[0], players[0], rt)
				case "can_place":
					if s.MonsterSpawnCanPlaceWithRuntime50DE80(types.Ptf(c.Point[0], c.Point[1]), rt) {
						result = 1
					}
				case "tick":
					result = s.MonsterSpawnTick50D890(rt)
				default:
					t.Fatal("unknown original entry", observation.Entry)
				}
				if observation.ReturnValue != nil && result != *observation.ReturnValue {
					t.Fatal("scalar return", result, *observation.ReturnValue)
				}
				encoded, err := json.Marshal(events)
				if err != nil {
					t.Fatal(err)
				}
				var normalized []any
				if err = json.Unmarshal(encoded, &normalized); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(normalized, observation.Services) {
					t.Fatalf("ordered native service trace differs:\ngot %s\nwant %v", encoded, observation.Services)
				}
				globals, err := hex.DecodeString(observation.GlobalsHex)
				if err != nil || len(globals) != 156 {
					t.Fatal("global owner fixture", err)
				}
				if s.monsterSpawns50D780.visibleCount != binary.LittleEndian.Uint32(globals[:4]) || s.monsterSpawns50D780.screenCount != binary.LittleEndian.Uint32(globals[20:24]) {
					t.Fatal("visible/scratch count differs from original")
				}
				for i, value := range s.monsterSpawns50D780.screenCounts {
					if uint32(value) != binary.LittleEndian.Uint32(globals[24+4*i:]) {
						t.Fatal("per-player count", i, value)
					}
				}
				if s.monsterSpawns50D780.screenHead != nil || len(s.monsterSpawns50D780.screenPins) != 0 {
					t.Fatal("scratch owners remain live")
				}
				for pointer, before := range owners {
					if !bytes.Equal(before, unsafe.Slice((*byte)(pointer), len(before))) {
						t.Fatal("unrelated complete native owner changed")
					}
				}
			}
			if len(freed) != len(recordCases[row].Records) {
				t.Fatal("released owner count", len(freed), len(recordCases[row].Records))
			}
			for _, want := range recordCases[row].Records {
				if got, ok := freed[want.Role]; !ok || got != want {
					t.Fatal("released record value differs from original", want.Role, got, want)
				}
			}
		})
	}
}

func TestMonsterSpawnDistantNativeCollisionSurface(t *testing.T) {
	s := new(Server)
	defer s.MonsterSpawnFree50D820()
	s.MonsterSpawnInit50D780()
	generator, _ := monsterSpawnTestGenerator50D780(0)
	spawned, _ := monsterSpawnTestMonster50D780(0)
	s.MonsterSpawnRegister50E030(generator, spawned, MonsterSpawnRegisterRuntime50E030{})
	player := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: &Player{Field4792: 1}})}
	player.Shape.Kind = ShapeKindCircle
	player.Shape.Circle.R = 20
	deleted := 0
	rt := MonsterSpawnDeleteRuntime50E210{FirstPlayerUnit: func() *Object { return player }, NextPlayerUnit: func(*Object) *Object { return nil }, DelayedDelete: func(*Object) { deleted++ }}
	spawned.PosVec = types.Ptf(710, 0)
	s.MonsterSpawnCullDistant50D8D0(rt)
	if deleted != 0 {
		t.Fatal("center distance replaced original collision-surface distance")
	}
	spawned.PosVec = types.Ptf(721, 0)
	s.MonsterSpawnCullDistant50D8D0(rt)
	if deleted != 1 {
		t.Fatal("distant monster was not deleted")
	}
}
