//go:build !server

package ail

/*
#include <stdint.h>
#define MINIMP3_ONLY_MP3
#define MINIMP3_NO_SIMD
#include "../mp3/minimp3.h"
*/
import "C"
import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

func newMP3Decoder(r *wavReader, hr io.Reader, wf waveFormat, hdr []byte) (*mp3Decoder, error) {
	if len(hdr) < 14+16 {
		return nil, fmt.Errorf("invalid MP3 header size: %d", len(hdr))
	}
	wf3 := waveFormatMP3{waveFormat: wf}
	wf3.bitsPerSample = binary.LittleEndian.Uint16(hdr[14:16])
	wf3.size = binary.LittleEndian.Uint16(hdr[16:18])
	wf3.id = binary.LittleEndian.Uint16(hdr[18:20])
	wf3.flags = binary.LittleEndian.Uint32(hdr[20:24])
	wf3.blockSize = binary.LittleEndian.Uint16(hdr[24:26])
	wf3.framesPerBlock = binary.LittleEndian.Uint16(hdr[26:28])
	wf3.codecDelay = binary.LittleEndian.Uint16(hdr[28:30])
	d := &mp3Decoder{
		r:          r,
		sampleRate: int(wf.samplesPerSec),
		channel:    int(wf.channels),
	}
	audioLog.Printf("MP3 stream: %q, %d channels", r.Name(), wf.channels)
	return d, nil
}

type waveFormatMP3 struct {
	waveFormat
	bitsPerSample  uint16
	size           uint16
	id             uint16
	flags          uint32
	blockSize      uint16
	framesPerBlock uint16
	codecDelay     uint16
}

type mp3Decoder struct {
	r          *wavReader
	dec        C.mp3dec_t
	sync       mp3FrameSync
	sampleRate int
	channel    int

	bufSz  int
	bufOff int
	buf    [16 * 1024 * 3]byte
}

func (*mp3Decoder) Format() string {
	return "MP3"
}

func (d *mp3Decoder) Channels() int {
	return d.channel
}

func (d *mp3Decoder) SampleRate() int {
	return d.sampleRate
}

func (d *mp3Decoder) Close() {}

func (d *mp3Decoder) Decode(out []int16) (int, bool) {
	if d.bufOff >= len(d.buf)/2 && d.bufSz != 0 {
		copy(d.buf[:], d.buf[d.bufOff:d.bufOff+d.bufSz])
		d.bufOff = 0
	}
	// Synchronization needs enough data to look ahead to the next frame.
	if d.bufSz < len(d.buf) {
		n, err := d.r.Read(d.buf[d.bufOff+d.bufSz:])
		d.bufSz += n
		if err != nil && d.bufSz == 0 {
			return 0, false
		}
	}
	for d.bufSz != 0 {
		data := d.buf[d.bufOff : d.bufOff+d.bufSz]
		skip, size, reset := d.sync.next(data)
		if reset {
			d.dec = C.mp3dec_t{}
		}
		if skip+size == 0 {
			break // An incomplete trailing header cannot advance the stream.
		}
		d.bufOff += skip + size
		d.bufSz -= skip + size
		header := d.sync.header
		if size == 0 || header.layer() != 3 {
			continue
		}
		status := C.mp3dec_decode_layer3(&d.dec, (*C.uint8_t)(unsafe.Pointer(&data[skip])),
			C.int(size), (*C.int16_t)(unsafe.Pointer(&out[0])))
		if status < 0 {
			d.sync.header[0] = 0
		} else if status > 0 {
			return header.samples(), true
		}
	}
	return 0, false
}

func (d *mp3Decoder) Seek(position int) {
	d.bufOff = 0
	d.bufSz = 0
	d.sync = mp3FrameSync{}
}

func (d *mp3Decoder) Position() int {
	return 0
}
