package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestToxicCloudCNativePointers53D850(t *testing.T) {
	if value := os.Getenv("OPENNOX_TOXIC_CLOUD_NATIVE_PROBE"); value != "" {
		kind, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		cloudFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		targetFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		dataFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.ToxicCloudUpdateData{}))
		cloud := (*server.Object)(unsafe.Pointer(&cloudFrame[32]))
		target := (*server.Object)(unsafe.Pointer(&targetFrame[32]))
		data := (*server.ToxicCloudUpdateData)(unsafe.Pointer(&dataFrame[32]))
		*cloud = server.Object{Field34: 11, UpdateData: unsafe.Pointer(data)}
		*target = server.Object{}
		*data = server.ToxicCloudUpdateData{Duration: 2}
		beforeCloud := append([]byte(nil), cloudFrame...)
		beforeTarget := append([]byte(nil), targetFrame...)
		wantData := append([]byte(nil), dataFrame...)
		gameFrameHook = func() uint32 { return 10 }
		toxicCloudRawRuntime53D850 = func() server.ToxicCloudRawRuntime53D850 {
			return server.ToxicCloudRawRuntime53D850{Frame: gameFrameHook}
		}
		fmt.Printf("NATIVE_CLOUD_BEGIN kind=%d cloud=%#x target=%#x data=%#x\n", kind, uintptr(unsafe.Pointer(cloud)), uintptr(unsafe.Pointer(target)), uintptr(unsafe.Pointer(data)))
		toxicCloudNativeProbe53D850(kind, cloud, target)
		if kind == 0 || kind == 2 {
			binary.LittleEndian.PutUint32(wantData[32:], 1)
		}
		if !bytes.Equal(beforeCloud, cloudFrame) || !bytes.Equal(beforeTarget, targetFrame) || !bytes.Equal(wantData, dataFrame) {
			t.Fatal("native waiting/non-unit frames or canaries")
		}
		fmt.Println("NATIVE_CLOUD_OK")
		return
	}
	for kind, name := range []string{"cloud-root", "cloud-target", "small-root", "small-target"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestToxicCloudCNativePointers53D850$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENNOX_TOXIC_CLOUD_NATIVE_PROBE="+strconv.Itoa(kind))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "NATIVE_CLOUD_OK") {
				t.Fatal("native marker missing")
			}
		})
	}
}
