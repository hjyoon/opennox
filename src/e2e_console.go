package opennox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/client/gui"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
)

func e2eConsoleInput() string {
	return eventRespStr(guiCon.input.Func94(gui.AsWindowEvent(0x401d, 0, 0)))
}

func e2eConsoleLines() string {
	data := (*gui.ScrollListBoxData)(guiCon.scrollbox.WidgetData)
	if data == nil || data.Items == nil || data.Count > 128 || data.Field_11_0 > data.Count {
		return ""
	}
	var lines []string
	for _, item := range unsafe.Slice(data.Items, int(data.Field_11_0)) {
		text := item.Text[:]
		for i, c := range text {
			if c == 0 {
				text = text[:i]
				break
			}
		}
		lines = append(lines, string(utf16.Decode(text)))
	}
	return strings.Join(lines, "\n")
}

func e2eConsoleFocused() bool {
	return guiCon.root != nil && guiCon.input != nil && guiCon.scrollbox != nil &&
		!guiCon.root.GetFlags().IsHidden() && noxClient.GUI.Focused() == guiCon.input
}

func e2eConsoleHasHelp(path string) bool {
	cmd := consoleCommandAt(noxConsole, path)
	if cmd == nil {
		return false
	}
	help := noxConsole.HelpString(cmd)
	return help != "" && strings.Contains(e2eConsoleLines(), help)
}

func e2eConsoleListsMaps() bool {
	names := legacy.ConsoleMapNames4D09B0()
	if len(names) == 0 {
		return false
	}
	lines := e2eConsoleLines()
	for _, name := range names {
		if !strings.Contains(lines, name+".map") {
			return false
		}
	}
	return true
}

func e2eConsoleListsClass(class object.Class) bool {
	var entries []string
	for _, typ := range noxServer.Types.List() {
		if typ.Class().Has(class) {
			entries = append(entries, fmt.Sprintf("%d\t%s\t", typ.Ind(), typ.ID()))
		}
	}
	if len(entries) == 0 {
		return false
	}
	// The native console retains 128 lines, including the command echo.
	if len(entries) > 127 {
		entries = entries[len(entries)-127:]
	}
	lines := e2eConsoleLines()
	for _, entry := range entries {
		if !strings.Contains(lines, entry) {
			return false
		}
	}
	return true
}

// Commands enter only through seat text/keyboard events. The observer reads
// real entry/scroll widgets and native state; it never calls Exec, injects
// console output, replaces a handler or edits a rendering baseline.
func (sc *e2eScenario) consoleCommand(line string, output string, check func() bool) {
	sc.add(1, "console input ready", func() {
		if !e2eConsoleFocused() || e2eConsoleInput() != "" {
			e2eError(fmt.Errorf("console lacks focused empty input before %q", line))
		}
	})
	sc.Input(0, "type console command", &seat.TextInputEvent{Text: line})
	sc.add(2, "verify console entry", func() {
		if !e2eConsoleFocused() || e2eConsoleInput() != line {
			e2eError(fmt.Errorf("console text event did not reach the entry for %q", line))
		}
	})
	sc.Key(keybind.KeyEnter, "execute console command")
	sc.add(4, "verify console command", func() {
		if !e2eConsoleFocused() || e2eConsoleInput() != "" || check != nil && !check() || output != "" && !strings.Contains(e2eConsoleLines(), output) {
			e2eError(fmt.Errorf("console native state/output/input-clear assertion failed for %q", line))
			return
		}
		e2eLog.Printf("CONSOLE COMMAND PASS: input=%q frame=%d focused=true input-cleared=true", line, noxServer.Frame())
	})
}

func (sc *e2eScenario) CheckConsoleCommands(name string) {
	var gold uint32
	var file string
	sc.add(0, name, func() {
		unit := noxServer.Players.HostUnit()
		if !nox_client_isConnected() || unit == nil || unit.HealthData == nil || unit.ControllingPlayer() == nil || guiCon.root == nil {
			e2eError(fmt.Errorf("console fixture requires a connected native host/player/GUI"))
			return
		}
		gold = unit.ControllingPlayer().GoldVal
		file = filepath.Join(e2e.path, "Console Case Log.txt")
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			e2eError(fmt.Errorf("console log fixture destination is not fresh"))
		}
	})
	sc.Key(keybind.KeyByName("F1"), "open console with F1")
	sc.add(2, "verify F1 console focus", func() {
		if !e2eConsoleFocused() {
			e2eError(fmt.Errorf("F1 failed to open and focus the native console"))
		}
	})
	sc.consoleCommand("RACOIAWS", "", func() bool { return noxConsole.Cheats() })
	sc.consoleCommand("HELP LIST", "", func() bool { return e2eConsoleHasHelp("list maps") })
	sc.consoleCommand("HELP TELNET ON", "", func() bool { return e2eConsoleHasHelp("telnet on") })
	sc.consoleCommand(`SET NAME "한글Console"`, "", func() bool { return legacy.Nox_xxx_serverOptionsGetServername_40A4C0() == "한글Console" })
	for _, setting := range []struct {
		name string
		mask uint32
	}{{"MONSTERS", 4}, {"WEAPONS", 1}, {"STAFFS", 16}} {
		sc.consoleCommand("SET "+setting.name+" OFF", "", func() bool { return legacy.Nox_xxx_getServerSubFlags_409E60()&setting.mask == 0 })
		sc.consoleCommand("SET "+setting.name+" ON", "", func() bool { return legacy.Nox_xxx_getServerSubFlags_409E60()&setting.mask != 0 })
	}
	sc.consoleCommand("SET CYCLE OFF", "", func() bool { return legacy.Sub_4D0D70() == 0 })
	sc.consoleCommand("SET CYCLE ON", "", func() bool { return legacy.Sub_4D0D70() != 0 })
	sc.consoleCommand("SET PLAYERS 7", "", func() bool { return legacy.Nox_xxx_servGetPlrLimit_409FA0() == 7 })
	sc.consoleCommand("SET QUALITY lan", "", func() bool { return legacy.Get_nox_server_connectionType_3596() == 1 })
	sc.consoleCommand("SET FRAMERATELIMITER", "", func() bool { return useFrameLimit })
	sc.consoleCommand("UNSET FRAMERATELIMITER", "", func() bool { return !useFrameLimit })
	// E2E starts with the limiter disabled for its deterministic clock. Both
	// real commands above are verified; leave it restored for the next inputs.
	sc.consoleCommand("SHOW MMX", "", func() bool { return strings.Contains(e2eConsoleLines(), "MMX") })
	sc.consoleCommand("CLEAR", "", func() bool { return e2eConsoleLines() == "" })
	sc.consoleCommand("LIST MAPS", "", e2eConsoleListsMaps)
	sc.consoleCommand("CLEAR", "", func() bool { return e2eConsoleLines() == "" })
	sc.consoleCommand("LIST USERS", "", func() bool { return strings.Contains(e2eConsoleLines(), noxServer.Players.Host().Name()) })
	sc.consoleCommand("CLEAR", "", func() bool { return e2eConsoleLines() == "" })
	sc.consoleCommand("LIST SPELLS", "SPELL_", nil)
	for _, group := range []struct {
		name  string
		class object.Class
	}{{"ARMOR", object.ClassArmor}, {"WEAPONS", object.ClassWeapon}, {"STAFFS", object.ClassWand}} {
		sc.consoleCommand("CLEAR", "", func() bool { return e2eConsoleLines() == "" })
		sc.consoleCommand("LIST "+group.name, "", func() bool { return e2eConsoleListsClass(group.class) })
	}
	sc.consoleCommand("CHEAT GOLD 25", "", func() bool {
		return noxServer.Players.Host().GoldVal == gold+25 && legacy.Nox_client_gold_4674A0() == gold+25
	})
	sc.consoleCommand("CHEAT HEALTH 777", "", func() bool { return noxServer.Players.HostUnit().HealthData.Max == 777 })
	sc.consoleCommand("CHEAT MANA 333", "", func() bool { _, max := noxServer.Players.HostUnit().Mana(); return max == 333 })
	sc.consoleCommand("SET SPELL SPELL_FIREBALL OFF", "", func() bool {
		return noxflags.HasGame(noxflags.GameModeChat) || !noxServer.Spells.DefByInd(spell.SPELL_FIREBALL).Enabled
	})
	sc.consoleCommand("SET SPELL SPELL_FIREBALL ON", "", func() bool {
		return noxflags.HasGame(noxflags.GameModeChat) || noxServer.Spells.DefByInd(spell.SPELL_FIREBALL).Enabled
	})
	sc.consoleCommand("MACROS OFF", "", func() bool { return !noxConsole.Macros() })
	sc.consoleCommand("MACROS ON", "", func() bool { return noxConsole.Macros() })
	sc.consoleCommand(`BIND f2 SET NAME "Macro Proof"`, "", nil)
	sc.consoleCommand("SHOW BINDINGS", "Macro Proof", nil)
	sc.Key(keybind.KeyByName("F1"), "close F1 console before macro")
	sc.add(2, "verify console hidden", func() {
		if !guiCon.root.GetFlags().IsHidden() || noxClient.GUI.Focused() == guiCon.input {
			e2eError(fmt.Errorf("F1 did not hide/release console input"))
		}
	})
	sc.Key(keybind.KeyByName("F2"), "execute bound macro with real F2")
	sc.add(4, "verify real macro effect", func() {
		if legacy.Nox_xxx_serverOptionsGetServername_40A4C0() != "Macro Proof" {
			e2eError(fmt.Errorf("F2 macro failed to preserve and execute quoted command"))
		}
	})
	sc.Key(keybind.KeyByName("F1"), "reopen console")
	sc.Wait(2, "wait for reopened console")
	sc.consoleCommand("UNBIND F2", "", nil)
	sc.consoleCommand("LOG CONSOLE", "", func() bool { return noxflags.HasEngine(noxflags.EngineLogToConsole) })
	// The runtime path is private; construct it when its scheduled input runs.
	sc.add(0, "queue private log command", func() {
		e2e.input = append(e2e.input, e2eQueuedInput{event: &seat.TextInputEvent{Text: `LOG FILE "` + file + `"`}, space: e2eScenarioInputSpace})
	})
	sc.Wait(2, "wait for log command input")
	sc.Key(keybind.KeyEnter, "enable private file log")
	sc.add(4, "verify private log enabled", func() {
		if !noxflags.HasEngine(noxflags.EngineLogToFile) {
			e2eError(fmt.Errorf("log file command did not enable native log flag"))
		}
	})
	sc.consoleCommand("SHOW MMX", "MMX", nil)
	sc.consoleCommand("LOG STOP", "", func() bool {
		data, err := os.ReadFile(file)
		return err == nil && strings.Contains(string(data), "MMX") && !noxflags.HasEngine(noxflags.EngineLogToFile|noxflags.EngineLogToConsole)
	})
	sc.consoleCommand("CLEAR", "", func() bool { return e2eConsoleLines() == "" })
	sc.consoleCommand("HELP SET", "", func() bool { return e2eConsoleHasHelp("set sysop") })
	sc.checkConsoleKeyboardText()
	sc.add(2, "capture actual rendered F1 console", func() {
		path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
		if err != nil {
			e2eError(err)
			return
		}
		e2eLog.Printf("CONSOLE F1 PASS: native-text-and-key-events=true macro-F2=true native-state-and-output=true frame=%d path=%s", noxServer.Frame(), path)
	})
	sc.Key(keybind.KeyByName("F1"), "close verified console")
}
