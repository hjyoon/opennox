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

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

func TestSentryGlobeCNativePointers(t *testing.T) {
	if value := os.Getenv("OPENNOX_SENTRY_GLOBE_PROBE"); value != "" {
		kind, err := strconv.Atoi(value)
		if err != nil {
			t.Fatal(err)
		}
		frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
		source := (*server.Object)(unsafe.Pointer(&frame[32]))
		*source = server.Object{PosVec: types.Ptf(10, 20), Pos39: types.Ptf(30, 40)}
		df := radialCFrameOriginal(t, 12)
		data := (*server.SentryUpdateData)(unsafe.Pointer(&df[32]))
		*data = server.SentryUpdateData{Field4: 0x3f800000}
		source.UpdateData = unsafe.Pointer(data)
		if kind == 1 {
			source.ObjFlags = object.FlagMarked
		}
		if kind == 2 {
			source.ObjFlags = object.FlagBelow
		}
		world, _ := commonRandomFloatFixture416030(t, 0)
		world.NetList = netlist.New()
		world.NetList.Init()
		h := world.SentryGlobeRuntime510E60()
		if kind == 4 || kind == 5 {
			h.List.SetHead(source)
		}
		if kind == 5 {
			pf := radialCFrameOriginal(t, unsafe.Sizeof(server.Player{}))
			player := (*server.Player)(unsafe.Pointer(&pf[32]))
			*player = server.Player{Field10: 20, Field12: 20, Pos3632Vec: types.Ptf(100, 100)}
			h.Player = func(int32) *server.Player { return player }
		}
		old := sentryGlobeRawRuntime510E60
		defer func() { sentryGlobeRawRuntime510E60 = old }()
		sentryGlobeRawRuntime510E60 = func() server.SentryGlobeRawRuntime510E60 { return h }
		wantObject, wantData := append([]byte(nil), frame...), append([]byte(nil), df...)
		if kind == 0 {
			binary.LittleEndian.PutUint32(wantObject[32+unsafe.Offsetof(source.ObjFlags):], uint32(source.ObjFlags|object.FlagMarked))
		}
		if kind == 0 || kind == 5 {
			binary.LittleEndian.PutUint32(wantData[32:], data.Field4)
		}
		if kind == 1 {
			binary.LittleEndian.PutUint32(wantObject[32+unsafe.Offsetof(source.ObjFlags):], uint32(source.ObjFlags&^object.FlagMarked))
		}
		fmt.Printf("NATIVE_SENTRY_BEGIN kind=%d source=%#x data=%#x\n", kind, uintptr(unsafe.Pointer(source)), uintptr(unsafe.Pointer(data)))
		result := sentryGlobeProbe(kind, source, nil)
		switch kind {
		case 0:
			if !source.ObjFlags.Has(object.FlagMarked) || data.Field0 != data.Field4 {
				t.Fatal("native list insertion/disabled angle")
			}
		case 1:
			if result != source || source.ObjFlags.Has(object.FlagMarked) {
				t.Fatal("native unlink/return identity")
			}
		case 3:
			packet := world.NetList.CopyPacketsA(ntype.PlayerInd(2), netlist.Kind1)
			if !bytes.Equal(packet, []byte{149, 10, 0, 20, 0, 30, 0, 40, 0}) {
				t.Fatalf("native ray packet %x", packet)
			}
		case 4:
			if h.List.Head() != nil {
				t.Fatal("C reset retained the shared Go head")
			}
		}
		if !bytes.Equal(wantObject, frame) || !bytes.Equal(wantData, df) {
			t.Fatal("native object/data frame or canary")
		}
		fmt.Println("NATIVE_SENTRY_OK")
		return
	}
	for kind, name := range []string{"update", "unlink", "candidate", "packet", "reset", "send"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestSentryGlobeCNativePointers$", "-test.count=1")
			cmd.Env = append(os.Environ(), "OPENNOX_SENTRY_GLOBE_PROBE="+strconv.Itoa(kind))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("native child:\n%s", out)
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "NATIVE_SENTRY_OK") {
				t.Fatal("native marker")
			}
		})
	}
}
