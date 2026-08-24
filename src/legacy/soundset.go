package legacy

import (
	"bufio"
	"fmt"
	"io"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	soundSetRecordSize = uintptr(84)
	soundSetFieldTable = uintptr(64704)
)

type soundSetRecord struct {
	name string
	data unsafe.Pointer
	next *soundSetRecord
}

var (
	soundSetHead   *soundSetRecord
	soundSetByName = make(map[string]*soundSetRecord)
)

func parseSoundSetBin(path string) error {
	f, err := binfile.BinfileOpen(path, binfile.ReadOnly)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.SetKey(5); err != nil {
		return err
	}
	r := bufio.NewReader(f)
	for {
		name, ok := readSoundSetToken(r)
		if !ok {
			return nil
		}
		data, _ := alloc.Malloc(soundSetRecordSize)
		rec := &soundSetRecord{name: name, data: data, next: soundSetHead}
		soundSetHead = rec
		soundSetByName[name] = rec

		for {
			field, ok := readSoundSetToken(r)
			if !ok || field == "END" {
				break
			}
			soundName, ok := readSoundSetToken(r)
			if !ok {
				break
			}
			off, ok := soundSetFieldOffset(field)
			if !ok {
				return fmt.Errorf("unknown sound-set field %q", field)
			}
			*(*uint32)(unsafe.Add(data, off)) = uint32(sound.ByName(soundName))
		}
	}
}

func readSoundSetToken(r *bufio.Reader) (string, bool) {
	var out []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", false
		}
		if b == '/' {
			if next, err := r.Peek(1); err == nil && next[0] == '/' {
				_, _ = r.ReadByte()
				for {
					b, err = r.ReadByte()
					if err != nil || b == '\n' {
						break
					}
				}
				out = out[:0]
				if err == io.EOF {
					return "", false
				}
				continue
			}
		}
		if isSoundSetSpace(b) {
			if len(out) == 0 {
				continue
			}
			return string(out), true
		}
		out = append(out, b)
	}
}

func isSoundSetSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

func soundSetFieldOffset(name string) (uintptr, bool) {
	for off := soundSetFieldTable; ; off += 8 {
		p := *memmap.PtrPtr(0x587000, off)
		if p == nil {
			return 0, false
		}
		if alloc.GoString((*byte)(p)) == name {
			fieldOff := uintptr(*memmap.PtrUint32(0x587000, off+4))
			if fieldOff < 4 || fieldOff+4 > soundSetRecordSize {
				return 0, false
			}
			return fieldOff, true
		}
	}
}

func getDefaultSoundSet(name string) unsafe.Pointer {
	if rec := soundSetByName[name]; rec != nil {
		return rec.data
	}
	return nil
}
