package screen

/* TODO some images are more sensitive to textured fills, and will render incorrectly if the *exact* random
   number generator is not used. It looks like they targeted ~25% fill-rate for their textured fills. This
   doesn't look like its the 100% proper implementation, I'll have to try to hunt down something that is more
   accurate.
*/

/* Found at https://github.com/wjp/freesci-archive/blob/master/src/gfx/resource/sci_pic_0.c */
/* 'Random' fill patterns, provided by Carl Muckenhoupt: */
var noise = [...]bool{
	false, false, true, false, false, false, false, false,
	true, false, false, true, false, true, false, false,
	false, false, false, false, false, false, true, false,
	false, false, true, false, false, true, false, false,
	true, false, false, true, false, false, false, false,
	true, false, false, false, false, false, true, false,
	true, false, true, false, false, true, false, false,
	true, false, true, false, false, false, true, false,
	true, false, false, false, false, false, true, false,
	false, false, false, false, true, false, false, true,
	false, false, false, false, true, false, true, false,
	false, false, true, false, false, false, true, false,
	false, false, false, true, false, false, true, false,
	false, false, false, true, false, false, false, false,
	false, true, false, false, false, false, true, false,
	false, false, false, true, false, true, false, false,
	true, false, false, true, false, false, false, true,
	false, true, false, false, true, false, true, false,
	true, false, false, true, false, false, false, true,
	false, false, false, true, false, false, false, true,
	false, false, false, false, true, false, false, false,
	false, false, false, true, false, false, true, false,
	false, false, true, false, false, true, false, true,
	false, false, false, true, false, false, false, false,
	false, false, true, false, false, false, true, false,
	true, false, true, false, true, false, false, false,
	false, false, false, true, false, true, false, false,
	false, false, true, false, false, true, false, false,
	false, false, false, false, false, false, false, false,
	false, true, false, true, false, false, false, false,
	false, false, true, false, false, true, false, false,
	false, false, false, false, false, true, false,
}

/* Found at https://github.com/wjp/freesci-archive/blob/master/src/gfx/resource/sci_pic_0.c */
/* 'Random' fill offsets, provided by Carl Muckenhoupt: */
var vectorPatternTextureOffset = [...]int{
	0x00, 0x18, 0x30, 0xc4, 0xdc, 0x65, 0xeb, 0x48,
	0x60, 0xbd, 0x89, 0x05, 0x0a, 0xf4, 0x7d, 0x7d,
	0x85, 0xb0, 0x8e, 0x95, 0x1f, 0x22, 0x0d, 0xdf,
	0x2a, 0x78, 0xd5, 0x73, 0x1c, 0xb4, 0x40, 0xa1,
	0xb9, 0x3c, 0xca, 0x58, 0x92, 0x34, 0xcc, 0xce,
	0xd7, 0x42, 0x90, 0x0f, 0x8b, 0x7f, 0x32, 0xed,
	0x5c, 0x9d, 0xc8, 0x99, 0xad, 0x4e, 0x56, 0xa6,
	0xf7, 0x68, 0xb7, 0x25, 0x82, 0x37, 0x3a, 0x51,
	0x69, 0x26, 0x38, 0x52, 0x9e, 0x9a, 0x4f, 0xa7,
	0x43, 0x10, 0x80, 0xee, 0x3d, 0x59, 0x35, 0xcf,
	0x79, 0x74, 0xb5, 0xa2, 0xb1, 0x96, 0x23, 0xe0,
	0xbe, 0x05, 0xf5, 0x6e, 0x19, 0xc5, 0x66, 0x49,
	0xf0, 0xd1, 0x54, 0xa9, 0x70, 0x4b, 0xa4, 0xe2,
	0xe6, 0xe5, 0xab, 0xe4, 0xd2, 0xaa, 0x4c, 0xe3,
	0x06, 0x6f, 0xc6, 0x4a, 0xa4, 0x75, 0x97, 0xe1,
	/* NOTE not sure where the last 8 items are in this array */
}
