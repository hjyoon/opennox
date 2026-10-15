package legacy

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

func deathBallDrawableEntryNative224(t *testing.T, kind int, fn unsafe.Pointer) {
	t.Helper()
	if os.Getenv("OPENNOX_DEATHBALL_UPDATE_224") != t.Name() {
		cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v")
		cmd.Env = append(os.Environ(), "OPENNOX_DEATHBALL_UPDATE_224="+t.Name())
		out, err := cmd.CombinedOutput()
		if err != nil {
			if len(out) > 8000 {
				out = append(out[:5000], out[len(out)-3000:]...)
			}
			t.Fatalf("actual registered DeathBall CGo entry: %v\n%s", err, out)
		}
		t.Log(string(out))
		return
	}
	vp, freeVP := alloc.New(noxrender.Viewport{})
	dr, freeDR := alloc.New(client.Drawable{})
	defer freeVP()
	defer freeDR()
	for _, pointer := range []unsafe.Pointer{vp.C(), dr.C()} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native fixture address below4GiB: %p", pointer)
		}
	}
	// Initialize only the old type-ID caches. A pre-port C entry must reach
	// its drawable read instead of failing in renderer/type-lookup setup.
	*memmap.PtrUint32(0x5D4594, 1523008) = 71
	*memmap.PtrUint32(0x5D4594, 1523012) = 72
	priorSparks, priorCharge, priorOrb := ClientUpdateDeathBallSparks4CDF80, ClientUpdateDeathBallCharge4CE0C0, ClientUpdateLinearOrb4CA650
	defer func() {
		ClientUpdateDeathBallSparks4CDF80, ClientUpdateDeathBallCharge4CE0C0, ClientUpdateLinearOrb4CA650 = priorSparks, priorCharge, priorOrb
	}()
	want := math.MinInt32 + 224 + kind
	calls := 0
	hook := func(v *noxrender.Viewport, d *client.Drawable) int {
		if (calls == 0 && (v != vp || d != dr)) || (calls == 1 && (v != nil || d != nil)) || calls > 1 {
			t.Fatalf("callback%d changed pointers: %p/%p", calls, v, d)
		}
		calls++
		return want
	}
	switch kind {
	case 0, 1:
		count := 3
		if kind == 1 {
			count = 1
		}
		ClientUpdateDeathBallSparks4CDF80 = func(v *noxrender.Viewport, d *client.Drawable, n int) int {
			if n != count {
				t.Fatalf("spark count=%d, want%d", n, count)
			}
			return hook(v, d)
		}
	case 2:
		ClientUpdateDeathBallCharge4CE0C0 = hook
	case 3:
		ClientUpdateLinearOrb4CA650 = hook
	default:
		t.Fatal("unknown callback kind")
	}
	beforeVP := bytes.Clone(unsafe.Slice((*byte)(vp.C()), unsafe.Sizeof(*vp)))
	beforeDR := bytes.Clone(unsafe.Slice((*byte)(dr.C()), unsafe.Sizeof(*dr)))
	t.Logf("registered DeathBall fixture: kind=%d viewport=%p drawable=%p", kind, vp, dr)
	if got := ccall.CallIntPtr2(fn, vp.C(), dr.C()); got != want || calls != 1 {
		t.Fatalf("return=%d want=%d calls=%d", got, want, calls)
	}
	if got := ccall.CallIntPtr2(fn, nil, nil); got != want || calls != 2 {
		t.Fatalf("nil forwarding return=%d calls=%d", got, calls)
	}
	if !bytes.Equal(beforeVP, unsafe.Slice((*byte)(vp.C()), unsafe.Sizeof(*vp))) || !bytes.Equal(beforeDR, unsafe.Slice((*byte)(dr.C()), unsafe.Sizeof(*dr))) {
		t.Fatal("native entry changed whole viewport/drawable records")
	}
	ClientUpdateDeathBallSparks4CDF80, ClientUpdateDeathBallCharge4CE0C0, ClientUpdateLinearOrb4CA650 = nil, nil, nil
	if got := ccall.CallIntPtr2(fn, nil, nil); got != 1 {
		t.Fatalf("renderer-free return=%d, want1", got)
	}
	t.Logf("registered DeathBall kind%d: native pointers, original counts, nil forwarding, signed return, input records unchanged, renderer-free keep-alive", kind)
}

func TestDeathBallUpdate4CDF80RegisteredBallNativePointers(t *testing.T) {
	deathBallDrawableEntryNative224(t, 0, Get_nox_xxx_updDrawDBall_4CDF80())
}

func TestDeathBallUpdate4CE0A0RegisteredFragmentNativePointers(t *testing.T) {
	deathBallDrawableEntryNative224(t, 1, Get_sub_4CE0A0())
}

func TestDeathBallUpdate4CE0C0RegisteredChargeNativePointers(t *testing.T) {
	deathBallDrawableEntryNative224(t, 2, Get_nox_xxx_updDrawDBallCharge_4CE0C0())
}

func TestDeathBallUpdate4CA650RegisteredLinearOrbNativePointers(t *testing.T) {
	deathBallDrawableEntryNative224(t, 3, Get_sub_4CA650())
}
