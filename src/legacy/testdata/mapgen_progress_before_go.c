void nox_xxx_mapGenSetFlags_5235F0(char a1) {
	unsigned char packet[3];

	// Dedicated servers have no client renderer for the progress display.
	if (nox_common_getEngineFlag(NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) {
		return;
	}
	if (!nox_common_gameFlags_check_40A5C0(0x200000)) {
		dword_5d4594_2487564 = nox_platform_get_ticks();
		if ((int32_t)dword_5d4594_2487568 > (int32_t)dword_5d4594_2487564) {
			dword_5d4594_2487568 = 0;
		}
		nox_input_pollEvents_4453A0();
		if ((int32_t)(dword_5d4594_2487564 - dword_5d4594_2487568) > *getMemIntPtr(0x587000, 254948)) {
			uint16_t sequence = *getMemU16Ptr(0x5D4594, 2487572);
			packet[0] = (unsigned char)a1;
			packet[1] = (unsigned char)sequence;
			packet[2] = (unsigned char)(sequence >> 8);
			++*getMemU32Ptr(0x5D4594, 2487572);
			nox_xxx_mapGenClientText_4A9D00(packet);
			dword_5d4594_2487568 = dword_5d4594_2487564;
		}
	}
}

