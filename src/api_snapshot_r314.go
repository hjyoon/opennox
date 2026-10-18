package opennox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type apiSnapshotHTTPError struct {
	status  int
	message string
}

func (err apiSnapshotHTTPError) Error() string { return err.message }

// Encode live world objects on the server loop. Only owned JSON bytes may
// cross back to the HTTP goroutine; object pointers can expire on the next tick.
func writeJSONRespInLoop(w http.ResponseWriter, r *http.Request, queue func(context.Context, func()), snapshot func() any) {
	writeJSONRespInLoopWithError(w, r, queue, func() (any, error) { return snapshot(), nil })
}

func writeJSONRespInLoopWithError(w http.ResponseWriter, r *http.Request, queue func(context.Context, func()), snapshot func() (any, error)) {
	ctx := r.Context()
	if ctx.Err() != nil {
		return
	}
	type response struct {
		body []byte
		err  error
	}
	done := make(chan response, 1)
	queue(ctx, func() {
		if ctx.Err() != nil {
			return
		}
		var buf bytes.Buffer
		value, err := snapshot()
		if err == nil {
			enc := json.NewEncoder(&buf)
			enc.SetIndent("", "\t")
			err = enc.Encode(value)
		}
		done <- response{body: buf.Bytes(), err: err}
	})
	select {
	case <-ctx.Done():
		return
	case v := <-done:
		if ctx.Err() != nil {
			return
		}
		if v.err != nil {
			var httpErr apiSnapshotHTTPError
			if errors.As(v.err, &httpErr) {
				http.Error(w, httpErr.message, httpErr.status)
				return
			}
			http.Error(w, "Cannot encode world snapshot", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(v.body)
	}
}
