package legacy

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"sync"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/ifs"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

const mapIntroVersion505080 = 1

type mapIntroReadSeekCloser505080 interface {
	io.Reader
	io.Seeker
	io.Closer
}

type mapIntroHooks505080 struct {
	external  bool
	parseOnly bool
	open      func(string) (mapIntroReadSeekCloser505080, error)
	create    func(string) (io.WriteCloser, error)
	clear     func()
	store     func([]byte)
}

var mapIntroState505080 struct {
	sync.RWMutex
	data []byte
}

func clearMapIntro505080() {
	mapIntroState505080.Lock()
	mapIntroState505080.data = nil
	mapIntroState505080.Unlock()
}

func storeMapIntro505080(data []byte) {
	mapIntroState505080.Lock()
	mapIntroState505080.data = data
	mapIntroState505080.Unlock()
}

func mapIntroSnapshot505080() []byte {
	mapIntroState505080.RLock()
	defer mapIntroState505080.RUnlock()
	return append([]byte(nil), mapIntroState505080.data...)
}

func mapIntroPath505080(root, mapName string) string {
	return filepath.Join(root, "maps", mapName, mapName+".txt")
}

func mapRWMapIntro505080(cf *cryptfile.CryptFile, path string, hooks mapIntroHooks505080) error {
	if cf == nil {
		return fmt.Errorf("map intro: nil crypt file")
	}
	if hooks.clear != nil {
		hooks.clear()
	}
	if cf.ReadOnly() {
		return mapReadIntro505080(cf, path, hooks)
	}
	return mapWriteIntro505080(cf, path, hooks)
}

func mapReadIntro505080(cf *cryptfile.CryptFile, path string, hooks mapIntroHooks505080) error {
	version, err := cf.ReadU16()
	if err != nil {
		return fmt.Errorf("map intro version: %w", err)
	}
	if int16(version) < mapIntroVersion505080 {
		return fmt.Errorf("unsupported map intro version: %d", int16(version))
	}
	size, err := cf.ReadU32()
	if err != nil {
		return fmt.Errorf("map intro length: %w", err)
	}
	if int32(size) <= 0 {
		return nil
	}
	if hooks.parseOnly {
		if err := cf.Seek(int64(size), io.SeekCurrent); err != nil {
			return fmt.Errorf("skip map intro payload: %w", err)
		}
		return nil
	}
	if hooks.external {
		if hooks.create == nil {
			return fmt.Errorf("map intro destination is not configured")
		}
		out, err := hooks.create(path)
		if err != nil {
			return fmt.Errorf("create map intro %q: %w", path, err)
		}
		if out == nil {
			return fmt.Errorf("create map intro %q: nil writer", path)
		}
		if _, err := io.CopyN(out, cf, int64(size)); err != nil {
			_ = out.Close()
			return fmt.Errorf("read map intro payload: %w", err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("close map intro %q: %w", path, err)
		}
		return nil
	}
	if hooks.store == nil {
		return fmt.Errorf("map intro storage is not configured")
	}
	var buf bytes.Buffer
	if size <= 1<<20 {
		buf.Grow(int(size))
	}
	if _, err := io.CopyN(&buf, cf, int64(size)); err != nil {
		return fmt.Errorf("read map intro payload: %w", err)
	}
	hooks.store(buf.Bytes())
	return nil
}

func mapWriteIntro505080(cf *cryptfile.CryptFile, path string, hooks mapIntroHooks505080) error {
	if err := cf.WriteU16(mapIntroVersion505080); err != nil {
		return fmt.Errorf("write map intro version: %w", err)
	}
	if !hooks.external || hooks.open == nil {
		return cf.WriteU32(0)
	}
	in, err := hooks.open(path)
	if err != nil || in == nil {
		return cf.WriteU32(0)
	}
	size, err := in.Seek(0, io.SeekEnd)
	if err != nil {
		_ = in.Close()
		return fmt.Errorf("size map intro %q: %w", path, err)
	}
	if size < 0 || size > math.MaxInt32 {
		_ = in.Close()
		return fmt.Errorf("map intro %q has unsupported length %d", path, size)
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		_ = in.Close()
		return fmt.Errorf("rewind map intro %q: %w", path, err)
	}
	if err := cf.WriteU32(uint32(size)); err != nil {
		_ = in.Close()
		return fmt.Errorf("write map intro length: %w", err)
	}
	if size != 0 {
		if _, err := io.CopyN(cf, in, size); err != nil {
			_ = in.Close()
			return fmt.Errorf("write map intro payload: %w", err)
		}
	}
	if err := in.Close(); err != nil {
		return fmt.Errorf("close map intro %q: %w", path, err)
	}
	return nil
}

func Nox_server_mapRWMapIntro_505080(cf *cryptfile.CryptFile, _ unsafe.Pointer) error {
	path := mapIntroPath505080(datapath.Data(), Nox_xxx_mapGetMapName_409B40())
	return mapRWMapIntro505080(cf, path, mapIntroHooks505080{
		external:  noxflags.HasGame(noxflags.GameFlag22),
		parseOnly: noxflags.HasGame(noxflags.GameFlag23),
		open: func(path string) (mapIntroReadSeekCloser505080, error) {
			return ifs.Open(path)
		},
		create: func(path string) (io.WriteCloser, error) {
			return ifs.Create(path)
		},
		clear: clearMapIntro505080,
		store: storeMapIntro505080,
	})
}
