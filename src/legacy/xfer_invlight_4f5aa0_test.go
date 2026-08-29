package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"testing"
)

type invLightXferTestObject4F5AA0 struct {
	field34 uint32
	class   uint32
	extent  uint32
	netCode uint32
}

type invLightXferTestDrawable4F5AA0 struct {
	name  string
	code  uint32
	next  *invLightXferTestDrawable4F5AA0
	light [invLightXferPayloadSize4F5AA0]byte
}

type invLightXferTestWorld4F5AA0 struct {
	t               *testing.T
	events          []string
	version         uint16
	mapResult       int32
	modes           []int32
	flags           []int32
	first           *invLightXferTestDrawable4F5AA0
	static          *invLightXferTestDrawable4F5AA0
	dynamic         *invLightXferTestDrawable4F5AA0
	readWire        bool
	wire            []byte
	wireAt          int
	written         []byte
	legacyField43   uint8
	forceTooBright  *bool
	clamps          int
	applied         *[invLightXferPayloadSize4F5AA0]byte
	inventoryCalls  int
	inventoryVer    uint16
	inventoryCount  int32
	inventoryResult int32
}

func newInvLightXferTestWorld4F5AA0(t *testing.T) *invLightXferTestWorld4F5AA0 {
	return &invLightXferTestWorld4F5AA0{
		t:               t,
		version:         invLightXferCurrentVersion4F5AA0,
		mapResult:       1,
		inventoryResult: 1,
	}
}

func (w *invLightXferTestWorld4F5AA0) event(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

func (w *invLightXferTestWorld4F5AA0) pop(values *[]int32, name string) int32 {
	w.t.Helper()
	if len(*values) == 0 {
		w.t.Fatalf("unexpected %s read after events %v", name, w.events)
	}
	value := (*values)[0]
	*values = (*values)[1:]
	return value
}

func (w *invLightXferTestWorld4F5AA0) deps() invLightXferDeps4F5AA0[
	*invLightXferTestObject4F5AA0,
	*invLightXferTestDrawable4F5AA0,
] {
	return invLightXferDeps4F5AA0[
		*invLightXferTestObject4F5AA0,
		*invLightXferTestDrawable4F5AA0,
	]{
		loadField34: func(object *invLightXferTestObject4F5AA0) uint32 {
			w.event("field34:%#x", object.field34)
			return object.field34
		},
		rwVersion: func(value uint16) uint16 {
			w.event("version:%d", value)
			return w.version
		},
		mapReadWrite: func(_ *invLightXferTestObject4F5AA0, version int32) int32 {
			w.event("map:%d", version)
			return w.mapResult
		},
		readMode: func() int32 {
			value := w.pop(&w.modes, "mode")
			w.event("mode:%d", value)
			return value
		},
		gameFlags: func(mask uint32) int32 {
			value := w.pop(&w.flags, "game flags")
			w.event("flags:%#x:%d", mask, value)
			return value
		},
		firstDrawable: func() *invLightXferTestDrawable4F5AA0 {
			w.event("first")
			return w.first
		},
		nextDrawable: func(drawable *invLightXferTestDrawable4F5AA0) *invLightXferTestDrawable4F5AA0 {
			w.event("next:%s", drawable.name)
			return drawable.next
		},
		loadDrawableCode: func(drawable *invLightXferTestDrawable4F5AA0) uint32 {
			w.event("drawable-code:%s:%#x", drawable.name, drawable.code)
			return drawable.code
		},
		loadExtent: func(object *invLightXferTestObject4F5AA0) uint32 {
			w.event("extent:%#x", object.extent)
			return object.extent
		},
		loadClass: func(object *invLightXferTestObject4F5AA0) uint32 {
			w.event("class:%#x", object.class)
			return object.class
		},
		loadNetCode: func(object *invLightXferTestObject4F5AA0) uint32 {
			w.event("net-code:%#x", object.netCode)
			return object.netCode
		},
		staticDrawable: func(code uint32) *invLightXferTestDrawable4F5AA0 {
			w.event("static:%#x", code)
			return w.static
		},
		dynamicDrawable: func(code uint32) *invLightXferTestDrawable4F5AA0 {
			w.event("dynamic:%#x", code)
			return w.dynamic
		},
		copyDrawableLight: func(drawable *invLightXferTestDrawable4F5AA0, light *[invLightXferPayloadSize4F5AA0]byte) {
			w.event("copy-drawable:%s", drawable.name)
			*light = drawable.light
		},
		rwLight: func(light *[invLightXferPayloadSize4F5AA0]byte, offset, size int) {
			w.event("rw:%d:%d", offset, size)
			if w.readWire {
				if w.wireAt+size > len(w.wire) {
					w.t.Fatalf("wire exhausted at %d+%d of %d", w.wireAt, size, len(w.wire))
				}
				copy(light[offset:offset+size], w.wire[w.wireAt:w.wireAt+size])
				w.wireAt += size
				return
			}
			w.written = append(w.written, light[offset:offset+size]...)
		},
		rwLegacyField43: func(value uint8) uint8 {
			w.event("rw-legacy-43:%#x", value)
			return w.legacyField43
		},
		legacyTooBright: func(light *[invLightXferPayloadSize4F5AA0]byte) bool {
			value := invLightLegacyTooBright4F5AA0(light)
			if w.forceTooBright != nil {
				value = *w.forceTooBright
			}
			w.event("too-bright:%t", value)
			return value
		},
		clampLegacyLight: func(light *[invLightXferPayloadSize4F5AA0]byte) {
			w.event("clamp")
			w.clamps++
			invLightClampLegacy4F5AA0(light, func(float32) uint32 { return 0xa1b2c3d4 })
		},
		copyObjectLight: func(_ *invLightXferTestObject4F5AA0, light *[invLightXferPayloadSize4F5AA0]byte) {
			w.event("copy-object")
			copyValue := *light
			w.applied = &copyValue
		},
		transferInventory: func(version uint16, _ *invLightXferTestObject4F5AA0, count int32) int32 {
			w.event("inventory:%#x:%d", version, count)
			w.inventoryCalls++
			w.inventoryVer = version
			w.inventoryCount = count
			return w.inventoryResult
		},
		storeField34: func(object *invLightXferTestObject4F5AA0, value uint32) {
			w.event("store-field34:%#x", value)
			object.field34 = value
		},
	}
}

func invLightXferExpectedWire4F5AA0(
	light [invLightXferPayloadSize4F5AA0]byte,
	version int16,
) []byte {
	var out []byte
	parts := invLightXferParts4F5AA0[:]
	if version < 2 {
		parts = invLightXferLegacyParts4F5AA0[:]
	}
	for _, part := range parts {
		out = append(out, light[part.offset:part.offset+part.size]...)
	}
	if version >= 42 {
		out = append(out, light[36:40]...)
	}
	return out
}

func TestInvLightXfer4F5AA0Version60WriteUsesExactStaticWire(t *testing.T) {
	object := &invLightXferTestObject4F5AA0{class: invLightXferStaticMask4F5AA0, extent: 0x11223344}
	drawable := &invLightXferTestDrawable4F5AA0{name: "static"}
	for i := range drawable.light {
		drawable.light[i] = byte(i)
	}
	w := newInvLightXferTestWorld4F5AA0(t)
	w.modes = []int32{0, 0}
	w.flags = []int32{0}
	w.static = drawable

	if got := invLightXfer4F5AA0(object, w.deps()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if want := invLightXferExpectedWire4F5AA0(drawable.light, 60); !bytes.Equal(w.written, want) {
		t.Fatalf("wire = %x, want %x", w.written, want)
	}
	wantPrefix := []string{
		"field34:0x0", "version:60", "map:60", "mode:0", "flags:0x600000:0",
		"class:0x20400000", "extent:0x11223344", "static:0x11223344", "copy-drawable:static",
	}
	if !reflect.DeepEqual(w.events[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("event prefix = %v, want %v", w.events[:len(wantPrefix)], wantPrefix)
	}
	if got := w.events[len(w.events)-3:]; !reflect.DeepEqual(got, []string{"mode:0", "field34:0x0", "store-field34:0x0"}) {
		t.Fatalf("event suffix = %v", got)
	}
}

func TestInvLightXfer4F5AA0PreviewListReloadsExtent(t *testing.T) {
	object := &invLightXferTestObject4F5AA0{extent: 10}
	second := &invLightXferTestDrawable4F5AA0{name: "second", code: 20}
	first := &invLightXferTestDrawable4F5AA0{name: "first", code: 1, next: second}
	w := newInvLightXferTestWorld4F5AA0(t)
	w.modes = []int32{0, 0}
	w.flags = []int32{1}
	w.first = first
	deps := w.deps()
	extentLoads := 0
	deps.loadExtent = func(object *invLightXferTestObject4F5AA0) uint32 {
		extentLoads++
		w.event("extent-load:%d", extentLoads)
		if extentLoads == 1 {
			object.extent = 20
			return 10
		}
		return object.extent
	}

	if got := invLightXfer4F5AA0(object, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if extentLoads != 2 || !reflect.DeepEqual(w.events[4:12], []string{
		"flags:0x600000:1", "first", "drawable-code:first:0x1", "extent-load:1",
		"next:first", "drawable-code:second:0x14", "extent-load:2", "copy-drawable:second",
	}) {
		t.Fatalf("preview events = %v, extent loads %d", w.events, extentLoads)
	}
}

func TestInvLightXfer4F5AA0VersionBranches(t *testing.T) {
	tests := []struct {
		name          string
		version       uint16
		wireSize      int
		legacyField43 uint8
		wantField43   uint32
		wantLastEvent string
	}{
		{name: "version 42", version: 42, wireSize: 137, wantField43: 0x88878685, wantLastEvent: "rw:36:4"},
		{name: "version 41", version: 41, wireSize: 133, legacyField43: 0x7e, wantField43: 0x7e, wantLastEvent: "rw-legacy-43:0x0"},
		{name: "version 40", version: 40, wireSize: 133, wantField43: 0, wantLastEvent: "rw:138:1"},
		{name: "version 2", version: 2, wireSize: 133, wantField43: 0, wantLastEvent: "rw:138:1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := &invLightXferTestObject4F5AA0{}
			w := newInvLightXferTestWorld4F5AA0(t)
			w.version = test.version
			w.modes = []int32{1, 1}
			w.flags = []int32{1}
			w.readWire = true
			w.wire = make([]byte, test.wireSize)
			for i := range w.wire {
				w.wire[i] = byte(i)
			}
			w.legacyField43 = test.legacyField43

			if got := invLightXfer4F5AA0(object, w.deps()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if w.applied == nil {
				t.Fatal("read payload was not applied")
			}
			if got := binary.LittleEndian.Uint32(w.applied[36:40]); got != test.wantField43 {
				t.Fatalf("field 43 = %#x, want %#x", got, test.wantField43)
			}
			if w.wireAt != len(w.wire) {
				t.Fatalf("wire consumed = %d, want %d", w.wireAt, len(w.wire))
			}
			modeAt := -1
			for i, event := range w.events {
				if event == "mode:1" && i != 3 {
					modeAt = i
					break
				}
			}
			if modeAt <= 0 || w.events[modeAt-1] != test.wantLastEvent {
				t.Fatalf("events before apply-mode = %v, want %q", w.events, test.wantLastEvent)
			}
		})
	}
}

func TestInvLightXfer4F5AA0LegacyDefaultsAndClampReload(t *testing.T) {
	t.Run("signed negative version keeps exact defaults and untouched bytes", func(t *testing.T) {
		object := &invLightXferTestObject4F5AA0{netCode: 17}
		drawable := &invLightXferTestDrawable4F5AA0{name: "dynamic"}
		for i := range drawable.light {
			drawable.light[i] = 0xaa
		}
		w := newInvLightXferTestWorld4F5AA0(t)
		w.version = 0xffff
		w.modes = []int32{0, 1}
		w.flags = []int32{0, 1}
		w.dynamic = drawable
		w.readWire = true
		w.wire = make([]byte, 36)
		binary.LittleEndian.PutUint32(w.wire[4:8], math.Float32bits(1))

		if got := invLightXfer4F5AA0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.applied == nil {
			t.Fatal("legacy payload was not applied")
		}
		light := w.applied[:]
		for _, span := range [][2]int{{40, 42}, {122, 132}, {134, 136}} {
			if !bytes.Equal(light[span[0]:span[1]], make([]byte, span[1]-span[0])) {
				t.Errorf("cleared span %v = %x", span, light[span[0]:span[1]])
			}
		}
		if light[138] != 0x80 {
			t.Errorf("legacy byte 138 = %#x, want 0x80", light[138])
		}
		for _, span := range [][2]int{{36, 40}, {42, 122}, {132, 134}, {136, 138}, {139, 140}} {
			for index, value := range light[span[0]:span[1]] {
				if value != 0xaa {
					t.Errorf("untouched byte %d = %#x, want 0xaa", span[0]+index, value)
				}
			}
		}
	})

	t.Run("clamp reload can revoke apply", func(t *testing.T) {
		object := &invLightXferTestObject4F5AA0{}
		w := newInvLightXferTestWorld4F5AA0(t)
		w.version = 1
		w.modes = []int32{1, 1, 0}
		w.readWire = true
		w.wire = make([]byte, 36)
		binary.LittleEndian.PutUint32(w.wire[4:8], math.Float32bits(64))

		if got := invLightXfer4F5AA0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.clamps != 1 || w.applied != nil || len(w.flags) != 0 {
			t.Fatalf("clamps/applied/flags = %d/%v/%v, want 1/nil/empty", w.clamps, w.applied, w.flags)
		}
		want := []string{"too-bright:true", "clamp", "mode:0", "field34:0x0", "store-field34:0x0"}
		if got := w.events[len(w.events)-len(want):]; !reflect.DeepEqual(got, want) {
			t.Fatalf("event suffix = %v, want %v", got, want)
		}
	})
}

func TestInvLightLegacyThresholdAndClamp4F5AA0(t *testing.T) {
	var light [invLightXferPayloadSize4F5AA0]byte
	set := func(intensity float32, fixed int32) {
		clear(light[:])
		binary.LittleEndian.PutUint32(light[4:8], math.Float32bits(intensity))
		binary.LittleEndian.PutUint32(light[12:16], uint32(fixed))
	}
	set(63, 63*65536)
	if invLightLegacyTooBright4F5AA0(&light) {
		t.Fatal("exact legacy limit was treated as too bright")
	}
	set(math.Float32frombits(math.Float32bits(63)+1), 0)
	if !invLightLegacyTooBright4F5AA0(&light) {
		t.Fatal("float intensity above 63 was accepted")
	}
	set(float32(math.NaN()), 63*65536+1)
	if !invLightLegacyTooBright4F5AA0(&light) {
		t.Fatal("fixed intensity above 63 was accepted when float was NaN")
	}
	set(float32(math.NaN()), -1)
	if invLightLegacyTooBright4F5AA0(&light) {
		t.Fatal("unordered float and negative fixed intensity were clamped")
	}

	for i := range light {
		light[i] = byte(i)
	}
	fixedBefore := append([]byte(nil), light[12:16]...)
	invLightClampLegacy4F5AA0(&light, func(value float32) uint32 {
		if value != 63 {
			t.Fatalf("radius intensity = %v, want 63", value)
		}
		return 0x11223344
	})
	if got := math.Float32frombits(binary.LittleEndian.Uint32(light[4:8])); got != 63 {
		t.Errorf("clamped intensity = %v, want 63", got)
	}
	if got := binary.LittleEndian.Uint32(light[8:12]); got != 0x11223344 {
		t.Errorf("clamped radius = %#x", got)
	}
	if !bytes.Equal(light[12:16], fixedBefore) {
		t.Errorf("fixed intensity changed = %x, want %x", light[12:16], fixedBefore)
	}
}

func TestInvLightXfer4F5AA0FailurePrefixesAndField34Restore(t *testing.T) {
	t.Run("newer version does not restore", func(t *testing.T) {
		object := &invLightXferTestObject4F5AA0{field34: 0x11}
		w := newInvLightXferTestWorld4F5AA0(t)
		deps := w.deps()
		deps.rwVersion = func(uint16) uint16 {
			object.field34 = 0x22
			return 61
		}
		if got := invLightXfer4F5AA0(object, deps); got != 0 || object.field34 != 0x22 {
			t.Fatalf("result/Field34 = %d/%#x, want 0/0x22", got, object.field34)
		}
	})

	t.Run("common transfer failure does not restore", func(t *testing.T) {
		object := &invLightXferTestObject4F5AA0{field34: 0x11}
		w := newInvLightXferTestWorld4F5AA0(t)
		deps := w.deps()
		deps.mapReadWrite = func(*invLightXferTestObject4F5AA0, int32) int32 {
			object.field34 = 0x33
			return 0
		}
		if got := invLightXfer4F5AA0(object, deps); got != 0 || object.field34 != 0x33 {
			t.Fatalf("result/Field34 = %d/%#x, want 0/0x33", got, object.field34)
		}
	})

	for _, inventoryResult := range []int32{0, 1} {
		name := "inventory failure"
		if inventoryResult != 0 {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			object := &invLightXferTestObject4F5AA0{field34: 0x11}
			w := newInvLightXferTestWorld4F5AA0(t)
			w.modes = []int32{1, 1, 1}
			w.flags = []int32{0}
			w.inventoryResult = inventoryResult
			deps := w.deps()
			deps.mapReadWrite = func(*invLightXferTestObject4F5AA0, int32) int32 {
				object.field34 = 0x80000002
				return 1
			}
			baseInventory := deps.transferInventory
			deps.transferInventory = func(version uint16, objectArg *invLightXferTestObject4F5AA0, count int32) int32 {
				result := baseInventory(version, objectArg, count)
				object.field34 = 0x44
				return result
			}

			got := invLightXfer4F5AA0(object, deps)
			if w.inventoryVer != 60 || w.inventoryCount != -0x7ffffffe || w.inventoryCalls != 1 {
				t.Fatalf("inventory = version %#x, count %d, calls %d", w.inventoryVer, w.inventoryCount, w.inventoryCalls)
			}
			if inventoryResult == 0 {
				if got != 0 || object.field34 != 0x44 {
					t.Fatalf("failure result/Field34 = %d/%#x, want 0/0x44", got, object.field34)
				}
			} else if got != 1 || object.field34 != 0x11 {
				t.Fatalf("success result/Field34 = %d/%#x, want 1/0x11", got, object.field34)
			}
		})
	}
}

func TestInvLightXfer4F5AA0ExactModeAndMissingDrawableContracts(t *testing.T) {
	t.Run("nonzero initial mode skips drawable but exact one later applies", func(t *testing.T) {
		object := &invLightXferTestObject4F5AA0{field34: 3}
		w := newInvLightXferTestWorld4F5AA0(t)
		w.modes = []int32{2, 1, 2}
		w.flags = []int32{1}
		if got := invLightXfer4F5AA0(object, w.deps()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.applied == nil || w.inventoryCalls != 0 || object.field34 != 3 {
			t.Fatalf("applied/inventory/Field34 = %v/%d/%d", w.applied != nil, w.inventoryCalls, object.field34)
		}
		for _, event := range w.events {
			if event == "first" || len(event) >= 7 && event[:7] == "static:" || len(event) >= 8 && event[:8] == "dynamic:" {
				t.Fatalf("initial mode 2 resolved a drawable: %v", w.events)
			}
		}
	})

	t.Run("static miss is still passed to the copy boundary", func(t *testing.T) {
		object := &invLightXferTestObject4F5AA0{class: invLightXferStaticMask4F5AA0}
		w := newInvLightXferTestWorld4F5AA0(t)
		w.modes = []int32{0, 0}
		w.flags = []int32{0}
		deps := w.deps()
		called := false
		deps.copyDrawableLight = func(drawable *invLightXferTestDrawable4F5AA0, _ *[invLightXferPayloadSize4F5AA0]byte) {
			called = true
			if drawable != nil {
				t.Fatalf("missing static drawable = %v, want nil", drawable)
			}
		}
		if got := invLightXfer4F5AA0(object, deps); got != 1 || !called {
			t.Fatalf("result/copy called = %d/%t, want 1/true", got, called)
		}
	})

	t.Run("preview miss deliberately keeps the zero payload", func(t *testing.T) {
		object := &invLightXferTestObject4F5AA0{}
		w := newInvLightXferTestWorld4F5AA0(t)
		w.modes = []int32{0, 0}
		w.flags = []int32{1}
		deps := w.deps()
		deps.copyDrawableLight = func(*invLightXferTestDrawable4F5AA0, *[invLightXferPayloadSize4F5AA0]byte) {
			t.Fatal("preview miss copied a drawable")
		}
		if got := invLightXfer4F5AA0(object, deps); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if !bytes.Equal(w.written, make([]byte, len(w.written))) {
			t.Fatalf("preview miss wire = %x, want zero", w.written)
		}
	})
}
