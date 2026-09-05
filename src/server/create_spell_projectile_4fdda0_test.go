package server

import (
	"fmt"
	"image"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

const (
	createSpellProjectileTestNil4FDDA0        = 0
	createSpellProjectileTestSource4FDDA0     = 1
	createSpellProjectileTestTarget4FDDA0     = 2
	createSpellProjectileTestProjectile4FDDA0 = 3
	createSpellProjectileTestUpdate4FDDA0     = 10
	createSpellProjectileTestPlayerX4FDDA0    = 20
	createSpellProjectileTestPlayerY4FDDA0    = 21
)

type createSpellProjectileTestWorld4FDDA0 struct {
	t              *testing.T
	sourceArg      int
	targetArg      int
	spellArg       int32
	radius         float32
	classLow       uint8
	searchResult   int
	traceResult    int32
	newResult      int
	hasEnchant     int32
	directionLoads map[int]int
	posXLoads      int
	posYLoads      int
	velXLoads      map[int]int
	velYLoads      map[int]int
	speedLoads     int
	velX           map[int]float32
	velY           map[int]float32
	projectileDir  uint16
	events         []string
	faultAt        int
	aim            *types.Pointf
	ray            *createSpellProjectileRay4FDDA0
}

func newCreateSpellProjectileTestWorld4FDDA0(t *testing.T) *createSpellProjectileTestWorld4FDDA0 {
	return &createSpellProjectileTestWorld4FDDA0{
		t:              t,
		sourceArg:      createSpellProjectileTestSource4FDDA0,
		targetArg:      createSpellProjectileTestNil4FDDA0,
		spellArg:       math.MinInt32,
		radius:         6,
		classLow:       createSpellProjectilePlayerClass4FDDA0,
		searchResult:   createSpellProjectileTestTarget4FDDA0,
		traceResult:    math.MinInt32,
		newResult:      createSpellProjectileTestProjectile4FDDA0,
		hasEnchant:     math.MinInt32,
		directionLoads: make(map[int]int),
		velXLoads:      make(map[int]int),
		velYLoads:      make(map[int]int),
		velX:           make(map[int]float32),
		velY:           make(map[int]float32),
	}
}

func (w *createSpellProjectileTestWorld4FDDA0) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *createSpellProjectileTestWorld4FDDA0) hooks() createSpellProjectileHooks4FDDA0[int, int, int] {
	return createSpellProjectileHooks4FDDA0[int, int, int]{
		loadSourceArg: func() int {
			w.event("source-arg")
			return w.sourceArg
		},
		loadTargetArg: func() int {
			w.event("target-arg")
			return w.targetArg
		},
		loadSpellArg: func() int32 {
			w.event("spell-arg")
			return w.spellArg
		},
		loadRadius: func(obj int) float32 {
			w.event("radius")
			if obj != createSpellProjectileTestSource4FDDA0 {
				w.t.Fatalf("radius object = %d", obj)
			}
			return w.radius
		},
		loadClassLow: func(obj int) uint8 {
			w.event("class")
			if obj != createSpellProjectileTestSource4FDDA0 {
				w.t.Fatalf("class object = %d", obj)
			}
			return w.classLow
		},
		loadUpdate: func(obj int) int {
			w.event("source-update")
			if obj != createSpellProjectileTestSource4FDDA0 {
				w.t.Fatalf("source update object = %d", obj)
			}
			return createSpellProjectileTestUpdate4FDDA0
		},
		loadPlayer: func(update int) int {
			w.event("player")
			if update != createSpellProjectileTestUpdate4FDDA0 {
				w.t.Fatalf("player update = %d", update)
			}
			if len(w.events) != 0 && w.events[len(w.events)-2] == "cursor-x" {
				return createSpellProjectileTestPlayerY4FDDA0
			}
			return createSpellProjectileTestPlayerX4FDDA0
		},
		loadCursorX: func(player int) int32 {
			w.event("cursor-x")
			if player != createSpellProjectileTestPlayerX4FDDA0 {
				w.t.Fatalf("cursor X player = %d", player)
			}
			return 16_777_217
		},
		loadCursorY: func(player int) int32 {
			w.event("cursor-y")
			if player != createSpellProjectileTestPlayerY4FDDA0 {
				w.t.Fatalf("cursor Y player = %d", player)
			}
			return -16_777_217
		},
		spellFlags: func(id int32) uint32 {
			w.event("spell-flags")
			if id != math.MinInt32 {
				w.t.Fatalf("flags spell = %d", id)
			}
			return 0xf1234567
		},
		searchTarget: func(aim *types.Pointf, source int, flags uint32, distance float32, mode int32, self int) int {
			w.event("search-target")
			if aim == nil || math.Float32bits(aim.X) != math.Float32bits(16_777_216) || math.Float32bits(aim.Y) != math.Float32bits(-16_777_216) {
				w.t.Fatalf("search aim = %v", aim)
			}
			if source != createSpellProjectileTestSource4FDDA0 || self != source || flags != 0xf1234567 || distance != 600 || mode != 0 {
				w.t.Fatalf("search args = %d/%08x/%v/%d/%d", source, flags, distance, mode, self)
			}
			w.aim = aim
			aim.X, aim.Y = 123, 456
			return w.searchResult
		},
		loadDirection: func(obj int) int16 {
			w.event(fmt.Sprintf("direction:%d", obj))
			w.directionLoads[obj]++
			switch obj {
			case createSpellProjectileTestSource4FDDA0:
				switch w.directionLoads[obj] {
				case 1:
					return 5
				case 2:
					return int16(-32477) // raw bits 0x8123
				default:
					return 9
				}
			case createSpellProjectileTestProjectile4FDDA0:
				if w.directionLoads[obj] == 1 {
					return 11
				}
				return 12
			default:
				w.t.Fatalf("direction object = %d", obj)
				return 0
			}
		},
		loadPosX: func(obj int) float32 {
			w.event("pos-x")
			if obj != createSpellProjectileTestSource4FDDA0 {
				w.t.Fatalf("position X object = %d", obj)
			}
			w.posXLoads++
			if w.posXLoads == 1 {
				return 10
			}
			return 20
		},
		loadPosY: func(obj int) float32 {
			w.event("pos-y")
			if obj != createSpellProjectileTestSource4FDDA0 {
				w.t.Fatalf("position Y object = %d", obj)
			}
			w.posYLoads++
			if w.posYLoads == 1 {
				return 30
			}
			return 40
		},
		directionX: func(direction int16) float32 {
			w.event(fmt.Sprintf("direction-x:%d", direction))
			switch direction {
			case 5:
				return 0.25
			case 11:
				return 2
			default:
				w.t.Fatalf("direction X index = %d", direction)
				return 0
			}
		},
		directionY: func(direction int16) float32 {
			w.event(fmt.Sprintf("direction-y:%d", direction))
			switch direction {
			case 5:
				return 0.5
			case 12:
				return -3
			default:
				w.t.Fatalf("direction Y index = %d", direction)
				return 0
			}
		},
		loadVelX: func(obj int) float32 {
			w.event(fmt.Sprintf("vel-x:%d", obj))
			w.velXLoads[obj]++
			if obj == createSpellProjectileTestSource4FDDA0 {
				if w.velXLoads[obj] == 1 {
					return 1
				}
				return 100
			}
			return w.velX[obj]
		},
		loadVelY: func(obj int) float32 {
			w.event(fmt.Sprintf("vel-y:%d", obj))
			w.velYLoads[obj]++
			if obj == createSpellProjectileTestSource4FDDA0 {
				if w.velYLoads[obj] == 1 {
					return 2
				}
				return 200
			}
			return w.velY[obj]
		},
		mapTrace: func(ray *createSpellProjectileRay4FDDA0, point *types.Pointf, grid *image.Point, flags int32) int32 {
			w.event("map-trace")
			if point != nil || grid != nil || flags != 5 {
				w.t.Fatalf("trace extra args = %p/%p/%d", point, grid, flags)
			}
			if ray.Origin != (types.Pointf{X: 10, Y: 30}) || ray.Destination != (types.Pointf{X: 23.5, Y: 47}) {
				w.t.Fatalf("ray = %+v", *ray)
			}
			w.ray = ray
			ray.Destination = types.Pointf{X: -12.5, Y: 91.25}
			return w.traceResult
		},
		newObject: func(id string) int {
			w.event("new-object")
			if id != "Magic" {
				w.t.Fatalf("object type = %q", id)
			}
			return w.newResult
		},
		loadProjectileUpdate: func(obj int) int {
			w.event("projectile-update")
			if obj != createSpellProjectileTestProjectile4FDDA0 {
				w.t.Fatalf("projectile update object = %d", obj)
			}
			return createSpellProjectileTestUpdate4FDDA0
		},
		spellPower: func(id int32, source int) int32 {
			w.event("spell-power")
			if id != math.MinInt32 || source != createSpellProjectileTestSource4FDDA0 {
				w.t.Fatalf("power args = %d/%d", id, source)
			}
			return math.MinInt32
		},
		storeLevel: func(update int, value uint32) {
			w.event("store-level")
			if update != createSpellProjectileTestUpdate4FDDA0 || value != uint32(0x80000000) {
				w.t.Fatalf("level store = %d/%08x", update, value)
			}
		},
		createAt: func(projectile, owner int, position types.Pointf, reserved int32) {
			w.event("create-at")
			if projectile != createSpellProjectileTestProjectile4FDDA0 || owner != createSpellProjectileTestSource4FDDA0 || position != (types.Pointf{X: -12.5, Y: 91.25}) || reserved != 0 {
				w.t.Fatalf("create args = %d/%d/%v/%d", projectile, owner, position, reserved)
			}
		},
		storeDirection1: func(projectile int, value uint16) {
			w.event("store-direction-1")
			if projectile != createSpellProjectileTestProjectile4FDDA0 || value != 0x8123 {
				w.t.Fatalf("direction 1 store = %d/%04x", projectile, value)
			}
			w.projectileDir = value
		},
		storeDirection2: func(projectile int, value uint16) {
			w.event("store-direction-2")
			if projectile != createSpellProjectileTestProjectile4FDDA0 || value != 0x8123 {
				w.t.Fatalf("direction 2 store = %d/%04x", projectile, value)
			}
		},
		storeField0: func(update, value int) {
			w.event("store-field-0")
			if update != createSpellProjectileTestUpdate4FDDA0 || value != createSpellProjectileTestSource4FDDA0 {
				w.t.Fatalf("field 0 store = %d/%d", update, value)
			}
		},
		storeTarget: func(update, value int) {
			w.event("store-target")
			if update != createSpellProjectileTestUpdate4FDDA0 || value != createSpellProjectileTestTarget4FDDA0 {
				w.t.Fatalf("target store = %d/%d", update, value)
			}
		},
		storeField8: func(update, value int) {
			w.event("store-field-8")
			if update != createSpellProjectileTestUpdate4FDDA0 || value != createSpellProjectileTestSource4FDDA0 {
				w.t.Fatalf("field 8 store = %d/%d", update, value)
			}
		},
		storeSpell: func(update int, value uint32) {
			w.event("store-spell")
			if update != createSpellProjectileTestUpdate4FDDA0 || value != uint32(0x80000000) {
				w.t.Fatalf("spell store = %d/%08x", update, value)
			}
		},
		indexedDirection: func(direction int16, scratch *types.Pointf) {
			w.event("indexed-direction")
			if direction != 9 || scratch != w.aim || *scratch != (types.Pointf{X: 123, Y: 456}) {
				w.t.Fatalf("indexed direction = %d/%p/%v, aim %p", direction, scratch, *scratch, w.aim)
			}
		},
		loadSpeed: func(projectile int) float32 {
			w.event("speed")
			if projectile != createSpellProjectileTestProjectile4FDDA0 {
				w.t.Fatalf("speed object = %d", projectile)
			}
			w.speedLoads++
			if w.speedLoads == 1 {
				return 4
			}
			return 5
		},
		storeVelX: func(projectile int, value float32) {
			w.event(fmt.Sprintf("store-vel-x:%g", value))
			if projectile != createSpellProjectileTestProjectile4FDDA0 {
				w.t.Fatalf("velocity X object = %d", projectile)
			}
			w.velX[projectile] = value
		},
		storeVelY: func(projectile int, value float32) {
			w.event(fmt.Sprintf("store-vel-y:%g", value))
			if projectile != createSpellProjectileTestProjectile4FDDA0 {
				w.t.Fatalf("velocity Y object = %d", projectile)
			}
			w.velY[projectile] = value
		},
		hasEnchant: func(source int, enchant int32) int32 {
			w.event("has-enchant")
			if source != createSpellProjectileTestSource4FDDA0 || enchant != 21 {
				w.t.Fatalf("has enchant args = %d/%d", source, enchant)
			}
			return w.hasEnchant
		},
		enchantPower: func(source int, enchant int32) int32 {
			w.event("enchant-power")
			if source != createSpellProjectileTestSource4FDDA0 || enchant != 21 {
				w.t.Fatalf("enchant power args = %d/%d", source, enchant)
			}
			return 0x123
		},
		enchantTimer: func(source int, enchant int32) int32 {
			w.event("enchant-timer")
			if source != createSpellProjectileTestSource4FDDA0 || enchant != 21 {
				w.t.Fatalf("enchant timer args = %d/%d", source, enchant)
			}
			return 0x12345
		},
		applyEnchant: func(projectile int, enchant int32, duration int16, power uint8) {
			w.event("apply-enchant")
			if projectile != createSpellProjectileTestProjectile4FDDA0 || enchant != 21 || duration != 0x2345 || power != 0x23 {
				w.t.Fatalf("apply enchant args = %d/%d/%04x/%02x", projectile, enchant, uint16(duration), power)
			}
		},
		spellAudio: func(id, field int32) int32 {
			w.event("spell-audio")
			if id != math.MinInt32 || field != 0 {
				w.t.Fatalf("spell audio args = %d/%d", id, field)
			}
			return -77
		},
		audio: func(id int32, source int, kind int32, code uint32) {
			w.event("audio")
			if id != -77 || source != createSpellProjectileTestSource4FDDA0 || kind != 0 || code != 0 {
				w.t.Fatalf("audio args = %d/%d/%d/%d", id, source, kind, code)
			}
		},
	}
}

var createSpellProjectileExactEvents4FDDA0 = []string{
	"source-arg", "target-arg", "spell-arg", "radius",
	"class", "source-update", "player", "cursor-x", "player", "cursor-y",
	"spell-flags", "search-target",
	"direction:1", "pos-x", "pos-y", "direction-x:5", "pos-x", "direction-y:5", "pos-y", "vel-x:1", "vel-y:1",
	"map-trace", "new-object", "projectile-update", "spell-power", "store-level", "create-at",
	"direction:1", "store-direction-1", "store-direction-2",
	"store-field-0", "store-target", "store-field-8", "store-spell",
	"direction:1", "indexed-direction", "direction:3", "direction:3",
	"direction-x:11", "speed", "store-vel-x:8",
	"direction-y:12", "speed", "store-vel-y:-15",
	"vel-x:3", "vel-x:1", "store-vel-x:108",
	"vel-y:3", "vel-y:1", "store-vel-y:185",
	"has-enchant", "enchant-power", "enchant-timer", "apply-enchant",
	"spell-audio", "audio",
}

func TestCreateSpellProjectile4FDDA0ExactSuccessOrderAndWidths(t *testing.T) {
	w := newCreateSpellProjectileTestWorld4FDDA0(t)
	if got := createSpellProjectile4FDDA0(w.hooks()); got != createSpellProjectileTestProjectile4FDDA0 {
		t.Fatalf("result = %d", got)
	}
	if !reflect.DeepEqual(w.events, createSpellProjectileExactEvents4FDDA0) {
		t.Fatalf("events:\n got  %v\n want %v", w.events, createSpellProjectileExactEvents4FDDA0)
	}
	if w.velX[createSpellProjectileTestProjectile4FDDA0] != 108 || w.velY[createSpellProjectileTestProjectile4FDDA0] != 185 {
		t.Fatalf("projectile velocity = %v/%v", w.velX[3], w.velY[3])
	}
	if w.projectileDir != 0x8123 {
		t.Fatalf("projectile direction bits = %04x", w.projectileDir)
	}
}

func TestCreateSpellProjectile4FDDA0ExactFaultPrefixes(t *testing.T) {
	for faultAt := 1; faultAt <= len(createSpellProjectileExactEvents4FDDA0); faultAt++ {
		t.Run(fmt.Sprintf("%02d_%s", faultAt, createSpellProjectileExactEvents4FDDA0[faultAt-1]), func(t *testing.T) {
			w := newCreateSpellProjectileTestWorld4FDDA0(t)
			w.faultAt = faultAt
			panicked := false
			func() {
				defer func() { panicked = recover() != nil }()
				_ = createSpellProjectile4FDDA0(w.hooks())
			}()
			if !panicked {
				t.Fatal("expected injected fault")
			}
			want := createSpellProjectileExactEvents4FDDA0[:faultAt]
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("fault prefix:\n got  %v\n want %v", w.events, want)
			}
		})
	}
}

func TestCreateSpellProjectile4FDDA0SuppliedTargetSkipsSearch(t *testing.T) {
	w := newCreateSpellProjectileTestWorld4FDDA0(t)
	w.targetArg = createSpellProjectileTestTarget4FDDA0
	hooks := w.hooks()
	hooks.loadClassLow = func(int) uint8 { t.Fatal("supplied target loaded class"); return 0 }
	hooks.spellFlags = func(int32) uint32 { t.Fatal("supplied target loaded flags"); return 0 }
	hooks.searchTarget = func(*types.Pointf, int, uint32, float32, int32, int) int {
		t.Fatal("supplied target performed search")
		return 0
	}
	hooks.indexedDirection = func(direction int16, scratch *types.Pointf) {
		w.event("indexed-direction")
		if direction != 9 || scratch == nil || *scratch != (types.Pointf{}) {
			t.Fatalf("non-player scratch = %d/%p/%v", direction, scratch, *scratch)
		}
	}
	if got := createSpellProjectile4FDDA0(hooks); got != createSpellProjectileTestProjectile4FDDA0 {
		t.Fatalf("result = %d", got)
	}
	for _, event := range w.events {
		if event == "class" || event == "spell-flags" || event == "search-target" {
			t.Fatalf("unexpected event %q", event)
		}
	}
}

func TestCreateSpellProjectile4FDDA0NonPlayerSearchUsesNilAim(t *testing.T) {
	w := newCreateSpellProjectileTestWorld4FDDA0(t)
	w.classLow = 0xfb
	hooks := w.hooks()
	hooks.loadUpdate = func(int) int { t.Fatal("non-player loaded update"); return 0 }
	hooks.searchTarget = func(aim *types.Pointf, source int, flags uint32, distance float32, mode int32, self int) int {
		w.event("search-target")
		if aim != nil || source != 1 || flags != 0xf1234567 || distance != 600 || mode != 0 || self != 1 {
			t.Fatalf("search args = %p/%d/%08x/%v/%d/%d", aim, source, flags, distance, mode, self)
		}
		return createSpellProjectileTestTarget4FDDA0
	}
	hooks.indexedDirection = func(direction int16, scratch *types.Pointf) {
		w.event("indexed-direction")
		if direction != 9 || scratch == nil || *scratch != (types.Pointf{}) {
			t.Fatalf("scratch = %d/%p/%v", direction, scratch, *scratch)
		}
	}
	if got := createSpellProjectile4FDDA0(hooks); got != createSpellProjectileTestProjectile4FDDA0 {
		t.Fatalf("result = %d", got)
	}
}

func TestCreateSpellProjectile4FDDA0NilExitsAndTraceTruthiness(t *testing.T) {
	t.Run("blocked ray", func(t *testing.T) {
		w := newCreateSpellProjectileTestWorld4FDDA0(t)
		w.traceResult = 0
		hooks := w.hooks()
		hooks.newObject = func(string) int { t.Fatal("blocked ray created object"); return 0 }
		if got := createSpellProjectile4FDDA0(hooks); got != 0 {
			t.Fatalf("result = %d", got)
		}
	})

	t.Run("nil factory", func(t *testing.T) {
		w := newCreateSpellProjectileTestWorld4FDDA0(t)
		w.newResult = 0
		hooks := w.hooks()
		hooks.loadProjectileUpdate = func(int) int { t.Fatal("nil projectile loaded update"); return 0 }
		if got := createSpellProjectile4FDDA0(hooks); got != 0 {
			t.Fatalf("result = %d", got)
		}
	})

	for _, value := range []int32{-1, 1, 2, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("trace_%d", value), func(t *testing.T) {
			w := newCreateSpellProjectileTestWorld4FDDA0(t)
			w.traceResult = value
			if got := createSpellProjectile4FDDA0(w.hooks()); got != createSpellProjectileTestProjectile4FDDA0 {
				t.Fatalf("trace value %d result = %d", value, got)
			}
		})
	}
}

func TestCreateSpellProjectile4FDDA0EveryNonzeroEnchantPredicateIsTrue(t *testing.T) {
	for _, value := range []int32{-1, 1, 2, math.MinInt32, math.MaxInt32} {
		w := newCreateSpellProjectileTestWorld4FDDA0(t)
		w.hasEnchant = value
		if got := createSpellProjectile4FDDA0(w.hooks()); got != createSpellProjectileTestProjectile4FDDA0 {
			t.Fatalf("predicate %d result = %d", value, got)
		}
		if !containsCreateSpellProjectileEvent4FDDA0(w.events, "apply-enchant") {
			t.Fatalf("predicate %d skipped enchant", value)
		}
	}
	w := newCreateSpellProjectileTestWorld4FDDA0(t)
	w.hasEnchant = 0
	hooks := w.hooks()
	hooks.enchantPower = func(int, int32) int32 { t.Fatal("absent enchant loaded power"); return 0 }
	if got := createSpellProjectile4FDDA0(hooks); got != createSpellProjectileTestProjectile4FDDA0 {
		t.Fatalf("zero predicate result = %d", got)
	}
}

func TestCreateSpellProjectile4FDDA0PreservesAsymmetricX87Spills(t *testing.T) {
	t.Run("X remains binary64 through velocity add", func(t *testing.T) {
		w := newCreateSpellProjectileTestWorld4FDDA0(t)
		w.targetArg = createSpellProjectileTestTarget4FDDA0
		hooks := w.hooks()
		hooks.loadRadius = func(int) float32 { return math.Float32frombits(0xcaf25060) }
		hooks.loadDirection = func(int) int16 { return 0 }
		hooks.directionX = func(int16) float32 { return math.Float32frombits(0x3ef5d338) }
		hooks.directionY = func(int16) float32 { return 0 }
		xLoads := 0
		hooks.loadPosX = func(int) float32 {
			xLoads++
			if xLoads == 1 {
				return 0
			}
			return math.Float32frombits(0x4ca8d37c)
		}
		hooks.loadPosY = func(int) float32 { return 0 }
		hooks.loadVelX = func(int) float32 { return math.Float32frombits(0xc2701db9) }
		hooks.loadVelY = func(int) float32 { return 0 }
		hooks.mapTrace = func(ray *createSpellProjectileRay4FDDA0, _ *types.Pointf, _ *image.Point, _ int32) int32 {
			if got := math.Float32bits(ray.Destination.X); got != 0x4ca18dfe {
				t.Fatalf("destination X bits = %08x, want x87 %08x (premature spill would be %08x)", got, uint32(0x4ca18dfe), uint32(0x4ca18dfd))
			}
			return 0
		}
		if got := createSpellProjectile4FDDA0(hooks); got != 0 {
			t.Fatalf("blocked result = %d", got)
		}
	})

	t.Run("Y spills before velocity add", func(t *testing.T) {
		w := newCreateSpellProjectileTestWorld4FDDA0(t)
		w.targetArg = createSpellProjectileTestTarget4FDDA0
		hooks := w.hooks()
		hooks.loadRadius = func(int) float32 { return math.Float32frombits(0x4b441048) }
		hooks.loadDirection = func(int) int16 { return 0 }
		hooks.directionX = func(int16) float32 { return 0 }
		hooks.directionY = func(int16) float32 { return math.Float32frombits(0x3ffe58de) }
		hooks.loadPosX = func(int) float32 { return 0 }
		yLoads := 0
		hooks.loadPosY = func(int) float32 {
			yLoads++
			if yLoads == 1 {
				return 0
			}
			return math.Float32frombits(0xcc7e44af)
		}
		hooks.loadVelX = func(int) float32 { return 0 }
		hooks.loadVelY = func(int) float32 { return math.Float32frombits(0x42c665c6) }
		hooks.mapTrace = func(ray *createSpellProjectileRay4FDDA0, _ *types.Pointf, _ *image.Point, _ int32) int32 {
			if got := math.Float32bits(ray.Destination.Y); got != 0xcc1cde78 {
				t.Fatalf("destination Y bits = %08x, want x87 spill %08x (unspilled would be %08x)", got, uint32(0xcc1cde78), uint32(0xcc1cde79))
			}
			return 0
		}
		if got := createSpellProjectile4FDDA0(hooks); got != 0 {
			t.Fatalf("blocked result = %d", got)
		}
	})
}

func containsCreateSpellProjectileEvent4FDDA0(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}
