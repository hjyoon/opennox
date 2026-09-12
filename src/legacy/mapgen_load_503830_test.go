package legacy

import (
	"encoding/binary"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

type mapgenLoadServer503830 struct {
	Server
	freed *int
}

func (s mapgenLoadServer503830) Nox_xxx_free503F40() { *s.freed++ }

func mapgenLoadWire503830(name, section string, payload []byte) []byte {
	extra := binary.LittleEndian.AppendUint32(nil, mapgenMagic502ED0)
	for _, value := range []uint32{7, 9, 10, 20, 10, 40, 30, 20, 30, 40} {
		extra = binary.LittleEndian.AppendUint32(extra, value)
	}
	encoded := make([]byte, 0, 1+len(section)+4+len(payload)+1)
	if section != "" {
		encoded = append(encoded, byte(len(section)))
		encoded = append(encoded, section...)
		encoded = binary.LittleEndian.AppendUint32(encoded, uint32(len(payload)))
		encoded = append(encoded, payload...)
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
