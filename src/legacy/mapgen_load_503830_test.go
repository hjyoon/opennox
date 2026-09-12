package legacy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type mapgenLoadServer503830 struct {
	Server
	freed *int
}

func (s mapgenLoadServer503830) Nox_xxx_free503F40() { *s.freed++ }

type mapgenWireSection503830 struct {
	name    string
	payload []byte
}

func mapgenLoadWireSections503830(name string, sections ...mapgenWireSection503830) []byte {
	extra := binary.LittleEndian.AppendUint32(nil, mapgenMagic502ED0)
	for _, value := range []uint32{7, 9, 10, 20, 10, 40, 30, 20, 30, 40} {
		extra = binary.LittleEndian.AppendUint32(extra, value)
	}
	var encoded []byte
	for _, section := range sections {
		encoded = append(encoded, byte(len(section.name)))
		encoded = append(encoded, section.name...)
		encoded = binary.LittleEndian.AppendUint32(encoded, uint32(len(section.payload)))
		encoded = append(encoded, section.payload...)
	}
	encoded = append(encoded, 0)
	for i := range encoded {
		encoded[i] ^= 126
	}
	extra = append(extra, encoded...)
	record := mapgenRecordWire502B10([]byte(name), 0x3f800000, 0x40000000, extra)
	record[4+1+len(name)+1] = 1 // no optional attachment block
	return mapgenStreamWire502B10(record)
}

func mapgenLoadWire503830(name, section string, payload []byte) []byte {
	if section == "" {
		return mapgenLoadWireSections503830(name)
	}
	return mapgenLoadWireSections503830(name, mapgenWireSection503830{section, payload})
}

func mapgenLoadWrite503830(t *testing.T, wire []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "AreaMap.dat")
	if err := os.WriteFile(path, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func mapgenLoadSetup503830(t *testing.T) *int {
	t.Helper()
	initMapgenReaderHandles502B10(t)
	oldGetServer := GetServer
	oldSection := Nox_xxx_mapReadSection_426EA0
	oldLoadObject := mapgenLoadPlaceObject503830
	oldCrypt := cryptfile.Global()
	freed := new(int)
	GetServer = func() Server { return mapgenLoadServer503830{freed: freed} }
	t.Cleanup(func() {
		GetServer = oldGetServer
		Nox_xxx_mapReadSection_426EA0 = oldSection
		mapgenLoadPlaceObject503830 = oldLoadObject
		cryptfile.SetGlobal(oldCrypt)
	})
	return freed
}

func TestMapgenLoad503830KnownSection(t *testing.T) {
	freed := mapgenLoadSetup503830(t)
	var seenName string
	var seenContext uintptr
	Nox_xxx_mapReadSection_426EA0 = func(context unsafe.Pointer, name string) (bool, error) {
		seenContext = uintptr(context)
		seenName = name
		buf := make([]byte, 2)
		if n, err := cryptfile.Global().ReadWrite(buf); err != nil || n != len(buf) || buf[0] != 0x42 || buf[1] != 0x99 {
			t.Fatalf("section payload = %x (%d, %v)", buf, n, err)
		}
		return true, nil
	}
	got := mapgenLoadViaC503830(mapgenLoadWrite503830(t, mapgenLoadWire503830("Target", "Known", []byte{0x42, 0x99})))
	if got.indexed != 1 || !got.loaded || !got.fileClosed || *freed != 1 || seenName != "Known" {
		t.Fatalf("known section = %+v, frees=%d, name=%q", got, *freed, seenName)
	}
	if got.selected != math.MaxUint32 || got.loadedIndex != 0 || got.walls != [2]uint32{7, 9} ||
		got.bounds != [8]uint32{10, 20, 10, 20, 30, 40, 10, 40} {
		t.Fatalf("map state = %+v", got)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && (got.recordsAddress <= math.MaxUint32 || seenContext <= math.MaxUint32) {
		t.Fatalf("C pointers narrowed: records=%#x context=%#x", got.recordsAddress, seenContext)
	}
}

func TestMapgenLoad503830UnknownObjectUsesNativeBounds(t *testing.T) {
	freed := mapgenLoadSetup503830(t)
	Nox_xxx_mapReadSection_426EA0 = func(_ unsafe.Pointer, _ string) (bool, error) { return false, nil }
	var seenName string
	var seenBounds [4]int32
	var seenPtr uintptr
	mapgenLoadPlaceObject503830 = func(name string, bounds unsafe.Pointer) bool {
		seenName = name
		seenPtr = uintptr(bounds)
		seenBounds = *(*[4]int32)(bounds)
		return true
	}
	got := mapgenLoadViaC503830(mapgenLoadWrite503830(t, mapgenLoadWire503830("Target", "Monster", nil)))
	if !got.loaded || !got.fileClosed || *freed != 1 || seenName != "Monster" || seenBounds != [4]int32{10, 20, 30, 40} {
		t.Fatalf("unknown object = %+v, frees=%d name=%q bounds=%v", got, *freed, seenName, seenBounds)
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && seenPtr <= math.MaxUint32 {
		t.Fatalf("C bounds pointer narrowed: %#x", seenPtr)
	}
}

func TestMapgenLoad503830UnknownObjectXferThenPlace(t *testing.T) {
	freed := mapgenLoadSetup503830(t)
	Nox_xxx_mapReadSection_426EA0 = func(_ unsafe.Pointer, _ string) (bool, error) { return false, nil }
	mapgenTestXferReset503830()
	obj := &server.Object{Xfer: mapgenTestXferFunc503830()}
	var events []string
	var xferBounds [4]int32
	var placeBounds ntype.Point32
	var entryBounds uintptr
	deps := mapgenLoadObjectDeps503830{
		newObject: func(name string) *server.Object {
			events = append(events, "new:"+name)
			return obj
		},
		xfer: func(got *server.Object, bounds unsafe.Pointer) error {
			events = append(events, "xfer")
			if got != obj {
				return errors.New("wrong object passed to xfer")
			}
			xferBounds = *(*[4]int32)(bounds)
			return obj.CallXfer(bounds)
		},
		freeObject: func(_ *server.Object) { events = append(events, "free") },
		placeObject: func(got *server.Object, bounds *ntype.Point32) int32 {
			events = append(events, "place")
			if got == obj {
				placeBounds = *bounds
			}
			return 0 // GAME.EXE ignores placement rejection here.
		},
	}
	mapgenLoadPlaceObject503830 = func(name string, bounds unsafe.Pointer) bool {
		entryBounds = uintptr(bounds)
		if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(bounds) <= math.MaxUint32) {
			t.Errorf("object or bounds pointer narrowed: object=%p bounds=%p", obj, bounds)
		}
		return mapgenLoadObjectWithDeps503830(name, bounds, deps)
	}
	got := mapgenLoadViaC503830(mapgenLoadWrite503830(t, mapgenLoadWire503830("Target", "TestObject", []byte{0x11, 0x22, 0x33})))
	seenObj, seenBounds, payloadRead := mapgenTestXferSnapshot503830()
	if !got.loaded || !got.fileClosed || *freed != 1 || strings.Join(events, ",") != "new:TestObject,xfer,place" ||
		xferBounds != [4]int32{10, 20, 30, 40} || placeBounds != (ntype.Point32{X: 10, Y: 20}) ||
		seenObj != uintptr(unsafe.Pointer(obj)) || seenBounds != entryBounds || payloadRead != [3]byte{0x11, 0x22, 0x33} {
		t.Fatalf("object xfer/placement = %+v, frees=%d, events=%v, xfer bounds=%v, place bounds=%+v, C object=%#x C bounds=%#x payload=%x",
			got, *freed, events, xferBounds, placeBounds, seenObj, seenBounds, payloadRead)
	}
}

func TestMapgenLoad503830MixedSectionsKeepCryptPosition(t *testing.T) {
	freed := mapgenLoadSetup503830(t)
	mapgenTestXferReset503830()
	obj := &server.Object{Xfer: mapgenTestXferFunc503830()}
	var events []string
	var context unsafe.Pointer
	Nox_xxx_mapReadSection_426EA0 = func(gotContext unsafe.Pointer, name string) (bool, error) {
		events = append(events, "dispatch:"+name)
		if context == nil {
			context = gotContext
		} else if context != gotContext {
			return false, fmt.Errorf("section context changed: %p to %p", context, gotContext)
		}
		var want []byte
		switch name {
		case "First":
			want = []byte{0x42, 0x99}
		case "Last":
			want = []byte{0xa5}
		default:
			return false, nil
		}
		buf := make([]byte, len(want))
		if n, err := cryptfile.Global().ReadWrite(buf); err != nil || n != len(buf) || !bytes.Equal(buf, want) {
			return false, fmt.Errorf("section %q payload %x (%d, %v), want %x", name, buf, n, err, want)
		}
		return true, nil
	}
	mapgenLoadPlaceObject503830 = func(name string, bounds unsafe.Pointer) bool {
		return mapgenLoadObjectWithDeps503830(name, bounds, mapgenLoadObjectDeps503830{
			newObject: func(name string) *server.Object {
				events = append(events, "new:"+name)
				return obj
			},
			xfer: func(obj *server.Object, bounds unsafe.Pointer) error {
				events = append(events, "xfer")
				return obj.CallXfer(bounds)
			},
			freeObject: func(*server.Object) { events = append(events, "free") },
			placeObject: func(_ *server.Object, bounds *ntype.Point32) int32 {
				events = append(events, "place")
				if *bounds != (ntype.Point32{X: 10, Y: 20}) {
					t.Errorf("placement bounds = %+v", *bounds)
				}
				return 1
			},
		})
	}
	wire := mapgenLoadWireSections503830("Target",
		mapgenWireSection503830{"First", []byte{0x42, 0x99}},
		mapgenWireSection503830{"TestObject", []byte{0x11, 0x22, 0x33}},
		mapgenWireSection503830{"Last", []byte{0xa5}},
	)
	got := mapgenLoadViaC503830(mapgenLoadWrite503830(t, wire))
	_, _, payloadRead := mapgenTestXferSnapshot503830()
	wantEvents := "dispatch:First,dispatch:TestObject,new:TestObject,xfer,place,dispatch:Last"
	if !got.loaded || !got.fileClosed || *freed != 1 || context == nil ||
		strings.Join(events, ",") != wantEvents || payloadRead != [3]byte{0x11, 0x22, 0x33} {
		t.Fatalf("mixed sections = %+v, frees=%d context=%p events=%v C payload=%x",
			got, *freed, context, events, payloadRead)
	}
}

func TestMapgenLoad503830ObjectLifecycleFailures(t *testing.T) {
	for _, tc := range []struct {
		name       string
		create     bool
		hasXfer    bool
		xferError  bool
		wantEvents string
	}{
		{name: "unknown type", wantEvents: "new"},
		{name: "missing xfer", create: true, wantEvents: "new,free"},
		{name: "failed xfer", create: true, hasXfer: true, xferError: true, wantEvents: "new,xfer,free"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			var marker byte
			obj := &server.Object{}
			if tc.hasXfer {
				obj.Xfer = unsafe.Pointer(&marker)
			}
			bounds := [4]int32{10, 20, 30, 40}
			deps := mapgenLoadObjectDeps503830{
				newObject: func(_ string) *server.Object {
					events = append(events, "new")
					if tc.create {
						return obj
					}
					return nil
				},
				xfer: func(_ *server.Object, _ unsafe.Pointer) error {
					events = append(events, "xfer")
					if tc.xferError {
						return errors.New("xfer failed")
					}
					return nil
				},
				freeObject: func(_ *server.Object) { events = append(events, "free") },
				placeObject: func(_ *server.Object, _ *ntype.Point32) int32 {
					events = append(events, "place")
					return 1
				},
			}
			if mapgenLoadObjectWithDeps503830("TestObject", unsafe.Pointer(&bounds[0]), deps) {
				t.Fatal("failed object lifecycle reported success")
			}
			if got := strings.Join(events, ","); got != tc.wantEvents {
				t.Fatalf("lifecycle events = %q, want %q", got, tc.wantEvents)
			}
		})
	}
}

func TestMapgenLoad503830OptionalAttachment(t *testing.T) {
	freed := mapgenLoadSetup503830(t)
	Nox_xxx_mapReadSection_426EA0 = func(_ unsafe.Pointer, _ string) (bool, error) {
		t.Fatal("zero-section record invoked a section handler")
		return false, nil
	}
	wire := mapgenLoadWire503830("Target", "", nil)
	attachment := []byte{0, 1, 2, 3, 0xff}
	block := binary.LittleEndian.AppendUint32(nil, uint32(len(attachment)))
	block = append(block, attachment...)
	magicAt := 4 + 4 + 1 + len("Target") + 2 + 8
	wire = append(wire[:magicAt], append(block, wire[magicAt:]...)...)
	binary.LittleEndian.PutUint32(wire[4:], binary.LittleEndian.Uint32(wire[4:])+uint32(len(block)))
	wire[4+4+1+len("Target")+1] = 2
	got := mapgenLoadViaC503830(mapgenLoadWrite503830(t, wire))
	if got.indexed != 1 || !got.loaded || !got.fileClosed || *freed != 1 || got.walls != [2]uint32{7, 9} {
		t.Fatalf("optional attachment = %+v, frees=%d", got, *freed)
	}
}

func TestMapgenLoad503830UnknownObjectFailure(t *testing.T) {
	freed := mapgenLoadSetup503830(t)
	Nox_xxx_mapReadSection_426EA0 = func(_ unsafe.Pointer, _ string) (bool, error) { return false, nil }
	mapgenLoadPlaceObject503830 = func(_ string, _ unsafe.Pointer) bool { return false }
	got := mapgenLoadViaC503830(mapgenLoadWrite503830(t, mapgenLoadWire503830("Target", "Missing", nil)))
	if got.loaded || !got.fileClosed || *freed != 1 {
		t.Fatalf("missing object = %+v, frees=%d", got, *freed)
	}
}

func TestMapgenLoad503830RejectsMalformedRecords(t *testing.T) {
	for _, tc := range []struct {
		name string
		wire func() []byte
	}{
		{"bad magic", func() []byte {
			wire := mapgenLoadWire503830("Target", "", nil)
			wire[4+4+1+len("Target")+2+8] = 0
			return wire
		}},
		{"oversized attachment", func() []byte {
			wire := mapgenLoadWire503830("Target", "", nil)
			version := 4 + 4 + 1 + len("Target") + 1
			wire[version] = 2
			attachment := version + 1 + 8
			binary.LittleEndian.PutUint32(wire[attachment:], math.MaxUint32)
			return wire
		}},
		{"truncated section", func() []byte {
			wire := mapgenLoadWire503830("Target", "Known", []byte{1})
			return append(wire[:len(wire)-5], wire[len(wire)-4:]...)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			freed := mapgenLoadSetup503830(t)
			Nox_xxx_mapReadSection_426EA0 = func(_ unsafe.Pointer, _ string) (bool, error) { return true, nil }
			got := mapgenLoadViaC503830(mapgenLoadWrite503830(t, tc.wire()))
			if got.loaded || !got.fileClosed || *freed != 1 {
				t.Fatalf("malformed record = %+v, frees=%d", got, *freed)
			}
		})
	}
}

func TestMapgenLoad503830SectionFailureClosesFile(t *testing.T) {
	freed := mapgenLoadSetup503830(t)
	Nox_xxx_mapReadSection_426EA0 = func(_ unsafe.Pointer, _ string) (bool, error) {
		return false, errors.New("broken section")
	}
	got := mapgenLoadViaC503830(mapgenLoadWrite503830(t, mapgenLoadWire503830("Target", "Broken", nil)))
	if got.loaded || !got.fileClosed || *freed != 1 {
		t.Fatalf("failed section = %+v, frees=%d", got, *freed)
	}
}
