package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"log/slog"
	"math"
	"testing"

	"github.com/opennox/libs/client/seat"

	"github.com/opennox/opennox/v1/client/input"
	seatheadless "github.com/opennox/opennox/v1/client/seat/headless"
)

func TestE2ELockIdleInputUnsignedCadence(t *testing.T) {
	for _, test := range []struct {
		name        string
		last, frame uint32
		want        bool
	}{
		{"before", 0, 299, false},
		{"boundary", 0, 300, true},
		{"late", 0, 301, true},
		{"same", 300, 300, false},
		{"next-before", 300, 599, false},
		{"next-boundary", 300, 600, true},
		{"wrap-before", math.MaxUint32 - 200, 98, false},
		{"wrap-boundary", math.MaxUint32 - 200, 99, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			keep := e2eLockIdleInput{lastFrame: test.last}
			cursor := image.Pt(417, 328)
			var events []seat.InputEvent
			got := keep.observe(test.frame, cursor, func(ev ...seat.InputEvent) { events = append(events, ev...) })
			if got != test.want {
				t.Fatalf("queued=%t want=%t", got, test.want)
			}
			wantFrame := test.last
			if test.want {
				wantFrame = test.frame
				if len(events) != 1 {
					t.Fatalf("events=%d want one motion", len(events))
				}
				move, ok := events[0].(*seat.MouseMoveEvent)
				if !ok || move.Pos != cursor || move.Relative || move.Rel.X != 0 || move.Rel.Y != 0 {
					t.Fatalf("motion changed actual cursor or supplied relative movement: %#v", events[0])
				}
			} else if len(events) != 0 {
				t.Fatalf("early observation produced input: %v", events)
			}
			if keep.lastFrame != wantFrame {
				t.Fatalf("cadence checkpoint=%d want=%d", keep.lastFrame, wantFrame)
			}
		})
	}
}

func TestE2ELockIdleMotionUsesRealHeadlessInput(t *testing.T) {
	for _, scaled := range []bool{false, true} {
		name := "native"
		if scaled {
			name = "scaled"
		}
		t.Run(name, func(t *testing.T) {
			st := seatheadless.New(image.Pt(1024, 768))
			t.Cleanup(func() { _ = st.Close() })
			h := input.New(slog.Default(), st, false, 0)
			if scaled {
				h.SetWinSize(image.Rect(12, 18, 652, 498))
				h.SetDrawWinSize(image.Pt(1024, 768))
			} else {
				h.SetWinSize(image.Rect(0, 0, 1024, 768))
				h.SetDrawWinSize(image.Pt(1024, 768))
			}
			st.QueueInput(&seat.MouseMoveEvent{Pos: image.Pt(240, 170)})
			h.Tick()
			cursor := h.GetMousePos()
			for ticks := 0; ticks < 2701; ticks++ {
				h.Tick()
			}
			if h.SeqDelay() <= 2700 {
				t.Fatal("control did not reach the unchanged idle-observer threshold")
			}
			var keep e2eLockIdleInput
			if !keep.observe(300, cursor, func(events ...seat.InputEvent) {
				for _, event := range events {
					st.QueueInput(e2ePlaybackInput(event, e2eCanvasInputSpace, h.DrawPosToWindow, nil))
				}
			}) {
				t.Fatal("Lock motion did not use the ordinary canvas input queue")
			}
			h.Tick()
			if h.SeqDelay() != 1 || h.GetMousePos() != cursor || h.GetMouseRel() != (image.Point{}) {
				t.Fatalf("motion changed cursor/movement or failed to reset ordinary input activity: delay=%d cursor=%v/%v delta=%v", h.SeqDelay(), h.GetMousePos(), cursor, h.GetMouseRel())
			}
			for _, button := range []seat.MouseButton{seat.MouseButtonLeft, seat.MouseButtonRight, seat.MouseButtonMiddle} {
				if h.IsMousePressed(button) {
					t.Fatalf("Lock motion pressed button %v", button)
				}
			}
			if len(h.KeyboardKeys()) != 0 || h.GetMouseWheel() != 0 {
				t.Fatal("Lock motion supplied keyboard/wheel input")
			}
		})
	}
}

func TestE2ELockNaturalExpiryWaitKeepsOrdinaryInputActive(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_lock_spell.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var wait *ast.FuncLit
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "CheckLockSpell" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 5 {
				return true
			}
			method, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || method.Sel.Name != "addWhen" {
				return true
			}
			name, ok := call.Args[1].(*ast.BinaryExpr)
			if !ok {
				return true
			}
			suffix, ok := name.Y.(*ast.BasicLit)
			if ok && suffix.Value == "\" natural expiry\"" {
				wait, _ = call.Args[3].(*ast.FuncLit)
			}
			return true
		})
	}
	if wait == nil {
		t.Fatal("Lock natural expiry observation is missing")
	}
	motion, invariants, naturalDeadline := false, false, false
	ast.Inspect(wait.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if method, ok := call.Fun.(*ast.SelectorExpr); ok {
				if method.Sel.Name == "observe" && len(call.Args) == 3 {
					queue, ok := call.Args[2].(*ast.Ident)
					motion = motion || ok && queue.Name == "e2eQueueInput"
				}
				invariants = invariants || method.Sel.Name == "unchangedOutsideAndUnits"
			}
		}
		if compare, ok := n.(*ast.BinaryExpr); ok && compare.Op == token.GEQ {
			deadline, ok := compare.Y.(*ast.SelectorExpr)
			naturalDeadline = naturalDeadline || ok && deadline.Sel.Name == "expiry"
		}
		return true
	})
	if !motion || !invariants || !naturalDeadline {
		t.Fatalf("expiry wait must observe actual input and unchanged state/deadline: motion=%t invariants=%t natural=%t", motion, invariants, naturalDeadline)
	}
}
