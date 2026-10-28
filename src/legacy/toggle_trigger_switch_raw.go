package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TriggerImmediateTypeNative53B1B0(unit *server.Object, lookup func(string) uint32) bool {
	return server.TriggerImmediateTypeRaw53B1B0(unit, server.TriggerTypeCacheRuntime53B1B0{
		Load:   func(n int) uint32 { return *memmap.PtrUint32(0x5D4594, uintptr(2488680-4*n)) },
		Store:  func(n int, v uint32) { *memmap.PtrUint32(0x5D4594, uintptr(2488680-4*n)) = v },
		Lookup: lookup,
	})
}

var toggleTriggerRawRuntime53B060 = func() server.ToggleTriggerRawRuntime53B060 {
	return server.ToggleTriggerRawRuntime53B060{
		Frame: func() uint32 { return GetServer().S().Frame() }, FPS: func() uint32 { return GetServer().S().TickRate() },
		MarkAnimFrame: (*server.Object).MarkAnimFrame,
		CollideTarget: func(o *server.Object, _ *server.TriggerUpdateData) *server.Object { return o.TriggerCollideTarget() },
		AudioEvent:    func(id uint32, o *server.Object) { GetServer().S().Audio.EventObj(sound.ID(id), o, 0, 0) },
		ScriptCallback: func(b *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) {
			GetServer().NoxScriptC().ScriptCallback(b, caller, trigger, event)
		},
	}
}
var triggerImmediateType53B1B0 = func(o *server.Object) bool {
	return TriggerImmediateTypeNative53B1B0(o, func(name string) uint32 { return uint32(GetServer().S().Types.IndByID(name)) })
}
var switchRawRuntime53B320 = func() server.SwitchUpdateRawRuntime53B320 {
	return server.SwitchUpdateRawRuntime53B320{NeedSync: (*server.Object).NeedSync, QueueCollision: pendingCollisionEnqueue537610}
}

//export nox_toggle_trigger_switch_native
func nox_toggle_trigger_switch_native(kind C.int, unit *nox_object_t) C.uchar {
	switch kind {
	case 0:
		return C.uchar(server.ToggleUpdateRaw53B060(asObjectS(unit), toggleTriggerRawRuntime53B060()))
	case 1:
		return C.uchar(server.TriggerUpdateRaw53B1B0(asObjectS(unit), server.TriggerUpdateRawRuntime53B1B0{ToggleTriggerRawRuntime53B060: toggleTriggerRawRuntime53B060(), ImmediateType: triggerImmediateType53B1B0}))
	default:
		return C.uchar(server.SwitchUpdateRaw53B320(asObjectS(unit), switchRawRuntime53B320()))
	}
}
