package legacy

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/player_vital_report_original_v1.json.gz
var playerVitalReportOriginal []byte

type playerVitalReportCase struct {
	Input struct {
		Name, Path, Presence string
		Recipient, Code      uint32
		Quest                bool
		QueueResult          uint32 `json:"queue_result"`
		HP                   []uint16
		Health, Mana         [4]uint32
		Frame, FPS           uint32
		Changes              [2]bool
	}
	Observed struct {
		ReturnKind   string `json:"return_kind"`
		ReturnWord   uint32 `json:"return_word"`
		Events       [][]json.RawMessage
		Health, Mana [4]uint32
		Frame        uint32
	}
}

func playerVitalReportCases(t *testing.T) []playerVitalReportCase {
	t.Helper()
	return tradeOfferOriginalRows[playerVitalReportCase](t, playerVitalReportOriginal,
		"665a3c5583cd4f69d5c60ed80348200090280364d764033ca6eee9156685ec5e", 272)
}

func playerVitalReportWorld(t *testing.T) *server.Server {
	t.Helper()
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	getServer, flags := GetServer, noxflags.GetGame()
	t.Cleanup(func() {
		GetServer = getServer
		noxflags.ResetGame()
		noxflags.SetGame(flags)
	})
	host := &playerReportPacketServer{world: s}
	GetServer = func() Server { return host }
	return s
}

type playerVitalPacket struct {
	Recipient uint32
	Remove    int
	Bytes     string
}

func TestPlayerVitalPacketNativeCABI(t *testing.T) {
	rows, world := playerVitalReportCases(t), playerVitalReportWorld(t)
	for route := range 2 {
		for _, row := range rows {
			in, want := row.Input, row.Observed
			if in.Path == "throttle" {
				continue
			}
			t.Run(fmt.Sprintf("route%d-%s", route, in.Name), func(t *testing.T) {
				unit := game4ConsumerValue(t, server.Object{})
				word := uint32(uint16(in.Code))
				unit.NetCode = word
				if word&0x8000 != 0 {
					unit.ObjClass, unit.NetCode, unit.Extent = object.ClassClientPersist, 1, word&0x7fff
				}
				if in.HP != nil {
					unit.HealthData = game4ConsumerValue(t, server.HealthData{Cur: in.HP[0], Max: in.HP[1]})
				}
				if player := world.Players.ByIndRaw(ntype.PlayerInd(int32(in.Recipient))); player != nil {
					*player = server.Player{}
					defer func() { *player = server.Player{} }()
					if in.Presence != "missing" {
						player.Active = 1
					}
					if in.Presence == "active" {
						player.PlayerUnit = unit
					}
				}
				noxflags.ResetGame()
				if in.Quest {
					noxflags.SetGame(noxflags.GameModeQuest)
				}
				var expected, observed []playerVitalPacket
				for _, event := range want.Events {
					if string(event[0]) != `"packet"` {
						continue
					}
					var packet playerVitalPacket
					if string(event[2]) != "0" || json.Unmarshal(event[1], &packet.Recipient) != nil ||
						json.Unmarshal(event[3], &packet.Remove) != nil || json.Unmarshal(event[4], &packet.Bytes) != nil {
						t.Fatal("invalid original packet trace")
					}
					expected = append(expected, packet)
				}
				world.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
					if related != nil || sequence != 1 {
						t.Fatal("health report changed the reliable transport flags")
					}
					observed = append(observed, playerVitalPacket{uint32(recipient), remove, hex.EncodeToString(packet)})
					return int(int32(in.QueueResult))
				}
				var got uintptr
				target := uintptr(want.ReturnWord)
				if in.Path == "team" {
					got = uintptr(callTeamHealthNative4D8670(route, in.Recipient, unit))
				} else {
					got = callOwnerHealthNative4D86E0(route, in.Recipient)
					if want.ReturnKind == "unit" {
						target = uintptr(unsafe.Pointer(unit))
					}
				}
				if got != target || !reflect.DeepEqual(observed, expected) {
					t.Fatalf("result %#x, packets %v, want %#x and %v", got, observed, target, expected)
				}
			})
		}
	}
}

func TestPlayerVitalsThrottledNative528030(t *testing.T) {
	rows, world := playerVitalReportCases(t), playerVitalReportWorld(t)
	noxflags.ResetGame()
	for _, row := range rows {
		in, want := row.Input, row.Observed
		if in.Path != "throttle" {
			continue
		}
		for variant := range 3 {
			if variant < 2 && (in.Changes[0] || in.Changes[1]) {
				continue
			}
			t.Run(fmt.Sprintf("route%d-%s", variant, in.Name), func(t *testing.T) {
				world.SetFrame(in.Frame)
				world.SetTickRate(in.FPS)
				health := game4ConsumerValue(t, server.HealthData{Cur: uint16(in.Health[0]), Max: uint16(in.Health[2])})
				unit := game4ConsumerValue(t, server.Object{ObjClass: object.ClassPlayer, HealthData: health})
				player := world.Players.ByIndRaw(ntype.PlayerInd(in.Recipient))
				if player == nil {
					player = game4ConsumerValue(t, server.Player{})
				}
				*player = server.Player{Active: 1, PlayerInd: byte(in.Recipient), PlayerUnit: unit,
					HealthReportFrame: in.Health[3], ManaReportFrame: in.Mana[3]}
				defer func() { *player = server.Player{} }()
				update := game4ConsumerValue(t, server.PlayerUpdateData{Player: player,
					ManaCur: uint16(in.Mana[0]), ManaPrev: uint16(in.Mana[1]), ManaMax: uint16(in.Mana[2]),
					Field2_1: uint16(in.Health[1])})
				unit.UpdateData = unsafe.Pointer(update)
				var got uint16
				if variant < 2 {
					world.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
						if uint32(recipient) != in.Recipient || len(data) != 3 || data[0] != 0x43 || related != nil || remove != 1 || sequence != 1 {
							t.Fatal("unexpected live transport from throttled vitals", recipient, data, related, remove, sequence)
						}
						return -1
					}
					got = callVitalsThrottledNative528030(variant, unit)
				} else {
					var events [][2]string
					got = uint16(world.ReportPlayerVitalsThrottled528030(unit, func(ind byte) {
						events = append(events, [2]string{"health", fmt.Sprint(ind)})
						if in.Changes[0] {
							health.Cur = 0x8456
							world.SetFrame(world.Frame() + 7)
						}
					}, func(ind byte, obj *server.Object) {
						events = append(events, [2]string{"mana", fmt.Sprint(ind)})
						if obj != unit {
							t.Fatal("mana callback received the wrong object")
						}
						if in.Changes[1] {
							update.ManaCur = 0xa153
							world.SetFrame(world.Frame() + 4)
						}
					}))
					var expected [][2]string
					for _, event := range want.Events {
						var name string
						if json.Unmarshal(event[0], &name) != nil {
							t.Fatal("invalid original throttle trace")
						}
						expected = append(expected, [2]string{name, string(event[1])})
					}
					if !reflect.DeepEqual(events, expected) {
						t.Fatalf("report callbacks %v, want %v", events, expected)
					}
				}
				hp := [4]uint32{uint32(health.Cur), uint32(update.Field2_1), uint32(health.Max), player.HealthReportFrame}
				mana := [4]uint32{uint32(update.ManaCur), uint32(update.ManaPrev), uint32(update.ManaMax), player.ManaReportFrame}
				if hp != want.Health || mana != want.Mana || world.Frame() != want.Frame || uint32(got) != want.ReturnWord {
					t.Fatalf("vitals %v/%v frame %#x result %#x, want %v/%v frame %#x result %#x",
						hp, mana, world.Frame(), got, want.Health, want.Mana, want.Frame, want.ReturnWord)
				}
			})
		}
	}
}
