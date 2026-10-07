package server

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

func earthquakeAlloc52DE40[T comparable](t *testing.T, value T) *T {
	t.Helper()
	ptr, free := alloc.New(value)
	t.Cleanup(free)
	*ptr = value
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ptr)) <= math.MaxUint32 {
		t.Fatalf("native earthquake pointer below 4 GiB: %p", ptr)
	}
	return ptr
}

func earthquakeTestDeps52DE40(level *int32) earthquakeCastDeps52DE40 {
	return earthquakeCastDeps52DE40{
		storeLevel: func(value int32) { *level = value },
		loadLevel:  func() int32 { return *level },
		balance:    func(string) float32 { return 100 },
		indexed:    func(string, int32) float32 { return 40 },
		circle:     func(*Object, float32, func(*Object)) {},
		owner:      (*Object).FindOwnerChainPlayer,
		hp:         UnitGetHP4EE780,
		distance:   ObjectDistance4E6C00,
		damage:     func(*Object, *Object, *Object, int32, object.DamageType) {},
		castSound:  func(int32) sound.ID { return 123 },
		audio:      func(sound.ID, *Object, int, uint32) {},
		quake:      func(*Object, int32) {},
	}
}

func TestEarthquakeCast52DE40NativePointersLiveOrderAndReentrancy(t *testing.T) {
	owner := earthquakeAlloc52DE40(t, Object{ObjClass: object.ClassPlayer})
	laterOwner := earthquakeAlloc52DE40(t, Object{ObjClass: object.ClassPlayer})
	caster := earthquakeAlloc52DE40(t, Object{ObjOwner: owner, PosVec: types.Ptf(10, 20), Field29: 29})
	health := earthquakeAlloc52DE40(t, HealthData{Cur: 0xffff, Max: 0xffff})
	target := earthquakeAlloc52DE40(t, Object{HealthData: health, Field29: 31})
	var level int32
	var events []string
	h := earthquakeTestDeps52DE40(&level)
	h.storeLevel = func(value int32) { events = append(events, "store"); level = value }
	rangeCalls := 0
	h.balance = func(key string) float32 {
		events = append(events, "range")
		if key != "EarthquakeRange" {
			t.Fatal(key)
		}
		rangeCalls++
		if rangeCalls == 1 {
			if level != 3 {
				t.Fatal("range lookup preceded DWORD store")
			}
			caster.PosVec = types.Ptf(30, 40)
		} else {
			level = 5 // Nested cast changes the live global before its read.
		}
		return 100
	}
	h.circle = func(got *Object, radius float32, callback func(*Object)) {
		events = append(events, "circle")
		if got != caster || got.PosVec != types.Ptf(30, 40) || radius != 100 {
			t.Fatal("circle used a stale position or narrowed caster")
		}
		callback(target)
		caster.PosVec = types.Ptf(50, 60)
	}
	h.owner = func(got *Object) *Object {
		events = append(events, "owner")
		if got != caster {
			t.Fatal("owner source narrowed")
		}
		return got.FindOwnerChainPlayer()
	}
	h.hp = func(got *Object) uint16 { events = append(events, "hp"); return UnitGetHP4EE780(got) }
	h.distance = func(got, source *Object) float64 {
		events = append(events, "distance")
		if got != target || source != caster {
			t.Fatal("distance source narrowed")
		}
		caster.ObjOwner = laterOwner // The earlier resolved owner stays cached.
		return 25
	}
	h.loadLevel = func() int32 { events = append(events, "load"); return level }
	h.indexed = func(key string, index int32) float32 {
		events = append(events, key)
		if key == "EarthquakeDamage" {
			if index != 4 {
				t.Fatalf("damage ignored live DWORD level: %d", index)
			}
			level = -123 // Index must already have been cached.
			return 80
		}
		if key != "EarthquakeJiggle" || index != 2 {
			t.Fatalf("jiggle did not retain entry power: %s/%d", key, index)
		}
		caster.PosVec = types.Ptf(90, 100)
		return 9.75
	}
	h.damage = func(got, by, source *Object, amount int32, typ object.DamageType) {
		events = append(events, "damage")
		if got != target || by != owner || source != caster || amount != 60 || typ != object.DamageImpact {
			t.Fatalf("damage=%p/%p/%p/%d/%d", got, by, source, amount, typ)
		}
	}
	h.castSound = func(id int32) sound.ID {
		events = append(events, "sound")
		if id != int32(spell.SPELL_EARTHQUAKE) {
			t.Fatal(id)
		}
		caster.PosVec = types.Ptf(70, 80)
		return 321
	}
	h.audio = func(id sound.ID, got *Object, kind int, code uint32) {
		events = append(events, "audio")
		if id != 321 || got != caster || kind != 0 || code != 0 {
			t.Fatal("sound callback changed")
		}
	}
	h.quake = func(got *Object, jiggle int32) {
		events = append(events, "quake")
		if got != caster || got.PosVec != types.Ptf(90, 100) || jiggle != 9 {
			t.Fatalf("quake source/value=%p/%v/%d", got, got.PosVec, jiggle)
		}
	}
	want := []string{"store", "range", "circle", "owner", "hp", "distance", "range", "load", "EarthquakeDamage", "damage", "sound", "audio", "EarthquakeJiggle", "quake"}
	if got := earthquakeCast52DE40(int32(spell.SPELL_EARTHQUAKE), caster, 3, h); got != 1 || !reflect.DeepEqual(events, want) || level != -123 || health.Cur != 0xffff || target.Field29 != 31 || caster.Field29 != 29 {
		t.Fatalf("result/order/cache/untouched=%d/%v/%d", got, events, level)
	}
}

func TestEarthquakeDamage52DEC0OriginalGatesAndLiveOwnerRead(t *testing.T) {
	for _, mode := range []string{"self", "self-nil", "airborne", "live-airborne", "zero-hp", "missing-health", "hp-high-word", "ordinary", "destroyed", "invulnerable", "same-team"} {
		t.Run(mode, func(t *testing.T) {
			level := int32(3)
			caster := &Object{}
			target := &Object{HealthData: &HealthData{Cur: 1}}
			if mode == "self" {
				target = caster
			}
			if mode == "self-nil" {
				target, caster = nil, nil
			}
			if mode == "airborne" {
				target.ObjFlags = object.FlagAirborne
			}
			if mode == "destroyed" {
				target.ObjFlags = object.FlagDestroyed
			}
			if mode == "invulnerable" {
				target.Buffs = 1 << ENCHANT_INVULNERABLE
			}
			if mode == "same-team" {
				target.TeamVal.ID, caster.TeamVal.ID = 7, 7
			}
			if mode == "zero-hp" {
				target.HealthData.Cur = 0
			}
			if mode == "missing-health" {
				target.HealthData = nil
			}
			if mode == "hp-high-word" {
				target.HealthData.Cur = 0x8000
			}
			h := earthquakeTestDeps52DE40(&level)
			ownerCalls, hpCalls, damageCalls := 0, 0, 0
			h.owner = func(got *Object) *Object {
				ownerCalls++
				if got != caster {
					t.Fatal("owner source changed")
				}
				if mode == "live-airborne" {
					target.ObjFlags |= object.FlagAirborne
				}
				return nil
			}
			h.hp = func(got *Object) uint16 { hpCalls++; return UnitGetHP4EE780(got) }
			h.distance = func(*Object, *Object) float64 { return 25 }
			h.damage = func(got, by, source *Object, amount int32, typ object.DamageType) {
				damageCalls++
				if got != target || by != nil || source != caster || amount != 30 || typ != 11 {
					t.Fatal("damage contract changed")
				}
			}
			earthquakeDamage52DEC0(target, caster, h)
			wantHP, wantDamage := 1, 1
			switch mode {
			case "self", "self-nil", "airborne", "live-airborne":
				wantHP, wantDamage = 0, 0
			case "zero-hp", "missing-health":
				wantDamage = 0
			}
			if ownerCalls != 1 || hpCalls != wantHP || damageCalls != wantDamage {
				t.Fatalf("owner/HP/damage=%d/%d/%d", ownerCalls, hpCalls, damageCalls)
			}
		})
	}
}

// Independent precision-53 big.Float instructions and exact rational spill
// comparison model the original arithmetic, including both FSTPS boundaries.
func earthquakeReferenceSpill52DEC0(value float64) float64 {
	narrow := float32(value)
	if math.IsInf(float64(narrow), 0) && !math.IsInf(value, 0) {
		return math.Copysign(math.MaxFloat32, value)
	}
	exact := new(big.Rat).SetFloat64(value)
	stored := new(big.Rat).SetFloat64(float64(narrow))
	if value > 0 && stored.Cmp(exact) > 0 || value < 0 && stored.Cmp(exact) < 0 {
		narrow = math.Nextafter32(narrow, 0)
	}
	return float64(narrow)
}

func earthquakeReferenceDamage52DEC0(distance float64, radius, base float32) int32 {
	operation := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	distance = earthquakeReferenceSpill52DEC0(distance)
	portion := operation().Quo(big.NewFloat(distance), big.NewFloat(float64(radius)))
	fraction := operation().Sub(big.NewFloat(1), portion)
	value, _ := fraction.Float64()
	falloff := earthquakeReferenceSpill52DEC0(value)
	product := operation().Mul(big.NewFloat(float64(base)), big.NewFloat(falloff))
	if product.Cmp(big.NewFloat(-0x1p63)) < 0 || product.Cmp(big.NewFloat(0x1p63)) >= 0 {
		return 0
	}
	integer, _ := product.Int(nil)
	integer.Mod(integer, new(big.Int).Lsh(big.NewInt(1), 32))
	return int32(uint32(integer.Uint64()))
}

func TestEarthquakeDamage52DEC0Precision53SpillsAndQwordLow(t *testing.T) {
	distances := []float64{0, float64(float32(0.01)), 0.1, 1 + 0x1p-25, 1 - 0x1p-25, 25, 99.999998, 100.000008, -0.1, 0x1p-150, 0x1p-149 * 1.5, 0x1p127}
	ranges := []float32{0.01, 1, 3, 100, 127, -3, math.SmallestNonzeroFloat32, math.MaxFloat32}
	bases := []float32{0, 1, -7, 40, 2147483648, 4294967296, 0x1p63, math.MaxFloat32}
	for di, distance := range distances {
		for ri, radius := range ranges {
			for bi, base := range bases {
				t.Run(fmt.Sprintf("d%d/r%d/b%d", di, ri, bi), func(t *testing.T) {
					level := int32(5)
					h := earthquakeTestDeps52DE40(&level)
					h.distance = func(*Object, *Object) float64 { return distance }
					h.balance = func(key string) float32 {
						if key != "EarthquakeRange" {
							t.Fatal(key)
						}
						return radius
					}
					h.indexed = func(key string, index int32) float32 {
						if key != "EarthquakeDamage" || index != 4 {
							t.Fatal(key, index)
						}
						return base
					}
					calls := 0
					h.damage = func(_, _, _ *Object, amount int32, typ object.DamageType) {
						calls++
						if want := earthquakeReferenceDamage52DEC0(distance, radius, base); amount != want || typ != 11 {
							t.Fatalf("damage=%d/%d want=%d distance=%x radius=%x base=%x", amount, typ, want, math.Float64bits(distance), math.Float32bits(radius), math.Float32bits(base))
						}
					}
					earthquakeDamage52DEC0(&Object{HealthData: &HealthData{Cur: 1}}, &Object{}, h)
					if calls != 1 {
						t.Fatal("zero/negative/overflow damage callback was suppressed")
					}
				})
			}
		}
	}
}

func TestEarthquakeDamage52DEC0NonfiniteStillCallsHandler(t *testing.T) {
	for _, tc := range []struct {
		name         string
		distance     float64
		radius, base float32
		want         int32
	}{
		{"nan-distance", math.NaN(), 100, 40, 0}, {"nan-range", 25, float32(math.NaN()), 40, 0},
		{"nan-base", 25, 100, float32(math.NaN()), 0}, {"infinite-distance", math.Inf(1), 100, 40, 0},
		{"infinite-base", 25, 100, float32(math.Inf(1)), 0}, {"zero-range", 25, 0, 40, 0},
		{"zero-zero-range", 0, 0, 40, 0}, {"infinite-range", 25, float32(math.Inf(1)), 40, 40},
	} {
		t.Run(tc.name, func(t *testing.T) {
			level := int32(1)
			h := earthquakeTestDeps52DE40(&level)
			h.distance = func(*Object, *Object) float64 { return tc.distance }
			h.balance = func(string) float32 { return tc.radius }
			h.indexed = func(string, int32) float32 { return tc.base }
			calls := 0
			h.damage = func(_, _, _ *Object, amount int32, typ object.DamageType) {
				calls++
				if amount != tc.want || typ != 11 {
					t.Fatalf("damage=%d/%d want=%d", amount, typ, tc.want)
				}
			}
			earthquakeDamage52DEC0(&Object{HealthData: &HealthData{Cur: 1}}, &Object{}, h)
			if calls != 1 {
				t.Fatal("nonfinite result suppressed callback")
			}
		})
	}
}

func TestEarthquakeCast52DE40DWORDLevelsAndJiggle(t *testing.T) {
	for _, power := range []int32{1, 2, 3, 4, 5, 0, -129, 128, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(power), func(t *testing.T) {
			var level int32
			h := earthquakeTestDeps52DE40(&level)
			h.indexed = func(key string, index int32) float32 {
				if key != "EarthquakeJiggle" || index != power-1 {
					t.Fatalf("table index=%s/%d", key, index)
				}
				return -9.75
			}
			calls := 0
			h.quake = func(caster *Object, jiggle int32) {
				calls++
				if caster == nil || jiggle != -9 {
					t.Fatal("jiggle changed")
				}
			}
			if got := earthquakeCast52DE40(23, &Object{}, power, h); got != 1 || level != power || calls != 1 {
				t.Fatal(got, level, calls)
			}
		})
	}
	for _, tc := range []struct {
		name  string
		input float32
		want  int32
	}{
		{"positive", 2.75, 2}, {"negative", -2.75, -2}, {"zero", 0, 0}, {"negative-zero", float32(math.Copysign(0, -1)), 0},
		{"max-finite-int", math.Nextafter32(2147483648, 0), 2147483520}, {"minimum", -2147483648, math.MinInt32},
		{"positive-overflow", 2147483648, math.MinInt32}, {"negative-overflow", math.Nextafter32(-2147483648, float32(math.Inf(-1))), math.MinInt32},
		{"nan", float32(math.NaN()), math.MinInt32}, {"infinity", float32(math.Inf(1)), math.MinInt32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := earthquakeJiggle52DE40(tc.input); got != tc.want {
				t.Fatalf("FISTP=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestEarthquakeCast52DE40AllFaultPrefixes(t *testing.T) {
	want := []string{"store", "range", "circle", "sound", "audio", "indexed", "quake"}
	for stop := range want {
		t.Run(want[stop], func(t *testing.T) {
			var events []string
			step := func(name string) {
				events = append(events, name)
				if len(events) == stop+1 {
					panic("injected service fault")
				}
			}
			h := earthquakeCastDeps52DE40{
				storeLevel: func(int32) { step("store") }, balance: func(string) float32 { step("range"); return 100 },
				circle: func(*Object, float32, func(*Object)) { step("circle") }, castSound: func(int32) sound.ID { step("sound"); return 1 },
				audio: func(sound.ID, *Object, int, uint32) { step("audio") }, indexed: func(string, int32) float32 { step("indexed"); return 2 }, quake: func(*Object, int32) { step("quake") },
			}
			var fault any
			func() { defer func() { fault = recover() }(); earthquakeCast52DE40(23, &Object{}, 1, h) }()
			if fault == nil || !reflect.DeepEqual(events, want[:stop+1]) {
				t.Fatalf("fault/prefix=%v/%v", fault, events)
			}
		})
	}
	// The original LEA does not read the caster before balance lookup.
	level := int32(0)
	h := earthquakeTestDeps52DE40(&level)
	calls := 0
	h.balance = func(string) float32 { calls++; return 100 }
	h.circle = func(caster *Object, _ float32, _ func(*Object)) { _ = caster.PosVec }
	var fault any
	func() { defer func() { fault = recover() }(); earthquakeCast52DE40(23, nil, 3, h) }()
	if fault == nil || calls != 1 || level != 3 {
		t.Fatal("nil caster fault moved before cache/balance", fault, calls, level)
	}
}

func TestEarthquakeDamage52DEC0AllFaultPrefixes(t *testing.T) {
	want := []string{"owner", "hp", "distance", "range", "load", "indexed", "damage"}
	for stop := range want {
		t.Run(want[stop], func(t *testing.T) {
			var events []string
			step := func(name string) {
				events = append(events, name)
				if len(events) == stop+1 {
					panic("injected callback fault")
				}
			}
			h := earthquakeCastDeps52DE40{
				owner: func(*Object) *Object { step("owner"); return nil }, hp: func(*Object) uint16 { step("hp"); return 1 }, distance: func(*Object, *Object) float64 { step("distance"); return 25 },
				balance: func(string) float32 { step("range"); return 100 }, loadLevel: func() int32 { step("load"); return 3 }, indexed: func(string, int32) float32 { step("indexed"); return 40 },
				damage: func(*Object, *Object, *Object, int32, object.DamageType) { step("damage") },
			}
			var fault any
			func() { defer func() { fault = recover() }(); earthquakeDamage52DEC0(&Object{}, &Object{}, h) }()
			if fault == nil || !reflect.DeepEqual(events, want[:stop+1]) {
				t.Fatalf("fault/prefix=%v/%v", fault, events)
			}
		})
	}
	ownerCalls := 0
	h := earthquakeCastDeps52DE40{owner: func(*Object) *Object { ownerCalls++; return nil }}
	var fault any
	func() { defer func() { fault = recover() }(); earthquakeDamage52DEC0(nil, &Object{}, h) }()
	if fault == nil || ownerCalls != 1 {
		t.Fatal("nil target suppressed/moved original flags fault")
	}
}

func TestEarthquakeCast52DE40NativeMapAndDamageRegistry(t *testing.T) {
	s := new(Server)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	s.Balance.file = &balance.File{Global: balance.Config{
		"earthquakerange": balance.Float(100), "earthquakedamage": balance.Array{20, 30, 40, 50, 60}, "earthquakejiggle": balance.Array{5, 6, 7, 8, 9},
	}}
	level := uint32(0)
	owner := earthquakeAlloc52DE40(t, Object{ObjClass: object.ClassPlayer})
	caster := earthquakeAlloc52DE40(t, Object{ObjOwner: owner, PosVec: types.Ptf(400, 400), NewPos: types.Ptf(400, 400), ObjFlags: object.FlagActive, ObjClass: object.ClassMonster})
	targetHP := earthquakeAlloc52DE40(t, HealthData{Cur: 0xffff, Max: 0xffff})
	callback := earthquakeAlloc52DE40(t, byte(1))
	target := earthquakeAlloc52DE40(t, Object{HealthData: targetHP, PosVec: types.Ptf(450, 400), NewPos: types.Ptf(450, 400), ObjFlags: object.FlagActive, ObjClass: object.ClassMonster, Damage: unsafe.Pointer(callback)})
	caster.Shape.Kind, target.Shape.Kind = ShapeKindCircle, ShapeKindCircle
	caster.Shape.Circle.R, target.Shape.Circle.R = 10, 5 // Surface distance=35, not center distance=50.
	airborne := earthquakeAlloc52DE40(t, Object{HealthData: targetHP, PosVec: types.Ptf(420, 400), NewPos: types.Ptf(420, 400), ObjFlags: object.FlagActive | object.FlagAirborne, ObjClass: object.ClassMonster})
	dead := earthquakeAlloc52DE40(t, Object{HealthData: earthquakeAlloc52DE40(t, HealthData{}), PosVec: types.Ptf(430, 400), NewPos: types.Ptf(430, 400), ObjFlags: object.FlagActive, ObjClass: object.ClassMonster})
	outside := earthquakeAlloc52DE40(t, Object{HealthData: targetHP, PosVec: types.Ptf(501, 400), NewPos: types.Ptf(501, 400), ObjFlags: object.FlagActive, ObjClass: object.ClassMonster})
	for _, unit := range []*Object{caster, target, airborne, dead, outside} {
		s.Map.AddObjectToIndex(unit)
	}
	damageCalls := 0
	previousRegistry := objDamage
	objDamage = ccall.NewFuncs(func(ptr unsafe.Pointer) DamageFunc { return previousRegistry.Get(ptr) })
	t.Cleanup(func() { objDamage = previousRegistry })
	objDamage.Register(unsafe.Pointer(callback), func(got, by, source *Object, amount int32, typ object.DamageType) bool {
		damageCalls++
		want := earthquakeReferenceDamage52DEC0(35, 100, 40)
		if got != target || by != owner || source != caster || amount != want || typ != 11 {
			t.Fatalf("native damage=%p/%p/%p/%d/%d want=%d", got, by, source, amount, typ, want)
		}
		got.HealthData.Cur -= uint16(amount)
		return false // A false damage return does not stop enumeration or cast.
	})
	// Other three native pointer arguments are genuinely unused, even nil.
	if got := s.CastEarthquake52DE40(int32(spell.SPELL_EARTHQUAKE), nil, nil, caster, nil, 3, EarthquakeCastRuntime52DE40{LevelCache: &level}); got != 1 || level != 3 || damageCalls != 1 || targetHP.Cur != 0xffff-uint16(earthquakeReferenceDamage52DEC0(35, 100, 40)) {
		t.Fatalf("native result/level/calls/HP=%d/%d/%d/%d", got, level, damageCalls, targetHP.Cur)
	}
	if len(s.Audio.delayedObj) != 1 || s.Audio.delayedObj[0].Obj != caster {
		t.Fatal("native sound omitted or source narrowed")
	}
}

func TestEarthquakeDamage52DEC0NativeLiveDamageSlotAndMissingCallback(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(fmt.Sprint(missing), func(t *testing.T) {
			s := new(Server)
			level := uint32(1)
			caster := earthquakeAlloc52DE40(t, Object{})
			target := earthquakeAlloc52DE40(t, Object{HealthData: earthquakeAlloc52DE40(t, HealthData{Cur: 1})})
			first, live := earthquakeAlloc52DE40(t, byte(1)), earthquakeAlloc52DE40(t, byte(2))
			target.Damage = unsafe.Pointer(first)
			previousRegistry := objDamage
			objDamage = ccall.NewFuncs(func(ptr unsafe.Pointer) DamageFunc { return previousRegistry.Get(ptr) })
			t.Cleanup(func() { objDamage = previousRegistry })
			objDamage.Register(unsafe.Pointer(first), func(*Object, *Object, *Object, int32, object.DamageType) bool {
				t.Fatal("stale callback slot used")
				return true
			})
			calls := 0
			objDamage.Register(unsafe.Pointer(live), func(got, by, source *Object, amount int32, typ object.DamageType) bool {
				calls++
				if got != target || by != caster || source != caster || amount != 40 || typ != 11 {
					t.Fatal("native live callback args changed")
				}
				return false
			})
			h := s.earthquakeCastDeps52DE40(EarthquakeCastRuntime52DE40{LevelCache: &level})
			h.balance = func(string) float32 { return 100 }
			h.indexed = func(string, int32) float32 {
				target.Damage = unsafe.Pointer(live)
				if missing {
					target.Damage = nil
				}
				return 40
			}
			h.distance = func(*Object, *Object) float64 { return 0 }
			var fault any
			func() { defer func() { fault = recover() }(); earthquakeDamage52DEC0(target, caster, h) }()
			if missing && (fault == nil || calls != 0) || !missing && (fault != nil || calls != 1) {
				t.Fatalf("callback fault/calls=%v/%d", fault, calls)
			}
		})
	}
}
