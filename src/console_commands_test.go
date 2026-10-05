package opennox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/console"
	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type consoleCommandCapture struct{ lines []string }

func (p *consoleCommandCapture) Print(_ console.Color, text string) { p.lines = append(p.lines, text) }
func (p *consoleCommandCapture) Printf(cl console.Color, format string, args ...interface{}) {
	p.Print(cl, fmt.Sprintf(format, args...))
}

func cloneConsoleCommand(cmd *console.Command) *console.Command {
	out := *cmd
	out.Sub = nil
	for _, sub := range cmd.Sub {
		out.Sub = append(out.Sub, cloneConsoleCommand(sub))
	}
	return &out
}

func consoleCommandTestConsole(t *testing.T) (*console.Console, *consoleCommandCapture) {
	t.Helper()
	p := new(consoleCommandCapture)
	c := console.New(p)
	for _, cmd := range noxConsole.Commands() {
		if existing := consoleCommandAt(c, cmd.Token); existing != nil {
			*existing = *cloneConsoleCommand(cmd)
		} else {
			c.Register(cloneConsoleCommand(cmd))
		}
	}
	restoreConsoleCommands(c)
	c.Localize(strman.New(), "on", "off", "respawn", "all")
	c.SetCheats(true)
	return c, p
}

func consoleCommandTestFlags(t *testing.T) {
	t.Helper()
	game, engine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.UnsetGame(game)
	noxflags.UnsetEngine(engine)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.SetGame(game)
		noxflags.SetEngine(engine)
	})
}

func consoleCommandTestExec(c *console.Console, input string) bool {
	return c.Exec(console.AsServer(context.Background()), normalizeConsoleInput(c, input))
}

func TestConsoleRestoredLeavesKeepPermissions(t *testing.T) {
	c, _ := consoleCommandTestConsole(t)
	for _, path := range []string{"set name", "set monsters", "set spell", "set mode", "mute", "unmute", "list users", "list maps", "list spells", "bind", "log file", "log stop"} {
		before, after := consoleCommandAt(noxConsole, path), consoleCommandAt(c, path)
		if before == nil || after == nil || after.Func == nil || after.LegacyFunc != nil {
			t.Fatalf("missing native handler for %s", path)
		}
		if before.Flags != after.Flags || before.HelpID != after.HelpID {
			t.Fatalf("metadata changed for %s", path)
		}
	}
	if cmd := consoleCommandAt(c, "show mmx"); cmd == nil || cmd.Flags != console.ClientServer || cmd.Func == nil {
		t.Fatal("missing original show mmx")
	}
}

func TestConsoleSetNameAndQuotedMacro(t *testing.T) {
	c, _ := consoleCommandTestConsole(t)
	oldName := legacy.Nox_xxx_serverOptionsGetServername_40A4C0()
	t.Cleanup(func() { legacy.Nox_xxx_gameSetServername_40A440(oldName) })
	for _, name := range []string{"Red Dragon", "한글 서버", strings.Repeat("한", 10), "abcdefghijklmnop"} {
		if !consoleCommandTestExec(c, `SET NAME "`+name+`"`) {
			t.Fatal("name command was rejected")
		}
		got := legacy.Nox_xxx_serverOptionsGetServername_40A4C0()
		if !utf8.ValidString(got) || len(got) > 15 || !strings.HasPrefix(name, got) {
			t.Fatalf("unsafe server name: %q from %q", got, name)
		}
	}
	c.SetExec(func(_ context.Context, input string) bool { return consoleCommandTestExec(c, input) })
	if !consoleCommandTestExec(c, `BIND f2 SET NAME "Red Dragon"`) || !c.ExecMacros(context.Background(), keybind.KeyByName("F2")) {
		t.Fatal("quoted macro failed")
	}
	if got := legacy.Nox_xxx_serverOptionsGetServername_40A4C0(); got != "Red Dragon" {
		t.Fatalf("macro lost name quoting: %q", got)
	}
	if !consoleCommandTestExec(c, "UNBIND f2") || c.ExecMacros(context.Background(), keybind.KeyByName("F2")) {
		t.Fatal("macro was not removed")
	}
}

func TestConsoleSetMonstersAndOriginalModeFlags(t *testing.T) {
	consoleCommandTestFlags(t)
	c, _ := consoleCommandTestConsole(t)
	oldSub := legacy.Nox_xxx_getServerSubFlags_409E60()
	t.Cleanup(func() {
		legacy.Sub_409EC0(4 | 8)
		legacy.Sub_409E70(int(oldSub & (4 | 8)))
	})
	for _, tt := range []struct {
		input string
		mask  uint32
		on    bool
	}{
		{"SET MONSTERS ON", 4, true}, {"set monsters off", 4, false},
		{"set monsters respawn on", 8, true}, {"set monsters respawn off", 8, true},
	} {
		if !consoleCommandTestExec(c, tt.input) || (legacy.Nox_xxx_getServerSubFlags_409E60()&tt.mask != 0) != tt.on {
			t.Fatalf("%s: flags %#x", tt.input, legacy.Nox_xxx_getServerSubFlags_409E60())
		}
	}
	if consoleSetMonsters(context.Background(), c, []string{"respawn"}) || consoleSetMonsters(context.Background(), c, []string{"invalid"}) {
		t.Fatal("invalid monsters arguments accepted")
	}
	settings := getSettings2ByInd(1)
	oldMode := settings.Field52
	t.Cleanup(func() { settings.Field52 = oldMode })
	for name, mode := range map[string]uint16{"Arena": 0x100, "Elimination": 0x400, "CTF": 0x20, "KOTR": 0x10, "Flagball": 0x40, "Chat": 0x4000, "Coop": 0x8000} {
		settings.Field52 = 0xFFFF
		if !consoleSetMode(context.Background(), c, []string{strings.ToUpper(name)}) || settings.Field52 != 0xE80F|mode {
			t.Fatalf("mode %s: %#x", name, settings.Field52)
		}
	}
}

func TestConsoleMuteUsesNativePlayersAndBothMasks(t *testing.T) {
	consoleCommandTestFlags(t)
	noxflags.SetGame(noxflags.GameClient)
	c, output := consoleCommandTestConsole(t)
	s := NewServer(nil, nil, strman.New())
	oldServer := noxServer
	noxServer = s
	t.Cleanup(func() { s.Close(); noxServer = oldServer })
	p := s.Players.ByIndRaw(1)
	p.Active, p.PlayerInd = 1, 1
	p.SetName("Red Dragon")
	p.Field3680 = 0x100
	for _, tt := range []struct {
		input string
		want  uint32
	}{
		{`MUTE "red dragon"`, 0x108}, {`MUTE ALL "Red Dragon"`, 0x10C},
		{`UNMUTE "Red Dragon"`, 0x104}, {`UNMUTE ALL "Red Dragon"`, 0x100},
	} {
		if !consoleCommandTestExec(c, tt.input) || p.Field3680 != tt.want {
			t.Fatalf("%s: status %#x want %#x", tt.input, p.Field3680, tt.want)
		}
	}
	host := s.Players.ByIndRaw(server.HostPlayerIndex)
	host.Active, host.PlayerInd = 1, server.HostPlayerIndex
	host.SetName("Host")
	consoleCommandTestExec(c, "mute Host")
	if host.Field3680&8 != 0 {
		t.Fatal("host muted itself locally")
	}
	consoleCommandTestExec(c, "mute all Host")
	if host.Field3680&4 == 0 {
		t.Fatal("global mute did not include host")
	}
	consoleCommandTestExec(c, "list users")
	if got := strings.Join(output.lines, "\n"); !strings.Contains(got, "Red Dragon") || !strings.Contains(got, "Host") {
		t.Fatalf("native users missing: %q", got)
	}
}

func TestConsoleSetSpellAndSparseList(t *testing.T) {
	consoleCommandTestFlags(t)
	c, output := consoleCommandTestConsole(t)
	s := NewServer(nil, nil, strman.New())
	oldServer := noxServer
	noxServer = s
	t.Cleanup(func() { s.Close(); noxServer = oldServer })
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	defs := []things.Spell{{ID: "SPELL_MAGIC_MISSILE", ManaCost: 5, Title: "MagicTitle"}, {ID: spell.ID(132).String(), ManaCost: 9, Title: "BallTitle"}}
	if err := things.WriteSpellsYAML(filepath.Join(dir, "spells.yml"), defs); err != nil {
		t.Fatal(err)
	}
	data, _ := alloc.Make([]byte{}, 4) // An empty original SPEL section, skipped in favour of YAML.
	f := binfile.NewMemFile(unsafe.Pointer(&data[0]), len(data))
	t.Cleanup(f.Free)
	if err := s.Spells.Read(f, nil, false); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{`set spell SPELL_MAGIC_MISSILE OFF`, `set spell MagicTitle ON`, `set spell spell_magic_missile off`} {
		if !consoleCommandTestExec(c, input) {
			t.Fatalf("rejected %q", input)
		}
	}
	if s.Spells.DefByInd(spell.SPELL_MAGIC_MISSILE).Enabled {
		t.Fatal("spell disable did not apply")
	}
	noxflags.SetGame(noxflags.GameModeChat)
	consoleCommandTestExec(c, "set spell SPELL_MAGIC_MISSILE on")
	if s.Spells.DefByInd(spell.SPELL_MAGIC_MISSILE).Enabled {
		t.Fatal("chat-only gate bypassed")
	}
	noxflags.UnsetGame(noxflags.GameModeChat)
	s.Spells.Enable(132, false)
	noxflags.SetGame(noxflags.GameModeFlagBall)
	consoleCommandTestExec(c, "set spell "+spell.ID(132).String()+" on")
	if s.Spells.DefByInd(132).Enabled {
		t.Fatal("flagball forbidden spell enabled")
	}
	consoleCommandTestExec(c, "list spells")
	if got := strings.Join(output.lines, "\n"); !strings.Contains(got, "132\t"+spell.ID(132).String()) {
		t.Fatalf("sparse list used enumeration instead of spell ID: %q", got)
	}
}

func TestConsoleLogFileStopAndMMX(t *testing.T) {
	consoleCommandTestFlags(t)
	c, output := consoleCommandTestConsole(t)
	if logFile != nil {
		t.Fatal("test requires an unconfigured log file")
	}
	t.Cleanup(closeLog)
	path := filepath.Join(t.TempDir(), "console log.txt")
	if err := os.WriteFile(path, []byte("old log\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if !consoleCommandTestExec(c, `LOG FILE "`+path+`"`) || !noxflags.HasEngine(noxflags.EngineLogToFile) {
		t.Fatal("file logging not enabled")
	}
	_, _ = logBuf.WriteString("new log\n")
	noxflags.SetEngine(noxflags.EngineLogToConsole)
	if !consoleCommandTestExec(c, "LOG STOP") || logFile != nil || noxflags.HasEngine(noxflags.EngineLogToFile|noxflags.EngineLogToConsole) {
		t.Fatal("log stop failed")
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "old log\nnew log\n" {
		t.Fatalf("log lost content: %q %v", got, err)
	}
	if consoleLogFile(context.Background(), c, nil) || consoleLogStop(context.Background(), c, []string{"invalid"}) {
		t.Fatal("log arity")
	}
	mmx := memmap.PtrUint32(0x5D4594, 805836)
	oldMMX := *mmx
	t.Cleanup(func() { *mmx = oldMMX })
	for _, enabled := range []uint32{0, 1} {
		*mmx = enabled
		if !consoleCommandTestExec(c, "SHOW MMX") {
			t.Fatal("mmx command missing")
		}
		id := strman.ID("MMXNotEnabled")
		if enabled != 0 {
			id = "MMXEnabled"
		}
		if got := output.lines[len(output.lines)-1]; got != c.Strings().GetStringInFile(id, "parsecmd.c") {
			t.Fatalf("MMX status = %q", got)
		}
	}
}
