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

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestSparkTrailPushCNativePointers(t *testing.T) {
	if value := os.Getenv("OPENNOX_SPARK_TRAIL_PUSH_PROBE"); value != "" {
		kind, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&frame[32]))
		*source = server.Object{}
		dataFrame := radialCFrameOriginal(t, 16)
		data := unsafe.Pointer(&dataFrame[32])
		clear(dataFrame[32:48])
		source.UpdateData = data
		if kind == 0 {
			*(*server.SparkUpdateData)(data) = server.SparkUpdateData{LifetimeRemaining: 2, Kind: 4}
		}
		if kind == 2 {
			*(*server.PushUpdateData53B030)(data) = server.PushUpdateData53B030{Radius: 5, Force: 6}
		}
		wantObject, wantData := append([]byte(nil), frame...), append([]byte(nil), dataFrame...)
		if kind == 0 {
			binary.LittleEndian.PutUint32(wantObject[32+unsafe.Offsetof(source.Float28):], 0x3f800000)
			binary.LittleEndian.PutUint32(wantData[36:], 1)
		}
		created, pushed := 0, 0
		sparkTrailPushRawRuntime = func() sparkTrailPushRuntime {
			return sparkTrailPushRuntime{
				DelayedDelete: func(*server.Object) { t.Fatal("unexpected deletion") },
				Trail: server.ProjectileTrailRawRuntime53AEC0{
					DirectionFloat: func(int32) float32 { return 0 },
					RandomFloat:    func(float32, float32) float64 { return 0 },
					CreateSpark: func(p types.Pointf, k, l int, v types.Pointf, z float32, o *server.Object) *server.Object {
						if p != (types.Pointf{}) || v != (types.Pointf{}) || k != 1 || l != 6 || z != 0 || o != source {
							t.Fatal("spark arguments/owner")
						}
						created++
						return nil
					},
				},
				PushUnits: func(p *types.Pointf, a, b, c float32) {
					if p != &source.PosVec || a != 5 || b != 0 || c != 6 {
						t.Fatal("live push arguments/point")
					}
					pushed++
				},
			}
		}
		fmt.Printf("NATIVE_SPARK_TRAIL_PUSH_BEGIN kind=%d source=%#x data=%#x\n", kind, uintptr(unsafe.Pointer(source)), uintptr(data))
		got := sparkTrailPushProbe(kind, source)
		if got != nil || !bytes.Equal(wantObject, frame) || !bytes.Equal(wantData, dataFrame) {
			t.Fatal("native result/object/data/canary differs")
		}
		if (kind == 1 && created != 8) || (kind == 2 && pushed != 1) {
			t.Fatal("missing service calls")
		}
		fmt.Println("NATIVE_SPARK_TRAIL_PUSH_OK")
		return
	}
	for kind, name := range []string{"spark", "trail", "push"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestSparkTrailPushCNativePointers$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENNOX_SPARK_TRAIL_PUSH_PROBE="+strconv.Itoa(kind))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "NATIVE_SPARK_TRAIL_PUSH_OK") {
				t.Fatal("native marker")
			}
		})
	}
}
