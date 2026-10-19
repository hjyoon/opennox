package opennox

import (
	"encoding/json"
	"net/http"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func init() {
	http.HandleFunc("/debug/nox/flags", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		enc.Encode(struct {
			Game   noxflags.GameFlag   `json:"game"`
			Engine noxflags.EngineFlag `json:"engine"`
		}{
			Game:   noxflags.GetGame(),
			Engine: noxflags.GetEngine(),
		})
	})
	http.HandleFunc("/debug/nox/flags/game", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var set, unset noxflags.GameFlag
		if v := q.Get("set"); v != "" {
			if v, err := noxflags.ParseGameFlag(v); err == nil {
				set = v
			}
		}
		if v := q.Get("unset"); v != "" {
			if v, err := noxflags.ParseGameFlag(v); err == nil {
				unset = v
			}
		}
		var value noxflags.GameFlag
		if set == 0 && unset == 0 {
			value = noxflags.GetGame()
		} else {
			ctx := r.Context()
			if ctx.Err() != nil {
				return
			}
			done := make(chan noxflags.GameFlag, 1)
			noxServer.QueueInLoop(ctx, func() {
				if ctx.Err() != nil {
					return
				}
				// Game flag hooks can access native GUI and live server state.
				// Only the copied flag value may return to the HTTP encoder.
				noxflags.SetGame(set)
				noxflags.UnsetGame(unset)
				done <- noxflags.GetGame()
			})
			select {
			case <-ctx.Done():
				return
			case value = <-done:
				if ctx.Err() != nil {
					return
				}
			}
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		enc.Encode(value)
	})
	http.HandleFunc("/debug/nox/flags/engine", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if v := q.Get("set"); v != "" {
			if v, err := noxflags.ParseEngineFlag(v); err == nil {
				noxflags.SetEngine(v)
			}
		}
		if v := q.Get("unset"); v != "" {
			if v, err := noxflags.ParseEngineFlag(v); err == nil {
				noxflags.UnsetEngine(v)
			}
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		enc.Encode(noxflags.GetEngine())
	})
}
