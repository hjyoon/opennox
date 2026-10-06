package server

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

type cleansingFlameWorld52D5C0 struct {
	t                                                                              *testing.T
	cache                                                                          [10]uint32
	coop, missing, blocked                                                         bool
	powerRoll                                                                      int32
	flames                                                                         []*Object
	rolls, directions, timers, traces, creates, deletes, adds, predictions, sounds int
	owner                                                                          *Object
}

func newCleansingFlameWorld52D5C0(t *testing.T) *cleansingFlameWorld52D5C0 {
	w := &cleansingFlameWorld52D5C0{t: t, coop: true, owner: &Object{PosVec: types.Ptf(10, 20)}}
	w.owner.Shape.Circle.R = 7
	for i := range w.cache {
		w.cache[i] = uint32(300 + i)
	}
	return w
}

func (w *cleansingFlameWorld52D5C0) hooks() cleansingFlameCastDeps52D5C0 {
	return cleansingFlameCastDeps52D5C0{
		loadCache:   func(slot int32) uint32 { return w.cache[slot] },
		storeCache:  func(slot int32, value uint32) { w.cache[slot] = value },
		lookupType:  func(string) uint32 { w.t.Fatal("warm cache performed lookup"); return 0 },
		priorityMsg: func(*Object, string, byte) { w.t.Fatal("no owned unit matched") },
		hasGameFlag: func(flag uint32) bool {
			if flag != 2048 {
				w.t.Fatalf("flag=%d", flag)
			}
			return w.coop
		},
		random: func(min, max int32) int32 {
			switch {
			case min == 0 && max == 1:
				w.rolls++
				return w.powerRoll
			case min == 0 && max == 255:
				w.directions++
				return 64
			case min == 90 && max == 180:
				w.timers++
				return 123
			default:
				w.t.Fatalf("RNG bounds=%d,%d", min, max)
				return 0
			}
		},
		newObject: func(kind uint32) *Object {
			if w.missing {
				return nil
			}
			flame := &Object{TypeInd: uint16(kind), Float28: 9}
			flame.Shape.Circle.R = 3
			w.flames = append(w.flames, flame)
			return flame
		},
		directionVector: func(direction int16) types.Pointf {
			if direction != 64 {
				w.t.Fatalf("direction=%d", direction)
			}
			return types.Ptf(0.5, -0.25)
		},
		traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			w.traces++
			if from != types.Ptf(10, 20) || to != types.Ptf(17, 16.5) || flags != 65 {
				w.t.Fatalf("trace=%v,%v,%d", from, to, flags)
			}
			return !w.blocked
		},
		createAt: func(flame, owner *Object, point types.Pointf) {
			w.creates++
			if owner != w.owner || point != types.Ptf(17, 16.5) {
				w.t.Fatal("wrong owner/position")
			}
			flame.PosVec = point
		},
		delayedDelete: func(flame *Object) {
			w.deletes++
			if flame != w.flames[len(w.flames)-1] {
				w.t.Fatal("deletion lost flame pointer")
			}
		},
		fps: func() uint32 { return 30 }, frame: func() uint32 { return 1000 },
		addUpdatable: func(flame *Object) {
			w.adds++
			if flame.Direction2 != 64 || flame.VelVec != types.Ptf(2, -1) || flame.Field34 != 1123 || flame.Pos39 != w.owner.PosVec {
				w.t.Fatalf("incomplete updatable flame=%+v", flame)
			}
		},
		predictLinear: func(flame *Object) {
			w.predictions++
			if uint32(flame.ObjClass) != 0x40000000 || math.Float32bits(flame.Float28) != 0 {
				w.t.Fatal("prediction before client-predict class/zero damping")
			}
		},
		castSound: func(int32) sound.ID { return sound.ID(123) },
		audio: func(id sound.ID, caster *Object, kind int, code uint32) {
			w.sounds++
			if id != 123 || caster != w.owner || kind != 0 || code != 0 {
				w.t.Fatal("cast audio contract")
			}
		},
	}
}

func TestCleansingFlameCast52D5C0All48Attempts(t *testing.T) {
	for _, id := range []int32{10, 11, -1} {
		for level := int32(1); level <= 5; level++ {
			for _, roll := range []int32{0, 1} {
				for _, coop := range []bool{false, true} {
					for _, branch := range []string{"placed", "blocked", "missing"} {
						t.Run(fmt.Sprintf("id%d/level%d/roll%d/coop%t/%s", id, level, roll, coop, branch), func(t *testing.T) {
							w := newCleansingFlameWorld52D5C0(t)
							w.coop, w.powerRoll, w.blocked, w.missing = coop, roll, branch == "blocked", branch == "missing"
							if got := cleansingFlameCast52D5C0(id, nil, nil, w.owner, level, w.hooks()); got != 1 {
								t.Fatalf("result=%d", got)
							}
							power := level - roll
							if !coop {
								power = 4 - roll
							}
							want := 48
							if power < 1 || w.missing {
								want = 0
							}
							if w.rolls != 48 || len(w.flames) != want || w.directions != want || w.traces != want || w.sounds != 1 {
								t.Fatalf("attempt/allocate/dir/trace/audio=%d/%d/%d/%d/%d", w.rolls, len(w.flames), w.directions, w.traces, w.sounds)
							}
							base := int32(5)
							if id == 10 {
								base = 0
							}
							for _, flame := range w.flames {
								if flame.TypeInd != uint16(w.cache[base+power-1]) {
									t.Fatalf("flame kind=%d", flame.TypeInd)
								}
							}
							if w.blocked {
								if w.deletes != want || w.creates != 0 || w.timers != 0 || w.adds != 0 || w.predictions != 0 {
									t.Fatal("blocked flame placed/predicted or consumed timer RNG")
								}
							} else if w.deletes != 0 || w.creates != want || w.timers != want || w.adds != want || w.predictions != want {
								t.Fatal("clear flame lifecycle count mismatch")
							}
							if unsafe.Sizeof(uintptr(0)) == 8 {
								for _, flame := range w.flames {
									if uintptr(unsafe.Pointer(flame)) <= math.MaxUint32 {
										t.Fatal("native flame pointer not above 4 GiB")
									}
								}
							}
							for i, value := range w.cache {
								if value != uint32(300+i) {
									t.Fatal("warm type cache changed")
								}
							}
						})
					}
				}
			}
		}
	}
}

func TestCleansingFlameCast52D5C0CacheAndOwnedGuards(t *testing.T) {
	names := [10]string{"SmallFlameCleanse", "SmallFlameCleanse", "MediumFlameCleanse", "FlameCleanse", "LargeFlameCleanse", "SmallBlueFlameCleanse", "SmallBlueFlameCleanse", "MediumBlueFlameCleanse", "BlueFlameCleanse", "LargeBlueFlameCleanse"}
	for slot := 0; slot < 10; slot++ {
		t.Run(fmt.Sprint(slot), func(t *testing.T) {
			w := newCleansingFlameWorld52D5C0(t)
			w.cache = [10]uint32{}
			recipient := &Object{}
			second := &Object{Field129: &Object{TypeInd: 77, Field128: &Object{TypeInd: uint16(300 + slot)}}}
			before := *second
			h, lookups, messages := w.hooks(), 0, 0
			h.lookupType = func(name string) uint32 {
				if name != names[lookups] {
					t.Fatalf("lookup %d=%q", lookups, name)
				}
				for previous := 0; previous < lookups; previous++ {
					if w.cache[previous] != uint32(300+previous) {
						t.Fatal("cache store was postponed")
					}
				}
				lookups++
				return uint32(299 + lookups)
			}
			h.priorityMsg = func(unit *Object, message string, code byte) {
				messages++
				if unit != recipient || message != "plyrspel.c:TooManySpells" || code != 0 {
					t.Fatal("wrong priority-message recipient")
				}
			}
			h.hasGameFlag = func(uint32) bool { t.Fatal("owned rejection reached game flag"); return false }
			if got := cleansingFlameCast52D5C0(10, second, recipient, nil, 1, h); got != 0 || lookups != 10 || messages != 1 || w.rolls != 0 || w.sounds != 0 || *second != before {
				t.Fatalf("owned guard result/lookups/messages=%d/%d/%d", got, lookups, messages)
			}
		})
	}
	// A TypeInd WORD is zero extended, not narrowed from the DWORD cache.
	w := newCleansingFlameWorld52D5C0(t)
	w.cache[0] = 0x1004d
	second := &Object{Field129: &Object{TypeInd: 77}}
	if got := cleansingFlameCast52D5C0(10, second, nil, w.owner, 0, w.hooks()); got != 1 || w.rolls != 48 || len(w.flames) != 0 {
		t.Fatal("high cache DWORD became an owned type match")
	}
}

func TestCleansingFlameCast52D5C0LiveWritesAndCallOrder(t *testing.T) {
	w := newCleansingFlameWorld52D5C0(t)
	h := w.hooks()
	callback := unsafe.Pointer(new(byte))
	h.updateCallback = callback
	var events []string
	rolls := 0
	h.loadCache = func(slot int32) uint32 { events = append(events, fmt.Sprintf("cache:%d", slot)); return w.cache[slot] }
	h.hasGameFlag = func(uint32) bool { events = append(events, "flag"); return true }
	h.random = func(min, max int32) int32 {
		if max == 1 {
			rolls++
			if rolls > 1 {
				return 1
			}
			events = append(events, "power")
			return 0
		}
		if max == 255 {
			events = append(events, "direction")
			return 0xffff
		}
		events = append(events, "timer")
		if min != -3 || max != -6 {
			t.Fatalf("wrapped signed FPS bounds=%d,%d", min, max)
		}
		w.owner.PosVec = types.Ptf(50, 60)
		return 20
	}
	originalNew := h.newObject
	h.newObject = func(kind uint32) *Object { events = append(events, "new"); return originalNew(kind) }
	h.directionVector = func(dir int16) types.Pointf {
		events = append(events, "vector")
		if dir != -1 {
			t.Fatalf("signed low direction=%d", dir)
		}
		return types.Ptf(0.5, -0.25)
	}
	h.traceRay = func(from, to types.Pointf, flags MapTraceFlags) bool {
		events = append(events, "trace")
		if from != types.Ptf(10, 20) || to != types.Ptf(17, 16.5) || flags != 65 {
			t.Fatal("trace arguments")
		}
		w.owner.PosVec = types.Ptf(30, 40)
		return true
	}
	h.createAt = func(flame, owner *Object, point types.Pointf) {
		events = append(events, "create")
		if owner != w.owner || point != types.Ptf(17, 16.5) || flame.Direction1 != 0xffff {
			t.Fatal("create snapshot/owner")
		}
		flame.Direction1 = 23
	}
	h.fps = func() uint32 { events = append(events, "fps"); return math.MaxUint32 }
	h.frame = func() uint32 {
		events = append(events, "frame")
		w.owner.PosVec = types.Ptf(70, 80)
		return math.MaxUint32 - 10
	}
	h.addUpdatable = func(flame *Object) {
		events = append(events, "add")
		if flame.Direction2 != 23 || flame.VelVec != types.Ptf(2, -1) || flame.Field34 != 9 || flame.Pos39 != types.Ptf(70, 80) || flame.Update != callback {
			t.Fatalf("post-callback flame=%+v", flame)
		}
		flame.ObjClass = object.ClassPlayer
		flame.Float28 = 999
	}
	h.predictLinear = func(flame *Object) {
		events = append(events, "predict")
		if flame.ObjClass != object.ClassPlayer|object.Class(0x40000000) || math.Float32bits(flame.Float28) != 0 {
			t.Fatal("prediction saw cached class/damping")
		}
	}
	h.castSound = func(id int32) sound.ID {
		events = append(events, "sound")
		if id != 10 {
			t.Fatal("sound id")
		}
		return 12
	}
	h.audio = func(id sound.ID, owner *Object, kind int, code uint32) {
		events = append(events, "audio")
		if id != 12 || owner != w.owner || kind != 0 || code != 0 {
			t.Fatal("audio arguments")
		}
	}
	if got := cleansingFlameCast52D5C0(10, nil, nil, w.owner, 1, h); got != 1 || rolls != 48 || len(w.flames) != 1 {
		t.Fatalf("result/rolls/flames=%d/%d/%d", got, rolls, len(w.flames))
	}
	want := []string{"cache:0", "flag", "power", "cache:0", "new", "direction", "vector", "trace", "create", "fps", "timer", "frame", "add", "predict", "sound", "audio"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("call order=%v, want %v", events, want)
	}
}

func TestCleansingFlameCast52D5C0ZeroCacheMarkerRetries(t *testing.T) {
	w := newCleansingFlameWorld52D5C0(t)
	w.cache = [10]uint32{}
	h := w.hooks()
	lookups := 0
	h.lookupType = func(string) uint32 {
		slot := lookups % 10
		lookups++
		if slot == 0 {
			return 0
		}
		return uint32(300 + slot)
	}
	for pass := 1; pass <= 2; pass++ {
		if got := cleansingFlameCast52D5C0(10, nil, nil, w.owner, 0, h); got != 1 || lookups != 10*pass || w.rolls != 48*pass || w.sounds != pass {
			t.Fatalf("pass%d result/lookups/rolls/sounds=%d/%d/%d/%d", pass, got, lookups, w.rolls, w.sounds)
		}
	}
}

func TestCleansingFlameCast52D5C0OwnedTypeSnapshotAndLiveCache(t *testing.T) {
	w := newCleansingFlameWorld52D5C0(t)
	owned := &Object{TypeInd: 77}
	second, recipient := &Object{Field129: owned}, &Object{}
	h := w.hooks()
	var slots []int32
	h.loadCache = func(slot int32) uint32 {
		slots = append(slots, slot)
		if len(slots) == 2 {
			// The owned TypeInd is already in ECX, but each cache DWORD
			// remains live. Reloading TypeInd would reject at slot zero.
			owned.TypeInd = uint16(w.cache[0])
			w.cache[1] = 77
		}
		return w.cache[slot]
	}
	messages := 0
	h.priorityMsg = func(unit *Object, message string, code byte) {
		messages++
		if unit != recipient || message != "plyrspel.c:TooManySpells" || code != 0 {
			t.Fatal("live-cache guard recipient")
		}
	}
	h.hasGameFlag = func(uint32) bool { t.Fatal("owned match reached game flag"); return false }
	if got := cleansingFlameCast52D5C0(10, second, recipient, nil, 1, h); got != 0 || messages != 1 || !reflect.DeepEqual(slots, []int32{0, 0, 1}) {
		t.Fatalf("owned snapshot/live-cache result=%d messages=%d slots=%v", got, messages, slots)
	}
}

func TestCleansingFlameCast52D5C0UnclampedWrappedLevel(t *testing.T) {
	w := newCleansingFlameWorld52D5C0(t)
	h := w.hooks()
	w.powerRoll = 1
	h.loadCache = func(slot int32) uint32 {
		if slot != 0 && slot != math.MaxInt32-1 {
			t.Fatalf("wrapped slot=%d", slot)
		}
		return 300
	}
	h.newObject = func(uint32) *Object { return nil }
	h.audio = func(id sound.ID, caster *Object, kind int, code uint32) {
		w.sounds++
		if id != 123 || caster != nil || kind != 0 || code != 0 {
			t.Fatal("allocation-free cast did not preserve nil-caster audio")
		}
	}
	if got := cleansingFlameCast52D5C0(10, nil, nil, nil, math.MinInt32, h); got != 1 || w.rolls != 48 || w.sounds != 1 {
		t.Fatal("signed DWORD level subtraction was clamped")
	}
}

func TestCleansingFlameCast52D5C0CacheOffsetWrap(t *testing.T) {
	for _, slot := range []int32{-1, 0, 9, math.MinInt32, math.MaxInt32} {
		want := uintptr(uint32(uint64(2487760) + 4*uint64(uint32(slot))))
		if got := cleansingFlameCacheOffset52D5C0(slot); got != want {
			t.Fatalf("cache offset(%d)=%x, want %x", slot, got, want)
		}
	}
}

func TestCleansingFlameCast52D5C0IndependentChopCoordinates(t *testing.T) {
	// Independent precision-53 ToZero operations and a precision-24 spill;
	// this reference never calls production chop/spill/Nextafter helpers.
	op := func(a, b float64, multiply bool) float64 {
		x, y := new(big.Float).SetFloat64(a), new(big.Float).SetFloat64(b)
		z := new(big.Float).SetPrec(53).SetMode(big.ToZero)
		if multiply {
			z.Mul(x, y)
		} else {
			z.Add(x, y)
		}
		value, _ := z.Float64()
		return value
	}
	spill := func(value float64) float32 {
		v, _ := new(big.Float).SetPrec(24).SetMode(big.ToZero).SetFloat64(value).Float32()
		return v
	}
	seed := uint32(0x5235d5c0)
	for i := 0; i < 1024; i++ {
		word := func() float32 {
			seed = seed*1664525 + 1013904223
			return math.Float32frombits((seed & 0x807fffff) | uint32(80+int(seed%90))<<23)
		}
		casterRadius, flameRadius, cosine, sine, x, y := word(), word(), word(), word(), word(), word()
		w := newCleansingFlameWorld52D5C0(t)
		w.owner.Shape.Circle.R, w.owner.PosVec = casterRadius, types.Ptf(x, y)
		h := w.hooks()
		rolls := 0
		h.random = func(_, max int32) int32 {
			if max == 1 {
				rolls++
				if rolls > 1 {
					return 1
				}
			}
			return 0
		}
		h.newObject = func(uint32) *Object { obj := new(Object); obj.Shape.Circle.R = flameRadius; return obj }
		h.directionVector = func(int16) types.Pointf { return types.Ptf(cosine, sine) }
		distance := op(op(float64(flameRadius), float64(casterRadius), false), 4, false)
		want := types.Ptf(spill(op(op(distance, float64(cosine), true), float64(x), false)), spill(op(op(distance, float64(sine), true), float64(y), false)))
		h.traceRay = func(from, to types.Pointf, flags MapTraceFlags) bool {
			if from != w.owner.PosVec || math.Float32bits(to.X) != math.Float32bits(want.X) || math.Float32bits(to.Y) != math.Float32bits(want.Y) || flags != 65 {
				t.Fatalf("case%d: coordinate bits=%08x,%08x, want %08x,%08x", i, math.Float32bits(to.X), math.Float32bits(to.Y), math.Float32bits(want.X), math.Float32bits(want.Y))
			}
			return false
		}
		h.delayedDelete = func(*Object) {}
		cleansingFlameCast52D5C0(10, nil, nil, w.owner, 1, h)
		if rolls != 48 {
			t.Fatalf("reference case%d roll count=%d", i, rolls)
		}
	}
}
