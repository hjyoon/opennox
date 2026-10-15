package opennox

import (
	"bytes"
	"encoding/binary"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestDedicatedHostSettingsPackedNativeCopyAndSelection(t *testing.T) {
	consoleCommandTestFlags(t)
	const size = 3 * 58
	raw := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 371380)), size)
	before, oldIndex := bytes.Clone(raw), getCurrentSettings2Ind()
	t.Cleanup(func() { copy(raw, before); setCurrentSettings2Ind(oldIndex) })
	for _, initial := range []int{0, 1, 2, -1} {
		for _, pattern := range []byte{0, 0x5a, 0xff} {
			for _, mode := range []noxflags.GameFlag{noxflags.GameModeChat, noxflags.GameModeCTF, noxflags.GameModeKOTR, noxflags.GameModeQuest} {
				for i := range raw {
					raw[i] = pattern ^ byte(i*13)
				}
				setCurrentSettings2Ind(initial)
				noxflags.ResetGame()
				noxflags.SetGame(noxflags.GameHost | noxflags.GameOnline | mode)
				noxflags.SetEngine(noxflags.EngineNoRendering)
				game, engine := noxflags.GetGame(), noxflags.GetEngine()
				guard := bytes.Clone(raw[116:])
				calls := 0
				initDedicatedHostSettings(func() {
					calls++
					if getCurrentSettings2Ind() != 0 || getCurrentSettings2() != getSettings2ByInd(0) {
						t.Fatal("refresh did not see active slot0")
					}
					// The declared refresh service writes all58 bytes, including the
					// last two bytes that a padded60-byte Go assignment would overrun.
					for i := 0; i < 58; i++ {
						raw[i] = pattern ^ byte(i*17+7)
					}
					binary.LittleEndian.PutUint16(raw[52:], uint16(mode))
				})
				if calls != 1 || getCurrentSettings2Ind() != 1 || getCurrentSettings2() != getSettings2ByInd(1) {
					t.Fatal("dedicated settings not refreshed/copied/selected once")
				}
				if !bytes.Equal(raw[:58], raw[58:116]) || !bytes.Equal(raw[116:], guard) {
					t.Fatal("native packed copy did not preserve exact58-byte boundary")
				}
				if noxflags.GetGame() != game || noxflags.GetEngine() != engine {
					t.Fatal("settings selection changed gameplay or engine flags")
				}
			}
		}
	}
	t.Log("actual packed C copy/native selection: 48 initial-index/mode/guard conditions")
}

func TestDedicatedHostSettingsRefreshFaultKeepsOriginalOrdering(t *testing.T) {
	raw := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 371380)), 116)
	before, oldIndex := bytes.Clone(raw), getCurrentSettings2Ind()
	t.Cleanup(func() { copy(raw, before); setCurrentSettings2Ind(oldIndex) })
	for i := range raw {
		raw[i] = byte(i ^ 0x5a)
	}
	destination := bytes.Clone(raw[58:])
	setCurrentSettings2Ind(1)
	var caught any
	func() {
		defer func() { caught = recover() }()
		initDedicatedHostSettings(func() {
			if getCurrentSettings2Ind() != 0 {
				t.Fatal("refresh was called before selecting slot0")
			}
			panic("controlled refresh fault")
		})
	}()
	if caught != "controlled refresh fault" || getCurrentSettings2Ind() != 0 || !bytes.Equal(raw[58:], destination) {
		t.Fatal("failed refresh still copied or selected slot1")
	}
}

func TestDedicatedHostSettingsPreparedAfterMapBeforeCommandsAndTicks(t *testing.T) {
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
		t.Fatal("dedicated MainLoop missing")
	}
	var initial, post, setup, commands, ticks token.Pos
	count, posts := 0, 0
	ast.Inspect(body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "InitialMapLoad" {
				initial = call.Pos()
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Sub_43AA70" {
				post, posts = call.Pos(), posts+1
			}
			if id, ok := call.Fun.(*ast.Ident); ok {
				if id.Name == "initDedicatedHostSettings" {
					setup = call.Pos()
					count++
					if len(call.Args) != 1 {
						t.Fatal("original refresh callback missing")
					}
					sel, ok := call.Args[0].(*ast.SelectorExpr)
					if !ok || sel.Sel.Name != "Sub_4161E0" {
						t.Fatal("original native refresh must be used")
					}
					pkg, ok := sel.X.(*ast.Ident)
					if !ok || pkg.Name != "legacy" {
						t.Fatal("native refresh provider changed")
					}
				}
				if id.Name == "execGameStartupCommands" {
					commands = call.Pos()
				}
			}
		}
		if loop, ok := node.(*ast.ForStmt); ok {
			ticks = loop.Pos()
		}
		return true
	})
	if count != 1 || posts != 1 || initial == token.NoPos || post <= initial || setup <= post || commands <= setup || ticks <= setup {
		t.Fatal("dedicated settings must be prepared after the original host reset and before commands or ticks")
	}
}
