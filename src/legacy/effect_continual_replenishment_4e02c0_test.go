package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type continualReplenishmentServer218 struct {
	Server
	srv *server.Server
}

func (s *continualReplenishmentServer218) S() *server.Server { return s.srv }

func runContinualReplenishmentNative218(t *testing.T, dispatch bool) {
	t.Helper()
	child := "callback"
	if dispatch {
		child = "dispatch"
	}
	if os.Getenv("OPENNOX_CONTINUAL_REPLENISHMENT_218") != child {
		cmd := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v")
		cmd.Env = append(os.Environ(), "OPENNOX_CONTINUAL_REPLENISHMENT_218="+child)
		out, err := cmd.CombinedOutput()
		if err != nil {
			if len(out) > 5000 {
				out = append(out[:2500], out[len(out)-2500:]...)
			}
			t.Fatalf("actual CGo equipped-item callback failed: %v\n%s", err, out)
		}
		t.Log(string(out))
		return
	}
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	item, freeItem := alloc.New(server.Object{})
	defer freeItem()
	use, freeUse := alloc.New(server.WandUseData{})
	defer freeUse()
	effect, freeEffect := alloc.New(server.ModifierEff{})
	defer freeEffect()
	attrs, freeAttrs := alloc.New(server.ModifierInitData{})
	defer freeAttrs()
	for _, p := range []unsafe.Pointer{owner.CObj(), item.CObj(), unsafe.Pointer(use), effect.C(), unsafe.Pointer(attrs)} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= 0xffffffff {
			t.Fatalf("native pointer below4GiB: %p", p)
		}
	}
	var srv server.Server
	srv.SetFrame(240)
	prior := GetServer
	t.Cleanup(func() { GetServer = prior })
	GetServer = func() Server { return &continualReplenishmentServer218{srv: &srv} }
	owner.InvFirstItem = item
	item.ObjClass = object.ClassWand
	item.ObjFlags = object.FlagEquipped
	item.UseData.Ptr = unsafe.Pointer(use)
	item.InitData = unsafe.Pointer(attrs)
	// No holder: the original still updates charges, but must not report a player.
	use.Charge = 4
	use.MaxCharge = 5
	use.Progress = 0x5a5a5a5a
	attrs.Modifiers[2] = effect
	effect.Update100 = server.ModifierEffFnc{Fnc: continualReplenishmentEffectPointer4E02C0(), Val: 240}
	snapshot := func(p unsafe.Pointer, n uintptr) []byte { return bytes.Clone(unsafe.Slice((*byte)(p), n)) }
	beforeOwner := snapshot(owner.CObj(), unsafe.Sizeof(*owner))
	beforeItem := snapshot(item.CObj(), unsafe.Sizeof(*item))
	beforeEffect := snapshot(effect.C(), unsafe.Sizeof(*effect))
	beforeAttrs := snapshot(unsafe.Pointer(attrs), unsafe.Sizeof(*attrs))
	if dispatch {
		srv.ItemsApplyUpdateEffect(owner)
	} else {
		effect.CallUpdateNil(item)
	}
	if use.Charge != 5 || use.MaxCharge != 5 || use.Progress != 0x5a5a5a5a {
		t.Fatalf("charge/max/progress=%d/%d/%x", use.Charge, use.MaxCharge, use.Progress)
	}
	if !bytes.Equal(beforeOwner, snapshot(owner.CObj(), unsafe.Sizeof(*owner))) || !bytes.Equal(beforeItem, snapshot(item.CObj(), unsafe.Sizeof(*item))) || !bytes.Equal(beforeEffect, snapshot(effect.C(), unsafe.Sizeof(*effect))) || !bytes.Equal(beforeAttrs, snapshot(unsafe.Pointer(attrs), unsafe.Sizeof(*attrs))) {
		t.Fatal("callback changed native input records")
	}
	t.Logf("actual native CGo continual replenishment: dispatch=%t charge4->5 max5 progress unchanged", dispatch)
}

func TestContinualReplenishment4E02C0NativeCallback(t *testing.T) {
	runContinualReplenishmentNative218(t, false)
}

func TestContinualReplenishment4E02C0EquippedDispatcher(t *testing.T) {
	runContinualReplenishmentNative218(t, true)
}

func TestContinualReplenishment4E02C0NilItemDoesNotReadEffectOrFrame(t *testing.T) {
	old := GetServer
	t.Cleanup(func() { GetServer = old })
	GetServer = func() Server { t.Fatal("nil item queried server"); return nil }
	effect := server.ModifierEff{Update100: server.ModifierEffFnc{Fnc: continualReplenishmentEffectPointer4E02C0()}}
	effect.CallUpdateNil(nil)
}
