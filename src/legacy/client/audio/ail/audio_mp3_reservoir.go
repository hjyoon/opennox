//go:build !server

package ail

const mp3ReservoirSize = 511

// MP3 frames can borrow main data from the end of earlier frames.
type mp3BitReservoir struct {
	data [mp3ReservoirSize]byte
	used int
}

func (r *mp3BitReservoir) restore(dst, payload []byte, backReference int) ([]byte, bool) {
	prefix := min(r.used, backReference)
	copy(dst, r.data[r.used-prefix:r.used])
	copy(dst[prefix:], payload)
	return dst[:prefix+len(payload)], prefix == backReference
}

func (r *mp3BitReservoir) save(data []byte, consumedBits int) {
	start := max((consumedBits+7)/8, len(data)-mp3ReservoirSize)
	r.used = copy(r.data[:], data[start:])
}
