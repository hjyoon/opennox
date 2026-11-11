package legacy

import (
	"bytes"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func game2ServerSaveOwners(fields ...int) func() {
	old := make([]unsafe.Pointer, len(fields))
	for i, field := range fields {
		old[i] = game2ServerFixtureOwner(field)
	}
	return func() {
		for i, field := range fields {
			game2ServerFixtureSetOwner(field, old[i])
		}
	}
}
func game2ServerTestList(g *gui.GUI, parent *gui.Window, id int, multi bool) *gui.Window {
	opts := gui.ScrollListBoxData{Count: 8, Line_height: 10}
	if multi {
		opts.Field_4 = 1
	}
	d := gui.WindowData{Window: parent, Style: gui.StyleScrollListBox}
	w := gui.NewScrollListBoxRaw(g, parent, gui.StatusEnabled, 0, 0, 180, 100, &d, &opts)
	w.SetID(uint(id))
	return w
}

func TestGame2ServerNativeLayoutAndFirstVisible(t *testing.T) {
	var w gui.Window
	var d gui.WindowData
	var list gui.ScrollListBoxData
	var item gui.ScrollListBoxItem
	for field, want := range []uintptr{unsafe.Sizeof(w), unsafe.Sizeof(d), unsafe.Sizeof(list), unsafe.Sizeof(item), unsafe.Sizeof(accessListNode416720{})} {
		if got := game2ServerFixtureLayout(field); got != want {
			t.Fatalf("C/Go named native layout field%d got%d want%d", field, got, want)
		}
	}
	items, free := alloc.MakeHigh([]gui.ScrollListBoxItem{}, 4)
	defer free()
	for i, v := range []uint32{11, 22, 33, 44} {
		items[i].Field_0 = v
	}
	list = gui.ScrollListBoxData{Count: 4, Items: &items[0], Field_11_0: 4}
	for _, scroll := range []uint16{0, 10, 11, 22, 32, 33, 44, 65535} {
		list.Field_13_1 = scroll
		if got, want := g2ServerFirstVisible(&list), game2ServerFixtureFirstVisible(unsafe.Pointer(&list)); got != want {
			t.Fatalf("first visible at scroll%d got%d current-C%d", scroll, got, want)
		}
	}
	if g2ServerFirstVisible(nil) != 0 {
		t.Fatal("nil list")
	}
	list.Items = nil
	if g2ServerFirstVisible(&list) != 0 {
		t.Fatal("missing native item owner")
	}
}

func TestGame2ServerNativeFilterCallbacks(t *testing.T) {
	g := newWidgetCallbackGUI(t)
	restore := game2ServerSaveOwners(0, 1, 100)
	defer func() { restore(); g.DestroyAll() }()
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	list := game2ServerTestList(g, root, 1512, false)
	game2ServerFixtureSetOwner(0, root.C())
	game2ServerFixtureSetOwner(1, list.C())
	for id := 1520; id <= 1533; id++ {
		w := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 20, 10, nil)
		w.SetID(uint(id))
		w.DrawData().Field0Set(4, true)
	}
	data := (*gui.ScrollListBoxData)(list.WidgetData)
	data.Field_11_0 = 0
	// A real GUI tree and real C owner globals, not PE32 window tokens.
	if sub_453750() != 0 {
		t.Fatal("empty filter rows should hide every checkbox")
	}
	for id := 1520; id <= 1533; id++ {
		if !root.ChildByID(uint(id)).GetFlags().IsHidden() {
			t.Fatal("empty checkbox remained visible")
		}
	}
	controls := unsafe.Slice((*byte)(game2ServerFixtureControls()), 12)
	saved := append([]byte(nil), controls...)
	defer copy(controls, saved)
	game2ServerFixtureSetOwner(100, root.C())
	for _, spec := range []struct {
		id   int
		text string
		off  int
		want []byte
	}{{10126, "999", 1, []byte{0xAE}}, {10130, "65535", 5, []byte{0xFF, 0xFF}}} {
		draw := gui.WindowData{Window: root, Style: gui.StyleEntryField}
		edit := gui.NewEntryFieldRaw(g, root, gui.StatusEnabled, 0, 0, 100, 20, &draw, &gui.EntryFieldData{Field_1040: 32})
		edit.SetID(uint(spec.id))
		requireNativeWindowAddress(t, edit)
		for i := range controls {
			controls[i] = byte(0xA0 + i)
		}
		before := append([]byte(nil), controls...)
		text := alloc.InternCString16(spec.text)
		edit.Func94(gui.AsWindowEvent(16414, uintptr(unsafe.Pointer(text)), 0))
		if game2ServerFixtureAccessEvent(nil, 16387, 0, uintptr(spec.id)) != 0 {
			t.Fatal("edit notification return")
		}
		copy(before[spec.off:], spec.want)
		if !bytes.Equal(controls, before) {
			t.Fatalf("BYTE/unalignedWORD edit escaped neighbors: id%d", spec.id)
		}
		if spec.id == 10126 {
			p := gui.EventRespPtr(edit.Func94(gui.AsWindowEvent(16413, 0, 0)))
			if alloc.GoString16((*uint16)(p)) != "14" {
				t.Fatal("clamped edit text must return through full pointer callback")
			}
		}
	}
}

func TestGame2ServerNativePlayerListOwners(t *testing.T) {
	for _, old := range []bool{true, false} {
		t.Run(map[bool]string{true: "private-current-C", false: "public-Go"}[old], func(t *testing.T) {
			g := newWidgetCallbackGUI(t)
			restore := game2ServerSaveOwners(5, 6, 7)
			players, teams := game2ServerFixtureHead(false), game2ServerFixtureHead(true)
			oldPlayers, oldTeams := *players, *teams
			defer func() {
				game2ServerFixtureClearHead(false)
				game2ServerFixtureClearHead(true)
				*players = oldPlayers
				*teams = oldTeams
				restore()
				g.DestroyAll()
			}()
			accessListNativeClear416720(players)
			accessListNativeClear416720(teams)
			root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
			list := game2ServerTestList(g, root, 10501, true)
			game2ServerTestList(g, root, 10502, false)
			game2ServerFixtureSetOwner(5, root.C())
			names := []string{"Alpha Ω", "Beta 中"}
			ids := []uint32{0x80000001, 0xFFFFFFFF}
			for i, name := range names {
				p := unsafe.Pointer(alloc.InternCString16(name))
				if game2ServerFixtureCall(old, 0x457140, p, int32(ids[i]), 0) != 1 {
					t.Fatal("actual list add-line result")
				}
			}
			first, second := players.Next, players.Next.Next
			if first == players || second == players || first == second {
				t.Fatal("two full native allocated nodes")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(first)) <= 0xFFFFFFFF {
				t.Fatal("C list node must exercise high native pointers")
			}
			for i, node := range []*accessListNode416720{first, second} {
				name := alloc.GoString16((*uint16)(unsafe.Add(unsafe.Pointer(node), game2ServerFixtureLayout(6))))
				id := *(*uint32)(unsafe.Add(unsafe.Pointer(node), game2ServerFixtureLayout(7)))
				if name != names[i] || id != ids[i] {
					t.Fatal("UTF16 name or DWORD net code changed width")
				}
			}
			for i := 0; i < 3; i++ {
				runtime.GC()
				if game2ServerFixtureOwner(5) != root.C() || players.Next != first || first.Next != second || second.Next != players {
					t.Fatal("GC/current C head lifetime")
				}
			}
			for i, id := range ids {
				if int32(game2ServerFixtureCall(old, 0x456E40, nil, int32(id), 0)) != int32(i) {
					t.Fatal("raw high-bit DWORD identity lookup")
				}
			}
			teamA := game2ServerFixtureTeam(1, 7, unsafe.Pointer(alloc.InternCString16("Red")))
			teamB := game2ServerFixtureTeam(0x101, 3, unsafe.Pointer(alloc.InternCString16("Blue")))
			if teamA == nil || teamB == nil {
				t.Fatal("C team allocation")
			}
			accessListNativeAppend416720(teams, (*accessListNode416720)(teamA))
			accessListNativeAppend416720(teams, (*accessListNode416720)(teamB))
			game2ServerFixtureCall(old, 0x457350, nil, 0x101, 0x1FE)
			off := game2ServerFixtureLayout(8)
			if *(*uint32)(unsafe.Add(teamA, off)) != 0xFE || *(*uint32)(unsafe.Add(teamB, off)) != 0x101 {
				t.Fatal("only ID arguments narrow to BYTE; stored team IDs stay DWORD")
			}
			if int32(game2ServerFixtureCall(old, 0x456F10, unsafe.Pointer(alloc.InternCString16("rED")), 0, 0)) != 0 {
				t.Fatal("case-insensitive native team lookup")
			}
			if game2ServerFixtureCall(old, 0x4571A0, nil, int32(ids[1]), 0xFE) != 1 {
				t.Fatal("team-color item update")
			}
			data := (*gui.ScrollListBoxData)(list.WidgetData)
			if data.Field_11_0 != 2 {
				t.Fatal("list item ownership/count")
			}
			if game2ServerFixtureCall(old, 0x456E40, nil, int32(ids[0]), 1) != 0 || players.Next != second || second.Prev != players {
				t.Fatal("remove must unlink the actual C head before freeing")
			}
			if game2ServerFixtureCall(old, 0x456F10, unsafe.Pointer(alloc.InternCString16("RED")), 1, 0) != 0 || teams.Next != (*accessListNode416720)(teamB) {
				t.Fatal("remove native team by case-insensitive name")
			}
		})
	}
}

func TestGame2ServerNativeHUDNoRendering(t *testing.T) {
	g := newWidgetCallbackGUI(t)
	defer g.DestroyAll()
	old := game2ServerFixtureNoRender(true)
	defer game2ServerFixtureNoRender(old)
	roots := [2]unsafe.Pointer{game2ServerFixtureOwner(9), game2ServerFixtureOwner(10)}
	before := append([]byte(nil), unsafe.Slice((*byte)(g2ServerMem(1045608)), 40)...)
	if game2ServerFixtureHUD(true, 16) != 0 || game2ServerFixtureHUD(false, 0) != 0 {
		t.Fatal("no-render must not construct/show either HUD")
	}
	if roots != [2]unsafe.Pointer{game2ServerFixtureOwner(9), game2ServerFixtureOwner(10)} || !bytes.Equal(before, unsafe.Slice((*byte)(g2ServerMem(1045608)), 40)) {
		t.Fatal("disabled HUD changed native owners or neighboring scalar state")
	}
}

func TestGame2ServerNativeDrawableABI(t *testing.T) {
	size := game2ServerFixtureLayout(5)
	link := game2ServerFixtureAlloc(size)
	if link == nil {
		t.Fatal("C linked drawable")
	}
	defer game2ServerFixtureFree(link)
	for _, input := range []struct{ class, sub, anim uint32 }{{0, 0, 8}, {2, 0x40000, 8}, {2, 0x40000, 7}, {2, 0, 8}} {
		var storage [2][]byte
		var owners [2]unsafe.Pointer
		for backend := 0; backend < 2; backend++ {
			p := game2ServerFixtureAlloc(size + 16)
			if p == nil {
				t.Fatal("C drawable")
			}
			owners[backend] = p
			defer game2ServerFixtureFree(p)
			all := unsafe.Slice((*byte)(p), int(size)+16)
			copy(all[int(size):], bytes.Repeat([]byte{0xB6}, 16))
			game2ServerFixtureDrawable(p, input.class, input.sub, 0x80000000, input.anim, 0x12345678, 0xABCDEF01, link)
			if got := game2ServerFixtureCall(backend == 0, 0x45AB80, p, -17, 0); got != uintptr(p) {
				t.Fatal("typed C frame entry must return the full pointer")
			}
			if got := game2ServerFixtureCall(backend == 0, 0x45A990, p, 0, 0); got != uintptr(p) {
				t.Fatal("typed active entry identity")
			}
			for _, entry := range []uint32{0x459EC0, 0x45A010, 0x45A070} {
				if game2ServerFixtureCall(backend == 0, entry, p, 0, 0) != uintptr(link) {
					t.Fatal("native linked drawable identity")
				}
			}
			storage[backend] = append([]byte(nil), all...)
		}
		if !bytes.Equal(storage[0], storage[1]) {
			t.Fatal("preserved current C/native Go complete drawable and canaries differ")
		}
		game2ServerFixtureCall(true, 0x45AB80, owners[0], -17, 0)
		if game2ServerFixtureLegacyFrame(owners[1], -17) != owners[1] || !bytes.Equal(unsafe.Slice((*byte)(owners[0]), int(size)+16), unsafe.Slice((*byte)(owners[1]), int(size)+16)) {
			t.Fatal("legacy spelling must use the same full pointer input/output and repeated-call state transitions")
		}
	}
	var dr client.Drawable
	if game2ServerFixtureLayout(5) > unsafe.Sizeof(dr) {
		t.Fatal("C drawable prefix exceeds named Go owner")
	}
}

func TestGame2ServerNativePinnedDrawables(t *testing.T) {
	parent, child := new(client.Drawable), new(client.Drawable)
	var pins runtime.Pinner
	pins.Pin(parent)
	pins.Pin(child)
	defer pins.Unpin()
	parent.NextPtr = child
	parent.Field_102 = child
	parent.Field_104 = child
	parent.ObjClass = 2
	parent.ObjSubClass = 0
	parent.ObjFlags = 0x80000000
	parent.AnimInd = 7
	parent.AnimFrameSlave = 0x12345678
	parent.Field_78 = 0xABCDEF01
	want := *parent
	childBefore := *child
	for i := 0; i < 3; i++ {
		runtime.GC()
		if game2ServerFixtureCall(false, 0x45AB80, parent.C(), -17, 0) != uintptr(parent.C()) {
			t.Fatal("C return narrowed a pinned native Go owner")
		}
		if game2ServerFixtureCall(false, 0x45A990, parent.C(), 0, 0) != uintptr(parent.C()) {
			t.Fatal("pinned owner active entry identity")
		}
		want.Field_78 = want.AnimFrameSlave
		want.AnimFrameSlave = 0xFFFFFFEF
		want.ObjFlags |= 4
		for _, entry := range []uint32{0x459EC0, 0x45A010, 0x45A070} {
			if game2ServerFixtureCall(false, entry, parent.C(), 0, 0) != uintptr(child.C()) {
				t.Fatal("nested individually pinned Go link")
			}
		}
		if !reflect.DeepEqual(*parent, want) || !reflect.DeepEqual(*child, childBefore) {
			t.Fatal("native callbacks changed unrelated typed owner fields")
		}
	}
	runtime.KeepAlive(parent)
	runtime.KeepAlive(child)
}
