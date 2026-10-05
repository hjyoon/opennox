package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"
	"unsafe"

	"gopkg.in/yaml.v2"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestE2ENPCDialogRepeatBoundedMouseSchedule(t *testing.T) {
	for _, count := range []int{-2, -1, 0, 1, 2, 3, 100} {
		t.Run(fmt.Sprintf("count_%d", count), func(t *testing.T) {
			var sc e2eScenario
			if count < 1 || count > 2 {
				defer func() {
					if recover() == nil || len(sc.steps) != 0 {
						t.Fatal("invalid count scheduled gameplay")
					}
				}()
				sc.CheckNPCDialogRepeat(count, "invalid")
				return
			}
			sc.CheckNPCDialogRepeat(count, "NPC Repeat")
			if len(sc.steps) != 8+10*count {
				t.Fatalf("steps=%d", len(sc.steps))
			}
			var before time.Duration
			bounded := 0
			for _, step := range sc.steps {
				if step.time < before {
					t.Fatal("mouse steps go backwards")
				}
				before = step.time
				if step.ready == nil {
					continue
				}
				bounded++
				want := time.Duration(30000)
				if strings.HasSuffix(step.name, "stock opening voice") {
					want = 6000
				}
				if strings.HasSuffix(step.name, "native stream") {
					want = 1500
				}
				if strings.HasSuffix(step.name, "actual close") {
					want = 1200
				}
				if step.waitTimeout != want || step.fnc == nil {
					t.Fatalf("unbounded observation: %+v", step)
				}
			}
			if bounded != 3+2*count {
				t.Fatalf("bounded steps=%d", bounded)
			}
		})
	}
}

func TestE2ENPCDialogRepeatPublicScenarioAndDispatch(t *testing.T) {
	const path = "../scripts/e2e/solo-warrior-dialog-repeat.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	checks := 0
	for _, step := range file.Steps {
		if step.Action == "check-npc-dialog-repeat" {
			checks++
			if step.Count != 2 || step.Name == "" {
				t.Fatalf("wrong public Repeat observation: %+v", step)
			}
		}
	}
	if checks != 1 {
		t.Fatalf("Repeat actions=%d", checks)
	}
	var sc e2eScenario
	sc.Load(path)
	streams, finishes, close := 0, 0, 0
	for _, step := range sc.steps {
		if strings.HasSuffix(step.name, " native stream") {
			streams++
		}
		if strings.HasSuffix(step.name, " natural end") {
			finishes++
		}
		if strings.HasSuffix(step.name, " actual close") {
			close++
		}
	}
	if streams != 2 || finishes != 3 || close != 1 {
		t.Fatalf("loaded streams/finishes/close=%d/%d/%d", streams, finishes, close)
	}
}

func TestE2ENPCDialogRepeatObserverDoesNotSupplyResults(t *testing.T) {
	for _, path := range []string{"e2e_npc_dialog_repeat.go", "e2e_npc_dialog_repeat_audio.go", "e2e_npc_dialog_repeat_audio_server.go"} {
		t.Run(path, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			forbidden := map[string]bool{"PlayFile": true, "SetStream": true, "Start": true, "Close": true, "Sub_44D8F0": true, "Sub_44D3A0": true, "SetRaw": true, "SetInterp": true, "SetText": true, "SetHidden": true, "Func94": true, "NPCDialogProc479B00": true, "AddToMsgListCli": true, "ResetByInd": true}
			ast.Inspect(file, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.AssignStmt:
					for _, lhs := range n.Lhs {
						if _, ok := lhs.(*ast.StarExpr); ok {
							t.Error("observer writes native memory")
						}
					}
				case *ast.CallExpr:
					var name string
					switch fun := n.Fun.(type) {
					case *ast.Ident:
						name = fun.Name
					case *ast.SelectorExpr:
						name = fun.Sel.Name
					}
					if forbidden[name] {
						t.Errorf("observer supplies result via %s", name)
					}
				}
				return true
			})
		})
	}
}

func TestE2ENPCDialogRepeatRejectsMockAudio(t *testing.T) {
	t.Setenv("NOX_E2E_AUDIO", "mock")
	if _, err := e2eNPCDialogAudioStats479B00(); err == nil {
		t.Fatal("mock accepted as native playback evidence")
	}
}

func TestE2ENPCDialogRepeatTextSnapshot(t *testing.T) {
	g := gui.New(nil)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 400, 140, nil)
	title := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 80, 20, nil)
	list := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 300, 100, nil)
	title.SetID(3910)
	list.SetID(3901)
	title.DrawData().SetText("선장")
	d, free := alloc.New(gui.ScrollListBoxData{})
	defer free()
	items, freeItems := alloc.Make([]gui.ScrollListBoxItem(nil), 2)
	defer freeItems()
	defer g.DestroyAll()
	alloc.StrCopyZero16(items[0].Text[:], "다시 듣기")
	alloc.StrCopyZero16(items[1].Text[:], "Original dialog text")
	d.Count, d.Field_11_0, d.Items = 2, 2, &items[0]
	list.WidgetData = unsafe.Pointer(d)
	if got, err := e2eNPCDialogText479B00(root); err != nil || got != "선장\n다시 듣기\nOriginal dialog text" {
		t.Fatalf("snapshot=%q err=%v", got, err)
	}
	for _, n := range []uint16{0, 3} {
		d.Field_11_0 = n
		if _, err := e2eNPCDialogText479B00(root); err == nil {
			t.Fatalf("invalid item population %d accepted", n)
		}
	}
	d.Field_11_0, d.Items = 1, nil
	if _, err := e2eNPCDialogText479B00(root); err == nil {
		t.Fatal("nil item storage accepted")
	}
}
