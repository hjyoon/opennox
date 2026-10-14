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

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Run real selector -> reset -> code encoder -> C sender -> Go host queue.
// Only the host/list and inventory input records are prepared; no reset,
// pointer lookup, packet encoding or send callback is substituted.
func TestClientSecondaryActualHostPacketQueue(t *testing.T) {
	const child = "OPENNOX_TEST_SECONDARY_GRID_PACKETS"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestClientSecondaryActualHostPacketQueue$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated actual secondary packets: %v\n%s", err, output)
		}
		t.Log(strings.TrimSpace(string(output)))
		return
	}
	InitBlobData()
	grid, layout := clientSecondaryGridLayout198()
	stride, cells, codeOff, equipOff, markerOff, countOff := layout[0], layout[1], layout[2], layout[3], layout[4], layout[5]
	drawSize, drawCodeOff, drawFlagsOff := layout[6], layout[7], layout[8]
	pointerSize := int(unsafe.Sizeof(uintptr(0)))
	if cells != 84 || stride != 148+pointerSize-4 || markerOff != 136+pointerSize-4 || countOff != 140+pointerSize-4 {
		t.Fatalf("native inventory header layout=%v pointer=%d", layout, pointerSize)
	}
	base := uintptr(unsafe.Pointer(&grid[0]))
	if pointerSize > 4 && base <= math.MaxUint32 {
		t.Fatalf("native grid below 4 GiB: %#x", base)
	}
	drawables, free := alloc.Make([]byte{0}, cells*drawSize)
	defer free()
	before := bytes.Clone(grid)
	oldSelected, oldCode := clientSecondaryGridGlobals198()
	oldServer, oldFlags := GetServer, noxflags.GetGame()
	defer func() {
		copy(grid, before)
		clientSecondaryGridRestore198(oldSelected, oldCode)
		GetServer = oldServer
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	}()
	list := netlist.New()
	list.Init()
	defer list.Free()
	GetServer = func() Server { return &netClientSendTestServer{srv: &server.Server{NetList: list}} }
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	codes := []uint32{0, 0x7fff, 0x8000, 0xffffffff}
	flags := []uint32{0, 0x00400000, 0x20000000}
	amounts := []byte{0, 1, 32, 255}
	cases := 0
	for pattern := 0; pattern < 4; pattern++ {
		for _, previous := range []int{-1, 0, 41, 83} {
			for selected := -1; selected < cells; selected++ {
				for i := range grid {
					grid[i] = 0x5a
				}
				for i := range drawables {
					drawables[i] = 0x69
				}
				for slot := 0; slot < cells; slot++ {
					row := grid[slot*stride:]
					ptr := uintptr(unsafe.Pointer(&drawables[slot*drawSize]))
					if pointerSize > 4 {
						if ptr <= math.MaxUint32 {
							t.Fatal("native drawable below 4 GiB")
						}
						binary.LittleEndian.PutUint64(row, uint64(ptr))
					} else {
						binary.LittleEndian.PutUint32(row, uint32(ptr))
					}
					binary.LittleEndian.PutUint32(row[codeOff:], codes[(slot+pattern)%4])
					binary.LittleEndian.PutUint32(row[equipOff:], 0xdead0000+uint32(slot))
					binary.LittleEndian.PutUint32(row[markerOff:], 0x80000000+uint32(slot))
					row[countOff] = amounts[(slot+pattern)%4]
					binary.LittleEndian.PutUint32(drawables[slot*drawSize+drawCodeOff:], 0xfedc0000+uint32(slot))
					binary.LittleEndian.PutUint32(drawables[slot*drawSize+drawFlagsOff:], flags[(slot+pattern)%3])
				}
				clientSecondaryGridPrepare198(previous)
				wantGrid, wantDraw := bytes.Clone(grid), bytes.Clone(drawables)
				wantPrevious := uint32(0)
				if previous >= 0 {
					wantPrevious = binary.LittleEndian.Uint32(grid[previous*stride+codeOff:])
				}
				for slot := 0; slot < cells; slot++ {
					if wantGrid[slot*stride+countOff] > 0 {
						binary.LittleEndian.PutUint32(wantGrid[slot*stride+markerOff:], 0)
					}
				}
				wantSelected, wire := uintptr(0), uint16(0)
				if selected >= 0 {
					wantSelected = base + uintptr(selected*stride)
					code := binary.LittleEndian.Uint32(grid[selected*stride+codeOff:])
					binary.LittleEndian.PutUint32(wantGrid[selected*stride+markerOff:], 1)
					binary.LittleEndian.PutUint32(wantDraw[selected*drawSize+drawCodeOff:], code)
					if code < 0x8000 {
						wire = uint16(code)
						if binary.LittleEndian.Uint32(drawables[selected*drawSize+drawFlagsOff:])&0x20400000 != 0 {
							wire |= 0x8000
						}
					}
				}
				if got := clientSecondaryGridSelect198(selected); got != 1 {
					t.Fatalf("actual selector return=%d", got)
				}
				gotSelected, gotPrevious := clientSecondaryGridGlobals198()
				wantPacket := []byte{0xe0, byte(wire), byte(wire >> 8)}
				if gotSelected != wantSelected || gotPrevious != wantPrevious || !bytes.Equal(grid, wantGrid) || !bytes.Equal(drawables, wantDraw) ||
					!bytes.Equal(list.CopyPacketsA(31, netlist.Kind0), wantPacket) {
					t.Fatalf("actual secondary selection pattern%d previous%d selected%d: pointer/code/whole state/three-byte host packet contract", pattern, previous, selected)
				}
				if len(list.CopyPacketsA(31, netlist.Kind0)) != 0 {
					t.Fatal("secondary selection emitted more than one packet")
				}
				cases++
			}
		}
	}
	if cases != 1360 {
		t.Fatalf("actual secondary packet corpus=%d", cases)
	}
	t.Logf("actual secondary host packets: %d; every cell/nil, native pointer, previous DWORD, 84 markers/empty flags, code limits/static bits, whole state and genuine C-to-Go queue verified", cases)
}
