//go:build amd64 || arm64

package legacy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func TestRWInvLightPayload4F5AA0Version60WriteOrder(t *testing.T) {
	light := make([]byte, client.DrawableLightXferSize)
	for i := range light {
		light[i] = byte(i)
	}
	path := filepath.Join(t.TempDir(), "invlight.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	apply, err := rwInvLightPayload4F5AA0(cf, light, 60)
	if err != nil {
		t.Fatal(err)
	}
	if apply {
		t.Fatal("write path requested a preview payload apply")
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := make([]byte, 0, 137)
	want = append(want, light[0:36]...)
	want = append(want, light[40:132]...)
	want = append(want, light[134:139]...)
	want = append(want, light[36:40]...)
	if !bytes.Equal(got, want) {
		t.Fatalf("version 60 payload = %x, want %x", got, want)
	}
}

func TestRWInvLightPayload4F5AA0RejectsWrongBlockSize(t *testing.T) {
	if _, err := rwInvLightPayload4F5AA0(nil, make([]byte, client.DrawableLightXferSize-1), 60); err == nil {
		t.Fatal("short light block was accepted")
	}
}
