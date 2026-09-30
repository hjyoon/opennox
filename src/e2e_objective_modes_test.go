package opennox

import (
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestE2EObjectiveGameMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		want noxflags.GameFlag
	}{
		{"ctf", noxflags.GameModeCTF},
		{" FlagBall ", noxflags.GameModeFlagBall},
		{"KOTR", noxflags.GameModeKOTR},
	} {
		if got, err := e2eObjectiveGameMode(tc.name); err != nil || got != tc.want {
			t.Fatalf("mode %q = %s/%v, want %s", tc.name, got, err, tc.want)
		}
	}
	for _, mode := range []string{"", "arena", "ctff"} {
		if _, err := e2eObjectiveGameMode(mode); err == nil {
			t.Fatalf("unexpected success for mode %q", mode)
		}
	}
}
