package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type game2BookTestClient struct {
	Client
	sequence uint
}

func (c *game2BookTestClient) GetInputSeq() uint { return c.sequence }

func TestGame2BookSelectedRowsNativePointers(t *testing.T) {
	oldTrap := *memmap.PtrUint8(0x5D4594, 1048140)
	defer func() { *memmap.PtrUint8(0x5D4594, 1048140) = oldTrap }()
	backing := make([]byte, 256)
	for i := range backing {
		backing[i] = 0xa7
	}
	var pin runtime.Pinner
	pin.Pin(&backing[0])
	defer pin.Unpin()
	base := unsafe.Pointer(&backing[0])
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(base) <= math.MaxUint32 {
		t.Fatal("Go owner must be above4GiB")
	}
	for _, row := range []byte{0, 1, 2, 3, 4, 5, 127, 255} {
		backing[200] = row
		before := append([]byte(nil), backing...)
		for _, slot := range []int{-1, 0, 1, 4, 5, math.MaxInt32} {
			t.Run(fmt.Sprintf("row%d/slot%d", row, slot), func(t *testing.T) {
				runtime.GC()
				got := game2BookTestRow(false, base, slot, 1)
				want := game2BookTestRow(true, base, slot, 1)
				if got != want {
					t.Fatalf("native slot %p != reference %p", got, want)
				}
				if row < 5 && slot >= 0 && slot < 5 && got != unsafe.Add(base, 40*int(row)+8*slot) {
					t.Fatal("slot stride/address")
				}
				if !bytes.Equal(backing, before) {
					t.Fatal("row helper mutated data/canary")
				}
			})
		}
		if game2BookTestRow(false, base, 0, 0) != game2BookTestRow(true, base, 0, 0) {
			t.Fatal("row ABI")
		}
		*memmap.PtrUint8(0x5D4594, 1048140) = row
		if game2BookTestRow(false, nil, 0, 2) != game2BookTestRow(true, nil, 0, 2) {
			t.Fatal("trap row")
		}
	}
	if game2BookTestRow(false, nil, 0, 0) != nil || game2BookTestRow(false, nil, 0, 1) != nil {
		t.Fatal("nil row")
	}
	runtime.KeepAlive(backing)
}

func TestGame2BookAnimationNativeLayoutAndPinnedOwners(t *testing.T) {
	client := &game2BookTestClient{}
	oldClient := GetClient
	GetClient = func() Client { return client }
	defer func() { GetClient = oldClient }()
	type ref struct {
		Name [96]byte
		Data unsafe.Pointer
		Tags [4]byte
	}
	type anim struct {
		End    unsafe.Pointer
		Images unsafe.Pointer
		Tags   [4]byte
		Frame  uint32
	}
	if unsafe.Sizeof(ref{}) != game2BookTestLayout(0) || unsafe.Offsetof(ref{}.Data) != game2BookTestLayout(1) || unsafe.Sizeof(anim{}) != game2BookTestLayout(2) || unsafe.Offsetof(anim{}.End) != game2BookTestLayout(3) || unsafe.Offsetof(anim{}.Images) != game2BookTestLayout(4) || unsafe.Offsetof(anim{}.Frame) != game2BookTestLayout(5) {
		t.Fatal("C/Go native animation layout")
	}
	a := new(anim)
	r := new(ref)
	r.Data = unsafe.Pointer(a)
	var pin runtime.Pinner
	pin.Pin(a)
	pin.Pin(r)
	defer pin.Unpin()
	for _, creature := range []bool{false, true} {
		fn := game2BookTestCallback(creature)
		if uintptr(fn) <= math.MaxUint32 && unsafe.Sizeof(uintptr(0)) == 8 {
			t.Fatal("callback must retain full width")
		}
		for i := 0; i < 3; i++ {
			client.sequence = uint([]uint32{0, 0x80000001, 0xffffffff}[i])
			runtime.GC()
			if game2BookTestAnim(false, unsafe.Pointer(r), nil, 0) != unsafe.Pointer(a) {
				t.Fatal("native ref pointer")
			}
			game2BookTestAnim(false, unsafe.Pointer(r), fn, 1)
			if a.End != fn {
				t.Fatal("native callback truncation")
			}
			a.Frame = 0xdecafbad
			game2BookTestAnim(true, unsafe.Pointer(r), nil, 2)
			want := a.Frame
			a.Frame = 0xdecafbad
			game2BookTestAnim(false, unsafe.Pointer(r), nil, 2)
			if a.Frame != want {
				t.Fatal("sequence write")
			}
		}
	}
	if game2BookTestAnim(false, nil, nil, 0) != nil {
		t.Fatal("nil ref")
	}
	game2BookTestAnim(false, nil, nil, 1)
	game2BookTestAnim(false, nil, nil, 2)
	r.Data = nil
	if game2BookTestAnim(false, unsafe.Pointer(r), nil, 0) != nil {
		t.Fatal("nil anim")
	}
	runtime.KeepAlive(r)
	runtime.KeepAlive(a)
}

func TestGame2BookClientPacketsNativeReference(t *testing.T) {
	cursor, send := game2BookCursor, game2BookSend
	defer func() { game2BookCursor = cursor; game2BookSend = send }()
	input := []uint32{0x89abcdef, 0x10203040, 0xffffffff, 0x80000000, 0x76543210}
	var pin runtime.Pinner
	pin.Pin(&input[0])
	defer pin.Unpin()
	for _, ability := range []bool{false, true} {
		for _, busy := range []int32{0, 7} {
			for _, count := range []int32{-1, 0, 1, 2, 5, 6, 0x12345678} {
				for _, flag := range []byte{0, 1, 0x80, 0xff} {
					t.Run(fmt.Sprintf("ability%t/cursor%d/count%d/flag%x", ability, busy, count, flag), func(t *testing.T) {
						game2BookTestConfig(busy, -12345)
						want := game2BookTestSend(true, ability, unsafe.Pointer(&input[0]), count, flag)
						packet, calls := game2BookTestCapture()
						gotCalls := 0
						var gotPacket []byte
						game2BookCursor = func() int32 { return busy }
						game2BookSend = func(p []byte) int32 { gotCalls++; gotPacket = append([]byte(nil), p...); runtime.GC(); return -12345 }
						got := game2BookTestSend(false, ability, unsafe.Pointer(&input[0]), count, flag)
						if got != want || calls != gotCalls || !bytes.Equal(gotPacket, packet) {
							t.Fatalf("C entry/reference got=%d/%x/%d want=%d/%x/%d", got, gotPacket, gotCalls, want, packet, calls)
						}
						if !ability && busy == 0 {
							if len(gotPacket) != 22 || gotPacket[0] != 121 || gotPacket[21] != flag {
								t.Fatal("wire22")
							}
							for i := 0; i < 5; i++ {
								v := uint32(0)
								if int32(i) < count {
									v = input[i]
								}
								if binary.LittleEndian.Uint32(gotPacket[1+4*i:]) != v {
									t.Fatal("DWORD packet")
								}
							}
						}
					})
				}
			}
		}
	}
	input[0] = 0
	game2BookCursor = func() int32 { return 0 }
	game2BookSend = func([]byte) int32 { t.Fatal("zero-first must not send"); return 0 }
	if game2BookTestSend(false, false, unsafe.Pointer(&input[0]), 5, 0) != 0 {
		t.Fatal("zero-first result")
	}
	runtime.KeepAlive(input)
}

func TestGame2BookFlagsMotionAndStateNativeReference(t *testing.T) {
	// Foreign native allocation is independent of packed memmap DWORD storage.
	owner, free := alloc.Make([]byte{}, int(game2BookTestLayout(6)))
	defer free()
	original := *memmap.PtrPtr(0x852978, 8)
	defer func() { *memmap.PtrPtr(0x852978, 8) = original }()
	for _, flags := range []uint32{0, 1, 2, 3, 0x80000000, 0xffffffff} {
		*(*uint32)(unsafe.Add(unsafe.Pointer(&owner[0]), game2BookTestLayout(7))) = flags
		*memmap.PtrPtr(0x852978, 8) = unsafe.Pointer(&owner[0])
		for _, mask := range []uint32{0, 1, 2, 3, 0x80000000, 0xffffffff} {
			if game2BookTestLocalFlags(false, mask) != game2BookTestLocalFlags(true, mask) {
				t.Fatal("native drawable flags")
			}
		}
	}
	*memmap.PtrPtr(0x852978, 8) = nil
	if game2BookTestLocalFlags(false, 0) != 0 {
		t.Fatal("nil local flags")
	}
	region := memmap.Slice(0x5D4594, 1046676)[:192]
	saved := append([]byte(nil), region...)
	defer copy(region, saved)
	a := [2]int32{16777217, -16777217}
	b := [2]int32{math.MinInt32, math.MaxInt32}
	for _, count := range []uint32{0, 1, 18, 19, 20, 21, 0xffffffff} {
		for i := range region {
			region[i] = 0xa5
		}
		*game2BookU(1046680) = count
		seed := append([]byte(nil), region...)
		want := game2BookTestMove(true, unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]))
		after := append([]byte(nil), region...)
		copy(region, seed)
		got := game2BookTestMove(false, unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]))
		if got != want || !bytes.Equal(region, after) {
			t.Fatal("ordered overlapping motion DWORD/float writes")
		}
	}
	old0 := game2BookTestMode(0, 0, false)
	old1 := game2BookTestMode(1, 0, false)
	defer game2BookTestMode(0, old0, true)
	defer game2BookTestMode(1, old1, true)
	for _, guide := range []bool{false, true} {
		game2BookTestMode(0, 0xface, true)
		game2BookTestMode(1, 0xbeef, true)
		want := game2BookTestClick(true, guide)
		w0, w1 := game2BookTestMode(0, 0, false), game2BookTestMode(1, 0, false)
		game2BookTestMode(0, 0xface, true)
		game2BookTestMode(1, 0xbeef, true)
		if game2BookTestClick(false, guide) != want || game2BookTestMode(0, 0, false) != w0 || game2BookTestMode(1, 0, false) != w1 {
			t.Fatal("click state")
		}
	}
}

func TestGame2BookBusyCallbacksUseNativeABI(t *testing.T) {
	busy := game2BookTestBusy(0, false)
	defer game2BookTestBusy(busy, true)
	game2BookTestBusy(1, true)
	// A pinned foreign-looking high pointer is passed through the actual public
	// C callback ABI; busy input must return before dereferencing window/player.
	window := make([]byte, 256)
	var pin runtime.Pinner
	pin.Pin(&window[0])
	defer pin.Unpin()
	p := unsafe.Pointer(&window[0])
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
		t.Fatal("native callback arg")
	}
	for which := 0; which < 4; which++ {
		for _, event := range []uint32{0, 5, 6, 7, 8, 10, 11, 0xffffffff} {
			runtime.GC()
			if got, want := game2BookTestBlockedCallback(false, which, p, event, 0xffff8001), game2BookTestBlockedCallback(true, which, p, event, 0xffff8001); got != want || got != 1 {
				t.Fatal("busy gate/native window ABI")
			}
		}
	}
	runtime.KeepAlive(window)
}

func TestGame2BookResetAndTargetNativeReference(t *testing.T) {
	oldTimeout := InputSetKeyTimeoutLegacy
	defer func() { InputSetKeyTimeoutLegacy = oldTimeout }()
	var keyCalls []byte
	InputSetKeyTimeoutLegacy = func(key byte) { keyCalls = append(keyCalls, key) }
	span := memmap.Slice(0x5D4594, 1047916)[:1800]
	saved := append([]byte(nil), span...)
	defer copy(span, saved)
	for i := range span {
		span[i] = byte(i*37 + 11)
	}
	seed := append([]byte(nil), span...)
	want := game2BookTestReset(true)
	after := append([]byte(nil), span...)
	copy(span, seed)
	if game2BookTestReset(false) != want || !bytes.Equal(span, after) {
		t.Fatal("exact136byte reset plus separate byte and guards")
	}
	for _, direct := range []bool{false, true} {
		for _, busy := range []uint32{0, 1, 0xffffffff} {
			for _, id := range []int32{0, 1, -1, math.MinInt32, math.MaxInt32} {
				copy(span, seed)
				*game2BookU(1047916) = busy
				before := append([]byte(nil), span...)
				keyCalls = nil
				want := game2BookTestTarget(true, direct, id)
				wantKeys := append([]byte(nil), keyCalls...)
				after := append([]byte(nil), span...)
				copy(span, before)
				keyCalls = nil
				if game2BookTestTarget(false, direct, id) != want || !bytes.Equal(span, after) || !bytes.Equal(keyCalls, wantKeys) {
					t.Fatal("target state order/domain/key timeout")
				}
			}
		}
	}
}
