package server

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func lockTestDoor52CE90(pos types.Pointf, dir, tileX, tileY int32) *Object {
	data := &DoorUpdateData{CurrentDirection: dir, TileX: tileX, TileY: tileY, LockCode: 6}
	return &Object{ObjClass: object.ClassDoor, PosVec: pos, UpdateData: unsafe.Pointer(data)}
}

type lockTestWorld52CE90 struct {
	state                      LockCastState52CE90
	first, group               []*Object
	events                     []string
	rects                      []types.Rectf
	fps, frame                 uint32
	trace                      bool
	from, to                   types.Pointf
	audioTarget, messageTarget *Object
	after                      func(string)
}

func (w *lockTestWorld52CE90) record(event string) {
	w.events = append(w.events, event)
	if w.after != nil {
		w.after(event)
	}
}

func (w *lockTestWorld52CE90) hooks(t *testing.T) lockCastHooks52CE90 {
	return lockCastHooks52CE90{
		state: &w.state,
		rect: func(rect types.Rectf, visit func(*Object)) {
			index := len(w.rects)
			w.rects = append(w.rects, rect)
			w.record(fmt.Sprintf("rect-%d", index))
			objects := w.first
			if index != 0 {
				objects = w.group
			}
			for _, obj := range objects {
				visit(obj)
			}
		},
		trace: func(from, to types.Pointf, flags MapTraceFlags) bool {
			if flags != 0 {
				t.Fatalf("trace flags=%d", flags)
			}
			w.from, w.to = from, to
			w.record("trace")
			return w.trace
		},
		fps:   func() uint32 { w.record("fps"); return w.fps },
		frame: func() uint32 { w.record("frame"); return w.frame },
		message: func(obj *Object, id string, value uint8) {
			if id != "ExecSpel.c:DoorAlreadyLocked" || value != 0 {
				t.Fatalf("message=%s/%d", id, value)
			}
			w.messageTarget = obj
			w.record("message")
		},
		castSound: func(id int32) sound.ID {
			if id != math.MinInt32 {
				t.Fatalf("signed sound ID=%d", id)
			}
			w.record("sound")
			return sound.SoundLockCast
		},
		audio: func(id sound.ID, obj *Object, kind int, code uint32) {
			if id != sound.SoundLockCast || kind != 0 || code != 0 {
				t.Fatalf("audio=%d/%d/%d", id, kind, code)
			}
			w.audioTarget = obj
			w.record("audio")
		},
	}
}

func TestLockCast52CE90ThirdCasterFourthAimAndGroupOrder(t *testing.T) {
	caster := &Object{PosVec: types.Ptf(300, 500)}
	aim := &Object{PosVec: types.Ptf(330, 520)}
	far := lockTestDoor52CE90(types.Ptf(350, 520), 8, 15, 23)
	door := lockTestDoor52CE90(types.Ptf(340, 520), 8, 15, 23)
	tie := lockTestDoor52CE90(types.Ptf(320, 520), 8, 15, 23)
	neighbor := lockTestDoor52CE90(types.Ptf(345, 525), 24, 15, 23)
	previousOwner := new(Object)
	neighbor.ObjOwner, neighbor.Field34, neighbor.ObjFlags = previousOwner, 99, object.FlagDestroyed
	other := &Object{ObjClass: object.ClassMonster, ObjOwner: previousOwner, Field34: 97}
	w := &lockTestWorld52CE90{first: []*Object{other, far, door, tie}, group: []*Object{door, neighbor, other}, trace: true, fps: 30, frame: 100}
	beforeCaster, beforeAim, beforeOther, beforeFar, beforeTie := *caster, *aim, *other, *far, *tie
	beforeUpdate := *neighbor.UpdateDataDoor()
	if got := lockCast52CE90(math.MinInt32, caster, aim, w.hooks(t)); got != 1 {
		t.Fatalf("result=%d", got)
	}
	if w.state.Selected != door || w.state.Nearest != 100 || w.state.GroupOwner != nil || w.audioTarget != door || w.messageTarget != nil {
		t.Fatalf("selected/state/audio/message=%+v/%p/%p", w.state, w.audioTarget, w.messageTarget)
	}
	if door.ObjOwner != caster || neighbor.ObjOwner != caster || door.Field34 != 1900 || neighbor.Field34 != 1900 {
		t.Fatalf("actual lock state=%p/%p/%d/%d", door.ObjOwner, neighbor.ObjOwner, door.Field34, neighbor.Field34)
	}
	wantRects := []types.Rectf{{Min: types.Ptf(150, 350), Max: types.Ptf(450, 650)}, {Min: types.Ptf(311, 495), Max: types.Ptf(379, 563)}}
	wantEvents := []string{"rect-0", "trace", "trace", "fps", "frame", "rect-1", "fps", "frame", "fps", "frame", "sound", "audio"}
	if !reflect.DeepEqual(w.rects, wantRects) || !reflect.DeepEqual(w.events, wantEvents) || w.from != aim.PosVec || w.to != types.Ptf(351.5, 508.5) {
		t.Fatalf("geometry/events=%v/%v trace=%v..%v", w.rects, w.events, w.from, w.to)
	}
	if *caster != beforeCaster || *aim != beforeAim || *other != beforeOther || *far != beforeFar || *tie != beforeTie || *neighbor.UpdateDataDoor() != beforeUpdate {
		t.Fatal("selectors, rejected objects, or door update record changed")
	}
}

func TestLockCast52CE90OwnerGateDoesNotExpireOrFilterFlags(t *testing.T) {
	for _, owner := range []string{"nil", "self", "other-expired", "other-live"} {
		for _, flags := range []object.Flags{0, object.FlagDestroyed, object.FlagActive | object.FlagDead} {
			t.Run(fmt.Sprintf("%s/%d", owner, flags), func(t *testing.T) {
				caster := &Object{PosVec: types.Ptf(300, 300), ObjClass: object.ClassMonster}
				door := lockTestDoor52CE90(types.Ptf(305, 300), 4, 13, 13)
				door.ObjFlags = flags
				if owner == "self" {
					door.ObjOwner = caster
				}
				if owner == "other-expired" || owner == "other-live" {
					door.ObjOwner = new(Object)
				}
				door.Field34 = 1
				if owner == "other-live" {
					door.Field34 = 5000
				}
				before := *door
				w := &lockTestWorld52CE90{first: []*Object{door}, trace: true, fps: 30, frame: 100}
				got := lockCast52CE90(math.MinInt32, caster, caster, w.hooks(t))
				if owner == "other-expired" || owner == "other-live" {
					if got != 0 || *door != before || w.messageTarget != caster || !reflect.DeepEqual(w.events, []string{"rect-0", "trace", "message"}) {
						t.Fatalf("denial=%d/%+v", got, w)
					}
				} else if got != 1 || door.ObjOwner != caster || door.Field34 != 1900 || w.messageTarget != nil {
					t.Fatalf("lock=%d/%+v", got, w)
				}
			})
		}
	}
}

func TestLockCandidate52CF90SpillCutoffStrictTiesAndUnordered(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pos     types.Pointf
		nearest float32
		accept  bool
	}{
		{"zero", types.Ptf(0, 0), 1e8, true},
		{"boundary", types.Ptf(150, 0), 1e8, true},
		{"spilled-boundary", types.Ptf(150, .0001), 1e8, true},
		{"outside", types.Ptf(math.Nextafter32(150, 151), 0), 1e8, false},
		{"strict-tie", types.Ptf(150, .0001), 22500, false},
		{"closer", types.Ptf(150, .0001), math.Nextafter32(22500, 23000), true},
		{"nan-distance", types.Ptf(float32(math.NaN()), 0), 1e8, true},
		{"nan-nearest", types.Ptf(1, 0), float32(math.NaN()), true},
		{"infinite-distance", types.Ptf(float32(math.Inf(1)), 0), 1e8, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			door := lockTestDoor52CE90(types.Ptf(0, 0), 0, 0, 0)
			aim := &Object{PosVec: tc.pos}
			w := &lockTestWorld52CE90{state: LockCastState52CE90{Nearest: tc.nearest}, trace: true}
			lockCastCandidate52CF90(door, aim, w.hooks(t))
			if (w.state.Selected == door) != tc.accept || len(w.events) != map[bool]int{true: 1, false: 0}[tc.accept] {
				t.Fatalf("selection/events=%p/%v", w.state.Selected, w.events)
			}
			if tc.accept {
				want := lockRefDistance52CF90(tc.pos, door.PosVec)
				if !lockSameFloat52CE90(w.state.Nearest, want) {
					t.Fatalf("spilled nearest=%x want=%x", math.Float32bits(w.state.Nearest), math.Float32bits(want))
				}
			}
		})
	}
}

func TestLockCandidate52CF90OriginalDirectionTable(t *testing.T) {
	table := [32][2]int32{{-23, -23}, {-18, -27}, {-12, -30}, {-6, -31}, {0, -32}, {6, -31}, {12, -30}, {18, -27}, {23, -23}, {27, -18}, {30, -12}, {31, -6}, {32, 0}, {31, 6}, {30, 12}, {27, 18}, {23, 23}, {18, 27}, {12, 30}, {6, 31}, {0, 32}, {-6, 31}, {-12, 30}, {-18, 27}, {-23, 23}, {-27, 18}, {-30, 12}, {-31, 6}, {-32, 0}, {-31, -6}, {-30, -12}, {-27, -18}}
	for dir, offset := range table {
		t.Run(fmt.Sprint(dir), func(t *testing.T) {
			door := lockTestDoor52CE90(types.Ptf(300, 500), int32(dir), 0, 0)
			aim := &Object{PosVec: types.Ptf(310, 505)}
			w := &lockTestWorld52CE90{state: LockCastState52CE90{Nearest: 1e8}, trace: true}
			lockCastCandidate52CF90(door, aim, w.hooks(t))
			want := types.Ptf(300+float32(offset[0])*.5, 500+float32(offset[1])*.5)
			if w.state.Selected != door || w.from != aim.PosVec || w.to != want {
				t.Fatalf("trace=%v..%v want endpoint=%v", w.from, w.to, want)
			}
		})
	}
}

func lockRefOp52CE90(a, b float64, multiply bool) float64 {
	if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) {
		if multiply {
			return a * b
		}
		return a + b
	}
	x := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(a)
	y := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(b)
	if multiply {
		x.Mul(x, y)
	} else {
		x.Add(x, y)
	}
	v, _ := x.Float64()
	return v
}

func lockRefSpill52CE90(value float64) float32 {
	v := float32(value)
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return v
	}
	if math.Abs(float64(v)) > math.Abs(value) {
		v = math.Nextafter32(v, 0)
	}
	return v
}

func lockRefDistance52CF90(a, b types.Pointf) float32 {
	dx := lockRefOp52CE90(float64(a.X), -float64(b.X), false)
	dy := lockRefOp52CE90(float64(a.Y), -float64(b.Y), false)
	yy := lockRefOp52CE90(dy, dy, true)
	xx := lockRefOp52CE90(dx, dx, true)
	return lockRefSpill52CE90(lockRefOp52CE90(yy, xx, false))
}

func lockSameFloat52CE90(a, b float32) bool {
	return math.Float32bits(a) == math.Float32bits(b) || math.IsNaN(float64(a)) && math.IsNaN(float64(b))
}

func TestLockCast52CE90Precision53ChopIndependentReference(t *testing.T) {
	points := []types.Pointf{types.Ptf(0, 0), types.Ptf(150, .0001), types.Ptf(math.Nextafter32(150, 0), .5), types.Ptf(-150, math.SmallestNonzeroFloat32), types.Ptf(300, 500), types.Ptf(16777216, -16777216), types.Ptf(math.MaxFloat32, -math.MaxFloat32), types.Ptf(1e30, -1e30), types.Ptf(-1e-20, 1e-20)}
	for ai, a := range points {
		for bi, b := range points {
			t.Run(fmt.Sprintf("distance-%d-%d", ai, bi), func(t *testing.T) {
				if got, want := lockCastDistance52CF90(a, b), lockRefDistance52CF90(a, b); !lockSameFloat52CE90(got, want) {
					t.Fatalf("distance bits=%x want=%x", math.Float32bits(got), math.Float32bits(want))
				}
			})
		}
		t.Run(fmt.Sprintf("rect-%d", ai), func(t *testing.T) {
			w := &lockTestWorld52CE90{}
			if got := lockCast52CE90(math.MinInt32, &Object{PosVec: a}, nil, w.hooks(t)); got != 0 {
				t.Fatalf("empty rect result=%d", got)
			}
			want := types.Rectf{Min: types.Ptf(lockRefSpill52CE90(lockRefOp52CE90(float64(a.X), -150, false)), lockRefSpill52CE90(lockRefOp52CE90(float64(a.Y), -150, false))), Max: types.Ptf(lockRefSpill52CE90(lockRefOp52CE90(float64(a.X), 150, false)), lockRefSpill52CE90(lockRefOp52CE90(float64(a.Y), 150, false)))}
			if len(w.rects) != 1 || w.rects[0] != want {
				t.Fatalf("rect=%v want=%v", w.rects, want)
			}
		})
	}
}

func TestLockGroup52D060SignedWrappingTileBounds(t *testing.T) {
	for _, tileX := range []int32{math.MinInt32, -93368854, -1, 0, 1, 93368854, math.MaxInt32} {
		for _, tileY := range []int32{math.MinInt32, -93368854, -1, 0, 1, 93368854, math.MaxInt32} {
			t.Run(fmt.Sprintf("%d/%d", tileX, tileY), func(t *testing.T) {
				caster := new(Object)
				door := lockTestDoor52CE90(types.Ptf(0, 0), 0, tileX, tileY)
				w := &lockTestWorld52CE90{}
				w.after = func(event string) {
					if event == "rect-0" && w.state.GroupOwner != caster {
						t.Fatal("group owner not stored before query")
					}
				}
				lockCastGroup52D060(door, caster, w.hooks(t))
				x, y := float64(int32(uint32(tileX)*23)), float64(int32(uint32(tileY)*23))
				want := types.Rectf{Min: types.Ptf(lockRefSpill52CE90(x-34), lockRefSpill52CE90(y-34)), Max: types.Ptf(lockRefSpill52CE90(x+34), lockRefSpill52CE90(y+34))}
				if len(w.rects) != 1 || w.rects[0] != want || w.state.GroupOwner != nil {
					t.Fatalf("rect/state=%v/%+v want=%v", w.rects, w.state, want)
				}
			})
		}
	}
}

func TestLockGroup52CE60OwnerSnapshotFPSBeforeFrameAndUint32Wrap(t *testing.T) {
	for _, fps := range []uint32{0, 1, 30, math.MaxUint32} {
		for _, frame := range []uint32{0, 100, math.MaxUint32} {
			t.Run(fmt.Sprintf("%d/%d", fps, frame), func(t *testing.T) {
				owner, replacement := new(Object), new(Object)
				door := lockTestDoor52CE90(types.Ptf(0, 0), 0, 0, 0)
				w := &lockTestWorld52CE90{state: LockCastState52CE90{GroupOwner: owner}, fps: fps, frame: frame}
				w.after = func(event string) {
					if event == "fps" {
						if door.ObjOwner != owner {
							t.Fatal("owner write must precede FPS")
						}
						w.state.GroupOwner = replacement
					}
					if event == "frame" {
						w.fps = 0
					}
				}
				lockCastGroupDoor52CE60(door, w.hooks(t))
				if door.ObjOwner != owner || door.Field34 != frame+60*fps || !reflect.DeepEqual(w.events, []string{"fps", "frame"}) {
					t.Fatalf("state/order=%+v/%v", door, w.events)
				}
			})
		}
	}
}

func TestLockCast52CE90ReentrantTraceUsesLiveSharedSelection(t *testing.T) {
	for _, accepted := range []bool{false, true} {
		t.Run(fmt.Sprint(accepted), func(t *testing.T) {
			candidate := lockTestDoor52CE90(types.Ptf(2, 0), 0, 0, 0)
			reentrant := lockTestDoor52CE90(types.Ptf(7, 0), 0, 0, 0)
			w := &lockTestWorld52CE90{state: LockCastState52CE90{Nearest: 1e8}, trace: accepted}
			w.after = func(event string) {
				if event == "trace" {
					w.state.Selected, w.state.Nearest = reentrant, 49
					candidate.PosVec = types.Ptf(1000, 1000)
				}
			}
			lockCastCandidate52CF90(candidate, new(Object), w.hooks(t))
			want, wantDistance := reentrant, float32(49)
			if accepted {
				want, wantDistance = candidate, 4
			}
			if w.state.Selected != want || w.state.Nearest != wantDistance {
				t.Fatalf("nested selection=%+v want=%p/%v", w.state, want, wantDistance)
			}
		})
	}
}

func TestLockCast52CE90SharedReloadAfterTimerAndAudioSnapshot(t *testing.T) {
	caster := new(Object)
	initial := lockTestDoor52CE90(types.Ptf(1, 0), 0, 1, 2)
	reloaded := lockTestDoor52CE90(types.Ptf(2, 0), 0, 3, 4)
	audioTarget := lockTestDoor52CE90(types.Ptf(3, 0), 0, 5, 6)
	w := &lockTestWorld52CE90{first: []*Object{initial}, trace: true, fps: 30, frame: 100}
	w.after = func(event string) {
		switch event {
		case "fps":
			if initial.ObjOwner != caster {
				t.Fatal("initial owner write missing")
			}
			w.state.Selected = reloaded
		case "rect-1":
			if reloaded.Field34 != 1900 || w.state.GroupOwner != caster {
				t.Fatal("live timer/group selected pointer missing")
			}
			w.state.Selected = audioTarget
		case "sound":
			w.state.Selected = nil
		}
	}
	if got := lockCast52CE90(math.MinInt32, caster, caster, w.hooks(t)); got != 1 || initial.Field34 != 0 || initial.ObjOwner != caster || reloaded.ObjOwner != nil || reloaded.Field34 != 1900 || w.audioTarget != audioTarget || w.state.Selected != nil || w.state.GroupOwner != nil {
		t.Fatalf("result/state=%d/%+v", got, w)
	}
	if w.rects[1] != (types.Rectf{Min: types.Ptf(35, 58), Max: types.Ptf(103, 126)}) {
		t.Fatalf("reloaded group rect=%v", w.rects[1])
	}
}

func TestLockCast52CE90FaultPrefixesDoNotCleanSharedGroupOwner(t *testing.T) {
	want := []string{"rect-0", "trace", "fps", "frame", "rect-1", "fps", "frame", "sound", "audio"}
	for fault := 0; fault <= len(want); fault++ {
		t.Run(fmt.Sprint(fault), func(t *testing.T) {
			caster := new(Object)
			door := lockTestDoor52CE90(types.Ptf(1, 0), 0, 0, 0)
			w := &lockTestWorld52CE90{first: []*Object{door}, group: []*Object{door}, trace: true, fps: 30, frame: 100}
			w.after = func(string) {
				if fault != 0 && len(w.events) == fault {
					panic("service fault")
				}
			}
			var recovered any
			got := int32(-99)
			func() {
				defer func() { recovered = recover() }()
				got = lockCast52CE90(math.MinInt32, caster, caster, w.hooks(t))
			}()
			count := len(want)
			if fault != 0 {
				count = fault
			}
			if !reflect.DeepEqual(w.events, want[:count]) || (recovered != nil) != (fault != 0) || (fault == 0 && got != 1) {
				t.Fatalf("fault/result/prefix=%v/%d/%v", recovered, got, w.events)
			}
			wantOwner := (*Object)(nil)
			if fault >= 5 && fault <= 7 {
				wantOwner = caster
			}
			if w.state.GroupOwner != wantOwner {
				t.Fatalf("fault group owner=%p want=%p", w.state.GroupOwner, wantOwner)
			}
		})
	}
}

func TestLockCast52CE90RequiredPointersAndServicesFailAtUse(t *testing.T) {
	for _, stage := range []string{"nil-caster", "nil-state", "nil-rect", "nil-candidate", "nil-aim", "nil-update", "nil-trace", "nil-fps", "nil-frame", "nil-sound", "nil-audio", "nil-message"} {
		t.Run(stage, func(t *testing.T) {
			caster, aim := new(Object), new(Object)
			door := lockTestDoor52CE90(types.Ptf(1, 0), 0, 0, 0)
			w := &lockTestWorld52CE90{first: []*Object{door}, trace: true, fps: 30, frame: 100}
			h := w.hooks(t)
			switch stage {
			case "nil-caster":
				caster = nil
			case "nil-state":
				h.state = nil
			case "nil-rect":
				h.rect = nil
			case "nil-candidate":
				w.first = []*Object{nil}
			case "nil-aim":
				aim = nil
			case "nil-update":
				door.UpdateData = nil
			case "nil-trace":
				h.trace = nil
			case "nil-fps":
				h.fps = nil
			case "nil-frame":
				h.frame = nil
			case "nil-sound":
				h.castSound = nil
			case "nil-audio":
				h.audio = nil
			case "nil-message":
				door.ObjOwner = aim
				h.message = nil
			}
			var recovered any
			func() { defer func() { recovered = recover() }(); lockCast52CE90(math.MinInt32, caster, aim, h) }()
			if recovered == nil {
				t.Fatal("required load/service was silently skipped")
			}
		})
	}
	for _, class := range []object.Class{0, object.ClassPlayer, object.ClassMonster} {
		t.Run(fmt.Sprintf("non-door-%d", class), func(t *testing.T) {
			obj := &Object{ObjClass: class}
			w := &lockTestWorld52CE90{first: []*Object{obj}}
			h := lockCastHooks52CE90{state: &w.state, rect: w.hooks(t).rect}
			if got := lockCast52CE90(math.MinInt32, new(Object), nil, h); got != 0 || len(w.events) != 1 {
				t.Fatalf("unused nil services/aim=%d/%v", got, w.events)
			}
		})
	}
}

func TestLockCast52CE90RealIndexedDoorsNativeOwnersAndAudio(t *testing.T) {
	s := newDirectedVisionMap(t)
	s.SetTickRate(30)
	s.SetFrame(123)
	s.Spells.byID = map[spell.ID]*SpellDef{spell.SPELL_LOCK: {CastSound: sound.SoundLockCast}}
	newObject := func(value Object) *Object {
		obj, free := alloc.New(value)
		t.Cleanup(free)
		*obj = value // alloc.New takes a type witness, not an initial value.
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
			t.Fatalf("actual native object=%p want >4 GiB", obj)
		}
		return obj
	}
	caster := newObject(Object{PosVec: types.Ptf(300, 300)})
	aim := newObject(Object{PosVec: types.Ptf(310, 310)})
	otherOwner := newObject(Object{PosVec: types.Ptf(300, 300)})
	makeDoor := func(pos types.Pointf) *Object {
		data, free := alloc.New(DoorUpdateData{CurrentDirection: 8, TileX: 14, TileY: 14, LockCode: 6})
		t.Cleanup(free)
		*data = DoorUpdateData{CurrentDirection: 8, TileX: 14, TileY: 14, LockCode: 6}
		obj := newObject(Object{ObjClass: object.ClassDoor, ObjFlags: object.FlagActive, PosVec: pos, NewPos: pos, UpdateData: unsafe.Pointer(data)})
		obj.Shape.Kind = ShapeKindCircle
		obj.Shape.Circle.R, obj.Shape.Circle.R2 = 4, 16
		s.Map.AddObjectToIndex(obj)
		t.Cleanup(func() { s.Map.RemoveObjectFromIndex(obj) })
		return obj
	}
	door := makeDoor(types.Ptf(322, 322))
	neighbor := makeDoor(types.Ptf(345, 340))
	outside := makeDoor(types.Ptf(400, 400))
	neighbor.ObjOwner, neighbor.Field34 = otherOwner, math.MaxUint32
	beforeOutside := lockTestWithoutVisitToken52CE90(outside)
	state := new(LockCastState52CE90)
	if got := s.CastLock52CE90(int32(spell.SPELL_LOCK), nil, caster, aim, nil, math.MinInt32, LockCastRuntime52CE90{State: state}); got != 1 || state.Selected != door || state.GroupOwner != nil || door.ObjOwner != caster || neighbor.ObjOwner != caster || door.Field34 != 1923 || neighbor.Field34 != 1923 || lockTestWithoutVisitToken52CE90(outside) != beforeOutside || len(s.Audio.delayedObj) != 1 {
		t.Fatalf("native indexed result=%d state=%+v door=%p/%d neighbor=%p/%d outside=%p/%d audio=%v", got, state, door.ObjOwner, door.Field34, neighbor.ObjOwner, neighbor.Field34, outside.ObjOwner, outside.Field34, s.Audio.delayedObj)
	}
	if event := s.Audio.delayedObj[0]; event.ID != sound.SoundLockCast || event.Obj != door || event.Kind != 0 || event.Code != 0 {
		t.Fatalf("actual audio=%+v", event)
	}
	messages := 0
	beforeDoor, beforeNeighbor := lockTestWithoutVisitToken52CE90(door), lockTestWithoutVisitToken52CE90(neighbor)
	if got := s.CastLock52CE90(int32(spell.SPELL_LOCK), nil, otherOwner, aim, nil, math.MaxInt32, LockCastRuntime52CE90{State: state, PriorityMessage: func(unit *Object, id string, value uint8) {
		messages++
		if unit != otherOwner || id != "ExecSpel.c:DoorAlreadyLocked" || value != 0 {
			t.Fatalf("real indexed denial=%p/%s/%d", unit, id, value)
		}
	}}); got != 0 || messages != 1 || lockTestWithoutVisitToken52CE90(door) != beforeDoor || lockTestWithoutVisitToken52CE90(neighbor) != beforeNeighbor || len(s.Audio.delayedObj) != 1 {
		t.Fatalf("native foreign denial=%d messages=%d", got, messages)
	}
}

// Map.EachObjInRect writes its visit token, not spell state. Compare every
// other field without mutating the actual world object.
func lockTestWithoutVisitToken52CE90(obj *Object) Object {
	value := *obj
	value.Field62 = [2]uint32{}
	return value
}
