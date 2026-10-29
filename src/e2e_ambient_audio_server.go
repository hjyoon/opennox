//go:build server

package opennox

import "fmt"

func (sc *e2eScenario) CheckAmbientAudio(_ int, name string) {
	sc.add(0, name, func() { e2eError(fmt.Errorf("ambient playback requires a client build")) })
}
