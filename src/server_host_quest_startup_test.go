package opennox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

// Exercise the actual native mode words and C setters, not a fake startup
// service. The real-process probe separately verifies the Quest transition.
func TestDedicatedHostQuestSelectionAndFreshHostResets(t *testing.T) {
	for _, quest := range []bool{false, true} {
		for _, stale := range []uint32{0, 1, 0xffffffff} {
			t.Run(fmt.Sprintf("quest_%t/stale_%08x", quest, stale), func(t *testing.T) {
				consoleCommandTestFlags(t)
				oldJoin, oldResume := legacy.Get_nox_game_createOrJoin_815048(), questFlag_1556156
				offsets := []uintptr{1049688, 1556104, 1556152, 1556160, 1556164}
				oldWords := make([]uint32, len(offsets))
				for i, off := range offsets {
					oldWords[i] = memmap.Uint32(0x5D4594, off)
				}
				t.Cleanup(func() {
					legacy.Set_nox_game_createOrJoin_815048(oldJoin)
					questFlag_1556156 = oldResume
					for i, off := range offsets {
						*memmap.PtrUint32(0x5D4594, off) = oldWords[i]
					}
				})
				for _, off := range offsets {
					*memmap.PtrUint32(0x5D4594, off) = stale
				}
				legacy.Set_nox_game_createOrJoin_815048(0)
				questFlag_1556156 = true
				base := noxflags.GameHost | noxflags.GameClient | noxflags.GameOnline |
					noxflags.GameModeCTF | noxflags.GameModeCoop | noxflags.GameNotQuest | noxflags.GameFlag26
				if stale == 0 {
					base &^= noxflags.GameNotQuest
				}
				noxflags.SetGame(base)
				noxflags.SetEngine(noxflags.EngineNoRendering | noxflags.EngineAdmin | noxflags.EngineGodMode)
				engine := noxflags.GetEngine()
				s := new(Server)
				s.setQuestFlag(13)
				s.setQuestMapName("pending-map")
				s.initDedicatedHostQuest(quest)
				want := base &^ noxflags.GameModeCoop
				if quest {
					want &^= noxflags.GameNotQuest
				} else {
					want |= noxflags.GameNotQuest
				}
				if noxflags.GetGame() != want || noxflags.GetEngine() != engine {
					t.Fatalf("host mode changed unrelated flags: game=%v want=%v engine=%v", noxflags.GetGame(), want, noxflags.GetEngine())
				}
				if legacy.Get_nox_game_createOrJoin_815048() != 1 || s.nox_xxx_isQuest_4D6F50() != quest || memmap.Uint32(0x5D4594, 1556160) != uint32(bool2int(quest)) {
					t.Fatal("dedicated entry did not establish the selected host Quest mode")
				}
				if questFlag_1556156 || memmap.Uint32(0x5D4594, 1049688) != 0 || memmap.Uint32(0x5D4594, 1556104) != 0 || memmap.Uint32(0x5D4594, 1556164) != 0 {
					t.Fatal("fresh-host reset differs from startServer")
				}
				if memmap.Uint32(0x5D4594, 1556152) != stale || s.getQuestFlag() != 13 || s.getQuestMapName() != "pending-map" {
					t.Fatal("mode selection overwrote the separate transition request or saved closed state")
				}
			})
		}
	}
}

func TestDedicatedHostQuestConfiguredBeforeSessionAndStartupCommands(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "gameloop.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "MainLoop" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("dedicated entry missing")
	}
	var setup, session, commands, ticks token.Pos
	var count int
	ast.Inspect(body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
				if selector.Sel.Name == "initDedicatedHostQuest" {
					setup, count = call.Pos(), count+1
					if len(call.Args) != 1 {
						t.Fatal("Quest CLI intent must be passed explicitly")
					}
					arg, ok := call.Args[0].(*ast.Ident)
					if !ok || arg.Name != "isServerQuest" {
						t.Fatal("dedicated entry does not consume the parsed -autoquest option")
					}
				}
				if selector.Sel.Name == "newSession" {
					session = call.Pos()
				}
			}
			if fn, ok := call.Fun.(*ast.Ident); ok && fn.Name == "execGameStartupCommands" {
				commands = call.Pos()
			}
		}
		if loop, ok := node.(*ast.ForStmt); ok {
			ticks = loop.Pos()
		}
		return true
	})
	if count != 1 || setup == token.NoPos || session == token.NoPos || commands == token.NoPos || ticks == token.NoPos || setup >= session || setup >= commands || setup >= ticks {
		t.Fatal("selected host Quest mode must be initialized once before accepting players, startup commands, or ticking")
	}
}
