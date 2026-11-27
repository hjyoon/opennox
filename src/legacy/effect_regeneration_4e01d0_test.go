package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestRegenerationUpdate4E01D0NativeCallback(t *testing.T) {
	owner, freeOwner := alloc.NewHigh(server.Object{})
	defer freeOwner()
	health, freeHealth := alloc.NewHigh(server.HealthData{Cur: 149, Max: 150})
	defer freeHealth()
	// alloc.New reserves zeroed C memory; its argument determines the type,
	// not the allocation's initial contents.
	*health = server.HealthData{Cur: 149, Max: 150}
	item, freeItem := alloc.NewHigh(server.Object{})
	defer freeItem()
	effect, freeEffect := alloc.NewHigh(server.ModifierEff{})
	defer freeEffect()
	owner.HealthData = health
	item.InvHolder = owner
	effect.Update100 = server.ModifierEffFnc{Fnc: regenerationEffectPointerNative4E01D0(), Val: 60}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{owner.CObj(), item.CObj(), effect.C(), unsafe.Pointer(health)} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatalf("expected native high pointer, got %p", ptr)
			}
		}
	}
	previous := regenerationEffectCall4E01D0
	defer func() { regenerationEffectCall4E01D0 = previous }()
	var srv server.Server
	srv.SetFrame(36)
	srv.SetTickRate(30)
	calls := 0
	regenerationEffectCall4E01D0 = func(got *server.ModifierEff, it *server.Object) {
		if got != effect || it != item {
			t.Fatal("C callback truncated/changed native arguments")
		}
		srv.EffectRegeneration4E01D0(got, it, func(unit *server.Object, delta int32) {
			if unit != owner || delta != 1 {
				t.Fatal("wrong owner or delta")
			}
			calls++
			health.Cur++
		})
	}
	ccall.CallVoidPtr3(effect.Update100.Fnc, effect.C(), item.CObj(), nil)
	if calls != 1 || health.Cur != 150 {
		t.Fatalf("heals/HP=%d/%d, want 1/150", calls, health.Cur)
	}
}

func TestItemsApplyUpdateEffect4FA490NativeRegeneration(t *testing.T) {
	health := &server.HealthData{Cur: 149, Max: 150}
	owner := &server.Object{HealthData: health}
	effect := &server.ModifierEff{Update100: server.ModifierEffFnc{Fnc: regenerationEffectPointerNative4E01D0(), Val: 60}}
	attrs := &server.ModifierInitData{Modifiers: [4]*server.ModifierEff{2: effect}}
	item := &server.Object{
		InvHolder: owner,
		ObjFlags:  0x100,
		ObjClass:  0x01000000,
		InitData:  unsafe.Pointer(attrs),
	}
	owner.InvFirstItem = item
	previous := regenerationEffectCall4E01D0
	defer func() { regenerationEffectCall4E01D0 = previous }()
	var srv server.Server
	srv.SetFrame(36)
	srv.SetTickRate(30)
	calls := 0
	regenerationEffectCall4E01D0 = func(got *server.ModifierEff, it *server.Object) {
		if got != effect || it != item {
			t.Fatal("dispatcher must pass the native equipped item")
		}
		srv.EffectRegeneration4E01D0(got, it, func(unit *server.Object, delta int32) {
			if unit != owner || delta != 1 {
				t.Fatal("healing owner/delta")
			}
			calls++
			health.Cur++
		})
	}
	srv.ItemsApplyUpdateEffect(owner)
	if calls != 1 || health.Cur != 150 {
		t.Fatalf("heals/HP=%d/%d, want 1/150", calls, health.Cur)
	}
}
