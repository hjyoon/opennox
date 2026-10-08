package opennox

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Extracted MMX status (original VA 0x699160). The portable renderer leaves
// it disabled; the corresponding mapped-variable blob bytes remain reserved.
var dword_5d4594_805836 uint32

func consoleCommandAt(c *console.Console, path string) *console.Command {
	commands := c.Commands()
	var found *console.Command
	for _, token := range strings.Fields(path) {
		found = nil
		for _, cmd := range commands {
			if cmd.Token == token {
				found = cmd
				commands = cmd.Sub
				break
			}
		}
		if found == nil {
			return nil
		}
	}
	return found
}

// Replace only the unsafe/absent leaves. Their original help and permission
// flags remain on the registered command objects.
func restoreConsoleCommands(c *console.Console) {
	for path, fn := range map[string]console.CommandFunc{
		"set name": consoleSetName, "set monsters": consoleSetMonsters, "set sysop": consoleSetSysop,
		"set spell": consoleSetSpell, "set mode": consoleDedicatedOnly(consoleSetMode),
		"mute": consoleMute, "unmute": consoleUnmute,
		"list users": consoleListUsers, "list maps": consoleListMaps,
		"list spells": consoleListSpells, "bind": consoleBind,
		"log file": consoleLogFile, "log stop": consoleLogStop,
	} {
		if cmd := consoleCommandAt(c, path); cmd != nil {
			cmd.Func, cmd.LegacyFunc = fn, nil
		}
	}
	// libs/console.AsDedicated stores false in its context. Preserve the
	// original dedicated-only gate explicitly, without modifying that module.
	for _, path := range []string{"set mode", "set team"} {
		if cmd := consoleCommandAt(c, path); cmd != nil && cmd.Flags.Has(console.FlagDedicated) {
			cmd.Flags &^= console.FlagDedicated
			if fn := cmd.LegacyFunc; fn != nil {
				cmd.LegacyFunc = func(ctx context.Context, c *console.Console, ind int, tokens []string) bool {
					if console.IsClient(ctx) || !noxflags.HasEngine(noxflags.EngineNoRendering) {
						return true
					}
					return fn(ctx, c, ind, tokens)
				}
			}
		}
	}
	if show := consoleCommandAt(c, "show"); show != nil && consoleCommandAt(c, "show mmx") == nil {
		show.Register(&console.Command{Token: "mmx", HelpID: "showmmxhelp", Flags: console.ClientServer, Func: consoleShowMMX})
	}
}

func consoleDedicatedOnly(fn console.CommandFunc) console.CommandFunc {
	return func(ctx context.Context, c *console.Console, args []string) bool {
		if console.IsClient(ctx) || !noxflags.HasEngine(noxflags.EngineNoRendering) {
			return true
		}
		return fn(ctx, c, args)
	}
}

func consoleMessage(c *console.Console, id strman.ID, args ...interface{}) {
	format := c.Strings().GetStringInFile(id, "parsecmd.c")
	format = strings.ReplaceAll(format, "%S", "%s")
	c.Printf(console.ColorRed, format, args...)
}

func consoleSetName(_ context.Context, c *console.Console, args []string) bool {
	if len(args) == 0 {
		return false
	}
	name := strings.Join(args, " ")
	if name != "" {
		bounded := name
		if len(bounded) > 15 {
			bounded = bounded[:15]
			for !utf8.ValidString(bounded) {
				bounded = bounded[:len(bounded)-1]
			}
		}
		legacy.Nox_xxx_gameSetServername_40A440(bounded)
		consoleMessage(c, "setgamename", name)
	}
	return true
}

func consoleSetMonsters(_ context.Context, c *console.Console, args []string) bool {
	if len(args) < 1 || len(args) > 2 {
		return false
	}
	mask, message, state := 4, strman.ID("monsters"), args[0]
	respawn := strings.EqualFold(state, "respawn")
	if respawn {
		if len(args) != 2 {
			return false
		}
		mask, message, state = 8, "monsterrespawn", args[1]
	}
	switch strings.ToLower(state) {
	case "on":
		legacy.Sub_409E70(mask)
	case "off":
		if respawn {
			// GAME.EXE 00442A9C also sets bit eight for respawn off.
			legacy.Sub_409E70(mask)
		} else {
			legacy.Sub_409EC0(mask)
		}
	default:
		return false
	}
	legacy.Nox_server_gameSettingsUpdated_40A670()
	consoleMessage(c, message, c.Strings().GetStringInFile(strman.ID("cmd_token:"+strings.ToLower(state)), "parsecmd.c"))
	return true
}

func consoleSetSpell(_ context.Context, c *console.Console, args []string) bool {
	if len(args) != 2 {
		return false
	}
	if noxflags.HasGame(noxflags.GameModeChat) {
		consoleMessage(c, "NotInChat", args[0])
		return true
	}
	if noxServer == nil {
		c.Print(console.ColorRed, "No game server is available.")
		return true
	}
	id := noxServer.Spells.ByTitle(args[0])
	if id == spell.SPELL_INVALID {
		id = spell.ParseID(strings.ToUpper(args[0]))
	}
	sp := noxServer.Spells.DefByInd(id)
	if sp == nil {
		consoleMessage(c, "invalidspell", args[0])
		return false
	}
	var enabled bool
	var message strman.ID
	switch strings.ToLower(args[1]) {
	case "on":
		if id == 132 && (noxflags.HasGame(noxflags.GameModeFlagBall) || getCurrentSettings2().Field52&0x40 != 0) {
			return true
		}
		enabled, message = true, "spellenabled"
	case "off":
		message = "spelldisabled"
	default:
		return false
	}
	if sp.Enabled != enabled {
		noxServer.Spells.Enable(id, enabled)
		legacy.Nox_server_gameSettingsUpdated_40A670()
		consoleMessage(c, message, args[0])
	}
	return true
}

func consoleSetMode(_ context.Context, _ *console.Console, args []string) bool {
	if len(args) != 1 {
		return false
	}
	// The original table is seven (wchar pointer, uint16 flags) pairs. Keeping
	// these values outside the PE32 blob prevents pointers overwriting flags.
	for name, flags := range map[string]uint16{"Arena": 0x100, "Elimination": 0x400, "CTF": 0x20, "KOTR": 0x10, "Flagball": 0x40, "Chat": 0x4000, "Coop": 0x8000} {
		if strings.EqualFold(args[0], name) {
			settings := getSettings2ByInd(1)
			settings.Field52 = settings.Field52&0xE80F | flags
			break
		}
	}
	return true
}

func consoleChangeMute(ctx context.Context, c *console.Console, args []string, muted bool) bool {
	if len(args) < 1 || len(args) > 2 {
		return false
	}
	mask, name := uint32(8), args[0]
	if !console.IsClient(ctx) && strings.EqualFold(name, "all") {
		if len(args) != 2 {
			return false
		}
		mask, name = 4, args[1]
	}
	var found *server.Player
	if noxServer != nil && (mask == 4 || noxflags.HasGame(noxflags.GameClient)) {
		for _, p := range noxServer.Players.List() {
			if strings.EqualFold(p.Name(), name) && (!muted || mask == 4 || p.Index() != server.HostPlayerIndex) {
				found = p
				break
			}
		}
	}
	message := strman.ID("UserNotFound")
	if found != nil {
		if muted {
			noxServer.NeedPlayerStatus4174F0(found, mask)
			message = "Muted"
		} else {
			noxServer.UnsetPlayerStatus417530(found, mask, server.PlayerUnsetStatusRuntime417530{})
			message = "UnMuted"
		}
	}
	consoleMessage(c, message, name)
	return true
}

func consoleMute(ctx context.Context, c *console.Console, args []string) bool {
	return consoleChangeMute(ctx, c, args, true)
}

func consoleUnmute(ctx context.Context, c *console.Console, args []string) bool {
	return consoleChangeMute(ctx, c, args, false)
}

func consoleListUsers(ctx context.Context, c *console.Console, _ []string) bool {
	consoleMessage(c, "userslist")
	if noxServer != nil {
		for _, p := range noxServer.Players.List() {
			name := p.Name()
			if !console.IsClient(ctx) && p.Field3680&4 != 0 {
				name += ", " + c.Strings().GetStringInFile("SysMuted", "parsecmd.c")
			}
			if p.Field3680&8 != 0 {
				name += ", " + c.Strings().GetStringInFile("ClientMuted", "parsecmd.c")
			}
			c.Print(console.ColorRed, name)
		}
	}
	return true
}

func consoleListMaps(_ context.Context, c *console.Console, _ []string) bool {
	names := legacy.ConsoleMapNames4D09B0()
	for i := 0; i < len(names); i += 4 {
		var line strings.Builder
		for j := i; j < len(names) && j < i+4; j++ {
			fmt.Fprintf(&line, "%-20.20s", names[j]+".map")
		}
		c.Print(console.ColorRed, line.String())
	}
	return true
}

func consoleListSpells(_ context.Context, c *console.Console, _ []string) bool {
	if noxServer != nil {
		for _, sp := range noxServer.Spells.Defs() {
			state := "X"
			if !sp.IsValid() {
				state = "INVALID"
			} else if sp.Enabled {
				state = "ok"
			}
			c.Printf(console.ColorRed, "%3d\t%-40.40s\t%-40.40q\tcost: %-3d\t%-8s", sp.ID, sp.ID.String(), sp.Title, sp.Def.ManaCost, state)
		}
	}
	return true
}

func consoleBind(ctx context.Context, c *console.Console, args []string) bool {
	if len(args) < 2 {
		return false
	}
	input := console.CurCommand(ctx)
	tokens := consoleInputTokens(input)
	if len(tokens) < 3 {
		return false
	}
	// Preserve quotes in the bound command, including multi-word names.
	return c.BindKeyByName(args[0], strings.TrimSpace(input[tokens[1].end:]))
}

func consoleLogFile(_ context.Context, c *console.Console, args []string) bool {
	if len(args) != 1 || args[0] == "" {
		return false
	}
	// Append prevents a repeated console command from destroying earlier logs.
	f, err := os.OpenFile(args[0], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		c.Print(console.ColorRed, err.Error())
		return true
	}
	setLogFile(f)
	noxflags.SetEngine(noxflags.EngineLogToFile)
	return true
}

func consoleLogStop(_ context.Context, _ *console.Console, args []string) bool {
	if len(args) != 0 {
		return false
	}
	closeLog()
	noxflags.UnsetEngine(noxflags.EngineLogToFile | noxflags.EngineLogToConsole)
	return true
}

func consoleShowMMX(_ context.Context, c *console.Console, _ []string) bool {
	id := strman.ID("MMXNotEnabled")
	if dword_5d4594_805836 != 0 {
		id = "MMXEnabled"
	}
	consoleMessage(c, id)
	return true
}
