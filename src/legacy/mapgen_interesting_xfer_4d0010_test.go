package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type interestingXferTestMarkers4D0010 struct {
	values [8]byte
}

func (m *interestingXferTestMarkers4D0010) ptr(ind int) unsafe.Pointer {
	return unsafe.Pointer(&m.values[ind])
}

func interestingXferTestDeps4D0010(
	first *server.Object,
	xfers map[*server.Object]unsafe.Pointer,
	markers *interestingXferTestMarkers4D0010,
) interestingXferNativeDeps4D0010 {
	return interestingXferNativeDeps4D0010{
		firstPending: func() *server.Object { return first },
		nextPending:  (*server.Object).Next,
		typeXfer: func(obj *server.Object) unsafe.Pointer {
			return xfers[obj]
		},
		findPendingObject: func(scriptID uint32) *server.Object {
			for obj := first; obj != nil; obj = obj.Next() {
				if obj.ScriptIDVal == int32(scriptID) {
					return obj
				}
			}
			return nil
		},
		findPendingWaypoint: func(uint32) *server.Waypoint { return nil },
		mapWallSize: func() *[2]uint32 {
			panic("unexpected map-wall lookup")
		},
		setElevatorLink:    func(*server.Object, unsafe.Pointer, *server.Object) {},
		setTransporterLink: func(*server.Object, *server.TransporterUpdateData, *server.Object) {},
		elevatorXfer:       markers.ptr(0),
		elevatorShaftXfer:  markers.ptr(1),
		transporterXfer:    markers.ptr(2),
		holeXfer:           markers.ptr(3),
		exitXfer:           markers.ptr(4),
		moverXfer:          markers.ptr(5),
		glyphXfer:          markers.ptr(6),
	}
}

func TestInterestingXfer4D0010RenumbersThenResolvesIndependentNativeLinks(t *testing.T) {
	markers := &interestingXferTestMarkers4D0010{}
	elevatorData := &server.ElevatorUpdateData{Field_1: 0xffffffff, Field_2: 20}
	shaftData := &server.ElevatorShaftUpdateData{Field_1: 0xeeeeeeee, Field_2: 10}
	transporterData := &server.TransporterUpdateData{TargetPE32: 0xdddddddd, TargetExtent: 40}
	plain := &server.Object{Extent: 40}
	transporter := &server.Object{
		Extent:     30,
		UpdateData: unsafe.Pointer(transporterData),
		ObjNext:    plain,
	}
	shaft := &server.Object{
		Extent:     20,
		UpdateData: unsafe.Pointer(shaftData),
		ObjNext:    transporter,
	}
	elevator := &server.Object{
		Extent:     10,
		UpdateData: unsafe.Pointer(elevatorData),
		ObjNext:    shaft,
	}
	xfers := map[*server.Object]unsafe.Pointer{
		elevator:    markers.ptr(0),
		shaft:       markers.ptr(1),
		transporter: markers.ptr(2),
		plain:       markers.ptr(7),
	}
	deps := interestingXferTestDeps4D0010(elevator, xfers, markers)
	type elevatorLink struct {
		data   unsafe.Pointer
		target *server.Object
	}
	elevatorLinks := make(map[*server.Object]elevatorLink)
	transporterLinks := make(map[*server.Object]*server.Object)
	deps.setElevatorLink = func(obj *server.Object, data unsafe.Pointer, target *server.Object) {
		elevatorLinks[obj] = elevatorLink{data: data, target: target}
	}
	deps.setTransporterLink = func(obj *server.Object, data *server.TransporterUpdateData, target *server.Object) {
		if data != transporterData {
			t.Fatalf("transporter data = %p, want %p", data, transporterData)
		}
		transporterLinks[obj] = target
	}

	if got := interestingXferNative4D0010(nil, 100, deps); got != 104 {
		t.Fatalf("next extent = %d, want 104", got)
	}
	objects := []*server.Object{elevator, shaft, transporter, plain}
	oldExtents := []int32{10, 20, 30, 40}
	newExtents := []uint32{100, 101, 102, 103}
	for i, obj := range objects {
		if obj.ScriptIDVal != oldExtents[i] || obj.Extent != newExtents[i] {
			t.Fatalf("object %d script/extent = %d/%d, want %d/%d",
				i, obj.ScriptIDVal, obj.Extent, oldExtents[i], newExtents[i])
		}
	}
	if got := elevatorLinks[elevator]; got.data != unsafe.Pointer(elevatorData) || got.target != shaft {
		t.Fatalf("elevator data/target = %p/%p, want %p/%p",
			got.data, got.target, elevatorData, shaft)
	}
	if got := elevatorLinks[shaft]; got.data != unsafe.Pointer(shaftData) || got.target != elevator {
		t.Fatalf("shaft data/target = %p/%p, want %p/%p",
			got.data, got.target, shaftData, elevator)
	}
	if got := transporterLinks[transporter]; got != plain {
		t.Fatalf("transporter target = %p, want arbitrary-class target %p", got, plain)
	}
	if elevatorData.Field_1 != 0 || elevatorData.Field_2 != shaft.Extent {
		t.Fatalf("elevator PE32/extent = %#x/%d, want 0/%d",
			elevatorData.Field_1, elevatorData.Field_2, shaft.Extent)
	}
	if shaftData.Field_1 != 0 || shaftData.Field_2 != elevator.Extent {
		t.Fatalf("shaft PE32/extent = %#x/%d, want 0/%d",
			shaftData.Field_1, shaftData.Field_2, elevator.Extent)
	}
	if transporterData.TargetPE32 != 0 || transporterData.TargetExtent != plain.Extent {
		t.Fatalf("transporter PE32/extent = %#x/%d, want 0/%d",
			transporterData.TargetPE32, transporterData.TargetExtent, plain.Extent)
	}
}

func TestInterestingXfer4D0010UsesEntryDataAndOriginalStoreOrder(t *testing.T) {
	markers := &interestingXferTestMarkers4D0010{}
	entryData := &server.ElevatorUpdateData{Field_1: 0xffffffff, Field_2: 22}
	replacementData := &server.ElevatorUpdateData{Field_1: 0xeeeeeeee, Field_2: 0x55667788}
	target := &server.Object{Extent: 22}
	source := &server.Object{Extent: 11, UpdateData: unsafe.Pointer(entryData), ObjNext: target}
	xfers := map[*server.Object]unsafe.Pointer{
		source: markers.ptr(0),
		target: markers.ptr(7),
	}
	deps := interestingXferTestDeps4D0010(source, xfers, markers)
	deps.findPendingObject = func(scriptID uint32) *server.Object {
		if scriptID != 22 {
			t.Fatalf("lookup script ID = %d, want 22", scriptID)
		}
		source.UpdateData = unsafe.Pointer(replacementData)
		return target
	}
	var callbackData unsafe.Pointer
	deps.setElevatorLink = func(obj *server.Object, data unsafe.Pointer, gotTarget *server.Object) {
		if obj != source || gotTarget != target {
			t.Fatalf("callback object/target = %p/%p, want %p/%p", obj, gotTarget, source, target)
		}
		callbackData = data
		if entryData.Field_1 != 0 || entryData.Field_2 != 22 {
			t.Fatalf("callback observed PE32/extent = %#x/%d, want pointer-store state 0/22",
				entryData.Field_1, entryData.Field_2)
		}
	}

	if got := interestingXferNative4D0010(nil, 500, deps); got != 502 {
		t.Fatalf("next extent = %d, want 502", got)
	}
	if callbackData != unsafe.Pointer(entryData) {
		t.Fatalf("callback data = %p, want entry data %p", callbackData, entryData)
	}
	if entryData.Field_1 != 0 || entryData.Field_2 != target.Extent {
		t.Fatalf("entry PE32/extent = %#x/%d, want 0/%d",
			entryData.Field_1, entryData.Field_2, target.Extent)
	}
	if replacementData.Field_1 != 0xeeeeeeee || replacementData.Field_2 != 0x55667788 {
		t.Fatalf("replacement data changed: %+v", *replacementData)
	}
}

func TestInterestingXfer4D0010FailureClearsExtentBeforeNativeLink(t *testing.T) {
	markers := &interestingXferTestMarkers4D0010{}
	data := &server.TransporterUpdateData{TargetPE32: 0xffffffff, TargetExtent: 0xf1234567}
	transporter := &server.Object{Extent: 7, UpdateData: unsafe.Pointer(data)}
	deps := interestingXferTestDeps4D0010(
		transporter,
		map[*server.Object]unsafe.Pointer{transporter: markers.ptr(2)},
		markers,
	)
	deps.setTransporterLink = func(obj *server.Object, gotData *server.TransporterUpdateData, target *server.Object) {
		if obj != transporter || gotData != data || target != nil {
			t.Fatalf("callback object/data/target = %p/%p/%p", obj, gotData, target)
		}
		if data.TargetExtent != 0 || data.TargetPE32 != 0 {
			t.Fatalf("callback observed PE32/extent = %#x/%#x, want 0/0",
				data.TargetPE32, data.TargetExtent)
		}
	}

	if got := interestingXferNative4D0010(nil, 80, deps); got != 81 {
		t.Fatalf("next extent = %d, want 81", got)
	}
	if data.TargetExtent != 0 || data.TargetPE32 != 0 {
		t.Fatalf("final PE32/extent = %#x/%#x, want 0/0", data.TargetPE32, data.TargetExtent)
	}
}

func TestInterestingXfer4D0010UsesNativeMoverAndGlyphLayouts(t *testing.T) {
	markers := &interestingXferTestMarkers4D0010{}
	holeData := &server.HoleCollideData{DestinationX: -5, DestinationY: 7}
	exitData := &server.ExitCollideData{DestinationX: 1.25, DestinationY: -2.5}
	moverTargetSentinel := &server.Object{}
	moverData := &server.MoverUpdateData{
		Field_2: 77,
		Field_7: moverTargetSentinel,
		Field_8: 11,
	}
	glyphObjectSentinel := &server.Object{}
	glyphData := &server.GlyphInitData{}
	glyphData.SpellArg.Obj = glyphObjectSentinel
	glyphData.SpellArg.Pos.X = 0.5
	glyphData.SpellArg.Pos.Y = -0.75
	glyph := &server.Object{Extent: 44, InitData: unsafe.Pointer(glyphData)}
	mover := &server.Object{Extent: 33, UpdateData: unsafe.Pointer(moverData), ObjNext: glyph}
	exit := &server.Object{Extent: 22, CollideData: unsafe.Pointer(exitData), ObjNext: mover}
	hole := &server.Object{Extent: 11, CollideData: unsafe.Pointer(holeData), ObjNext: exit}
	xfers := map[*server.Object]unsafe.Pointer{
		hole:  markers.ptr(3),
		exit:  markers.ptr(4),
		mover: markers.ptr(5),
		glyph: markers.ptr(6),
	}
	deps := interestingXferTestDeps4D0010(hole, xfers, markers)
	waypoint := &server.Waypoint{Index: 900}
	deps.findPendingWaypoint = func(id uint32) *server.Waypoint {
		if id == 77 {
			return waypoint
		}
		return nil
	}
	wallCalls := 0
	wallSize := &[2]uint32{10, 20}
	deps.mapWallSize = func() *[2]uint32 {
		wallCalls++
		return wallSize
	}

	origin := &[2]uint32{1000, 2000}
	if got := interestingXferNative4D0010(origin, 100, deps); got != 104 {
		t.Fatalf("next extent = %d, want 104", got)
	}
	if holeData.DestinationX != 765 || holeData.DestinationY != 1547 {
		t.Fatalf("hole destination = %d/%d, want 765/1547",
			holeData.DestinationX, holeData.DestinationY)
	}
	if exitData.DestinationX != 771.25 || exitData.DestinationY != 1537.5 {
		t.Fatalf("exit destination = %v/%v, want 771.25/1537.5",
			exitData.DestinationX, exitData.DestinationY)
	}
	if moverData.Field_2 != 900 || moverData.Field_8 != hole.Extent {
		t.Fatalf("mover waypoint/object = %d/%d, want 900/%d",
			moverData.Field_2, moverData.Field_8, hole.Extent)
	}
	if moverData.Field_7 != moverTargetSentinel {
		t.Fatalf("mover native Field_7 pointer changed: got %p, want %p",
			moverData.Field_7, moverTargetSentinel)
	}
	if glyphData.SpellArg.Obj != glyphObjectSentinel {
		t.Fatalf("glyph native object pointer changed: got %p, want %p",
			glyphData.SpellArg.Obj, glyphObjectSentinel)
	}
	if glyphData.SpellArg.Pos.X != 770.5 || glyphData.SpellArg.Pos.Y != 1539.25 {
		t.Fatalf("glyph destination = %v/%v, want 770.5/1539.25",
			glyphData.SpellArg.Pos.X, glyphData.SpellArg.Pos.Y)
	}
	if wallCalls != 3 {
		t.Fatalf("map-wall calls = %d, want one per Hole/Exit/Glyph (3)", wallCalls)
	}
}

func TestInterestingXfer4D0010LoadsWallHeightAfterXStore(t *testing.T) {
	markers := &interestingXferTestMarkers4D0010{}
	data := &server.HoleCollideData{DestinationX: 1, DestinationY: 7}
	hole := &server.Object{Extent: 11, CollideData: unsafe.Pointer(data)}
	deps := interestingXferTestDeps4D0010(
		hole,
		map[*server.Object]unsafe.Pointer{hole: markers.ptr(3)},
		markers,
	)
	// Alias wall height to DestinationX. GAME.EXE loads height only after
	// storing translated X, so the Y calculation must observe the new 1001.
	deps.mapWallSize = func() *[2]uint32 {
		return (*[2]uint32)(unsafe.Add(unsafe.Pointer(data), 4))
	}
	origin := &[2]uint32{1000, 2000}
	interestingXferNative4D0010(origin, 100, deps)
	if data.DestinationX != 1001 || data.DestinationY != -21016 {
		t.Fatalf("aliased destination = %d/%d, want 1001/-21016",
			data.DestinationX, data.DestinationY)
	}
}

func TestInterestingXferTranslateFloat4D0010UsesSingleBinary32Spill(t *testing.T) {
	value := float32(2)
	delta := int32(1<<30 + 63)
	got := interestingXferTranslateFloat4D0010(value, delta)
	if bits := math.Float32bits(got); bits != 0x4e800001 {
		t.Fatalf("translated bits = %#08x, want x87-compatible 0x4e800001", bits)
	}
	if naive := float32(delta) + value; math.Float32bits(naive) != 0x4e800000 {
		t.Fatalf("test vector lost its binary32-intermediate distinction: %#08x", math.Float32bits(naive))
	}
}

func TestInterestingXfer4D0010NativeRecordLayouts(t *testing.T) {
	wantMoverSize := uintptr(36)
	wantMoverObject := uintptr(28)
	wantMoverExtent := uintptr(32)
	wantGlyphPoint := uintptr(28)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantMoverSize = 48
		wantMoverObject = 32
		wantMoverExtent = 40
		wantGlyphPoint = 32
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"MoverUpdateData size", unsafe.Sizeof(server.MoverUpdateData{}), wantMoverSize},
		{"MoverUpdateData.Field_2", unsafe.Offsetof(server.MoverUpdateData{}.Field_2), 8},
		{"MoverUpdateData.Field_7", unsafe.Offsetof(server.MoverUpdateData{}.Field_7), wantMoverObject},
		{"MoverUpdateData.Field_8", unsafe.Offsetof(server.MoverUpdateData{}.Field_8), wantMoverExtent},
		{"GlyphInitData point", unsafe.Offsetof(server.GlyphInitData{}.SpellArg) + unsafe.Offsetof(server.SpellAcceptArg{}.Pos), wantGlyphPoint},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestInterestingXfer4D0010LeavesOriginLazyForUnrelatedTypes(t *testing.T) {
	markers := &interestingXferTestMarkers4D0010{}
	plain := &server.Object{Extent: 0xffffffff}
	deps := interestingXferTestDeps4D0010(
		plain,
		map[*server.Object]unsafe.Pointer{plain: markers.ptr(7)},
		markers,
	)
	if got := interestingXferNative4D0010(nil, -1, deps); got != 0 {
		t.Fatalf("wrapped next extent = %d, want 0", got)
	}
	if plain.ScriptIDVal != -1 || plain.Extent != 0xffffffff {
		t.Fatalf("plain script/extent = %d/%#x, want -1/0xffffffff",
			plain.ScriptIDVal, plain.Extent)
	}
}
