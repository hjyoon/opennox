package legacy

/*
#include "GAME4_1.h"

typedef const char* nox_monster_bin_name_t;
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/binfile"
)

const monsterBinTokenCapacity = 256

// monsterBinFileTokens consumes the same file position across the C token,
// line and definition entries.
type monsterBinFileTokens struct {
	file *binfile.File
}

func (r monsterBinFileTokens) readByte() (byte, error) {
	var buf [1]byte
	n, err := r.file.Bin.Read(buf[:])
	if err != nil {
		r.file.Err = err
	}
	if r.file.Err != nil {
		return 0, r.file.Err
	}
	if n == 0 {
		return 0, io.ErrNoProgress
	}
	return buf[0], nil
}

func (r monsterBinFileTokens) readToken(dst []byte) (int, error) {
	var previous byte
	n := 0
	for {
		b, err := r.readByte()
		if err != nil {
			return n, err
		}
		switch {
		case isLegacyTokenSpace(b):
			if n != 0 {
				dst[n] = 0
				return n, nil
			}
		case b == '/' && previous == '/':
			_ = r.skipLine()
			n = 0
		default:
			if n >= len(dst)-1 {
				return n, fmt.Errorf("monster.bin token exceeds %d bytes", len(dst)-1)
			}
			dst[n] = b
			n++
		}
		previous = b
	}
}

func (r monsterBinFileTokens) next() (string, error) {
	var token [monsterBinTokenCapacity]byte
	n, err := r.readToken(token[:])
	if err != nil {
		return "", err
	}
	return GoStringS(token[:n]), nil
}

func (r monsterBinFileTokens) skipLine() error {
	return skipMonsterBinLine(r.readByte)
}

func (r monsterBinFileTokens) readMissileName(dst []byte) error {
	// The original writes directly into the record and clears the first
	// byte for the literal NULL. Bytes after the terminator stay in place.
	if _, err := r.readToken(dst); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if GoStringS(dst) == "NULL" {
		dst[0] = 0
	}
	return nil
}

//export nox_xxx_readStr_517090
func nox_xxx_readStr_517090(input *C.FILE, dst *C.uint8_t) C.int {
	if input == nil || dst == nil {
		return 0
	}
	file := fileByHandle(input)
	if file == nil || file.Bin == nil {
		return 0
	}
	tokens := monsterBinFileTokens{file: file}
	if _, err := tokens.readToken(unsafe.Slice((*byte)(unsafe.Pointer(dst)), monsterBinTokenCapacity)); err != nil {
		return 0
	}
	return 1
}

//export sub_517140
func sub_517140(input *C.FILE) C.int {
	if input == nil {
		return -1
	}
	file := fileByHandle(input)
	if file == nil || file.Bin == nil {
		return -1
	}
	if err := (monsterBinFileTokens{file: file}).skipLine(); err != nil {
		return -1
	}
	return 0
}

//export nox_xxx_servParseMonsterDef_517170
func nox_xxx_servParseMonsterDef_517170(input *C.FILE, name C.nox_monster_bin_name_t) C.int {
	if input == nil || name == nil {
		return 0
	}
	file := fileByHandle(input)
	if file == nil || file.Bin == nil {
		return 0
	}
	tokens := monsterBinFileTokens{file: file}
	if err := loadedMonsterDefs.readDefinition(tokens, GoString((*C.char)(unsafe.Pointer(name))), currentMonsterBinMode()); err != nil {
		return 0
	}
	return 1
}
