package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type worldPlayerTestEnv40A830 struct {
	pin       runtime.Pinner
	players   [4]*server.Player
	units     [4]*server.Object
	updates   [4]*server.PlayerUpdateData
	teams     [2]*server.Team
	wall      *server.Wall
	target    []uint16
	r         worldPlayerRuntime40A830
	events    []string
	packets   [][]byte
	badLesson bool
}

func worldPlayerTestHigh40A830(t *testing.T, p unsafe.Pointer) {
	t.Helper()
	if p == nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32) {
		t.Fatalf("native full pointer: %p", p)
	}
}
func worldPlayerTestEnvNew40A830(t *testing.T, teams, host bool) *worldPlayerTestEnv40A830 {
	e := new(worldPlayerTestEnv40A830)
	t.Cleanup(e.pin.Unpin)
	for i := range e.players {
		e.players[i] = new(server.Player)
		e.units[i] = new(server.Object)
		e.updates[i] = new(server.PlayerUpdateData)
		for _, p := range []unsafe.Pointer{unsafe.Pointer(e.players[i]), unsafe.Pointer(e.units[i]), unsafe.Pointer(e.updates[i])} {
			e.pin.Pin(p)
			worldPlayerTestHigh40A830(t, p)
		}
		e.players[i].PlayerUnit = e.units[i]
		e.players[i].PlayerInd = byte(i + 2)
		e.players[i].NetCodeVal = 0xf000be00 + uint32(i)
		e.players[i].Lessons = int32(5 + i)
		e.players[i].Field2140 = 0x80000100 + uint32(i)
		e.updates[i].Player = e.players[i]
		e.units[i].UpdateData = unsafe.Pointer(e.updates[i])
		e.units[i].TeamVal.ID = server.TeamID(1 + i/2)
	}
	e.players[0].Field3680 = 0x80000422
	e.players[1].Field3680 = 1
	e.players[2].Field3680 = 0x21
	e.players[3].Field3680 = 0
	e.players[3].PlayerInd = 31
	for i := range e.teams {
		e.teams[i] = new(server.Team)
		e.pin.Pin(e.teams[i])
		e.teams[i].IDVal = server.TeamID(i + 1)
	}
	e.wall = new(server.Wall)
	e.pin.Pin(e.wall)
	for i, s := range []string{"ALPHA", "Beta", "GAMMA", "DELTA"} {
		for j, c := range []byte(s) {
			e.players[i].NameFinal[j] = uint16(c)
		}
	}
	e.target = []uint16{'b', 'E', 't', 'A', 0}
	e.pin.Pin(&e.target[0])
	pi := func(p *server.Player) int {
		for i, v := range e.players {
			if p == v {
				return i
			}
		}
		return -1
	}
	ui := func(p *server.Object) int {
		for i, v := range e.units {
			if p == v {
				return i
			}
		}
		return -1
	}
	ti := func(p *server.Team) int {
		for i, v := range e.teams {
			if p == v {
				return i
			}
		}
		return -1
	}
	e.r = worldPlayerDefaults40A830
	e.r.PlayerFirst = func() *server.Player { e.events = append(e.events, "player:first"); return e.players[0] }
	e.r.PlayerNext = func(p *server.Player) *server.Player {
		i := pi(p)
		e.events = append(e.events, fmt.Sprint("player:next:", i))
		if i >= 0 && i < 3 {
			return e.players[i+1]
		}
		return nil
	}
	e.r.UnitFirst = func() *server.Object { e.events = append(e.events, "unit:first"); return e.units[0] }
	e.r.UnitNext = func(p *server.Object) *server.Object {
		i := ui(p)
		e.events = append(e.events, fmt.Sprint("unit:next:", i))
		if i >= 0 && i < 3 {
			return e.units[i+1]
		}
		return nil
	}
	e.r.TeamFirst = func() *server.Team { e.events = append(e.events, "team:first"); return e.teams[0] }
	e.r.TeamNext = func(p *server.Team) *server.Team {
		i := ti(p)
		e.events = append(e.events, fmt.Sprint("team:next:", i))
		if i == 0 {
			return e.teams[1]
		}
		return nil
	}
	e.r.TeamContains = func(value *server.ObjectTeam, id server.TeamID) bool {
		e.events = append(e.events, fmt.Sprintf("member:%d:%d", value.ID, id))
		return value.ID == id
	}
	e.r.EngineNoRendering = func() bool { e.events = append(e.events, "no-rendering"); return true }
	e.r.GameplayTeams = func() bool { e.events = append(e.events, "gameplay-teams"); return teams }
	e.r.GameFlag = func(mask uint32) uint32 {
		e.events = append(e.events, fmt.Sprintf("flag:%x", mask))
		if mask == 0x40000 || (host && mask == 1) {
			return 1
		}
		return 0
	}
	e.r.GameFlags = func() uint32 { e.events = append(e.events, "flags-value"); runtime.GC(); return 0x80008123 }
	e.r.TimerStatus = func() int32 { e.events = append(e.events, "timer-status"); return 0 }
	e.r.TimerMode = func(v int16) byte {
		e.events = append(e.events, fmt.Sprintf("timer-mode:%04x", uint16(v)))
		return 0x81
	}
	e.r.TimerStart = func() { e.events = append(e.events, "timer-start") }
	e.r.TimerSet = func(v int32) int32 { e.events = append(e.events, fmt.Sprint("timer-set:", v)); return -2 }
	e.r.Send = func(recipient int32, packet []byte) int32 {
		e.events = append(e.events, fmt.Sprintf("send:%d:%x", recipient, packet))
		e.packets = append(e.packets, append([]byte(nil), packet...))
		e.players[0].PlayerInd++
		e.players[1].Field3680 ^= 0x20
		return -77
	}
	e.r.Lesson = func(unit *server.Object) {
		i := ui(unit)
		e.events = append(e.events, fmt.Sprint("lesson:", i))
		if i < 0 || e.players[i].Lessons != 0 || e.players[i].Field2140 != 0 {
			e.badLesson = true
		}
		runtime.GC()
	}
	e.r.Minimap = func(mark bool, index byte, object *server.Object, flags uint32) {
		e.events = append(e.events, fmt.Sprintf("minimap:%t:%d:%d:%08x", mark, index, ui(object), flags))
	}
	e.r.WallCreate = func(x, y int32) *server.Wall {
		e.events = append(e.events, fmt.Sprintf("create:%d:%d", x, y))
		return e.wall
	}
	e.r.WallAttach = func(w *server.Wall, o *server.Object) unsafe.Pointer {
		e.events = append(e.events, fmt.Sprintf("attach:%t:%d", w == e.wall, ui(o)))
		return unsafe.Pointer(w)
	}
	t.Cleanup(pushWorldPlayerRuntime40A830(&e.r))
	return e
}

func TestWorldPlayerNativeCurrentC40A830(t *testing.T) {
	type result struct {
		words   []uint32
		events  []string
		packets [][]byte
		fields  [4][3]uint32
	}
	for _, scenario := range []struct {
		name        string
		teams, host bool
	}{{"players-host", false, true}, {"teams-host", true, true}, {"nonhost", false, false}} {
		t.Run(scenario.name, func(t *testing.T) {
			var reference result
			for _, backend := range []string{"old", "go", "c"} {
				t.Run(backend, func(t *testing.T) {
					e := worldPlayerTestEnvNew40A830(t, scenario.teams, scenario.host)
					got := result{}
					for _, team := range e.teams {
						got.words = append(got.words, worldPlayerFixtureWord40A830(backend, 0, unsafe.Pointer(team), 0))
					}
					got.words = append(got.words, worldPlayerFixtureWord40A830(backend, 1, nil, 0), worldPlayerFixtureWord40A830(backend, 7, nil, 0))
					if p := worldPlayerFixturePointer40A830(backend, 7, unsafe.Pointer(&e.target[0]), nil, 0, 0); p != unsafe.Pointer(e.players[1]) {
						t.Fatalf("name full identity: %p", p)
					}
					if p := worldPlayerFixturePointer40A830(backend, 7, nil, nil, 0, 0); p != nil {
						t.Fatal("nil name")
					}
					if p := worldPlayerFixturePointer40A830(backend, 0, unsafe.Pointer(e.units[0]), nil, -17, math.MaxInt32); p != unsafe.Pointer(e.wall) {
						t.Fatal("door create/attach identity")
					}
					got.words = append(got.words, worldPlayerFixtureWord40A830(backend, 9, unsafe.Pointer(e.players[0]), 1), worldPlayerFixtureWord40A830(backend, 10, unsafe.Pointer(e.players[2]), 0))
					for _, mask := range []uint32{0x80, 0x400, 0x80000000} {
						got.words = append(got.words, worldPlayerFixtureWord40A830(backend, 9, unsafe.Pointer(e.players[0]), mask))
					}
					if p := worldPlayerFixturePointer40A830(backend, 10, unsafe.Pointer(e.players[0]), nil, 0, 0); p != nil {
						t.Fatal("send all return")
					}
					for _, entry := range []int{8, 9} {
						if p := worldPlayerFixturePointer40A830(backend, entry, unsafe.Pointer(e.units[2]), nil, int32(-2147483648), 0); p != nil {
							t.Fatal("minimap return")
						}
					}
					if p := worldPlayerFixturePointer40A830(backend, 6, nil, nil, 1, 0); p != nil || e.badLesson {
						t.Fatal("lessons reset before report")
					}
					for i, p := range e.players {
						got.fields[i] = [3]uint32{p.Field3680, uint32(p.Lessons), p.Field2140}
					}
					got.events, got.packets = e.events, e.packets
					for _, packet := range got.packets {
						if len(packet) != 7 || packet[0] != 106 || binary.LittleEndian.Uint32(packet[3:])&^uint32(0x423) != 0 {
							t.Fatalf("7-byte status wire: %x", packet)
						}
					}
					if backend == "old" {
						reference = got
					} else if !reflect.DeepEqual(got, reference) {
						t.Fatalf("%s current-C/source-order mismatch\ngot=%+v\nold=%+v", backend, got, reference)
					}
				})
			}
		})
	}
	for _, backend := range []string{"old", "go", "c"} {
		t.Run("empty-"+backend, func(t *testing.T) {
			e := worldPlayerTestEnvNew40A830(t, false, true)
			e.r.PlayerFirst = func() *server.Player { return nil }
			e.r.UnitFirst = func() *server.Object { return nil }
			e.r.TeamFirst = func() *server.Team { return nil }
			e.r.PlayerNext = func(*server.Player) *server.Player { panic("empty list next") }
			e.r.UnitNext = func(*server.Object) *server.Object { panic("empty unit next") }
			e.r.TeamNext = func(*server.Team) *server.Team { panic("empty team next") }
			if got := worldPlayerFixtureWord40A830(backend, 0, nil, 0); got != 0 {
				t.Fatal("empty count must not dereference team")
			}
			for _, teams := range []bool{false, true} {
				e.r.GameplayTeams = func() bool { return teams }
				if got := worldPlayerFixtureWord40A830(backend, 1, nil, 0); got != 0 {
					t.Fatal("empty any-player gate")
				}
			}
			if got := worldPlayerFixturePointer40A830(backend, 7, unsafe.Pointer(&e.target[0]), nil, 0, 0); got != nil {
				t.Fatal("empty name list")
			}
			for _, entry := range []int{6, 8, 9, 10} {
				if got := worldPlayerFixturePointer40A830(backend, entry, nil, nil, 1, 0); got != nil {
					t.Fatal("empty broadcast must not access recipient/object")
				}
			}
		})
	}
}

func TestWorldPlayerNativeScalarOwners4100F0(t *testing.T) {
	bank := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 371300)), 512)
	saved := append([]byte(nil), bank...)
	t.Cleanup(func() { copy(bank, saved) })
	cases := [][]uint16{{0}, {'A', 'z', 0}, {0x00ff, 0}, {0x0100, 0}, {'a', 0x8001, 0}, {0x00ff, 0x00fe, 0}}
	for i, text := range cases {
		var pin runtime.Pinner
		pin.Pin(&text[0])
		func() {
			defer pin.Unpin()
			want := worldPlayerFixtureWord40A830("old", 2, unsafe.Pointer(&text[0]), 0)
			for _, backend := range []string{"go", "c"} {
				if got := worldPlayerFixtureWord40A830(backend, 2, unsafe.Pointer(&text[0]), 0); got != want {
					t.Fatalf("UTF16 BYTE-half case%d %s", i, backend)
				}
			}
		}()
	}
	for _, index := range []int32{-1, 0, 1} {
		want := worldPlayerFixturePointer40A830("old", 5, nil, nil, index, 0)
		for _, backend := range []string{"go", "c"} {
			if got := worldPlayerFixturePointer40A830(backend, 5, nil, nil, index, 0); got != want {
				t.Fatal("58-byte record full pointer")
			}
			worldPlayerTestHigh40A830(t, want)
		}
	}
	for _, value := range []uint32{0, 1, 0x80000000, 0xffffffff} {
		for _, backend := range []string{"old", "go", "c"} {
			copy(bank, saved)
			before := append([]byte(nil), bank...)
			got := worldPlayerFixtureWord40A830(backend, 8, nil, value)
			expected := append([]byte(nil), before...)
			binary.LittleEndian.PutUint32(expected[404:], value)
			if got != value || !bytes.Equal(bank, expected) {
				t.Fatalf("DWORD settings owner %s/%08x", backend, value)
			}
		}
	}
}

func TestWorldPlayerNativeDefaultPause413A50(t *testing.T) {
	saved := noxflags.GetGame()
	defer func() { noxflags.ResetGame(); noxflags.SetGame(saved) }()
	defer pushWorldPlayerRuntime40A830(&worldPlayerDefaults40A830)()
	for _, flags := range []uint32{0, 0x40000, 0x80040000, 0x80000000} {
		noxflags.ResetGame()
		noxflags.SetGame(noxflags.GameFlag(flags))
		want := uint32(0)
		if flags&0x40000 != 0 {
			want = 1
		}
		for _, backend := range []string{"old", "go", "c"} {
			if got := worldPlayerFixtureWord40A830(backend, 7, nil, 0); got != want || uint32(noxflags.GetGame()) != flags {
				t.Fatalf("actual flags owner %s/%08x", backend, flags)
			}
		}
	}
}

func TestWorldPlayerNativeListsPinned410760(t *testing.T) {
	restore := mapgenHighFixtureMemory()
	defer restore()
	for _, backend := range []string{"old", "go", "c"} {
		t.Run(backend, func(t *testing.T) {
			nodes := [2]*server.SecretWall{new(server.SecretWall), new(server.SecretWall)}
			var pin runtime.Pinner
			for _, n := range nodes {
				pin.Pin(n)
			}
			defer pin.Unpin()
			cell := (*unsafe.Pointer)(secretWallNativeHeadCell410550())
			old := *cell
			defer func() { *cell = old }()
			*cell = nil
			for i, n := range nodes {
				worldPlayerTestHigh40A830(t, unsafe.Pointer(n))
				n.X = int32(0x12345678 + i)
				n.Y = -19
				n.Flags = 0xfe
				n.PlayerBits = 0x87654321
				if got := worldPlayerFixturePointer40A830(backend, 1, unsafe.Pointer(n), nil, 0, 0); got != unsafe.Pointer(n) {
					t.Fatal("insert return full pointer")
				}
				runtime.GC()
			}
			if got := worldPlayerFixturePointer40A830(backend, 2, nil, nil, 0, 0); got != unsafe.Pointer(nodes[1]) {
				t.Fatal("real typed C head")
			}
			if got := worldPlayerFixturePointer40A830(backend, 3, unsafe.Pointer(nodes[1]), nil, 0, 0); got != unsafe.Pointer(nodes[0]) {
				t.Fatal("native Next slot")
			}
			if got := worldPlayerFixturePointer40A830(backend, 3, nil, nil, 0, 0); got != nil {
				t.Fatal("nil Next")
			}
			for i, n := range nodes {
				if n.X != int32(0x12345678+i) || n.Y != -19 || n.Flags != 0xfe || n.PlayerBits != 0x87654321 {
					t.Fatal("neighbor scalars changed")
				}
			}
			// Real C head is restored by the later defer before individually pinned
			// Go owners are unpinned. No C free is ever applied to a Go allocation.
		})
	}
	first := worldPlayerFixtureAlloc40A830(unsafe.Sizeof(server.ModifierEff{}))
	next := worldPlayerFixtureAlloc40A830(unsafe.Sizeof(server.ModifierEff{}))
	defer worldPlayerFixtureFree40A830(first)
	defer worldPlayerFixtureFree40A830(next)
	worldPlayerTestHigh40A830(t, first)
	worldPlayerTestHigh40A830(t, next)
	worldPlayerFixtureModifierNext40A830(first, next)
	for _, backend := range []string{"old", "go", "c"} {
		if got := worldPlayerFixturePointer40A830(backend, 4, first, nil, 0, 0); got != next {
			t.Fatal("modifier native Next full pointer")
		}
	}
}

func TestWorldPlayerNativeArmorCallback415C00(t *testing.T) {
	restore := mapgenHighFixtureMemory()
	defer restore()
	item := (*server.Object)(worldPlayerFixtureAlloc40A830(unsafe.Sizeof(server.Object{})))
	attrs := (*server.ModifierInitData)(worldPlayerFixtureAlloc40A830(unsafe.Sizeof(server.ModifierInitData{})))
	effect := (*server.ModifierEff)(worldPlayerFixtureAlloc40A830(unsafe.Sizeof(server.ModifierEff{})))
	definition := (*server.Modifier)(worldPlayerFixtureAlloc40A830(unsafe.Sizeof(server.Modifier{})))
	for _, p := range []unsafe.Pointer{unsafe.Pointer(item), unsafe.Pointer(attrs), unsafe.Pointer(effect), unsafe.Pointer(definition)} {
		worldPlayerTestHigh40A830(t, p)
		defer worldPlayerFixtureFree40A830(p)
	}
	r := worldPlayerDefaults40A830
	r.ArmorDefinition = func(kind uint16) *server.Modifier {
		if kind != 0xbeef {
			return nil
		}
		return definition
	}
	defer pushWorldPlayerRuntime40A830(&r)()
	for _, scenario := range []struct {
		name                                         string
		itemNil, armor, def, attrs, effect, callback bool
	}{{"nil", true, false, false, false, false, false}, {"other-class", false, false, false, false, false, false}, {"no-definition", false, true, false, false, false, false}, {"base", false, true, true, false, false, false}, {"no-effect", false, true, true, true, false, false}, {"no-callback", false, true, true, true, true, false}, {"full-callback", false, true, true, true, true, true}} {
		t.Run(scenario.name, func(t *testing.T) {
			*item = server.Object{}
			*attrs = server.ModifierInitData{}
			*effect = server.ModifierEff{}
			*definition = server.Modifier{}
			if scenario.armor {
				item.ObjClass = 0x02000000
			}
			if scenario.def {
				item.TypeInd = 0xbeef
			}
			if scenario.attrs {
				item.InitData = unsafe.Pointer(attrs)
			}
			if scenario.effect {
				attrs.Modifiers[0] = effect
			}
			if scenario.callback {
				effect.Defend76.Fnc = worldPlayerFixtureDefendCallback40A830()
			}
			definition.DamageCoeffOrArmor64 = 0.375
			arg := item
			if scenario.itemNil {
				arg = nil
			}
			want := worldPlayerFixtureArmor40A830("old", arg)
			for _, backend := range []string{"go", "c"} {
				got := worldPlayerFixtureArmor40A830(backend, arg)
				if math.Float64bits(got) != math.Float64bits(want) {
					t.Fatalf("%s double result %g/%g", backend, got, want)
				}
				if scenario.callback && !worldPlayerFixtureDefendSeen40A830(unsafe.Pointer(effect), unsafe.Pointer(item)) {
					t.Fatal("full typed callback args/item identity")
				}
			}
		})
	}
}

func TestWorldPlayerNativeTileOwnership411160(t *testing.T) {
	restore := mapgenHighFixtureMemory()
	defer restore()
	var player server.Player
	var object server.Object
	var cell worldTileCell411160
	expected := []uintptr{unsafe.Sizeof(player), unsafe.Offsetof(player.PlayerUnit), unsafe.Offsetof(player.Field3680), unsafe.Offsetof(player.Lessons), unsafe.Offsetof(player.Field2140), unsafe.Offsetof(player.NameFinal), unsafe.Offsetof(object.UpdateData), unsafe.Offsetof(object.TeamVal), 0, unsafe.Sizeof(server.ModifierEff{}), unsafe.Sizeof(cell), unsafe.Offsetof(cell.List5), unsafe.Offsetof(cell.List10), unsafe.Sizeof(server.TileDef{})}
	for i, v := range expected {
		if i != 8 && worldPlayerFixtureLayout40A830(i) != v {
			t.Fatalf("actual C/Go named layout field%d", i)
		}
	}
	oldHead := *(*unsafe.Pointer)(worldTileHeadCell411160())
	oldCap := worldTileCapacity411160()
	defer worldPlayerFixtureTileOwner40A830(oldHead, oldCap)
	edge := unsafe.Slice((*byte)(memmap.PtrOff(0x85B3FC, 28696)), 2)
	edgeBefore := append([]byte(nil), edge...)
	defer copy(edge, edgeBefore)
	copy(edge, []byte{3, 3})
	defs := worldTileDefinitions411160()
	savedDefs := append([]server.TileDef(nil), defs...)
	defer copy(defs, savedDefs)
	cache := unsafe.Slice((*byte)(memmap.PtrOff(0x587000, 26516)), 20)
	savedCache := append([]byte(nil), cache...)
	defer copy(cache, savedCache)
	points := []types.Pointf{{X: 184, Y: 184}, {X: 184 + 10, Y: 184 + 35}, {X: 184 + 35, Y: 184 + 10}, {X: 184 + 35, Y: 184 + 35}, {X: 184 + 1, Y: 184 + 1}, {X: -200, Y: 0}, {X: 6000, Y: 6000}}
	var reference []uint32
	for _, backend := range []string{"old", "go", "c"} {
		func() {
			worldPlayerFixtureTileOwner40A830(nil, 128)
			if worldPlayerFixtureWord40A830(backend, 3, nil, 0) != 1 {
				t.Fatal("tile allocation")
			}
			head := *(*unsafe.Pointer)(worldTileHeadCell411160())
			worldPlayerTestHigh40A830(t, head)
			defer worldPlayerFixtureFree40A830(head)
			rowsFreed := false
			defer func() {
				if !rowsFreed {
					worldPlayerFixtureWord40A830(backend, 4, nil, 0)
				}
			}()
			node := (*tileListNativeNode411350)(worldPlayerFixtureAlloc40A830(unsafe.Sizeof(tileListNativeNode411350{})))
			defer worldPlayerFixtureFree40A830(unsafe.Pointer(node))
			node.Value = 9001
			node.Variation = 6
			for x := int32(0); x < 128; x++ {
				row := worldTileRow411160(head, x)
				worldPlayerTestHigh40A830(t, row)
				for _, b := range unsafe.Slice((*byte)(row), 128*int(unsafe.Sizeof(cell))) {
					if b != 0 {
						t.Fatal("calloc row not zero")
					}
				}
				for y := 0; y < 128; y++ {
					c := (*worldTileCell411160)(unsafe.Add(row, uintptr(y)*unsafe.Sizeof(cell)))
					c.Value1, c.Value6 = 1000+x*128+int32(y), 2000+x*128+int32(y)
					c.List5, c.List10 = node, node
				}
			}
			var got []uint32
			for i := range points {
				got = append(got, worldPlayerFixtureWord40A830(backend, 5, unsafe.Pointer(&points[i]), 0))
			}
			for _, missing := range []int{3, 4} {
				func() {
					slot := (*unsafe.Pointer)(unsafe.Add(head, uintptr(missing)*unsafe.Sizeof(head)))
					saved := *slot
					defer func() { *slot = saved }()
					*slot = nil
					if got := worldPlayerFixtureWord40A830(backend, 5, unsafe.Pointer(&points[0]), 0); got != math.MaxUint32 {
						t.Fatal("current missing-row transition policy")
					}
				}()
			}
			for x := int32(0); x < 128; x++ {
				row := worldTileRow411160(head, x)
				for y := 0; y < 128; y++ {
					c := (*worldTileCell411160)(unsafe.Add(row, uintptr(y)*unsafe.Sizeof(cell)))
					c.Value1, c.Value6 = 17, 17
					c.List5, c.List10 = nil, nil
				}
			}
			for i := range defs {
				defs[i].NameBuf = [32]byte{}
			}
			copy(defs[17].NameBuf[:], "WaterDeepNoTeleport")
			for i := range cache {
				cache[i] = 0
			}
			binary.LittleEndian.PutUint32(cache[4:], 0xffffffff)
			for i := range points {
				got = append(got, worldPlayerFixtureWord40A830(backend, 6, unsafe.Pointer(&points[i]), 0))
			}
			if binary.LittleEndian.Uint32(cache[4:]) != 17 {
				t.Fatal("actual teleport sentinel/cache owner")
			}
			worldPlayerFixtureTileOwner40A830(nil, 128)
			if got := worldPlayerFixtureWord40A830(backend, 5, unsafe.Pointer(&points[0]), 0); got != math.MaxUint32 {
				t.Fatal("current nil-map transition policy")
			}
			worldPlayerFixtureTileOwner40A830(head, 128)
			if backend == "old" {
				reference = got
			} else if !reflect.DeepEqual(got, reference) {
				t.Fatalf("%s tile/teleport current C: %v/%v", backend, got, reference)
			}
			slots := unsafe.Slice((*unsafe.Pointer)(head), 128)
			savedSlots := append([]unsafe.Pointer(nil), slots...)
			worldPlayerFixtureWord40A830(backend, 4, nil, 0)
			rowsFreed = true
			if !reflect.DeepEqual(slots, savedSlots) {
				t.Fatal("current tileFree must leave head/row slots unchanged")
			}
		}()
	}
}
