package opennox

import (
	"context"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/opennox/libs/console"
	"github.com/opennox/opennox/v1/legacy"
)

func consoleSetSysop(_ context.Context, c *console.Console, args []string) bool {
	if len(args) != 1 {
		return false
	}
	// GAME.EXE 0040A610 uses 005D5368; the adjacent game password starts
	// at 005D537C. Stock access.wnd entry 10136 also has a nine-unit limit.
	// Reject, rather than silently changing, an oversized password.
	password := args[0]
	if !utf8.ValidString(password) || strings.ContainsRune(password, 0) || len(utf16.Encode([]rune(password))) > 9 {
		c.Print(console.ColorRed, "Sysop password must fit in 9 UTF-16 units without NUL characters.")
		return true
	}
	legacy.Nox_xxx_sysopSetPass_40A610(password)
	consoleMessage(c, "sysoppasswordset")
	return true
}
