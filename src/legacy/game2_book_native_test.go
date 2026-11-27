package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/client/gui"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

type game2BookTestClient struct {
	Client
	sequence uint
}

func (c *game2BookTestClient) GetInputSeq() uint { return c.sequence }

func TestGame2BookForwardAndPositionReturnABI(t *testing.T) {
	g := gui.NewHigh(nil)
	defer g.DestroyAll()
	parent := g.NewWindowRaw(nil, gui.StatusEnabled, -133, 81, 100, 100, nil)
	icon := g.NewWindowRaw(parent, gui.StatusEnabled, 47, -15, 8, 8, nil)
	oldList, oldDetails, oldState := bookUI.list_window, *bookDetails, *bookState
	oldSound := ClientPlaySoundSpecial
	cursor := memmap.PtrUint32(0x5D4594, 1096672)
	oldCursor := *cursor
	ids := unsafe.Slice(game2BookU(1046960), 137)
	oldIDs := append([]uint32(nil), ids...)
	defer func() {
		bookUI.list_window, *bookDetails, *bookState = oldList, oldDetails, oldState
		ClientPlaySoundSpecial = oldSound
		*cursor = oldCursor
		copy(ids, oldIDs)
	}()
	ClientPlaySoundSpecial = func(sound.ID, int) {}
	*cursor = 1
	bookUI.list_window = (*nox_window)(parent.C())
	bookDetails.forward, bookDetails.right_arrow, bookDetails.left_arrow = nil, nil, nil
	bookDetails.page_icon = nil
	clear(ids)

	for _, tc := range []struct {
		kind uintptr
		want uintptr
	}{
		{2, uintptr(unsafe.Pointer(game2BookU(1047508)))},
		{3, uintptr(unsafe.Pointer(game2BookU(1046984)))},
		{4, uintptr(unsafe.Pointer(game2BookU(1047508)))},
		{^uintptr(0), ^uintptr(0)},
	} {
		parent.SetPos(image.Pt(6, 4))
		got := cfixture.Game2BookForward(tc.kind, 87, -133, 81)
		if got != tc.want || parent.Offs() != image.Pt(-133, 81) {
			t.Fatalf("kind %#x returned %#x at %v, want %#x at (-133,81)", tc.kind, got, parent.Offs(), tc.want)
		}
	}
	if high := ^uintptr(0) &^ uintptr(math.MaxUint32); high != 0 {
		kind := high | 2
		if got := cfixture.Game2BookForward(kind, 87, -133, 81); got != kind {
			t.Fatalf("kind %#x was narrowed to %#x", kind, got)
		}
	}
	ids[5] = 87
	if got := cfixture.Game2BookForward(3, 87, -133, 81); got != ^uintptr(0)-1 || bookState.entry_index != 5 {
		t.Fatalf("found page returned %#x with entry %d, want sign-extended -2 and entry 5", got, bookState.entry_index)
	}

	if got := cfixture.Game2BookPosition(false, nil); got != 0 {
		t.Fatalf("nil output returned %d, want 0", got)
	}
	p := [2]uint32{0x89abcdef, 0x11223344}
	if got := cfixture.Game2BookPosition(false, &p); got != -2 || p != [2]uint32{0x89abcdef, 0x11223344} {
		t.Fatalf("missing icon returned %d and wrote %x", got, p)
	}
	bookDetails.page_icon = (*nox_window)(icon.C())
	var want [2]uint32
	wantStatus := cfixture.Game2BookPosition(true, &want)
	if got := cfixture.Game2BookPosition(false, &p); got != wantStatus || got != 0 || p != want ||
		int32(p[0]) != -86 || int32(p[1]) != 66 {
		t.Fatalf("position returned %d, wrote %v, reference is %d and %v", got, p, wantStatus, want)
	}
}

type game2BookRewardTestClient struct {
	Client
	frames []uint
	reads  int
}

func (c *game2BookRewardTestClient) GetInputSeq() uint {
	frame := c.frames[c.reads]
	c.reads++
	return frame
}

type game2BookRewardTestServer struct {
	Server
	native server.Server
}

func (s *game2BookRewardTestServer) S() *server.Server { return &s.native }

func TestGame2BookRewardReadsLiveSequenceAfterCoopGate(t *testing.T) {
	oldClient, oldServer := GetClient, GetServer
	oldGame := noxflags.GetGame()
	last := g2ServerWord(1217504)
	cursor := memmap.PtrUint32(0x5D4594, 1096672)
	oldLast, oldCursor := *last, *cursor
	oldList, oldPool := bookUI.list_window, legacyPointers.screenParticleAllocator
	defer func() {
		GetClient, GetServer = oldClient, oldServer
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
		*last, *cursor = oldLast, oldCursor
		bookUI.list_window, legacyPointers.screenParticleAllocator = oldList, oldPool
	}()
	*cursor = 1
	bookUI.list_window = nil
	legacyPointers.screenParticleAllocator = nil
	srv := &game2BookRewardTestServer{}
	srv.native.Rand.Other = prand.New(0)
	GetServer = func() Server { return srv }
	for _, tc := range []struct {
		name  string
		coop  bool
		last  uint32
		reads []uint
		want  uint32
	}{
		{"cooperative reject", true, 35, []uint{36}, 35},
		{"cooperative accept", true, 35, []uint{37, 49}, 49},
		{"cooperative wrap", true, math.MaxUint32, []uint{1, 4}, 4},
		{"ordinary reward", false, 35, []uint{36}, 36},
	} {
		t.Run(tc.name, func(t *testing.T) {
			noxflags.ResetGame()
			if tc.coop {
				noxflags.SetGame(noxflags.GameModeCoop)
			}
			client := &game2BookRewardTestClient{frames: tc.reads}
			GetClient = func() Client { return client }
			*last = tc.last
			cfixture.Game2BookReward(^uintptr(0), 87, 0)
			if client.reads != len(tc.reads) || *last != tc.want {
				t.Fatalf("read %d frames and saved %d, want %d and %d", client.reads, *last, len(tc.reads), tc.want)
			}
		})
	}
}

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
				got := cfixture.Game2BookRow(false, base, slot, 1)
				want := cfixture.Game2BookRow(true, base, slot, 1)
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
		if cfixture.Game2BookRow(false, base, 0, 0) != cfixture.Game2BookRow(true, base, 0, 0) {
			t.Fatal("row ABI")
		}
		*memmap.PtrUint8(0x5D4594, 1048140) = row
		if cfixture.Game2BookRow(false, nil, 0, 2) != cfixture.Game2BookRow(true, nil, 0, 2) {
			t.Fatal("trap row")
		}
	}
	if cfixture.Game2BookRow(false, nil, 0, 0) != nil || cfixture.Game2BookRow(false, nil, 0, 1) != nil {
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
	if unsafe.Sizeof(ref{}) != cfixture.Game2BookLayout(0) || unsafe.Offsetof(ref{}.Data) != cfixture.Game2BookLayout(1) || unsafe.Sizeof(anim{}) != cfixture.Game2BookLayout(2) || unsafe.Offsetof(anim{}.End) != cfixture.Game2BookLayout(3) || unsafe.Offsetof(anim{}.Images) != cfixture.Game2BookLayout(4) || unsafe.Offsetof(anim{}.Frame) != cfixture.Game2BookLayout(5) {
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
		fn := cfixture.Game2BookCallback(creature)
		if uintptr(fn) <= math.MaxUint32 && unsafe.Sizeof(uintptr(0)) == 8 {
			t.Fatal("callback must retain full width")
		}
		for i := 0; i < 3; i++ {
			client.sequence = uint([]uint32{0, 0x80000001, 0xffffffff}[i])
			runtime.GC()
			if cfixture.Game2BookAnim(false, unsafe.Pointer(r), nil, 0) != unsafe.Pointer(a) {
				t.Fatal("native ref pointer")
			}
			cfixture.Game2BookAnim(false, unsafe.Pointer(r), fn, 1)
			if a.End != fn {
				t.Fatal("native callback truncation")
			}
			a.Frame = 0xdecafbad
			cfixture.Game2BookAnim(true, unsafe.Pointer(r), nil, 2)
			want := a.Frame
			a.Frame = 0xdecafbad
			cfixture.Game2BookAnim(false, unsafe.Pointer(r), nil, 2)
			if a.Frame != want {
				t.Fatal("sequence write")
			}
		}
	}
	if cfixture.Game2BookAnim(false, nil, nil, 0) != nil {
		t.Fatal("nil ref")
	}
	cfixture.Game2BookAnim(false, nil, nil, 1)
	cfixture.Game2BookAnim(false, nil, nil, 2)
	r.Data = nil
	if cfixture.Game2BookAnim(false, unsafe.Pointer(r), nil, 0) != nil {
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
						cfixture.Game2BookConfig(busy, -12345)
						want := cfixture.Game2BookSend(true, ability, unsafe.Pointer(&input[0]), count, flag)
						packet, calls := cfixture.Game2BookCapture()
						gotCalls := 0
						var gotPacket []byte
						game2BookCursor = func() int32 { return busy }
						game2BookSend = func(p []byte) int32 { gotCalls++; gotPacket = append([]byte(nil), p...); runtime.GC(); return -12345 }
						got := cfixture.Game2BookSend(false, ability, unsafe.Pointer(&input[0]), count, flag)
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
	if cfixture.Game2BookSend(false, false, unsafe.Pointer(&input[0]), 5, 0) != 0 {
		t.Fatal("zero-first result")
	}
	runtime.KeepAlive(input)
}

func TestGame2BookFlagsMotionAndStateNativeReference(t *testing.T) {
	// Foreign native allocation is independent of packed memmap DWORD storage.
	owner, free := alloc.MakeHigh([]byte{}, int(cfixture.Game2BookLayout(6)))
	defer free()
	original := *memmap.PtrPtr(0x852978, 8)
	defer func() { *memmap.PtrPtr(0x852978, 8) = original }()
	for _, flags := range []uint32{0, 1, 2, 3, 0x80000000, 0xffffffff} {
		*(*uint32)(unsafe.Add(unsafe.Pointer(&owner[0]), cfixture.Game2BookLayout(7))) = flags
		*memmap.PtrPtr(0x852978, 8) = unsafe.Pointer(&owner[0])
		for _, mask := range []uint32{0, 1, 2, 3, 0x80000000, 0xffffffff} {
			if cfixture.Game2BookLocalFlags(false, mask) != cfixture.Game2BookLocalFlags(true, mask) {
				t.Fatal("native drawable flags")
			}
		}
	}
	*memmap.PtrPtr(0x852978, 8) = nil
	if cfixture.Game2BookLocalFlags(false, 0) != 0 {
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
		want := cfixture.Game2BookMove(true, unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]))
		after := append([]byte(nil), region...)
		copy(region, seed)
		got := cfixture.Game2BookMove(false, unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]))
		if got != want || !bytes.Equal(region, after) {
			t.Fatal("ordered overlapping motion DWORD/float writes")
		}
	}
	old0 := cfixture.Game2BookMode(0, 0, false)
	old1 := cfixture.Game2BookMode(1, 0, false)
	defer cfixture.Game2BookMode(0, old0, true)
	defer cfixture.Game2BookMode(1, old1, true)
	for _, guide := range []bool{false, true} {
		cfixture.Game2BookMode(0, 0xface, true)
		cfixture.Game2BookMode(1, 0xbeef, true)
		want := cfixture.Game2BookClick(true, guide)
		w0, w1 := cfixture.Game2BookMode(0, 0, false), cfixture.Game2BookMode(1, 0, false)
		cfixture.Game2BookMode(0, 0xface, true)
		cfixture.Game2BookMode(1, 0xbeef, true)
		if cfixture.Game2BookClick(false, guide) != want || cfixture.Game2BookMode(0, 0, false) != w0 || cfixture.Game2BookMode(1, 0, false) != w1 {
			t.Fatal("click state")
		}
	}
}

func TestGame2BookBusyCallbacksUseNativeABI(t *testing.T) {
	busy := cfixture.Game2BookBusy(0, false)
	defer cfixture.Game2BookBusy(busy, true)
	cfixture.Game2BookBusy(1, true)
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
			if got, want := cfixture.Game2BookBlockedCallback(false, which, p, event, 0xffff8001), cfixture.Game2BookBlockedCallback(true, which, p, event, 0xffff8001); got != want || got != 1 {
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
	want := cfixture.Game2BookReset(true)
	after := append([]byte(nil), span...)
	copy(span, seed)
	if cfixture.Game2BookReset(false) != want || !bytes.Equal(span, after) {
		t.Fatal("exact136byte reset plus separate byte and guards")
	}
	for _, direct := range []bool{false, true} {
		for _, busy := range []uint32{0, 1, 0xffffffff} {
			for _, id := range []int32{0, 1, -1, math.MinInt32, math.MaxInt32} {
				copy(span, seed)
				*game2BookU(1047916) = busy
				before := append([]byte(nil), span...)
				keyCalls = nil
				want := cfixture.Game2BookTarget(true, direct, id)
				wantKeys := append([]byte(nil), keyCalls...)
				after := append([]byte(nil), span...)
				copy(span, before)
				keyCalls = nil
				if cfixture.Game2BookTarget(false, direct, id) != want || !bytes.Equal(span, after) || !bytes.Equal(keyCalls, wantKeys) {
					t.Fatal("target state order/domain/key timeout")
				}
			}
		}
	}
}
