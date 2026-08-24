package client

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func newAnimationMemFile(t *testing.T, data []byte) *binfile.MemFile {
	t.Helper()
	buf, _ := alloc.CloneSlice(data)
	f := binfile.NewMemFile(unsafe.Pointer(&buf[0]), len(buf))
	t.Cleanup(f.Free)
	return f
}

func appendAnimationU32(dst []byte, v uint32) []byte {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	return append(dst, buf[:]...)
}

func appendZeroFrameState(dst []byte, params uint32, delay byte, kind string) []byte {
	dst = appendAnimationU32(dst, 0x53544154) // "STAT"
	dst = appendAnimationU32(dst, params)
	dst = append(dst, 4, 'N', 'U', 'L', 'L', 0)
	dst = append(dst, 0, delay, byte(len(kind)))
	return append(dst, kind...)
}

func TestParseAnimStateDrawDataUsesNativeLayout(t *testing.T) {
	var data []byte
	data = appendZeroFrameState(data, 2, 7, "Loop")
	data = appendZeroFrameState(data, 4, 8, "OneShot")
	data = appendZeroFrameState(data, 8, 9, "Slave")
	data = appendAnimationU32(data, 0x454E4420) // "END "

	f := newAnimationMemFile(t, data)
	dd, free := alloc.New(AnimationStateDrawData{})
	t.Cleanup(free)
	if err := ParseAnimStateDrawData(dd, f, nil); err != nil {
		t.Fatalf("parse animate-state draw: %v", err)
	}
	if got, want := dd.Size, uint32(unsafe.Sizeof(*dd)); got != want {
		t.Fatalf("draw-data size = %d, want native size %d", got, want)
	}
	want := [...]struct {
		delay uint16
		kind  AnimKind
	}{{7, AnimLoop}, {8, AnimOneShot}, {9, AnimSlave}}
	for i, v := range want {
		if dd.Anim[i].Cnt40 != 0 || dd.Anim[i].Val42 != v.delay || dd.Anim[i].Kind != v.kind {
			t.Errorf("animation %d = %+v", i, dd.Anim[i])
		}
	}
	if len(f.Data()) != 0 {
		t.Fatalf("parser left %d unread bytes", len(f.Data()))
	}
}

func TestParseAnimStateDrawDataRejectsMissingState(t *testing.T) {
	var data []byte
	data = appendAnimationU32(data, 0x53544154)
	data = appendAnimationU32(data, 0)

	f := newAnimationMemFile(t, data)
	dd, free := alloc.New(AnimationStateDrawData{})
	t.Cleanup(free)
	if err := ParseAnimStateDrawData(dd, f, nil); err == nil {
		t.Fatal("parse animate-state draw succeeded without a state bit")
	}
}

func TestParseAnimStateDrawDataLoadsSingleFrameSequence(t *testing.T) {
	var data []byte
	data = appendAnimationU32(data, 0x53544154)
	data = appendAnimationU32(data, 0xE) // bit 2 wins over bits 4 and 8
	data = append(data, 4, 'N', 'U', 'L', 'L', 0)
	data = append(data, 2, 3, 4, 'L', 'o', 'o', 'p')
	data = appendAnimationU32(data, 42)
	data = appendAnimationU32(data, ^uint32(0))
	data = append(data, 7, 3, 'f', 'o', 'o')
	data = appendAnimationU32(data, 0x454E4420)

	f := newAnimationMemFile(t, data)
	dd, free := alloc.New(AnimationStateDrawData{})
	t.Cleanup(free)
	type imageRequest struct {
		id   int
		typ  byte
		name string
	}
	var requests []imageRequest
	err := ParseAnimStateDrawData(dd, f, func(id int, typ byte, name string) noxrender.ImageHandle {
		requests = append(requests, imageRequest{id: id, typ: typ, name: name})
		return nil
	})
	if err != nil {
		t.Fatalf("parse animate-state frames: %v", err)
	}
	frames := dd.Anim[0].FramesSlice(0)
	t.Cleanup(func() { alloc.FreeSlice(frames) })
	if len(frames) != 2 {
		t.Fatalf("frame count = %d, want 2", len(frames))
	}
	if dd.Anim[1].Cnt40 != 0 || dd.Anim[2].Cnt40 != 0 {
		t.Fatal("state-bit priority did not select animation zero")
	}
	if dd.Anim[0].Frames[1] != nil || dd.Anim[0].Frames[2] != nil {
		t.Fatal("single-sequence loader populated a directional frame slot")
	}
	want := []imageRequest{{id: 42}, {id: -1, typ: 7, name: "foo"}}
	if len(requests) != len(want) {
		t.Fatalf("image requests = %+v, want %+v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("image request %d = %+v, want %+v", i, requests[i], want[i])
		}
	}
}
