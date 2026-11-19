package legacy

import (
	"bytes"
	"errors"
	"io"
	"unsafe"

	"github.com/opennox/libs/ifs"

	"github.com/opennox/opennox/v1/internal/binfile"
)

func loadAccessList4E41B0(path string) bool {
	f, err := ifs.Open(path)
	if err != nil {
		return false
	}
	file := binfile.NewTextFile(f)
	defer file.Close()

	for {
		token, atStart, read := readAccessListToken4E41B0(file)
		if !read {
			return true
		}
		if !atStart {
			continue
		}
		switch string(token) {
		case "[Banned Users]":
			if !loadAccessBanned4E42C0(file) {
				return false
			}
		case "[Allowed Users]":
			loadAccessAllowed4E4390(file)
		}
	}
}

func loadAccessBanned4E42C0(file *binfile.File) bool {
	for {
		name, _, read := readAccessListToken4E41B0(file)
		if !read || len(name) == 0 {
			return true
		}
		serial, _, read := readAccessListToken4E41B0(file)
		if !read {
			return true
		}
		if len(serial) == 0 {
			return false
		}
		wide := accessListWideName4E41B0(name)
		var serialBuf [24]byte
		var serialPtr unsafe.Pointer
		if string(serial) != "0" {
			copy(serialBuf[:len(serialBuf)-1], serial)
			serialPtr = unsafe.Pointer(&serialBuf[0])
		}
		sub416770Native(0, unsafe.Pointer(&wide[0]), serialPtr)
	}
}

func loadAccessAllowed4E4390(file *binfile.File) {
	for {
		name, _, read := readAccessListToken4E41B0(file)
		if !read || len(name) == 0 {
			return
		}
		wide := accessListWideName4E41B0(name)
		sub4168A0Native(unsafe.Pointer(&wide[0]))
	}
}

func readAccessListToken4E41B0(file *binfile.File) (token []byte, atStart, read bool) {
	const separators = "\r\t\n"
	line, err := file.ReadStringLimit(1023)
	if err != nil && (len(line) == 0 || !errors.Is(err, io.EOF)) {
		return nil, false, false
	}
	if i := bytes.IndexByte(line, 0); i >= 0 {
		line = line[:i]
	}
	token = bytes.TrimLeft(line, separators)
	atStart = len(token) == len(line)
	if i := bytes.IndexAny(token, separators); i >= 0 {
		token = token[:i]
	}
	return token, atStart, true
}

func accessListWideName4E41B0(token []byte) (name [26]uint16) {
	for i, value := range token[:min(len(token), len(name)-1)] {
		name[i] = uint16(value)
	}
	return
}
