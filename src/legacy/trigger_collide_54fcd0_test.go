package legacy

import "testing"

func TestTriggerCollideCompatibilityAllowed54FCD0(t *testing.T) {
	tests := []struct {
		name      string
		mapName   string
		callback  string
		candidate string
		want      bool
	}{
		{
			name:      "Con02A Necromancer",
			mapName:   "Con02a.map",
			callback:  "NecroInPosition",
			candidate: "Con02a:Necromancer",
			want:      true,
		},
		{
			name:      "Con02A accidental Tanya2 crossing",
			mapName:   `/maps/Con02A.MAP`,
			callback:  "necroinposition",
			candidate: "Con02a:Tanya2",
			want:      false,
		},
		{
			name:      "Con02A accidental Julie2 crossing",
			mapName:   "Con02a.map",
			callback:  "NecroInPosition",
			candidate: "Con02a:Julie2",
			want:      false,
		},
		{
			name:      "Con02A unrelated actor",
			mapName:   "Con02a.map",
			callback:  "NecroInPosition",
			candidate: "Con02a:Horvath",
			want:      false,
		},
		{
			name:      "Other Con02A trigger",
			mapName:   "Con02a",
			callback:  "M4Trigger",
			candidate: "Tanya2",
			want:      true,
		},
		{
			name:      "Other map",
			mapName:   "Con03a.map",
			callback:  "NecroInPosition",
			candidate: "Tanya2",
			want:      true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := triggerCollideCompatibilityAllowed54FCD0(tc.mapName, tc.callback, tc.candidate); got != tc.want {
				t.Fatalf("triggerCollideCompatibilityAllowed54FCD0(%q, %q, %q) = %t, want %t",
					tc.mapName, tc.callback, tc.candidate, got, tc.want)
			}
		})
	}
}
