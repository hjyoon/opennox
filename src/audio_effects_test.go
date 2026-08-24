package opennox

import (
	"encoding/binary"
	"testing"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
)

func TestParseNativeAudioBank(t *testing.T) {
	idx := make([]byte, 12+36)
	copy(idx[:4], "GABA")
	binary.LittleEndian.PutUint32(idx[4:8], 2)
	binary.LittleEndian.PutUint32(idx[8:12], 1)
	rec := idx[12:]
	copy(rec[:16], "sselect")
	binary.LittleEndian.PutUint32(rec[16:20], 2)
	binary.LittleEndian.PutUint32(rec[20:24], 8)
	binary.LittleEndian.PutUint32(rec[24:28], 22050)
	binary.LittleEndian.PutUint32(rec[28:32], 12)
	binary.LittleEndian.PutUint32(rec[32:36], 4)
	bag := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	bank, err := parseNativeAudioBank(idx, bag)
	if err != nil {
		t.Fatal(err)
	}
	entry := bank.entries["sselect"]
	if entry == nil {
		t.Fatal("sample was not indexed")
	}
	if entry.rate != 22050 || entry.flags != 12 || entry.blockSize != 4 {
		t.Fatalf("metadata = rate %d, flags %#x, block %d", entry.rate, entry.flags, entry.blockSize)
	}
	if got := entry.data; len(got) != 8 || got[0] != 2 || got[7] != 9 {
		t.Fatalf("sample data = %v", got)
	}
}

func TestParseNativeAudioBankRejectsOutOfBoundsSample(t *testing.T) {
	idx := make([]byte, 12+36)
	copy(idx[:4], "GABA")
	binary.LittleEndian.PutUint32(idx[4:8], 2)
	binary.LittleEndian.PutUint32(idx[8:12], 1)
	rec := idx[12:]
	copy(rec[:16], "broken")
	binary.LittleEndian.PutUint32(rec[16:20], 8)
	binary.LittleEndian.PutUint32(rec[20:24], 8)
	binary.LittleEndian.PutUint32(rec[24:28], 22050)
	binary.LittleEndian.PutUint32(rec[28:32], 12)
	binary.LittleEndian.PutUint32(rec[32:36], 4)

	if _, err := parseNativeAudioBank(idx, make([]byte, 10)); err == nil {
		t.Fatal("out-of-bounds sample was accepted")
	}
}

func TestParseNativeAUDAndAVNT(t *testing.T) {
	var state nativeAudioEffectsState
	state.resetDefsLocked()

	aud := make([]byte, 4)
	binary.LittleEndian.PutUint32(aud, 1)
	aud = appendAudioString8(aud, "ShellClick")
	aud = appendAudioI16(aud, -7)
	aud = append(aud, 80)
	aud = appendAudioI16(aud, 10)
	aud = append(aud, 4, 0xfe, 3, 1)
	aud = appendAudioString8(aud, "first.wav")
	aud = appendAudioString8(aud, "second.WAV")
	aud = append(aud, 0)

	n, records, err := state.parseAUDLocked(aud)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(aud) || records != 1 {
		t.Fatalf("consumed %d/%d bytes and %d records", n, len(aud), records)
	}
	def := state.defs[sound.SoundShellClick]
	if !def.enabled || def.behavior != 2 || def.priority != -7 || def.volume != 163*80 || def.maxDist != 150 {
		t.Fatalf("AUD definition = %+v", def)
	}
	if def.field14 != 4 || def.field19 != -2 || def.field20 != 3 || def.mode != 1 {
		t.Fatalf("AUD fields = %+v", def)
	}
	if len(def.sampleIDs) != 2 || def.sampleIDs[0] != "first" || def.sampleIDs[1] != "second" {
		t.Fatalf("AUD samples = %q", def.sampleIDs)
	}

	avnt := appendAudioString8(nil, "ShellClick")
	avnt = append(avnt, 2, 4)
	avnt = append(avnt, 3, 90)
	avnt = append(avnt, 5, 6)
	avnt = append(avnt, 6, 0xfd, 8)
	avnt = append(avnt, 7)
	avnt = appendAudioString8(avnt, "override")
	avnt = append(avnt, 0)
	avnt = append(avnt, 8)
	avnt = appendAudioU32(avnt, 33)
	avnt = appendAudioU32(avnt, 44)
	avnt = append(avnt, 9)
	avnt = appendAudioI16(avnt, 25)
	avnt = append(avnt, 10)
	avnt = appendAudioI16(avnt, -3)
	avnt = append(avnt, 0)

	n, err = state.parseAVNTLocked(avnt)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(avnt) {
		t.Fatalf("consumed %d/%d AVNT bytes", n, len(avnt))
	}
	def = state.defs[sound.SoundShellClick]
	if def.behavior != 4 || def.volume != 163*90 || def.field15 != 6 || def.field19 != -3 || def.field20 != 8 {
		t.Fatalf("AVNT definition = %+v", def)
	}
	if def.delayMin != 33 || def.delayMax != 44 || def.maxDist != 375 || def.priority != -3 {
		t.Fatalf("AVNT timing fields = %+v", def)
	}
	if len(def.sampleIDs) != 1 || def.sampleIDs[0] != "override" {
		t.Fatalf("AVNT samples = %q", def.sampleIDs)
	}
}

func TestParseNativeAVNTRejectsUnknownProperty(t *testing.T) {
	var state nativeAudioEffectsState
	state.resetDefsLocked()
	data := appendAudioString8(nil, "ShellClick")
	data = append(data, 0xff)
	if _, err := state.parseAVNTLocked(data); err == nil {
		t.Fatal("unknown AVNT property was accepted")
	}
}

func TestLegacyClientPlaySoundSpecialUsesNativeCallback(t *testing.T) {
	prev := legacy.ClientPlaySoundSpecial
	defer func() { legacy.ClientPlaySoundSpecial = prev }()
	var gotID sound.ID
	var gotVolume int
	legacy.ClientPlaySoundSpecial = func(id sound.ID, volume int) {
		gotID, gotVolume = id, volume
	}
	legacy.Nox_xxx_clientPlaySoundSpecial_452D80(sound.SoundShellClick, 73)
	if gotID != sound.SoundShellClick || gotVolume != 73 {
		t.Fatalf("callback = (%v, %d)", gotID, gotVolume)
	}
}

func appendAudioString8(dst []byte, s string) []byte {
	if len(s) > 255 {
		panic("test string too long")
	}
	dst = append(dst, byte(len(s)))
	return append(dst, s...)
}

func appendAudioI16(dst []byte, v int16) []byte {
	return append(dst, byte(v), byte(uint16(v)>>8))
}

func appendAudioU32(dst []byte, v uint32) []byte {
	return append(dst, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}
