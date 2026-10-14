package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Exercise the complete C dispatcher and real 32-slot ally table. Only
// frontend connection/debug state is supplied; no lookup or HP callback
// is substituted. Keep the original raw-ID versus normalized local-ID gates.
func TestClientAllyHealthPacketsActualDispatcher(t *testing.T) {
	const child = "OPENNOX_TEST_ALLY_HEALTH_PACKETS"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestClientAllyHealthPacketsActualDispatcher$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated actual ally-health packets: %v\n%s", err, output)
		}
		t.Log(strings.TrimSpace(string(output)))
		return
	}
	InitBlobData()
	rows := clientAllyHealthTable495020()
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&rows[0])) <= math.MaxUint32 {
		t.Fatalf("native ally table below 4 GiB: %p", &rows[0])
	}
	oldConnected, oldCode, oldEngine := Nox_client_isConnected, ClientPlayerNetCode(), noxflags.GetEngine()
	defer func() {
		Nox_client_isConnected = oldConnected
		ClientSetPlayerNetCode(oldCode)
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	}()
	ClientSetPlayerNetCode(0x7ffe) // No case is a local-player HUD packet.
	cases := 0
	for _, mark := range []uint16{0, 0x8000} {
		seed := bytes.Repeat([]byte{0x5a}, 512)
		for slot := 0; slot < 32; slot++ {
			binary.LittleEndian.PutUint32(seed[slot*16+12:], 0)
		}
		copy(rows, seed)
		for slot := 0; slot < 32; slot++ {
			code := uint16(0x4000+slot) | mark
			if got := Sub_495060(int(code), 100+slot, 500+slot); got != 1 {
				t.Fatalf("real monitor insertion slot%d: %d", slot, got)
			}
			if got := clientAllyHealthColor4950F0(uint32(code), byte(0xff-slot)); got != 1 {
				t.Fatalf("real alternate-color slot%d: %d", slot, got)
			}
			row := seed[slot*16:]
			binary.LittleEndian.PutUint32(row, uint32(code))
			row[4] = byte(0xff - slot)
			binary.LittleEndian.PutUint16(row[6:], uint16(100+slot))
			binary.LittleEndian.PutUint16(row[8:], uint16(500+slot))
			binary.LittleEndian.PutUint32(row[12:], 1)
		}
		if !bytes.Equal(rows, seed) || clientAllyHealthFree494FF0() != nil {
			t.Fatal("real 32-slot monitor population differs from original scalar layout")
		}
		for _, connected := range []bool{false, true} {
			Nox_client_isConnected = func() bool { return connected }
			for _, debug := range []bool{false, true} {
				noxflags.ResetEngine()
				if debug {
					noxflags.SetEngine(noxflags.EngineNetDebug)
				}
				for _, op := range []netmsg.Op{netmsg.MSG_REPORT_HEALTH, netmsg.MSG_REPORT_TOTAL_HEALTH} {
					for slot := 0; slot <= 32; slot++ {
						for _, value := range []uint16{0, 0x7fff, 0x8000, 0xffff} {
							for _, offset := range []int{1, 2} {
								copy(rows, seed)
								size := 4
								if op == netmsg.MSG_REPORT_TOTAL_HEALTH {
									size = 7
								}
								storage, free := alloc.Make([]byte{0xa5}, size+8)
								for i := range storage {
									storage[i] = 0xa5
								}
								packet := storage[offset : offset+size]
								packet[0] = byte(op)
								code := uint16(0x4000+slot) | mark
								binary.LittleEndian.PutUint16(packet[1:], code)
								if op == netmsg.MSG_REPORT_HEALTH {
									packet[3] = byte(value)
								} else {
									binary.LittleEndian.PutUint16(packet[3:], value)
									binary.LittleEndian.PutUint16(packet[5:], ^value)
								}
								before := bytes.Clone(storage)
								want := bytes.Clone(seed)
								found := connected && slot < 32
								if found {
									cur := 2 * uint16(byte(value))
									if op == netmsg.MSG_REPORT_TOTAL_HEALTH {
										cur = value
										binary.LittleEndian.PutUint16(want[slot*16+8:], ^value)
									}
									binary.LittleEndian.PutUint16(want[slot*16+6:], cur)
								}
								if got := Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, op, packet); got != size {
									free()
									t.Fatalf("actual packet op%d code%#x consumed=%d want=%d", op, code, got, size)
								}
								if !bytes.Equal(rows, want) || !bytes.Equal(storage, before) {
									free()
									t.Fatalf("actual packet op%d slot%d high=%#x connected=%t debug=%t value=%#x offset%d changed table/padding/packet", op, slot, mark, connected, debug, value, offset)
								}
								cur, max, alt, ok := Sub_495180(int(code))
								if slot < 32 {
									if !ok || cur != int(binary.LittleEndian.Uint16(want[slot*16+6:])) ||
										max != int(binary.LittleEndian.Uint16(want[slot*16+8:])) || !alt {
										free()
										t.Fatalf("actual native ally getter op%d slot%d=%d/%d alt=%t ok=%t", op, slot, cur, max, alt, ok)
									}
								} else if cur != 0 || max != 0 || alt || ok {
									free()
									t.Fatalf("missing ally getter created data: %d/%d alt=%t ok=%t", cur, max, alt, ok)
								}
								free()
								cases++
							}
						}
					}
				}
			}
		}
		copy(rows, seed)
		for slot := 31; slot >= 0; slot-- {
			code := uint16(0x4000+slot) | mark
			if Sub_4950C0(int(code)) != 1 {
				t.Fatalf("real unmonitor slot%d", slot)
			}
			binary.LittleEndian.PutUint32(seed[slot*16+12:], 0)
			if !bytes.Equal(rows, seed) {
				t.Fatalf("unmonitor slot%d changed unrelated fields", slot)
			}
		}
		reset := clientAllyHealthReset4951C0()
		// The original return is table+518 in the containing memmap blob.
		// Compare addresses without constructing an out-of-slice pointer.
		if uintptr(reset) != uintptr(unsafe.Pointer(&rows[0]))+518 {
			t.Fatalf("reset native return=%p", reset)
		}
		for slot := 0; slot < 32; slot++ {
			row := seed[slot*16:]
			binary.LittleEndian.PutUint32(row, 0)
			row[4] = 0
			binary.LittleEndian.PutUint16(row[6:], 0)
			binary.LittleEndian.PutUint16(row[8:], 0)
			binary.LittleEndian.PutUint32(row[12:], 0)
		}
		if !bytes.Equal(rows, seed) {
			t.Fatal("actual reset did not clear all records while preserving padding")
		}
	}
	if cases != 4224 {
		t.Fatalf("actual packet corpus=%d", cases)
	}
	t.Logf("actual ally-health packets: %d; both active C opcodes, every slot, raw IDs, debug/connection, native getters, packet guards and reset verified", cases)
}
