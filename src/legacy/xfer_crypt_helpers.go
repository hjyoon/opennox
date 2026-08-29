package legacy

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

// These byte-stream helpers are shared by native-width transfer routines and
// inventory transfer code that must also compile in the original 32-bit ABI.
func monsterRWBytes528DB0(cf *cryptfile.CryptFile, p []byte) error {
	if len(p) == 0 {
		return nil
	}
	if cf.ReadOnly() {
		_, err := io.ReadFull(cf, p)
		return err
	}
	n, err := cf.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	return err
}

func monsterRWU8(cf *cryptfile.CryptFile, value byte) (byte, error) {
	buf := [1]byte{value}
	err := monsterRWBytes528DB0(cf, buf[:])
	return buf[0], err
}

func monsterRWU16(cf *cryptfile.CryptFile, value uint16) (uint16, error) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], value)
	err := monsterRWBytes528DB0(cf, buf[:])
	return binary.LittleEndian.Uint16(buf[:]), err
}

func monsterRWU32(cf *cryptfile.CryptFile, value uint32) (uint32, error) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	err := monsterRWBytes528DB0(cf, buf[:])
	return binary.LittleEndian.Uint32(buf[:]), err
}

func monsterRWI32(cf *cryptfile.CryptFile, value int32) (int32, error) {
	v, err := monsterRWU32(cf, uint32(value))
	return int32(v), err
}

func monsterRWF32(cf *cryptfile.CryptFile, value float32) (float32, error) {
	v, err := monsterRWU32(cf, math.Float32bits(value))
	return math.Float32frombits(v), err
}

func monsterRWString8(cf *cryptfile.CryptFile, value string) (string, error) {
	if len(value) > math.MaxUint8 {
		value = value[:math.MaxUint8]
	}
	sz, err := monsterRWU8(cf, byte(len(value)))
	if err != nil {
		return "", err
	}
	if cf.ReadOnly() {
		buf := make([]byte, int(sz))
		if err := monsterRWBytes528DB0(cf, buf); err != nil {
			return "", err
		}
		return string(buf), nil
	}
	if err := monsterRWBytes528DB0(cf, []byte(value)); err != nil {
		return "", err
	}
	return value, nil
}

func cStringBytes528DB0(buf []byte) string {
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

func setCStringBytes528DB0(buf []byte, value string) {
	clear(buf)
	if len(buf) == 0 {
		return
	}
	copy(buf[:len(buf)-1], value)
}
