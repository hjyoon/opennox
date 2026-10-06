package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func burnCastAlloc52C3E0[T comparable](t *testing.T, value T) *T {
	t.Helper()
	ptr, free := alloc.New(value)
	t.Cleanup(free)
	*ptr = value
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ptr)) <= math.MaxUint32 {
		t.Fatalf("native allocation %p must be above 4 GiB", ptr)
	}
	return ptr
}

func TestBurnCastNativePointersLiveOrder52C3E0(t *testing.T) {
	for _, isGlyph := range []bool{false, true} {
		t.Run(fmt.Sprintf("glyph-%t", isGlyph), func(t *testing.T) {
			second := burnCastAlloc52C3E0(t, Object{Field29: 123})
			owner := burnCastAlloc52C3E0(t, Object{Field129: second})
			caster := burnCastAlloc52C3E0(t, Object{TypeInd: 12, PosVec: types.Ptf(10, 20)})
			if isGlyph {
				caster.TypeInd = 11
			}
			arg := burnCastAlloc52C3E0(t, SpellAcceptArg{Obj: second, Pos: types.Ptf(30, 40)})
			flame := burnCastAlloc52C3E0(t, Object{Field5: 0x1234, ObjFlags: 0x100, Field29: 0x12345678})
			glyphCache := burnCastAlloc52C3E0(t, uint32(0))
			flameCache := burnCastAlloc52C3E0(t, uint32(0))
			beforeSecond, beforeOwner, wantCaster, wantFlame, wantArg := *second, *owner, *caster, *flame, *arg
			var events []string
			const id = int32(math.MinInt32 + 5)
			h := burnCastDeps52C3E0{
				loadGlyph: func() uint32 { events = append(events, "glyph.load"); return *glyphCache },
				storeGlyph: func(kind uint32) {
					events = append(events, "glyph.store")
					if kind != 11 {
						t.Fatal(kind)
					}
					// Comparison keeps EAX, not this live cache or its low WORD.
					*glyphCache = 0x1000b
				},
				loadFlame: func() uint32 { events = append(events, "flame.load"); return *flameCache },
				storeFlame: func(kind uint32) {
					events = append(events, "flame.store")
					if kind != 23 {
						t.Fatal(kind)
					}
					*flameCache = 0x80000017
				},
				lookupType: func(key string) uint32 {
					events = append(events, "lookup:"+key)
					switch key {
					case "Glyph":
						caster.PosVec, arg.Pos = types.Ptf(50, 60), types.Ptf(70, 80)
						return 11
					case "MediumFlame":
						caster.PosVec, arg.Pos = types.Ptf(90, 100), types.Ptf(110, 120)
						return 23
					default:
						t.Fatal(key)
						return 0
					}
				},
				traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
					events = append(events, "trace")
					if isGlyph || from != types.Ptf(50, 60) || to != types.Ptf(70, 80) || flags != 9 {
						t.Fatalf("trace=%v/%v/%d glyph=%t", from, to, flags, isGlyph)
					}
					arg.Pos = types.Ptf(130, 140)
					return true
				},
				newObject: func(kind uint32) *Object {
					events = append(events, "new")
					if kind != 23 {
						t.Fatalf("local allocation DWORD=%08x", kind)
					}
					caster.PosVec, arg.Pos = types.Ptf(150, 160), types.Ptf(170, 180)
					return flame
				},
				createAt: func(got, by *Object, pos types.Pointf) {
					events = append(events, "create")
					want := types.Ptf(170, 180)
					if isGlyph {
						want = types.Ptf(150, 160)
					}
					if got != flame || by != caster || by == owner || pos != want || arg.Obj != second {
						t.Fatalf("placement=%p/%p/%v want=%v", got, by, pos, want)
					}
					flame.PosVec = types.Ptf(200, 220)
				},
				balance: func(key string) float64 {
					events = append(events, "balance:"+key)
					if key != "BurnDuration" || flame.PosVec != types.Ptf(200, 220) {
						t.Fatal("balance before placement")
					}
					return 2.50000001 // Binary32 chop spills 2.5; integer chop keeps 2.
				},
				setDecay: func(got *Object, delay uint32) {
					events = append(events, "decay")
					if got != flame || delay != 2 {
						t.Fatalf("decay=%p/%08x", got, delay)
					}
					flame.PosVec = types.Ptf(250, 260)
				},
				spark: func(pos types.Pointf, count byte) {
					events = append(events, "spark")
					if pos != types.Ptf(250, 260) || count != 64 {
						t.Fatalf("live spark=%v/%d", pos, count)
					}
				},
				castSound: func(got int32) sound.ID {
					events = append(events, "cast-sound")
					if got != id {
						t.Fatalf("sound spell DWORD=%08x", uint32(got))
					}
					arg.Pos = types.Ptf(270, 280)
					return sound.ID(123)
				},
				audio: func(id sound.ID, pos types.Pointf, kind int, code uint32) {
					events = append(events, "audio")
					if id != 123 || pos != types.Ptf(270, 280) || kind != 0 || code != 0 {
						t.Fatalf("live argument audio=%d/%v/%d/%d", id, pos, kind, code)
					}
				},
				inform: func(uint8, byte, int32) { t.Fatal("unexpected notification") },
			}
			got := burnCast52C3E0(id, caster, arg, h)
			want := []string{"glyph.load", "lookup:Glyph", "glyph.store"}
			if !isGlyph {
				want = append(want, "trace")
			}
			want = append(want, "flame.load", "lookup:MediumFlame", "flame.store", "new", "create", "balance:BurnDuration", "decay", "spark", "cast-sound", "audio")
			wantCaster.PosVec, wantFlame.PosVec, wantArg.Pos = types.Ptf(150, 160), types.Ptf(250, 260), types.Ptf(270, 280)
			if got != 1 || !reflect.DeepEqual(events, want) || *second != beforeSecond || *owner != beforeOwner ||
				*caster != wantCaster || *flame != wantFlame || *arg != wantArg || *glyphCache != 0x1000b || *flameCache != 0x80000017 {
				t.Fatalf("result/order/raw records=%d/%v want=%v", got, events, want)
			}
		})
	}
}

func TestBurnCastCacheAndAllocation52C3E0(t *testing.T) {
	for _, word := range []uint32{0, 11, 0x1000b, 0x7fffffff, 0x80000000, 0xffffffff} {
		for _, allocated := range []bool{false, true} {
			t.Run(fmt.Sprintf("cache-%08x/allocated-%t", word, allocated), func(t *testing.T) {
				glyph, flameKind := word, word
				flame, caster := &Object{}, &Object{TypeInd: uint16(word), PosVec: types.Ptf(10, 20)}
				arg := &SpellAcceptArg{Obj: flame, Pos: types.Ptf(30, 40)}
				beforeFlame, beforeCaster, beforeArg := *flame, *caster, *arg
				glyphLoads, flameLoads, lookups, stores, traces, allocations, placements, balances, decays, sparks, sounds, audios := 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
				h := burnCastDeps52C3E0{
					loadGlyph:  func() uint32 { glyphLoads++; return glyph },
					storeGlyph: func(kind uint32) { stores++; glyph = kind },
					loadFlame:  func() uint32 { flameLoads++; return flameKind },
					storeFlame: func(kind uint32) { stores++; flameKind = kind },
					lookupType: func(key string) uint32 {
						lookups++
						switch key {
						case "Glyph":
							return 11
						case "MediumFlame":
							return 23
						default:
							t.Fatal(key)
							return 0
						}
					},
					traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
						traces++
						if from != caster.PosVec || to != arg.Pos || flags != 9 {
							t.Fatal("wrong trace")
						}
						return true
					},
					newObject: func(kind uint32) *Object {
						allocations++
						if kind != flameKind {
							t.Fatalf("allocation DWORD=%08x want=%08x", kind, flameKind)
						}
						if allocated {
							return flame
						}
						return nil
					},
					createAt: func(got, by *Object, pos types.Pointf) {
						placements++
						want := arg.Pos
						if uint32(caster.TypeInd) == glyph {
							want = caster.PosVec
						}
						if got != flame || by != caster || pos != want {
							t.Fatal("wrong placement")
						}
					},
					balance: func(key string) float64 {
						balances++
						if key != "BurnDuration" {
							t.Fatal(key)
						}
						return 90.5
					},
					setDecay: func(got *Object, delay uint32) {
						decays++
						if got != flame || delay != 90 {
							t.Fatal("wrong decay")
						}
					},
					spark: func(pos types.Pointf, count byte) {
						sparks++
						if pos != flame.PosVec || count != 64 {
							t.Fatal("wrong spark")
						}
					},
					castSound: func(id int32) sound.ID { sounds++; return sound.ID(id) },
					audio: func(id sound.ID, pos types.Pointf, kind int, code uint32) {
						audios++
						if id != 5 || pos != arg.Pos || kind != 0 || code != 0 {
							t.Fatal("wrong audio")
						}
					},
					inform: func(uint8, byte, int32) { t.Fatal("unexpected notification") },
				}
				got := burnCast52C3E0(5, caster, arg, h)
				wantLookups, wantTrace, wantPlacement := 0, 1, 0
				if word == 0 {
					wantLookups = 2
				}
				if uint32(caster.TypeInd) == glyph {
					wantTrace = 0
				}
				if allocated {
					wantPlacement = 1
				}
				if got != 1 || glyphLoads != 1 || flameLoads != 1 || lookups != wantLookups || stores != wantLookups || traces != wantTrace || allocations != 1 ||
					placements != wantPlacement || balances != wantPlacement || decays != wantPlacement || sparks != wantPlacement || sounds != 1 || audios != 1 ||
					*flame != beforeFlame || *caster != beforeCaster || *arg != beforeArg {
					t.Fatalf("result/callback counts=%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d", got, glyphLoads, flameLoads, lookups, stores, traces, allocations, placements, balances, decays, sparks, sounds, audios)
				}
			})
		}
	}
}

func TestBurnCastBlockedLivePlayer52C3E0(t *testing.T) {
	for _, initialPlayer := range []bool{false, true} {
		for _, livePlayer := range []bool{false, true} {
			for _, index := range []uint8{0, 31, 255} {
				t.Run(fmt.Sprintf("initial-%t/live-%t/index-%d", initialPlayer, livePlayer, index), func(t *testing.T) {
					player := burnCastAlloc52C3E0(t, Player{PlayerInd: index})
					data := burnCastAlloc52C3E0(t, PlayerUpdateData{Player: player})
					caster := burnCastAlloc52C3E0(t, Object{ObjClass: object.ClassMonster})
					if initialPlayer {
						caster.ObjClass = object.ClassPlayer
					}
					arg := burnCastAlloc52C3E0(t, SpellAcceptArg{Pos: types.Ptf(30, 40)})
					calls := 0
					h := burnCastDeps52C3E0{
						loadGlyph: func() uint32 { return 11 },
						traceRay: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
							caster.ObjClass = object.ClassMonster
							if livePlayer {
								caster.ObjClass = object.ClassPlayer
							}
							caster.UpdateData = unsafe.Pointer(data)
							return false
						},
						inform: func(got uint8, code byte, value int32) {
							calls++
							if got != index || code != 0 || value != 2 {
								t.Fatalf("inform=%d/%d/%d", got, code, value)
							}
						},
					}
					beforePlayer, beforeData, beforeArg := *player, *data, *arg
					got := burnCast52C3E0(5, caster, arg, h)
					wantCalls := 0
					if livePlayer {
						wantCalls = 1
					}
					if got != 0 || calls != wantCalls || *player != beforePlayer || *data != beforeData || *arg != beforeArg {
						t.Fatalf("blocked result=%d calls=%d", got, calls)
					}
				})
			}
		}
	}
}

func TestBurnCastNilGuardsAfterCache52C3E0(t *testing.T) {
	for _, word := range []uint32{0, 321} {
		for _, missing := range []string{"argument", "caster", "both"} {
			t.Run(fmt.Sprintf("cache-%d/%s", word, missing), func(t *testing.T) {
				caster, arg := &Object{}, &SpellAcceptArg{}
				switch missing {
				case "argument":
					arg = nil
				case "caster":
					caster = nil
				case "both":
					caster, arg = nil, nil
				}
				cache := word
				var events []string
				h := burnCastDeps52C3E0{
					loadGlyph: func() uint32 { events = append(events, "load"); return cache },
					lookupType: func(key string) uint32 {
						events = append(events, "lookup:"+key)
						return 11
					},
					storeGlyph: func(kind uint32) { events = append(events, "store"); cache = kind },
				}
				want, wantCache := []string{"load"}, word
				if word == 0 {
					want, wantCache = append(want, "lookup:Glyph", "store"), 11
				}
				if got := burnCast52C3E0(5, caster, arg, h); got != 0 || cache != wantCache || !reflect.DeepEqual(events, want) {
					t.Fatalf("guard result/cache/order=%d/%d/%v", got, cache, events)
				}
			})
		}
	}
}

func TestBurnCastDurationSpillAndChop52C3E0(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value float64
		want  uint32
	}{
		{"zero", 0, 0}, {"negative-zero", math.Copysign(0, -1), 0},
		{"half", 0.5, 0}, {"one-half", 1.5, 1}, {"two-half", 2.5, 2},
		{"negative-half", -0.5, 0}, {"negative-one-half", -1.5, 0xffffffff}, {"negative-two-half", -2.5, 0xfffffffe},
		{"binary32-tie", 2.50000001, 2}, {"binary32-integer", 16777217, 16777216},
		{"binary32-round-down", 16777219, 16777218}, {"negative-binary32-round-up", -16777219, 0xfefffffe},
		{"binary32-fractional-below-integer", 16777215.75, 0x00ffffff}, {"negative-binary32-fractional-above-integer", -16777215.75, 0xff000001},
		{"largest-binary32-int", 2147483520, 0x7fffff80}, {"positive-overflow", 2147483648, 0x80000000},
		{"positive-binary32-boundary", 2147483647, 0x7fffff80}, {"negative-binary32-boundary", -2147483649, 0x80000000},
		{"minimum-int", -2147483648, 0x80000000}, {"negative-overflow", -2147483904, 0x80000000},
		{"nan", math.NaN(), 0x80000000}, {"positive-infinity", math.Inf(1), 0x80000000}, {"negative-infinity", math.Inf(-1), 0x80000000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flame, caster, arg := &Object{}, &Object{TypeInd: 11}, &SpellAcceptArg{}
			var events []string
			h := burnCastDeps52C3E0{
				loadGlyph: func() uint32 { return 11 },
				loadFlame: func() uint32 { return 23 },
				newObject: func(uint32) *Object { events = append(events, "new"); return flame },
				createAt: func(got, owner *Object, _ types.Pointf) {
					events = append(events, "create")
					if got != flame || owner != caster {
						t.Fatal("wrong placement")
					}
				},
				balance: func(key string) float64 {
					events = append(events, "balance")
					if key != "BurnDuration" {
						t.Fatal(key)
					}
					return tc.value
				},
				setDecay: func(got *Object, delay uint32) {
					events = append(events, "decay")
					if got != flame || delay != tc.want {
						t.Fatalf("duration=%08x want=%08x", delay, tc.want)
					}
				},
				spark:     func(types.Pointf, byte) { events = append(events, "spark") },
				castSound: func(int32) sound.ID { events = append(events, "sound"); return 123 },
				audio:     func(sound.ID, types.Pointf, int, uint32) { events = append(events, "audio") },
			}
			want := []string{"new", "create", "balance", "decay", "spark", "sound", "audio"}
			if got := burnCast52C3E0(5, caster, arg, h); got != 1 || !reflect.DeepEqual(events, want) {
				t.Fatalf("result/order=%d/%v", got, events)
			}
		})
	}
}

func TestBurnCastMissingGlyphAndFlameRetry52C3E0(t *testing.T) {
	for _, typeInd := range []uint16{0, 99} {
		t.Run(fmt.Sprintf("type-%d", typeInd), func(t *testing.T) {
			glyph, flame, sounds := uint32(0), uint32(0), 0
			var events []string
			h := burnCastDeps52C3E0{
				loadGlyph:  func() uint32 { events = append(events, "glyph.load"); return glyph },
				storeGlyph: func(value uint32) { events = append(events, "glyph.store"); glyph = value },
				loadFlame:  func() uint32 { events = append(events, "flame.load"); return flame },
				storeFlame: func(value uint32) { events = append(events, "flame.store"); flame = value },
				lookupType: func(key string) uint32 { events = append(events, "lookup:"+key); return 0 },
				traceRay:   func(types.Pointf, types.Pointf, MapTraceFlags) bool { events = append(events, "trace"); return true },
				newObject: func(kind uint32) *Object {
					events = append(events, "new")
					if kind != 0 {
						t.Fatal(kind)
					}
					return nil
				},
				castSound: func(int32) sound.ID { events = append(events, "sound"); sounds++; return 123 },
				audio:     func(sound.ID, types.Pointf, int, uint32) { events = append(events, "audio") },
			}
			want := []string{"glyph.load", "lookup:Glyph", "glyph.store"}
			if typeInd != 0 {
				want = append(want, "trace")
			}
			want = append(want, "flame.load", "lookup:MediumFlame", "flame.store", "new", "sound", "audio")
			want = append(want, want...)
			for range 2 {
				if got := burnCast52C3E0(5, &Object{TypeInd: typeInd}, &SpellAcceptArg{}, h); got != 1 {
					t.Fatal(got)
				}
			}
			if glyph != 0 || flame != 0 || sounds != 2 || !reflect.DeepEqual(events, want) {
				t.Fatalf("retry glyph/flame/sounds/order=%d/%d/%d/%v", glyph, flame, sounds, events)
			}
		})
	}
}

func TestBurnCastBlockedPlayerFaultPrefix52C3E0(t *testing.T) {
	for _, missing := range []string{"update-data", "player"} {
		t.Run(missing, func(t *testing.T) {
			caster := &Object{ObjClass: object.ClassPlayer}
			if missing == "player" {
				caster.UpdateData = unsafe.Pointer(&PlayerUpdateData{})
			}
			var events []string
			h := burnCastDeps52C3E0{
				loadGlyph: func() uint32 { events = append(events, "load"); return 11 },
				traceRay:  func(types.Pointf, types.Pointf, MapTraceFlags) bool { events = append(events, "trace"); return false },
			}
			panicked := false
			func() {
				defer func() { panicked = recover() != nil }()
				burnCast52C3E0(5, caster, &SpellAcceptArg{}, h)
			}()
			if !panicked || !reflect.DeepEqual(events, []string{"load", "trace"}) {
				t.Fatalf("original unguarded player fault=%t order=%v", panicked, events)
			}
		})
	}
}
