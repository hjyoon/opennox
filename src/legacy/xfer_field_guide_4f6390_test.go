package legacy

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

type fieldGuideXferTestData4F6390 struct {
	buf []byte
}

type fieldGuideXferTestObject4F6390 struct {
	data    *fieldGuideXferTestData4F6390
	field34 uint32
}

type fieldGuideXferTestInventoryCall4F6390 struct {
	version uint16
	object  *fieldGuideXferTestObject4F6390
	count   int32
}

type fieldGuideXferTestWorld4F6390 struct {
	version         uint16
	mapResult       int32
	byteResult      *uint8
	readPayload     []byte
	modes           []int32
	inventoryResult int32

	field34Loads   int
	useDataLoads   int
	mapVersions    []int32
	modeCalls      int
	byteInputs     []uint8
	creatureCounts []uint8
	creatureBytes  [][]byte
	terminators    []uint8
	inventoryCalls []fieldGuideXferTestInventoryCall4F6390
	field34Stores  int
	events         []string
	after          map[string]func()
}

func newFieldGuideXferTestWorld4F6390() *fieldGuideXferTestWorld4F6390 {
	return &fieldGuideXferTestWorld4F6390{
		version:         fieldGuideXferCurrentVersion4F6390,
		mapResult:       1,
		inventoryResult: 1,
		after:           make(map[string]func()),
	}
}

func fieldGuideXferTestByte4F6390(value uint8) *uint8 {
	return &value
}

func fieldGuideXferTestDataWithName4F6390(name string, trailing ...byte) *fieldGuideXferTestData4F6390 {
	buf := make([]byte, len(name)+1+len(trailing))
	copy(buf, name)
	copy(buf[len(name)+1:], trailing)
	return &fieldGuideXferTestData4F6390{buf: buf}
}

func (w *fieldGuideXferTestWorld4F6390) event(name string) {
	w.events = append(w.events, name)
	if callback := w.after[name]; callback != nil {
		callback()
	}
}

func (w *fieldGuideXferTestWorld4F6390) deps() fieldGuideXferDeps4F6390[
	*fieldGuideXferTestObject4F6390,
	*fieldGuideXferTestData4F6390,
] {
	return fieldGuideXferDeps4F6390[
		*fieldGuideXferTestObject4F6390,
		*fieldGuideXferTestData4F6390,
	]{
		loadUseData: func(object *fieldGuideXferTestObject4F6390) *fieldGuideXferTestData4F6390 {
			w.useDataLoads++
			w.event("load-use-data")
			return object.data
		},
		loadField34: func(object *fieldGuideXferTestObject4F6390) uint32 {
			w.field34Loads++
			w.event(fmt.Sprintf("load-field34:%d", w.field34Loads))
			return object.field34
		},
		rwVersion: func(value uint16) uint16 {
			w.event(fmt.Sprintf("rw-version:%d", value))
			return w.version
		},
		mapReadWrite: func(_ *fieldGuideXferTestObject4F6390, version int32) int32 {
			w.mapVersions = append(w.mapVersions, version)
			w.event(fmt.Sprintf("map-read-write:%d", version))
			return w.mapResult
		},
		readMode: func() int32 {
			call := w.modeCalls
			w.modeCalls++
			value := int32(0)
			if call < len(w.modes) {
				value = w.modes[call]
			}
			w.event(fmt.Sprintf("read-mode:%d=%d", call+1, value))
			return value
		},
		creatureLength: func(data *fieldGuideXferTestData4F6390) uint32 {
			w.event("creature-length")
			for i, value := range data.buf {
				if value == 0 {
					return uint32(i)
				}
			}
			return uint32(len(data.buf))
		},
		rwByte: func(value uint8) uint8 {
			w.byteInputs = append(w.byteInputs, value)
			w.event(fmt.Sprintf("rw-byte:%d", value))
			if w.byteResult != nil {
				return *w.byteResult
			}
			return value
		},
		rwCreature: func(data *fieldGuideXferTestData4F6390, size uint8) {
			w.creatureCounts = append(w.creatureCounts, size)
			if size != 0 {
				if len(data.buf) < int(size) {
					panic("FieldGuideXfer test use-data underrun")
				}
				if w.readPayload != nil {
					if len(w.readPayload) < int(size) {
						panic("FieldGuideXfer test byte-stream underrun")
					}
					copy(data.buf[:size], w.readPayload[:size])
				}
				w.creatureBytes = append(w.creatureBytes, append([]byte(nil), data.buf[:size]...))
			} else {
				w.creatureBytes = append(w.creatureBytes, []byte{})
			}
			w.event(fmt.Sprintf("rw-creature:%d", size))
		},
		storeCreatureTerminator: func(data *fieldGuideXferTestData4F6390, index uint8) {
			w.terminators = append(w.terminators, index)
			w.event(fmt.Sprintf("store-terminator:%d", index))
			data.buf[index] = 0
		},
		transferInventory: func(version uint16, object *fieldGuideXferTestObject4F6390, count int32) int32 {
			w.inventoryCalls = append(w.inventoryCalls, fieldGuideXferTestInventoryCall4F6390{
				version: version,
				object:  object,
				count:   count,
			})
			w.event("transfer-inventory")
			return w.inventoryResult
		},
		storeField34: func(object *fieldGuideXferTestObject4F6390, value uint32) {
			w.field34Stores++
			object.field34 = value
			w.event("store-field34")
		},
	}
}

func TestFieldGuideXfer4F6390CachesEntryUseDataAndOrdersWrite(t *testing.T) {
	entry := fieldGuideXferTestDataWithName4F6390("old")
	replacement := fieldGuideXferTestDataWithName4F6390("replacement")
	object := &fieldGuideXferTestObject4F6390{data: entry, field34: 0x10203040}
	w := newFieldGuideXferTestWorld4F6390()
	w.mapResult = -7
	w.modes = []int32{0}
	w.after["map-read-write:60"] = func() {
		entry.buf = append([]byte("UrchinShaman"), 0)
		object.data = replacement
		object.field34 = 0
	}

	if got := fieldGuideXfer4F6390(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if object.data != replacement || object.field34 != 0x10203040 {
		t.Fatalf("live data/Field34 = %p/%#x, want replacement/restored", object.data, object.field34)
	}
	if !reflect.DeepEqual(w.creatureBytes, [][]byte{[]byte("UrchinShaman")}) {
		t.Fatalf("write payload = %q, want UrchinShaman", w.creatureBytes)
	}
	wantEvents := []string{
		"load-use-data",
		"load-field34:1",
		"rw-version:60",
		"map-read-write:60",
		"read-mode:1=0",
		"creature-length",
		"rw-byte:12",
		"rw-creature:12",
		"load-field34:2",
		"store-field34",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, wantEvents)
	}
}

func TestFieldGuideXfer4F6390SignedVersionAndIndependentModes(t *testing.T) {
	data := &fieldGuideXferTestData4F6390{buf: bytes.Repeat([]byte{0xcc}, 64)}
	object := &fieldGuideXferTestObject4F6390{data: data, field34: 0x11223344}
	w := newFieldGuideXferTestWorld4F6390()
	w.version = 0xffff
	w.mapResult = -9
	w.byteResult = fieldGuideXferTestByte4F6390(3)
	w.readPayload = []byte("Imp")
	w.modes = []int32{2, 1}
	w.inventoryResult = -3
	w.after["map-read-write:-1"] = func() { object.field34 = 0x80000002 }

	if got := fieldGuideXfer4F6390(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if !reflect.DeepEqual(w.mapVersions, []int32{-1}) {
		t.Fatalf("map versions = %v, want signed -1", w.mapVersions)
	}
	wantInventory := []fieldGuideXferTestInventoryCall4F6390{{
		version: 0xffff,
		object:  object,
		count:   -2147483646,
	}}
	if !reflect.DeepEqual(w.inventoryCalls, wantInventory) {
		t.Fatalf("inventory calls = %#v, want zero-extended version and signed count", w.inventoryCalls)
	}
	if string(data.buf[:3]) != "Imp" || data.buf[3] != 0 || object.field34 != 0x11223344 {
		t.Fatalf("creature/NUL/Field34 = %q/%#x/%#x, want Imp/0/restored",
			data.buf[:3], data.buf[3], object.field34)
	}
	if !reflect.DeepEqual(w.byteInputs, []uint8{0}) || !reflect.DeepEqual(w.terminators, []uint8{3}) {
		t.Fatalf("byte inputs/terminators = %v/%v, want [0]/[3]", w.byteInputs, w.terminators)
	}
}

func TestFieldGuideXfer4F6390WriteLengthUsesLowByteWithout64ByteGate(t *testing.T) {
	for _, length := range []int{0, 1, 63, 64, 255, 256, 257} {
		t.Run(fmt.Sprintf("length-%d", length), func(t *testing.T) {
			data := &fieldGuideXferTestData4F6390{buf: append(bytes.Repeat([]byte{'X'}, length), 0)}
			object := &fieldGuideXferTestObject4F6390{data: data}
			w := newFieldGuideXferTestWorld4F6390()
			w.modes = []int32{0}

			if got := fieldGuideXfer4F6390(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			wantSize := uint8(length)
			if !reflect.DeepEqual(w.byteInputs, []uint8{wantSize}) ||
				!reflect.DeepEqual(w.creatureCounts, []uint8{wantSize}) || len(w.creatureBytes) != 1 ||
				len(w.creatureBytes[0]) != int(wantSize) {
				t.Fatalf("length byte/count/payload = %v/%v/%d, want %#x/%#x/%d",
					w.byteInputs, w.creatureCounts, len(w.creatureBytes[0]), wantSize, wantSize, wantSize)
			}
			if w.modeCalls != 1 || len(w.terminators) != 0 {
				t.Fatalf("mode calls/terminators = %d/%v, want 1/none", w.modeCalls, w.terminators)
			}
		})
	}
}

func TestFieldGuideXfer4F6390WriteUsesTransferredLengthByte(t *testing.T) {
	data := &fieldGuideXferTestData4F6390{buf: []byte{'A', 0, 'B', 'C'}}
	object := &fieldGuideXferTestObject4F6390{data: data}
	w := newFieldGuideXferTestWorld4F6390()
	w.modes = []int32{0}
	w.byteResult = fieldGuideXferTestByte4F6390(3)

	if got := fieldGuideXfer4F6390(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.byteInputs, []uint8{1}) ||
		!reflect.DeepEqual(w.creatureBytes, [][]byte{{'A', 0, 'B'}}) {
		t.Fatalf("length input/payload = %v/%v, want [1]/[A NUL B]", w.byteInputs, w.creatureBytes)
	}
}

func TestFieldGuideXfer4F6390ReadLengthBoundaryAndTerminator(t *testing.T) {
	for _, size := range []uint8{0, 1, 63, 64, 255} {
		t.Run(fmt.Sprintf("size-%d", size), func(t *testing.T) {
			data := &fieldGuideXferTestData4F6390{buf: bytes.Repeat([]byte{0xcc}, 65)}
			object := &fieldGuideXferTestObject4F6390{data: data, field34: 7}
			w := newFieldGuideXferTestWorld4F6390()
			w.byteResult = fieldGuideXferTestByte4F6390(size)
			w.readPayload = bytes.Repeat([]byte{'Q'}, int(size))
			w.modes = []int32{2, 0}
			w.after["map-read-write:60"] = func() { object.field34 = 9 }

			got := fieldGuideXfer4F6390(object, w.deps())
			if size >= 64 {
				if got != 0 || object.field34 != 9 || w.field34Loads != 1 || w.field34Stores != 0 ||
					len(w.creatureCounts) != 0 || len(w.terminators) != 0 || w.modeCalls != 1 {
					t.Fatalf("reject result/Field34/loads/stores/payload/NUL/modes = %d/%d/%d/%d/%v/%v/%d",
						got, object.field34, w.field34Loads, w.field34Stores,
						w.creatureCounts, w.terminators, w.modeCalls)
				}
				return
			}

			if got != 1 || object.field34 != 7 || w.field34Loads != 2 || w.field34Stores != 1 {
				t.Fatalf("success result/Field34/loads/stores = %d/%d/%d/%d, want 1/7/2/1",
					got, object.field34, w.field34Loads, w.field34Stores)
			}
			if !reflect.DeepEqual(w.creatureCounts, []uint8{size}) ||
				!reflect.DeepEqual(w.terminators, []uint8{size}) || data.buf[size] != 0 {
				t.Fatalf("payload/NUL/data = %v/%v/%#x, want [%d]/[%d]/0",
					w.creatureCounts, w.terminators, data.buf[size], size, size)
			}
			if size != 0 && !bytes.Equal(data.buf[:size], bytes.Repeat([]byte{'Q'}, int(size))) {
				t.Fatalf("read payload = %q, want Q bytes", data.buf[:size])
			}
		})
	}
}

func TestFieldGuideXfer4F6390SuffixGatesAndInventoryFailure(t *testing.T) {
	t.Run("live zero skips second mode read", func(t *testing.T) {
		object := &fieldGuideXferTestObject4F6390{data: fieldGuideXferTestDataWithName4F6390("A"), field34: 3}
		w := newFieldGuideXferTestWorld4F6390()
		w.modes = []int32{0, 1}
		w.after["rw-creature:1"] = func() { object.field34 = 0 }

		if got := fieldGuideXfer4F6390(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 1 || len(w.inventoryCalls) != 0 || object.field34 != 3 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 1/0/3",
				w.modeCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("suffix mode must equal one", func(t *testing.T) {
		object := &fieldGuideXferTestObject4F6390{data: fieldGuideXferTestDataWithName4F6390("A"), field34: 3}
		w := newFieldGuideXferTestWorld4F6390()
		w.modes = []int32{0, 2}

		if got := fieldGuideXfer4F6390(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.modeCalls != 2 || len(w.inventoryCalls) != 0 || object.field34 != 3 {
			t.Fatalf("mode/inventory/Field34 = %d/%d/%d, want 2/0/3",
				w.modeCalls, len(w.inventoryCalls), object.field34)
		}
	})

	t.Run("inventory failure keeps live count", func(t *testing.T) {
		object := &fieldGuideXferTestObject4F6390{data: fieldGuideXferTestDataWithName4F6390("A"), field34: 7}
		w := newFieldGuideXferTestWorld4F6390()
		w.modes = []int32{0, 1}
		w.inventoryResult = 0
		w.after["rw-creature:1"] = func() { object.field34 = 0x80000004 }

		if got := fieldGuideXfer4F6390(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []fieldGuideXferTestInventoryCall4F6390{{
			version: fieldGuideXferCurrentVersion4F6390,
			object:  object,
			count:   -2147483644,
		}}
		if !reflect.DeepEqual(w.inventoryCalls, want) || object.field34 != 0x80000004 || w.field34Stores != 0 {
			t.Fatalf("inventory/Field34/stores = %#v/%#x/%d, want live call/live/0",
				w.inventoryCalls, object.field34, w.field34Stores)
		}
	})
}

func TestFieldGuideXfer4F6390FailurePrefixes(t *testing.T) {
	for _, version := range []uint16{61, 0x7fff} {
		t.Run(fmt.Sprintf("reject-%#x", version), func(t *testing.T) {
			object := &fieldGuideXferTestObject4F6390{field34: 7}
			w := newFieldGuideXferTestWorld4F6390()
			w.version = version

			if got := fieldGuideXfer4F6390(object, w.deps()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			wantEvents := []string{"load-use-data", "load-field34:1", "rw-version:60"}
			if !reflect.DeepEqual(w.events, wantEvents) || len(w.mapVersions) != 0 || w.field34Stores != 0 {
				t.Fatalf("events/map/store = %v/%v/%d, want entry prefix/none/0",
					w.events, w.mapVersions, w.field34Stores)
			}
		})
	}

	t.Run("common failure keeps callback mutation", func(t *testing.T) {
		object := &fieldGuideXferTestObject4F6390{
			data: fieldGuideXferTestDataWithName4F6390("A"), field34: 0x11223344,
		}
		w := newFieldGuideXferTestWorld4F6390()
		w.mapResult = 0
		w.after["map-read-write:60"] = func() { object.field34 = 0x55667788 }

		if got := fieldGuideXfer4F6390(object, w.deps()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		if object.field34 != 0x55667788 || w.modeCalls != 0 || len(w.creatureCounts) != 0 ||
			w.field34Loads != 1 || w.field34Stores != 0 {
			t.Fatalf("Field34/modes/payload/loads/stores = %#x/%d/%v/%d/%d, want callback state/0/none/1/0",
				object.field34, w.modeCalls, w.creatureCounts, w.field34Loads, w.field34Stores)
		}
	})
}

func TestFieldGuideXfer4F6390FaultBoundaries(t *testing.T) {
	expectPanic := func(t *testing.T, run func()) {
		t.Helper()
		deferred := false
		func() {
			defer func() { deferred = recover() != nil }()
			run()
		}()
		if !deferred {
			t.Fatal("call did not panic")
		}
	}

	t.Run("nil object faults on entry UseData load", func(t *testing.T) {
		w := newFieldGuideXferTestWorld4F6390()
		expectPanic(t, func() { fieldGuideXfer4F6390((*fieldGuideXferTestObject4F6390)(nil), w.deps()) })
		if !reflect.DeepEqual(w.events, []string{"load-use-data"}) || w.field34Loads != 0 {
			t.Fatalf("events/Field34 loads = %v/%d, want attempted UseData only", w.events, w.field34Loads)
		}
	})

	t.Run("nil UseData write faults at strlen after common", func(t *testing.T) {
		object := &fieldGuideXferTestObject4F6390{field34: 3}
		w := newFieldGuideXferTestWorld4F6390()
		w.modes = []int32{0}
		expectPanic(t, func() { fieldGuideXfer4F6390(object, w.deps()) })
		wantSuffix := []string{"map-read-write:60", "read-mode:1=0", "creature-length"}
		if len(w.events) < len(wantSuffix) || !reflect.DeepEqual(w.events[len(w.events)-len(wantSuffix):], wantSuffix) ||
			w.field34Loads != 1 || w.field34Stores != 0 {
			t.Fatalf("events/loads/stores = %v/%d/%d, want common->mode->strlen fault/1/0",
				w.events, w.field34Loads, w.field34Stores)
		}
	})

	t.Run("nil UseData empty read faults at trailing NUL store", func(t *testing.T) {
		object := &fieldGuideXferTestObject4F6390{field34: 3}
		w := newFieldGuideXferTestWorld4F6390()
		w.byteResult = fieldGuideXferTestByte4F6390(0)
		w.modes = []int32{2}
		expectPanic(t, func() { fieldGuideXfer4F6390(object, w.deps()) })
		wantSuffix := []string{"read-mode:1=2", "rw-byte:0", "rw-creature:0", "store-terminator:0"}
		if len(w.events) < len(wantSuffix) || !reflect.DeepEqual(w.events[len(w.events)-len(wantSuffix):], wantSuffix) ||
			w.field34Loads != 1 || w.field34Stores != 0 {
			t.Fatalf("events/loads/stores = %v/%d/%d, want empty transfer then NUL fault/1/0",
				w.events, w.field34Loads, w.field34Stores)
		}
	})
}
