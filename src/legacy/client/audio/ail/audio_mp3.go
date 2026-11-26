//go:build !server

package ail

import (
	"encoding/binary"
	"fmt"
	"io"
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
	synthesis  mp3SynthesisState
	overlap    [2][mp3OverlapSamples]float32
	sync       mp3FrameSync
	reservoir  mp3BitReservoir
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
			clear(d.synthesis[:])
			clear(d.overlap[:])
			d.reservoir.used = 0
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
		frame := data[skip : skip+size]
		info, valid := readMP3SideInfo(frame, header)
		if !valid {
			d.sync.header[0] = 0
			continue
		}
		if d.decodeLayer3(frame, header, &info, out) {
			return header.samples(), true
		}
	}
	return 0, false
}

func (d *mp3Decoder) decodeLayer3(frame []byte, header mp3FrameHeader, info *mp3SideInfo, out []int16) bool {
	var scratch [mp3ReservoirSize + mp3MaxFreeFormatBytes]byte
	data, ready := d.reservoir.restore(scratch[:], frame[info.mainDataOffset:], info.mainDataBegin)
	if !ready {
		d.reservoir.save(data, 0)
		return false
	}
	bits := mp3BitReader{data: data}
	var spectrum mp3SpectralWorkspace
	var intensity [2][39]uint8
	channels, count := header.channelCount(), header.granuleCount()
	for g := range count {
		side := info.granules[g*channels:]
		spectrum.unpack(header, side, &bits, &intensity)
		for ch := range channels {
			mp3Hybrid(spectrum.channel(ch), &d.overlap[ch], side[ch].blockType, side[ch].longSubbands(header))
		}
		d.synthesis.synthesize(spectrum.channel(0), spectrum.channel(1), channels, out[g*channels*mp3SpectralSamples:])
	}
	d.reservoir.save(data, bits.pos)
	return true
}

func (d *mp3Decoder) Seek(position int) {
	d.bufOff = 0
	d.bufSz = 0
	d.sync = mp3FrameSync{}
}

func (d *mp3Decoder) Position() int {
	return 0
}
