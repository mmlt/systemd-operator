// Code generated by "stringer -type=UnitCmd"; DO NOT EDIT.

package systemctl

import "strconv"

const _UnitCmd_name = "StartStopReloadRestart"

var _UnitCmd_index = [...]uint8{0, 5, 9, 15, 22}

func (i UnitCmd) String() string {
	if i < 0 || i >= UnitCmd(len(_UnitCmd_index)-1) {
		return "UnitCmd(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _UnitCmd_name[_UnitCmd_index[i]:_UnitCmd_index[i+1]]
}
