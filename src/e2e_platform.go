package opennox

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/opennox/libs/env"
	"github.com/opennox/libs/platform"
)

func newPlayformE2E() *platformE2E {
	p := &platformE2E{}
	if os.Getenv("NOX_E2E_REAL_SLEEP") == "true" {
		p.realSleep = time.Sleep
	}
	p.RandSeed(1234)
	return p
}

type platformE2E struct {
	mu        sync.Mutex
	t         time.Duration
	rnd       *rand.Rand
	realSleep func(time.Duration)
}

func (p *platformE2E) ticks() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.t
}

func (p *platformE2E) tick(dt time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.t += dt
}

func (p *platformE2E) Ticks() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.t
}

func (p *platformE2E) Sleep(dt time.Duration) {
	p.mu.Lock()
	p.t += dt
	sleep := p.realSleep
	p.mu.Unlock()
	if sleep != nil {
		sleep(dt)
	}
}

func (p *platformE2E) TimeSeed() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return int64(p.t)
}

func (p *platformE2E) RandInt() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.rnd.Int()
}

func (p *platformE2E) RandSeed(seed int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rnd = rand.New(rand.NewSource(seed))
}

func (p *platformE2E) RandSeedTime() {
	p.RandSeed(p.TimeSeed())
}

// setE2EGameClock keeps the scenario iteration counter separate from the game
// platform in ordinary mode. The production clock, Sleep and random source
// then remain installed, while the headless seat still accepts scripted input.
func setE2EGameClock(p *platformE2E, mode string) error {
	switch mode {
	case "", "virtual":
		platform.Set(p)
	case "ordinary":
		// Keep the production platform initialized before E2E playback.
	default:
		return fmt.Errorf("NOX_E2E_CLOCK must be virtual or ordinary (got %q)", mode)
	}
	return nil
}

// e2eFrameLimitDisabled preserves an explicit user flag in ordinary playback.
// Virtual playback retains its existing uncapped input-iteration behavior.
func e2eFrameLimitDisabled(requested bool) bool {
	return requested || env.IsE2EVirtualTime()
}
