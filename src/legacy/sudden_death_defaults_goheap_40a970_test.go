//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

// This loader reads only the frozen independent input. Original observations
// are opened by the separate post-call assertion helper below.
func suddenDeathExtraInput40A970(t *testing.T, name string) sdInput40A970 {
	t.Helper()
	f, z, err := rskOpenGzip4E23C0(filepath.Join("testdata", "sudden_death_40a970", "inputs-v1.json.gz"), sdInputSHA40A970)
	if err != nil {
		t.Fatal(err)
	}
	var input struct {
		Cases []sdInput40A970 `json:"cases"`
		CW    []uint16        `json:"FPU_controls"`
	}
	dec := json.NewDecoder(z)
	err = dec.Decode(&input)
	var extra any
	if err == nil {
		if end := dec.Decode(&extra); end != io.EOF {
			err = fmt.Errorf("input document EOF: %v", end)
		}
	}
	ze, fe := z.Close(), f.Close()
	if err != nil || ze != nil || fe != nil || len(input.Cases) != 403 || len(input.CW) != 2 || input.CW[0] != 0x0e3f || input.CW[1] != 0x023f {
		t.Fatalf("frozen input/CRC/control words: %v %v %v", err, ze, fe)
	}
	var selected *sdInput40A970
	for i := range input.Cases {
		if input.Cases[i].Name == name {
			if selected != nil {
				t.Fatal("unique selected independent input")
			}
			selected = &input.Cases[i]
		}
	}
	if selected == nil {
		t.Fatal("selected frozen independent input absent")
	}
	return *selected
}

func suddenDeathExtraCall40A970(t *testing.T, backend string) {
	t.Helper()
	switch backend {
	case "Go":
		sub40A970Native()
	case "C":
		suddenDeathFixtureCall40A970()
	default:
		t.Fatal("declared void backend")
	}
}

type suddenDeathExtraServer40A970 struct {
	Server
	native *server.Server
}

func (s *suddenDeathExtraServer40A970) S() *server.Server { return s.native }

func suddenDeathExtraWorld40A970(t *testing.T, in *sdInput40A970, threshold string) *server.Server {
	t.Helper()
	var oldCells [4]uint32
	for i := range oldCells {
		cell := suddenDeathNativeCell40A970(i + 1)
		witness := twoSettingsPointerFixturePointerWitness40A740(cell)
		if !witness.Above4GiB || !witness.Low32LosesBits || witness.Native != uintptr(cell) {
			t.Fatal("actual production C state pointer retains all native bits")
		}
		t.Logf("SUDDEN_DEATH_PRODUCTION_CELL_POINTER_OK field=%d native=%#x c-witness=%#x high=true low32-loses-bits=true", i+1, uintptr(cell), witness.Native)
		oldCells[i] = *(*uint32)(cell)
	}
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() {
		for i, value := range oldCells {
			*(*uint32)(suddenDeathNativeCell40A970(i + 1)) = value
		}
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	srv := server.NewHigh(nil, nil, strman.New())
	previous, oldFrame, oldData := GetServer, gameFrameHook, datapath.Data("")
	GetServer = func() Server { return &suddenDeathExtraServer40A970{native: srv} }
	gameFrameHook = func() uint32 { return GetServer().S().Frame() }
	t.Cleanup(func() { srv.Close(); GetServer = previous; gameFrameHook = oldFrame; datapath.SetData(oldData) })
	root := ""
	if base := os.Getenv("OPENNOX_LOCAL_TEST_ARTIFACTS"); base != "" {
		owned, err := filepath.Abs(filepath.Join(base, "sudden-death-extra-controls"))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(owned, 0700); err != nil {
			t.Fatal(err)
		}
		root, err = os.MkdirTemp(owned, "fixture-")
		if err != nil {
			t.Fatal(err)
		}
	} else {
		root = t.TempDir()
	}
	data := "SuddenDeathPlayerThreshold: " + threshold + "\nSuddenDeathLifeTime: 0.5\n"
	path := filepath.Join(root, "gamedata.yml")
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		raw, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(raw, []byte(data)) {
			t.Errorf("actual balance fixture changed: %v", err)
		}
	})
	t.Logf("SUDDEN_DEATH_LINKED_BALANCE_INPUT path=%s sha256=%s source-input-sha256=%s", path, rskHash4E23C0([]byte(data)), sdInputSHA40A970)
	datapath.SetData(root)
	if err := srv.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	srv.SetFrame(in.InitialWords["0084ea04"])
	for field := 1; field <= 4; field++ {
		*(*uint32)(suddenDeathNativeCell40A970(field)) = in.InitialWords[fmt.Sprintf("%08x", sdCellPE40A970[field])]
	}
	noxflags.ResetGame()
	return srv
}

// Call only after the actual native entry. This reads unchanged authentic
// original fields; it does not calculate conversion/branch results.
func suddenDeathExtraReferenceAfterCall40A970(t *testing.T, name string, cw uint16, srv *server.Server) *sdReference40A970 {
	t.Helper()
	cases, err := sdLoadCases40A970()
	if err != nil {
		t.Fatal(err)
	}
	var ref *sdReference40A970
	for _, item := range cases {
		if item.Input.Name == name {
			ref = item.References[cw]
			break
		}
	}
	if ref == nil || ref.Status != "returned" || !ref.Valid {
		t.Fatal("authentic successful selected original reference")
	}
	for field := 1; field <= 4; field++ {
		want, ok := ref.AfterFields[sdCellPE40A970[field]]
		if !ok || len(want) != 4 {
			t.Fatal("whole original public DWORD")
		}
		if got := *(*uint32)(suddenDeathNativeCell40A970(field)); got != sdWord40A970(want) {
			t.Fatalf("actual production state field%d got=%08x original=%x", field, got, want)
		}
	}
	if want := ref.AfterFields[sdCellPE40A970[0]]; len(want) != 4 || srv.Frame() != sdWord40A970(want) {
		t.Fatal("actual Go frame owner versus original selected input")
	}
	return ref
}

func TestSuddenDeathDefaultLinked40A970(t *testing.T) {
	const selected = "0040a970/raw-gamedata-format-and-bits/0040a98a-f64-positive-one-and-half"
	for _, backend := range []string{"Go", "C"} {
		for _, withPlayers := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/players-%t", backend, withPlayers), func(t *testing.T) {
				if suddenDeathNativeObserverCurrent40A970() != nil {
					t.Fatal("actual nil-observer control")
				}
				in := suddenDeathExtraInput40A970(t, selected)
				srv := suddenDeathExtraWorld40A970(t, &in, "1.5")
				var players []*server.Player
				if withPlayers {
					for i, status := range []uint32{1, 0x100, 0x80000100} {
						p := srv.Players.NewRaw(9700 + i)
						if p == nil {
							t.Fatal("real default player iterator slot")
						}
						p.Field3680 = status
						players = append(players, p)
					}
				}
				noxflags.SetGame(noxflags.GameFlag(0x84000000))
				suddenDeathExtraCall40A970(t, backend)
				suddenDeathExtraReferenceAfterCall40A970(t, in.Name, 0x0e3f, srv)
				if noxflags.GetGame() != noxflags.GameFlag(0x80000000) {
					t.Fatal("actual linked GameFlags unset preserves unrelated flag")
				}
				if withPlayers {
					// These are the linked417530 helper's clear-bit controls, not
					// claimed observations of its excluded original helper body.
					for i, want := range []uint32{1, 0, 0x80000000} {
						if players[i].Field3680 != want {
							t.Fatal("actual linked status helper DWORD mask0x100")
						}
					}
				}
				t.Logf("SUDDEN_DEATH_DEFAULT_LINKED_OK backend=%s players=%t observer=nil actual-C-double-parser=true actual-Go-frame=true actual-production-C-state4=true actual-player-helper-mask0100=true actual-flags-unset=true helper-original-body-credit=0 private-FP-stack-credit=0", backend, withPlayers)
			})
		}
	}
}

type suddenDeathExtraService40A970 struct {
	Site     uint32
	Args     []suddenDeathNativeArg40A970
	Response int
}

func TestSuddenDeathPinnedGoHeap40A970(t *testing.T) {
	const selected = "0040a970/after-completed-status-read/00000100-to-00000000"
	for _, backend := range []string{"Go", "C"} {
		for _, cw := range []uint16{0x0e3f, 0x023f} {
			t.Run(fmt.Sprintf("%s/%04x", backend, cw), func(t *testing.T) {
				in := suddenDeathExtraInput40A970(t, selected)
				srv := suddenDeathExtraWorld40A970(t, &in, "1.0")
				var players [3]*server.Player
				tokens := map[uint32]unsafe.Pointer{0: nil}
				var before [3][]byte
				var pin runtime.Pinner
				defer pin.Unpin()
				for i := range players {
					role := fmt.Sprintf("player%d", i)
					token := in.Roles[role]
					var initial []byte
					for _, r := range in.Regions {
						if r.Base == token {
							initial = sdHex40A970(r.Initial)
							break
						}
					}
					if len(initial) != 3684 {
						t.Fatal("selected full distinct player input")
					}
					p := &server.Player{Field3680: sdWord40A970(initial[3680:3684]), PlayerInd: byte(i), Field3688: 0xa0b0c000 + uint32(i), Level: 0x5a + uint8(i)}
					players[i] = p
					pin.Pin(p)
					ptr := unsafe.Pointer(p)
					w := twoSettingsPointerFixturePointerWitness40A740(ptr)
					if !w.Above4GiB || !w.Low32LosesBits || w.Native != uintptr(ptr) {
						t.Fatal("individually pinned full native Go player passed through actual C witness")
					}
					if _, ok := tokens[token]; ok {
						t.Fatal("selected role identities distinct")
					}
					tokens[token] = ptr
					before[i] = bytes.Clone(unsafe.Slice((*byte)(ptr), int(unsafe.Sizeof(*p))))
				}
				gcs, mutations, firsts, nexts, unsets := 0, 0, 0, 0, 0
				forceGC := func() { runtime.GC(); gcs++ }
				var services []suddenDeathExtraService40A970
				var reads []uint32
				counts := map[uint32]int{}
				apply := func(writes []sdWrite40A970) {
					for _, w := range writes {
						data := sdHex40A970(w.Hex)
						var dst unsafe.Pointer
						for i, p := range players {
							base := in.Roles[fmt.Sprintf("player%d", i)] + 3680
							if w.Address >= base && uint64(w.Address)+uint64(len(data)) <= uint64(base)+4 {
								dst = unsafe.Add(unsafe.Pointer(p), unsafe.Offsetof(server.Player{}.Field3680)+uintptr(w.Address-base))
								break
							}
						}
						if dst == nil {
							t.Fatal("selected input effect remains within typed status scalar")
						}
						copy(unsafe.Slice((*byte)(dst), len(data)), data)
					}
				}
				observer := &suddenDeathNativeObserver40A970{ControlWord: cw}
				observer.Operand = func(e suddenDeathNativeOperand40A970) {
					if e.Field != 5 || !e.Complete {
						return
					}
					if e.Site != 0x40a9c5 || e.Width != 4 || !e.ValueValid || !e.AddressValid || e.Write {
						t.Fatal("actual DWORD status observation")
					}
					var token uint32
					for original, p := range tokens {
						if p == e.Base {
							token = original
							break
						}
					}
					if token == 0 || e.Offset != unsafe.Offsetof(server.Player{}.Field3680) || e.Address != unsafe.Add(e.Base, e.Offset) {
						t.Fatal("actual full Go player field address")
					}
					reads = append(reads, e.Value)
					for _, stim := range in.AfterReads {
						if stim.PC == e.Site && stim.Address == token+3680 && stim.Width == 4 && stim.Occurrence == len(reads) {
							apply(stim.Writes)
							mutations++
						}
					}
					forceGC()
				}
				observer.Service = func(site uint32, args []suddenDeathNativeArg40A970) (suddenDeathNativeReply40A970, bool) {
					key := fmt.Sprintf("%08x", site)
					schedule, ok := in.Services[key]
					if !ok || len(args) != schedule.Argc || len(schedule.Responses) == 0 {
						t.Fatal("selected actual service site/arity")
					}
					index := counts[site]
					counts[site]++
					if index >= len(schedule.Responses) {
						if !schedule.Repeat {
							t.Fatal("selected repeated reply")
						}
						index = len(schedule.Responses) - 1
					}
					packet := schedule.Responses[index]
					services = append(services, suddenDeathExtraService40A970{site, append([]suddenDeathNativeArg40A970(nil), args...), index})
					reply := suddenDeathNativeReply40A970{Word: packet.Registers["eax"]}
					if site == 0x40a98a || site == 0x40a9a1 {
						var err error
						reply.FloatFormat, reply.FloatPayload, err = sdRawFPInput40A970(packet.X87)
						if err != nil {
							t.Fatal(err)
						}
					}
					if site == 0x40a9ba || site == 0x40a9df {
						var ok bool
						reply.Pointer, ok = tokens[packet.Registers["eax"]]
						if !ok {
							t.Fatal("selected reply requires an actual typed Go identity")
						}
					}
					if site == 0x40a9ba {
						firsts++
						forceGC()
					}
					if site == 0x40a9d6 {
						unsets++
						forceGC()
					}
					if site == 0x40a9df {
						nexts++
						forceGC()
					}
					apply(packet.Writes)
					if len(in.AfterServices) != 0 || len(in.AfterWrites) != 0 {
						t.Fatal("selected witness has only declared after-read mutation")
					}
					return reply, true
				}
				pop := pushSuddenDeathNativeObserver40A970(observer)
				func() { defer pop(); suddenDeathExtraCall40A970(t, backend) }()
				ref := suddenDeathExtraReferenceAfterCall40A970(t, in.Name, cw, srv)
				if len(services) != len(ref.Services) {
					t.Fatal("whole original selected service sequence")
				}
				for i, actual := range services {
					want := ref.Services[i]
					if actual.Site != want.Site || actual.Response != want.ResponseIndex || len(actual.Args) != len(want.Args) {
						t.Fatal("actual cached service order/response index")
					}
					for j, arg := range actual.Args {
						switch arg.Kind {
						case "player":
							if p, ok := tokens[want.Args[j]]; !ok || arg.Pointer != p {
								t.Fatal("cached full player argument after mutation and GC")
							}
						case "word":
							if arg.Word != want.Args[j] {
								t.Fatal("full actual service DWORD")
							}
						case "string":
							key := "SuddenDeathPlayerThreshold"
							if actual.Site == 0x40a9a1 {
								key = "SuddenDeathLifeTime"
							}
							if arg.String != key {
								t.Fatal("actual gamedata string")
							}
						default:
							t.Fatal("declared selected service argument kind")
						}
					}
				}
				var wantReads []uint32
				for _, event := range ref.Events {
					if event.Kind == "read" && event.Completed && event.PC == 0x40a9c5 {
						var raw string
						if err := json.Unmarshal(event.Value, &raw); err != nil {
							t.Fatal(err)
						}
						p, err := hex.DecodeString(raw)
						if err != nil || len(p) != 4 {
							t.Fatal("original DWORD read bytes")
						}
						wantReads = append(wantReads, sdWord40A970(p))
					}
				}
				if len(reads) != len(wantReads) {
					t.Fatal("actual original status read count")
				}
				for i := range reads {
					if reads[i] != wantReads[i] {
						t.Fatal("actual cached DWORD before/after mutation")
					}
				}
				if mutations != len(in.AfterReads) || mutations != 1 || firsts != 1 || nexts != 3 || unsets != 2 || gcs != len(reads)+firsts+nexts+unsets {
					t.Fatalf("meaningful GC/cache witness mutation=%d first=%d next=%d unset=%d GC=%d", mutations, firsts, nexts, unsets, gcs)
				}
				for i, p := range players {
					want := bytes.Clone(before[i])
					field := ref.AfterFields[in.Roles[fmt.Sprintf("player%d", i)]+3680]
					if len(field) != 4 {
						t.Fatal("authentic original whole selected status")
					}
					offset := int(unsafe.Offsetof(server.Player{}.Field3680))
					copy(want[offset:offset+4], field)
					if !bytes.Equal(want, unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))) {
						t.Fatal("whole typed Go allocation/canary changed outside actual selected status")
					}
				}
				runtime.KeepAlive(players)
				t.Logf("SUDDEN_DEATH_PINNED_GO_OK backend=%s cw=%04x typed-player-pins=3 actual-C-pointer-witnesses=3 in-call-GC=%d source-input-mutations=1 actual-cached-status-reads=%d actual-unset-calls=2 actual-next-calls=3 whole-typed-allocations=3 original-public-DWORDs4=true private-FP-stack-credit=0", backend, cw, gcs, len(reads))
			})
		}
	}
}
