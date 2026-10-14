package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type deathMessagePrinter4952E0 struct {
	colors []console.Color
	lines  []string
}

func (p *deathMessagePrinter4952E0) Print(color console.Color, line string) {
	p.colors = append(p.colors, color)
	p.lines = append(p.lines, line)
}

func (p *deathMessagePrinter4952E0) Printf(color console.Color, format string, args ...interface{}) {
	p.Print(color, fmt.Sprintf(format, args...))
}

// Use the genuine C network dispatcher, CGo player registry/localization,
// UTF-16 formatter, console export and packet ring. Only the output printer
// is observed. No fabricated Player/name lookup or packet-processing hook.
func TestClientDeathMessage4952E0ActualPacketDispatcher(t *testing.T) {
	const child = "OPENNOX_TEST_DEATH_MESSAGE_4952E0"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestClientDeathMessage4952E0ActualPacketDispatcher$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated death-message dispatcher: %v\n%s", err, output)
		}
		t.Log(strings.TrimSpace(string(output)))
		return
	}
	// Start the actual original constant blobs in a child, without clearing
	// packed pointer side slots or changing any other parent test's state.
	InitBlobData()
	for _, tc := range []struct {
		name                string
		attacker, assistant uint16
		want                string
	}{
		{"environment", 0, 0x8000, "[피해자🚀] < 나쁜 것 >"},
		{"missing attacker", 0x1234, 0x8000, "[피해자🚀] < 나쁜 것 >"},
		{"attacker", 0x7fff, 0, "[피해자🚀] < 공격자🚀 >"},
		{"missing assistant", 0x7fff, 0x1234, "[피해자🚀] < 공격자🚀 >"},
		{"assistant", 0x7fff, 0x8000, "[피해자🚀] < 공격자🚀 > + 조력자🚀"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := playerIDNativeFixture417040(t)
			if err := s.Strings().ReadJSON("testdata/client_death_message_4952e0_strings.json"); err != nil {
				t.Fatal(err)
			}
			s.SetFrame(0xf1234567)
			slots := s.Players.ListSlots()
			for _, p := range slots {
				*p = server.Player{}
			}
			for i, spec := range []struct {
				code uint32
				name string
			}{{0x7fff, "공격자🚀"}, {0x8000, "조력자🚀"}, {0xffff, "피해자🚀"}} {
				p := slots[i+1]
				*p = server.Player{Active: 1, NetCodeVal: spec.code, PlayerInd: byte(i + 1)}
				p.SetName(spec.name)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= 0xffffffff {
					t.Fatalf("Player fixture not above 4 GiB: %p", p)
				}
			}
			before := make([][]byte, len(slots))
			for i, p := range slots {
				before[i] = playerIDBytes417040(p)
			}
			printer := new(deathMessagePrinter4952E0)
			cn := console.New(printer)
			oldConsole := GetConsole
			GetConsole = func() *console.Console { return cn }
			t.Cleanup(func() { GetConsole = oldConsole })
			next, oldest := deathMessageRingIndexes495210()
			t.Cleanup(func() { setDeathMessageRingIndexes495210(next, oldest) })
			setDeathMessageRingIndexes495210(99, 0)
			ring := memmap.PtrT[[600]uint32](0x5d4594, 1201428)
			oldRing := *ring
			t.Cleanup(func() { *ring = oldRing })
			for i := range ring {
				ring[i] = 0xa5a5a5a5
			}
			packet := []byte{byte(netmsg.MSG_INFORM), 0x0e, 0, 0, 0, 0, 0xff, 0xff, 0x34, 0x12, 2}
			binary.LittleEndian.PutUint16(packet[2:4], tc.attacker)
			binary.LittleEndian.PutUint16(packet[4:6], tc.assistant)
			packetBefore := bytes.Clone(packet)
			if got := Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, netmsg.MSG_INFORM, packet); got != 11 {
				t.Fatalf("MSG_INFORM death consumed %d bytes, want 11", got)
			}
			if !reflect.DeepEqual(printer.lines, []string{tc.want}) || !reflect.DeepEqual(printer.colors, []console.Color{console.ColorWhite}) {
				t.Fatalf("death console = %#v/%v, want %q/white", printer.lines, printer.colors, tc.want)
			}
			wantRing := [600]uint32{}
			for i := range wantRing {
				wantRing[i] = 0xa5a5a5a5
			}
			copy(wantRing[594:], []uint32{uint32(tc.attacker), uint32(tc.assistant), 0xffff, 0x1234, 2, 0xf1234567})
			if *ring != wantRing || !bytes.Equal(packet, packetBefore) {
				t.Fatal("packet or ring changed outside original death-event stores")
			}
			if next, oldest := deathMessageRingIndexes495210(); next != 0 || oldest != 1 {
				t.Fatalf("full ring did not wrap/advance oldest: %d/%d", next, oldest)
			}
			for i, p := range slots {
				if !bytes.Equal(before[i], playerIDBytes417040(p)) {
					t.Fatalf("Player %d changed during name formatting", i)
				}
			}
			t.Logf("actual MSG_INFORM death: Player=%p name=%p UTF-16=%q ring99->0 oldest0->1", slots[3], &slots[3].NameFinal[0], printer.lines[0])
		})
	}
}
