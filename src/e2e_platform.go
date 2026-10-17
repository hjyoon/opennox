package opennox

import (
	"math/rand"
	"os"
	"sync"
	"time"
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
