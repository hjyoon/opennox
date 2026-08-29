package server

import (
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func newTransporterStateServer(t *testing.T) *Server {
	t.Helper()
	handle := atomic.AddUintptr(&serverLast, 1)
	s := &Server{handle: handle}
	s.Objs.init(handle)
	servers.Store(handle, s)
	t.Cleanup(func() { servers.Delete(handle) })
	return s
}

func TestTransporterUpdateDataLayoutAndNativeTargetIdentity(t *testing.T) {
	if got := unsafe.Sizeof(TransporterUpdateData{}); got != 20 {
		t.Fatalf("Transporter update-data size = %d, want 20", got)
	}
	if got := unsafe.Offsetof(TransporterUpdateData{}.TargetPE32); got != 12 {
		t.Fatalf("TargetPE32 offset = %d, want 12", got)
	}
	if got := unsafe.Offsetof(TransporterUpdateData{}.TargetExtent); got != 16 {
		t.Fatalf("TargetExtent offset = %d, want 16", got)
	}

	s := newTransporterStateServer(t)
	entryData := &TransporterUpdateData{TargetPE32: 0xffffffff, TargetExtent: 0x11223344}
	liveData := &TransporterUpdateData{TargetPE32: 0xeeeeeeee, TargetExtent: 0x55667788}
	obj := &Object{UpdateData: unsafe.Pointer(entryData), serverHandle: s.handle}
	target := &Object{Extent: entryData.TargetExtent, serverHandle: s.handle}

	obj.SetTransporterTargetFor(entryData, target)
	if entryData.TargetPE32 != 0 || entryData.TargetExtent != 0x11223344 {
		t.Fatalf("entry PE32/extent = %#x/%#x, want 0/0x11223344",
			entryData.TargetPE32, entryData.TargetExtent)
	}
	if got := obj.TransporterTargetFor(entryData); got != target {
		t.Fatalf("entry target = %p, want %p", got, target)
	}

	obj.UpdateData = unsafe.Pointer(liveData)
	if got := obj.TransporterTarget(); got != nil {
		t.Fatalf("replacement-data target = %p, want nil", got)
	}
	if got := obj.TransporterTargetFor(entryData); got != target {
		t.Fatalf("cached-data target = %p, want %p", got, target)
	}
	obj.SetTransporterTargetFor(liveData, nil)
	if liveData.TargetPE32 != 0 || liveData.TargetExtent != 0x55667788 {
		t.Fatalf("live PE32/extent = %#x/%#x, want 0/0x55667788",
			liveData.TargetPE32, liveData.TargetExtent)
	}
	if got := obj.TransporterTargetFor(entryData); got != nil {
		t.Fatalf("stale-data target = %p, want nil", got)
	}

	standaloneData := &TransporterUpdateData{TargetPE32: 0xabcdef01, TargetExtent: 7}
	standalone := &Object{UpdateData: unsafe.Pointer(standaloneData)}
	standalone.SetTransporterTarget(target)
	if standaloneData.TargetPE32 != 0 || standaloneData.TargetExtent != 7 || standalone.TransporterTarget() != nil {
		t.Fatalf("standalone PE32/extent/target = %#x/%d/%p, want 0/7/nil",
			standaloneData.TargetPE32, standaloneData.TargetExtent, standalone.TransporterTarget())
	}
}

func TestAttachPendingTransporterKeepsExtentBesideNativeTarget(t *testing.T) {
	s := newTransporterStateServer(t)
	const targetExtent = uint32(0xf1234567)
	sourceData := &TransporterUpdateData{TargetPE32: 0xffffffff, TargetExtent: targetExtent}
	targetData := &TransporterUpdateData{TargetPE32: 0xeeeeeeee}
	target := &Object{
		ObjClass:     object.ClassTransporter,
		Extent:       targetExtent,
		UpdateData:   unsafe.Pointer(targetData),
		serverHandle: s.handle,
	}
	source := &Object{
		ObjClass:     object.ClassTransporter,
		UpdateData:   unsafe.Pointer(sourceData),
		ObjNext:      target,
		serverHandle: s.handle,
	}
	s.Objs.Pending = source

	s.AttachPending()

	if got := source.TransporterTarget(); got != target {
		t.Fatalf("source target = %p, want %p", got, target)
	}
	if sourceData.TargetPE32 != 0 || sourceData.TargetExtent != targetExtent {
		t.Fatalf("source PE32/extent = %#x/%#x, want 0/%#x",
			sourceData.TargetPE32, sourceData.TargetExtent, targetExtent)
	}
	if target.TransporterTarget() != nil || targetData.TargetPE32 != 0 {
		t.Fatalf("target link/PE32 = %p/%#x, want nil/0",
			target.TransporterTarget(), targetData.TargetPE32)
	}
}

func TestAttachPendingTransporterClearsUnresolvedTarget(t *testing.T) {
	s := newTransporterStateServer(t)
	const missingExtent = uint32(0xf7654321)
	sourceData := &TransporterUpdateData{TargetPE32: 0xffffffff, TargetExtent: missingExtent}
	source := &Object{
		ObjClass:     object.ClassTransporter,
		UpdateData:   unsafe.Pointer(sourceData),
		serverHandle: s.handle,
	}
	staleTarget := &Object{Extent: missingExtent, serverHandle: s.handle}
	source.SetTransporterTargetFor(sourceData, staleTarget)
	sourceData.TargetPE32 = 0xffffffff
	s.Objs.Pending = source

	s.AttachPending()

	if got := source.TransporterTarget(); got != nil {
		t.Fatalf("source target = %p, want nil", got)
	}
	if sourceData.TargetPE32 != 0 || sourceData.TargetExtent != 0 {
		t.Fatalf("source PE32/extent = %#x/%#x, want 0/0",
			sourceData.TargetPE32, sourceData.TargetExtent)
	}
}
