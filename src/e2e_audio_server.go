//go:build server

package opennox

import "errors"

func (sc *e2eScenario) AssertOpenALPlayback(_ bool, name string) {
	sc.add(0, name, func() {
		e2eError(errors.New("OpenAL playback is unavailable in a server-only build"))
	})
}
