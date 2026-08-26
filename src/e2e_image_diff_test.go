package opennox

import (
	"bytes"
	"testing"
)

func TestImageDiffHandlesResolutionChanges(t *testing.T) {
	got := imageDiff(
		[]byte{10, 20, 30, 40, 50, 60, 70, 80},
		[]byte{5, 25, 30, 50},
	)
	want := []byte{50, 50, 0, 155, 50, 60, 70, 255}
	if !bytes.Equal(got, want) {
		t.Fatalf("imageDiff = %v, want %v", got, want)
	}

	got = imageDiff([]byte{10, 20, 30, 40}, []byte{5, 25, 30, 50, 60, 70, 80, 90})
	want = []byte{50, 50, 0, 155}
	if !bytes.Equal(got, want) {
		t.Fatalf("imageDiff with larger reference = %v, want %v", got, want)
	}
}
