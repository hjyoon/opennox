package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestTrapDoorNativeStatusServices53DE80(t *testing.T) {
	frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
	source := (*server.Object)(unsafe.Pointer(&frame[32]))
	dataFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.TrapDoorCollideData{}))
	data := (*server.TrapDoorCollideData)(unsafe.Pointer(&dataFrame[32]))
	*data = server.TrapDoorCollideData{NextFrame: 10, Activated: 7}
	*source = server.Object{ObjClass: object.ClassImmobile, Field5: 2, CollideData: unsafe.Pointer(data)}
	h := trapDoorUpdateRawRuntime53DE80()
	h.Frame = func() uint32 { return 10 }
	h.FPS = func() uint32 { return 30 }
	audio := 0
	h.AudioEvent = func(id uint32, o *server.Object) {
		if id != 874 || o != source || o.Field5 != 4 || data.NextFrame != 160 || !o.ObjFlags.Has(object.FlagEnabled) {
			t.Fatal("audio before real transition/deadline")
		}
		audio++
	}
	old := trapDoorUpdateRawRuntime53DE80
	defer func() { trapDoorUpdateRawRuntime53DE80 = old }()
	trapDoorUpdateRawRuntime53DE80 = func() server.TrapDoorUpdateRawRuntime53DE80 { return h }
	trapDoorUndeadNativeProbe(0, source)
	if source.Field38 != math.MaxUint32 || audio != 1 {
		t.Fatal("real status synchronization/audio")
	}
	for _, v := range source.Field140 {
		if v&0xc0000 != 0xc0000 {
			t.Fatal("on/off and status dirty words")
		}
	}
	// The reached opening branch needs status services but no record or clock.
	source.CollideData = nil
	source.Field5 = 6
	h.Frame = func() uint32 { t.Fatal("opening read clock"); return 0 }
	trapDoorUndeadNativeProbe(0, source)
	if source.Field5 != 12 {
		t.Fatal("opening bit priority")
	}
}

func TestUndeadKillerNativeCancellationBeforeClock53E190(t *testing.T) {
	frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
	source := (*server.Object)(unsafe.Pointer(&frame[32]))
	*source = server.Object{}
	dataFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.UndeadKillerCollideData{}))
	data := (*server.UndeadKillerCollideData)(unsafe.Pointer(&dataFrame[32]))
	spellFrame := radialCFrameOriginal(t, unsafe.Sizeof(server.DurSpell{}))
	spell := (*server.DurSpell)(unsafe.Pointer(&spellFrame[32]))
	*spell = server.DurSpell{Flags88: 0x101}
	*data = server.UndeadKillerCollideData{Spell: spell}
	source.CollideData = unsafe.Pointer(data)
	h := undeadKillerUpdateRawRuntime53E190()
	deletes := 0
	h.Frame = func() uint32 { t.Fatal("cancelled spell read clock"); return 0 }
	h.DelayedDelete = func(o *server.Object) {
		if o != source {
			t.Fatal("delete identity")
		}
		deletes++
	}
	old := undeadKillerUpdateRawRuntime53E190
	defer func() { undeadKillerUpdateRawRuntime53E190 = old }()
	undeadKillerUpdateRawRuntime53E190 = func() server.UndeadKillerUpdateRawRuntime53E190 { return h }
	trapDoorUndeadNativeProbe(1, source)
	undeadKillerUpdateCall53E190(source)
	if deletes != 2 {
		t.Fatal("C/registered-Go cancellation", deletes)
	}
}
