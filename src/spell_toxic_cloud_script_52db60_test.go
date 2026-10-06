package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/noxscript/ns/asm"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const toxicCloudScriptChild52DB60 = "OPENNOX_TEST_TOXIC_CLOUD_SCRIPT_NATIVE_WIDTH"
const toxicCloudScriptReturn52DB60 = 0x52db60

// Encode a real SCRIPT03, including the string/float operands consumed by
// builtin 134 and a FrameTimer callback. No builtin, resolver, trace, cast
// result, or activator callback is substituted.
func toxicCloudScript52DB60(t *testing.T, source, target types.Pointf) []byte {
	t.Helper()
	var data bytes.Buffer
	word := func(v uint32) {
		if err := binary.Write(&data, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	data.WriteString("SCRIPT03STRG")
	word(1)
	name := spell.SPELL_TOXIC_CLOUD.String()
	word(uint32(len(name)))
	data.WriteString(name)
	data.WriteString("CODE")
	word(4)
	for i, name := range []string{"GLOBAL", "GLOBAL", "test_toxic_cloud", "test_schedule_toxic_cloud"} {
		data.WriteString("FUNC")
		word(uint32(len(name)))
		data.WriteString(name)
		returns := uint32(0)
		if i >= 2 {
			returns = 1
		}
		word(returns)
		word(0) // arguments
		data.WriteString("SYMB")
		word(0) // locals
		word(0) // unused
		code := []uint32{uint32(asm.OpReturn0)}
		switch i {
		case 2:
			code = []uint32{
				uint32(asm.OpPushString), 0,
				uint32(asm.OpPushFloat), math.Float32bits(source.X),
				uint32(asm.OpPushFloat), math.Float32bits(source.Y),
				uint32(asm.OpPushFloat), math.Float32bits(target.X),
				uint32(asm.OpPushFloat), math.Float32bits(target.Y),
				uint32(asm.OpCallBuiltin), uint32(asm.BuiltinCastSpellLocationLocation),
				uint32(asm.OpPushInt), toxicCloudScriptReturn52DB60,
				uint32(asm.OpReturn),
			}
		case 3:
			code = []uint32{
				uint32(asm.OpPushInt), 3,
				uint32(asm.OpPushInt), 2,
				uint32(asm.OpCallBuiltin), uint32(asm.BuiltinFrameTimer),
				uint32(asm.OpReturn),
			}
		}
		data.WriteString("DATA")
		word(uint32(len(code) * 4))
		for _, v := range code {
			word(v)
		}
	}
	data.WriteString("DONE")
	return data.Bytes()
}

// The crash report enters through a location-to-location map-script timer,
// not the legacy wrapper alone. Exercise that entire production call chain
// with a C-owned ImaginaryCaster above 4 GiB in an isolated test process.
// Empty-map/missing-type allocation is a control-flow/ABI witness, not a
// claim of stock cloud creation, damage, lifetime, or client rendering.
func TestToxicCloudScriptNativeEntry52DB60(t *testing.T) {
	if os.Getenv(toxicCloudScriptChild52DB60) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestToxicCloudScriptNativeEntry52DB60$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), toxicCloudScriptChild52DB60+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Toxic Cloud script entry: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	positions := []struct{ source, target types.Pointf }{
		{types.Ptf(300.25, 300.5), types.Ptf(330.75, 320.25)},
		{types.Ptf(650.5, 975.25), types.Ptf(620.25, 960.5)},
		{types.Ptf(1250.75, 750.25), types.Ptf(1275.5, 715.75)},
	}
	for _, route := range []string{"direct-vm", "frame-timer"} {
		for i, position := range positions {
			t.Run(fmt.Sprintf("%s/%d", route, i), func(t *testing.T) {
				s := NewServer(nil, nil, strman.New())
				t.Cleanup(s.Close)
				if !s.Objs.Init(1) {
					t.Fatal("cannot initialize native object allocator")
				}
				t.Cleanup(s.Objs.FreeObjects)
				s.Map.Init()
				t.Cleanup(s.Map.Free)
				if s.Walls.Init() == 0 {
					t.Fatal("cannot initialize empty native wall map")
				}
				t.Cleanup(s.Walls.Free)
				caster := s.Objs.NewObject(&server.ObjectType{})
				caster.TypeInd = 17
				caster.PosVec = types.Ptf(250, 250)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(caster)) <= math.MaxUint32 {
					t.Fatalf("script caster=%p, want actual allocation above 4 GiB", caster)
				}
				cache := memmap.PtrUint32(0x5D4594, 2487808)
				powerCache := memmap.PtrUint32(0x5D4594, 1569720)
				oldCache, oldPower := *cache, *powerCache
				oldServer, oldCaster, oldGame := noxServer, nox_xxx_imagCasterUnit_1569664, noxflags.GetGame()
				t.Cleanup(func() {
					*cache, *powerCache = oldCache, oldPower
					noxServer, nox_xxx_imagCasterUnit_1569664 = oldServer, oldCaster
					noxflags.UnsetGame(noxflags.GetGame())
					noxflags.SetGame(oldGame)
				})
				*cache, *powerCache = math.MaxUint32, uint32(caster.TypeInd)
				noxServer, nox_xxx_imagCasterUnit_1569664 = s, caster
				noxflags.UnsetGame(noxflags.GetGame())
				s.SetFrame(702)
				vm := &s.Server.NoxScriptVM
				if err := vm.ReadScript(bytes.NewReader(toxicCloudScript52DB60(t, position.source, position.target))); err != nil {
					t.Fatal(err)
				}
				if s.SpellPower4FE7B0(spell.SPELL_TOXIC_CLOUD, caster) != 1 {
					t.Fatal("script ImaginaryCaster did not select original default power")
				}
				if route == "direct-vm" {
					if err := vm.CallByIndex(2, nil, nil); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := vm.CallByIndex(3, nil, nil); err != nil {
						t.Fatal(err)
					}
					id := vm.PopU32()
					if id == 0 {
						t.Fatal("FrameTimer builtin did not schedule an activator")
					}
					s.SetFrame(704)
					vm.ActRun()
					if caster.PosVec != types.Ptf(250, 250) || vm.PopU32() != 0 {
						t.Fatal("script timer fired before its deadline")
					}
					s.SetFrame(705)
					vm.ActRun()
					if s.Activators.Cancel(id) {
						t.Fatal("script timer callback was not consumed at its deadline")
					}
				}
				if got := vm.PopU32(); got != toxicCloudScriptReturn52DB60 {
					t.Fatalf("map-script builtin did not complete: return=%08x", got)
				}
				if caster.PosVec != position.source || caster.NewPos != position.source || caster.PrevPos != position.source ||
					caster.Direction1 != server.DirFromVec(position.target.Sub(position.source)) ||
					*cache != math.MaxUint32 || *powerCache != 17 || nox_xxx_imagCasterUnit_1569664 != caster ||
					caster.FirstOwned516() != nil || s.Objs.Pending != nil || vm.Caller() != nil || vm.Trigger() != nil {
					t.Fatal("script cast changed its native caster, cache, coordinates, owner, or VM context")
				}
				vm.ActRun()
				if vm.PopU32() != 0 {
					t.Fatal("consumed script timer fired twice")
				}
				t.Logf("route=%s caster=%p source=%v target=%v power=1 completed=%08x", route, caster, position.source, position.target, toxicCloudScriptReturn52DB60)
			})
		}
	}
}
