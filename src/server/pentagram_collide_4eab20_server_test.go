package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPentagramCollide4EAB20NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantUpdateData := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantUpdateData = 872
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"PentagramUpdateDataPrefix size", unsafe.Sizeof(PentagramUpdateDataPrefix{}), 8},
		{"PentagramUpdateDataPrefix.Triggered", unsafe.Offsetof(PentagramUpdateDataPrefix{}.Triggered), 4},
		{"PentagramUpdateData size", unsafe.Sizeof(PentagramUpdateData{}), 24},
		{"PentagramUpdateData.Triggered", unsafe.Offsetof(PentagramUpdateData{}.Triggered), 4},
		{"PentagramUpdateData.AnimationFrame", unsafe.Offsetof(PentagramUpdateData{}.AnimationFrame), 8},
		{"PentagramUpdateData.DestinationPE32", unsafe.Offsetof(PentagramUpdateData{}.DestinationPE32), 12},
		{"PentagramUpdateData.DestinationExtent", unsafe.Offsetof(PentagramUpdateData{}.DestinationExtent), 16},
		{"PentagramUpdateData.AnimationStep", unsafe.Offsetof(PentagramUpdateData{}.AnimationStep), 20},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPentagramUpdateDataNativeDestinationUsesTransporterSidecar(t *testing.T) {
	s := newTransporterStateServer(t)
	entryData := &PentagramUpdateData{
		DestinationPE32:   math.MaxUint32,
		DestinationExtent: 0xf1234567,
		AnimationStep:     8,
	}
	liveData := &PentagramUpdateData{
		DestinationPE32:   0xeeeeeeee,
		DestinationExtent: 0x76543210,
	}
	obj := &Object{UpdateData: unsafe.Pointer(entryData), serverHandle: s.handle}
	target := &Object{Extent: entryData.DestinationExtent, serverHandle: s.handle}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= math.MaxUint32 {
		t.Fatalf("test destination address = %#x, want a value above PE32", uintptr(unsafe.Pointer(target)))
	}

	obj.SetPentagramDestinationFor(entryData, target)
	if entryData.DestinationPE32 != 0 || entryData.DestinationExtent != 0xf1234567 || entryData.AnimationStep != 8 {
		t.Fatalf("entry PE32/extent/step = %#x/%#x/%d, want 0/0xf1234567/8",
			entryData.DestinationPE32, entryData.DestinationExtent, entryData.AnimationStep)
	}
	if got := obj.PentagramDestinationFor(entryData); got != target {
		t.Fatalf("entry destination = %p, want %p", got, target)
	}

	obj.UpdateData = unsafe.Pointer(liveData)
	if got := obj.PentagramDestination(); got != nil {
		t.Fatalf("replacement-data destination = %p, want nil", got)
	}
	if got := obj.PentagramDestinationFor(entryData); got != target {
		t.Fatalf("cached-data destination = %p, want %p", got, target)
	}
	obj.SetPentagramDestinationFor(liveData, nil)
	if liveData.DestinationPE32 != 0 || liveData.DestinationExtent != 0x76543210 {
		t.Fatalf("live PE32/extent = %#x/%#x, want 0/0x76543210",
			liveData.DestinationPE32, liveData.DestinationExtent)
	}
	if got := obj.PentagramDestinationFor(entryData); got != nil {
		t.Fatalf("stale-data destination = %p, want nil", got)
	}
}

func TestAttachPendingPentagramKeepsExtentBesideNativeDestination(t *testing.T) {
	s := newTransporterStateServer(t)
	const destinationExtent = uint32(0xf2345678)
	sourceData := &PentagramUpdateData{
		DestinationPE32:   math.MaxUint32,
		DestinationExtent: destinationExtent,
		AnimationStep:     7,
	}
	destinationData := &PentagramUpdateData{DestinationPE32: 0xeeeeeeee}
	destination := &Object{
		ObjClass:     object.ClassTransporter,
		Extent:       destinationExtent,
		UpdateData:   unsafe.Pointer(destinationData),
		serverHandle: s.handle,
	}
	source := &Object{
		ObjClass:     object.ClassTransporter,
		UpdateData:   unsafe.Pointer(sourceData),
		ObjNext:      destination,
		serverHandle: s.handle,
	}
	s.Objs.Pending = source

	s.AttachPending()

	if got := source.PentagramDestination(); got != destination {
		t.Fatalf("source destination = %p, want %p", got, destination)
	}
	if sourceData.DestinationPE32 != 0 || sourceData.DestinationExtent != destinationExtent ||
		sourceData.AnimationStep != 7 {
		t.Fatalf("source PE32/extent/step = %#x/%#x/%d, want 0/%#x/7",
			sourceData.DestinationPE32, sourceData.DestinationExtent,
			sourceData.AnimationStep, destinationExtent)
	}
	if destination.PentagramDestination() != nil || destinationData.DestinationPE32 != 0 {
		t.Fatalf("destination link/PE32 = %p/%#x, want nil/0",
			destination.PentagramDestination(), destinationData.DestinationPE32)
	}
}

func TestTeleportPentagramUpdate53BEF0UsesNativeDestinationSidecar(t *testing.T) {
	s := newTransporterStateServer(t)
	const frame = uint32(0xf1234567)
	s.SetFrame(frame)
	sourceData := &PentagramUpdateData{Triggered: 1, DestinationExtent: 0x87654321}
	destinationData := &PentagramUpdateData{
		State:          9,
		AnimationFrame: 8,
		AnimationTick:  7,
	}
	destination := &Object{
		UpdateData:   unsafe.Pointer(destinationData),
		serverHandle: s.handle,
	}
	source := &Object{
		ObjFlags:     object.FlagEnabled,
		UpdateData:   unsafe.Pointer(sourceData),
		serverHandle: s.handle,
	}
	source.SetPentagramDestinationFor(sourceData, destination)

	got := s.TeleportPentagramUpdate53BEF0(source, PentagramUpdateRuntime53BEF0{})
	if uint32(got) != frame {
		t.Fatalf("return = %#x, want %#x", uint32(got), frame)
	}
	if sourceData.State != 1 || sourceData.Triggered != 0 || source.Field34 != frame {
		t.Fatalf("source state/trigger/frame = %d/%d/%#x, want 1/0/%#x",
			sourceData.State, sourceData.Triggered, source.Field34, frame)
	}
	if destinationData.State != 2 || destinationData.AnimationFrame != 0 ||
		destinationData.AnimationTick != 0 || destination.Field34 != frame {
		t.Fatalf("destination data/frame = %+v/%#x", destinationData, destination.Field34)
	}
	if sourceData.DestinationPE32 != 0 || sourceData.DestinationExtent != 0x87654321 {
		t.Fatalf("source PE32/extent = %#x/%#x, want 0/0x87654321",
			sourceData.DestinationPE32, sourceData.DestinationExtent)
	}
}

func TestTeleportPentagramUpdate53BEF0DoesNotInterpretPE32Slot(t *testing.T) {
	s := newTransporterStateServer(t)
	data := &PentagramUpdateData{
		Triggered:         1,
		DestinationPE32:   0x37c1,
		DestinationExtent: 0x37c1,
	}
	source := &Object{
		ObjFlags:     object.FlagEnabled,
		UpdateData:   unsafe.Pointer(data),
		serverHandle: s.handle,
	}

	if got := s.TeleportPentagramUpdate53BEF0(source, PentagramUpdateRuntime53BEF0{}); got != 0 {
		t.Fatalf("return = %d, want 0", got)
	}
	if data.Triggered != 0 || data.DestinationPE32 != 0x37c1 || data.DestinationExtent != 0x37c1 {
		t.Fatalf("data = %+v, want consumed trigger and untouched PE32 dwords", data)
	}
}

func TestPentagramCollideNative4EAB20UsesNativeUpdatePointer(t *testing.T) {
	data := &PentagramUpdateDataPrefix{
		Reserved0: [4]byte{1, 2, 3, 4},
		Triggered: 0xaabbccdd,
	}
	source := &Object{UpdateData: unsafe.Pointer(data), Field188: 0x11223344}
	target := &Object{Field188: 0x55667788}
	collision := &types.Pointf{X: 3, Y: 4}
	got := pentagramCollideNative4EAB20(source, target, collision)
	if got != source {
		t.Fatalf("return = %p, want %p", got, source)
	}
	if data.Reserved0 != [4]byte{1, 2, 3, 4} || data.Triggered != 1 {
		t.Fatalf("data = %+v", data)
	}
	if source.Field188 != 0x11223344 || target.Field188 != 0x55667788 || collision.X != 3 || collision.Y != 4 {
		t.Fatalf("state = source %#x target %#x collision %+v", source.Field188, target.Field188, collision)
	}
}

func TestPentagramCollide4EAB20ServerBinding(t *testing.T) {
	s := &Server{}
	data := &PentagramUpdateDataPrefix{Triggered: 9}
	source := &Object{UpdateData: unsafe.Pointer(data)}
	target := &Object{Field188: 0x89abcdef}
	collision := &types.Pointf{X: 5, Y: 6}
	s.PentagramCollide4EAB20(source, target, collision)
	if data.Triggered != 1 || target.Field188 != 0x89abcdef || collision.X != 5 || collision.Y != 6 {
		t.Fatalf("data/target/collision = %#x/%#x/%+v", data.Triggered, target.Field188, collision)
	}
}

func TestPentagramCollideNative4EAB20NilUpdateDataFaults(t *testing.T) {
	source := &Object{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update data did not fault")
		}
	}()
	pentagramCollideNative4EAB20(source, nil, nil)
}
