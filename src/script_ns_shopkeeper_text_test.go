package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/noxscript/ns/asm"
	ns3 "github.com/opennox/noxscript/ns/v3"
	ns4 "github.com/opennox/noxscript/ns/v4"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The public localized API must reach the existing object setter. The actual
// SCRIPT03 parser/builtin and v3 handle resolver are controls, not substituted
// text-setting callbacks. This fixture supplies no UI/localization/audio service.
func shopkeeperTextProgram(t *testing.T, handle int32, text string) []byte {
	t.Helper()
	var data bytes.Buffer
	word := func(v uint32) {
		if err := binary.Write(&data, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	data.WriteString("SCRIPT03STRG")
	word(1)
	word(uint32(len(text)))
	data.WriteString(text)
	data.WriteString("CODE")
	word(3)
	for i, name := range []string{"GLOBAL", "GLOBAL", "test_shopkeeper_text"} {
		data.WriteString("FUNC")
		word(uint32(len(name)))
		data.WriteString(name)
		word(0)
		word(0)
		data.WriteString("SYMB")
		word(0)
		word(0)
		code := []uint32{uint32(asm.OpReturn0)}
		if i == 2 {
			code = []uint32{
				uint32(asm.OpPushInt), uint32(handle),
				uint32(asm.OpPushString), 0,
				uint32(asm.OpCallBuiltin), uint32(asm.BuiltinSetShopkeeperText),
				uint32(asm.OpReturn0),
			}
		}
		data.WriteString("DATA")
		word(uint32(4 * len(code)))
		for _, value := range code {
			word(value)
		}
	}
	data.WriteString("DONE")
	return data.Bytes()
}

func shopkeeperTextRuntime(t *testing.T) *Server {
	t.Helper()
	s := scriptNSRoamFlagServer515C80(t)
	previous := ns4.Runtime()
	ns4.SetRuntime(s.noxScriptP())
	t.Cleanup(func() { ns4.SetRuntime(previous) })
	// Do not retain a native cache entry after its C-owned fixture is freed.
	_ = legacy.Nox_xxx_script_511C50(0x41234567)
	legacy.Sub_511E20()
	t.Cleanup(legacy.Sub_511E20)
	return s
}

func TestScriptNSSetShopkeeperTextNativeRecordsAndClassicBuiltin(t *testing.T) {
	s := shopkeeperTextRuntime(t)
	texts := []struct{ name, value string }{
		{"empty", ""}, {"localized-id", "Con02:Trader01"},
		{"31-bytes", strings.Repeat("a", 31)}, {"32-bytes", strings.Repeat("b", 32)},
		{"63-bytes", strings.Repeat("c", 63)}, {"embedded-nul", "a\x00bc"},
		{"utf8-id", "상점:인사"}, {"utf8-boundary", "NPC:" + strings.Repeat("가", 12)},
	}
	classes := []struct {
		name     string
		class    object.Class
		subclass object.SubClass
		eligible bool
	}{
		{"class-zero", 0, object.SubClass(object.MonsterShopkeeper), false},
		{"player", object.ClassPlayer, object.SubClass(object.MonsterShopkeeper), false},
		{"monster-no-subclass", object.ClassMonster, 0, false},
		{"ordinary-npc", object.ClassMonster, object.SubClass(object.MonsterNPC), false},
		{"shopkeeper", object.ClassMonster, object.SubClass(object.MonsterShopkeeper), true},
		{"mixed-simple", object.ClassMonster | object.ClassSimple, object.SubClass(object.MonsterShopkeeper), true},
		{"mixed-player", object.ClassMonster | object.ClassPlayer, object.SubClass(object.MonsterShopkeeper), true},
		{"npc-shopkeeper", object.ClassMonster, object.SubClass(object.MonsterNPC | object.MonsterShopkeeper), true},
		{"female-shopkeeper", object.ClassMonster, object.SubClass(object.MonsterFemaleNPC | object.MonsterShopkeeper), true},
	}
	for _, cl := range classes {
		for _, flags := range []object.Flags{0, object.FlagEnabled, object.FlagDead, object.FlagDestroyed, object.FlagNoUpdate} {
			for _, text := range texts {
				t.Run(fmt.Sprintf("%s/flags-%x/%s", cl.name, flags, text.name), func(t *testing.T) {
					unit, freeUnit := alloc.New(server.Object{})
					update, freeUpdate := alloc.New(server.MonsterUpdateData{})
					arena, freeArena := alloc.New([1724 + 64]byte{})
					t.Cleanup(freeArena)
					t.Cleanup(freeUpdate)
					t.Cleanup(freeUnit)
					for i := range arena {
						arena[i] = 0xa5
					}
					init := (*server.ShopkeeperInitData)(unsafe.Pointer(&arena[32]))
					*unit = server.Object{
						ObjClass: cl.class, ObjSubClass: cl.subclass, ObjFlags: flags,
						ScriptIDVal: 0x41234567, Extent: 0xf1234567, NetCode: 73,
						InitData: unsafe.Pointer(init), UpdateData: unsafe.Pointer(update),
					}
					*update = server.MonsterUpdateData{AIStackInd: -1, Field333: 0x12345678, Field124: 0x87654321}
					for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(init)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
							t.Fatalf("text API fixture is not native-width: %p", ptr)
						}
					}
					oldList := s.Objs.List
					s.Objs.List = unit
					// This cleanup runs before the three backing allocations are freed.
					t.Cleanup(func() { legacy.Sub_511E20(); s.Objs.List = oldList })
					api := s.noxScriptP().toObj(unit)
					unitBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
					updateBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))
					arenaBefore := bytes.Clone(arena[:])
					for _, route := range []string{"object-control", "runtime-method", "public-v4", "public-v3", "classic-vm"} {
						copy(arena[:], arenaBefore)
						want := bytes.Clone(arenaBefore)
						resolved := route != "public-v3" && route != "classic-vm" || !flags.Has(object.FlagDestroyed)
						if cl.eligible && resolved {
							// Independent byte oracle for the unchanged bounded setter:
							// retain the suffix after a short ID and always terminate.
							start := 32 + int(unsafe.Offsetof(init.ShopText))
							n := len(text.value)
							if n > 31 {
								n = 31
							}
							for i := 0; i < n; i++ {
								want[start+i] = text.value[i]
							}
							want[start+n] = 0
						}
						var panicked any
						func() {
							defer func() { panicked = recover() }()
							switch route {
							case "object-control":
								api.SetShopText(ns4.StringID(text.value))
							case "runtime-method":
								s.noxScriptP().SetShopkeeperText(api, ns4.StringID(text.value))
							case "public-v4":
								ns4.SetShopkeeperText(api, ns4.StringID(text.value))
							case "public-v3":
								ns3.SetShopkeeperText(ns3.ObjectID(unit.ScriptIDVal), text.value)
							case "classic-vm":
								if err := s.Server.NoxScriptVM.ReadScript(bytes.NewReader(shopkeeperTextProgram(t, unit.ScriptIDVal, text.value))); err != nil {
									t.Fatal(err)
								}
								if err := s.Server.NoxScriptVM.CallByIndex(2, nil, nil); err != nil {
									t.Fatal(err)
								}
							}
						}()
						if panicked != nil {
							t.Errorf("%s localized shopkeeper text panicked: %v", route, panicked)
						}
						if !bytes.Equal(want, arena[:]) ||
							!bytes.Equal(unitBefore, scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
							!bytes.Equal(updateBefore, scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
							t.Errorf("%s changed unexpected object/update/init/guard bytes: shop text=%x", route, init.ShopText)
						}
					}
				})
			}
		}
	}
}

func TestScriptNSSetShopkeeperTextMissingObjects(t *testing.T) {
	s := shopkeeperTextRuntime(t)
	for _, spec := range []struct {
		name string
		obj  ns4.Obj
	}{
		{"nil-interface", nil}, {"empty-native-object", nsObj{}},
	} {
		t.Run(spec.name, func(t *testing.T) {
			for _, route := range []string{"runtime-method", "public-v4"} {
				var panicked any
				func() {
					defer func() { panicked = recover() }()
					if route == "runtime-method" {
						s.noxScriptP().SetShopkeeperText(spec.obj, "Con02:Trader01")
					} else {
						ns4.SetShopkeeperText(spec.obj, "Con02:Trader01")
					}
				}()
				if panicked != nil {
					t.Errorf("%s missing object panicked: %v", route, panicked)
				}
			}
		})
	}
	for _, handle := range []int32{0, -1, -2, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("unresolved-%d", handle), func(t *testing.T) {
			var panicked any
			func() {
				defer func() { panicked = recover() }()
				ns3.SetShopkeeperText(ns3.ObjectID(handle), "Con02:Trader01")
			}()
			if panicked != nil {
				t.Errorf("unresolved v3 handle panicked: %v", panicked)
			}
		})
	}
}

type shopkeeperTextAPIProbe struct {
	ns4.Obj
	calls []ns4.StringID
}

func (obj *shopkeeperTextAPIProbe) SetShopText(id ns4.StringID) {
	obj.calls = append(obj.calls, id)
}

func TestScriptNSSetShopkeeperTextDelegatesLocalizedIDExactlyOnce(t *testing.T) {
	s := shopkeeperTextRuntime(t)
	for _, id := range []ns4.StringID{"", "Con02:Trader01", "a\x00bc", "상점:인사"} {
		t.Run(fmt.Sprintf("id-%x", string(id)), func(t *testing.T) {
			obj := new(shopkeeperTextAPIProbe)
			var panicked any
			func() {
				defer func() { panicked = recover() }()
				ns4.SetShopkeeperText(obj, id)
			}()
			if panicked != nil || len(obj.calls) != 1 || obj.calls[0] != id {
				t.Fatalf("localized ID was not delegated once unchanged: panic=%v calls=%q", panicked, obj.calls)
			}
			_ = s // Keep the real implementation bound until the subtest returns.
		})
	}
}
