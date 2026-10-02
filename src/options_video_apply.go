package opennox

import "github.com/opennox/libs/env"

// applyOptionsVideoMode applies the existing OpenNox resolution extension.
// Keep the same E2E resize guard as the shell's options close callback.
func applyOptionsVideoMode() {
	if !env.IsE2E() {
		videoUpdateGameMode(guiOptionsRes)
	}
}
