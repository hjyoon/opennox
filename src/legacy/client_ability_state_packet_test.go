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
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// The actual C switch still consumes CE/CF/EB/CD. Exercise its genuine table
// and query function, not replacement callbacks. Positive reward UI is covered
// by the headless gameplay fixture; here CD removes an award or retains an
// already-owned level, which must not open the book.
func TestClientAbilityPacketsActualDispatcher(t *testing.T) {
	const child = "OPENNOX_TEST_CLIENT_ABILITY_PACKETS"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestClientAbilityPacketsActualDispatcher$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated actual ability packets: %v\n%s", err, output)
		}
		t.Log(strings.TrimSpace(string(output)))
		return
	}
	InitBlobData()
	rows := memmap.Slice(0x5D4594, 1047764)[:192]
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&rows[0])) <= math.MaxUint32 {
		t.Fatalf("native ability table below 4 GiB: %p", &rows[0])
	}
	before := bytes.Clone(rows)
	oldConnected, oldFrame, oldGame := Nox_client_isConnected, gameFrameHook, noxflags.GetGame()
	defer func() {
		copy(rows, before)
		Nox_client_isConnected, gameFrameHook = oldConnected, oldFrame
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
	}()
	noxflags.ResetGame() // These table/packet checks are not host reward UI.
	cases := 0
	for layout := 0; layout < 5; layout++ {
		seed := bytes.Clone(before)
		for slot := 1; slot <= 5; slot++ {
			row := seed[24*slot : 24*(slot+1)]
			for i := range row {
				row[i] = 0x5a
			}
			id := uint32((slot+layout-1)%5 + 1)
			binary.LittleEndian.PutUint32(row, id)
			binary.LittleEndian.PutUint32(row[8:], 0x80000000+uint32(slot))
			binary.LittleEndian.PutUint32(row[12:], 0xffffff80|1<<id)
			binary.LittleEndian.PutUint32(row[16:], uint32(2+slot))
			binary.LittleEndian.PutUint32(row[20:], 0xfedcba00+uint32(slot))
		}
		for _, connected := range []bool{false, true} {
			Nox_client_isConnected = func() bool { return connected }
			for _, frame := range []uint32{0, 0x7fffffff, 0x80000000, 0xffffffff} {
				gameFrameHook = func() uint32 { return frame }
				for _, op := range []netmsg.Op{netmsg.MSG_REPORT_ABILITY_STATE, netmsg.MSG_REPORT_ACTIVE_ABILITIES, netmsg.MSG_RESET_ABILITIES, netmsg.MSG_REPORT_ABILITY_AWARD} {
					values := []byte{0, 1, 0x80, 0xff}
					if op == netmsg.MSG_RESET_ABILITIES {
						values = []byte{0}
					} else if op == netmsg.MSG_REPORT_ABILITY_AWARD {
						values = []byte{0, 0x80, 1, 0x81}
					}
					for id := 0; id <= 6; id++ {
						for _, value := range values {
							for _, offset := range []int{1, 2} {
								copy(rows, seed)
								if op == netmsg.MSG_REPORT_ABILITY_AWARD && value&0x7f != 0 {
									for slot := 1; slot <= 5; slot++ {
										binary.LittleEndian.PutUint32(rows[24*slot+16:], uint32(value&0x7f))
									}
								}
								want := bytes.Clone(rows)
								size := 3
								if op == netmsg.MSG_RESET_ABILITIES {
									size = 2
								}
								storage, free := alloc.Make([]byte{0xa5}, size+8)
								for i := range storage {
									storage[i] = 0xa5
								}
								packet := storage[offset : offset+size]
								packet[0], packet[1] = byte(op), byte(id)
								if size == 3 {
									packet[2] = value
								}
								packetBefore := bytes.Clone(storage)
								if connected {
									key := uint32(id)
									if op == netmsg.MSG_RESET_ABILITIES && id != 6 {
										key = binary.LittleEndian.Uint32(want[24*id:])
									}
									for slot := 1; slot <= 5; slot++ {
										row := want[24*slot:]
										if binary.LittleEndian.Uint32(row) != key && !(op == netmsg.MSG_RESET_ABILITIES && id == 6) {
											continue
										}
										switch op {
										case netmsg.MSG_REPORT_ABILITY_STATE:
											binary.LittleEndian.PutUint32(row[8:], uint32(value))
											start := uint32(0)
											if value == 0 {
												start = frame
											}
											binary.LittleEndian.PutUint32(row[20:], start)
										case netmsg.MSG_REPORT_ACTIVE_ABILITIES:
											mask := binary.LittleEndian.Uint32(row[12:])
											if value == 0 {
												mask &^= 1 << id
											} else {
												mask |= 1 << id
											}
											binary.LittleEndian.PutUint32(row[12:], mask)
										case netmsg.MSG_RESET_ABILITIES:
											binary.LittleEndian.PutUint32(row[8:], 1)
											binary.LittleEndian.PutUint32(row[20:], 0)
										case netmsg.MSG_REPORT_ABILITY_AWARD:
											if id >= 1 && id < 6 {
												binary.LittleEndian.PutUint32(row[16:], uint32(value&0x7f))
											}
										}
									}
								}
								if got := Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, op, packet); got != size {
									free()
									t.Fatalf("actual ability packet op%d id%d consumed=%d want=%d", op, id, got, size)
								}
								if !bytes.Equal(rows, want) || !bytes.Equal(storage, packetBefore) {
									free()
									t.Fatalf("actual ability packet op%d id%d layout%d connected=%t frame=%#x value=%#x offset%d changed table/padding/packet", op, id, layout, connected, frame, value, offset)
								}
								for key := 0; key <= 6; key++ {
									active := false
									for slot := 1; slot <= 5; slot++ {
										row := want[24*slot:]
										if binary.LittleEndian.Uint32(row) == uint32(key) {
											active = binary.LittleEndian.Uint32(row[12:])&(1<<key) != 0
											break
										}
									}
									if clientAbilityActive461160(key) != active || !bytes.Equal(rows, want) {
										free()
										t.Fatalf("actual ability query id%d after op%d id%d", key, op, id)
									}
								}
								free()
								cases++
							}
						}
					}
				}
			}
		}
	}
	if cases != 7280 {
		t.Fatalf("actual ability packet corpus=%d", cases)
	}
	t.Logf("actual ability packets: %d; four active C opcodes, every row, connection/frame/byte boundaries, genuine queries and all table/packet bytes checked", cases)
}
