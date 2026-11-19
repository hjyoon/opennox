package legacy

import (
	"slices"
	"strings"
	"testing"
)

func TestMapRulePaths4D0550(t *testing.T) {
	nested := "/" + strings.Repeat("campaign-save/", 26) + "war03c"
	for _, test := range []struct {
		name        string
		mapFile     string
		firstExists bool
		valid       bool
		want        []string
	}{
		{
			name:        "directory rules",
			mapFile:     `maps\crypt\crypt.map`,
			firstExists: true,
			valid:       true,
			want:        []string{`maps\crypt\user.rul`},
		},
		{
			name:    "map rules fallback",
			mapFile: `maps\crypt\crypt.map`,
			valid:   true,
			want:    []string{`maps\crypt\user.rul`, `maps\crypt\crypt.rul`},
		},
		{
			name:    "absolute saved map",
			mapFile: nested + ".map",
			valid:   true,
			want:    []string{nested + "user.rul", nested + ".rul"},
		},
		{
			name:    "C string terminator",
			mapFile: "crypt.map\x00ignored",
			valid:   true,
			want:    []string{"cryptuser.rul", "crypt.rul"},
		},
		{name: "short filename", mapFile: "x"},
		{name: "empty filename"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var paths []string
			valid := loadMapRules4D0550(test.mapFile, func(path string) bool {
				paths = append(paths, path)
				return test.firstExists && len(paths) == 1
			})
			if valid != test.valid || !slices.Equal(paths, test.want) {
				t.Fatalf("valid=%t paths=%q, want valid=%t paths=%q", valid, paths, test.valid, test.want)
			}
		})
	}
}
