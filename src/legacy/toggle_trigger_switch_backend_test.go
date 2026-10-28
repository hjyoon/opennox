package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestTriggerNativeSharedCacheAndCaller53B1B0(t *testing.T) {
	cells := [2]*uint32{memmap.PtrUint32(0x5D4594, 2488680), memmap.PtrUint32(0x5D4594, 2488676)}
	old := [2]uint32{*cells[0], *cells[1]}
	defer func() { *cells[0] = old[0]; *cells[1] = old[1] }()
	*cells[0] = 0
	*cells[1] = 0
	frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
	source := (*server.Object)(unsafe.Pointer(&frame[32]))
	df := radialCFrameOriginal(t, 60)
	data := (*server.TriggerUpdateData)(unsafe.Pointer(&df[32]))
	*data = server.TriggerUpdateData{Flags: 9, Field4: 0xabbaabba}
	cf := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
	caller := (*server.Object)(unsafe.Pointer(&cf[32]))
	*caller = server.Object{}
	*source = server.Object{TypeInd: 17, ObjClass: object.ClassImmobile, ObjFlags: object.FlagEnabled, UpdateData: unsafe.Pointer(data)}
	h := toggleTriggerRawRuntime53B060()
	h.Frame = func() uint32 { return 10 }
	h.FPS = func() uint32 { t.Fatal("immediate trigger read FPS"); return 0 }
	h.AudioEvent = func(uint32, *server.Object) {}
	h.CollideTarget = func(o *server.Object, d *server.TriggerUpdateData) *server.Object {
		if o != source || d != data {
			t.Fatal("native source/record identity")
		}
		return caller
	}
	scripts := 0
	h.ScriptCallback = func(b *server.ScriptCallback, c, tr *server.Object, event server.ScriptEventType) {
		if b != &data.ScriptActivate || c != caller || tr != source || event != server.NoxEventTriggerActivated {
			t.Fatal("wide caller/script identity or event metadata")
		}
		scripts++
	}
	lookups := 0
	lookup := func(name string) uint32 {
		lookups++
		switch name {
		case "Trigger":
			return 17
		case "PressurePlate":
			if toggleTriggerCacheProbe(0) != 17 {
				t.Fatal("Trigger cache was not stored before PressurePlate lookup")
			}
			return 18
		}
		t.Fatal(name)
		return 0
	}
	oldH, oldI := toggleTriggerRawRuntime53B060, triggerImmediateType53B1B0
	defer func() { toggleTriggerRawRuntime53B060 = oldH; triggerImmediateType53B1B0 = oldI }()
	toggleTriggerRawRuntime53B060 = func() server.ToggleTriggerRawRuntime53B060 { return h }
	triggerImmediateType53B1B0 = func(o *server.Object) bool { return TriggerImmediateTypeNative53B1B0(o, lookup) }
	if got := toggleTriggerSwitchProbe(1, source); got != 12 || data.State != 1 || source.Field34 != 10 || lookups != 2 || scripts != 1 {
		t.Fatal("cold native cache/activation", got, lookups, scripts)
	}
	data.Flags = 9
	data.State = 0
	if got := toggleTriggerSwitchProbe(1, source); got != 12 || lookups != 2 || scripts != 2 {
		t.Fatal("warm cache repeated lookup")
	}
	*cells[0] = 0x10011
	*cells[1] = 0x10012
	if TriggerImmediateTypeNative53B1B0(source, lookup) {
		t.Fatal("full DWORD type caches were truncated to uint16")
	}
	if data.Field4 != 0xabbaabba || source.Field33 != 1 || source.Field38 != 0xffffffff {
		t.Fatal("fixed wire slot or real animation synchronization")
	}
}

func TestSwitchNativeQueueReturn53B320(t *testing.T) {
	oldQ, oldR := pendingCollisionQueue537580, pendingCollisionRuntime537610
	defer func() {
		pendingCollisionQueue537580 = oldQ
		pendingCollisionRuntime537610 = oldR
		pendingCollisionSync537580()
	}()
	pendingCollisionQueue537580 = server.PendingCollisionQueue537580{}
	pendingCollisionRuntime537610.TypeByName = func(name string) uint32 {
		if name == "PeriodicSpike" {
			return 2
		}
		return 1
	}
	frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
	source := (*server.Object)(unsafe.Pointer(&frame[32]))
	cf := radialCFrameOriginal(t, 1)
	*source = server.Object{ObjClass: object.ClassImmobile, ObjFlags: object.FlagActive | object.FlagEnabled | object.FlagNoCollide, Collide: unsafe.Pointer(&cf[32])}
	got := toggleTriggerSwitchProbe(2, source)
	if got != 2 || !pendingCollisionQueue537580.Empty() || source.ObjFlags.Has(object.FlagNoCollide) {
		t.Fatal("real immobile-class filter and helper byte", got)
	}
	// A generic active collidable object reaches the actual queue insertion.
	source.ObjClass = 0
	source.ObjFlags |= object.FlagNoCollide
	got = toggleTriggerSwitchProbe(2, source)
	if pendingCollisionQueue537580.Head() != source || pendingCollisionQueue537580.Contains(source) == 0 || source.ObjFlags.Has(object.FlagNoCollide) {
		t.Fatal("actual native collision queue")
	}
	// The helper's native return byte is observable independently of flags.
	if got != 1 || got == uint8(source.ObjFlags) {
		t.Fatal("queue helper byte replaced by flags", got, uint8(source.ObjFlags))
	}
}
