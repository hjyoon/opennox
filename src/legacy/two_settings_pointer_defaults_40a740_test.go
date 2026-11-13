//go:build darwin && arm64

package legacy

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func twoSettingsPointerExtraCall40A740(t *testing.T, backend string, entry uint32) int32 {
	t.Helper()
	if backend == "C" {
		return twoSettingsPointerFixtureCall40A740(entry)
	}
	switch entry {
	case 0x40a740:
		return sub40A740Native()
	case 0x40a770:
		return sub40A770Native()
	default:
		t.Fatal("declared extra-control entry")
		return 0
	}
}

func twoSettingsPointerExtraSaveFlags40A740(t *testing.T) {
	t.Helper()
	game, engine, play := noxflags.GetGame(), noxflags.GetEngine(), noxflags.GetGamePlay()
	noxflags.UnsetGame(game)
	noxflags.UnsetEngine(engine)
	noxflags.UnsetGamePlay(play)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGame(game)
		noxflags.SetEngine(engine)
		noxflags.SetGamePlay(play)
	})
}

func TestTwoSettingsPointerDefaultInfo40A740(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		for _, c := range []struct {
			name   string
			flags  noxflags.GameFlag
			value  byte
			result int32
		}{{"off-zero", 0, 0, 0}, {"off-highbit", 0, 0x80, 0}, {"on-positive-byte", 128, 0x7f, 0}, {"on-highbit", 128, 0x80, 1}} {
			t.Run(backend+"/"+c.name, func(t *testing.T) {
				if twoSettingsPointerNativeObserverCurrent40A740() != nil {
					t.Fatal("actual nil-observer default control")
				}
				twoSettingsPointerExtraSaveFlags40A740(t)
				info := memmap.Slice(0x5d4594, 371380)[:58]
				index := (*uint32)(memmap.PtrOff(0x5d4594, 371688))
				saved, savedIndex := bytes.Clone(info), *index
				defer func() { copy(info, saved); *index = savedIndex }()
				*index, info[53] = 0, c.value
				noxflags.SetGame(c.flags)
				before := bytes.Clone(info)
				w := twoSettingsPointerFixturePointerWitness40A740(unsafe.Pointer(&info[0]))
				if !w.Above4GiB || !w.Low32LosesBits || w.Native != uintptr(unsafe.Pointer(&info[0])) {
					t.Fatal("actual C info blob has full native pointer identity")
				}
				if got := twoSettingsPointerExtraCall40A740(t, backend, 0x40a740); got != c.result || *index != 0 || !bytes.Equal(info, before) {
					t.Fatalf("actual default InfoGetter/GameFlags got=%d want=%d", got, c.result)
				}
				t.Logf("TWO_SETTINGS_POINTER_DEFAULT_INFO_OK backend=%s control=%s observer=nil actual-info-cell=true actual-gameflags=true native-high=true helper-original-body-credit=0", backend, c.name)
			})
		}
	}
}

type twoSettingsPointerDefaultServer40A740 struct {
	Server
	native *server.Server
}

func (s *twoSettingsPointerDefaultServer40A740) S() *server.Server { return s.native }

func TestTwoSettingsPointerDefaultIterators40A740(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		t.Run(backend, func(t *testing.T) {
			if twoSettingsPointerNativeObserverCurrent40A740() != nil {
				t.Fatal("actual nil-observer default iterator control")
			}
			twoSettingsPointerExtraSaveFlags40A740(t)
			srv := server.NewHigh(nil, nil, strman.New())
			previous := GetServer
			GetServer = func() Server { return &twoSettingsPointerDefaultServer40A740{native: srv} }
			defer func() { srv.Close(); GetServer = previous }()
			players := [3]*server.Player{srv.Players.NewRaw(7100), srv.Players.NewRaw(7101), srv.Players.NewRaw(7102)}
			var free []func()
			defer func() {
				for _, p := range players {
					if p != nil {
						p.PlayerUnit = nil
					}
				}
				for i := len(free) - 1; i >= 0; i-- {
					free[i]()
				}
			}()
			for i, p := range players {
				if p == nil {
					t.Fatal("actual initialized native player slot")
				}
				unit, freeUnit := alloc.NewHigh(server.Object{})
				update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
				free = append(free, freeUnit, freeUpdate)
				*unit = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
				*update = server.PlayerUpdateData{Player: p}
				p.PlayerUnit = unit
				p.Field3680 = 0xffff0000
				if i == 1 {
					p.Field3680 |= 1
				}
				if i == 2 {
					p.PlayerInd = 31
				}
				for _, ptr := range []unsafe.Pointer{unsafe.Pointer(p), unsafe.Pointer(unit), unsafe.Pointer(update)} {
					if uintptr(ptr) <= 0xffffffff {
						t.Fatal("actual default player/unit/update pointer must exceed4GiB")
					}
				}
			}
			calls := 0
			check := func(name string, want int32) {
				t.Helper()
				calls++
				if got := twoSettingsPointerExtraCall40A740(t, backend, 0x40a770); got != want {
					t.Fatalf("actual default %s got=%d want=%d", name, got, want)
				}
			}
			check("nonteam-two-eligible", 2)
			noxflags.SetEngine(noxflags.EngineFlag(0x40000))
			check("nonteam-special-player-disabled", 1)
			noxflags.SetGamePlay(noxflags.GameplayFlag4)
			check("team-mode-no-teams", 0)
			noxflags.SetGame(noxflags.GameModeCoopTeam)
			first, second := srv.Teams.Create(1), srv.Teams.Create(2)
			if first == nil || second == nil || srv.Teams.MemberCount(first) != 0 || srv.Teams.MemberCount(second) != 0 {
				t.Fatal("two actual teams without attached members")
			}
			check("each-team-scans-global-unit-list", 2)
			for _, p := range players {
				p.Field3680 |= 1
			}
			check("all-global-units-eliminated", 0)
			if calls != 5 {
				t.Fatal("five actual nil-observer calls")
			}
			t.Logf("TWO_SETTINGS_POINTER_DEFAULT_ITERATORS_OK backend=%s observer=nil actual-server-player-team-unit-services=true calls=5 actual-no-membership-behavior=true helper-original-body-credit=0", backend)
		})
	}
}

func TestTwoSettingsPointerPinnedGoHeap40A740(t *testing.T) {
	for _, backend := range []string{"Go", "C"} {
		for _, teamMode := range []bool{false, true} {
			for _, c := range []struct {
				word   uint32
				result int32
			}{{0, 1}, {0x40000, 0}, {0x80000000, 1}, {0xffffffff, 0}} {
				t.Run(fmt.Sprintf("%s/team-%t/engine-%08x", backend, teamMode, c.word), func(t *testing.T) {
					twoSettingsPointerExtraSaveFlags40A740(t)
					noxflags.SetEngine(noxflags.EngineFlag(c.word))
					player := &server.Player{PlayerInd: 31, Field3680: 0xffff0000}
					update := &server.PlayerUpdateData{Player: player}
					unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
					teamOwner := new([32]byte)
					for i := range teamOwner {
						teamOwner[i] = byte(i*31 + 7)
					}
					team := unsafe.Pointer(&teamOwner[1])
					owners := []unsafe.Pointer{unsafe.Pointer(player), unsafe.Pointer(update), unsafe.Pointer(unit), unsafe.Pointer(teamOwner)}
					var pin runtime.Pinner
					defer pin.Unpin()
					for _, ptr := range owners {
						pin.Pin(ptr)
						w := twoSettingsPointerFixturePointerWitness40A740(ptr)
						if !w.Above4GiB || !w.Low32LosesBits || w.Native != uintptr(ptr) {
							t.Fatal("individually pinned high Go pointer C witness")
						}
					}
					if uintptr(team)%unsafe.Alignof(unsafe.Pointer(nil)) == 0 {
						t.Fatal("opaque team token must be actually unaligned")
					}
					sizes := []uintptr{unsafe.Sizeof(*player), unsafe.Sizeof(*update), unsafe.Sizeof(*unit), unsafe.Sizeof(*teamOwner)}
					before := make([][]byte, len(owners))
					for i, p := range owners {
						before[i] = bytes.Clone(unsafe.Slice((*byte)(p), int(sizes[i])))
					}
					gcs, engines := 0, 0
					forceGC := func() { runtime.GC(); gcs++ }
					observer := &twoSettingsPointerNativeObserver40A740{Operand: func(e twoSettingsPointerNativeOperand40A740) {
						if !e.Complete {
							return
						}
						if e.Field == 3 || e.Field == 4 {
							forceGC()
						}
						if e.Field == 5 {
							if e.AddressValid || e.Address != nil || e.Base != nil || e.Width != 4 || e.IsPointer || !e.ValueValid || e.ValueWord != c.word {
								t.Fatal("actual full Go EngineFlags getter, no diagnostic cell or PE alias")
							}
							engines++
							forceGC()
						}
					}}
					observer.Service = func(site uint32, args []twoSettingsPointerNativeArg40A740) (twoSettingsPointerNativeReply40A740, bool) {
						switch site {
						case 0x40a778:
							if len(args) != 1 || args[0].IsPointer || args[0].Word != 4 {
								t.Fatal("actual gameplay raw argument")
							}
							if teamMode {
								return twoSettingsPointerNativeReply40A740{Word: 0x100}, true
							}
							return twoSettingsPointerNativeReply40A740{}, true
						case 0x40a784:
							if !teamMode || len(args) != 0 {
								t.Fatal("team first only in selected team path")
							}
							return twoSettingsPointerNativeReply40A740{Pointer: team}, true
						case 0x40a79a:
							if !teamMode || len(args) != 0 {
								t.Fatal("unit first only in selected team path")
							}
							forceGC()
							return twoSettingsPointerNativeReply40A740{Pointer: unsafe.Pointer(unit)}, true
						case 0x40a7ee:
							if teamMode || len(args) != 0 {
								t.Fatal("player first only in selected nonteam path")
							}
							forceGC()
							return twoSettingsPointerNativeReply40A740{Pointer: unsafe.Pointer(player)}, true
						case 0x40a7c9, 0x40a7d9, 0x40a819:
							want := unsafe.Pointer(unit)
							if site == 0x40a7d9 {
								want = team
							}
							if site == 0x40a819 {
								want = unsafe.Pointer(player)
							}
							if len(args) != 1 || !args[0].IsPointer || args[0].Pointer != want {
								t.Fatal("actual cached full native iterator argument")
							}
							return twoSettingsPointerNativeReply40A740{}, true
						default:
							t.Fatalf("unexpected service %x", site)
							return twoSettingsPointerNativeReply40A740{}, true
						}
					}
					pop := pushTwoSettingsPointerNativeObserver40A740(observer)
					var result int32
					func() { defer pop(); result = twoSettingsPointerExtraCall40A740(t, backend, 0x40a770) }()
					if result != c.result || gcs != 4 || engines != 1 {
						t.Fatalf("actual pinned/default-engine result=%d want=%d GC=%d engineReads=%d", result, c.result, gcs, engines)
					}
					for i, p := range owners {
						if !bytes.Equal(before[i], unsafe.Slice((*byte)(p), int(sizes[i]))) {
							t.Fatal("whole pinned graph allocation changed")
						}
					}
					runtime.KeepAlive(player)
					runtime.KeepAlive(update)
					runtime.KeepAlive(unit)
					runtime.KeepAlive(teamOwner)
					t.Logf("TWO_SETTINGS_POINTER_PINNED_GO_OK backend=%s team=%t engine=%08x individually-pinned-allocations=4 forced-in-call-GC=4 actual-Go-engine-getter=true opaque-team-unaligned=true no-Go-pointer-cgo-bypass=true helper-original-body-credit=0", backend, teamMode, c.word)
				})
			}
		}
	}
}
