package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func objectDropBoundedRuntime4ED810() server.ObjectDropBoundedRuntime4ED810 {
	return server.ObjectDropBoundedRuntime4ED810{
		LoadCrownTypeCache: func() uint32 {
			return *memmap.PtrUint32(0x5D4594, dropOwnedCrownsTypeCacheOffset4ED050)
		},
		StoreCrownTypeCache: func(value uint32) {
			*memmap.PtrUint32(0x5D4594, dropOwnedCrownsTypeCacheOffset4ED050) = value
		},
		Dispatch: objectDropDispatchCall4ED790,
	}
}

func objectDropBoundedCall4ED810(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	return GetServer().S().ObjectDropBounded4ED810(
		owner,
		item,
		point,
		objectDropBoundedRuntime4ED810(),
	)
}
