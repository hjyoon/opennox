package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type defaultDropTestUpdate4ED290 struct {
	name   string
	byte2  uint8
	frame  uint32
	action uint32
	status uint32
}

type defaultDropTestObject4ED290 struct {
	name    string
	class   uint32
	flags   uint32
	netCode uint32
	teamID  uint8
	typeInd uint16
	holder  *defaultDropTestObject4ED290
	head    *defaultDropTestObject4ED290
	next    *defaultDropTestObject4ED290
	update  *defaultDropTestUpdate4ED290
}

type defaultDropTestPoint4ED290 struct {
	x float32
	y float32
}

type defaultDropTestWorld4ED290 struct {
	owner *defaultDropTestObject4ED290
	item  *defaultDropTestObject4ED290
	point *defaultDropTestPoint4ED290

	events  []string
	faultAt int

	droppable int32
	dropMask  int32
	material  uint32
	frame     uint32
	fps       uint32

	glyphCache   uint32
	torchCache   uint32
	lanternCache uint32
	lookup       map[string]uint32
	gameFlags    map[uint32]uint32
	weaponFlags  map[*defaultDropTestObject4ED290]uint32

	onCreate   func()
	onWeapon   func(*defaultDropTestObject4ED290)
	onMaterial func()
	onInform   func()
	onMinimap  func()
	onRaise    func()
	onAudio    func(uint32)

	informCode     uint8
	informNetCode  uint32
	informMaterial uint32
	statusTeam     uint8
	statusMaterial uint8
}

func defaultDropObjectName4ED290(obj *defaultDropTestObject4ED290) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func defaultDropUpdateName4ED290(update *defaultDropTestUpdate4ED290) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func (w *defaultDropTestWorld4ED290) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func (w *defaultDropTestWorld4ED290) hooks() defaultDropHooks4ED290[
	*defaultDropTestObject4ED290,
	*defaultDropTestPoint4ED290,
	*defaultDropTestUpdate4ED290,
] {
	return defaultDropHooks4ED290[
		*defaultDropTestObject4ED290,
		*defaultDropTestPoint4ED290,
		*defaultDropTestUpdate4ED290,
	]{
		loadOwnerArg: func() *defaultDropTestObject4ED290 {
			w.event("owner-arg")
			return w.owner
		},
		loadItemArg: func() *defaultDropTestObject4ED290 {
			w.event("item-arg")
			return w.item
		},
		loadInventoryOwner: func(obj *defaultDropTestObject4ED290) *defaultDropTestObject4ED290 {
			w.event("holder:" + defaultDropObjectName4ED290(obj))
			return obj.holder
		},
		loadObjectClassByte: func(obj *defaultDropTestObject4ED290) uint8 {
			w.event("class-byte:" + defaultDropObjectName4ED290(obj))
			return uint8(obj.class)
		},
		loadObjectClass: func(obj *defaultDropTestObject4ED290) uint32 {
			w.event("class:" + defaultDropObjectName4ED290(obj))
			return obj.class
		},
		loadObjectFlags: func(obj *defaultDropTestObject4ED290) uint32 {
			w.event("flags:" + defaultDropObjectName4ED290(obj))
			return obj.flags
		},
		loadObjectNetCode: func(obj *defaultDropTestObject4ED290) uint32 {
			w.event("net-code:" + defaultDropObjectName4ED290(obj))
			return obj.netCode
		},
		loadObjectTeamID: func(obj *defaultDropTestObject4ED290) uint8 {
			w.event("team:" + defaultDropObjectName4ED290(obj))
			return obj.teamID
		},
		loadObjectType: func(obj *defaultDropTestObject4ED290) uint16 {
			w.event("type:" + defaultDropObjectName4ED290(obj))
			return obj.typeInd
		},
		loadObjectUpdate: func(obj *defaultDropTestObject4ED290) *defaultDropTestUpdate4ED290 {
			w.event("update:" + defaultDropObjectName4ED290(obj))
			return obj.update
		},
		itemIsDroppable: func(obj *defaultDropTestObject4ED290) int32 {
			w.event("droppable:" + defaultDropObjectName4ED290(obj))
			return w.droppable
		},
		itemDropMask: func(obj *defaultDropTestObject4ED290, mask uint32) int32 {
			w.event(fmt.Sprintf("drop-mask:%s:%d", defaultDropObjectName4ED290(obj), mask))
			return w.dropMask
		},
		primaryMessage: func(obj *defaultDropTestObject4ED290, id string, value uint8) {
			w.event(fmt.Sprintf("message:%s:%s:%d", defaultDropObjectName4ED290(obj), id, value))
		},
		audio: func(id uint32, obj *defaultDropTestObject4ED290, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%d", id, defaultDropObjectName4ED290(obj), kind, code))
			if w.onAudio != nil {
				w.onAudio(id)
			}
		},
		detachInventory: func(owner, item *defaultDropTestObject4ED290) {
			w.event("detach:" + defaultDropObjectName4ED290(owner) + ":" + defaultDropObjectName4ED290(item))
		},
		loadPointArg: func() *defaultDropTestPoint4ED290 {
			w.event("point-arg")
			return w.point
		},
		loadPointY: func(point *defaultDropTestPoint4ED290) float32 {
			w.event("point-y")
			return point.y
		},
		loadPointX: func(point *defaultDropTestPoint4ED290) float32 {
			w.event("point-x")
			return point.x
		},
		createAt: func(item, owner *defaultDropTestObject4ED290, x, y float32, reserved uint32) {
			w.event(fmt.Sprintf("create:%s:%s:%g:%g:%d", defaultDropObjectName4ED290(item), defaultDropObjectName4ED290(owner), x, y, reserved))
			if w.onCreate != nil {
				w.onCreate()
			}
		},
		weaponEquipFlags: func(obj *defaultDropTestObject4ED290) uint32 {
			w.event("weapon:" + defaultDropObjectName4ED290(obj))
			if w.onWeapon != nil {
				w.onWeapon(obj)
			}
			return w.weaponFlags[obj]
		},
		loadInventoryHead: func(obj *defaultDropTestObject4ED290) *defaultDropTestObject4ED290 {
			w.event("head:" + defaultDropObjectName4ED290(obj))
			return obj.head
		},
		loadInventoryNext: func(obj *defaultDropTestObject4ED290) *defaultDropTestObject4ED290 {
			w.event("next:" + defaultDropObjectName4ED290(obj))
			return obj.next
		},
		loadUpdateByte2: func(update *defaultDropTestUpdate4ED290) uint8 {
			w.event("update-byte2:" + defaultDropUpdateName4ED290(update))
			return update.byte2
		},
		delayedDelete: func(obj *defaultDropTestObject4ED290) {
			w.event("delayed-delete:" + defaultDropObjectName4ED290(obj))
		},
		materialIndex: func(obj *defaultDropTestObject4ED290) uint32 {
			w.event("material:" + defaultDropObjectName4ED290(obj))
			value := w.material
			if w.onMaterial != nil {
				w.onMaterial()
			}
			return value
		},
		netInformFlagDrop: func(code uint8, netCode, material uint32) {
			w.event(fmt.Sprintf("inform:%d:%d:%d", code, netCode, material))
			w.informCode = code
			w.informNetCode = netCode
			w.informMaterial = material
			if w.onInform != nil {
				w.onInform()
			}
		},
		markMinimapForAll: func(obj *defaultDropTestObject4ED290, flags uint32) {
			w.event(fmt.Sprintf("minimap:%s:%d", defaultDropObjectName4ED290(obj), flags))
			if w.onMinimap != nil {
				w.onMinimap()
			}
		},
		loadFrame: func() uint32 {
			w.event("frame")
			return w.frame
		},
		storeUpdateFrame: func(update *defaultDropTestUpdate4ED290, frame uint32) {
			w.event(fmt.Sprintf("store-frame:%s:%d", defaultDropUpdateName4ED290(update), frame))
			update.frame = frame
		},
		setTeamFlagStatus: func(team, status, material uint8, carrier uint16) {
			w.event(fmt.Sprintf("flag-status:%d:%d:%d:%d", team, status, material, carrier))
			w.statusTeam = team
			w.statusMaterial = material
		},
		loadMonsterStatus: func(update *defaultDropTestUpdate4ED290) uint32 {
			w.event("monster-status:" + defaultDropUpdateName4ED290(update))
			return update.status
		},
		storeMonsterAction: func(update *defaultDropTestUpdate4ED290, action uint32) {
			w.event(fmt.Sprintf("store-action:%s:%d", defaultDropUpdateName4ED290(update), action))
			update.action = action
		},
		storeMonsterStatus: func(update *defaultDropTestUpdate4ED290, status uint32) {
			w.event(fmt.Sprintf("store-status:%s:%d", defaultDropUpdateName4ED290(update), status))
			update.status = status
		},
		loadGlyphCache: func() uint32 {
			w.event("glyph-cache")
			return w.glyphCache
		},
		storeGlyphCache: func(value uint32) {
			w.event(fmt.Sprintf("store-glyph:%d", value))
			w.glyphCache = value
		},
		loadTorchCache: func() uint32 {
			w.event("torch-cache")
			return w.torchCache
		},
		storeTorchCache: func(value uint32) {
			w.event(fmt.Sprintf("store-torch:%d", value))
			w.torchCache = value
		},
		loadLanternCache: func() uint32 {
			w.event("lantern-cache")
			return w.lanternCache
		},
		storeLanternCache: func(value uint32) {
			w.event(fmt.Sprintf("store-lantern:%d", value))
			w.lanternCache = value
		},
		lookupType: func(name string) uint32 {
			w.event("lookup:" + name)
			return w.lookup[name]
		},
		gameFlag: func(flag uint32) uint32 {
			w.event(fmt.Sprintf("game-flag:%d", flag))
			return w.gameFlags[flag]
		},
		loadGameFPS: func() uint32 {
			w.event("game-fps")
			return w.fps
		},
		setDecayTime: func(obj *defaultDropTestObject4ED290, delay uint32) {
			w.event(fmt.Sprintf("decay:%s:%d", defaultDropObjectName4ED290(obj), delay))
		},
		raise: func(obj *defaultDropTestObject4ED290, z float32) {
			w.event(fmt.Sprintf("raise:%s:%g", defaultDropObjectName4ED290(obj), z))
			if w.onRaise != nil {
				w.onRaise()
			}
		},
		buffOff: func(obj *defaultDropTestObject4ED290, enchant uint32) {
			w.event(fmt.Sprintf("buff-off:%s:%d", defaultDropObjectName4ED290(obj), enchant))
		},
	}
}

func newDefaultDropFullWorld4ED290() *defaultDropTestWorld4ED290 {
	owner := &defaultDropTestObject4ED290{name: "owner", netCode: 4660}
	itemUpdate := &defaultDropTestUpdate4ED290{name: "item-update", status: 0x20}
	item := &defaultDropTestObject4ED290{
		name:    "item",
		class:   defaultDropWeaponClass4ED290 | defaultDropFlagClass4ED290 | defaultDropAudioClass4ED290 | defaultDropMonsterClass4ED290,
		teamID:  7,
		typeInd: 40,
		update:  itemUpdate,
	}
	item.holder = owner
	bad := &defaultDropTestObject4ED290{name: "bad"}
	quiver := &defaultDropTestObject4ED290{
		name:   "quiver",
		class:  defaultDropWeaponClass4ED290,
		update: &defaultDropTestUpdate4ED290{name: "quiver-update", byte2: 1},
	}
	bad.next = quiver
	owner.head = bad
	return &defaultDropTestWorld4ED290{
		owner:       owner,
		item:        item,
		point:       &defaultDropTestPoint4ED290{x: 12.5, y: -7.25},
		material:    3,
		frame:       99,
		fps:         30,
		lookup:      map[string]uint32{"Glyph": 30, "Torch": 40, "Lantern": 41},
		gameFlags:   make(map[uint32]uint32),
		weaponFlags: map[*defaultDropTestObject4ED290]uint32{item: 4, quiver: 2},
	}
}

var defaultDropFullTrace4ED290 = []string{
	"owner-arg", "item-arg", "holder:item", "class-byte:owner",
	"detach:owner:item", "point-arg", "point-y", "point-x", "create:item:nil:12.5:-7.25:0",
	"class:item", "weapon:item", "head:owner", "class:bad", "next:bad",
	"class:quiver", "weapon:quiver", "flags:quiver", "update:quiver", "update-byte2:quiver-update",
	"detach:owner:quiver", "delayed-delete:quiver",
	"class:item", "update:item", "material:item", "team:item", "net-code:owner",
	"inform:7:4660:3", "minimap:item:1", "frame", "store-frame:item-update:99", "flag-status:7:2:3:0",
	"glyph-cache", "lookup:Glyph", "store-glyph:30",
	"game-flag:2048", "game-flag:4096", "flags:item", "class:item",
	"raise:item:0", "class-byte:item", "audio:821:item:0:0", "class-byte:item",
	"update:item", "monster-status:item-update", "store-action:item-update:15", "store-status:item-update:288",
	"torch-cache", "lookup:Torch", "store-torch:40", "lookup:Lantern", "store-lantern:41",
	"torch-cache", "type:item", "buff-off:owner:15",
}

func TestDefaultDrop4ED290ExactFullTrace(t *testing.T) {
	w := newDefaultDropFullWorld4ED290()
	if got := defaultDrop4ED290(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, defaultDropFullTrace4ED290) {
		t.Fatalf("events =\n%q\nwant =\n%q", w.events, defaultDropFullTrace4ED290)
	}
	if w.item.update.frame != 99 || w.item.update.action != 15 || w.item.update.status != 0x120 {
		t.Fatalf("item update = %+v", *w.item.update)
	}
	if w.glyphCache != 30 || w.torchCache != 40 || w.lanternCache != 41 {
		t.Fatalf("caches = (%d, %d, %d)", w.glyphCache, w.torchCache, w.lanternCache)
	}
}

func TestDefaultDrop4ED290FullTraceFaultPrefixes(t *testing.T) {
	for faultAt := 1; faultAt <= len(defaultDropFullTrace4ED290); faultAt++ {
		w := newDefaultDropFullWorld4ED290()
		w.faultAt = faultAt
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("fault %d did not panic", faultAt)
				}
			}()
			defaultDrop4ED290(w.hooks())
		}()
		if want := defaultDropFullTrace4ED290[:faultAt]; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("fault %d events = %q, want %q", faultAt, w.events, want)
		}
	}
}

func TestDefaultDrop4ED290EntryAndRejectGates(t *testing.T) {
	t.Run("holder mismatch", func(t *testing.T) {
		w := newDefaultDropFullWorld4ED290()
		w.item.holder = nil
		if got := defaultDrop4ED290(w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"owner-arg", "item-arg", "holder:item"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("message and audio", func(t *testing.T) {
		w := newDefaultDropFullWorld4ED290()
		w.owner.class = defaultDropPlayerClass4ED290
		w.droppable = 1
		w.dropMask = 1
		if got := defaultDrop4ED290(w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{
			"owner-arg", "item-arg", "holder:item", "class-byte:owner",
			"droppable:item", "drop-mask:item:1", "flags:owner",
			"message:owner:drop.c:CantDropThat:0", "net-code:owner", "audio:925:owner:2:4660",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("silent owner flag", func(t *testing.T) {
		w := newDefaultDropFullWorld4ED290()
		w.owner.class = defaultDropPlayerClass4ED290
		w.owner.flags = 0x20
		w.droppable = 1
		w.dropMask = 1
		if got := defaultDrop4ED290(w.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{
			"owner-arg", "item-arg", "holder:item", "class-byte:owner",
			"droppable:item", "drop-mask:item:1", "flags:owner",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})
}

func TestDefaultDrop4ED290FlagCachesUpdateButReloadsTeamNetAndFrame(t *testing.T) {
	w := newDefaultDropFullWorld4ED290()
	w.owner.class = 0
	w.item.class = 0
	w.item.typeInd = 99
	w.glyphCache = 30
	w.torchCache = 40
	w.lanternCache = 41
	w.owner.head = nil
	oldUpdate := w.item.update
	newUpdate := &defaultDropTestUpdate4ED290{name: "replacement"}
	w.onCreate = func() {
		w.item.class = defaultDropFlagClass4ED290
	}
	w.onMaterial = func() {
		w.item.update = newUpdate
		w.item.teamID = 9
		w.owner.netCode = 77
	}
	w.onInform = func() {
		w.item.teamID = 11
		w.material = 8
		w.frame = 101
	}
	w.onMinimap = func() {
		w.frame = 202
	}

	if got := defaultDrop4ED290(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.informCode != 7 || w.informNetCode != 77 || w.informMaterial != 3 {
		t.Fatalf("inform = (%d, %d, %d)", w.informCode, w.informNetCode, w.informMaterial)
	}
	if w.statusTeam != 9 || w.statusMaterial != 3 {
		t.Fatalf("status = (%d, %d)", w.statusTeam, w.statusMaterial)
	}
	if oldUpdate.frame != 202 || newUpdate.frame != 0 {
		t.Fatalf("frames = old %d, replacement %d", oldUpdate.frame, newUpdate.frame)
	}
}

func TestDefaultDrop4ED290DecayGatesAndUint32Wrap(t *testing.T) {
	w := newDefaultDropFullWorld4ED290()
	w.item.class = 0
	w.owner.head = nil
	w.item.typeInd = 31
	w.glyphCache = 30
	w.torchCache = 40
	w.lanternCache = 41
	w.fps = math.MaxUint32

	if got := defaultDrop4ED290(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wrappedDelay := uint32(math.MaxUint32)
	wrappedDelay *= 10
	wantDecay := fmt.Sprintf("decay:item:%d", wrappedDelay)
	if !containsDefaultDropEvent4ED290(w.events, wantDecay) {
		t.Fatalf("events do not contain %q: %q", wantDecay, w.events)
	}

	w = newDefaultDropFullWorld4ED290()
	w.item.class = 0
	w.owner.head = nil
	w.glyphCache = 30
	w.torchCache = 40
	w.gameFlags[defaultDropCoopFlag4ED290] = 1
	defaultDrop4ED290(w.hooks())
	if containsDefaultDropEvent4ED290(w.events, "game-flag:4096") ||
		containsDefaultDropPrefix4ED290(w.events, "flags:item") ||
		containsDefaultDropPrefix4ED290(w.events, "decay:") {
		t.Fatalf("co-op gate read later decay state: %q", w.events)
	}
}

func TestDefaultDrop4ED290TorchCacheAsymmetry(t *testing.T) {
	w := newDefaultDropFullWorld4ED290()
	w.item.class = 0
	w.owner.head = nil
	w.item.typeInd = 40
	w.glyphCache = 30
	w.torchCache = 40
	w.lanternCache = 0
	w.gameFlags[defaultDropCoopFlag4ED290] = 1
	defaultDrop4ED290(w.hooks())
	if containsDefaultDropPrefix4ED290(w.events, "lookup:Torch") ||
		containsDefaultDropPrefix4ED290(w.events, "lookup:Lantern") ||
		containsDefaultDropEvent4ED290(w.events, "lantern-cache") {
		t.Fatalf("warm Torch cache touched Lantern initialization: %q", w.events)
	}
	if !containsDefaultDropEvent4ED290(w.events, "buff-off:owner:15") {
		t.Fatalf("Torch did not clear enchant: %q", w.events)
	}
}

func TestDefaultDrop4ED290NilOwnerFaultsOnlyAfterMatchingHolder(t *testing.T) {
	w := newDefaultDropFullWorld4ED290()
	w.owner = nil
	w.item.holder = nil
	defer func() {
		if recover() == nil {
			t.Fatal("nil owner did not fault")
		}
		want := []string{"owner-arg", "item-arg", "holder:item", "class-byte:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	defaultDrop4ED290(w.hooks())
}

func containsDefaultDropEvent4ED290(events []string, target string) bool {
	for _, event := range events {
		if event == target {
			return true
		}
	}
	return false
}

func containsDefaultDropPrefix4ED290(events []string, prefix string) bool {
	for _, event := range events {
		if len(event) >= len(prefix) && event[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
