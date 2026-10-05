package opennox

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/console"
	"github.com/opennox/libs/strman"
)

type consoleInputToken struct {
	text string
	end  int
}

// Tokenize only the command prefix; raw script/chat payloads are kept verbatim.
func consoleInputTokens(input string) []consoleInputToken {
	var out []consoleInputToken
	for i := 0; i < len(input); {
		r, n := utf8.DecodeRuneInString(input[i:])
		if unicode.IsSpace(r) {
			i += n
			continue
		}
		start := i
		if input[i] == '"' {
			i++
			start = i
			for i < len(input) && input[i] != '"' && input[i] != '\r' && input[i] != '\n' {
				i++
			}
			value := input[start:i]
			if i < len(input) && input[i] == '"' {
				i++
			}
			out = append(out, consoleInputToken{text: value, end: i})
			continue
		}
		for i < len(input) {
			r, n = utf8.DecodeRuneInString(input[i:])
			if unicode.IsSpace(r) {
				break
			}
			i += n
		}
		out = append(out, consoleInputToken{text: input[start:i], end: i})
	}
	return out
}

func consoleInputCommand(commands []*console.Command, token string) *console.Command {
	for _, cmd := range commands {
		if cmd.Flags.Has(console.Secret) {
			if console.EncodeSecret(strings.ToLower(token)) == cmd.Token {
				return cmd
			}
		} else if strings.EqualFold(token, cmd.Token) || (cmd.Alias != "" && strings.EqualFold(token, cmd.Alias)) {
			return cmd
		}
	}
	return nil
}

func consoleInputKeyword(c *console.Console, token string, keywords ...string) string {
	for _, keyword := range keywords {
		if strings.EqualFold(token, keyword) {
			return keyword
		}
		if alias, ok := c.Strings().GetVariantInFile(strman.ID("cmd_token:"+keyword), "parsecmd.c"); ok && strings.EqualFold(token, alias.Str) {
			return keyword
		}
	}
	return token
}

func consoleInputArgument(c *console.Console, path string, index int, token string) string {
	switch path {
	case "bind", "unbind":
		if index == 0 {
			for _, key := range keybind.ListKeys() {
				if strings.EqualFold(token, key.String()) {
					return key.String()
				}
				if alias, ok := c.Strings().GetVariantInFile(key.TitleID(), "parsecmd.c"); ok && strings.EqualFold(token, alias.Str) {
					return key.String()
				}
			}
		}
	case "set cycle", "set weapons", "set staffs":
		if index == 0 {
			return consoleInputKeyword(c, token, "on", "off")
		}
	case "set monsters":
		return consoleInputKeyword(c, token, "on", "off", "respawn")
	case "set armor", "set weapon", "set staff", "set spell":
		if index == 1 {
			return consoleInputKeyword(c, token, "on", "off")
		}
		if path == "set weapon" && index == 0 {
			return consoleInputKeyword(c, token, "respawn")
		}
	case "mute", "unmute":
		if index == 0 {
			return consoleInputKeyword(c, token, "all")
		}
	case "cheat god", "cheat equip.all", "cheat charm.all", "cheat summon.nolimit", "cheat spells", "cheat scrolls", "set maps allow.all":
		if index == 0 {
			switch consoleInputKeyword(c, token, "on", "off") {
			case "on":
				return "true"
			case "off":
				return "false"
			}
		}
	}
	return token
}

// The shared parser is case-sensitive and translates *every* unquoted alias,
// including player names and filenames. Canonicalize command words only, and
// quote ordinary arguments so their spelling cannot be changed by localization.
func normalizeConsoleInput(c *console.Console, input string) string {
	input = strings.TrimSpace(input)
	tokens := consoleInputTokens(input)
	if len(tokens) == 0 {
		return ""
	}
	commands := c.Commands()
	var path []string
	var leaf *console.Command
	i := 0
	for ; i < len(tokens); i++ {
		cmd := consoleInputCommand(commands, tokens[i].text)
		if cmd == nil {
			break
		}
		token := cmd.Token
		if cmd.Flags.Has(console.Secret) {
			token = strings.ToLower(tokens[i].text)
		}
		path = append(path, token)
		leaf = cmd
		if len(path) == 1 && (token == "help" || token == "ques") {
			commands = c.Commands()
			continue
		}
		if len(cmd.Sub) == 0 {
			i++
			break
		}
		commands = cmd.Sub
	}
	if len(path) == 0 {
		return input
	}
	prefix := strings.Join(path, " ")
	if leaf != nil && (leaf.Raw || prefix == "say" || prefix == "sysop" || prefix == "exec") {
		payload := strings.TrimSpace(input[tokens[i-1].end:])
		if payload != "" {
			return prefix + " " + payload
		}
		return prefix
	}
	for n, token := range tokens[i:] {
		value := consoleInputArgument(c, prefix, n, token.text)
		path = append(path, `"`+value+`"`)
	}
	return strings.Join(path, " ")
}
