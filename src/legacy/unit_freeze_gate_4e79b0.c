#include "GAME3_3.h"

enum {
	unit_freeze_gate_base_4e79b0 = 0x5D4594,
	unit_freeze_gate_offset_4e79b0 = 1567712,
};

//----- (004E79B0) --------------------------------------------------------
uint32_t sub_4E79B0(uint32_t value) {
	*getMemU32Ptr(unit_freeze_gate_base_4e79b0, unit_freeze_gate_offset_4e79b0) = value;
	return value;
}
