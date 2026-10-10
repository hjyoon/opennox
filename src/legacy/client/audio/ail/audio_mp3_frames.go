//go:build !server

package ail

// minimp3 returns frames per channel, while audioDecoder.Decode returns
// interleaved samples. Always decode a complete MP3 frame into private storage:
// callers (including stream refills) may have less than one frame left.
type mp3FrameDecoder struct {
	*mp3Decoder
	position int
	pending  []int16
	pcm      [1152 * 2]int16
	eof      bool
}

// Position counts delivered PCM frames, not the audible OpenAL queue offset.
func (d *mp3FrameDecoder) Position() int {
	return d.position
}

func (d *mp3FrameDecoder) Decode(out []int16) (int, bool) {
	channels := d.Channels()
	if channels < 1 || channels > 2 {
		return 0, false
	}
	if len(out) < channels {
		return 0, !d.eof
	}
	if len(d.pending) == 0 {
		if d.eof {
			return 0, false
		}
		frames, ok := d.mp3Decoder.Decode(d.pcm[:])
		if frames == 0 {
			d.eof = !ok
			return 0, ok
		}
		d.pending = d.pcm[:frames*channels]
	}
	n := min(len(out), len(d.pending))
	n -= n % channels
	copy(out[:n], d.pending[:n])
	d.pending = d.pending[n:]
	d.position += n / channels
	return n, true
}

func (d *mp3FrameDecoder) Seek(position int) {
	d.mp3Decoder.Seek(0)
	d.position = 0
	d.pending = nil
	d.eof = false
	channels := d.Channels()
	if channels < 1 || channels > 2 {
		d.eof = true
		return
	}
	position = max(position, 0)
	// MP3 frames share a bit reservoir and synthesis history. Decode from
	// the WAV data start (rewound by wavReader.Seek), never skip encoded bytes.
	for d.position < position {
		frames, ok := d.mp3Decoder.Decode(d.pcm[:])
		if frames == 0 {
			d.eof = !ok
			return
		}
		skip := min(position-d.position, frames)
		d.position += skip
		d.pending = d.pcm[skip*channels : frames*channels]
	}
}
