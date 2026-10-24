package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type playerCamperUpdateView4D7CC0 struct {
	Previous [2]uint32 `json:"previous_bits"`
	Stamp    uint32    `json:"stamp"`
	Marker   uint32    `json:"marker"`
	Info     string    `json:"info"`
	Canaries [2]uint32 `json:"canaries"`
}
type playerCamperUnitView4D7CC0 struct {
	Position  [2]uint32                    `json:"position_bits"`
	Flags     object.Flags                 `json:"flags"`
	Current   string                       `json:"current_update"`
	Entry     playerCamperUpdateView4D7CC0 `json:"entry"`
	Alternate playerCamperUpdateView4D7CC0 `json:"alternate"`
}
type playerCamperStateView4D7CC0 struct {
	Frame uint32                       `json:"frame"`
	Last  uint32                       `json:"last"`
	Start uint32                       `json:"start"`
	Fade  uint32                       `json:"fade"`
	Units []playerCamperUnitView4D7CC0 `json:"units"`
}
type playerCamperSnapshot4D7CC0 struct {
	At    string                      `json:"at"`
	State playerCamperStateView4D7CC0 `json:"state"`
}

func TestPlayerCamperOriginal4D7CC0(t *testing.T) {
	raw, err := os.ReadFile("testdata/player_camper_update_4d7cc0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"code_bytes_unchanged"`
		Region    struct {
			SHA  string `json:"sha256"`
			Size int    `json:"size"`
		} `json:"region"`
		Rows []struct {
			Input struct {
				Name  string `json:"name"`
				Units []struct {
					Position [2]uint32    `json:"position_bits"`
					Previous [2]uint32    `json:"previous_bits"`
					Stamp    uint32       `json:"stamp"`
					Marker   uint32       `json:"marker"`
					Flags    object.Flags `json:"flags"`
					Role     uint32       `json:"role"`
					Equip    uint32       `json:"equip"`
				} `json:"units"`
				Frame   uint32 `json:"frame"`
				Last    uint32 `json:"last"`
				Start   uint32 `json:"start"`
				Fade    uint32 `json:"fade"`
				Radius  uint32 `json:"radius_bits"`
				CTF     bool   `json:"ctf"`
				FPCW    uint32 `json:"fpcw"`
				Actions []struct {
					At   string `json:"at"`
					Kind string `json:"kind"`
				} `json:"actions"`
			} `json:"input"`
			Events    []string                     `json:"events"`
			Snapshots []playerCamperSnapshot4D7CC0 `json:"snapshots"`
			Final     playerCamperStateView4D7CC0  `json:"final"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.Region.SHA != "c120400a825fce27405bdfd46428516eb526315c2967586ff5ca8248abc606fa" || fixture.Region.Size != 400 || len(fixture.Rows) != 97 {
		t.Fatal("unchanged original camper fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			if row.Input.FPCW != 0xe7f {
				t.Fatal("gameplay53/chop context")
			}
			count := len(row.Input.Units)
			units := make([]Object, count)
			entries := make([]PlayerUpdateData, count)
			alternates := make([]PlayerUpdateData, count)
			infos := make([]Player, count)
			infos2 := make([]Player, count)
			indices := map[*Object]int{}
			for i, c := range row.Input.Units {
				units[i] = Object{ObjClass: object.ClassPlayer, ObjFlags: c.Flags, PosVec: types.Ptf(math.Float32frombits(c.Position[0]), math.Float32frombits(c.Position[1])), UpdateData: unsafe.Pointer(&entries[i])}
				infos[i].Field3680 = c.Role
				infos[i].WeaponEquip = c.Equip
				infos2[i].WeaponEquip = 1
				entries[i] = PlayerUpdateData{Player: &infos[i], Field62: c.Previous[0], Field63: c.Previous[1], Field64: c.Stamp, IsCamping: c.Marker, CurTraps: 0xabab0000 + uint32(i), Field66: 0xcdcd0000 + uint32(i)}
				alternates[i] = PlayerUpdateData{Player: &infos2[i], Field62: 0x3f000000, Field63: 0xbf000000, Field64: 0xabcd1234, IsCamping: 0x77, CurTraps: 0xabab0000 + uint32(i), Field66: 0xcdcd0000 + uint32(i)}
				indices[&units[i]] = i
				for name, ptr := range map[string]unsafe.Pointer{"unit": unsafe.Pointer(&units[i]), "entry": unsafe.Pointer(&entries[i]), "alternate": unsafe.Pointer(&alternates[i]), "info": unsafe.Pointer(&infos[i]), "alternate info": unsafe.Pointer(&infos2[i])} {
					assertQuestPlayerStatePointerAbove32Bits4D79C0(t, name, ptr)
				}
			}
			camper := playerCamper{LastCheck: row.Input.Last, StartTime: row.Input.Start, FadeTime: row.Input.Fade, RadiusSq: float64(math.Float32frombits(row.Input.Radius))}
			frame := row.Input.Frame
			updateView := func(ud *PlayerUpdateData) playerCamperUpdateView4D7CC0 {
				info := "alternate"
				for i := range infos {
					if ud.Player == &infos[i] {
						info = "entry"
					}
				}
				return playerCamperUpdateView4D7CC0{Previous: [2]uint32{ud.Field62, ud.Field63}, Stamp: ud.Field64, Marker: ud.IsCamping, Info: info, Canaries: [2]uint32{ud.CurTraps, ud.Field66}}
			}
			state := func() playerCamperStateView4D7CC0 {
				out := playerCamperStateView4D7CC0{Frame: frame, Last: camper.LastCheck, Start: camper.StartTime, Fade: camper.FadeTime, Units: make([]playerCamperUnitView4D7CC0, count)}
				for i := range units {
					current := "entry"
					if units[i].UpdateData == unsafe.Pointer(&alternates[i]) {
						current = "alternate"
					}
					out.Units[i] = playerCamperUnitView4D7CC0{Position: [2]uint32{math.Float32bits(units[i].PosVec.X), math.Float32bits(units[i].PosVec.Y)}, Flags: units[i].ObjFlags, Current: current, Entry: updateView(&entries[i]), Alternate: updateView(&alternates[i])}
				}
				return out
			}
			mutate := func(at string) {
				for _, a := range row.Input.Actions {
					if a.At != at {
						continue
					}
					switch a.Kind {
					case "swap-update":
						units[0].UpdateData = unsafe.Pointer(&alternates[0])
					case "swap-info":
						entries[0].Player = &infos2[0]
					case "entry-marker":
						entries[0].IsCamping = 0x55
					case "position":
						units[0].PosVec = types.Ptf(math.Float32frombits(0x43804000), math.Float32frombits(0xc3802000))
					case "last-check":
						camper.LastCheck = 999
					case "fade":
						camper.FadeTime = 0
					case "frame":
						frame += 3
					case "next-unit-dead":
						units[1].ObjFlags = object.FlagDead
					case "next-role":
						infos[1].Field3680 = 1
					default:
						t.Fatal("unreviewed mutation", a.Kind)
					}
				}
			}
			events := make([]string, 0)
			snapshots := make([]playerCamperSnapshot4D7CC0, 0)
			firstCalls, current := 0, 0
			phase := ""
			sampling := frame-camper.LastCheck > camper.StartTime
			service := func(kind string, u *Object) {
				i, ok := indices[u]
				if !ok {
					t.Fatal("native service identity")
				}
				at := fmt.Sprintf("%s:%d", kind, i)
				events = append(events, at)
				snapshots = append(snapshots, playerCamperSnapshot4D7CC0{At: at, State: state()})
				mutate(at)
			}
			camper.update4D7CC0(playerCamperRuntime4D7CC0{
				frame: func() uint32 { return frame },
				firstUnit: func() *Object {
					phase = "marker"
					if sampling && firstCalls == 0 {
						phase = "sample"
					}
					firstCalls++
					events = append(events, "first:"+phase)
					current = 0
					if count == 0 {
						return nil
					}
					return &units[0]
				},
				nextUnit: func(u *Object) *Object {
					i, ok := indices[u]
					if !ok {
						t.Fatal("roster identity")
					}
					at := fmt.Sprintf("next:%s:%d", phase, i)
					events = append(events, at)
					mutate(at)
					current = i + 1
					if current == count {
						return nil
					}
					return &units[current]
				},
				isCTF: func() bool {
					at := fmt.Sprintf("ctf:%d", current)
					events = append(events, fmt.Sprintf("%s:%t", at, row.Input.CTF))
					mutate(at)
					return row.Input.CTF
				},
				audio: func(id uint32, u *Object) {
					if id != 774 {
						t.Fatal("original audio ID", id)
					}
					service("audio", u)
				},
				markerOn:  func(u *Object) { service("on", u) },
				markerOff: func(u *Object) { service("off", u) },
			})
			if !reflect.DeepEqual(events, row.Events) || !reflect.DeepEqual(snapshots, row.Snapshots) || !reflect.DeepEqual(state(), row.Final) {
				t.Fatalf("events=%v/%v snapshots=%+v/%+v final=%+v/%+v", events, row.Events, snapshots, row.Snapshots, state(), row.Final)
			}
		})
	}
}
