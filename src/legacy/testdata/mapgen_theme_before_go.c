// Theme and spell name compatibility bodies from 36c78c5df.
//----- (0051E1D0) --------------------------------------------------------
int nox_xxx_mapGenSpellIdByName_51E1D0(const char* a1) {
	unsigned int i; // esi
	char v3[60];    // [esp+8h] [ebp-78h]
	char v4[60];    // [esp+44h] [ebp-3Ch]

	strcpy(v3, a1);
	for (i = 0; i < strlen(v3); ++i) {
		v3[i] = toupper(v3[i]);
	}
	nox_sprintf(v4, "SPELL_%s", v3);
	return nox_xxx_spellNameToN_4243F0(v4);
}
// 51E1D0: using guessed type char var_78[60];

//----- (0051E260) --------------------------------------------------------
int nox_mapgenReadThemeNative_51E260(int* a1, const char* a2) {
	FILE* v2;     // eax
	FILE* v3;     // edi
	int v4;       // eax
	int v5;       // ebx
	char v7[256]; // [esp+18h] [ebp-100h]

	memset(a1, 0, 0x45Cu);
	if (!nox_mapgenLegacyPtrRegister(a1)) {
		return 0;
	}
	nox_sprintf(v7, "mapgen/%s.thm", a2);
	a1[3] = 3;
	a1[5] = 3;
	a1[1] = 5;
	a1[2] = 2;
	a1[4] = 1;
	a1[6] = 40;
	a1[7] = 20;
	a1[8] = 10;
	a1[9] = 5;
	a1[10] = 20;
	a1[16] = 1157234688;
	a1[18] = 100;
	a1[19] = time(0);
	a1[14] = 1;
	a1[22] = 0;
	a1[11] = 0;
	a1[12] = 100;
	a1[13] = 100;
	a1[30] = 0;
	a1[38] = 0;
	a1[46] = 0;
	*getMemU32Ptr(0x5D4594, 2487520) = 0;
	v2 = nox_binfile_open_408CC0(v7, 0);
	v3 = v2;
	if (!v2 || !nox_binfile_cryptSet_408D40(v2, 1)) {
		return 0;
	}
	while (nox_xxx_mapGenReadLine_51E540(v3, getMemAt(0x5D4594, 2487264))) {
		if (!nox_strcmpi("ALGORITHM_DATA", (const char*)getMemAt(0x5D4594, 2487264))) {
			v4 = nox_mapgenReadAlgDataNative_51EBB0((uint8_t*)a1, v3);
			goto LABEL_23;
		}
		if (!nox_strcmpi("EXIT", (const char*)getMemAt(0x5D4594, 2487264))) {
			v4 = nox_mapgenReadExitNative_51F800((uint8_t*)a1, v3);
			goto LABEL_23;
		}
		if (!nox_strcmpi("PREFAB", (const char*)getMemAt(0x5D4594, 2487264))) {
			v4 = nox_xxx_genReadPrefab_520BF0((uint8_t*)a1, v3);
			goto LABEL_23;
		}
		if (!nox_strcmpi("AMBIENT_LIGHT", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (!nox_xxx_mapGenReadLine_51E540(v3, getMemAt(0x5D4594, 2487264)) ||
				(a1[134] = atoi((const char*)getMemAt(0x5D4594, 2487264)),
				 !nox_xxx_mapGenReadLine_51E540(v3, getMemAt(0x5D4594, 2487264))) ||
				(a1[135] = atoi((const char*)getMemAt(0x5D4594, 2487264)),
				 !nox_xxx_mapGenReadLine_51E540(v3, getMemAt(0x5D4594, 2487264)))) {
				v5 = 0;
				goto LABEL_27;
			}
			a1[136] = atoi((const char*)getMemAt(0x5D4594, 2487264));
			continue;
		}
		if (nox_strcmpi("SPELL_SET", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (nox_strcmpi("ARMOR_SET", (const char*)getMemAt(0x5D4594, 2487264))) {
				if (nox_strcmpi("WEAPON_SET", (const char*)getMemAt(0x5D4594, 2487264))) {
					if (nox_strcmpi("DECOR", (const char*)getMemAt(0x5D4594, 2487264))) {
						v5 = 0;
						goto LABEL_27;
					}
					v4 = nox_xxx_genReadDecor_51F9F0((uint32_t*)a1, v3);
				} else {
					v4 = nox_xxx_genReadWeaponSet_51F030((uint8_t*)a1, v3);
				}
			} else {
				v4 = nox_xxx_genReadArmorSet_51F640((uint8_t*)a1, v3);
			}
		} else {
			v4 = nox_mapgenReadSpellSetNative_51EFB0((uint8_t*)a1, v3);
		}
	LABEL_23:
		v5 = v4;
		if (!v4) {
			goto LABEL_27;
		}
	}
	v5 = 1;
LABEL_27:
	nox_binfile_close_408D90(v3);
	if (v5 != 1 || (nox_xxx_mapgenCheckSettings_520AD0(a1 + 22) && nox_xxx_mapgenCheckSettings_520AD0(a1 + 30) &&
					   (!a1[46] || nox_xxx_mapgenCheckSettings_520AD0(a1 + 46)))) {
		return v5;
	}
	return 0;
}

int nox_xxx_mapGenReadTheme_51E260(int* a1, int a2) {
	return nox_mapgenReadThemeNative_51E260(a1, (const char*)nox_mapgenLegacyPtrResolve((uint32_t)a2));
}

//----- (0051E540) --------------------------------------------------------
// Token reading and comment skipping at 51E540, 51E570 and 51E630
// are implemented in mapgen_token_reader_public.go.


// Theme section compatibility bodies from 4cc5069e4.
int nox_mapgenReadAlgDataNative_51EBB0(uint8_t* a1, FILE* a2) {
	int v2;            // edi
	char v7[60];           // [esp+10h] [ebp-3Ch]

	while (nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
		if (!nox_strcmpi("END", (const char*)getMemAt(0x5D4594, 2487264))) {
			return 1;
		}
		if (!nox_xxx_mapGenReadLine_51E540(a2, (uint8_t*)v7)) {
			break;
		}
		if (nox_strcmpi("midHallLength", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (nox_strcmpi("hallLengthVariance", (const char*)getMemAt(0x5D4594, 2487264))) {
				if (nox_strcmpi("midHallWidth", (const char*)getMemAt(0x5D4594, 2487264))) {
					if (nox_strcmpi("hallWidthVariance", (const char*)getMemAt(0x5D4594, 2487264))) {
						if (nox_strcmpi("hallLimit", (const char*)getMemAt(0x5D4594, 2487264))) {
							if (nox_strcmpi("hallBranchRate", (const char*)getMemAt(0x5D4594, 2487264))) {
								if (nox_strcmpi("hallRoomRate", (const char*)getMemAt(0x5D4594, 2487264))) {
									if (nox_strcmpi("midRoomSize", (const char*)getMemAt(0x5D4594, 2487264))) {
										if (nox_strcmpi("roomVariance", (const char*)getMemAt(0x5D4594, 2487264))) {
											if (nox_strcmpi("irregularRoomRate",
														 (const char*)getMemAt(0x5D4594, 2487264))) {
												if (nox_strcmpi("mapSize", (const char*)getMemAt(0x5D4594, 2487264))) {
													if (nox_strcmpi("recurs"
																 "ionLim"
																 "it",
																 (const char*)getMemAt(0x5D4594, 2487264))) {
														if (nox_strcmpi("seed",
																	 (const char*)getMemAt(0x5D4594, 2487264))) {
															if (nox_strcmpi("emptyRoomRate",
																		 (const char*)getMemAt(0x5D4594, 2487264))) {
																if (nox_strcmpi("mergeRate", (const char*)getMemAt(
																							  0x5D4594, 2487264))) {
																	if (nox_strcmpi("useDoors", (const char*)getMemAt(
																								 0x5D4594, 2487264))) {
																		if (nox_strcmpi("debug", (const char*)getMemAt(
																								  0x5D4594, 2487264))) {
																			if (nox_strcmpi("adjacentPortalRate",
																						 (const char*)getMemAt(
																							 0x5D4594, 2487264))) {
																				if (!nox_strcmpi("skeleton",
																							  (const char*)getMemAt(
																								  0x5D4594, 2487264))) {
																					v2 = 0;
																					const char* skeleton = (const char*)getMemPtr(0x587000, 253244);
																					while (skeleton) {
																						if (!nox_strcmpi(skeleton, v7)) {
																							break;
																						}
																						++v2;
																						skeleton = (const char*)getMemPtr(0x587000, 253244 + 4 * v2);
																					}
																					*(uint32_t*)a1 = skeleton ? (uint32_t)v2 : 0;
																				}
																			} else {
																				*(uint32_t*)(a1 + 52) = atoi(v7);
																			}
																		} else if (nox_strcmpi(v7, "true")) {
																			*(uint32_t*)(a1 + 60) = 0;
																		} else {
																			*(uint32_t*)(a1 + 60) = 1;
																		}
																	} else if (nox_strcmpi(v7, "true")) {
																		*(uint32_t*)(a1 + 56) = 0;
																	} else {
																		*(uint32_t*)(a1 + 56) = 1;
																	}
																} else {
																	*(uint32_t*)(a1 + 48) = atoi(v7);
																}
															} else {
																*(uint32_t*)(a1 + 44) = atoi(v7);
															}
														} else {
															*(uint32_t*)(a1 + 76) = atoi(v7);
														}
													} else {
														*(uint32_t*)(a1 + 72) = atoi(v7);
													}
												} else {
													*(float*)(a1 + 64) = atof(v7);
												}
											} else {
												*(uint32_t*)(a1 + 40) = atoi(v7);
											}
										} else {
											*(uint32_t*)(a1 + 36) = atoi(v7);
										}
									} else {
										*(uint32_t*)(a1 + 32) = atoi(v7);
									}
								} else {
									*(uint32_t*)(a1 + 28) = atoi(v7);
								}
							} else {
								*(uint32_t*)(a1 + 24) = atoi(v7);
							}
						} else {
							*(uint32_t*)(a1 + 20) = atoi(v7);
						}
					} else {
						*(uint32_t*)(a1 + 16) = atoi(v7);
					}
				} else {
					*(uint32_t*)(a1 + 12) = atoi(v7);
				}
			} else {
				*(uint32_t*)(a1 + 8) = atoi(v7);
			}
		} else {
			*(uint32_t*)(a1 + 4) = atoi(v7);
		}
	}
	return 0;
}

int nox_xxx_genReadAlgData_51EBB0(int a1, FILE* a2) {
	return nox_mapgenReadAlgDataNative_51EBB0((uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

//----- (0051EFB0) --------------------------------------------------------

int nox_mapgenReadSpellSetNative_51EFB0(uint8_t* a1, FILE* a2) {
	int v2; // eax

	while (1) {
		if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
			return 0;
		}
		if (!nox_strcmpi("END", (const char*)getMemAt(0x5D4594, 2487264)) || *(uint32_t*)(a1 + 1096) >= 137) {
			break;
		}
		v2 = nox_xxx_mapGenSpellIdByName_51E1D0((const char*)getMemAt(0x5D4594, 2487264));
		if (v2) {
			*(uint32_t*)(a1 + 4 * (*(uint32_t*)(a1 + 1096))++ + 548) = v2;
		}
	}
	return 1;
}

int nox_xxx_genReadSpellSet_51EFB0(int a1, FILE* a2) {
	return nox_mapgenReadSpellSetNative_51EFB0((uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

//----- (0051F030) --------------------------------------------------------

int nox_mapgenReadExitNative_51F800(uint8_t* a1, FILE* a2) {
	while (1) {
		if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
			return 0;
		}
		if (!nox_strcmpi("END", (const char*)getMemAt(0x5D4594, 2487264))) {
			break;
		}
		if (nox_strcmpi("OBJECT", (const char*)getMemAt(0x5D4594, 2487264))) {
			if (nox_strcmpi("LINKDATA", (const char*)getMemAt(0x5D4594, 2487264)) ||
				!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
				return 0;
			}
			strcpy((char*)(a1 + 476), (const char*)getMemAt(0x5D4594, 2487264));
		} else {
			if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
				return 0;
			}
			strcpy((char*)(a1 + 216 + (*(uint32_t*)(a1 + 472) << 6)), (const char*)getMemAt(0x5D4594, 2487264));
			if (!nox_xxx_mapGenReadLine_51E540(a2, getMemAt(0x5D4594, 2487264))) {
				return 0;
			}
			if (nox_strcmpi("NORTH", (const char*)getMemAt(0x5D4594, 2487264))) {
				if (nox_strcmpi("SOUTH", (const char*)getMemAt(0x5D4594, 2487264))) {
					if (nox_strcmpi("EAST", (const char*)getMemAt(0x5D4594, 2487264))) {
						if (!nox_strcmpi("WEST", (const char*)getMemAt(0x5D4594, 2487264))) {
							*(uint32_t*)((*(uint32_t*)(a1 + 472) << 6) + a1 + 216 + 60) = 3;
						}
						++*(uint32_t*)(a1 + 472);
					} else {
						*(uint32_t*)(((*(uint32_t*)(a1 + 472))++ << 6) + a1 + 216 + 60) = 2;
					}
				} else {
					*(uint32_t*)(((*(uint32_t*)(a1 + 472))++ << 6) + a1 + 216 + 60) = 1;
				}
			} else {
				*(uint32_t*)(((*(uint32_t*)(a1 + 472))++ << 6) + a1 + 216 + 60) = 0;
			}
		}
	}
	return 1;
}

int nox_xxx_genReadExit_51F800(int a1, FILE* a2) {
	return nox_mapgenReadExitNative_51F800((uint8_t*)nox_mapgenLegacyPtrResolve((uint32_t)a1), a2);
}

//----- (0051F9F0) --------------------------------------------------------
