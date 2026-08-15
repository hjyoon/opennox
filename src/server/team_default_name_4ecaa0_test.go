package server

import (
	"fmt"
	"testing"
)

func TestTeamDefaultName4ECAA0Table(t *testing.T) {
	want := []string{
		"NONE",
		"Team 1",
		"Team 2",
		"Team 3",
		"Team 4",
		"Team 5",
		"Team 6",
		"Team 7",
		"Team 8",
		"Team 9",
		"Team 10",
		"Team 11",
		"Team 12",
		"Team 13",
		"Team 14",
		"Team 15",
	}
	if got := len(teamDefaultNameTable4ECAA0); got != 17 {
		t.Fatalf("table length = %d, want 17", got)
	}
	for i, text := range want {
		entry := teamDefaultNameTable4ECAA0[i]
		if !entry.present || entry.text != text {
			t.Fatalf("entry %d = {%q, %t}, want {%q, true}", i, entry.text, entry.present, text)
		}
	}
	if entry := teamDefaultNameTable4ECAA0[16]; entry.present || entry.text != "" {
		t.Fatalf("nil-sentinel entry = {%q, %t}", entry.text, entry.present)
	}
}

func TestTeamDefaultName4ECAA0AllSignedInputs(t *testing.T) {
	for raw := 0; raw <= 0xff; raw++ {
		index := int8(uint8(raw))
		wantIndex := index
		if wantIndex > teamDefaultNameMax4ECAA0 {
			wantIndex = 0
		}

		loads := 0
		got := teamDefaultName4ECAA0(index, teamDefaultNameHooks4ECAA0[int8]{
			load: func(gotIndex int8) int8 {
				loads++
				if gotIndex != wantIndex {
					t.Fatalf("raw %#02x load index = %d, want %d", raw, gotIndex, wantIndex)
				}
				return gotIndex
			},
		})
		if got != wantIndex {
			t.Fatalf("raw %#02x result = %d, want %d", raw, got, wantIndex)
		}
		if loads != 1 {
			t.Fatalf("raw %#02x load count = %d, want 1", raw, loads)
		}
	}
}

func TestTeamDefaultName4ECAA0SelectionBoundaries(t *testing.T) {
	tests := []struct {
		index int8
		want  int8
	}{
		{index: -128, want: -128},
		{index: -1, want: -1},
		{index: 0, want: 0},
		{index: 15, want: 15},
		{index: 16, want: 16},
		{index: 17, want: 0},
		{index: 127, want: 0},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.index), func(t *testing.T) {
			got := teamDefaultName4ECAA0(tc.index, teamDefaultNameHooks4ECAA0[int8]{
				load: func(index int8) int8 { return index },
			})
			if got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTeamDefaultName4ECAA0LoadFault(t *testing.T) {
	const fault = "table load fault"
	defer func() {
		if got := recover(); got != fault {
			t.Fatalf("panic = %v, want %q", got, fault)
		}
	}()
	teamDefaultName4ECAA0(int8(17), teamDefaultNameHooks4ECAA0[struct{}]{
		load: func(index int8) struct{} {
			if index != 0 {
				t.Fatalf("load index = %d, want 0", index)
			}
			panic(fault)
		},
	})
}
