package legacy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func teamEntriesBytes417CF0[T any](values []T) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&values[0])), len(values)*int(unsafe.Sizeof(values[0])))
}

func TestRemainingTeamEntries4183C0FrozenOriginal(t *testing.T) {
	data, err := os.ReadFile("testdata/chat_nearest_original_4183c0_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var original chatNearestOriginal215
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if original.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || original.Function.Address != "0x4183c0" || original.Function.Bytes != 272 || original.ControlWord != "0x0e7f" || original.Executions != 168 || len(original.Faults) != 0 || !original.Guards || len(original.Cases) != 56 || original.Function.SHA256 != "3201fec59d0de584d3bbf363ef55e9a03d1549629400e26666f50468eea3ed47" {
		t.Fatal("frozen original scope")
	}
	for _, row := range original.Cases {
		for _, guard := range []byte{0, 0x5a, 0xff} {
			for _, public := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%02x/C=%t", row.Name, guard, public), func(t *testing.T) {
					teams, ft := alloc.Make([]server.Team{}, 3)
					defer ft()
					units, fu := alloc.Make([]server.Object{}, 3)
					defer fu()
					flags, ff := alloc.Make([]server.Object{}, 3)
					defer ff()
					updates, fd := alloc.Make([]server.PlayerUpdateData{}, 3)
					defer fd()
					players, fp := alloc.Make([]server.Player{}, 3)
					defer fp()
					banks := [][]byte{teamEntriesBytes417CF0(teams), teamEntriesBytes417CF0(units), teamEntriesBytes417CF0(flags), teamEntriesBytes417CF0(updates), teamEntriesBytes417CF0(players)}
					for _, b := range banks {
						for i := range b {
							b[i] = guard
						}
					}
					for i, v := range row.Teams {
						teams[i].IDVal = server.TeamID(v.ID)
						teams[i].Field_72 = nil
						if v.Flag != 0 {
							teams[i].Field_72 = unsafe.Pointer(&flags[i])
						}
						flags[i].PosVec.X = math.Float32frombits(v.X)
						flags[i].PosVec.Y = math.Float32frombits(v.Y)
					}
					for i, v := range row.Units {
						units[i].PosVec.X = math.Float32frombits(v.X)
						units[i].PosVec.Y = math.Float32frombits(v.Y)
						units[i].TeamVal.ID = server.TeamID(v.Team)
						units[i].NetCode = v.Code
						units[i].UpdateData = nil
						if v.Update != 0 {
							units[i].UpdateData = unsafe.Pointer(&updates[i])
						}
						updates[i].Player = &players[i]
						players[i].NetCodeVal = v.Code
						players[i].Field3680 = v.Status
					}
					if unsafe.Sizeof(uintptr(0)) > 4 {
						for _, b := range banks {
							if uintptr(unsafe.Pointer(&b[0])) <= math.MaxUint32 {
								t.Fatal("native owner not above4GiB")
							}
						}
					}
					before := make([][]byte, len(banks))
					for i, b := range banks {
						before[i] = bytes.Clone(b)
					}
					events := make([]any, 0)
					unitIndex := func(p *server.Object) int {
						for i := range units {
							if p == &units[i] {
								return i
							}
						}
						t.Fatal("wrong/truncated unit")
						return -1
					}
					teamIndex := func(p *server.Team) int {
						for i := range teams {
							if p == &teams[i] {
								return i
							}
						}
						t.Fatal("wrong/truncated team")
						return -1
					}
					memberIndex := func(p *server.ObjectTeam) int {
						for i := range units {
							if p == &units[i].TeamVal {
								return i
							}
						}
						t.Fatal("wrong/truncated member")
						return -1
					}
					rt := teamEntriesRuntime417CF0{
						FirstUnit: func() *server.Object {
							events = append(events, []any{"unit-first"})
							if len(row.Units) == 0 {
								return nil
							}
							return &units[0]
						},
						NextUnit: func(p *server.Object) *server.Object {
							i := unitIndex(p)
							events = append(events, []any{"unit-next", i})
							if i+1 == len(row.Units) {
								return nil
							}
							return &units[i+1]
						},
						FirstTeam: func() *server.Team {
							events = append(events, []any{"team-first"})
							if len(row.Teams) == 0 {
								return nil
							}
							return &teams[0]
						},
						NextTeam: func(p *server.Team) *server.Team {
							i := teamIndex(p)
							events = append(events, []any{"team-next", i})
							if i+1 == len(row.Teams) {
								return nil
							}
							return &teams[i+1]
						},
						ByID: func(id server.TeamID) *server.Team {
							events = append(events, []any{"by-id", uint32(id)})
							for i := range row.Teams {
								if teams[i].ID() == id {
									return &teams[i]
								}
							}
							return nil
						},
						ClientCode: func() uint32 { return row.Client }, NoRendering: func() bool { return row.Engine&0x40000 != 0 },
						Attach: func(id server.TeamID, m *server.ObjectTeam, code int32) {
							events = append(events, []any{"attach", uint32(id), memberIndex(m), 1, uint32(code), 0})
						},
						Change: func(m *server.ObjectTeam, tm *server.Team, code int32) {
							events = append(events, []any{"change", memberIndex(m), teamIndex(tm), uint32(code), 0})
						},
					}
					pop := pushTeamEntriesRuntime417CF0(&rt)
					defer pop()
					var got int32
					if public {
						got = teamEntriesCall4183C0()
					} else {
						got = sub4183C0Native()
					}
					if got != 0 {
						t.Fatal("exhausted iterator return", got)
					}
					actual, _ := json.Marshal(events)
					expected, _ := json.Marshal(row.Events)
					if !bytes.Equal(actual, expected) {
						t.Fatalf("original service transcript\ngot %s\nwant%s", actual, expected)
					}
					for i, b := range banks {
						if !bytes.Equal(b, before[i]) {
							t.Fatalf("owner bank%d modified", i)
						}
					}
				})
			}
		}
	}
}

func TestRemainingTeamEntries417CF0ABIAndServices(t *testing.T) {
	// Object has two Go-only handles after its shared C prefix.
	prefix := unsafe.Offsetof(server.Object{}.ScriptPickup) + unsafe.Sizeof(server.Object{}.ScriptPickup)
	if unsafe.Sizeof(server.Object{}) != prefix+2*unsafe.Sizeof(uintptr(0)) {
		t.Fatal("native object prefix followed by two Go-only handles")
	}
	want := []uintptr{unsafe.Sizeof(server.Team{}), 60, unsafe.Offsetof(server.Team{}.Field_72), prefix, unsafe.Offsetof(server.Object{}.UpdateData), unsafe.Offsetof(server.PlayerUpdateData{}.Player), unsafe.Offsetof(server.Player{}.NetCodeVal), unsafe.Offsetof(server.Player{}.Field3680)}
	for i, v := range want {
		if c := teamEntriesLayout417CF0(i); c != v {
			t.Fatalf("C/Go layout field%d %d/%d", i, c, v)
		}
	}
	for _, value := range []int32{0, 1, -1, math.MinInt32, math.MaxInt32} {
		var order []string
		rt := teamEntriesRuntime417CF0{RemoveActive: func() int32 { order = append(order, "remove-false"); return value }, Limit: func() int32 { order = append(order, "actual-limit"); return value }}
		a, b := func() (int32, int32) {
			pop := pushTeamEntriesRuntime417CF0(&rt)
			defer pop()
			return teamEntriesCall417CF0(), teamEntriesCall417DC0()
		}()
		if a != value || b != value || !reflect.DeepEqual(order, []string{"remove-false", "actual-limit"}) {
			t.Fatalf("service return/order %v %d/%d", order, a, b)
		}
	}
	for _, count := range []byte{0, 1, 128, 255} {
		var order []int
		rt := teamEntriesRuntime417CF0{Count: func() byte { order = append(order, 10); return count }, SetGameplay: func(v uint32) {
			if v != 2 {
				t.Fatal(v)
			}
			order = append(order, 20)
		}, AutoAssign: func(v int32) {
			if v != 0 {
				t.Fatal(v)
			}
			order = append(order, 30)
		}}
		got := func() int32 {
			pop := pushTeamEntriesRuntime417CF0(&rt)
			defer pop()
			return teamEntriesCall418390()
		}()
		expected := []int{10}
		result := int32(0)
		if count != 0 {
			expected = []int{10, 20, 30}
			result = 1
		}
		if got != result || !reflect.DeepEqual(order, expected) {
			t.Fatalf("count%d return/order%d %v", count, got, order)
		}
	}
	for _, count := range []int{0, 1, 127, 128, 255, 256, 300} {
		teams, free := alloc.Make([]server.Team{}, 301)
		nextCalls := 0
		for i := range teams {
			*(*uint32)(unsafe.Add(teams[i].C(), 60)) = 0x80000000
		}
		rt := teamEntriesRuntime417CF0{FirstTeam: func() *server.Team {
			if count == 0 {
				return nil
			}
			return &teams[0]
		}, NextTeam: func(p *server.Team) *server.Team {
			nextCalls++
			index := (uintptr(p.C()) - uintptr(teams[0].C())) / unsafe.Sizeof(teams[0])
			if int(index)+1 == count {
				return nil
			}
			return &teams[index+1]
		}}
		got := func() byte {
			pop := pushTeamEntriesRuntime417CF0(&rt)
			defer pop()
			return teamEntriesCall417DE0()
		}()
		free()
		if got != byte(count) || nextCalls != count {
			t.Fatalf("charwrap/count%d got%d next%d", count, got, nextCalls)
		}
	}
	cell := (*uint32)(settingEligibilityFixtureLimitCell40AA70())
	saved := *cell
	defer func() { *cell = saved }()
	for _, v := range []uint32{0, 0x80000000, 0xffffffff} {
		*cell = v
		if got := teamEntriesCall417DC0(); uint32(got) != v {
			t.Fatalf("actualCglobal %08x/%08x", got, v)
		}
	}
}

func TestRemainingTeamEntries4184D0NativePacket(t *testing.T) {
	data, err := os.ReadFile("testdata/team_packet_native_before_port_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		ActualOldNativeC bool `json:"actual_old_native_C"`
		Cases            []struct {
			Index                 int
			Nil, Flag512, Mutate  bool
			Before, After, Packet string
			Events                []string
		}
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	if !golden.ActualOldNativeC || len(golden.Cases) != 8 {
		t.Fatal("retained old native C packet scope")
	}
	for _, row := range golden.Cases {
		for _, public := range []bool{false, true} {
			t.Run(fmt.Sprintf("case%d/C=%t", row.Index, public), func(t *testing.T) {
				before, err := hex.DecodeString(row.Before)
				if err != nil || len(before) != 1024 {
					t.Fatal("before owner")
				}
				expectedAfter, _ := hex.DecodeString(row.After)
				expectedPacket, _ := hex.DecodeString(row.Packet)
				owner, release := alloc.Malloc(1024)
				defer release()
				bank := unsafe.Slice((*byte)(owner), 1024)
				copy(bank, before)
				team := (*server.Team)(owner)
				if row.Nil {
					team = nil
				}
				var order []string
				var packet []byte
				var allocated unsafe.Pointer
				var releasePacket func()
				rt := teamEntriesRuntime417CF0{
					Prepare: func(p *server.Team) {
						if p != team {
							t.Fatal("wrong Prepare owner")
						}
						order = append(order, "prepare")
						if row.Mutate {
							*(*uint16)(owner) = 'Z'
							*(*uint32)(unsafe.Add(owner, 60)) = 0x89abcdef
							*(*int32)(unsafe.Add(owner, 52)) = -17
							*(*byte)(unsafe.Add(owner, 57)) = 0x81
							*(*byte)(unsafe.Add(owner, 56)) = 0xe2
							*(*uint32)(unsafe.Add(owner, 68)) = 0xab234556
						}
					},
					Flag512: func() bool { order = append(order, "flag512"); return row.Flag512 },
					Alloc: func(n uintptr) unsafe.Pointer {
						order = append(order, "alloc")
						p, free := alloc.Malloc(n + 16)
						releasePacket = free
						raw := unsafe.Slice((*byte)(p), int(n)+16)
						for i := range raw {
							raw[i] = 0xa5
						}
						allocated = unsafe.Add(p, 8)
						return allocated
					},
					Send: func(p unsafe.Pointer, n int) {
						order = append(order, "send")
						if p != allocated {
							t.Fatal("wrong packet owner")
						}
						packet = bytes.Clone(unsafe.Slice((*byte)(p), n))
						for _, off := range []int{-8, -1, n, n + 7} {
							if *(*byte)(unsafe.Add(p, off)) != 0xa5 {
								t.Fatal("packet boundary overwritten")
							}
						}
						runtime.GC()
					},
					Free: func(p unsafe.Pointer) {
						order = append(order, "free")
						if p != allocated {
							t.Fatal("wrong Free owner")
						}
						releasePacket()
					},
				}
				func() {
					pop := pushTeamEntriesRuntime417CF0(&rt)
					defer pop()
					if public {
						teamEntriesCall4184D0(unsafe.Pointer(team))
					} else {
						sub4184D0Native(team)
					}
				}()
				if len(order) == 0 {
					order = []string{}
				}
				if !reflect.DeepEqual(order, row.Events) || !bytes.Equal(packet, expectedPacket) || !bytes.Equal(bank, expectedAfter) {
					t.Fatalf("oldC packet/order/wholeowner mismatch case%d events%v/%v packet%x/%x", row.Index, order, row.Events, packet, expectedPacket)
				}
			})
		}
	}
}
