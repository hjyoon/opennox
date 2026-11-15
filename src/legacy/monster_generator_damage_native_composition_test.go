package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

func generatorDamageCompositionServer(t *testing.T) (*server.Server, *server.Player) {
	t.Helper()
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.NetList.Init()
	t.Cleanup(s.NetList.Free)
	old := GetServer
	GetServer = func() Server { return &resourceSpellLegacyServer{srv: s} }
	t.Cleanup(func() { GetServer = old })
	pl := s.Players.NewRaw(1001)
	if pl == nil {
		t.Fatal("actual FX receiver")
	}
	update := resourceSpellCompositionOwner(t, server.PlayerUpdateData{Player: pl})
	pl.PlayerUnit = resourceSpellCompositionOwner(t, server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)})
	pl.Field10, pl.Field12 = 32767, 32767
	return s, pl
}

func TestMonsterGeneratorDamageNativeDefaultPointFXMatchesC(t *testing.T) {
	const child = "OPENNOX_GENERATOR_DAMAGE_WIRE_CHILD"
	if os.Getenv(child) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestMonsterGeneratorDamageNativeDefaultPointFXMatchesC$", "-test.v", "-test.timeout=3m")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual default/C wire parity:\n%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	s, receiver := generatorDamageCompositionServer(t)
	point := resourceSpellCompositionOwner(t, types.Pointf{})
	positions := []types.Pointf{{X: 1.5, Y: -1.5}, {X: 2.5, Y: -2.5}, {X: .5, Y: -.5}, {X: 65535.75, Y: -65535.75}, {X: 65536.75, Y: -65536.75}, {X: 32768.5, Y: -32768.5}, {X: 2147483520, Y: -2147483648}, {X: 2147483648, Y: -2147483904}, {}}
	for i, pos := range positions {
		for j, op := range [][2]byte{{0xf0, 26}, {0, 0}, {0xff, 0x80}} {
			t.Run(fmt.Sprintf("point%d/op%d", i, j), func(t *testing.T) {
				*point = pos
				receiver.PlayerUnit.PosVec = pos
				monsterGeneratorDamageDefaultRuntime4E27D0().PointFX(op[0], op[1], point)
				actual := s.NetList.CopyPacketsA(ntype.PlayerInd(receiver.PlayerInd), netlist.Kind1)
				generatorDamagePointFXExistingC(op[0], op[1], point)
				originalC := s.NetList.CopyPacketsA(ntype.PlayerInd(receiver.PlayerInd), netlist.Kind1)
				if len(actual) != 6 || len(originalC) != 6 || !bytes.Equal(actual, originalC) {
					t.Fatalf("actual default packet % x vs unchanged C % x at%v", actual, originalC, pos)
				}
			})
		}
	}
	t.Log("actual default wire cases=27; actual high-address receiver/point, NetList and unchanged public C523150, signed-DWORD conversion boundaries and char opcode/subtype bytes")
}

//go:embed testdata/monster_generator_damage_default_4e27d0_v1.json.gz
var generatorDamageDefaultOriginal []byte

// Actual HP, DefaultDamage's Invulnerable guard, status/sync services, network
// emission and queued audio run without a runtime override. Script dispatch is
// outside this unchanged-HP slice; the main reference matrix checks that link.
func TestMonsterGeneratorDamageNativeDefaultComposition(t *testing.T) {
	const child = "OPENNOX_GENERATOR_DAMAGE_DEFAULT_CHILD"
	if os.Getenv(child) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestMonsterGeneratorDamageNativeDefaultComposition$", "-test.v", "-test.timeout=3m")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual generator default composition:\n%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	s, receiver := generatorDamageCompositionServer(t)
	var rows []generatorDamageCase
	npcVoiceOriginalDecode(t, generatorDamageDefaultOriginal, "d4dc933fbbad2563311f33d56b312741ab057a9880d29c30c9756ee42d792699", "0670c655b42f818112e7989c24819f1ba83ca91871311f8736388ccafde46960", 48, &rows)
	comparisonPoint := resourceSpellCompositionOwner(t, types.Pointf{})
	compositions := 0
	for route := 0; route < 3; route++ {
		v := generatorDamageArenaNew(t, 64)
		for _, c := range rows {
			t.Run(fmt.Sprintf("route%d/%s", route, c.Input.Name), func(t *testing.T) {
				v.prepare(t, c)
				for i, o := range v.objects {
					state := c.Before.Objects[i]
					var h *server.HealthData
					var u unsafe.Pointer
					if state.Health >= 0 {
						h = v.health[state.Health]
					}
					if state.Update >= 0 {
						u = unsafe.Pointer(v.updates[state.Update])
					}
					*o = server.Object{ObjClass: object.ClassImmobile, ObjFlags: object.Flags(state.Flags), Field5: state.Status, PosVec: types.Ptf(math.Float32frombits(state.Position[0]), math.Float32frombits(state.Position[1])), Frame134: state.EffectFrame, HealthData: h, UpdateData: u, Buffs: 1 << server.ENCHANT_INVULNERABLE}
				}
				generatorDamageBindC(v.objects[0])
				s.SetFrame(c.Input.Frame)
				receiver.PlayerUnit.PosVec = v.objects[0].PosVec
				if got := v.state(); !reflect.DeepEqual(got, c.Before) {
					t.Fatalf("default native input got%+v want%+v", got, c.Before)
				}
				want := bytes.Clone(v.raw)
				targetOffset := v.offset(unsafe.Pointer(v.objects[0]))
				binary.LittleEndian.PutUint32(want[targetOffset+int(v.layout[2]):], c.After.Objects[0].Status)
				var fx *generatorDamageEvent
				statusChange := false
				for i := range c.Events {
					e := &c.Events[i]
					if e.Kind == "point-fx" {
						if fx != nil {
							t.Fatal("unexpected second original FX")
						}
						fx = e
					}
					if e.Kind == "script" {
						t.Fatal("unchanged-HP default slice must not enter script")
					}
					if e.Kind == "set-status" || e.Kind == "unset-status" {
						statusChange = true
					}
				}
				if statusChange {
					binary.LittleEndian.PutUint32(want[targetOffset+int(unsafe.Offsetof(server.Object{}.Field38)):], math.MaxUint32)
					for i := range v.objects[0].Field140 {
						binary.LittleEndian.PutUint32(want[targetOffset+int(unsafe.Offsetof(server.Object{}.Field140))+i*4:], 0x80000)
					}
				}
				queuedBefore := reflect.ValueOf(&s.Audio).Elem().FieldByName("delayedObj").Len()
				got := generatorDamageActualCall(route, v.objects[0], v.object(c.Input.Source), v.object(c.Input.Weapon), c.Input.Damage, c.Input.DamageType)
				if uint32(got) != c.Return {
					t.Fatalf("actual default return %#x want%#x", uint32(got), c.Return)
				}
				if state := v.state(); !reflect.DeepEqual(state, c.After) {
					t.Fatalf("actual default final got%+v want%+v", state, c.After)
				}
				v.expected = want
				v.checkArena(t)
				packet := s.NetList.CopyPacketsA(ntype.PlayerInd(receiver.PlayerInd), netlist.Kind1)
				wantAudio := 0
				if fx != nil {
					wantAudio = 1
					*comparisonPoint = types.Ptf(math.Float32frombits(fx.PointBits[0]), math.Float32frombits(fx.PointBits[1]))
					generatorDamagePointFXExistingC(byte(fx.Args[0]), byte(fx.Args[1]), comparisonPoint)
					control := s.NetList.CopyPacketsA(ntype.PlayerInd(receiver.PlayerInd), netlist.Kind1)
					if len(packet) != 6 || len(control) != 6 || !bytes.Equal(packet, control) {
						t.Fatalf("real default wire % x want unchanged C % x", packet, control)
					}
				} else if len(packet) != 0 {
					t.Fatalf("unexpected default FX packet % x", packet)
				}
				queued := reflect.ValueOf(&s.Audio).Elem().FieldByName("delayedObj")
				if queued.Len() != queuedBefore+wantAudio {
					t.Fatal("actual queued audio count", queued.Len(), queuedBefore, wantAudio)
				}
				if wantAudio != 0 {
					e := queued.Index(queuedBefore)
					if e.FieldByName("ID").Int() != 1001 || e.FieldByName("Obj").Pointer() != uintptr(unsafe.Pointer(v.objects[0])) || e.FieldByName("Kind").Int() != 0 || e.FieldByName("Code").Uint() != 0 {
						t.Fatal("actual queued generator-hurt audio/owner/kind/code")
					}
				}
				compositions++
			})
		}
	}
	if compositions != 144 {
		t.Fatal("missing real default compositions", compositions)
	}
	t.Log("actual generator default compositions=144; real frame/HP/defaultDamage Invulnerable guard/status+sync/net emission/queued sound1001, C wire controls; unchanged HP leaves script unexecuted in this slice")
}
