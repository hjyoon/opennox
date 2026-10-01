package legacy

import (
	"image"
	"testing"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type bubbleCreateClient499F60 struct {
	Client
	cli   *client.Client
	dr    *client.Drawable
	typ   int
	pos   image.Point
	calls int
}

func (c *bubbleCreateClient499F60) Cli() *client.Client { return c.cli }

func (c *bubbleCreateClient499F60) Nox_xxx_spriteLoadAdd_45A360_drawable(typ int, pos image.Point) *client.Drawable {
	c.typ, c.pos = typ, pos
	c.calls++
	return c.dr
}

func newBubbleCreateClient499F60(t *testing.T) *bubbleCreateClient499F60 {
	t.Helper()
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.SetFrame(100)
	c := &bubbleCreateClient499F60{cli: client.NewClient(nil, nil, s)}
	t.Cleanup(c.cli.Close)
	t.Cleanup(c.cli.GUI.DestroyAll)
	prev := GetClient
	GetClient = func() Client { return c }
	t.Cleanup(func() { GetClient = prev })
	for i := 0; i < 8; i++ {
		cell := memmap.PtrUint32(0x5D4594, 1217512+uintptr(i)*4)
		old := *cell
		*cell = uint32(100 + i)
		t.Cleanup(func() { *cell = old })
	}
	return c
}

func TestBubbleCreate499F60NativePayloadAndLists(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  int
		rgb  [3]byte
	}{
		{"red", 100, [3]byte{255, 128, 128}},
		{"white", 101, [3]byte{255, 255, 255}},
		{"light blue", 102, [3]byte{200, 200, 255}},
		{"orange", 103, [3]byte{255, 100, 50}},
		{"green", 104, [3]byte{64, 255, 64}},
		{"violet", 105, [3]byte{255, 100, 255}},
		{"light violet", 106, [3]byte{255, 200, 255}},
		{"yellow", 107, [3]byte{255, 255, 200}},
		{"unknown defaults to light blue", 200, [3]byte{200, 200, 255}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newBubbleCreateClient499F60(t)
			sentinel, freeSentinel := alloc.New(client.Drawable{})
			t.Cleanup(freeSentinel)
			dr, free := alloc.New(client.Drawable{})
			t.Cleanup(free)
			t.Logf("actual C drawable=%p next sprite=%p", dr, sentinel)
			dr.PosVec = image.Pt(-321, 654)
			dr.ObjClass = object.ClassMonster
			dr.ZVal, dr.ZVal2 = 17, 0xCAFE
			dr.LightColor = noxrender.RGB{R: 0x1AB, G: -2, B: 0x345}
			dr.NextPtr, dr.Field_94 = sentinel, sentinel
			dr.UnionEffect().Field_111 = 0xA5000000
			dr.UnionEffect().Field_112 = 0x11223344
			c.dr = dr
			c.cli.Objs.List1 = dr
			sentinel.NetCode32 = 0x1234
			sentinel.ObjClass = object.ClassMonster

			// Exercise the actual C entry and its real Go list callbacks, not a
			// surrogate payload setter. Scalar arguments retain the PE32 ABI.
			Sub_499F60(tc.typ, dr.PosVec, -123, -2, 0x180, -5, 0x1AB, -1, 7)

			// Check the vulnerable native links before following the list.
			if dr.NextPtr != sentinel || dr.Field_94 != sentinel {
				t.Fatalf("particle initialization corrupted native links: next=%p field94=%p; want %p", dr.NextPtr, dr.Field_94, sentinel)
			}
			if c.calls != 1 || c.typ != tc.typ || c.pos != dr.PosVec {
				t.Fatalf("allocation calls=%d type=%d position=%v", c.calls, c.typ, c.pos)
			}
			if dr.ZVal != uint16(0xFF85) || dr.ZVal2 != 0xCAFE || dr.ObjClass != object.ClassMonster {
				t.Fatalf("native height/adjacent fields = %#x/%#x/%#x", dr.ZVal, dr.ZVal2, dr.ObjClass)
			}
			if dr.LightColor != (noxrender.RGB{R: 0x1AB, G: -2, B: 0x345}) {
				t.Fatalf("native light color changed: %+v", dr.LightColor)
			}
			want := client.DrawableUnionEffect{
				Field_108: noxcolor.RGB5551Color(0xAB, 0xFE, 0x45).Color32(),
				Field_109: noxcolor.RGB5551Color(tc.rgb[0], tc.rgb[1], tc.rgb[2]).Color32(),
				Field_110: 0x808001FE,
				Field_111: 0xA5FBFFAB,
				Field_112: 0x11223344,
			}
			if got := *dr.UnionEffect(); got != want {
				t.Fatalf("effect payload = %+v; want %+v", got, want)
			}
			if c.cli.Objs.List6 != dr || c.cli.Objs.DeadlineList != dr || c.cli.Objs.List4 != dr ||
				dr.Deadline != 107 || dr.ObjFlags != 0x600000 {
				t.Fatalf("native list insertion failed: sight=%p decay=%p update=%p deadline=%d flags=%#x",
					c.cli.Objs.List6, c.cli.Objs.DeadlineList, c.cli.Objs.List4, dr.Deadline, dr.ObjFlags)
			}
			if got := c.cli.Objs.ByNetCodeDynamic(0x1234); got != sentinel {
				t.Fatalf("dynamic list lookup = %p; want unmodified next sprite %p", got, sentinel)
			}
		})
	}
}

func TestBubbleCreate499F60AllocationFailure(t *testing.T) {
	c := newBubbleCreateClient499F60(t)
	Sub_499F60(104, image.Pt(23, 45), 6, 3, 4, 1, 0, 0, 2)
	if c.calls != 1 || c.typ != 104 || c.pos != image.Pt(23, 45) {
		t.Fatalf("allocation calls=%d type=%d position=%v", c.calls, c.typ, c.pos)
	}
	if c.cli.Objs.List6 != nil || c.cli.Objs.DeadlineList != nil || c.cli.Objs.List4 != nil {
		t.Fatal("failed allocation inserted a sprite into a client list")
	}
}

func TestBubbleCreate499F60MissingTypeCacheStillAttemptsAllocation(t *testing.T) {
	c := newBubbleCreateClient499F60(t)
	*memmap.PtrUint32(0x5D4594, 1217512) = 0
	// With an empty real Things registry every name resolves to zero. The
	// original zero-cache path still attempts allocation after all lookups.
	Sub_499F60(200, image.Pt(23, 45), 6, 3, 4, 1, 0, 0, 2)
	for i := 0; i < 8; i++ {
		if got := *memmap.PtrUint32(0x5D4594, 1217512+uintptr(i)*4); got != 0 {
			t.Fatalf("missing type cache[%d]=%d; want zero", i, got)
		}
	}
	if c.calls != 1 || c.typ != 200 || c.pos != image.Pt(23, 45) {
		t.Fatalf("allocation calls=%d type=%d position=%v", c.calls, c.typ, c.pos)
	}
}

func TestBubbleCreate499F60ColorUsesFirstMatchingCache(t *testing.T) {
	c := newBubbleCreateClient499F60(t)
	dr, free := alloc.New(client.Drawable{})
	t.Cleanup(free)
	c.dr = dr
	*memmap.PtrUint32(0x5D4594, 1217528) = 100 // Red and green share a live type ID.
	Sub_499F60(100, image.Point{}, 1, 3, 4, 1, 0, 0, 2)
	if got, want := dr.UnionEffect().Field_109, noxcolor.RGB5551Color(255, 128, 128).Color32(); got != want {
		t.Fatalf("center color = %#x; want first (red) match %#x", got, want)
	}
}
