package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestAudioSourceNativeDeviceContextLifecycleCABI(t *testing.T) {
	a, _ := asnOwnedFixture(t)
	oldTarget := Get_dword_5d4594_805984()
	t.Cleanup(func() { Set_dword_5d4594_805984(oldTarget) })
	class := (*AudioClassNative)(a.take(unsafe.Sizeof(AudioClassNative{})))
	methods := (*AudioMethodsNative)(a.take(unsafe.Sizeof(AudioMethodsNative{})))
	class.Name, class.Flags, class.State, class.Methods = a.take(16), 2, 0x80, methods
	class.Create, class.Destroy = asnCallback(10, false), asnCallback(11, true)
	class.ContextCreate, class.ContextDestroy = asnCallback(12, false), asnCallback(13, true)
	methods.Create, methods.Destroy = asnCallback(14, false), asnCallback(15, true)
	methods.Play, methods.Stop, methods.Refresh = asnCallback(9, false), asnCallback(8, true), asnCallback(7, true)
	counts := make(map[int]int)
	deviceFailure, contextFailure, voiceFailure := false, false, false
	audioNativeFixtureCallbackService = func(kind int, p unsafe.Pointer) int32 {
		counts[kind]++
		switch kind {
		case 10:
			d := (*AudioDeviceNative)(p)
			if d.Class != class {
				t.Fatal("class native pointer")
			}
			d.Limit = 2
			if deviceFailure {
				return -11
			}
		case 11:
			if (*AudioDeviceNative)(p).Class != class {
				t.Fatal("destroy class pointer")
			}
		case 12:
			ctx := (*AudioContextNative)(p)
			if ctx.Device.Class != class || ctx.Methods != methods || ctx.Index < 0 || ctx.Index > 1 || ctx.Device.Contexts[ctx.Index] != ctx {
				t.Fatal("context native owner/slot/methods")
			}
			ctx.Limit = 2
			if contextFailure {
				return -12
			}
		case 13:
			ctx := (*AudioContextNative)(p)
			if ctx.Count != 0 || ctx.Voices.first() != nil {
				t.Fatal("context destroyed before its voices")
			}
		case 14:
			v := (*AudioVoiceNative)(p)
			if v.Methods != methods || v.Context.Methods != methods || v.Kind != 1 || v.Refill == nil || v.Restart == nil || v.Finish == nil || v.Master != &audioGlobalOwner.Group {
				t.Fatal("voice native construction")
			}
			if voiceFailure {
				return -13
			}
		case 15:
			if (*AudioVoiceNative)(p).Context.Methods != methods {
				t.Fatal("voice destruction native context")
			}
		case 7, 8, 9:
			if (*AudioVoiceNative)(p).Methods != methods {
				t.Fatal("voice service native argument")
			}
		default:
			t.Fatalf("unexpected callback %d", kind)
		}
		return 0
	}
	deviceFailure = true
	if asnPointer(6, unsafe.Pointer(class), nil, 0, 0, 0) != nil || counts[11] != 1 || audioGlobalOwner.Devices.first() != nil {
		t.Fatal("failed device rollback")
	}
	deviceFailure = false
	d := (*AudioDeviceNative)(asnPointer(6, unsafe.Pointer(class), nil, 0, 0, 0))
	if d == nil || d.Limit != 2 || class.State != 0x81 || memmap.Uint32(0x5D4594, 1193332) != 1 {
		t.Fatal("registered device state")
	}
	format := (*[7]uint32)(a.take(28))
	*format = [7]uint32{4, 0, 22050, 1, 2, 0, 0}
	contextFailure = true
	if asnPointer(7, nil, unsafe.Pointer(format), 0, 0, 0) != nil || d.Count != 0 || d.Contexts[0] != nil || counts[13] != 1 {
		t.Fatal("failed context rollback")
	}
	contextFailure = false
	ctx := (*AudioContextNative)(asnPointer(7, nil, unsafe.Pointer(format), 0, 0, 0))
	if ctx == nil || ctx.Format60 != *format || ctx.Users != 1 || d.Count != 1 {
		t.Fatal("context format/refcount")
	}
	if asnPointer(7, nil, nil, 0, 0, 0) != unsafe.Pointer(ctx) || asnPointer(7, nil, nil, -1, 0, 0) != unsafe.Pointer(ctx) || ctx.Users != 3 || counts[12] != 2 {
		t.Fatal("context reuse/default index")
	}
	other := (*AudioContextNative)(asnPointer(7, nil, unsafe.Pointer(format), 1, 0, 0))
	if other == nil || other == ctx || other.Index != 1 || other.Priority != 1 || asnPointer(7, nil, nil, 2, 0, 0) != nil || d.Count != 2 {
		t.Fatal("second slot / index limit")
	}
	voiceFailure = true
	if asnPointer(5, unsafe.Pointer(ctx), nil, 0, 0, 0) != nil || counts[15] != 1 || ctx.Count != 0 {
		t.Fatal("failed voice rollback")
	}
	voiceFailure = false
	if got := asnScalar(7, unsafe.Pointer(ctx), nil, 10); got != 2 || ctx.Count != 2 || asnScalar(7, unsafe.Pointer(ctx), nil, 10) != 0 {
		t.Fatal("voice capacity")
	}
	first := (*AudioVoiceNative)(unsafe.Pointer(ctx.Voices.first()))
	second := (*AudioVoiceNative)(unsafe.Pointer(first.Link.next()))
	if first == nil || second == nil || second.Link.next() != nil || asnPointer(8, unsafe.Pointer(ctx), nil, -1, 0, 0) != unsafe.Pointer(first) {
		t.Fatal("voice list/acquisition")
	}
	first.Flags = 0x10
	if asnPointer(8, unsafe.Pointer(ctx), nil, -1, 0, 0) != unsafe.Pointer(second) {
		t.Fatal("first free voice")
	}
	first.Flags, second.Flags = 1, 1
	first.Priority, second.Priority = 10, 20
	if asnPointer(8, unsafe.Pointer(ctx), nil, 1, 0, 0) != unsafe.Pointer(first) {
		t.Fatal("active priority order")
	}
	second.Priority = 10
	first.Mixed.Timers[0].Current, second.Mixed.Timers[0].Current = 0x2000, 0x1000
	if asnPointer(8, unsafe.Pointer(ctx), nil, 1, 0, 0) != unsafe.Pointer(second) {
		t.Fatal("same-priority gain threshold")
	}
	first.Flags, second.Flags = 1, 0
	buffer := (*AudioBufferNative)(a.take(unsafe.Sizeof(AudioBufferNative{})))
	asnPointer(19, unsafe.Pointer(buffer), nil, 0, 0, 0)
	first.Buffer = buffer
	ctx.Busy = 1
	if asnScalar(8, unsafe.Pointer(ctx), nil, 0) != -2146304000 {
		t.Fatal("busy context guard")
	}
	ctx.Busy = 0
	if asnScalar(8, unsafe.Pointer(ctx), nil, 0) != 0 || counts[7] != 1 || ctx.Last != 0x100000047 || ctx.Delta != 0x100000047 {
		t.Fatal("native context refresh/full tick width")
	}
	if asnScalar(8, unsafe.Pointer(ctx), nil, 0) != 0 || counts[7] != 1 {
		t.Fatal("interval gate")
	}
	Set_dword_5d4594_805984(unsafe.Pointer(ctx))
	asnPointer(25, nil, nil, 0, 0, 0)
	if first.Buffer != nil || counts[8] != 1 || Get_dword_5d4594_805984() != unsafe.Pointer(ctx) {
		t.Fatal("public ordinary-context stop wrapper")
	}
	asnPointer(24, nil, nil, 0, 0, 0)
	if Get_dword_5d4594_805984() != nil || d.Count != 1 {
		t.Fatal("public ordinary-context destroy/clear wrapper")
	}
	asnPointer(25, nil, nil, 0, 0, 0)
	asnPointer(24, nil, nil, 0, 0, 0)
	if asnScalar(13, nil, nil, 0) != 0 || d.Count != 0 || counts[13] != 3 || counts[15] != 3 || counts[8] != 2 || audioGlobalOwner.Contexts.first() != nil {
		t.Fatal("context/voice natural teardown")
	}
	asnPointer(22, nil, nil, 0, 0, 0)
	if counts[11] != 2 || class.State != 0x80 || memmap.Uint32(0x5D4594, 1193332) != 0 || audioGlobalOwner.Devices.first() != nil {
		t.Fatal("device teardown")
	}
}
