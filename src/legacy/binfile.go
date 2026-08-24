package legacy

/*
#include <stdbool.h>
#include <stdio.h>
*/
import "C"
import (
	"os"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/binfile"
)

//export nox_binfile_open_408CC0
func nox_binfile_open_408CC0(cpath *C.char, cmode_cgo int32) *C.FILE {
	cmode := int(cmode_cgo)
	file, err := binfile.BinfileOpen(GoString(cpath), binfile.Mode(cmode))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		binfile.Log.Println(err)
		return nil
	}
	return NewFileHandle(file.File)
}

//export nox_binfile_close_408D90
func nox_binfile_close_408D90(cfile *C.FILE) int32 {
	file := fileByHandle(cfile)
	file.Close()
	return int32(1)
}

//export nox_binfile_ftell_426A50
func nox_binfile_ftell_426A50(cfile *C.FILE) int32 {
	file := fileByHandle(cfile)
	off := file.Bin.Written()
	return int32(int(off))
}

//export nox_binfile_lastErr_409370
func nox_binfile_lastErr_409370(cfile *C.FILE) int32 {
	file := fileByHandle(cfile)
	if file.Err != nil {
		return int32(-1)
	}
	return int32(0)
}

//export nox_binfile_cryptSet_408D40
func nox_binfile_cryptSet_408D40(cfile *C.FILE, ckey_cgo int32) int32 {
	ckey := int(ckey_cgo)
	file := fileByHandle(cfile)
	bin := file.Bin
	err := bin.SetKey(ckey)
	if err != nil {
		binfile.Log.Println(err)
		return int32(0)
	}
	return int32(1)
}

//export nox_binfile_fread_408E40
func nox_binfile_fread_408E40(cbuf *C.char, sz_cgo, cnt_cgo int32, cfile *C.FILE) int32 {
	sz := int(sz_cgo)
	cnt := int(cnt_cgo)
	if sz*cnt == 0 {
		return int32(0)
	}
	file := fileByHandle(cfile)
	bin := file.Bin
	buf := unsafe.Slice((*byte)(unsafe.Pointer(cbuf)), sz*cnt)
	n, err := bin.Read(buf)
	if err != nil {
		file.Err = err
	}
	return int32(n / sz)
}

//export nox_binfile_fread_raw_40ADD0
func nox_binfile_fread_raw_40ADD0(cbuf *C.char, sz, cnt C.size_t, cfile *C.FILE) int32 {
	if sz*cnt == 0 {
		return int32(0)
	}
	n := int(nox_fs_fread(cfile, unsafe.Pointer(cbuf), int32(sz*cnt)))
	if n >= 0 {
		n /= int(sz)
	}
	return int32(n)
}

//export nox_binfile_fread_align_408FE0
func nox_binfile_fread_align_408FE0(cbuf *C.char, sz_cgo, cnt_cgo int32, cfile *C.FILE) int32 {
	sz := int(sz_cgo)
	cnt := int(cnt_cgo)
	if sz*cnt == 0 {
		return int32(0)
	}
	file := fileByHandle(cfile)
	bin := file.Bin
	buf := unsafe.Slice((*byte)(unsafe.Pointer(cbuf)), sz*cnt)
	n, err := bin.ReadAligned(buf)
	if err != nil {
		n = -1
	}
	if err != nil {
		file.Err = err
	}
	return int32(n / sz)
}

//export nox_binfile_skipLine_409520
func nox_binfile_skipLine_409520(cfile *C.FILE) int32 {
	file := fileByHandle(cfile)
	if err := file.Bin.SkipLine(); err != nil {
		binfile.Log.Println(err)
		return int32(-1)
	}
	return int32(0)
}

//export nox_binfile_fseek_409050
func nox_binfile_fseek_409050(cfile *C.FILE, coff_cgo, cwhence_cgo int32) int32 {
	coff := int(coff_cgo)
	cwhence := int(cwhence_cgo)
	file := fileByHandle(cfile)
	err := file.Bin.FileSeek(int64(coff), convWhence(cwhence))
	if err != nil {
		binfile.Log.Println(err)
	}
	return int32(0)
}
