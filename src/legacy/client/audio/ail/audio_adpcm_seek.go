//go:build !server

package ail

// adpcmSeekFrames is the WAV reader's frame-based seek adapter. Every IMA
// block includes its initial predictor frame; seeking past full blocks must
// also count it. Keep the decoded remainder when the requested frame is in
// the middle of a block, without changing the ordinary ADPCM decode path.
func adpcmSeekFrames(d *adpcmDecoder, offset int) audioDecoder {
	d.Seek(0)
	offset = min(max(offset, 0), d.samples)
	if offset == 0 {
		return d
	}
	framesPerBlock := (d.blockSize/d.channels-4)*2 + 1
	blocks := offset / framesPerBlock
	remaining := int64(blocks) * int64(d.blockSize)
	for remaining > 0 {
		size, err := d.r.ChunkSize()
		if err != nil || size == 0 {
			return d
		}
		skip := min(int64(size), remaining)
		d.r.Skip(int(skip))
		remaining -= skip
	}
	d.position = blocks * framesPerBlock
	inBlock := offset - d.position
	if inBlock == 0 {
		return d
	}
	buffer := make([]int16, framesPerBlock*d.channels)
	n, _ := d.Decode(buffer)
	skip := inBlock * d.channels
	if skip >= n {
		return d
	}
	return &adpcmSeekDecoder{adpcmDecoder: d, pending: buffer[skip:n]}
}

// The adapter belongs to wavReader. Its next seek unwraps it first, so a
// rewind cannot carry pending samples from an earlier block into the stream.
type adpcmSeekDecoder struct {
	*adpcmDecoder
	pending []int16
}

func (d *adpcmSeekDecoder) Position() int {
	return d.adpcmDecoder.Position() - len(d.pending)/d.channels
}

func (d *adpcmSeekDecoder) Decode(out []int16) (int, bool) {
	if len(d.pending) == 0 {
		return d.adpcmDecoder.Decode(out)
	}
	n := min(len(out), len(d.pending))
	n -= n % d.channels
	copy(out[:n], d.pending[:n])
	d.pending = d.pending[n:]
	return n, true
}
