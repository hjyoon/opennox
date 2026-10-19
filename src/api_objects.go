package opennox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

var _ json.Marshaler = &Object{}
var _ json.Marshaler = &server.ObjectType{}

func (obj *Object) MarshalJSON() ([]byte, error) {
	return obj.SObj().MarshalJSON()
}

func init() {
	http.HandleFunc("/debug/nox/types", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
			return noxServer.Types.List()
		})
	})
	http.HandleFunc("/debug/nox/objects", func(w http.ResponseWriter, r *http.Request) {
		qu := r.URL.Query()
		indStr := qu.Get("ind")
		if indStr == "" {
			writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
				return noxServer.Objs.AllObjects()
			})
			return
		}
		ind, err := strconv.ParseInt(indStr, 10, 32)
		if err != nil {
			http.Error(w, "Bad index", http.StatusBadRequest)
			return
		}
		writeJSONRespInLoopWithError(w, r, noxServer.QueueInLoop, func() (any, error) {
			obj := asObjectS(noxServer.Objs.GetObjectByInd(int(ind)))
			if obj == nil {
				return nil, apiSnapshotHTTPError{status: http.StatusFound, message: "Not found"}
			}
			all := obj.Flags()
			for _, word := range strings.Split(qu.Get("flags"), ",") {
				fname := word
				unset := false
				if strings.HasPrefix(fname, "-") {
					unset = true
					fname = fname[1:]
				}
				fl, err := object.ParseFlag(fname)
				if err != nil {
					return nil, apiSnapshotHTTPError{status: http.StatusBadRequest, message: err.Error()}
				}
				if unset {
					all &^= fl
				} else {
					all |= fl
				}
			}
			obj.ObjFlags = all
			return obj, nil
		})
	})
	http.HandleFunc("/debug/nox/objects/updatable", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
			var out []*Object
			for p := noxServer.Objs.UpdatableList; p != nil; p = p.Next() {
				out = append(out, asObjectS(p))
			}
			return out
		})
	})
	http.HandleFunc("/debug/nox/objects/non-missiles", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
			return noxServer.Objs.AllNonMissiles()
		})
	})
	http.HandleFunc("/debug/nox/objects/missiles", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
			return noxServer.Objs.AllMissiles()
		})
	})
	http.HandleFunc("/debug/nox/objects/index", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRespInLoop(w, r, noxServer.QueueInLoop, func() any {
			m := make(map[string]server.DebugMapIndex)
			for k, v := range noxServer.Map.DebugIndex() {
				m[fmt.Sprintf("%d,%d", k.X, k.Y)] = v
			}
			return m
		})
	})
}
