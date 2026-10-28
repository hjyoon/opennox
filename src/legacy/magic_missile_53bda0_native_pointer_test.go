package legacy

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestMagicMissileCNativePointer53BDA0(t *testing.T) {
	if os.Getenv("OPENNOX_MAGIC_MISSILE_CABI_PROBE") == "1" {
		mframe := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		oframe := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		uframe := radialCFrameOriginal(t, unsafe.Sizeof(server.MissileUpdateData{}))
		missile := (*server.Object)(unsafe.Pointer(&mframe[32]))
		owner := (*server.Object)(unsafe.Pointer(&oframe[32]))
		update := (*server.MissileUpdateData)(unsafe.Pointer(&uframe[32]))
		*owner = server.Object{ObjFlags: 0x20}
		*update = server.MissileUpdateData{Owner: owner}
		*missile = server.Object{UpdateData: unsafe.Pointer(update), Collide: magicMissileCollisionPointer53BDA0()}
		magicMissileUpdateRuntime53BDA0 = func() server.MagicMissileUpdateRuntime53BDA0 {
			return server.MagicMissileUpdateRuntime53BDA0{
				Frame:         func() uint32 { t.Fatal("destroyed owner must precede frame read"); return 0 },
				CollideResult: func(m *server.Object) int32 { return int32(ccall.CallIntPtr3(m.Collide, m.CObj(), nil, nil)) },
			}
		}
		fmt.Printf("NATIVE_MAGIC_MISSILE_CABI_BEGIN missile=%#x owner=%#x update=%#x\n", uintptr(unsafe.Pointer(missile)), uintptr(unsafe.Pointer(owner)), uintptr(unsafe.Pointer(update)))
		if got := uint32(magicMissileCProbe53BDA0(missile)); got != 0x89abcdef {
			t.Fatalf("collision result %#x", got)
		}
		if magicMissileCollisionSeen53BDA0(0) != uintptr(unsafe.Pointer(missile)) || magicMissileCollisionSeen53BDA0(1) != 0 || magicMissileCollisionSeen53BDA0(2) != 0 {
			t.Fatal("actual indirect C collision arguments")
		}
		fmt.Println("NATIVE_MAGIC_MISSILE_CABI_OK")
		return
	}
	command := exec.Command(os.Args[0], "-test.run=^TestMagicMissileCNativePointer53BDA0$", "-test.count=1")
	command.Env = append(os.Environ(), "OPENNOX_MAGIC_MISSILE_CABI_PROBE=1")
	out, err := command.CombinedOutput()
	if err != nil {
		t.Logf("native child:\n%s", out)
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "NATIVE_MAGIC_MISSILE_CABI_OK") {
		t.Fatal("native child success marker missing")
	}
}
