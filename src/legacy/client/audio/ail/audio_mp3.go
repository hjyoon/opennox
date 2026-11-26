//go:build !server

package ail

/*
#include <stdint.h>
#define MINIMP3_ONLY_MP3
#define MINIMP3_NO_SIMD
#include "../mp3/minimp3.h"
#cgo noescape mp3dec_unpack_layer3
#cgo noescape mp3dec_synthesize_layer3
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
	synthesis  C.mp3dec_synthesis_t
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
			d.synthesis = C.mp3dec_synthesis_t{}
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
	var granules [4]C.mp3dec_granule_t
	for i := range granules {
		src, dst := &info.granules[i], &granules[i]
		for n, width := range src.bandWidths {
			dst.sfbtab[n] = C.uint8_t(width)
		}
		dst.part_23_length = C.uint16_t(src.dataBits)
		dst.big_values = C.uint16_t(src.bigValues)
		dst.scalefac_compress = C.uint16_t(src.scalefacCompress)
		dst.global_gain = C.uint8_t(src.globalGain)
		dst.block_type = C.uint8_t(src.blockType)
		dst.mixed_block_flag = C.uint8_t(src.mixedBlock)
		dst.n_long_sfb = C.uint8_t(src.longBands)
		dst.n_short_sfb = C.uint8_t(src.shortBands)
		for n := range src.huffmanTable {
			dst.table_select[n] = C.uint8_t(src.huffmanTable[n])
			dst.region_count[n] = C.uint8_t(src.regionBandCount[n])
			dst.subblock_gain[n] = C.uint8_t(src.subblockGain[n])
		}
		dst.preflag = C.uint8_t(src.preflag)
		dst.scalefac_scale = C.uint8_t(src.scalefacScale)
		dst.count1_table = C.uint8_t(src.countOneTable)
		dst.scfsi = C.uint8_t(src.scalefacReuseMask)
	}
	var spectra C.mp3dec_spectral_frame_t
	consumed := C.mp3dec_unpack_layer3((*C.uint8_t)(unsafe.Pointer(&frame[0])), &granules[0],
		(*C.uint8_t)(unsafe.Pointer(&scratch[0])), C.int(len(data)), &spectra)
	spectrum := (*[2][2][mp3SpectralSamples]float32)(unsafe.Pointer(&spectra))
	channels, count := header.channelCount(), header.granuleCount()
	mixedLongSubbands := 2
	if header.bandRateIndex() == 2 {
		mixedLongSubbands = 4
	}
	for g := range count {
		for ch := range channels {
			side := &info.granules[g*channels+ch]
			longSubbands := 0
			if side.mixedBlock != 0 {
				longSubbands = mixedLongSubbands
			}
			mp3Hybrid(&spectrum[g][ch], &d.overlap[ch], side.blockType, longSubbands)
		}
	}
	C.mp3dec_synthesize_layer3(&d.synthesis, &spectra, C.int(count), C.int(channels),
		(*C.int16_t)(unsafe.Pointer(&out[0])))
	d.reservoir.save(data, int(consumed))
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
