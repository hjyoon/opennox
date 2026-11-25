unsigned char nox_xxx_tileReadOne_422A40(int version, uint32_t* fields, nox_tile_list_node_t** head) {
	(void)version;
	uint32_t byte_value = (uint8_t)fields[0];
	uint32_t short_value = (uint16_t)fields[1];
	nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
	nox_xxx_fileReadWrite_426AC0_file3_fread(&short_value, 2u);
	fields[0] = (uint8_t)byte_value;
	fields[1] = (int16_t)short_value;
	byte_value = (uint8_t)fields[2];
	nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
	fields[2] = (uint8_t)byte_value;
	byte_value = (uint8_t)fields[3];
	nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
	fields[3] = (uint8_t)byte_value;

	uint32_t count = 0;
	for (nox_tile_list_node_t* node = *head; node; node = node->next) {
		++count;
	}
	nox_xxx_fileReadWrite_426AC0_file3_fread(&count, 1u);
	if (nox_crypt_IsReadOnly()) {
		*head = NULL;
		nox_tile_list_node_t** tail = head;
		for (uint32_t i = 0; i < (uint8_t)count; ++i) {
			nox_tile_list_node_t* node = nox_xxx_tileListAddNewSubtile_422160(0, 0, 0, 0);
			if (!node) {
				return 0;
			}
			byte_value = 0;
			short_value = 0;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
			nox_xxx_fileReadWrite_426AC0_file3_fread(&short_value, 2u);
			node->field_0 = (uint8_t)byte_value;
			node->field_1 = (int16_t)short_value;
			byte_value = 0;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
			node->field_2 = (uint8_t)byte_value;
			byte_value = 0;
			nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
			node->field_3 = (uint8_t)byte_value;
			*tail = node;
			tail = &node->next;
		}
		return 1;
	}
	for (nox_tile_list_node_t* node = *head; node; node = node->next) {
		byte_value = (uint8_t)node->field_0;
		short_value = (uint16_t)node->field_1;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
		nox_xxx_fileReadWrite_426AC0_file3_fread(&short_value, 2u);
		byte_value = (uint8_t)node->field_2;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
		byte_value = (uint8_t)node->field_3;
		nox_xxx_fileReadWrite_426AC0_file3_fread(&byte_value, 1u);
	}
	return 1;
}
