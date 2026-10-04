package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"

	"gopkg.in/yaml.v2"
)

func TestE2ESecretWallTouchLaneClearanceAndFallback(t *testing.T) {
	grid := image.Pt(91, 71)
	for direction := byte(0); direction <= 1; direction++ {
		for _, radius := range []float32{1, 10, 32} {
			base, _, _, normal, err := e2eSecretWallTouchGeometry(grid, direction, radius)
			if err != nil {
				t.Fatal(err)
			}
			for _, useShort := range []bool{false, true} {
				var starts []types.Pointf
				var checks [][2]types.Pointf
				traceClear := func(from, to types.Pointf) bool {
					starts = append(starts, from)
					return false // Both real candidate traces must remain blocked.
				}
				laneClear := func(from, to types.Pointf) bool {
					checks = append(checks, [2]types.Pointf{from, to})
					return !useShort || from.Sub(base).Len() < float64(radius+26)
				}
				center, start, end, gotNormal, err := e2eSecretWallTouchLane(grid, direction, radius, traceClear, laneClear)
				gap, calls := radius+36, 1
				if useShort {
					gap, calls = radius+16, 2
				}
				if err != nil || center != base || gotNormal != normal || len(starts) != calls ||
					math.Abs(start.Sub(base).Len()-float64(gap)) > 0.001 ||
					math.Abs(end.Sub(base).Len()-float64(gap)) > 0.001 || len(checks) < 34 {
					t.Fatalf("dir=%d radius=%g short=%t center/start/end=%v/%v/%v starts=%v checks=%d err=%v",
						direction, radius, useShort, center, start, end, starts, len(checks), err)
				}
				// Both complete sets of sixteen starting-circle radial traces
				// use the same radius+4 clearance as the original wide lane.
				circles := checks[len(checks)-32:]
				for around := 0; around < 256; around += 16 {
					cx, cy := server.SinCosDir(byte(around))
					offset := types.Ptf((radius+4)*cx, (radius+4)*cy)
					for side, pos := range []types.Pointf{start, end} {
						if pair := circles[2*(around/16)+side]; pair != [2]types.Pointf{pos, pos.Add(offset)} {
							t.Fatalf("circle side=%d direction=%d clearance=%v, want radius+4", side, around, pair)
						}
					}
				}
			}
		}
	}
}

func TestE2ESecretWallTouchLaneRejectsUnsafeCandidates(t *testing.T) {
	grid := image.Pt(91, 71)
	for blockedRadial := 0; blockedRadial < 32; blockedRadial++ {
		calls := 0
		laneClear := func(from, to types.Pointf) bool {
			calls++
			if from == to {
				t.Fatal("clearance must not collapse to an empty ray")
			}
			// Each candidate first probes both approach segments, then
			// the two sixteen-ray circles in alternating start/end order.
			if calls == blockedRadial+3 {
				return false
			}
			// Do not allow a rejected wide circle to become an unchecked
			// short-circle fallback.
			return calls < blockedRadial+4
		}
		if _, _, _, _, err := e2eSecretWallTouchLane(grid, 1, 10,
			func(types.Pointf, types.Pointf) bool { return false }, laneClear); err == nil {
			t.Fatalf("obstructed circle ray %d passed", blockedRadial)
		}
	}
	for _, traceClear := range []bool{false, true} {
		for _, laneClear := range []bool{false, true} {
			if !traceClear && laneClear {
				continue
			}
			if _, _, _, _, err := e2eSecretWallTouchLane(grid, 0, 10,
				func(types.Pointf, types.Pointf) bool { return traceClear },
				func(types.Pointf, types.Pointf) bool { return laneClear }); err == nil {
				t.Fatalf("unsafe closed/approach trace=%t lane=%t passed", traceClear, laneClear)
			}
		}
	}
	for _, nilTrace := range []bool{false, true} {
		trace, clear := (func(types.Pointf, types.Pointf) bool)(nil), (func(types.Pointf, types.Pointf) bool)(nil)
		if nilTrace {
			clear = func(types.Pointf, types.Pointf) bool { t.Fatal("nil trace called clearance"); return true }
		} else {
			trace = func(types.Pointf, types.Pointf) bool { t.Fatal("nil clearance called trace"); return false }
		}
		if _, _, _, _, err := e2eSecretWallTouchLane(grid, 0, 10, trace, clear); err == nil {
			t.Fatal("missing clearance service passed")
		}
	}
	for _, invalid := range []struct {
		grid      image.Point
		direction byte
		radius    float32
	}{
		{image.Pt(-1, 71), 1, 10}, {image.Pt(91, 70), 1, 10}, {grid, 2, 10},
		{grid, 1, 0}, {grid, 1, 33}, {grid, 1, float32(math.NaN())}, {grid, 1, float32(math.Inf(1))},
	} {
		check := func(types.Pointf, types.Pointf) bool { t.Fatal("invalid geometry queried the map"); return false }
		if _, _, _, _, err := e2eSecretWallTouchLane(invalid.grid, invalid.direction, invalid.radius, check, check); err == nil {
			t.Fatalf("invalid lane passed: %+v", invalid)
		}
	}
}

func TestE2ESecretWallTouchOutcomeSensitivity(t *testing.T) {
	valid := e2eSecretWallTouchOutcome{
		beforeState: 1, contact: true, opening: true, moved: true,
		full: true, passable: true, stable: true, bound: true, progress: 0xFFFFFF, sounds: 1,
		crossed: [2]bool{true, true}, clientCrossed: [2]bool{true, true},
	}
	if err := valid.validate(); err != nil {
		t.Fatal(err)
	}
	for _, corrupt := range []func(*e2eSecretWallTouchOutcome){
		func(o *e2eSecretWallTouchOutcome) { o.beforeState = 0 },
		func(o *e2eSecretWallTouchOutcome) { o.beforeState = 4 },
		func(o *e2eSecretWallTouchOutcome) { o.beforeDelay = 1 },
		func(o *e2eSecretWallTouchOutcome) { o.contact = false },
		func(o *e2eSecretWallTouchOutcome) { o.opening = false },
		func(o *e2eSecretWallTouchOutcome) { o.moved = false },
		func(o *e2eSecretWallTouchOutcome) { o.full = false },
		func(o *e2eSecretWallTouchOutcome) { o.passable = false },
		func(o *e2eSecretWallTouchOutcome) { o.stable = false },
		func(o *e2eSecretWallTouchOutcome) { o.bound = false },
		func(o *e2eSecretWallTouchOutcome) { o.progress |= 1 << 24 },
		func(o *e2eSecretWallTouchOutcome) { o.sounds = 0 },
		func(o *e2eSecretWallTouchOutcome) { o.sounds = 2 },
		func(o *e2eSecretWallTouchOutcome) { o.crossed[0] = false },
		func(o *e2eSecretWallTouchOutcome) { o.crossed[1] = false },
		func(o *e2eSecretWallTouchOutcome) { o.clientCrossed[0] = false },
		func(o *e2eSecretWallTouchOutcome) { o.clientCrossed[1] = false },
	} {
		bad := valid
		corrupt(&bad)
		if err := bad.validate(); err == nil {
			t.Fatalf("corrupted secret-wall result passed: %+v", bad)
		}
	}
	for delay := 1; delay <= 23; delay++ {
		bad := valid
		bad.progress &^= 1 << delay
		if err := bad.validate(); err == nil {
			t.Fatalf("missing natural delay %d passed", delay)
		}
	}
	valid.progress &^= 1
	if err := valid.validate(); err != nil {
		t.Fatalf("pre-predicate collision delay zero must be optional: %v", err)
	}
}

func TestE2ESecretWallTouchGeometry(t *testing.T) {
	for direction := byte(0); direction <= 1; direction++ {
		center, start, end, normal, err := e2eSecretWallTouchGeometry(image.Pt(48, 62), direction, 10)
		if err != nil || center.X != 1115.5 || center.Y != 1437.5 || normal.X <= 0 ||
			(direction == 0) != (normal.Y > 0) || math.Abs(float64(normal.Len()-1)) > 1e-6 {
			t.Fatalf("native wall lane: center=%v normal=%v err=%v", center, normal, err)
		}
		for _, pos := range []struct {
			x, y, sign float32
		}{{start.X, start.Y, -1}, {end.X, end.Y, 1}} {
			dx, dy := pos.x-center.X, pos.y-center.Y
			if math.Abs(float64(dx*normal.X+dy*normal.Y-pos.sign*46)) > 0.001 ||
				math.Abs(float64(dx*normal.Y-dy*normal.X)) > 0.001 {
				t.Fatal("lane does not cross the wall midpoint perpendicularly with separate starting circles")
			}
		}
	}
	for _, grid := range []image.Point{image.Pt(-1, 1), image.Pt(1, -1), image.Pt(256, 0), image.Pt(0, 256), image.Pt(48, 63)} {
		if _, _, _, _, err := e2eSecretWallTouchGeometry(grid, 0, 10); err == nil {
			t.Fatalf("invalid stock wall grid passed: %v", grid)
		}
	}
	for direction := byte(2); direction <= 10; direction++ {
		if _, _, _, _, err := e2eSecretWallTouchGeometry(image.Pt(48, 62), direction, 10); err == nil {
			t.Fatalf("fixture must not assume a single diagonal for dir=%d", direction)
		}
	}
	for _, radius := range []float32{-1, 0, 33, float32(math.NaN()), float32(math.Inf(1))} {
		if _, _, _, _, err := e2eSecretWallTouchGeometry(image.Pt(48, 62), 0, radius); err == nil {
			t.Fatalf("invalid fixture radius passed: %g", radius)
		}
	}
}

func TestE2ESecretWallTouchSchedule(t *testing.T) {
	var sc e2eScenario
	sc.CheckSecretWallTouch("touch")
	if len(sc.steps) != 14 {
		t.Fatalf("steps=%d, want bounded preparation/contact/round-trip/stability", len(sc.steps))
	}
	for index, timeout := range map[int]int{0: 1200, 2: 120, 4: 180, 5: 120, 9: 180, 12: 120} {
		step := sc.steps[index]
		if step.ready == nil || step.fnc == nil || int(step.waitTimeout) != timeout {
			t.Fatalf("step %d lacks its bounded live check", index)
		}
	}
	for _, index := range []int{3, 8} {
		if sc.steps[index].ready != nil || sc.steps[index].fnc == nil || !strings.Contains(sc.steps[index].name, "actual") {
			t.Fatal("both crossings require actual queued movement")
		}
	}
	for _, index := range []int{6, 10} {
		if sc.steps[index].fnc == nil || !strings.Contains(sc.steps[index].name, "release") {
			t.Fatal("both crossings must release the real mouse button")
		}
	}
}

func TestE2ESecretWallTouchPublicScenario(t *testing.T) {
	const path = "../scripts/e2e/host-warrior-secret-wall-touch.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	checks, switches, mapsReady, warriorClicks := 0, 0, 0, 0
	for _, step := range file.Steps {
		switch step.Action {
		case "check-secret-wall-touch":
			checks++
		case "switch-map":
			if step.Map != "G_Crypts" {
				t.Fatalf("fixture must load the stock touch-enabled map, got %q", step.Map)
			}
			switches++
		case "wait-map":
			if step.Map != "G_Crypts" {
				t.Fatal("fixture must wait for the real stock map")
			}
			mapsReady++
		case "click":
			if step.X == 256 && step.Y == 208 {
				warriorClicks++
			}
		case "slow", "wait", "quit":
		default:
			t.Fatalf("secret-wall scenario supplies state/results or a new PNG baseline: %s", step.Action)
		}
	}
	if checks != 2 || switches != 1 || mapsReady != 1 || warriorClicks != 1 {
		t.Fatalf("scenario check/switch/map-ready/Warrior=%d/%d/%d/%d", checks, switches, mapsReady, warriorClicks)
	}
	var sc e2eScenario
	sc.Load(path)
	stable := 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " stable open wall") {
			stable++
		}
	}
	if stable != 2 {
		t.Fatal("public actions did not schedule both stock walls' stable observations")
	}
}

func TestE2ESecretWallTouchDoesNotSupplyWallOrCollisionResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_secret_wall_touch.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	forbiddenFields := map[string]bool{
		"State": true, "OpenDelay": true, "Flags": true, "OpenWait": true, "LastOpen": true, "PlayerBits": true,
		"Next": true, "Wall": true, "Data": true, "CollisionWall": true, "Dir0": true, "Tile1": true, "Flags4": true,
		"Field10": true, "UpdateData": true, "PosVec": true, "NewPos": true, "VelVec": true, "ForceVec": true,
	}
	forbiddenCalls := map[string]bool{
		"AttachSecret": true, "SetData": true, "Enable": true, "Toggle": true, "open": true, "close": true,
		"sub_548100": true, "sub_54FFC0": true, "Sub_4ED0C0": true, "EventPos": true, "EventObj": true,
		"NetSendPacketXxx0": true, "NetSendPacketXxx": true, "handleWallStatePacketNative48EA70": true,
		"Uint8Ptr": true, "Uint32Ptr": true, "PtrOff": true, "PtrT": true, "NewObjectByTypeID": true,
	}
	var inspectWrite func(ast.Expr)
	inspectWrite = func(expr ast.Expr) {
		switch expr := expr.(type) {
		case *ast.SelectorExpr:
			if forbiddenFields[expr.Sel.Name] {
				t.Errorf("observer writes native game result field %s", expr.Sel.Name)
			}
		case *ast.IndexExpr:
			inspectWrite(expr.X)
		case *ast.ParenExpr:
			inspectWrite(expr.X)
		case *ast.StarExpr:
			t.Error("observer writes through a native pointer")
		}
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				for _, expr := range node.Lhs {
					inspectWrite(expr)
				}
			case *ast.IncDecStmt:
				inspectWrite(node.X)
			case *ast.CallExpr:
				name := ""
				switch called := node.Fun.(type) {
				case *ast.SelectorExpr:
					name = called.Sel.Name
				case *ast.Ident:
					name = called.Name
				}
				if forbiddenCalls[name] || name == "SetPos" && fn.Name.Name != "prepare" && fn.Name.Name != "CheckSecretWallTouch" ||
					name == "e2eQueueInput" && fn.Name.Name != "input" {
					t.Errorf("%s supplies a gameplay result through %s", fn.Name.Name, name)
				}
			}
			return true
		})
	}
}
