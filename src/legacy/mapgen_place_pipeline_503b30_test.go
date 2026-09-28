package legacy

import (
	"runtime"
	"slices"
	"testing"
	"unsafe"
)

type mapgenPlaceTrace503B30 struct {
	calls         []string
	fail          string
	ready         bool
	scriptPayload bool
	geometry      mapgenPlaceGeometry503B30
	seenAt        mapgenPlacePoint503B30
	seenFixed     mapgenPlacePoint503B30
	seenGeometry  mapgenPlaceGeometry503B30
	seenDelta     [2]int32
	seenBounds    [4]int32
}

func (tr *mapgenPlaceTrace503B30) call(name string) bool {
	tr.calls = append(tr.calls, name)
	return tr.fail != name
}

func (tr *mapgenPlaceTrace503B30) deps() mapgenPlaceDeps503B30 {
	return mapgenPlaceDeps503B30{
		fixCoords: func(at mapgenPlacePoint503B30) (mapgenPlacePoint503B30, bool) {
			tr.seenAt = at
			return mapgenPlacePoint503B30{x: 123.5, y: -456.25}, tr.call("fix")
		},
		placementReady: func() bool {
			tr.call("ready")
			return tr.ready
		},
		loadSelected: func() bool { return tr.call("load") },
		geometry: func(at, fixed mapgenPlacePoint503B30) (mapgenPlaceGeometry503B30, bool) {
			tr.seenAt = at
			tr.seenFixed = fixed
			return tr.geometry, tr.call("geometry")
		},
		storeGeometry: func(geometry mapgenPlaceGeometry503B30) {
			tr.call("store-geometry")
			tr.seenGeometry = geometry
		},
		placeTiles: func(delta [2]int32) bool {
			tr.seenDelta = delta
			return tr.call("tiles")
		},
		placeWalls: func(delta [2]int32) bool {
			tr.seenDelta = delta
			return tr.call("walls")
		},
		placeWaypoints: func(delta [2]int32) bool {
			tr.seenDelta = delta
			return tr.call("waypoints")
		},
		placeObjects: func(delta [2]int32) bool {
			tr.seenDelta = delta
			return tr.call("objects")
		},
		prepareWaypoints:     func() { tr.call("prepare-waypoints") },
		markPendingWaypoints: func() { tr.call("mark-waypoints") },
		fixupPendingObjects: func(bounds [4]int32) {
			tr.seenBounds = bounds
			tr.call("fixup-objects")
		},
		placeGroups: func(delta [2]int32) bool {
			tr.seenDelta = delta
			return tr.call("groups")
		},
		clearWaypointTempIDs:   func() { tr.call("clear-waypoint-ids") },
		clearObjectScriptIDs:   func() { tr.call("clear-object-ids") },
		finalizeWaypoints:      func() { tr.call("finalize-waypoints") },
		finalizePendingObjects: func() { tr.call("finalize-objects") },
		markPlaced:             func() { tr.call("mark-placed") },
		hasScriptPayload: func() bool {
			tr.call("has-script")
			return tr.scriptPayload
		},
		finishScriptPayload: func(delta [2]int32) {
			tr.seenDelta = delta
			tr.call("finish-script")
		},
		advancePlacementIndex: func() { tr.call("advance-index") },
	}
}

func mapgenPlaceSuccessCalls503B30(withLoad, withScript bool) []string {
	calls := []string{"fix", "ready"}
	if withLoad {
		calls = append(calls, "load")
	}
	calls = append(calls,
		"geometry", "store-geometry", "tiles", "walls", "waypoints", "objects",
		"prepare-waypoints", "mark-waypoints", "fixup-objects", "groups",
		"clear-waypoint-ids", "clear-object-ids", "finalize-waypoints", "finalize-objects",
		"mark-placed", "has-script",
	)
	if withScript {
		calls = append(calls, "finish-script")
	}
	return append(calls, "advance-index")
}

func TestMapgenPlacePipeline503B30PreservesSuccessOrderAndArguments(t *testing.T) {
	geometry := mapgenPlaceGeometry503B30{
		bounds:   [4]int32{-12, 34, 567, 890},
		delta:    [2]int32{111, -222},
		wallSize: [2]uint32{19, 23},
		wallSpan: [2]float32{437, 529},
	}
	for _, tc := range []struct {
		name          string
		ready         bool
		scriptPayload bool
	}{
		{name: "already loaded", ready: true},
		{name: "load selected map", ready: false},
		{name: "merge script payload", ready: true, scriptPayload: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := &mapgenPlaceTrace503B30{
				ready: tc.ready, scriptPayload: tc.scriptPayload, geometry: geometry,
			}
			at := mapgenPlacePoint503B30{x: 17.25, y: 81.5}
			if !mapgenPlaceWithDeps503B30(at, tr.deps()) {
				t.Fatal("placement failed")
			}
			wantCalls := mapgenPlaceSuccessCalls503B30(!tc.ready, tc.scriptPayload)
			if !slices.Equal(tr.calls, wantCalls) {
				t.Fatalf("calls = %v, want %v", tr.calls, wantCalls)
			}
			if tr.seenAt != at {
				t.Fatalf("input = %+v, want %+v", tr.seenAt, at)
			}
			if want := (mapgenPlacePoint503B30{x: 123.5, y: -456.25}); tr.seenFixed != want {
				t.Fatalf("fixed = %+v, want %+v", tr.seenFixed, want)
			}
			if tr.seenGeometry != geometry || tr.seenDelta != geometry.delta || tr.seenBounds != geometry.bounds {
				t.Fatalf("forwarded geometry = (%+v, %v, %v), want %+v", tr.seenGeometry, tr.seenDelta, tr.seenBounds, geometry)
			}
		})
	}
}

func TestMapgenPlacePipeline503B30PreservesFailurePrefixes(t *testing.T) {
	full := mapgenPlaceSuccessCalls503B30(true, false)
	for _, stage := range []string{"fix", "load", "geometry", "tiles", "walls", "waypoints", "objects", "groups"} {
		t.Run(stage, func(t *testing.T) {
			tr := &mapgenPlaceTrace503B30{
				fail: stage,
				geometry: mapgenPlaceGeometry503B30{
					bounds: [4]int32{1, 2, 3, 4}, delta: [2]int32{5, 6},
				},
			}
			if mapgenPlaceWithDeps503B30(mapgenPlacePoint503B30{}, tr.deps()) {
				t.Fatal("placement unexpectedly succeeded")
			}
			end := slices.Index(full, stage)
			if end < 0 {
				t.Fatalf("stage %q is absent from canonical order", stage)
			}
			want := full[:end+1]
			if !slices.Equal(tr.calls, want) {
				t.Fatalf("calls = %v, want failure prefix %v", tr.calls, want)
			}
		})
	}
}

func TestMapgenPlaceCEntry503B30ForwardsValuesWithoutPE32Pointer(t *testing.T) {
	old := mapgenPlaceEntry503B30
	t.Cleanup(func() { mapgenPlaceEntry503B30 = old })
	var got mapgenPlacePoint503B30
	mapgenPlaceEntry503B30 = func(x, y float32) bool {
		got = mapgenPlacePoint503B30{x: x, y: y}
		return true
	}

	result := mapgenPlaceEntryFixture503B30(123.25, -456.5)
	if !result.result || result.nullResult {
		t.Fatalf("entry results = (%t, %t), want (true, false)", result.result, result.nullResult)
	}
	if want := (mapgenPlacePoint503B30{x: 123.25, y: -456.5}); got != want {
		t.Fatalf("forwarded point = %+v, want %+v", got, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && runtime.GOOS != "windows" && result.address <= 0xffffffff {
		t.Fatalf("entry pointer = %#x, want above 4 GiB", result.address)
	}
}
