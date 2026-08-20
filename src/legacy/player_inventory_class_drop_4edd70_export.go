package legacy

/*
#include "player_inventory_class_drop_4edd70.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func dropPlayerInventoryClassCall4EDD70() int32 {
	GetServer().S().DropPlayerInventoryClass4EDD70(
		server.DropPlayerInventoryClassRuntime4EDD70{
			Dispatch: objectDropDispatchCall4ED790,
		},
	)
	return 0
}

// Sub_4EDD70 preserves the existing Go call surface while bypassing the raw
// ABI32 C implementation.
func Sub_4EDD70() {
	_ = dropPlayerInventoryClassCall4EDD70()
}

//export sub_4EDD70
func sub_4EDD70() C.int {
	return C.int(dropPlayerInventoryClassCall4EDD70())
}
