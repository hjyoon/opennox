package legacy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestAntiSpellCNativePointers53BB00(t *testing.T) {
	names := []string{"root", "candidate"}
	if child := os.Getenv("OPENNOX_ANTI_SPELL_CABI_PROBE"); child != "" {
		kind := 0
		if child == "candidate" {
			kind = 1
		} else if child != "root" {
			t.Fatal("unknown C probe")
		}
		mframe := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		cframe := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		uframe := radialCFrameOriginal(t, unsafe.Sizeof(server.MissileUpdateData{}))
		sframe := radialCFrameOriginal(t, unsafe.Sizeof(server.AntiSpellTargetState53BD10{}))
		missile := (*server.Object)(unsafe.Pointer(&mframe[32]))
		candidate := (*server.Object)(unsafe.Pointer(&cframe[32]))
		update := (*server.MissileUpdateData)(unsafe.Pointer(&uframe[32]))
		state := (*server.AntiSpellTargetState53BD10)(unsafe.Pointer(&sframe[32]))
		*update = server.MissileUpdateData{}
		*state = server.AntiSpellTargetState53BD10{}
		*missile = server.Object{UpdateData: unsafe.Pointer(update)}
		*candidate = server.Object{}
		deleted := false
		antiSpellProjectileRuntime53BB00 = func() server.AntiSpellProjectileRuntime53BB00 {
			return server.AntiSpellProjectileRuntime53BB00{TargetState: state, Frame: func() uint32 { return 1000 }, FPS: func() uint32 { return 30 }, CanInteract: func(a, b *server.Object) bool {
				if a != missile || b != candidate {
					t.Fatal("truncated callback arguments")
				}
				return false
			}, DelayedDelete: func(o *server.Object) {
				if o != missile {
					t.Fatal("truncated expiry pointer")
				}
				deleted = true
			}}
		}
		fmt.Printf("NATIVE_ANTI_SPELL_CABI_BEGIN kind=%s missile=%#x candidate=%#x\n", child, uintptr(unsafe.Pointer(missile)), uintptr(unsafe.Pointer(candidate)))
		antiSpellCProbe53BB00(kind, missile, candidate)
		if deleted != (kind == 0) {
			t.Fatal("unexpected expiry result")
		}
		fmt.Printf("NATIVE_ANTI_SPELL_CABI_OK kind=%s\n", child)
		return
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestAntiSpellCNativePointers53BB00$", "-test.count=1")
			command.Env = append(os.Environ(), "OPENNOX_ANTI_SPELL_CABI_PROBE="+name)
			out, err := command.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatalf("actual C %s pointer boundary: %v", name, err)
			}
			if !strings.Contains(string(out), "NATIVE_ANTI_SPELL_CABI_OK kind="+name) {
				t.Fatal("native child success marker missing")
			}
		})
	}
}
