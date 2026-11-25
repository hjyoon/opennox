char* nox_xxx_net_4263C0() {
	extern char* nox_stats_bulk_flush_native(void);
	return nox_stats_bulk_flush_native();
}

//----- (004264D0) --------------------------------------------------------
int sub_4264D0() {
	extern int nox_stats_session_report_native(int mode);
	extern int nox_stats_quest_end_native(void);
	if (!nox_common_gameFlags_check_40A5C0(0x2000) || nox_common_gameFlags_check_40A5C0(4096)) {
		if (nox_common_gameFlags_check_40A5C0(4096)) {
			*getMemU32Ptr(0x5D4594, 739416) = time(0) - dword_5d4594_600116;
		}
	} else {
		for (nox_playerInfo* player = nox_common_playerInfoGetFirst_416EA0(); player;
			 player = nox_common_playerInfoGetNext_416EE0(player)) {
			sub_425E90(player, 1);
		}
		*getMemU32Ptr(0x5D4594, 599504) = time(0) - dword_5d4594_600116;
	}
	if (!nox_common_gameFlags_check_40A5C0(4096)) {
		return nox_stats_session_report_native(2);
	}
	return nox_stats_quest_end_native();
}
