package legacy

import (
	"strconv"
	"testing"
)

func TestMapgenOrientObject527C60PreservesNativePointers(t *testing.T) {
	got := mapgenOrientObjectFixture527C60()
	wantFrames := [4]uint32{0, 24, 8, 16}
	for i, want := range wantFrames {
		if got.nativeResult[i] != 1 {
			t.Fatalf("direction case %d returned %d, want 1", i, got.nativeResult[i])
		}
		for slot, frame := range got.nativeFrames[i] {
			if frame != want {
				t.Fatalf("direction case %d frame slot %d = %d, want %d", i, slot, frame, want)
			}
		}
	}
	if got.invalidDirectionResult != 0 {
		t.Fatalf("invalid direction returned %d, want 0", got.invalidDirectionResult)
	}
	if got.missingUpdateResult != 0 {
		t.Fatalf("missing update data returned %d, want 0", got.missingUpdateResult)
	}
	if got.legacyResult != 1 || got.legacyFrames != [3]uint32{24, 24, 24} {
		t.Fatalf("legacy token wrapper = result %d frames %v, want result 1 frames [24 24 24]", got.legacyResult, got.legacyFrames)
	}
	if got.objectToken == 0 {
		t.Fatal("legacy wrapper received a zero object token")
	}
	if strconv.IntSize == 64 {
		if got.objectAddress <= 1<<32 || got.updateAddress <= 1<<32 {
			t.Fatalf("fixture addresses = object %#x update %#x, want both above PE32 range", got.objectAddress, got.updateAddress)
		}
		if uintptr(got.objectToken) == got.objectAddress {
			t.Fatalf("object token %#x unexpectedly equals native address %#x", got.objectToken, got.objectAddress)
		}
	}
}
