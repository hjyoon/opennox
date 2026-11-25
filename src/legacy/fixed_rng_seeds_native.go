package legacy

/*
#include "common/platform/platform.h"
*/
import "C"

// Fixed seeds used by the original map and Quest entry points.

//export sub_4EED30
func sub_4EED30() { nox_platform_srand(0x22d6) }

//export sub_4EF560
func sub_4EF560() { nox_platform_srand(0x22d7) }

//export sub_4EF570
func sub_4EF570() { nox_platform_srand(0x7da) }

//export sub_4F0630
func sub_4F0630() { nox_platform_srand(0x7db) }

//export sub_4F3E20
func sub_4F3E20() { nox_platform_srand(0x4e2a) }

//export sub_4F78D0
func sub_4F78D0() { nox_platform_srand(0x4e34) }

//export sub_4FB940
func sub_4FB940() { nox_platform_srand(0x143d) }

//export sub_4FB950
func sub_4FB950() { nox_platform_srand(0x22ea) }

//export sub_4FC560
func sub_4FC560() { nox_platform_srand(0x22eb) }

//export sub_4FC950
func sub_4FC950() { nox_platform_srand(0x1446) }

//export sub_4FE840
func sub_4FE840() { nox_platform_srand(0x1447) }
