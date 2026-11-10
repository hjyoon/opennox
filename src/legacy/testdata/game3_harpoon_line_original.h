int sub_4B63B0(int2* a1, int2* a2) {
	int v2; // ebx
	int v3; // ebp

	nox_client_drawSetColor_434460(*getMemIntPtr(0x5D4594, 1312492));
	nox_client_drawAddPoint_49F500(a1->field_0, a1->field_4);
	nox_client_drawAddPoint_49F500(a2->field_0, a2->field_4);
	nox_client_drawLineFromPoints_49E4B0();
	v2 = a2->field_0 - a1->field_0;
	v3 = a2->field_4 - a1->field_4;
	nox_client_drawSetColor_434460(*getMemIntPtr(0x5D4594, 1312496));
	if (v2 < 0) {
		v2 = -v2;
	}
	if (v3 < 0) {
		v3 = -v3;
	}
	if (v2 <= v3) {
		nox_client_drawAddPoint_49F500(a1->field_0 - 1, a1->field_4);
		nox_client_drawAddPoint_49F500(a2->field_0 - 1, a2->field_4);
		nox_client_drawLineFromPoints_49E4B0();
		nox_client_drawAddPoint_49F500(a1->field_0 + 1, a1->field_4);
		nox_client_drawAddPoint_49F500(a2->field_0 + 1, a2->field_4);
	} else {
		nox_client_drawAddPoint_49F500(a1->field_0, a1->field_4 - 1);
		nox_client_drawAddPoint_49F500(a2->field_0, a2->field_4 - 1);
		nox_client_drawLineFromPoints_49E4B0();
		nox_client_drawAddPoint_49F500(a1->field_0, a1->field_4 + 1);
		nox_client_drawAddPoint_49F500(a2->field_0, a2->field_4 + 1);
	}
	return nox_client_drawLineFromPoints_49E4B0();
}
