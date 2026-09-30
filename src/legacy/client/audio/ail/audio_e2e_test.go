//go:build !server

package ail

import "testing"

func TestE2EAudioBackend(t *testing.T) {
	for _, tc := range []struct {
		name    string
		e2e     bool
		backend string
		mock    bool
	}{
		{"normal", false, "", false},
		{"normal ignores mock", false, "mock", false},
		{"normal ignores openal", false, "openal", false},
		{"playback default", true, "", true},
		{"playback mock", true, "mock", true},
		{"playback openal", true, "openal", false},
		{"unknown stays mock", true, "unknown", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := e2eMockAudio(tc.e2e, tc.backend); got != tc.mock {
				t.Fatalf("mock audio = %t, want %t", got, tc.mock)
			}
		})
	}
}
