package opennox

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"image"

	"github.com/opennox/libs/strman"
)

type e2eScreenReference struct {
	TextID     strman.ID         `yaml:"text-id"`
	TextSHA256 string            `yaml:"text-sha256"`
	Frames     map[string]string `yaml:"frames"`
}

func (r *e2eScreenReference) expectedFrame(name string, strings *strman.StringManager) string {
	if r == nil || r.Frames[name] == "" {
		return ""
	}
	entry, ok := strings.Get(r.TextID)
	if !ok || len(entry.Vals) != 1 {
		return ""
	}
	digest := sha256.Sum256([]byte(entry.Vals[0].Str))
	if hex.EncodeToString(digest[:]) == r.TextSHA256 {
		return r.Frames[name]
	}
	return ""
}

func e2eScreenDigest(frame *image.NRGBA) string {
	var dimensions [8]byte
	binary.LittleEndian.PutUint32(dimensions[:4], uint32(frame.Rect.Dx()))
	binary.LittleEndian.PutUint32(dimensions[4:], uint32(frame.Rect.Dy()))
	digest := sha256.New()
	digest.Write(dimensions[:])
	digest.Write(frame.Pix)
	return hex.EncodeToString(digest.Sum(nil))
}
