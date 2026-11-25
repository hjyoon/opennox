/* Exact six preserved current-C bodies. Private bounded test references only. */
nox_drawable* sub_4ABDA0(int a1, short a2, short a3, uint32_t* a4) {
	int v4;          // edi
	nox_drawable* result; // eax
	nox_drawable* v6;     // esi
	unsigned int v7; // edx
	int v8;          // eax
	float v9;        // [esp-14h] [ebp-148h]
	char v10;        // [esp+Fh] [ebp-125h]
	float v11;       // [esp+10h] [ebp-124h]
	int v12;         // [esp+14h] [ebp-120h]
	int v13[2];      // [esp+18h] [ebp-11Ch]
	float v14;       // [esp+20h] [ebp-114h]
	int v15;         // [esp+24h] [ebp-110h]
	int v16;         // [esp+28h] [ebp-10Ch]
	float v17;       // [esp+2Ch] [ebp-108h]
	int v18;         // [esp+30h] [ebp-104h]
	char v19[256];   // [esp+34h] [ebp-100h]

	nox_xxx_fileReadWrite_426AC0_file3_fread(&v15, 4u);
	*a4 += 4;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v16, 4u);
	*a4 += 4;
	if (a3 < 40 || a2 < 4) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(v13, 8u);
		*a4 += 8;
	} else {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v11, 4u);
		*a4 += 4;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v17, 4u);
		v9 = v11;
		*a4 += 4;
		v13[0] = nox_float2int(v9);
		v13[1] = nox_float2int(v17);
	}
	if (a3 >= 10) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v12, 1u);
		++*a4;
		nox_xxx_fileReadWrite_426AC0_file3_fread((unsigned char*)v19, (unsigned char)v12);
		*a4 += (unsigned char)v12;
	}
	if (a3 >= 20) {
		nox_xxx_fileReadWrite_426AC0_file3_fread((unsigned char*)&v10, 1u);
		++*a4;
	}
	if (a3 >= 30) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v12, 1u);
		++*a4;
	}
	if (a3 >= 40) {
		nox_xxx_cryptSeekCur_40E0A0(4);
		*a4 += 4;
		if (a2 >= 2) {
			if (a2 < 5) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 4u);
				*a4 += 4;
				v11 = v14;
			} else {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v11, 2u);
				*a4 += 2;
			}
			v4 = (unsigned short)(4 * LOWORD(v11));
			nox_xxx_cryptSeekCur_40E0A0(v4);
			*a4 += v4;
		}
		if (a2 >= 3) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v18, 4u);
			*a4 += 4;
		}
	}
	result = nox_xxx_spriteLoadAdd_45A360_drawable(a1, v13[0], v13[1]);
	v6 = result;
	if (result) {
		v7 = result->flags30 & 0xEEBF7E9D;
		result->flags30 = v16 | v7;
		v8 = result->flags70;
		LOBYTE(v8) = v8 & 0xA1;
		v6->flags70 = v18 | v8;
		v6->field_32 = v15;
		*(uint8_t*)&v6->field_7 = v10;
		if (!nox_common_gameFlags_check_40A5C0(1) && *(uint8_t*)&v6->field_7 &&
			(!(v6->flags28 & 0x10000000) || !nox_common_gameFlags_check_40A5C0(128))) {
			nox_xxx_createAtImpl_4191D0(*(uint8_t*)&v6->field_7, &v6->field_6, 0, v6->field_32, 0);
		}
		result = v6;
	}
	return result;
}

nox_drawable* nox_xxx_spriteLoadFromMap_4AC020(int thingInd, short a2, uint32_t* a3) {
	uint32_t* v3;      // esi
	short v4;          // di
	short v5;          // ax
	bool v6;           // cc
	int v7;            // eax
	bool v10;          // zf
	int v11;           // ecx
	int v12;           // ebx
	char v13;          // dl
	int v14;           // ebx
	int v15;           // ecx
	int v16;           // eax
	int v17;           // eax
	unsigned char v18; // cl
	int v19;           // [esp-30h] [ebp-60h]
	float v20;         // [esp-24h] [ebp-54h]
	int v21;           // [esp-14h] [ebp-44h]
	int v22;           // [esp-4h] [ebp-34h]
	int v23;           // [esp+Ch] [ebp-24h]
	int v24;           // [esp+10h] [ebp-20h]
	int v25;           // [esp+14h] [ebp-1Ch]
	int v26;           // [esp+18h] [ebp-18h]
	int v27;           // [esp+1Ch] [ebp-14h]
	int v28;           // [esp+20h] [ebp-10h]
	int v29;           // [esp+24h] [ebp-Ch]
	float v30;         // [esp+28h] [ebp-8h]
	float v31;         // [esp+2Ch] [ebp-4h]

	v3 = a3;
	v4 = a2;
	v5 = 1;
	v23 = 1;
	if (a2 < 40) {
		return sub_4ABDA0(thingInd, v5, v4, v3);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v23, 2u);
	v5 = v23;
	v6 = (short)v23 < 61;
	*v3 += 2;
	if (v6) {
		return sub_4ABDA0(thingInd, v5, v4, v3);
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v25, 4u);
	*v3 += 4;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v29, 4u);
	*v3 += 4;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v30, 4u);
	*v3 += 4;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&v31, 4u);
	v20 = v31;
	*v3 += 4;
	v22 = nox_float2int(v20);
	v7 = nox_float2int(v30);
	nox_drawable* v9 = nox_xxx_spriteLoadAdd_45A360_drawable(thingInd, v7, v22);
	if (!v9) {
		return 0;
	}

	v9->field_32 = v25;
	nox_xxx_fileReadWrite_426AC0_file3_fread(&a2, 1u);
	v10 = (uint8_t)a2 == 0;
	++*v3;
	if (!v10) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v26, 4u);
		v11 = v26;
		*v3 += 4;
		v9->flags30 = v11 | v9->flags30 & 0xEEBF7E9D;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&a3, 1u);
		v12 = (unsigned char)a3;
		v21 = (unsigned char)a3;
		++*v3;
		nox_xxx_cryptSeekCur_40E0A0(v21);
		*v3 += v12;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&thingInd, 1u);
		v13 = thingInd;
		++*v3;
		*(uint8_t*)&v9->field_7 = v13;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&a3, 1u);
		++*v3;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v27, 2u);
		v14 = 4 * (unsigned short)v27;
		v19 = 4 * (unsigned short)v27;
		*v3 += 2;
		nox_xxx_cryptSeekCur_40E0A0(v19);
		*v3 += v14;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v28, 4u);
		v15 = v28;
		*v3 += 4;
		v16 = v9->flags70;
		LOBYTE(v16) = v16 & 0xA1;
		v9->flags70 = v15 | v16;
		if ((short)v23 >= 63) {
			nox_xxx_cryptSeekCur_40E0A0(2);
			*v3 += 2;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v24, 4u);
			*v3 += 4;
			nox_xxx_cryptSeekCur_40E0A0(v24);
			*v3 += v24;
			nox_xxx_cryptSeekCur_40E0A0(4);
			*v3 += 4;
		}
		if ((short)v23 >= 64) {
			nox_xxx_cryptSeekCur_40E0A0(4);
			*v3 += 4;
		}
	}
	if (!nox_common_gameFlags_check_40A5C0(1)) {
		v17 = *getMemU32Ptr(0x5D4594, 1309788);
		if (!*getMemU32Ptr(0x5D4594, 1309788)) {
			v17 = nox_xxx_getTTByNameSpriteMB_44CFC0("FlagMarker");
			*getMemU32Ptr(0x5D4594, 1309788) = v17;
		}
		v18 = *(uint8_t*)&v9->field_7;
		if (v18) {
			if (v9->field_27 != v17) {
				nox_xxx_createAtImpl_4191D0(v18, &v9->field_6, 0, v9->field_32, 0);
			}
		}
	}
	return v9;
}

short sub_4B69F0(nox_drawable* dr) {
	char v1;      // cl
	short result; // ax
	int v3;       // edx

	v1 = dr->vel_z;
	dr->z += v1;
	result = dr->z;
	if (result >= 0) {
		if ((unsigned char)gameFrame() & 1) {
			dr->vel_z = v1 - 1;
		}
	} else {
		dr->z = -result;
		result = 26209 * v1;
		v3 = -9 * v1 / 10;
		dr->vel_z = v3;
		if ((char)v3 < 2) {
			dr->z = 0;
			dr->vel_z = 0;
		}
	}
	return result;
}

int sub_4AC7B0(int a1) {
	nox_drawable* v1; // eax
	nox_drawable* v2;
	int result; // eax
	int v4;     // eax
	char v5;    // al
	float v6;   // edi
	int v7;     // ecx
	double v9;  // st7
	double v10; // st7
	int v11;    // [esp+4h] [ebp-10h]
	int v12;    // [esp+8h] [ebp-Ch]
	float v13;  // [esp+Ch] [ebp-8h]
	int v14;    // [esp+10h] [ebp-4h]

	nox_xxx_fileReadWrite_426AC0_file3_fread(&v12, 2u);
	v11 = 2;
	v1 = nox_xxx_spriteLoadFromMap_4AC020(a1, v12, (uint32_t*)&v11);
	v2 = v1;
	if (v1) {
		v4 = v1->flags28;
		if ((v4 & 0x80u) == 0) {
			if (v4 & 0x200) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
				nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 4u);
				v9 = (double)a1;
				v11 += 8;
				v13 = v9;
				v2->shape.box_w = v9;
				v10 = (double)v14;
				v2->shape.box_h = v10;
				if (v13 > 60.0) {
					v2->shape.box_w = 60.0f;
				}
				if (v10 > 60.0) {
					v2->shape.box_h = 60.0f;
				}
				nox_shape_box_calc(&v2->shape);
			}
		} else {
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v13, 4u);
			v11 += 4;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&v14, 4u);
			v5 = v14;
			v11 += 4;
			v2->field_108_2 = v14;
			v2->field_108_1 = 0;
			if (v5) {
				v2->field_108_1 = 1;
			}
			v6 = v13;
			a1 = LODWORD(v13);
			if ((short)v12 >= 41) {
				nox_xxx_fileReadWrite_426AC0_file3_fread(&a1, 4u);
				v6 = *(float*)&a1;
				v11 += 4;
			}
			v2->field_74_4 = LOBYTE(v13);
			v7 = v2->pos.x + *getMemIntPtr(0x587000, 196184 + 8 * LODWORD(v6)) / 2;
			sub_410390(v2, v7 / 23,
					   (v2->pos.y + *getMemIntPtr(0x587000, 196188 + 8 * LODWORD(v6)) / 2) / 23);
		}
		v2->field_72 = 0;
		nox_xxx_spriteSetActiveMB_45A990_drawable(v2);
		result = v11;
	} else {
		nox_xxx_spriteLoadError_4356E0();
		result = 0;
	}
	return result;
}

int sub_4ACEF0(int a1) {
	nox_drawable* v1; // eax
	nox_drawable* v2; // edi
	int result;  // eax
	uint8_t* v4; // esi
	int v5;      // [esp+Ch] [ebp-10h]
	int v6;      // [esp+10h] [ebp-Ch]
	int v7;      // [esp+14h] [ebp-8h]
	uint8_t* v8; // [esp+18h] [ebp-4h]

	nox_xxx_fileReadWrite_426AC0_file3_fread(&v7, 2u);
	v5 = 2;
	v1 = nox_xxx_spriteLoadFromMap_4AC020(a1, v7, (uint32_t*)&v5);
	v2 = v1;
	if (v1) {
		v4 = &v1->field_108_1;
		v1->field_72 = 0;
		nox_xxx_spriteSetActiveMB_45A990_drawable(v1);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v6, 4u);
		v2->shape.box_w = (double)v6;
		v5 += 4;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v6, 4u);
		v2->shape.box_h = (double)v6;
		v5 += 4;
		nox_shape_box_calc(&v2->shape);
		v4[4] = 10;
		*v4 = 90;
		v4[1] = 90;
		v4[2] = 90;
		v4[3] = 10;
		v4[5] = 10;
		v8 = v4 + 5;
		if ((short)v7 >= 41) {
			nox_xxx_fileReadWrite_426AC0_file3_fread(v4, 1u);
			++v5;
			nox_xxx_fileReadWrite_426AC0_file3_fread(v4 + 1, 1u);
			++v5;
			nox_xxx_fileReadWrite_426AC0_file3_fread(v4 + 2, 1u);
			++v5;
			nox_xxx_fileReadWrite_426AC0_file3_fread(v4 + 3, 1u);
			++v5;
			nox_xxx_fileReadWrite_426AC0_file3_fread(v4 + 4, 1u);
			++v5;
			nox_xxx_fileReadWrite_426AC0_file3_fread(v8, 1u);
			++v5;
		}
		result = v5;
	} else {
		nox_xxx_spriteLoadError_4356E0();
		result = 0;
	}
	return result;
}

int sub_4B6970(uint32_t* a1, nox_drawable* dr, int a3, int a4) {
	int v4; // eax
	int v5; // eax

	v4 = dr->union_u32[2];
	dr->union_u32[0] += v4 * *getMemIntPtr(0x587000, 192088 + 8 * dr->field_74_4);
	v5 = v4 * *getMemIntPtr(0x587000, 192092 + 8 * dr->field_74_4) + dr->union_u32[1];
	dr->union_u32[1] = v5;
	nox_xxx_updateSpritePosition_49AA90(dr, dr->union_u32[0] >> 12, v5 >> 12);
	sub_4B69F0(dr);
	return sub_4B6880(a1, dr, a3, a4);
}

