package legacy

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type objectIsCoopPlayerPixieHooks struct {
	pixieTypeInd          func() int
	gameCoop              func() bool
	findParentChainPlayer func(*server.Object) *server.Object
}

func objectIsCoopPlayerPixie_4E5B80(obj *server.Object, hooks objectIsCoopPlayerPixieHooks) bool {
	pixieTypeInd := hooks.pixieTypeInd()
	if obj == nil || !obj.Class().Has(object.ClassMissile) {
		return false
	}
	if !hooks.gameCoop() || int(obj.TypeInd) != pixieTypeInd {
		return false
	}
	owner := hooks.findParentChainPlayer(obj)
	return owner != nil && owner.Class().Has(object.ClassPlayer)
}
