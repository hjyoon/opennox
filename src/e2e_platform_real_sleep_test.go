package opennox

import (
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestE2ERealSleepOptIn(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  string
		real bool
	}{
		{name: "unset"},
		{name: "false", env: "false"},
		{name: "true", env: "true", real: true},
		{name: "uppercase", env: "TRUE"},
		{name: "numeric", env: "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NOX_E2E_REAL_SLEEP", tc.env)
			if tc.name == "unset" {
				if err := os.Unsetenv("NOX_E2E_REAL_SLEEP"); err != nil {
					t.Fatal(err)
				}
			}
			p := newPlayformE2E()
			if got := p.realSleep != nil; got != tc.real {
				t.Fatalf("real Sleep selected=%v, want %v", got, tc.real)
			}
			if got, want := p.RandInt(), rand.New(rand.NewSource(1234)).Int(); got != want {
				t.Fatalf("initial random value=%d, want unchanged seed value %d", got, want)
			}
			var slept []time.Duration
			if p.realSleep != nil {
				p.realSleep = func(dt time.Duration) { slept = append(slept, dt) }
			}
			for _, dt := range []time.Duration{17 * time.Millisecond, 0, -3 * time.Millisecond} {
				p.Sleep(dt)
			}
			if got, want := p.Ticks(), 14*time.Millisecond; got != want {
				t.Fatalf("virtual ticks=%v, want %v", got, want)
			}
			if !tc.real {
				return
			}
			want := []time.Duration{17 * time.Millisecond, 0, -3 * time.Millisecond}
			if len(slept) != len(want) {
				t.Fatalf("real Sleep calls=%v, want %v", slept, want)
			}
			for i := range want {
				if slept[i] != want[i] {
					t.Fatalf("real Sleep call %d=%v, want %v", i, slept[i], want[i])
				}
			}
		})
	}
}

func TestE2ERealSleepReleasesPlatformLock(t *testing.T) {
	t.Setenv("NOX_E2E_REAL_SLEEP", "true")
	p := newPlayformE2E()
	entered := make(chan time.Duration, 1)
	release := make(chan struct{})
	done := make(chan struct{})
	p.realSleep = func(dt time.Duration) {
		entered <- dt
		<-release
	}
	go func() {
		defer close(done)
		p.Sleep(30 * time.Millisecond)
	}()
	defer func() { close(release); <-done }()
	select {
	case dt := <-entered:
		if dt != 30*time.Millisecond {
			t.Fatalf("real Sleep duration=%v", dt)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("real Sleep was not selected")
	}
	lockedOperations := make(chan time.Duration, 1)
	go func() {
		ticks := p.Ticks()
		p.tick(time.Millisecond)
		p.RandSeedTime()
		p.RandInt()
		lockedOperations <- ticks
	}()
	select {
	case ticks := <-lockedOperations:
		if ticks != 30*time.Millisecond {
			t.Fatalf("virtual ticks before blocking real Sleep=%v", ticks)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("platform clock or RNG locked during real Sleep")
	}
	select {
	case <-done:
		t.Fatal("blocking real Sleep returned before release")
	default:
	}
}
