// Code generated by "stringer -type VegType ./src"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PEPPER-0]
	_ = x[LETTUCE-1]
	_ = x[CARROT-2]
	_ = x[CABBAGE-3]
	_ = x[ONION-4]
	_ = x[TOMATO-5]
}

const _VegType_name = "PEPPERLETTUCECARROTCABBAGEONIONTOMATO"

var _VegType_index = [...]uint8{0, 6, 13, 19, 26, 31, 37}

func (i VegType) String() string {
	if i < 0 || i >= VegType(len(_VegType_index)-1) {
		return "VegType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _VegType_name[_VegType_index[i]:_VegType_index[i+1]]
}
