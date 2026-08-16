package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type trapDropTestWorld4ED580 struct {
	pointArg      string
	ownerArg      string
	glyphArg      string
	mapResult     int32
	defaultResult int32
	netCodes      map[string]uint32
	events        []string
	faultAt       int
	afterMap      func(*trapDropTestWorld4ED580)
	afterDefault  func(*trapDropTestWorld4ED580)
}

func newTrapDropTestWorld4ED580() *trapDropTestWorld4ED580 {
	return &trapDropTestWorld4ED580{
		pointArg:      "point-a",
		ownerArg:      "owner-a",
		glyphArg:      "glyph-a",
		defaultResult: 1,
		netCodes:      map[string]uint32{"owner-a": 0x89abcdef},
	}
}

func (w *trapDropTestWorld4ED580) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *trapDropTestWorld4ED580) hooks() trapDropHooks4ED580[string, string] {
	return trapDropHooks4ED580[string, string]{
		loadPointArg: func() string {
			w.event("point-arg:" + w.pointArg)
			return w.pointArg
		},
		mapTile: func(point string) int32 {
			w.event("map:" + point)
			if w.afterMap != nil {
				w.afterMap(w)
			}
			return w.mapResult
		},
		loadOwnerArg: func() string {
			w.event("owner-arg:" + w.ownerArg)
			return w.ownerArg
		},
		loadNetCode: func(owner string) uint32 {
			code := w.netCodes[owner]
			w.event(fmt.Sprintf("net-code:%s=%08x", owner, code))
			return code
		},
		audio: func(id uint32, owner string, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", id, owner, kind, code))
		},
		loadGlyphArg: func() string {
			w.event("glyph-arg:" + w.glyphArg)
			return w.glyphArg
		},
		defaultDrop: func(owner, glyph, point string) int32 {
			w.event("default:" + owner + ":" + glyph + ":" + point)
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return w.defaultResult
		},
		setOwner: func(owner, glyph string) {
			w.event("set-owner:" + owner + ":" + glyph)
		},
	}
}

func verifyTrapDropFaultPrefixes4ED580(
	t *testing.T,
	want []string,
	build func() *trapDropTestWorld4ED580,
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			trapDrop4ED580(w.hooks())
		})
	}
}

func TestTrapDrop4ED580ForbiddenTileUsesLiveOwnerAndNetCode(t *testing.T) {
	build := func() *trapDropTestWorld4ED580 {
		w := newTrapDropTestWorld4ED580()
		w.mapResult = -1
		w.afterMap = func(w *trapDropTestWorld4ED580) {
			w.pointArg = "point-b"
			w.ownerArg = "owner-b"
			w.netCodes["owner-b"] = 0xfedcba98
		}
		return w
	}
	want := []string{
		"point-arg:point-a", "map:point-a", "owner-arg:owner-b",
		"net-code:owner-b=fedcba98", "audio:925:owner-b:2:fedcba98",
	}
	w := build()
	if got := trapDrop4ED580(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyTrapDropFaultPrefixes4ED580(t, want, build)
}

func TestTrapDrop4ED580DefaultRejectPreservesCachedPoint(t *testing.T) {
	build := func() *trapDropTestWorld4ED580 {
		w := newTrapDropTestWorld4ED580()
		w.defaultResult = 0
		w.afterMap = func(w *trapDropTestWorld4ED580) {
			w.pointArg = "point-b"
			w.ownerArg = "owner-b"
			w.glyphArg = "glyph-b"
		}
		return w
	}
	want := []string{
		"point-arg:point-a", "map:point-a", "owner-arg:owner-b", "glyph-arg:glyph-b",
		"default:owner-b:glyph-b:point-a",
	}
	w := build()
	if got := trapDrop4ED580(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyTrapDropFaultPrefixes4ED580(t, want, build)
}

func TestTrapDrop4ED580SuccessCachesOwnerAndGlyphAcrossDefault(t *testing.T) {
	build := func() *trapDropTestWorld4ED580 {
		w := newTrapDropTestWorld4ED580()
		w.defaultResult = math.MinInt32
		w.afterDefault = func(w *trapDropTestWorld4ED580) {
			w.ownerArg = "owner-b"
			w.glyphArg = "glyph-b"
		}
		return w
	}
	want := []string{
		"point-arg:point-a", "map:point-a", "owner-arg:owner-a", "glyph-arg:glyph-a",
		"default:owner-a:glyph-a:point-a", "set-owner:owner-a:glyph-a",
	}
	w := build()
	if got := trapDrop4ED580(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyTrapDropFaultPrefixes4ED580(t, want, build)
}

func TestTrapDrop4ED580WholeEAXGates(t *testing.T) {
	for _, mapResult := range []int32{1, -1, math.MinInt32} {
		w := newTrapDropTestWorld4ED580()
		w.mapResult = mapResult
		if got := trapDrop4ED580(w.hooks()); got != 0 {
			t.Fatalf("map result %d: result = %d", mapResult, got)
		}
	}
	for _, defaultResult := range []int32{1, -1, math.MinInt32} {
		w := newTrapDropTestWorld4ED580()
		w.defaultResult = defaultResult
		if got := trapDrop4ED580(w.hooks()); got != 1 {
			t.Fatalf("default result %d: result = %d", defaultResult, got)
		}
	}
}
