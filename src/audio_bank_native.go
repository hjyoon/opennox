package opennox

import (
	"encoding/binary"
	"io"
	"sort"
	"strings"
	"unsafe"

	"github.com/opennox/libs/ifs"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func audioBankRecordNative(p unsafe.Pointer, index int32) unsafe.Pointer {
	b := (*audioStructXxx)(p)
	if b == nil || index < 0 || uint(index) >= b.size4 {
		return nil
	}
	return unsafe.Add(unsafe.Pointer(b.arr0), uintptr(index)*unsafe.Sizeof(audioStructYyy{}))
}

func audioBankLookupNative(p unsafe.Pointer, name string) uint32 {
	b := (*audioStructXxx)(p)
	rows := unsafe.Slice(b.arr0, b.size4)
	key := strings.ToLower(name)
	i := sort.Search(len(rows), func(i int) bool {
		return strings.ToLower(alloc.GoStringS(rows[i].field0[:])) >= key
	})
	if i < len(rows) && strings.EqualFold(alloc.GoStringS(rows[i].field0[:]), name) {
		return uint32(i)
	}
	return 0xffffffff
}

func audioBankCloseReadNative(p unsafe.Pointer) unsafe.Pointer {
	b := (*audioStructXxx)(p)
	b.field280 = nil
	if b.field272 != nil {
		legacy.Nox_fs_close(b.field272)
		b.field272 = nil
	}
	return nil
}

func audioBankRemainingNative(p unsafe.Pointer) uint32 { return (*audioStructXxx)(p).field284 }

func audioBankReadNative(p unsafe.Pointer, data *byte, size int32) int32 {
	b := (*audioStructXxx)(p)
	if b.field280 == nil {
		return 0
	}
	if size > int32(b.field284) {
		size = int32(b.field284)
	}
	if size <= 0 {
		return 0
	}
	n, _ := legacy.AudioBankFileNative(b.field280).Read(unsafe.Slice(data, int(size)))
	b.field284 -= uint32(n)
	return int32(n)
}

func audioBankSelectNative(p unsafe.Pointer, index int32) int32 {
	b := (*audioStructXxx)(p)
	record := (*[9]uint32)(audioBankRecordNative(p, index))
	if record == nil {
		return 0
	}
	audioBankCloseReadNative(p)
	b.field280, b.field284 = b.bagfile268, record[5]
	result := int32(1)
	if _, err := legacy.AudioBankFileNative(b.bagfile268).Seek(int64(record[4]), io.SeekStart); err != nil || record[5] == 0 {
		result = 0
	}
	if b.field276 == 0 {
		return result
	}
	name := alloc.GoString((*byte)(unsafe.Pointer(record)))
	f, err := ifs.Open(alloc.GoStringS(b.path2_8[:]) + name + ".wav")
	if err != nil {
		return result
	}
	wav := binfile.NewFile(f)
	b.field272 = legacy.NewFileHandle(wav)
	var header [12]byte
	if _, err := io.ReadFull(wav, header[:]); err != nil || string(header[:4]) != "RIFF" || string(header[8:]) != "WAVE" {
		legacy.Nox_fs_close(b.field272)
		b.field272 = nil
		return result
	}
	var chunk [8]byte
	var format [16]byte
	foundFormat := false
	var length uint32
	for {
		if _, err := io.ReadFull(wav, chunk[:]); err != nil {
			break
		}
		size := binary.LittleEndian.Uint32(chunk[4:])
		switch string(chunk[:4]) {
		case "fmt ":
			if size < uint32(len(format)) {
				break
			}
			if _, err := io.ReadFull(wav, format[:]); err != nil {
				break
			}
			foundFormat = true
			if _, err := wav.Seek(int64(size)-16, io.SeekCurrent); err != nil {
				break
			}
			continue
		case "data":
			length = size
		default:
			if _, err := wav.Seek(int64(size), io.SeekCurrent); err == nil {
				continue
			}
		}
		break
	}
	channels := binary.LittleEndian.Uint16(format[2:])
	if !foundFormat || channels == 0 {
		// The old C path read an uninitialized fmt record here. Preserve the
		// selected bag stream for malformed override files instead.
		legacy.Nox_fs_close(b.field272)
		b.field272 = nil
		return result
	}
	record[7] = 2
	if binary.LittleEndian.Uint16(format[12:])/channels == 2 {
		record[7] = 6
	}
	if channels == 2 {
		record[7] |= 1
	}
	record[6] = binary.LittleEndian.Uint32(format[4:])
	b.field284, b.field280 = length, b.field272
	return 1
}
