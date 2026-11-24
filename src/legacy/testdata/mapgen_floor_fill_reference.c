// Retained C bodies for the R305 floor fill callback fixtures.
static nox_mapgen_decor_fill_legacy* nox_mapgenDecorFillResolve(uintptr_t ref) {
	return (nox_mapgen_decor_fill_legacy*)nox_mapgenLegacyPtrResolve(ref);
}

//----- (00524950) --------------------------------------------------------
int sub_524950(int a1, int a2, float* a3, int* a4) {
	float center[2];
	nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve((uint32_t)a2);
	if (!fill) {
		return 0;
	}

	nox_xxx_tileGetDefByName_51D4D0(fill->tile_name);
	sub_5245A0(a1, a3, *a4, a4[1]);
	sub_544020(fill->object_name);
	center[0] = (double)*a4 * ((double)0x1.04371ep+4f) + *a3;
	center[1] = (double)a4[1] * ((double)0x1.04371ep+4f) + a3[1];
	return sub_543680(&center[0]);
}

//----- (005249C0) --------------------------------------------------------
int sub_5249C0(int a1, int a2, float* a3, int* a4) {
	int* v4;    // esi
	int v5;     // ecx
	int v6;     // edx
	int v7;     // eax
	int v8;     // ebp
	double v9;  // st7
	int v10;    // edi
	double v11; // st7
	int v12;    // eax
	bool v13;   // cc
	int v14;    // eax
	int v15;    // edi
	int v16;    // ebp
	int v17;    // eax
	float v19;  // [esp+10h] [ebp-14h]
	float point[2];
	float center[2];
	int v24;    // [esp+34h] [ebp+10h]
	nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve((uint32_t)a2);
	if (!fill) {
		return 0;
	}

	nox_xxx_tileGetDefByName_51D4D0(fill->tile_name);
	v4 = a4;
	v5 = *a4;
	v6 = *a4 / 3;
	v7 = a4[1];
	v8 = 2 * v6;
	v24 = 0;
	v9 = (double)v6 * ((double)0x1.04371ep+5f);
	v10 = v5 - 2 * v6;
	v19 = v9;
	v11 = v9 + *a3;
	point[1] = a3[1];
	point[0] = v11;
	if (v7 / 2 >= 0) {
		do {
			sub_5245A0(a1, &point[0], v10, 1);
			v12 = *v4;
			v10 += 2;
			v13 = v10 < *v4;
			point[0] = point[0] - ((double)0x1.04371ep+5f);
			point[1] = point[1] + ((double)0x1.04371ep+5f);
			if (!v13) {
				v10 = v12;
				point[0] = *a3;
			}
			++v24;
		} while (v24 <= v4[1] / 2);
	}
	v14 = v4[1];
	v15 = *v4 - v8;
	v16 = 0;
	point[0] = v19 + *a3;
	point[1] = (double)(v14 - 1) * ((double)0x1.04371ep+5f) + a3[1];
	if (v14 / 2 > 0) {
		do {
			sub_5245A0(a1, &point[0], v15, 1);
			v17 = *v4;
			v15 += 2;
			v13 = v15 < *v4;
			point[0] = point[0] - ((double)0x1.04371ep+5f);
			point[1] = point[1] - ((double)0x1.04371ep+5f);
			if (!v13) {
				v15 = v17;
				point[0] = *a3;
			}
			++v16;
		} while (v16 < v4[1] / 2);
	}
	sub_544020(fill->object_name);
	center[0] = (double)*v4 * ((double)0x1.04371ep+4f) + *a3;
	center[1] = (double)v4[1] * ((double)0x1.04371ep+4f) + a3[1];
	return sub_543680(&center[0]);
}

//----- (00524B50) --------------------------------------------------------
void sub_524B50(int a1, int a2, float* a3, int* a4) {
	int* v4;        // esi
	int v5;         // eax
	int v7;         // ecx
	int v8;         // eax
	signed int v9;  // eax
	int v10;        // ebp
	signed int v11; // eax
	signed int v12; // eax
	int v13;        // ebp
	signed int v14; // eax
	float v15;      // eax
	int v16;        // eax
	signed int v17; // eax
	int v18;        // ebp
	signed int v19; // eax
	int v20;        // ebp
	float v21;      // ecx
	float point[2];
	float center[2];
	int v26;        // [esp+24h] [ebp+10h]
	signed int v27; // [esp+24h] [ebp+10h]
	nox_mapgen_decor_fill_legacy* fill = nox_mapgenDecorFillResolve((uint32_t)a2);
	if (!fill) {
		return;
	}

	v4 = a4;
	if (*a4 >= 3 && a4[1] >= 3) {
		nox_xxx_tileGetDefByName_51D4D0(fill->tile_name);
		v5 = *a4 - 2;
		point[0] = *a3 + ((double)0x1.04371ep+5f);
		point[1] = a3[1] + ((double)0x1.04371ep+5f);
		if (v5 < 1) {
			v5 = 1;
		}
		v7 = a4[1] - 2;
		if (v7 < 1) {
			v7 = 1;
		}
		sub_5245A0(a1, &point[0], v5, v7);
		v8 = *a4;
		if ((int)*a4 < 4) {
			if (v8 == 3) {
				v15 = a3[1];
				point[0] = *a3 + ((double)0x1.04371ep+5f);
				point[1] = v15;
				sub_5245A0(a1, &point[0], 1, 1);
				point[1] = point[1] + ((double)0x1.04371ep+6f);
				sub_5245A0(a1, &point[0], 1, 1);
			}
		} else {
			v9 = nox_xxx_mapGenRandFunc_526AC0(1, v8 - 3);
			v10 = v9;
			v11 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - v9 - 2);
			point[1] = a3[1];
			point[0] = (double)v11 * ((double)0x1.04371ep+5f) + *a3;
			sub_5245A0(a1, &point[0], v10, 1);
			v12 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - 3);
			v13 = v12;
			v14 = nox_xxx_mapGenRandFunc_526AC0(1, *a4 - v12 - 2);
			v26 = a4[1] - 1;
			point[0] = (double)v14 * ((double)0x1.04371ep+5f) + *a3;
			point[1] = (double)v26 * ((double)0x1.04371ep+5f) + a3[1];
			sub_5245A0(a1, &point[0], v13, 1);
		}
		v16 = v4[1];
		if (v16 < 4) {
			if (v16 == 3) {
				v21 = *a3;
				point[1] = a3[1] + ((double)0x1.04371ep+5f);
				point[0] = v21;
				sub_5245A0(a1, &point[0], 1, 1);
				point[0] = point[0] + ((double)0x1.04371ep+6f);
				sub_5245A0(a1, &point[0], 1, 1);
			}
		} else {
			v17 = nox_xxx_mapGenRandFunc_526AC0(1, v16 - 3);
			v18 = v17;
			v19 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - v17 - 2);
			point[0] = *a3;
			point[1] = (double)v19 * ((double)0x1.04371ep+5f) + a3[1];
			sub_524610(a1, &point[0], v18);
			v20 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - 3);
			v27 = nox_xxx_mapGenRandFunc_526AC0(1, v4[1] - v20 - 2);
			point[0] = (double)(*v4 - 1) * ((double)0x1.04371ep+5f) + *a3;
			point[1] = (double)v27 * ((double)0x1.04371ep+5f) + a3[1];
			sub_524610(a1, &point[0], v20);
		}
		sub_544020(fill->object_name);
		center[0] = (double)*v4 * ((double)0x1.04371ep+4f) + *a3;
		center[1] = (double)v4[1] * ((double)0x1.04371ep+4f) + a3[1];
		sub_543680(&center[0]);
	}
}
