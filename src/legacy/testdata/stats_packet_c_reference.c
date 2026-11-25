//----- (0042A8B0) --------------------------------------------------------
uint16_t* sub_42A8B0(uint8_t* a1, int* a2) {
	// This packet has one CNTL field. Keep the original node lifecycle and
	// serializer here: their int-based entry points cannot carry LP64 pointers.
	typedef struct packet_field {
		char name[4];
		uint16_t type, length;
		void* data;
		struct packet_field* next;
	} packet_field;
	struct {
		uint16_t length, type;
		packet_field* first;
	} packet = {0};
	uint32_t* length = (uint32_t*)a2;
	uint32_t initial_length = *length;
	uint8_t* compressed = calloc(1, (uint32_t)(initial_length * 2u));
	sub_42A970(a1, compressed, a2);
	uint8_t* encoded = sub_42AC50(compressed, (size_t*)a2);
	if (compressed) {
		free(compressed);
	}
	if (!encoded) {
		return 0; // The original empty-list destructor has no service calls.
	}

	uint16_t field_length = (uint16_t)*length;
	packet_field* field = calloc(1, sizeof(*field));
	if (field) {
		field->data = 0;
		free(field->data);
		strcpy(field->name, (const char*)getMemAt(0x5D4594, 741688));
		field->type = field->length = 0;
		field->data = 0;
		field->next = 0;
		const char* name = (const char*)getMemAt(0x587000, 71480);
		// Keep all four wire bytes and strncpy's padding for a short tag.
		for (size_t i = 0; i < sizeof(field->name); ++i) {
			field->name[i] = *name;
			if (*name) ++name;
		}
		field->type = 20;
		field->length = field_length;
		field->data = calloc(1, field_length);
		memcpy(field->data, encoded, field->length);
		field->next = 0;
	}
	// Preserve the original unguarded link after a failed node allocation.
	field->next = packet.first;
	packet.first = field;
	free(encoded);

	*length = 4;
	for (field = packet.first; field; field = field->next) {
		*length += 8;
		*length = field->length + *length + ((0u - ((field->length + *length) & 3u)) & 3u);
	}
	uint16_t* result = calloc(1, *length);
	result[0] = htons((uint16_t)*length);
	result[1] = htons(packet.type);
	uint8_t* output = (uint8_t*)(result + 2);
	for (field = packet.first; field; field = field->next) {
		field->type = htons(field->type);
		field->length = htons(field->length);
		memcpy(output, field->name, 8);
		output += 8;
		memcpy(output, field->data, ntohs(field->length));
		output += ntohs(field->length) + ((0u - (ntohs(field->length) & 3u)) & 3u);
		field->type = ntohs(field->type);
		field->length = ntohs(field->length);
	}
	for (field = packet.first; field;) {
		packet_field* next = field->next;
		free(field->data);
		strcpy(field->name, (const char*)getMemAt(0x5D4594, 741688));
		field->type = field->length = 0;
		field->data = 0;
		field->next = 0;
		free(field);
		field = next;
	}
	return result;
}

//----- (0042A970) --------------------------------------------------------
int sub_42A970(uint8_t* a1, uint8_t* a2, int* a3) {
	int v3;            // eax
	char v4;           // bl
	int v5;            // edx
	int v6;            // ecx
	int v7;            // esi
	int v8;            // eax
	int* v9;           // edx
	uint8_t* v10;      // ebp
	int v11;           // edi
	char v12;          // cl
	int result;        // eax
	int v14;           // edx
	int v15;           // eax
	int v16;           // ebp
	char v17;          // [esp+13h] [ebp-405h]
	unsigned char v18; // [esp+14h] [ebp-404h]
	int v19[256];      // [esp+18h] [ebp-400h]

	v3 = 0;
	v4 = 1;
	memset(v19, 0, sizeof(v19));
	v18 = 1;
	v5 = *a3;
	if (*a3 > 0) {
		do {
			v6 = (unsigned char)a1[v3++];
			++v19[v6];
		} while (v3 < v5);
	}
	v7 = v19[0];
	v17 = 0;
	v8 = 0;
	v9 = v19;
	while (*v9) {
		if (*v9 < v7) {
			v7 = *v9;
			v17 = v8;
		}
		++v8;
		++v9;
		if (v8 >= 256) {
			goto LABEL_10;
		}
	}
	v17 = v8;
LABEL_10:
	v10 = a1;
	v11 = 1;
	*a2 = v17;
	v12 = *a1;
	result = 1;
	v14 = *a3;
	if ((int)*a3 < 1) {
		*a3 = 1;
	} else {
		do {
			if (v10[v11] != (uint8_t)v12 || v11 >= v14) {
				if ((unsigned char)v4 <= 3u) {
					if ((int)v18 > 0) {
						v16 = v18;
						do {
							if (v12 == v17) {
								a2[result++] = v12;
							}
							a2[result++] = v12;
							--v16;
						} while (v16);
					}
				} else {
					a2[result] = v17;
					v15 = result + 1;
					a2[v15++] = v4;
					a2[v15] = v12;
					result = v15 + 1;
				}
				v10 = a1;
				v4 = 1;
				v12 = a1[v11];
			} else {
				++v4;
			}
			++v11;
			v18 = v4;
			v14 = *a3;
		} while (v11 <= *a3);
		*a3 = result;
	}
	return result;
}
// 42A970: using guessed type int var_400[256];

//----- (0042AAA0) --------------------------------------------------------
// GAME.EXE uses precision53/toward-zero. Reproduce this multiply locally;
// no global FPU state or the independent 4096-word gameplay RNG is changed.
static double nox_stats_mul_chop53_42AAA0(double a, double b) {
 double product = a * b;
 double error = fma(a, b, -product);
 if ((product > 0.0 && error < 0.0) || (product < 0.0 && error > 0.0)) {
  return nextafter(product, 0.0);
 }
 return product;
}

double sub_42AAA0(int* a1) {
	int v1;            // eax
	int v2;            // esi
	int v3;            // edi
	int v4;            // ecx
	int v5;            // edi
	int v6;            // esi
	unsigned char* v7; // ecx
	int v8;            // eax
	int v9;            // edx
	int v10;           // ecx
	int v11;           // eax
	int v13;           // [esp+Ch] [ebp+4h]

	if ((int)*a1 < 0 || !*getMemU32Ptr(0x5D4594, 741656)) {
		*getMemU32Ptr(0x5D4594, 741656) = 1;
		v1 = *a1;
		if ((int)*a1 < 0) {
			v1 = (int32_t)(0u - (uint32_t)v1);
		}
		v2 = 1;
		v3 = 21;
		v4 = (int32_t)(UINT32_C(161803398) - (uint32_t)v1) % INT32_C(1000000000);
		*getMemU32Ptr(0x5D4594, 741604) = (int32_t)(UINT32_C(161803398) - (uint32_t)v1) % INT32_C(1000000000);
		do {
			*getMemU32Ptr(0x5D4594, 741384 + 4 * (v3 % 55)) = v2;
			v2 = v4 - v2 + (v4 - v2 < 0 ? 0x3B9ACA00 : 0);
			v4 = *getMemU32Ptr(0x5D4594, 741384 + 4 * (v3 % 55));
			v3 += 21;
		} while (v3 <= 1134);
		v5 = 4;
		do {
			v6 = 1;
			v7 = getMemAt(0x5D4594, 741388);
			do {
				v8 = *(uint32_t*)v7 - *getMemU32Ptr(0x5D4594, 741388 + 4 * ((v6 + 30) % 55));
				*(uint32_t*)v7 = v8;
				if (v8 < 0) {
					*(uint32_t*)v7 = v8 + 1000000000;
				}
				v7 += 4;
				++v6;
			} while (v6 <= 55);
			--v5;
		} while (v5);
		dword_5d4594_741648 = 0;
		dword_5d4594_741652 = 31;
		*a1 = 1;
	}
	v9 = ++dword_5d4594_741648;
	if (dword_5d4594_741648 == 56) {
		v9 = 1;
		dword_5d4594_741648 = 1;
	}
	v10 = ++dword_5d4594_741652;
	if (dword_5d4594_741652 == 56) {
		v10 = 1;
		dword_5d4594_741652 = 1;
	}
	v11 = *getMemU32Ptr(0x5D4594, 741384 + 4 * v9) - *getMemU32Ptr(0x5D4594, 741384 + 4 * v10);
	v13 = v11;
	if (v11 < 0) {
		v11 += 1000000000;
		v13 = v11;
	}
	*getMemU32Ptr(0x5D4594, 741384 + 4 * v9) = v11;
	return nox_stats_mul_chop53_42AAA0((double)v13, *getMemDoublePtr(0x581450, 8368));
}

//----- (0042ABF0) --------------------------------------------------------
void sub_42ABF0(int a1, int a2, int a3) {
	double v3; // st7
	int i;     // esi

	if (a3 > 0) {
		a3 = -a3;
	}
	v3 = sub_42AAA0(&a3);
	for (i = 0; i < a2; ++i) {
		int v5 = (int)(v3 * 255.0);
		if (v5 < 0) {
			v5 = -v5;
		}
		*(uint8_t*)(i + a1) = v5;
		v3 = sub_42AAA0(&a3);
	}
}

//----- (0042AC50) --------------------------------------------------------
uint8_t* sub_42AC50(uint8_t* a1, size_t* a2) {
	// The legacy signature carries a DWORD length, not a native size_t field.
	uint32_t* length = (uint32_t*)a2;
	uint8_t* mask = calloc(1, *length);
	uint8_t* result = calloc(1, (uint32_t)(*length + 5u));
	if ((int32_t)*length < 15) {
		*length = UINT32_C(0xfffffffe);
		free(mask);
		free(result);
		return 0;
	}

	int32_t maskSeed = (int32_t)(uint32_t)time(0);
	if (maskSeed > 0) {
		maskSeed = (int32_t)(0u - (uint32_t)maskSeed);
	}
	uint32_t key = htonl((uint32_t)maskSeed);
	uint32_t maskCount = *length;
	// Inline 0042ABF0's loop: its int buffer argument truncates native pointers.
	// Preserve the initial RNG draw and the draw after the final mask byte.
	double random = sub_42AAA0(&maskSeed);
	for (uint32_t i = 0; (int32_t)i < (int32_t)maskCount; ++i) {
		int32_t value = (int32_t)(uint32_t)(int64_t)(random * 255.0);
		if (value < 0) {
			value = (int32_t)(0u - (uint32_t)value);
		}
		mask[i] = (uint8_t)value;
		random = sub_42AAA0(&maskSeed);
	}

	int32_t selectorSeed = (int32_t)(uint32_t)time(0);
	if (selectorSeed > 0) {
		selectorSeed = (int32_t)(0u - (uint32_t)selectorSeed);
	}
	float factor = (int32_t)*length >= 241 ? 241.0f : (float)(int32_t)(*length - 14u);
	uint8_t selector = (uint8_t)((uint64_t)(int64_t)(sub_42AAA0(&selectorSeed) * factor) + 10u);
	uint32_t pos = 0;
	for (uint32_t i = 0; (int32_t)i < (int32_t)*length; ++i) {
		if (pos == 5) {
			result[pos++] = selector;
		}
		if (pos == selector) {
			memcpy(result + pos, &key, sizeof(key));
			pos += 4;
		}
		result[pos++] = a1[i] ^ mask[i];
	}
	*length += 5u;
	free(mask);
	return result;
}

//----- (0042ADA0) --------------------------------------------------------
