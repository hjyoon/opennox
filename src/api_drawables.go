package opennox

import (
	"encoding/json"
	"net/http"

	"github.com/opennox/opennox/v1/client"
)

var _ json.Marshaler = &client.Drawable{}

func init() {
	http.HandleFunc("/debug/nox/drawable", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
			return noxClient.Objs.AllList1()
		})
	})
}
