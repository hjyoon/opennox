package legacy

import (
	"bytes"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

func fireballUpdateEntryNative223(t *testing.T, power int, fn unsafe.Pointer) {
	t.Helper()
	vp, freeVP := alloc.New(noxrender.Viewport{})
	dr, freeDR := alloc.New(client.Drawable{})
	defer freeVP()
	defer freeDR()
	for _, pointer := range []unsafe.Pointer{vp.C(), dr.C()} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native fixture address below4GiB: %p", pointer)
		}
	}
	// Deliberately poison the obsolete PE32 gate. This is opaque input to
	// the C entry adapter, not a gameplay fixture. It keeps the old C body
	// from reaching uninitialized renderer services during the RED check.
	*(*uint32)(unsafe.Add(dr.C(), 120*4)) = 0x13579bdf
	beforeVP := bytes.Clone(unsafe.Slice((*byte)(vp.C()), unsafe.Sizeof(*vp)))
	beforeDR := bytes.Clone(unsafe.Slice((*byte)(dr.C()), unsafe.Sizeof(*dr)))
	previous := ClientUpdateFireball4CCE70
	defer func() { ClientUpdateFireball4CCE70 = previous }()
	want := math.MinInt32 + 223 + power
	calls := 0
	ClientUpdateFireball4CCE70 = func(v *noxrender.Viewport, d *client.Drawable, p int) int {
		if p != power || (calls == 0 && (v != vp || d != dr)) || (calls == 1 && (v != nil || d != nil)) || calls > 1 {
			t.Fatalf("callback%d changed pointers/power: %p/%p/%d", calls, v, d, p)
		}
		calls++
		return want
	}
	if got := ccall.CallIntPtr2(fn, vp.C(), dr.C()); got != want || calls != 1 {
		t.Fatalf("registered Fireball power%d return=%d want=%d calls=%d", power, got, want, calls)
	}
	if got := ccall.CallIntPtr2(fn, nil, nil); got != want || calls != 2 {
		t.Fatalf("nil forwarding return=%d calls=%d", got, calls)
	}
	if !bytes.Equal(beforeVP, unsafe.Slice((*byte)(vp.C()), unsafe.Sizeof(*vp))) || !bytes.Equal(beforeDR, unsafe.Slice((*byte)(dr.C()), unsafe.Sizeof(*dr))) {
		t.Fatal("native entry changed whole viewport/drawable records")
	}
	ClientUpdateFireball4CCE70 = nil
	if got := ccall.CallIntPtr2(fn, nil, nil); got != 1 {
		t.Fatalf("renderer-free entry return=%d, want1", got)
	}
	t.Logf("registered Fireball power%d: native viewport=%p drawable=%p, nil delegation, signed return, whole records unchanged, renderer-free keep-alive", power, vp, dr)
}

func TestFireballUpdate4CCE70RegisteredTitanNativePointers(t *testing.T) {
	fireballUpdateEntryNative223(t, 5, Get_sub_4CCE70())
}

func TestFireballUpdate4CD090RegisteredStrongNativePointers(t *testing.T) {
	fireballUpdateEntryNative223(t, 4, Get_sub_4CD090())
}

func TestFireballUpdate4CD0C0RegisteredRegularNativePointers(t *testing.T) {
	fireballUpdateEntryNative223(t, 3, Get_sub_4CD0C0())
}

func TestFireballUpdate4CD0F0RegisteredWeakNativePointers(t *testing.T) {
	fireballUpdateEntryNative223(t, 2, Get_sub_4CD0F0())
}

func TestFireballUpdate4CD120RegisteredPitifulNativePointers(t *testing.T) {
	fireballUpdateEntryNative223(t, 1, Get_sub_4CD120())
}
