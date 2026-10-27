package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestModifierStatusCastOriginalNativeC(t *testing.T) {
	blob, err := os.ReadFile("testdata/modifier_status_cast_original.json")
	if err != nil {
		t.Fatal(err)
	}
	type event struct {
		Service   string   `json:"service"`
		Spell     int      `json:"spell"`
		Power     uint32   `json:"power_bits32"`
		SourceNil bool     `json:"source_nil"`
		Position  []uint32 `json:"position_bits32"`
		Index     int      `json:"player_index"`
		Value     int      `json:"value"`
	}
	var fixture struct {
		SHA       string            `json:"original_sha256"`
		Unchanged bool              `json:"original_unchanged"`
		CW        uint16            `json:"actual_startup_control_word"`
		Faults    []json.RawMessage `json:"memory_faults"`
		Rows      []struct {
			Input struct {
				Which     int    `json:"which"`
				Name      string `json:"name"`
				Class     int    `json:"target_class"`
				Power     int32  `json:"power"`
				NilSource bool   `json:"nil_source"`
				Mutation  int    `json:"mutation"`
			} `json:"input"`
			Events    []event `json:"events"`
			Preserved bool    `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || len(fixture.Rows) != 56 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete original status effect fixture")
	}
	extract := func(file, sig string) string {
		t.Helper()
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		if strings.Count(text, sig) != 1 {
			t.Fatal("nonunique production body")
		}
		begin := strings.Index(text, sig)
		end := begin + strings.Index(text[begin:], "{") + 1
		for depth := 1; depth != 0; end++ {
			if end == len(text) {
				t.Fatal("incomplete body")
			}
			if text[end] == '{' {
				depth++
			}
			if text[end] == '}' {
				depth--
			}
		}
		return text[begin:end]
	}
	harness, err := os.ReadFile("testdata/modifier_status_cast_native_test.c")
	if err != nil {
		t.Fatal(err)
	}
	combined := string(harness)
	for _, item := range [][3]string{
		{"FLOAT2INT", "GAME1_1.c", "int nox_float2int(float a1)"},
		{"PLAYER", "GAME3_2.c", "static nox_playerInfo* nox_xxx_modifierEffectPlayer_4E04D0("},
		{"INFORM", "GAME3_2.c", "static void nox_xxx_modifierEffectInform_4E04D0("},
		{"STUN", "GAME3_2.c", "void nox_xxx_stunEffect_4E04D0("},
		{"CONFUSE", "GAME3_2.c", "void nox_xxx_confuseEffect_4E0670("},
	} {
		marker := "// PRODUCTION_" + item[0]
		if strings.Count(combined, marker) != 1 {
			t.Fatal("nonunique fixture marker")
		}
		combined = strings.Replace(combined, marker, extract(item[1], item[2]), 1)
	}
	dir := t.TempDir()
	input, binary := filepath.Join(dir, "status.c"), filepath.Join(dir, "status")
	if err := os.WriteFile(input, []byte(combined), 0600); err != nil {
		t.Fatal(err)
	}
	include, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	args := append(compiler[1:], "-std=c11", "-O2", "-fshort-wchar", "-D_Static_assert(x,y)=", "-I"+include, input, "-lm", "-o", binary)
	if out, err := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...).CombinedOutput(); err != nil {
		t.Fatalf("compile exact status bodies: %v\n%s", err, out)
	}
	var stdin strings.Builder
	for _, row := range fixture.Rows {
		n := 0
		if row.Input.NilSource {
			n = 1
		}
		fmt.Fprintf(&stdin, "%d %d %d %d %d\n", row.Input.Which, row.Input.Class, row.Input.Power, n, row.Input.Mutation)
	}
	command := exec.Command(binary)
	command.Stdin = strings.NewReader(stdin.String())
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("native status effects: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 56 {
		t.Fatalf("native rows=%d", len(lines))
	}
	for i, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			var got struct {
				Bits      int     `json:"native_pointer_bits"`
				Preserved bool    `json:"nonoutput_bytes_preserved"`
				Events    []event `json:"events"`
			}
			if err := json.Unmarshal([]byte(lines[i]), &got); err != nil {
				t.Fatal(err)
			}
			if got.Bits != strconv.IntSize || !got.Preserved || !row.Preserved || !reflect.DeepEqual(got.Events, row.Events) {
				t.Fatalf("native=%s original=%+v", lines[i], row.Events)
			}
		})
	}
}

// Keep the actual enchantment-slot gate, C callback, native CGo cast export,
// shared server spell logic and live player attribution. Buff application is
// an explicit service that retargets the local spell argument before attribution.
func TestModifierStatusCastActualPreAttack(t *testing.T) {
	for which, callback := range modifierStatusEffectPointers4E04D0() {
		for _, slot := range []int{2, 3} {
			for _, power := range []int32{0x12345678, -129} {
				t.Run(fmt.Sprintf("effect-%d/slot-%d/power-%d", which, slot, power), func(t *testing.T) {
					srv := npcReflectLegacyServer4E17B0(t)
					oldFrame, oldApply := gameFrameHook, buffApplyExportImpl4FF380
					oldStun, oldConfuse := castStunCall52C2C0, castConfuseCall52C1E0
					t.Cleanup(func() {
						gameFrameHook = oldFrame
						buffApplyExportImpl4FF380 = oldApply
						castStunCall52C2C0 = oldStun
						castConfuseCall52C1E0 = oldConfuse
					})
					gameFrameHook = func() uint32 { return 1400 }
					sourcePlayer := selfWeaponElectricNativeAlloc(t, server.Player{PlayerInd: 254})
					sourceUpdate := selfWeaponElectricNativeAlloc(t, server.PlayerUpdateData{Player: sourcePlayer})
					source := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(sourceUpdate)})
					sourcePlayer.PlayerUnit = source
					target := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassMonster, Mass: 12, PosVec: types.Ptf(-12.75, -4.5)})
					livePlayer := selfWeaponElectricNativeAlloc(t, server.Player{PlayerInd: 130})
					liveUpdate := selfWeaponElectricNativeAlloc(t, server.PlayerUpdateData{Player: livePlayer})
					live := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(liveUpdate)})
					livePlayer.PlayerUnit = live
					effect := selfWeaponElectricNativeAlloc(t, server.ModifierEff{Attack40: server.ModifierEffFnc{Val: 99}, AttackPreHit52: server.ModifierEffFnc{Fnc: callback, Val: power}})
					data := selfWeaponElectricNativeAlloc(t, server.ModifierInitData{})
					data.Modifiers[slot] = effect
					weapon := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(data)})
					var argument *server.SpellAcceptArg
					casts, applies := 0, 0
					wrap := func(delegate func(*server.Object, *server.SpellAcceptArg, int32) int32) func(*server.Object, *server.SpellAcceptArg, int32) int32 {
						return func(caster *server.Object, arg *server.SpellAcceptArg, level int32) int32 {
							casts++
							argument = arg
							if caster != source || arg.Obj != target || arg.Pos != target.PosVec || level != power {
								t.Fatal("C callback lost cast identity, full power or cached target position")
							}
							if unsafe.Sizeof(uintptr(0)) > 4 {
								for _, p := range []unsafe.Pointer{unsafe.Pointer(arg), unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(effect), unsafe.Pointer(weapon)} {
									if uintptr(p) <= math.MaxUint32 {
										t.Fatal("expected native pointer above 4 GiB")
									}
								}
							}
							return delegate(caster, arg, level)
						}
					}
					if which == 0 {
						castStunCall52C2C0 = wrap(oldStun)
					} else {
						castConfuseCall52C1E0 = wrap(oldConfuse)
					}
					buffApplyExportImpl4FF380 = func(unit *server.Object, buff int32, duration int16, strength int8) {
						applies++
						want := int32(5)
						if which == 1 {
							want = 3
						}
						if unit != target || buff != want || duration != 0 || strength != int8(power) || argument == nil {
							t.Fatal("shared spell application or narrowing changed")
						}
						argument.Obj = live
					}
					if result := srv.PlayerPreAttackEffects538290(target, source, weapon, nil); result != 0 || casts != 1 || applies != 1 {
						t.Fatalf("actual enchantment slot result=%d casts=%d applies=%d", result, casts, applies)
					}
					if pending, index, frame := livePlayer.LastAggressorState(); pending != 1 || index != 254 || frame != 1400 {
						t.Fatalf("live attribution=%d/%d/%d", pending, index, frame)
					}
					t.Logf("actual enchantment slot %d -> status C callback -> shared native cast -> live attribution", slot)
				})
			}
		}
	}
}
