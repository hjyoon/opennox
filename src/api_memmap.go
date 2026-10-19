package opennox

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"

	"github.com/opennox/opennox/v1/common/memmap"
)

func init() {
	http.HandleFunc("/debug/nox/memmap/validate", func(w http.ResponseWriter, r *http.Request) {
		memmap.ValidateZeros()
	})
	http.HandleFunc("/debug/nox/memmap/snapshot", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctx.Err() != nil {
			return
		}
		done := make(chan []byte, 1)
		noxServer.QueueInLoop(ctx, func() {
			if ctx.Err() != nil {
				return
			}
			// Blob.Data borrows mutable native memory. Complete the archive on
			// the loop so HTTP writes only owned bytes after the next tick.
			var buf bytes.Buffer
			zw := zip.NewWriter(&buf)
			for _, b := range memmap.Blobs() {
				w, err := zw.Create(fmt.Sprintf("blob_%X.dat", b.Addr))
				if err != nil {
					break
				}
				w.Write(b.Data)
			}
			zw.Close()
			done <- buf.Bytes()
		})
		select {
		case <-ctx.Done():
			return
		case body := <-done:
			if ctx.Err() != nil {
				return
			}
			w.Write(body)
		}
	})
}
