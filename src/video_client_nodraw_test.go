//go:build !server

package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestDrawGeneralNoRenderingSkipsMovie(t *testing.T) {
	flags := noxflags.GetEngine()
	drawing := dword_5d4594_1311936
	stackCur := movieFilesStackCur
	hook := func_5d4594_1311924
	arg := *memmap.PtrUint32(0x5D4594, 1311932)
	t.Cleanup(func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(flags)
		dword_5d4594_1311936 = drawing
		movieFilesStackCur = stackCur
		func_5d4594_1311924 = hook
		*memmap.PtrUint32(0x5D4594, 1311932) = arg
	})

	noxflags.SetEngine(noxflags.EngineNoRendering)
	movieFilesStackCur = 1
	func_5d4594_1311924 = nil
	if err := (&Client{}).DrawGeneral(false); err != nil {
		t.Fatal(err)
	}
	if dword_5d4594_1311936 || movieFilesStackCur != 0 {
		t.Fatal("movie state was not cleared in no-rendering mode")
	}
}

func TestDrawAndPresentNoRenderingSkipsInput(t *testing.T) {
	flags := noxflags.GetEngine()
	guiFlag := nox_client_gui_flag_815132
	t.Cleanup(func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(flags)
		nox_client_gui_flag_815132 = guiFlag
	})

	noxflags.SetEngine(noxflags.EngineNoRendering | noxflags.EnginePause)
	nox_client_gui_flag_815132 = 1
	(&Client{}).drawAndPresent()
	if noxflags.HasEngine(noxflags.EnginePause) {
		t.Fatal("GUI animation did not clear pause in no-rendering mode")
	}
}

func TestClientDrawNoRenderingAllowsNilInput(t *testing.T) {
	flags := noxflags.GetEngine()
	t.Cleanup(func() {
		noxflags.ResetEngine()
		noxflags.SetEngine(flags)
	})

	noxflags.SetEngine(noxflags.EngineNoRendering)
	if !(&Client{Client: &client.Client{}}).nox_xxx_client_435F80_draw() {
		t.Fatal("headless draw stopped client update")
	}
}
