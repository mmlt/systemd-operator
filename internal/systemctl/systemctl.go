package systemctl

import (
	"github.com/mmlt/systemd-operator/internal/tableconv"
	"strings"
)

// Executer interface is used to perform systemctl commands.
type Executer interface {
	Exec(cmd string, args ...string) (string, error)
}

// SystemCtl instance data.
type SystemCtl struct {
	cmd Executer
}

// New systemctl instance.
func New(cmd Executer) *SystemCtl {
    return &SystemCtl{
		cmd: cmd,
    }
}

// Unit data as returned by systemctl list-units.
type Unit struct{
	Name string
	Load string
	Active string
	Sub string
	Description string
}

// ListUnits returns status of all systemd units that match a pattern.
func (sc *SystemCtl) ListUnits(pattern string) ([]Unit, error) {
	s, err := sc.systemctl("list-units", pattern, "--plain", "--full")
	if err != nil {
		return nil, err
	}

	var res []Unit

	if strings.HasPrefix(s, "0 loaded units listed") {
		return res, nil
	}

	tab := tableconv.Scan(strings.NewReader(s), 5)
	c := tab.ColNamesToIndices()
	for _, r := range tab.Rows {
		u := Unit{
			Name: r[c["UNIT"]],
			Load: r[c["LOAD"]],
			Active: r[c["ACTIVE"]],
			Sub: r[c["SUB"]],
			Description: r[c["DESCRIPTION"]],
		}
		res = append(res, u)
	}

	return res, nil
}

// Timer data as returned by systemctl list-timers.
type Timer struct{
	Next string
	Left string
	Last string
	Passed string
	Name string
	Activates string
}

// ListTimers returns status of all systemd timers that match a pattern.
func (sc *SystemCtl) ListTimers(pattern string) ([]Timer, error) {
	s, err := sc.systemctl("list-timers", pattern, "--plain", "--full")
	if err != nil {
		return nil, err
	}

	var res []Timer

	if strings.HasPrefix(s, "0 timers listed") {
		return res, nil
	}

	tab := tableconv.Scan(strings.NewReader(s), 5)
	c := tab.ColNamesToIndices()
	for _, r := range tab.Rows {
		t := Timer{
			Next: r[c["NEXT"]],
			Left: r[c["LEFT"]],
			Last: r[c["LAST"]],
			Passed: r[c["PASSED"]],
			Name: r[c["UNIT"]],
			Activates: r[c["ACTIVATES"]],
		}
		res = append(res, t)
	}

	return res, nil
}

func (sc *SystemCtl) DaemonReload() (string, error) {
	return sc.sudoSystemctl("daemon-reload")
}

// UnitCmd represents the actions to perform on an unit.
//go:generate stringer -type=UnitCmd
type UnitCmd int

const (
	Start UnitCmd = iota
	Stop
	Reload
	Restart
)

// Unit performs one of:
// Start (activate) one or more units
// Stop (deactivate) one or more units
// Reload one or more units
// Start or restart one or more units
func (sc *SystemCtl) Unit(cmd UnitCmd, name string) (string, error) {
	return sc.sudoSystemctl(strings.ToLower(cmd.String()), name)
}

// UnitCmd represents the actions to perform on an unit.
//go:generate stringer -type=UnitFileCmd
type UnitFileCmd int

const (
	Enable UnitFileCmd = iota
	Disable
	Reenable
)

// UnitFile performs one of:
// Enable one or more unit files
// Disable one or more unit files
// Reenable one or more unit files
func (sc *SystemCtl) UnitFile(cmd UnitFileCmd, name string) (string, error) {
	return sc.sudoSystemctl(strings.ToLower(cmd.String()), name)
}

func (sc *SystemCtl) sudoSystemctl(arg ...string) (string, error) {
	a := append([]string{"systemctl"}, arg...)
	return sc.cmd.Exec("sudo", a...)
}

func (sc *SystemCtl) systemctl(arg ...string) (string, error) {
	return sc.cmd.Exec("systemctl", arg...)
}


