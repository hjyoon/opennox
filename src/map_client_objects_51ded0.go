package opennox

import (
	"fmt"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type saveClientObjectsHooks51DED0 struct {
	first         func() *client.Drawable
	serverVisible func(typeInd int) bool
	saveable      func(typeInd int) bool
	typeName      func(typeInd int) string
	newObject     func(typeName string) *server.Object
	saveObject    func(obj *server.Object) int
	freeObject    func(obj *server.Object)
}

// saveClientObjectsNative51DED0 mirrors GAME.EXE 0051DED0 without treating a
// native Drawable or Object as a PE32 int array. It materializes the client-
// only invisible drawables as temporary server objects in their original list
// order so ObjectData retains the exact on-disk representation.
func saveClientObjectsNative51DED0(h saveClientObjectsHooks51DED0) error {
	if h.first == nil || h.serverVisible == nil || h.saveable == nil || h.typeName == nil ||
		h.newObject == nil || h.saveObject == nil || h.freeObject == nil {
		return fmt.Errorf("incomplete client-object save hooks")
	}
	for dr := h.first(); dr != nil; dr = dr.NextPtr {
		typeInd := int(dr.TypeIDVal)
		if h.serverVisible(typeInd) || !h.saveable(typeInd) {
			continue
		}
		typeName := h.typeName(typeInd)
		if typeName == "" {
			return fmt.Errorf("client drawable type %d has no name", typeInd)
		}
		obj := h.newObject(typeName)
		if obj == nil {
			return fmt.Errorf("cannot allocate temporary %q object", typeName)
		}
		savedNetCode := obj.NetCode
		obj.PosVec.X = float32(dr.PosVec.X) + 0.5
		obj.PosVec.Y = float32(dr.PosVec.Y) + 0.5
		obj.Extent = dr.NetCode32
		obj.ScriptIDVal = int32(dr.NetCode32)
		obj.NetCode = dr.NetCode32
		obj.ObjFlags = dr.ObjFlags
		obj.Field5 = dr.Flags70Val
		h.saveObject(obj) // GAME.EXE intentionally ignores the nested return value.
		obj.NetCode = savedNetCode
		h.freeObject(obj)
	}
	return nil
}

func (s *Server) saveClientObjects51DED0(cf *cryptfile.CryptFile) error {
	return saveClientObjectsNative51DED0(saveClientObjectsHooks51DED0{
		first: noxClient.Objs.FirstList1,
		serverVisible: func(typeInd int) bool {
			return sub_4E3AD0(typeInd) != 0
		},
		saveable: sub_4E3B80,
		typeName: func(typeInd int) string {
			typ := noxClient.Things.TypeByInd(typeInd)
			if typ == nil {
				return ""
			}
			return typ.ID()
		},
		newObject: s.NewObjectByTypeID,
		saveObject: func(obj *server.Object) int {
			return nox_xxx_xfer_saveObj51DF90(cf, obj)
		},
		freeObject: func(obj *server.Object) {
			s.Objs.FreeObject(obj)
		},
	})
}
