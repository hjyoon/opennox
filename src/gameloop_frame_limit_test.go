package opennox

import (
	"testing"
	"time"

	"github.com/opennox/libs/platform"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

// Freeze the platform independently of the actual Go timer. A passing wait
// must consume real wall time through the production LoopSleep, not advance
// an injected simulation clock or server frame.
type frameLimitTestClock struct {
	platform.Platform
	now   time.Duration
	reads int
}

func (p *frameLimitTestClock) Ticks() time.Duration {
	p.reads++
	return p.now
}

func frameLimitTestClient(t *testing.T) (*Client, *frameLimitTestClock) {
	t.Helper()
	consoleCommandTestFlags(t)
	previousServer, previousPlatform := noxServer, platform.Get()
	previousEnabled := useFrameLimit
	previousTicks, previousFrame := nox_gameTicks_371764, nox_gameFrame_371772
	t.Cleanup(func() {
		noxServer = previousServer
		platform.Set(previousPlatform)
		useFrameLimit = previousEnabled
		nox_gameTicks_371764, nox_gameFrame_371772 = previousTicks, previousFrame
	})
	p := &frameLimitTestClock{Platform: previousPlatform, now: 9 * time.Second}
	platform.Set(p)
	s := &Server{Server: &server.Server{}}
	s.SetFrame(480)
	noxServer = s
	useFrameLimit = true
	nox_gameFrame_371772 = s.Frame()
	nox_gameTicks_371764 = platformTicks() - 500
	noxflags.SetGame(noxflags.GameHost | noxflags.GameClient | noxflags.GameFlag29)
	return &Client{srv: s}, p
}

func frameLimitTestWait(t *testing.T, c *Client) {
	t.Helper()
	frame, ticks, anchor := c.srv.Frame(), nox_gameTicks_371764, nox_gameFrame_371772
	c.srv.SetRateLimit(30)
	started := time.Now()
	c.mainloopFrameLimit()
	elapsed := time.Since(started)
	if elapsed < 30*time.Millisecond {
		t.Fatalf("30 FPS deadline was bypassed: waited %s with frozen game frame %d", elapsed, frame)
	}
	if c.srv.Frame() != frame || nox_gameTicks_371764 != ticks || nox_gameFrame_371772 != anchor {
		t.Fatal("frame limiter advanced simulation or reset the synchronization anchor")
	}
	t.Logf("production frame limiter waited %s; game frame remains %d", elapsed, frame)
}

func TestMainloopFrameLimitGamePauseWithFrozenFrame(t *testing.T) {
	for _, enginePause := range []bool{false, true} {
		name := "rendering"
		if enginePause {
			name = "stale_render_catchup"
		}
		t.Run(name, func(t *testing.T) {
			c, _ := frameLimitTestClient(t)
			noxflags.SetGame(noxflags.GamePause)
			if enginePause {
				noxflags.SetEngine(noxflags.EnginePause)
			}
			if dt := nox_ticks_getNext(); dt >= 0 {
				t.Fatalf("fixture must reproduce an expired simulation deadline: %s", dt)
			}
			game, engine := noxflags.GetGame(), noxflags.GetEngine()
			frameLimitTestWait(t, c)
			if noxflags.GetGame() != game || noxflags.GetEngine() != engine {
				t.Fatal("frame pacing changed gameplay/render pause flags")
			}
		})
	}
}

func TestMainloopFrameLimitGamePauseEveryIteration(t *testing.T) {
	c, _ := frameLimitTestClient(t)
	noxflags.SetGame(noxflags.GamePause)
	for i := 0; i < 4; i++ {
		frameLimitTestWait(t, c)
	}
}

func TestMainloopFrameLimitGamePauseResumePreservesCatchup(t *testing.T) {
	c, clock := frameLimitTestClient(t)
	noxflags.SetGame(noxflags.GamePause)
	frameLimitTestWait(t, c)
	noxflags.UnsetGame(noxflags.GamePause)
	noxflags.SetEngine(noxflags.EnginePause)
	reads := clock.reads
	c.mainloopFrameLimit()
	if clock.reads != reads {
		t.Fatal("unpaused render catch-up did not keep its original no-wait return")
	}
	noxflags.UnsetEngine(noxflags.EnginePause)
	if dt := nox_ticks_getNext(); dt >= 0 {
		t.Fatal("resume fixture unexpectedly reset the expired simulation deadline")
	}
	c.mainloopFrameLimit()
}

func TestMainloopFrameLimitGamePauseHonorsDisabledLimiter(t *testing.T) {
	_, clock := frameLimitTestClient(t)
	noxflags.SetGame(noxflags.GamePause)
	useFrameLimit = false
	noxServer = nil
	reads := clock.reads
	// A disabled limiter must not touch either a client or the server timer.
	var c *Client
	c.mainloopFrameLimit()
	if clock.reads != reads {
		t.Fatal("disabled limiter accessed the clock")
	}
}

func TestMainloopFrameLimitOtherTopologiesKeepRateWait(t *testing.T) {
	for _, name := range []string{"server_only", "client_only", "ordinary_host_client", "no_rendering"} {
		t.Run(name, func(t *testing.T) {
			c, _ := frameLimitTestClient(t)
			noxflags.SetGame(noxflags.GamePause)
			switch name {
			case "server_only":
				noxflags.UnsetGame(noxflags.GameClient)
			case "client_only":
				noxflags.UnsetGame(noxflags.GameHost)
			case "ordinary_host_client":
				noxflags.UnsetGame(noxflags.GameFlag29)
			case "no_rendering":
				noxflags.SetEngine(noxflags.EngineNoRendering)
			}
			frameLimitTestWait(t, c)
		})
	}
}

func TestMainloopFrameLimitUnpausedSimulationDeadline(t *testing.T) {
	c, _ := frameLimitTestClient(t)
	nox_gameTicks_371764 = platformTicks()
	nox_gameFrame_371772 = c.srv.Frame() - 1
	if dt := nox_ticks_getNext(); dt != 33*time.Millisecond {
		t.Fatalf("original one-frame synchronization deadline changed: %s", dt)
	}
	frameLimitTestWait(t, c)
}
