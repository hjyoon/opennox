package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type playerStatOutputsOriginal map[string]uint64
type playerStatEventOriginal struct {
	Service string                    `json:"service"`
	Mask    uint32                    `json:"mask"`
	Value   uint32                    `json:"value"`
	Level   uint32                    `json:"level"`
	Reward  uint32                    `json:"reward"`
	Bytes   uint32                    `json:"bytes"`
	Token   uint32                    `json:"token"`
	Result  uint32                    `json:"result"`
	Visible playerStatOutputsOriginal `json:"visible_outputs"`
}

func TestPlayerStatPrecisionOriginal4EEDC0Native(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit player stat frames")
	}
	const path = "testdata/player_stat_precision_original.json"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Stats    []struct {
			Name    string     `json:"name"`
			Base    []uint32   `json:"base_bits32"`
			Classes [][]uint32 `json:"class_bits32"`
		} `json:"stats_sets"`
		Cases []struct {
			Name     string                    `json:"name"`
			Stats    string                    `json:"stats_set"`
			Class    uint8                     `json:"class_id"`
			Level    uint8                     `json:"level"`
			Flags    uint32                    `json:"game_flags"`
			Reward   int32                     `json:"reward"`
			Solo     int32                     `json:"solo_mode"`
			Expected playerStatOutputsOriginal `json:"expected_outputs"`
			Result   int32                     `json:"expected_result"`
			Events   []playerStatEventOriginal `json:"services"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 288 || len(fixture.Cases) != 288 || len(fixture.Stats) != 2 || unsafe.Sizeof(Object{}) != 928 || unsafe.Sizeof(PlayerUpdateData{}) != 656 || unsafe.Sizeof(HealthData{}) != 20 || unsafe.Sizeof(Player{}) != 6160 {
		t.Fatal("incomplete original inputs or native layout changed")
	}
	stats := make(map[string]*serverPlayers)
	statValue := func(v []uint32) ClassStats {
		if len(v) != 4 {
			t.Fatal("incomplete stat table")
		}
		return ClassStats{Health: math.Float32frombits(v[0]), Mana: math.Float32frombits(v[1]), Speed: math.Float32frombits(v[2]), Strength: math.Float32frombits(v[3])}
	}
	for _, s := range fixture.Stats {
		if len(s.Classes) != 3 {
			t.Fatal("incomplete class tables")
		}
		v := new(serverPlayers)
		v.Stats.Base = statValue(s.Base)
		v.Stats.Warrior = statValue(s.Classes[0])
		v.Stats.Wizard = statValue(s.Classes[1])
		v.Stats.Conjurer = statValue(s.Classes[2])
		stats[s.Name] = v
	}
	for class, label := range []string{"Warrior", "Wizard", "Conjurer"} {
		t.Run(label, func(t *testing.T) {
			count := 0
			for _, c := range fixture.Cases {
				if c.Class != uint8(class) {
					continue
				}
				count++
				t.Run(c.Name, func(t *testing.T) {
					sizes := []int{928, 656, 20, 6160}
					frames := make([][]byte, 4)
					before := make([][]byte, 4)
					allowed := make([]map[int]bool, 4)
					for i, n := range sizes {
						frame, free := alloc.Make([]byte(nil), n+128)
						defer free()
						for j := range frame {
							frame[j] = 0xa5
						}
						frames[i] = frame
						allowed[i] = make(map[int]bool)
						if uintptr(unsafe.Pointer(&frame[64])) <= 0xffffffff {
							t.Fatal("native frame below4GiB")
						}
					}
					unit := (*Object)(unsafe.Pointer(&frames[0][64]))
					update := (*PlayerUpdateData)(unsafe.Pointer(&frames[1][64]))
					health := (*HealthData)(unsafe.Pointer(&frames[2][64]))
					player := (*Player)(unsafe.Pointer(&frames[3][64]))
					unit.UpdateData = unsafe.Pointer(update)
					unit.HealthData = health
					unit.InvFirstItem = nil
					update.Player = player
					player.Level = c.Level
					player.Info().SetPlayerClass(playerlib.Class(c.Class))
					player.Info().SetName("A")
					player.ProtPlayerField2239 = 0x80001000
					player.ProtPlayerField2235 = 0x80002000
					player.ProtUnitManaMax = 0x80003000
					player.ProtUnitHPMax = 0x80004000
					player.ProtPlayerOrigName = 0x80005000
					health.Cur = 0x1234
					health.Max = 0x5678
					update.ManaCur = 0x2345
					update.ManaMax = 0x6789
					outputs := func() playerStatOutputsOriginal {
						return playerStatOutputsOriginal{"health_current": uint64(health.Cur), "health_maximum": uint64(health.Max), "mana_current": uint64(update.ManaCur), "mana_maximum": uint64(update.ManaMax), "strength": uint64(player.Info().Field2239()), "speed_stat": uint64(player.Info().Field2235()), "speed_base_bits32": uint64(math.Float32bits(unit.SpeedBase)), "mass_bits32": uint64(math.Float32bits(unit.Mass)), "capacity": uint64(uint16(player.field3652)), "carry": uint64(unit.CarryCapacity), "overweight": uint64(player.Field3656), "initialized": uint64(player.StatsReportPending)}
					}
					addAllowed := func(frame int, offset uintptr, n int) {
						for i := 0; i < n; i++ {
							allowed[frame][64+int(offset)+i] = true
						}
					}
					addAllowed(0, unsafe.Offsetof(Object{}.Mass), 4)
					addAllowed(0, unsafe.Offsetof(Object{}.SpeedBase), 4)
					addAllowed(0, unsafe.Offsetof(Object{}.CarryCapacity), 2)
					addAllowed(1, 4, 2)
					addAllowed(1, 8, 2)
					addAllowed(2, unsafe.Offsetof(HealthData{}.Cur), 2)
					addAllowed(2, unsafe.Offsetof(HealthData{}.Max), 2)
					addAllowed(3, unsafe.Offsetof(Player{}.info)+50, 8)
					addAllowed(3, unsafe.Offsetof(Player{}.field3652), 2)
					addAllowed(3, unsafe.Offsetof(Player{}.Field3656), 4)
					addAllowed(3, unsafe.Offsetof(Player{}.StatsReportPending), 1)
					for i := range frames {
						before[i] = append([]byte(nil), frames[i]...)
					}
					table := stats[c.Stats]
					if table == nil {
						t.Fatal("missing declared stats")
					}
					tableBefore := table.Stats
					events := make([]playerStatEventOriginal, 0, len(c.Events))
					record := func(e playerStatEventOriginal) { e.Visible = outputs(); events = append(events, e) }
					checkUnit := func(got *Object) {
						if got != unit {
							t.Fatal("native object service pointer changed")
						}
					}
					result := playerReadValuesNative4EEDC0(unit, c.Reward, table, playerReadValuesNativeDeps4EEDC0{
						gameFlagsCheck: func(mask uint32) int32 {
							var result int32
							if c.Flags&mask != 0 {
								result = 1
							}
							record(playerStatEventOriginal{Service: "flag", Mask: mask, Result: uint32(result)})
							return result
						},
						setHP: func(got *Object, v uint16) {
							checkUnit(got)
							health.Cur = v
							record(playerStatEventOriginal{Service: "HP", Value: uint32(v)})
						},
						soloMode: func() int32 { record(playerStatEventOriginal{Service: "solo", Result: uint32(c.Solo)}); return c.Solo },
						abilityGiveAll: func(got *Object, level int8, reward int32) {
							checkUnit(got)
							record(playerStatEventOriginal{Service: "ability", Level: uint32(uint8(level)), Reward: uint32(reward)})
						},
						protectInt: func(token, v uint32) { record(playerStatEventOriginal{Service: "protect_int", Token: token, Value: v}) },
						protectUint16: func(token uint32, v uint16) {
							record(playerStatEventOriginal{Service: "protect_u16", Token: token, Value: uint32(v)})
						},
						wideLen: func(info *PlayerInfo) uint32 {
							if info != player.Info() || info.Name() != "A" {
								t.Fatal("native wide-name service identity changed")
							}
							return 1
						},
						protectName: func(info *PlayerInfo, n, token uint32) int32 {
							if info != player.Info() {
								t.Fatal("native name protection pointer changed")
							}
							record(playerStatEventOriginal{Service: "protect_name", Bytes: n, Token: token, Result: 0x10203040})
							return 0x10203040
						},
					})
					for i, frame := range frames {
						for j, b := range frame {
							if b != before[i][j] && !allowed[i][j] {
								t.Fatalf("native frame%d nonoutput/canary byte changed at%d", i, j)
							}
						}
					}
					if table.Stats != tableBefore {
						t.Fatal("declared native class tables changed")
					}
					if result != c.Result || !reflect.DeepEqual(outputs(), c.Expected) || !reflect.DeepEqual(events, c.Events) {
						t.Fatalf("actual native player stats=%v result=%08x services=%+v want original=%v result=%08x services=%+v", outputs(), uint32(result), events, c.Expected, uint32(c.Result), c.Events)
					}
				})
			}
			if count != 96 {
				t.Fatal("incomplete class contexts", count)
			}
		})
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatal("original stat fixture changed", err)
	}
}
