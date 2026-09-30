//go:build !server

package ail

import "github.com/timshannon/go-openal/openal"

// PlaybackStats reports real OpenAL sources and hardware-consumed buffers.
// Mock E2E handles never have a native driver and therefore return zero stats.
type PlaybackStats struct {
	Native                 bool
	SampleSources          int
	StreamSources          int
	PlayingSources         int
	StreamsOpened          uint64
	MusicStreamsOpened     uint64
	DialogStreamsOpened    uint64
	SampleBuffersQueued    uint64
	StreamBuffersQueued    uint64
	DialogBuffersQueued    uint64
	SampleBuffersProcessed uint64
	StreamBuffersProcessed uint64
	DialogBuffersProcessed uint64
}

func (h Driver) PlaybackStats() PlaybackStats {
	d := h.get()
	if d == nil {
		return PlaybackStats{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	stats := d.stats
	stats.Native = true
	for s := d.sampleHead; s != nil; s = s.next {
		stats.SampleSources++
		stats.SampleBuffersProcessed += uint64(s.processedBuffers())
		if s.source.State() == openal.Playing {
			stats.PlayingSources++
		}
	}
	for s := d.streamHead; s != nil; s = s.next {
		stats.StreamSources++
		processed := uint64(s.source.BuffersProcessed())
		stats.StreamBuffersProcessed += processed
		if s.dialog {
			stats.DialogBuffersProcessed += processed
		}
		if s.source.State() == openal.Playing {
			stats.PlayingSources++
		}
	}
	return stats
}
