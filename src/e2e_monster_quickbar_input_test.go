package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
)

func e2eTestQuickbarSample(frame uint32) e2eMonsterQuickbarSample {
	return e2eMonsterQuickbarSample{
		frame: frame, mouse: image.Pt(550, 400),
		target: e2eNamedMonsterInputTarget{wire: 17, point: image.Pt(550, 400), distance: 90, health: 10, playerHP: 15},
	}
}

func e2eTestQuickbarKeys(events []seat.InputEvent) []seat.KeyboardEvent {
	var out []seat.KeyboardEvent
	for _, event := range events {
		switch event := event.(type) {
		case *seat.KeyboardEvent:
			out = append(out, *event)
		case *seat.MouseMoveEvent:
		default:
			panic("quickbar aiming must not emit mouse buttons or other input")
		}
	}
	return out
}

func TestMonsterQuickbarInputConfigBoundsAndOriginalKeys(t *testing.T) {
	keys := [...]keybind.Key{keybind.KeyA, keybind.KeyS, keybind.KeyD, keybind.KeyF, keybind.KeyG}
	for slot, want := range keys {
		key, count, err := e2eMonsterQuickbarConfig("AlbinoSpider", slot+1, 18)
		if err != nil || key != want || count != 18 {
			t.Fatalf("original quickbar key mapping: %v %v %v", key, count, err)
		}
	}
	key, count, err := e2eMonsterQuickbarConfig("Imp", 0, 0)
	if err != nil || key != keybind.KeyA || count != 18 {
		t.Fatal("defaults must use the original A key and a bounded tracking window")
	}
	for _, values := range [][2]int{{-1, 18}, {6, 18}, {1, -1}, {1, 121}} {
		if _, _, err := e2eMonsterQuickbarConfig("Imp", values[0], values[1]); err == nil {
			t.Fatal("invalid input configuration must be rejected before queuing any input")
		}
	}
	if _, _, err := e2eMonsterQuickbarConfig("", 1, 18); err == nil {
		t.Fatal("an original creature type is required")
	}
}

func TestMonsterQuickbarInputQueuesOneOriginalKeyAndRelease(t *testing.T) {
	st := e2eMonsterQuickbarTracker{key: keybind.KeyA, ticks: 18}
	var events []seat.InputEvent
	for _, frame := range []uint32{1, 2, 2, 3, 20} {
		next, done := st.step(e2eTestQuickbarSample(frame), true)
		events = append(events, next...)
		if done != (frame == 20) {
			t.Fatalf("unexpected tracking completion at frame%d", frame)
		}
	}
	want := []seat.KeyboardEvent{{Key: keybind.KeyA, Pressed: true}, {Key: keybind.KeyA, Pressed: false}}
	if !reflect.DeepEqual(e2eTestQuickbarKeys(events), want) || !st.started || st.pressed {
		t.Fatal("aiming must queue exactly one normal key press and release, without direct casting")
	}
}

func TestMonsterQuickbarInputTracksCurrentSpriteAndCamera(t *testing.T) {
	st := e2eMonsterQuickbarTracker{key: keybind.KeyS, ticks: 5}
	st.step(e2eTestQuickbarSample(10), true)
	st.step(e2eTestQuickbarSample(11), true)
	for frame := uint32(12); frame <= 16; frame++ {
		sample := e2eTestQuickbarSample(frame)
		sample.target.point = image.Pt(550+int(frame)*2, 400-int(frame))
		events, done := st.step(sample, true)
		var last image.Point
		for _, event := range events {
			if move, ok := event.(*seat.MouseMoveEvent); ok {
				last = move.Pos
			}
		}
		if last != sample.target.point || done != (frame == 16) {
			t.Fatal("tracking must use each new projected sprite position, not a stored screen coordinate")
		}
	}
}

func TestMonsterQuickbarInputRejectsCompetingNativeCursor(t *testing.T) {
	st := e2eMonsterQuickbarTracker{key: keybind.KeyA, ticks: 18}
	for frame := uint32(1); frame < 10; frame++ {
		sample := e2eTestQuickbarSample(frame)
		sample.competing = true
		events, done := st.step(sample, true)
		if done || len(e2eTestQuickbarKeys(events)) != 0 {
			t.Fatal("a different original selected object must not authorize a key press")
		}
	}
	if st.probe == 0 || e2eMonsterQuickbarProbe(st.probe) == (image.Point{}) {
		t.Fatal("overlap must be explored with ordinary nearby mouse points")
	}
	events, done := st.step(e2eTestQuickbarSample(10), true)
	if done || len(e2eTestQuickbarKeys(events)) != 1 || !st.started {
		t.Fatal("only a cleared competing cursor permits the original key input")
	}
}

func TestMonsterQuickbarInputDisappearanceReleasesKeyWithoutRetargeting(t *testing.T) {
	st := e2eMonsterQuickbarTracker{key: keybind.KeyA, ticks: 18}
	st.step(e2eTestQuickbarSample(1), true)
	st.step(e2eTestQuickbarSample(2), true)
	events, done := st.step(e2eMonsterQuickbarSample{}, false)
	if !done || !reflect.DeepEqual(e2eTestQuickbarKeys(events), []seat.KeyboardEvent{{Key: keybind.KeyA, Pressed: false}}) {
		t.Fatal("missing target must release a held key and finish without selecting a replacement")
	}
	events, done = st.step(e2eTestQuickbarSample(4), true)
	if !done || len(events) != 0 {
		t.Fatal("a finished action must not cast again when another creature appears")
	}
}

func TestMonsterQuickbarInputRejectsUnavailableDeadAndInvalidRange(t *testing.T) {
	for _, distance := range []float32{-1, 301, float32(math.NaN()), float32(math.Inf(1))} {
		st := e2eMonsterQuickbarTracker{key: keybind.KeyA, ticks: 18}
		sample := e2eTestQuickbarSample(1)
		sample.target.distance = distance
		if events, done := st.step(sample, true); !done || len(events) != 0 {
			t.Fatal("invalid or excessive distance must never queue a cast")
		}
	}
	for _, field := range []string{"absent", "dead", "target-health", "player-health", "unavailable"} {
		st := e2eMonsterQuickbarTracker{key: keybind.KeyA, ticks: 18}
		sample, available := e2eTestQuickbarSample(1), true
		switch field {
		case "absent":
			sample.target.absent = true
		case "dead":
			sample.target.dead = true
		case "target-health":
			sample.target.health = 0
		case "player-health":
			sample.target.playerHP = 0
		case "unavailable":
			available = false
		}
		if events, done := st.step(sample, available); !done || len(events) != 0 {
			t.Fatalf("%s must not authorize any mouse or key input", field)
		}
	}
}

func TestMonsterQuickbarInputUsesNativeFramesAcrossPauseAndWrap(t *testing.T) {
	st := e2eMonsterQuickbarTracker{key: keybind.KeyA, ticks: 18}
	st.step(e2eTestQuickbarSample(math.MaxUint32-2), true)
	st.step(e2eTestQuickbarSample(math.MaxUint32-1), true)
	for i := 0; i < 20; i++ {
		if events, done := st.step(e2eTestQuickbarSample(math.MaxUint32-1), true); done || len(e2eTestQuickbarKeys(events)) != 0 {
			t.Fatal("paused native frames must not consume tracking time or duplicate key events")
		}
	}
	if events, done := st.step(e2eTestQuickbarSample(0), true); done || len(e2eTestQuickbarKeys(events)) != 1 {
		t.Fatal("native frame wrap must release the key and continue the bounded window")
	}
	if _, done := st.step(e2eTestQuickbarSample(16), true); !done {
		t.Fatal("unsigned native frame delta must finish at the original eighteen-frame window")
	}
}

func TestMonsterQuickbarInputBoundsReadsAndRetainsNoLiveReferences(t *testing.T) {
	st := e2eMonsterQuickbarTracker{key: keybind.KeyA, ticks: 18}
	sample := e2eTestQuickbarSample(1)
	sample.competing = true
	for i := 1; i <= 600; i++ {
		events, done := st.step(sample, true)
		if done != (i == 600) || len(e2eTestQuickbarKeys(events)) != 0 {
			t.Fatal("a blocked/paused original cursor must terminate without casting")
		}
	}
	var check func(reflect.Type)
	check = func(typ reflect.Type) {
		switch typ.Kind() {
		case reflect.Pointer, reflect.UnsafePointer, reflect.Map, reflect.Interface, reflect.Slice, reflect.Func, reflect.Chan:
			t.Fatalf("input tracking retains a live reference: %v", typ)
		case reflect.Struct:
			for i := 0; i < typ.NumField(); i++ {
				check(typ.Field(i).Type)
			}
		case reflect.Array:
			check(typ.Elem())
		}
	}
	check(reflect.TypeOf(st))
	check(reflect.TypeOf(sample))
	check(reflect.TypeOf(e2eMonsterQuickbarPositionWatch{}))
	check(reflect.TypeOf(e2eMonsterQuickbarSwapTarget{}))
}

func TestMonsterQuickbarInputRuntimeCallsOnlyReadersAndOriginalInput(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_monster_quickbar_input.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{}
	for _, name := range []string{"Errorf", "len", "uint32", "int", "uint16", "float64", "append", "Hypot",
		"e2eMonsterInputFiniteDistance", "Class", "Has", "Flags", "HasAny", "Viewport", "ClientPlayerUnit", "HostUnit",
		"First", "Next", "ByInd", "ID", "GetUnitNetCode", "e2eReadMonsterInputTarget", "CanInteract", "IsEnemyTo",
		"observe", "Frame", "GetMousePos", "Nox_xxx_clientGetSpriteAtCursor_476F90", "UpdateDataPlayer",
		"e2eMonsterQuickbarConfig", "e2eError", "addWhen", "e2eReadMonsterQuickbarSample", "Add",
		"e2eMonsterQuickbarProbe", "step", "e2eQueueInput", "Printf",
		"e2eLogMonsterQuickbarPosition", "e2eWatchMonsterQuickbarPositions", "e2eObserveMonsterQuickbarSwapTarget"} {
		allowed[name] = true
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			name := ""
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				name = fn.Name
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			}
			if !allowed[name] {
				t.Errorf("quickbar runtime invokes unreviewed operation %q", name)
			}
		}
		if literal, ok := node.(*ast.CompositeLit); ok {
			if typ, ok := literal.Type.(*ast.SelectorExpr); ok && typ.X.(*ast.Ident).Name == "seat" &&
				typ.Sel.Name != "MouseMoveEvent" && typ.Sel.Name != "KeyboardEvent" {
				t.Errorf("input helper queues an unexpected native event: %s", typ.Sel.Name)
			}
		}
		return true
	})
}

func TestMonsterQuickbarPositionWatchAfterInput(t *testing.T) {
	w := e2eMonsterQuickbarPositionWatch{start: 645, last: 645, limit: 60, playerScriptID: 7}
	sample := e2eMonsterQuickbarSample{frame: 646, playerScriptID: 7, hasPositions: true}
	if !w.observe(sample, true) || w.observe(sample, true) {
		t.Fatal("observe a new native frame once after the input action finishes")
	}
	sample.frame = 662
	if !w.observe(sample, true) {
		t.Fatal("late cast/save frames must remain observable without input")
	}
}

func TestMonsterQuickbarPositionWatchStopsAtDeadline(t *testing.T) {
	w := e2eMonsterQuickbarPositionWatch{start: 645, last: 645, limit: 60, playerScriptID: 7}
	sample := e2eMonsterQuickbarSample{frame: 705, playerScriptID: 7, hasPositions: true}
	if !w.observe(sample, true) {
		t.Fatal("last bounded frame")
	}
	sample.frame++
	if w.observe(sample, true) || !w.stopped {
		t.Fatal("observer must stop after its native-frame budget")
	}
	sample.frame = 650
	if w.observe(sample, true) {
		t.Fatal("an expired observer cannot restart on another map/frame")
	}
}

func TestMonsterQuickbarPositionWatchRejectsMissingOrReplacedPair(t *testing.T) {
	for _, mode := range []string{"missing", "positions", "player"} {
		w := e2eMonsterQuickbarPositionWatch{start: 645, last: 645, limit: 60, playerScriptID: 7}
		sample := e2eMonsterQuickbarSample{frame: 646, playerScriptID: 7, hasPositions: true}
		found := true
		switch mode {
		case "missing":
			found = false
		case "positions":
			sample.hasPositions = false
		case "player":
			sample.playerScriptID = 8
		}
		if w.observe(sample, found) || !w.stopped {
			t.Fatalf("%s must end observation instead of retaining/replacing a native reference", mode)
		}
	}
}

func TestMonsterQuickbarPositionWatchHandlesPauseAndFrameWrap(t *testing.T) {
	w := e2eMonsterQuickbarPositionWatch{start: math.MaxUint32 - 1, last: math.MaxUint32 - 1, limit: 4, playerScriptID: 7}
	sample := e2eMonsterQuickbarSample{frame: math.MaxUint32 - 1, playerScriptID: 7, hasPositions: true}
	if w.observe(sample, true) || w.stopped {
		t.Fatal("a paused native frame must neither duplicate nor end observation")
	}
	sample.frame = 0
	if !w.observe(sample, true) {
		t.Fatal("unsigned native-frame wrap")
	}
	sample.frame = 3
	if w.observe(sample, true) || !w.stopped {
		t.Fatal("budget still applies across wrap")
	}
}

func TestMonsterQuickbarPositionRuntimeCallsOnlyReaders(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_monster_quickbar_position.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{}
	for _, name := range []string{"liveFrame", "Frame", "HostUnit", "Flags", "HasAny", "First", "Next", "Class", "Has",
		"GetUnitNetCode", "int", "ByInd", "ID", "TickRate", "TickHook", "observe", "Printf",
		"e2eReadMonsterQuickbarPosition", "e2eLogMonsterQuickbarPosition", "e2eObserveMonsterQuickbarSwapTarget"} {
		allowed[name] = true
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			name := ""
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				name = fn.Name
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			}
			if !allowed[name] {
				t.Errorf("passive position observer invokes unreviewed operation %q", name)
			}
		}
		return true
	})
}
