package opennox

import (
	"context"
	"strings"
	"testing"

	"github.com/opennox/libs/console"
)

// Primary oracle: GAME.EXE's 24-byte command tables at 0059EC58 (root),
// 0059E5C8 (set), 0059E550 (unset), 0059E808 (show), 0059E9A0 (list),
// 0059EAF0 (cheat), 0059E958 (allow), 0059EA90 (log), 0059EA48 (menu),
// 0059EBC8 (macros), 0059EC10 (telnet), and 0059E4C0 (quality).
// Do not mistake frameratelimiter/savedebugcmd for shortened new aliases.
func TestConsoleOriginalCommandInventory(t *testing.T) {
	c, _ := consoleCommandTestConsole(t)
	groups := map[string]string{
		"":      "allow audtest ban bind broadcast cheat clear exec execrul exit gamma help image kick list lock load log macros menu mute quit say set show sysop telnet unset unmute unbind unlock watch window startSoloQuest ques",
		"set":   "armor cycle frameratelimiter god lessons monsters name netdebug ob players quality sage savedebugcmd spell spellpoints staff staffs sysop time weapon weapons team mode",
		"unset": "god frameratelimiter netdebug sage",
		"show":  "bindings game motd rank perfmon extents gui ai info mem netstat mmx seq",
		"list":  "armor maps spells staffs weapons users",
		"cheat": "ability goto health mana level spells gold re-enter",
		"allow": "user IP", "log": "console file stop", "menu": "vidopt options",
		"macros": "on off", "telnet": "on off", "set quality": "modem isdn cable T1 LAN",
	}
	count := 0
	for parent, leaves := range groups {
		for _, leaf := range strings.Fields(leaves) {
			path := strings.TrimSpace(parent + " " + leaf)
			cmd := consoleCommandAt(c, path)
			if cmd == nil || len(cmd.Sub) == 0 && cmd.Func == nil && cmd.LegacyFunc == nil {
				t.Errorf("missing original command implementation: %s", path)
				continue
			}
			count++
			if !cmd.Flags.Has(console.NoHelp) && cmd.HelpID != "" {
				if !c.Exec(console.AsServer(context.Background()), normalizeConsoleInput(c, "HELP "+path)) {
					t.Errorf("help did not resolve original command: %s", path)
				}
			}
		}
	}
	secret := consoleCommandAt(c, console.EncodeSecret("racoiaws"))
	if secret == nil || secret.Func == nil || !secret.Flags.Has(console.Secret) || !secret.Flags.Has(console.NoHelp) {
		t.Error("stock cheat authorization command lost its hidden gate")
	}
	t.Logf("verified %d original paths plus the hidden authorization command", count)
}
