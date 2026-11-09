package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// These helpers author input bytes only. Expected return/cursor/whole globals
// and scratch bytes come from the private exact preserved-C caller, not from a
// second Go implementation or a generated golden array.
type thingWire411540 []byte

func (w *thingWire411540) b(v ...byte) { *w = append(*w, v...) }
func (w *thingWire411540) word(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.b(b[:]...)
}
func (w *thingWire411540) short(v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	w.b(b[:]...)
}
func (w *thingWire411540) str(b []byte) { w.b(byte(len(b))); w.b(b...) }
func (w *thingWire411540) aligned(v byte) {
	for len(*w)%8 != 0 {
		w.b(0x95)
	}
	w.b(v, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87)
}
func (w *thingWire411540) ref(named bool) {
	if named {
		w.word(0xffffffff)
		w.b(7)
		w.str([]byte("img.bin"))
	} else {
		w.word(0x81002000)
	}
}

type thingInput411540 struct {
	name        string
	entry       uint32
	data        []byte
	start       int
	floor, edge uint32
}

func thingInputs411540() []thingInput411540 {
	var out []thingInput411540
	add := func(name string, entry uint32, w thingWire411540, start int) {
		out = append(out, thingInput411540{name, entry, append([]byte(nil), w...), start, 0, 0})
	}
	for _, start := range []int{0, 1} {
		prefix := func() thingWire411540 { return bytes.Repeat([]byte{0x93}, start) }
		for _, title := range [][]byte{nil, []byte("STONE"), bytes.Repeat([]byte{'X'}, 31), []byte{'A', 0, 'B'}} {
			w := prefix()
			w.word(0x11223344)
			w.str(title)
			w.b(255, 0, 255)
			w.word(0x80000000)
			w.word(0xffffffff)
			w.b(0xD1, 1, 2, 1, 0xD2)
			w.ref(false)
			w.ref(true)
			add(fmt.Sprintf("floor-name-%d-start-%d", len(title), start), 0x411540, w, start)
			q := append(thingWire411540(nil), w...)
			q.word(0x454e4420)
			add(fmt.Sprintf("floor-skip-name-%d-start-%d", len(title), start), 0x414DB0, q, start)
		}
		w := prefix()
		for _, badEND := range []bool{false, true} {
			w = prefix()
			w.word(0)
			w.str([]byte("EDGE"))
			w.word(0x87654321)
			w.word(0x80000000)
			w.b(0x42, 1, 0x77, 0, 1, 0)
			w.ref(false)
			w.ref(true)
			if badEND {
				w.word(0xdeadbeef)
			} else {
				w.word(0x454e4420)
			}
			add(fmt.Sprintf("edge-end-%t-start-%d", badEND, start), 0x411850, w, start)
			add(fmt.Sprintf("edge-skip-end-%t-start-%d", badEND, start), 0x414E70, w, start)
		}
		w = prefix()
		w.word(0)
		w.str([]byte("REJECT"))
		w.word(3)
		w.word(4)
		w.b(5, 6, 7, 1)
		add(fmt.Sprintf("edge-flag-reject-%d", start), 0x411850, w, start)
		add(fmt.Sprintf("edge-skip-flag-reject-%d", start), 0x414E70, w, start)
		for _, groups := range []byte{0, 2, 9} {
			w = prefix()
			w.word(0)
			w.str([]byte("WALL"))
			w.b(make([]byte, 14)...)
			w.aligned(groups)
			for i := byte(0); i < groups; i++ {
				w.str([]byte{byte('A' + i)})
			}
			for i := 0; i < 3; i++ {
				w.str([]byte{byte('a' + i)})
			}
			w.b(0)
			for i := 0; i < 15; i++ {
				n := byte(0)
				if i == 0 {
					n = 1
				}
				w.aligned(n)
				if n != 0 {
					for k := 0; k < 4; k++ {
						w.b(make([]byte, 8)...)
						w.ref(k&1 != 0)
					}
				}
			}
			w.word(0x454e4420)
			add(fmt.Sprintf("wall-groups-%d-start-%d", groups, start), 0x414F60, w, start)
		}
		for _, entry := range []uint32{0x414D40, 0x415100, 0x415240, 0x415320, 0x415660} {
			for _, count := range []uint32{0, 0x80000000} {
				w = prefix()
				w.word(count)
				add(fmt.Sprintf("empty-%x-%x-%d", entry, count, start), entry, w, start)
			}
		}
		w = prefix()
		w.word(2)
		for i := 0; i < 2; i++ {
			w.str([]byte("snd"))
			w.b(make([]byte, 9)...)
			w.str([]byte("a"))
			w.str([]byte("bb"))
			w.b(0)
		}
		add(fmt.Sprintf("skip-audio-two-%d", start), 0x414D40, w, start)
		w = prefix()
		w.word(1)
		w.str([]byte("spell"))
		w.b(1, 2, 3)
		w.str([]byte("value"))
		w.ref(false)
		w.ref(true)
		w.word(7)
		w.str([]byte("payload"))
		w.short(3)
		w.b(4, 5, 6)
		for i := 0; i < 3; i++ {
			w.str([]byte{byte(i + 1)})
		}
		add(fmt.Sprintf("spells-record-%d", start), 0x415100, w, start)
		w = prefix()
		w.word(3)
		w.str([]byte("one"))
		w.b(0)
		w.ref(true)
		w.str([]byte("empty"))
		w.b(2, 0, 8)
		w.str(nil)
		w.str([]byte("two"))
		w.b(2, 2, 0)
		w.str([]byte("more"))
		w.ref(false)
		w.ref(true)
		add(fmt.Sprintf("image-records-%d", start), 0x415240, w, start)
		w = prefix()
		w.word(1)
		w.str([]byte("ability"))
		w.b(0x77)
		w.ref(true)
		w.ref(false)
		w.ref(true)
		w.str([]byte("p"))
		w.short(2)
		w.b(1, 2)
		for i := 0; i < 3; i++ {
			w.str([]byte{byte(i + 1)})
		}
		add(fmt.Sprintf("ability-record-%d", start), 0x415320, w, start)
		w = prefix()
		w.word(2)
		for i := 0; i < 2; i++ {
			w.str([]byte("__NO_THING_PORT_SOUND__"))
			w.b(make([]byte, 9)...)
			w.str([]byte("one.wav"))
			w.str([]byte("two.wav"))
			w.b(0)
		}
		add(fmt.Sprintf("audio-real-missing-sound-%d", start), 0x415660, w, start)
	}
	// Last legal native row exercises 64-byte tile stride and 60-byte EDGE
	// records; full native image pointers in other rows remain byte-canaries.
	last := out[0]
	last.name = "floor-last-slot"
	last.floor = 175
	out = append(out, last)
	for _, x := range out {
		if x.entry == 0x411850 && bytes.Contains([]byte(x.name), []byte("edge-end-false")) {
			x.name = "edge-last-slot"
			x.edge = 63
			out = append(out, x)
			break
		}
	}
	return out
}
