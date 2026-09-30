package legacy

import "testing"

func TestInventoryViewportNativeFields(t *testing.T) {
	for _, tc := range []struct {
		name  string
		words [13]uint32
	}{
		{name: "stock 1024x768", words: [13]uint32{0, 0, 1024, 768, 0, 0, 0, 0, 1024, 768}},
		{name: "stock 640x480", words: [13]uint32{0, 0, 640, 480, 0, 0, 0, 0, 640, 480}},
		{name: "all original words", words: [13]uint32{10, 20, 300, 400, 50, 60, 700, 800, 290, 380, 0xfedcba98, 0x89abcdef, 12}},
		{name: "signed coordinates unsigned flags", words: [13]uint32{0x80000000, 0xffffffff, 0x7fffffff, 0xfffffffe, 0xfffffff0, 0x80000001, 0x80000002, 0x80000003, 0xffffffe0, 0xffffffd0, 0xffffffff, 0x80000000, 0xfffffff4}},
		{name: "refresh after nonzero words", words: [13]uint32{0, 0, 1280, 720, 0, 0, 0, 0, 1280, 720}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fields, unchanged := inventoryViewportNativeFields(tc.words)
			if !unchanged {
				t.Fatal("conversion modified the PE32 record or its neighboring globals")
			}
			for i, word := range tc.words {
				want := int64(int32(word))
				if i == 10 || i == 11 {
					want = int64(word)
				}
				if fields[i] != want {
					t.Errorf("native field %d = %d, want %d", i, fields[i], want)
				}
			}
		})
	}
}
