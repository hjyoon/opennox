package legacy

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestLifetimeEffectCNativePointers53CC90(t *testing.T) {
	if v := os.Getenv("OPENNOX_LIFETIME_CABI_PROBE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatal(err)
		}
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		ownerFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&frame[32]))
		owner := (*server.Object)(unsafe.Pointer(&ownerFrame[32]))
		*source = server.Object{ZVal: 1}
		*owner = server.Object{ObjFlags: 0x20}
		if n == 3 {
			source.ObjOwner = owner
		}
		gameFrameHook = func() uint32 { return 0 }
		gameFPSHook = func() uint32 { return 30 }
		lifetimeEffectRuntime53CC90 = func() server.LifetimeEffectRuntime53CC90 {
			return server.LifetimeEffectRuntime53CC90{
				Frame: gameFrameHook, FPS: gameFPSHook, DelayedDelete: func(o *server.Object) {
					if o != source {
						t.Fatal("delete native identity")
					}
				},
				BuffOff: func(*server.Object, server.EnchantID) {},
				SpellSound: func(id spell.ID, kind int) sound.ID {
					if id != spell.SPELL_TELEKINESIS || kind != 2 {
						t.Fatal("off-sound arguments")
					}
					return 48
				},
				Audio: func(sound.ID, *server.Object) {},
			}
		}
		fmt.Printf("NATIVE_LIFETIME_CABI_BEGIN kind=%d source=%#x owner=%#x\n", n, uintptr(unsafe.Pointer(source)), uintptr(unsafe.Pointer(owner)))
		lifetimeEffectCProbe53CC90(server.LifetimeEffectKind53CC90(n), source)
		fmt.Println("NATIVE_LIFETIME_CABI_OK")
		return
	}
	for n, name := range []string{"SelfDestruct", "DeathBallFragment", "Moonglow", "Telekinesis", "Fist", "FlameCleanse"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestLifetimeEffectCNativePointers53CC90$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENNOX_LIFETIME_CABI_PROBE="+strconv.Itoa(n))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "NATIVE_LIFETIME_CABI_OK") {
				t.Fatal("native child marker missing")
			}
		})
	}
}
