package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/opennox/libs/ifs"
	"github.com/opennox/libs/log"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
)

const (
	nativeAudioVoiceCount = 16
	nativeSoundDefCount   = 1024
)

var (
	audioEffectsLog   = log.New("audio-effects")
	audioEffectsDebug = os.Getenv("NOX_DEBUG_AUDIO") == "true"
	nativeAudioFX     nativeAudioEffectsState
)

type nativeAudioBankEntry struct {
	name      string
	rate      uint32
	flags     uint32
	blockSize uint32
	data      []byte
}

type nativeAudioBank struct {
	entries map[string]*nativeAudioBankEntry
	data    []byte
}

func loadNativeAudioBank(base string) (*nativeAudioBank, error) {
	if i := strings.LastIndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	idx, err := readNativeAudioFile(base + ".idx")
	if err != nil {
		return nil, err
	}
	bag, err := readNativeAudioFile(base + ".bag")
	if err != nil {
		return nil, err
	}
	return parseNativeAudioBank(idx, bag)
}

func readNativeAudioFile(path string) ([]byte, error) {
	f, err := ifs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return b, nil
}

func parseNativeAudioBank(idx, bag []byte) (*nativeAudioBank, error) {
	if len(idx) < 12 {
		return nil, fmt.Errorf("audio index header is truncated")
	}
	if string(idx[:4]) != "GABA" {
		return nil, fmt.Errorf("unexpected audio index signature %q", idx[:4])
	}
	version := binary.LittleEndian.Uint32(idx[4:8])
	recordSize := uint64(0)
	switch version {
	case 1:
		recordSize = 32
	case 2:
		recordSize = 36
	default:
		return nil, fmt.Errorf("unsupported audio index version %d", version)
	}
	count := uint64(binary.LittleEndian.Uint32(idx[8:12]))
	want := uint64(12) + count*recordSize
	if want > uint64(len(idx)) {
		return nil, fmt.Errorf("audio index records are truncated: need %d bytes, have %d", want, len(idx))
	}

	bank := &nativeAudioBank{
		entries: make(map[string]*nativeAudioBankEntry, int(count)),
		data:    bag,
	}
	for i := uint64(0); i < count; i++ {
		off := uint64(12) + i*recordSize
		rec := idx[off : off+recordSize]
		nameBytes := rec[:16]
		if end := bytes.IndexByte(nameBytes, 0); end >= 0 {
			nameBytes = nameBytes[:end]
		}
		name := string(nameBytes)
		if name == "" {
			return nil, fmt.Errorf("audio index record %d has an empty name", i)
		}
		dataOff := uint64(binary.LittleEndian.Uint32(rec[16:20]))
		dataSize := uint64(binary.LittleEndian.Uint32(rec[20:24]))
		dataEnd := dataOff + dataSize
		if dataEnd < dataOff || dataEnd > uint64(len(bag)) {
			return nil, fmt.Errorf("audio sample %q points outside the bag", name)
		}
		entry := &nativeAudioBankEntry{
			name:  name,
			rate:  binary.LittleEndian.Uint32(rec[24:28]),
			flags: binary.LittleEndian.Uint32(rec[28:32]),
			data:  bag[dataOff:dataEnd],
		}
		if version >= 2 {
			entry.blockSize = binary.LittleEndian.Uint32(rec[32:36])
		}
		if entry.flags&8 != 0 {
			if entry.blockSize == 0 || dataSize%uint64(entry.blockSize) != 0 {
				return nil, fmt.Errorf("audio sample %q has invalid ADPCM block size %d", name, entry.blockSize)
			}
		}
		bank.entries[strings.ToLower(name)] = entry
	}
	return bank, nil
}

type nativeSoundDef struct {
	enabled   bool
	behavior  uint8
	priority  int16
	volume    uint32
	maxDist   int
	mode      int8
	field14   uint8
	field15   uint8
	field19   int8
	field20   int8
	delayMin  uint32
	delayMax  uint32
	sampleIDs []string
}

func defaultNativeSoundDef() nativeSoundDef {
	return nativeSoundDef{
		volume:  0x4000,
		maxDist: 600,
		mode:    1,
	}
}

type nativeAudioEffectsState struct {
	mu       sync.Mutex
	bank     *nativeAudioBank
	defs     [nativeSoundDefCount]nativeSoundDef
	voices   []ail.Sample
	next     int
	sequence uint32
}

func (s *nativeAudioEffectsState) init(driver ail.Driver, base string) error {
	s.close()
	bank, err := loadNativeAudioBank(base)
	if err != nil {
		return err
	}
	voices := make([]ail.Sample, 0, nativeAudioVoiceCount)
	for len(voices) < nativeAudioVoiceCount {
		h := driver.AllocateSample()
		if h == 0 {
			for _, allocated := range voices {
				allocated.Release()
			}
			return fmt.Errorf("allocate audio voice %d of %d", len(voices)+1, nativeAudioVoiceCount)
		}
		voices = append(voices, h)
	}

	s.mu.Lock()
	s.bank = bank
	s.voices = voices
	s.next = 0
	s.sequence = 0
	s.resetDefsLocked()
	s.mu.Unlock()
	audioEffectsLog.Printf("loaded %d samples and allocated %d voices", len(bank.entries), len(voices))
	return nil
}

func (s *nativeAudioEffectsState) close() {
	s.mu.Lock()
	voices := s.voices
	s.voices = nil
	s.bank = nil
	s.next = 0
	s.mu.Unlock()
	for _, h := range voices {
		h.Release()
	}
}

func (s *nativeAudioEffectsState) resetDefs() {
	s.mu.Lock()
	s.resetDefsLocked()
	s.mu.Unlock()
}

func (s *nativeAudioEffectsState) resetDefsLocked() {
	for i := range s.defs {
		s.defs[i] = defaultNativeSoundDef()
	}
}

func (s *nativeAudioEffectsState) readAUD(f *binfile.MemFile) error {
	s.mu.Lock()
	n, records, err := s.parseAUDLocked(f.Data())
	s.mu.Unlock()
	f.Skip(n)
	if err == nil && audioEffectsDebug {
		audioEffectsLog.Printf("loaded %d AUD definitions", records)
	}
	return err
}

func (s *nativeAudioEffectsState) readAVNT(f *binfile.MemFile) error {
	s.mu.Lock()
	n, err := s.parseAVNTLocked(f.Data())
	s.mu.Unlock()
	f.Skip(n)
	return err
}

func (s *nativeAudioEffectsState) parseAUDLocked(data []byte) (int, int, error) {
	r := nativeAudioDefDecoder{data: data}
	count, err := r.i32()
	if err != nil {
		return r.off, 0, err
	}
	if count <= 0 {
		return r.off, 0, nil
	}
	for i := 0; i < int(count); i++ {
		name, err := r.string8()
		if err != nil {
			return r.off, i, fmt.Errorf("AUD record %d name: %w", i, err)
		}
		def := defaultNativeSoundDef()
		def.behavior = 2
		if def.priority, err = r.i16(); err != nil {
			return r.off, i, fmt.Errorf("AUD record %q priority: %w", name, err)
		}
		volume, err := r.u8()
		if err != nil {
			return r.off, i, fmt.Errorf("AUD record %q volume: %w", name, err)
		}
		def.volume = 163 * uint32(volume)
		duration, err := r.i16()
		if err != nil {
			return r.off, i, fmt.Errorf("AUD record %q distance: %w", name, err)
		}
		if duration > 0 {
			def.maxDist = 15 * int(duration)
		}
		if def.field14, err = r.u8(); err != nil {
			return r.off, i, fmt.Errorf("AUD record %q field14: %w", name, err)
		}
		v19, err := r.i8()
		if err != nil {
			return r.off, i, fmt.Errorf("AUD record %q field19: %w", name, err)
		}
		def.field19 = v19
		v20, err := r.i8()
		if err != nil {
			return r.off, i, fmt.Errorf("AUD record %q field20: %w", name, err)
		}
		def.field20 = v20
		mode, err := r.i8()
		if err != nil {
			return r.off, i, fmt.Errorf("AUD record %q mode: %w", name, err)
		}
		def.mode = mode
		if mode >= 3 {
			return r.off, i, fmt.Errorf("AUD record %q has unsupported mode %d", name, mode)
		}
		for {
			sample, err := r.string8()
			if err != nil {
				return r.off, i, fmt.Errorf("AUD record %q sample: %w", name, err)
			}
			if sample == "" {
				break
			}
			if ext := strings.LastIndexByte(sample, '.'); ext >= 0 {
				sample = sample[:ext]
			}
			if len(def.sampleIDs) < 32 {
				def.sampleIDs = append(def.sampleIDs, sample)
			}
		}
		def.enabled = true
		if id, ok := nativeSoundID(name); ok {
			s.defs[id] = def
		}
	}
	return r.off, int(count), nil
}

func (s *nativeAudioEffectsState) parseAVNTLocked(data []byte) (int, error) {
	r := nativeAudioDefDecoder{data: data}
	name, err := r.string8()
	if err != nil {
		return r.off, fmt.Errorf("AVNT name: %w", err)
	}
	id, update := nativeSoundID(name)
	def := defaultNativeSoundDef()
	if update {
		def = s.defs[id]
	}
	for {
		typ, err := r.u8()
		if err != nil {
			return r.off, fmt.Errorf("AVNT %q property: %w", name, err)
		}
		switch typ {
		case 0:
			if update {
				def.enabled = true
				s.defs[id] = def
			}
			return r.off, nil
		case 1:
			v, err := r.i8()
			if err != nil {
				return r.off, err
			}
			def.mode = v
		case 2:
			v, err := r.u8()
			if err != nil {
				return r.off, err
			}
			def.behavior = v
		case 3:
			v, err := r.u8()
			if err != nil {
				return r.off, err
			}
			def.volume = 163 * uint32(v)
		case 4:
			v, err := r.u8()
			if err != nil {
				return r.off, err
			}
			def.field14 = v
		case 5:
			v, err := r.u8()
			if err != nil {
				return r.off, err
			}
			def.field15 = v
		case 6:
			v19, err := r.i8()
			if err != nil {
				return r.off, err
			}
			v20, err := r.i8()
			if err != nil {
				return r.off, err
			}
			def.field19, def.field20 = v19, v20
		case 7:
			def.sampleIDs = def.sampleIDs[:0]
			for {
				sample, err := r.string8()
				if err != nil {
					return r.off, err
				}
				if sample == "" {
					break
				}
				if len(def.sampleIDs) < 32 {
					def.sampleIDs = append(def.sampleIDs, sample)
				}
			}
		case 8:
			if def.delayMin, err = r.u32(); err != nil {
				return r.off, err
			}
			if def.delayMax, err = r.u32(); err != nil {
				return r.off, err
			}
		case 9:
			v, err := r.i16()
			if err != nil {
				return r.off, err
			}
			if v > 0 {
				def.maxDist = 15 * int(v)
			}
		case 10:
			if def.priority, err = r.i16(); err != nil {
				return r.off, err
			}
		default:
			return r.off, fmt.Errorf("AVNT %q has unknown property %d", name, typ)
		}
	}
}

func nativeSoundID(name string) (int, bool) {
	id := int(sound.ByName(name))
	return id, id > 0 && id < nativeSoundDefCount
}

type nativeAudioDefDecoder struct {
	data []byte
	off  int
}

func (r *nativeAudioDefDecoder) take(n int) ([]byte, error) {
	if n < 0 || n > len(r.data)-r.off {
		return nil, io.ErrUnexpectedEOF
	}
	b := r.data[r.off : r.off+n]
	r.off += n
	return b, nil
}

func (r *nativeAudioDefDecoder) u8() (uint8, error) {
	b, err := r.take(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *nativeAudioDefDecoder) i8() (int8, error) {
	v, err := r.u8()
	return int8(v), err
}

func (r *nativeAudioDefDecoder) i16() (int16, error) {
	b, err := r.take(2)
	if err != nil {
		return 0, err
	}
	return int16(binary.LittleEndian.Uint16(b)), nil
}

func (r *nativeAudioDefDecoder) u32() (uint32, error) {
	b, err := r.take(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *nativeAudioDefDecoder) i32() (int32, error) {
	v, err := r.u32()
	return int32(v), err
}

func (r *nativeAudioDefDecoder) string8() (string, error) {
	n, err := r.u8()
	if err != nil {
		return "", err
	}
	b, err := r.take(int(n))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *nativeAudioEffectsState) play(id sound.ID, requestedVolume int) bool {
	ind := int(id)
	if ind <= 0 || ind >= len(s.defs) {
		return false
	}
	if requestedVolume < 0 {
		requestedVolume = 0
	} else if requestedVolume > 100 {
		requestedVolume = 100
	}
	if requestedVolume == 0 || configGetVolume(VolumeFX) == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	def := &s.defs[ind]
	if !def.enabled || len(def.sampleIDs) == 0 || s.bank == nil || len(s.voices) == 0 {
		return false
	}

	played := false
	if def.behavior&4 != 0 {
		for _, sample := range def.sampleIDs {
			played = s.playSampleLocked(id, def, sample, requestedVolume) || played
		}
		return played
	}
	choice := 0
	if def.behavior&2 != 0 && len(def.sampleIDs) > 1 {
		if noxServer != nil && noxServer.Rand.Other != nil {
			choice = noxServer.Rand.Other.Int(0, len(def.sampleIDs)-1)
		} else {
			choice = int(s.sequence % uint32(len(def.sampleIDs)))
			s.sequence++
		}
	}
	return s.playSampleLocked(id, def, def.sampleIDs[choice], requestedVolume)
}

func (s *nativeAudioEffectsState) playSampleLocked(id sound.ID, def *nativeSoundDef, sample string, requestedVolume int) bool {
	entry := s.bank.entries[strings.ToLower(sample)]
	if entry == nil || len(entry.data) == 0 {
		if audioEffectsDebug {
			audioEffectsLog.Printf("sound %s references missing sample %q", id, sample)
		}
		return false
	}
	if entry.flags&8 == 0 || entry.blockSize == 0 {
		if audioEffectsDebug {
			audioEffectsLog.Printf("sound %s sample %q is not supported ADPCM", id, sample)
		}
		return false
	}

	voice := s.takeVoiceLocked()
	if voice == 0 {
		return false
	}
	voice.Init()
	format := int32(5)
	if entry.flags&1 != 0 {
		format = 7
	}
	voice.SetType(format, 0)
	voice.SetADPCMBlockSize(entry.blockSize)
	voice.SetPlaybackRate(int(entry.rate))
	voice.SetPan(63)

	eventVolume := uint32((uint64(163*requestedVolume) * uint64(def.volume)) >> 14)
	volume := int((uint64(127) * uint64(eventVolume)) >> 14)
	volume = volume * configGetVolume(VolumeFX) / VolumeMax
	if volume > 127 {
		volume = 127
	}
	voice.SetVolume(volume)
	ready := voice.BufferReady()
	if ready < 0 {
		return false
	}
	voice.LoadBuffer(uint32(ready), entry.data)
	if audioEffectsDebug {
		audioEffectsLog.Printf("playing %s via %s (%d bytes, %d Hz)", id, entry.name, len(entry.data), entry.rate)
	}
	return true
}

func (s *nativeAudioEffectsState) takeVoiceLocked() ail.Sample {
	for i := 0; i < len(s.voices); i++ {
		ind := (s.next + i) % len(s.voices)
		if s.voices[ind].Status() != 4 {
			s.next = (ind + 1) % len(s.voices)
			return s.voices[ind]
		}
	}
	voice := s.voices[s.next]
	s.next = (s.next + 1) % len(s.voices)
	return voice
}
