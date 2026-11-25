package opennox

import (
	"crypto/sha256"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/strman"
)

func TestE2EScreenReferenceMatchesLoadedText(t *testing.T) {
	reference := &e2eScreenReference{
		TextID:     "Briefing:Title",
		TextSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte("Intro"))),
		Frames:     map[string]string{"opening": "frame hash"},
	}
	path := filepath.Join(t.TempDir(), "strings.json")
	for _, test := range []struct {
		text string
		want string
	}{
		{"Intro", "frame hash"},
		{"소개", ""},
	} {
		if err := os.WriteFile(path, []byte(fmt.Sprintf(
			`{"lang":8,"entries":[{"id":"Briefing:Title","vals":[{"str":%q}]}]}`, test.text)), 0600); err != nil {
			t.Fatal(err)
		}
		strings := strman.New()
		if err := strings.ReadJSON(path); err != nil {
			t.Fatal(err)
		}
		if got := reference.expectedFrame("opening", strings); got != test.want {
			t.Errorf("reference for %q = %q, want %q", test.text, got, test.want)
		}
		if got := reference.expectedFrame("other frame", strings); got != "" {
			t.Errorf("unlisted frame uses reference %q", got)
		}
	}
}

func TestE2EScreenDigestIncludesDimensionsAndPixels(t *testing.T) {
	frame := image.NewNRGBA(image.Rect(0, 0, 2, 3))
	original := e2eScreenDigest(frame)
	if got := e2eScreenDigest(image.NewNRGBA(image.Rect(0, 0, 3, 2))); got == original {
		t.Error("changing frame dimensions kept the same digest")
	}
	frame.Pix[0] = 1
	if got := e2eScreenDigest(frame); got == original {
		t.Error("changing a pixel kept the same digest")
	}
}
