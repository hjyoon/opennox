package opennox

import (
	"context"
	"log/slog"
	"testing"
	"unicode/utf16"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/console"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/input"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

type consoleEntryTestInput struct{}

func (*consoleEntryTestInput) InputTick()                                        {}
func (*consoleEntryTestInput) ReplaceInputs(c seat.InputConfig) seat.InputConfig { return nil }
func (*consoleEntryTestInput) OnInput(func(seat.InputEvent))                     {}
func (*consoleEntryTestInput) SetTextInput(bool)                                 {}

func TestConsoleEnterAfterCommittedTextInput(t *testing.T) {
	consoleCommandTestFlags(t)
	noxflags.SetGame(noxflags.GameHost)
	oldClient, oldConsole := noxClient, noxConsole
	t.Cleanup(func() { noxClient, noxConsole = oldClient, oldConsole })
	g := gui.New(nil)
	defer g.DestroyAll()
	inp := input.New(slog.Default(), &consoleEntryTestInput{}, false, 6)
	noxClient = &Client{Client: &client.Client{Inp: inp, GUI: g}, ctrl: new(CtrlEventHandler)}
	g.SetInput(inp)
	draw := gui.WindowData{Style: gui.StyleEntryField}
	win := gui.NewEntryFieldRaw(g, nil, gui.StatusEnabled, 0, 0, 240, 20, &draw, &gui.EntryFieldData{Field_1040: 128})
	if win == nil {
		t.Fatal("native console entry allocation failed")
	}
	c, _ := consoleCommandTestConsole(t)
	noxConsole = c
	con := &guiConsole{c: c, input: win}
	var calls []string
	c.Register(&console.Command{Token: "probe", Flags: console.ClientServer, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		if len(args) != 1 {
			t.Fatalf("unexpected parsed arguments: %q", args)
		}
		calls = append(calls, args[0])
		return true
	}})
	inp.OnInputString(func(text string) {
		for _, ch := range utf16.Encode([]rune(text)) {
			gui.EntryFieldOnChar(win, ch)
		}
	})
	win.SetFunc93(con.inputProc)
	g.Focus(win)

	for _, key := range []keybind.Key{keybind.KeyEnter, keybind.KeyKpEnter} {
		inp.InputEvent(&seat.TextInputEvent{Text: `PROBE "한글 Name"`})
		data := (*gui.EntryFieldData)(win.WidgetData)
		if data.Field_1044 != 1 {
			t.Fatal("fixture did not reproduce the committed-character latch")
		}
		before := len(calls)
		win.Func93(gui.WindowKeyPress{Key: key, Pressed: false})
		if len(calls) != before {
			t.Fatal("released Enter executed the command")
		}
		win.Func93(gui.WindowKeyPress{Key: key, Pressed: true})
		if len(calls) != before+1 || calls[before] != "한글 Name" || eventRespStr(win.Func94(gui.AsWindowEvent(0x401d, 0, 0))) != "" {
			t.Fatalf("committed SDL text did not execute/clear through Enter %v: calls=%q latch=%d", key, calls, data.Field_1044)
		}
	}

	inp.InputEvent(&seat.TextInputEvent{Text: `PROBE "IME commit"`})
	inp.InputEvent(&seat.TextEditEvent{Text: "한"})
	before := len(calls)
	win.Func93(gui.WindowKeyPress{Key: keybind.KeyEnter, Pressed: true})
	if len(calls) != before || eventRespStr(win.Func94(gui.AsWindowEvent(0x401d, 0, 0))) != `PROBE "IME commit"` {
		t.Fatal("Enter executed or cleared a command while IME composition was active")
	}
	inp.InputEvent(&seat.TextInputEvent{Text: ""}) // ordinary SDL composition completion
	win.Func93(gui.WindowKeyPress{Key: keybind.KeyEnter, Pressed: true})
	if len(calls) != before+1 || calls[before] != "IME commit" {
		t.Fatal("completed IME composition did not permit the next Enter")
	}
}
