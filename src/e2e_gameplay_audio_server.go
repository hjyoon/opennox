//go:build server

package opennox

import "fmt"

func (sc *e2eScenario) CheckGameplayAudio(_ int, _ string, name string) {
	sc.add(0, name, func() {
		e2eError(fmt.Errorf("gameplay OpenAL observation is unavailable in a server-only build"))
	})
}
