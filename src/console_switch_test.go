package opennox

import (
	"context"
	"reflect"
	"testing"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/strman"
)

func TestConsoleSpellAndSageSwitches(t *testing.T) {
	c := console.New(&consoleCapture{})
	var got []string
	fn := func(_ context.Context, _ *console.Console, args []string) bool {
		got = append([]string(nil), args...)
		return true
	}
	c.Register(&console.Command{Token: "cheat", Flags: console.Server, Sub: []*console.Command{
		{Token: "spells", Flags: console.Server, Func: fn}, {Token: "sage", Flags: console.Server, Func: fn},
	}})
	c.Localize(strman.New(), "on", "off", "all")
	for _, tt := range []struct {
		input string
		want  []string
	}{
		{"CHEAT SAGE ON", []string{"true"}},
		{"cheat sage FaLsE", []string{"false"}},
		{"CHEAT SPELLS ALL OFF", []string{"all", "false"}},
		{"cheat spells all 3 ON", []string{"all", "3", "true"}},
		{"CHEAT SPELLS spell_magic_missile OFF", []string{"SPELL_MAGIC_MISSILE", "false"}},
		{"cheat spells magic_missile 3 off", []string{"MAGIC_MISSILE", "3", "false"}},
		{"cheat spells 3 OFF", []string{"3", "false"}},
		{"cheat spells ON", []string{"true"}},
	} {
		c.Exec(console.AsServer(context.Background()), normalizeConsoleInput(c, tt.input))
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("%s: got %#v want %#v", tt.input, got, tt.want)
		}
	}
}
