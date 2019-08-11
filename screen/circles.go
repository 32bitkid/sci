package screen

var circleBitmaps = map[int][][]bool{
	0: {{true}},
	1: {
		{false, true, false},
		{true, true, true},
		{false, true, false},
	},
	2: {
		{false, true, true, true, false},
		{true, true, true, true, true},
		{true, true, true, true, true},
		{true, true, true, true, true},
		{false, true, true, true, false},
	},
	3: {
		{false, false, true, true, true, false, false},
		{false, true, true, true, true, true, false},
		{true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true},
		{false, true, true, true, true, true, false},
		{false, false, true, true, true, false, false},
	},
	4: {
		{false, false, false, true, true, true, false, false, false},
		{false, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true},
		{false, true, true, true, true, true, true, true, false},
		{false, false, false, true, true, true, false},
	},
	5: {
		{false, false, false, false, true, true, true, false, false, false, false},
		{false, false, true, true, true, true, true, true, true, false, false},
		{false, true, true, true, true, true, true, true, true, true, false},
		{false, true, true, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true},
		{false, true, true, true, true, true, true, true, true, true, false},
		{false, true, true, true, true, true, true, true, true, true, false},
		{false, false, true, true, true, true, true, true, true, false, false},
		{false, false, false, true, true, true, true, true, false, false, false},
	},
	6: {
		{false, false, false, false, true, true, true, true, true, false, false, false, false},
		{false, false, true, true, true, true, true, true, true, true, true, false, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true},
		{false, true, true, true, true, true, true, true, true, true, true, true, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, false},
		{false, false, true, true, true, true, true, true, true, true, true, false, false},
		{false, false, false, false, true, true, true, true, true, false, false, false, false},
	},
	7: {
		{false, false, false, false, false, true, true, true, true, true, false, false, false, false, false},
		{false, false, false, true, true, true, true, true, true, true, true, true, false, false, false},
		{false, false, true, true, true, true, true, true, true, true, true, true, true, false, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{false, false, true, true, true, true, true, true, true, true, true, true, true, false, false},
		{false, false, false, true, true, true, true, true, true, true, true, true, false, false, false},
		{false, false, false, false, false, true, true, true, true, true, false, false, false, false, false},
	},
}