//go:build !server

package ail

// Frame synchronization follows the scalar minimp3 decoder by lieff, CC0-1.0.
const (
	mp3HeaderSize          = 4
	mp3MaxFreeFormatBytes  = 2304
	mp3MaxFrameSyncMatches = 10
)

type mp3FrameHeader [mp3HeaderSize]byte

func (h mp3FrameHeader) layer() int {
	return 4 - int(h[1]>>1&3)
}

func (h mp3FrameHeader) valid() bool {
	return h[0] == 0xff && (h[1]&0xf0 == 0xf0 || h[1]&0xfe == 0xe2) &&
		h.layer() != 4 && h[2]>>4 != 15 && h[2]>>2&3 != 3
}

func (h mp3FrameHeader) matches(next mp3FrameHeader) bool {
	return next.valid() && (h[1]^next[1])&0xfe == 0 && (h[2]^next[2])&0x0c == 0 &&
		h.freeFormat() == next.freeFormat()
}

func (h mp3FrameHeader) samples() int {
	if h.layer() == 1 {
		return 384
	}
	if h[1]&14 == 2 {
		return 576
	}
	return 1152
}

// Indexed by MPEG version group, Layer III through I, and the bitrate field.
var mp3HalfBitrates = [2][3][15]uint8{
	{
		{0, 4, 8, 12, 16, 20, 24, 28, 32, 40, 48, 56, 64, 72, 80},
		{0, 4, 8, 12, 16, 20, 24, 28, 32, 40, 48, 56, 64, 72, 80},
		{0, 16, 24, 28, 32, 40, 48, 56, 64, 72, 80, 88, 96, 112, 128},
	},
	{
		{0, 16, 20, 24, 28, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160},
		{0, 16, 24, 28, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192},
		{0, 16, 32, 48, 64, 80, 96, 112, 128, 144, 160, 176, 192, 208, 224},
	},
}

func (h mp3FrameHeader) freeFormat() bool {
	return h[2]>>4 == 0
}

func (h mp3FrameHeader) frameBytes(freeFormatBytes int) int {
	if h.freeFormat() {
		return freeFormatBytes
	}
	mpeg1 := int(h[1] >> 3 & 1)
	mpeg2Or1 := int(h[1] >> 4 & 1)
	sampleRate := [...]int{44100, 48000, 32000}[h[2]>>2&3] >> (1 - mpeg1) >> (1 - mpeg2Or1)
	layer := h.layer()
	bitrate := 2 * int(mp3HalfBitrates[mpeg1][3-layer][h[2]>>4])
	size := h.samples() * bitrate * 125 / sampleRate
	if layer == 1 {
		size &^= 3 // Layer I has four-byte slots.
	}
	return size
}

func (h mp3FrameHeader) padding() int {
	if h[2]&2 == 0 {
		return 0
	}
	if h.layer() == 1 {
		return 4
	}
	return 1
}

type mp3FrameSync struct {
	header          mp3FrameHeader
	freeFormatBytes int
}

// next returns the leading bytes to skip and the length of the next frame.
// Finding a new sequence resets the decoder history, including its reservoir.
func (s *mp3FrameSync) next(data []byte) (skip, size int, reset bool) {
	if len(data) > mp3HeaderSize && s.header[0] == 0xff {
		header := mp3FrameHeader(data)
		if s.header.matches(header) {
			size := header.frameBytes(s.freeFormatBytes) + header.padding()
			if size != 0 && (size == len(data) ||
				size+mp3HeaderSize <= len(data) && header.matches(mp3FrameHeader(data[size:]))) {
				s.header = header
				return 0, size, false
			}
		}
	}
	*s = mp3FrameSync{}
	skip, size, s.freeFormatBytes = findMP3Frame(data)
	if size != 0 {
		s.header = mp3FrameHeader(data[skip:])
	}
	return skip, size, true
}

func findMP3Frame(data []byte) (skip, size, freeFormatBytes int) {
	limit := max(0, len(data)-mp3HeaderSize)
	for i := 0; i < limit; i++ {
		header := mp3FrameHeader(data[i:])
		if !header.valid() {
			continue
		}
		frameBytes := header.frameBytes(0)
		total := frameBytes + header.padding()
		freeBytes := 0
		if frameBytes == 0 {
			total, frameBytes = inferMP3FreeFormatSize(data[i:], header)
			freeBytes = frameBytes
		}
		if (frameBytes != 0 && i+total <= len(data) && matchMP3Frames(data[i:], header, frameBytes)) ||
			(i == 0 && total == len(data)) {
			return i, total, freeBytes
		}
	}
	return limit, 0, 0
}

// A free-format stream encodes its frame size in the distance between headers.
func inferMP3FreeFormatSize(data []byte, header mp3FrameHeader) (total, frameBytes int) {
	for i := mp3HeaderSize; i < mp3MaxFreeFormatBytes && 2*i < len(data)-mp3HeaderSize; i++ {
		next := mp3FrameHeader(data[i:])
		if !header.matches(next) {
			continue
		}
		size := i - header.padding()
		third := i + size + next.padding()
		if third+mp3HeaderSize <= len(data) && header.matches(mp3FrameHeader(data[third:])) && size != 0 {
			return i, size
		}
	}
	return header.padding(), 0
}

func matchMP3Frames(data []byte, first mp3FrameHeader, frameBytes int) bool {
	header := first
	offset := 0
	for n := 0; n < mp3MaxFrameSyncMatches; n++ {
		offset += header.frameBytes(frameBytes) + header.padding()
		if offset+mp3HeaderSize > len(data) {
			return n != 0
		}
		header = mp3FrameHeader(data[offset:])
		if !first.matches(header) {
			return false
		}
	}
	return true
}
