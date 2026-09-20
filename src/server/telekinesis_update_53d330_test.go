package server

import (
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestTelekinesisUpdateNative53D330MovesToSignedCursorWithNativePointers(t *testing.T) {
	player := &Player{CursorVec: image.Pt(-321, 654)}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{
		PosVec:     types.Ptf(12, 34),
		UpdateData: unsafe.Pointer(update),
	}
	source := &Object{ObjOwner: owner, Field32: 100}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"source": uintptr(unsafe.Pointer(source)),
			"owner":  uintptr(unsafe.Pointer(owner)),
			"update": uintptr(unsafe.Pointer(update)),
			"player": uintptr(unsafe.Pointer(player)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want above the ABI32 range", name, ptr)
			}
		}
	}

	var tracedFrom, tracedTo types.Pointf
	var tracedFlags MapTraceFlags
	var movedObject *Object
	var movedTo types.Pointf
	telekinesisUpdateNative53D330(source, telekinesisUpdateDeps53D330{
		frame: func() uint32 { return 101 },
		fps:   func() uint32 { return 30 },
		spellOffSound: func(spell.ID) sound.ID {
			t.Fatal("live telekinesis requested its off sound")
			return 0
		},
		audioEvent: func(sound.ID, *Object) { t.Fatal("live telekinesis played off audio") },
		traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			tracedFrom, tracedTo, tracedFlags = from, to, flags
			return true
		},
		move: func(got *Object, point types.Pointf) {
			movedObject, movedTo = got, point
		},
		delayedDelete: func(*Object) { t.Fatal("live telekinesis was deleted") },
		buffOff:       func(*Object, EnchantID) { t.Fatal("live telekinesis removed its buff") },
	})

	wantCursor := types.Ptf(-321, 654)
	if tracedFrom != owner.PosVec || tracedTo != wantCursor || tracedFlags != MapTraceFlags(5) {
		t.Fatalf("trace = %v -> %v flags=%d, want %v -> %v flags=5", tracedFrom, tracedTo, tracedFlags, owner.PosVec, wantCursor)
	}
	if movedObject != source || movedTo != wantCursor {
		t.Fatalf("move = %p/%v, want %p/%v", movedObject, movedTo, source, wantCursor)
	}
}

func TestTelekinesisUpdateNative53D330BlockedTraceDoesNotMove(t *testing.T) {
	player := &Player{CursorVec: image.Pt(200, 300)}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{UpdateData: unsafe.Pointer(update)}
	source := &Object{ObjOwner: owner}
	traceCalls := 0

	telekinesisUpdateNative53D330(source, telekinesisUpdateDeps53D330{
		frame: func() uint32 { return 0 },
		fps:   func() uint32 { return 30 },
		traceRay: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
			traceCalls++
			return false
		},
		move:          func(*Object, types.Pointf) { t.Fatal("blocked trace moved telekinesis") },
		delayedDelete: func(*Object) { t.Fatal("blocked trace deleted telekinesis") },
	})
	if traceCalls != 1 {
		t.Fatalf("trace calls = %d, want 1", traceCalls)
	}
}

func TestTelekinesisUpdateNative53D330ExpiryPreservesOriginalOrderAndOwnerReload(t *testing.T) {
	update := &PlayerUpdateData{Player: &Player{}}
	owner := &Object{
		ObjFlags:   object.Flags(telekinesisOwnerStopFlags53D330),
		UpdateData: unsafe.Pointer(update),
	}
	afterSound := &Object{}
	afterAudio := &Object{}
	afterDelete := &Object{}
	source := &Object{ObjOwner: owner}
	const offSound = sound.ID(123)
	var events []string

	telekinesisUpdateNative53D330(source, telekinesisUpdateDeps53D330{
		frame: func() uint32 {
			t.Fatal("owner stop flags did not short-circuit the frame check")
			return 0
		},
		fps: func() uint32 {
			t.Fatal("owner stop flags did not short-circuit the FPS check")
			return 0
		},
		spellOffSound: func(id spell.ID) sound.ID {
			events = append(events, "sound")
			if id != spell.SPELL_TELEKINESIS {
				t.Fatalf("spell ID = %d, want telekinesis", id)
			}
			source.ObjOwner = afterSound
			return offSound
		},
		audioEvent: func(id sound.ID, got *Object) {
			events = append(events, "audio")
			if id != offSound || got != owner {
				t.Fatalf("audio = %d/%p, want %d/cached owner %p", id, got, offSound, owner)
			}
			source.ObjOwner = afterAudio
		},
		traceRay: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
			t.Fatal("expired telekinesis traced")
			return false
		},
		move: func(*Object, types.Pointf) { t.Fatal("expired telekinesis moved") },
		delayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("deleted = %p, want %p", got, source)
			}
			source.ObjOwner = afterDelete
		},
		buffOff: func(got *Object, enchant EnchantID) {
			events = append(events, "off")
			if got != afterDelete || enchant != ENCHANT_TELEKINESIS {
				t.Fatalf("buff off = %p/%d, want reloaded owner %p/telekinesis", got, enchant, afterDelete)
			}
		},
	})

	if want := []string{"sound", "audio", "delete", "off"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestTelekinesisUpdateNative53D330StrictLifetimeAndUnsignedWrap(t *testing.T) {
	tests := []struct {
		name          string
		frame         uint32
		creationFrame uint32
		expired       bool
	}{
		{name: "exact boundary stays live", frame: 120, creationFrame: 100},
		{name: "one tick past boundary expires", frame: 121, creationFrame: 100, expired: true},
		{name: "wrapped frame expires unsigned", frame: 11, creationFrame: math.MaxUint32 - 10, expired: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			player := &Player{}
			update := &PlayerUpdateData{Player: player}
			owner := &Object{UpdateData: unsafe.Pointer(update)}
			source := &Object{ObjOwner: owner, Field32: tc.creationFrame}
			frameCalls, fpsCalls, traceCalls := 0, 0, 0
			soundCalls, audioCalls, deleteCalls, offCalls := 0, 0, 0, 0

			telekinesisUpdateNative53D330(source, telekinesisUpdateDeps53D330{
				frame: func() uint32 {
					frameCalls++
					return tc.frame
				},
				fps: func() uint32 {
					fpsCalls++
					return 1
				},
				spellOffSound: func(id spell.ID) sound.ID {
					soundCalls++
					if id != spell.SPELL_TELEKINESIS {
						t.Fatalf("spell ID = %d", id)
					}
					return 77
				},
				audioEvent: func(id sound.ID, got *Object) {
					audioCalls++
					if id != 77 || got != owner {
						t.Fatalf("audio = %d/%p", id, got)
					}
				},
				traceRay: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
					traceCalls++
					return false
				},
				move: func(*Object, types.Pointf) { t.Fatal("false trace moved telekinesis") },
				delayedDelete: func(got *Object) {
					deleteCalls++
					if got != source {
						t.Fatalf("deleted = %p", got)
					}
				},
				buffOff: func(got *Object, enchant EnchantID) {
					offCalls++
					if got != owner || enchant != ENCHANT_TELEKINESIS {
						t.Fatalf("buff off = %p/%d", got, enchant)
					}
				},
			})

			if frameCalls != 1 || fpsCalls != 1 {
				t.Fatalf("timer calls = frame:%d fps:%d, want 1/1", frameCalls, fpsCalls)
			}
			if tc.expired {
				if soundCalls != 1 || audioCalls != 1 || deleteCalls != 1 || offCalls != 1 || traceCalls != 0 {
					t.Fatalf("expired calls = sound:%d audio:%d delete:%d off:%d trace:%d", soundCalls, audioCalls, deleteCalls, offCalls, traceCalls)
				}
			} else if soundCalls != 0 || audioCalls != 0 || deleteCalls != 0 || offCalls != 0 || traceCalls != 1 {
				t.Fatalf("live calls = sound:%d audio:%d delete:%d off:%d trace:%d", soundCalls, audioCalls, deleteCalls, offCalls, traceCalls)
			}
		})
	}
}

func TestTelekinesisUpdateNative53D330RejectsIncompleteOwnerGraph(t *testing.T) {
	emptyUpdate := &PlayerUpdateData{}
	tests := []struct {
		name  string
		owner *Object
	}{
		{name: "missing owner"},
		{name: "missing update data", owner: &Object{}},
		{name: "missing player", owner: &Object{UpdateData: unsafe.Pointer(emptyUpdate)}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &Object{ObjOwner: tc.owner}
			deleteCalls := 0
			telekinesisUpdateNative53D330(source, telekinesisUpdateDeps53D330{
				frame: func() uint32 { return 0 },
				fps:   func() uint32 { return 30 },
				spellOffSound: func(spell.ID) sound.ID {
					t.Fatal("incomplete owner requested off sound")
					return 0
				},
				audioEvent: func(sound.ID, *Object) { t.Fatal("incomplete owner played audio") },
				traceRay: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
					t.Fatal("incomplete owner traced")
					return false
				},
				move: func(*Object, types.Pointf) { t.Fatal("incomplete owner moved") },
				delayedDelete: func(got *Object) {
					deleteCalls++
					if got != source {
						t.Fatalf("deleted = %p, want %p", got, source)
					}
				},
				buffOff: func(*Object, EnchantID) { t.Fatal("incomplete owner removed buff") },
			})
			if deleteCalls != 1 {
				t.Fatalf("delete calls = %d, want 1", deleteCalls)
			}
		})
	}
}

func TestTelekinesisUpdate53D330NilSourceIsIgnored(t *testing.T) {
	new(Server).TelekinesisUpdate53D330(nil, TelekinesisUpdateRuntime53D330{
		Move:          func(*Object, types.Pointf) { t.Fatal("nil source moved") },
		DelayedDelete: func(*Object) { t.Fatal("nil source deleted") },
		BuffOff:       func(*Object, EnchantID) { t.Fatal("nil source removed buff") },
	})
}
