package opennox

import (
	"net/http"
)

func init() {
	http.HandleFunc("/debug/nox/waypoints", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
			return noxServer.WPs.All()
		})
	})
}
