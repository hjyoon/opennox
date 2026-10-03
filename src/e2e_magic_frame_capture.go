package opennox

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
)

// CaptureMagicFrame records the already-rendered frame for diagnosis only.
// Magic observers must assert their own live pixels, packets and expiry first:
// unrelated torches, particles and map animations cannot be fixed-frame goldens.
// The output is a fresh private temporary file, never a repository baseline.
func (sc *e2eScenario) CaptureMagicFrame(name string) {
	sc.add(0, name, func() {
		img := noxClient.r.CopyPixBuffer()
		path, err := e2eWriteMagicFrame("", img)
		if err != nil {
			e2eError(fmt.Errorf("capture magic frame %q: %w", name, err))
			return
		}
		e2eLog.Printf("MAGIC FRAME CAPTURE: name=%s frame=%d path=%s", name, noxServer.Frame(), path)
	})
}

func e2eWriteMagicFrame(dir string, img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, "opennox-e2e-magic-frame-*.png")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}
