package operator

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"github.com/golang/glog"
	"github.com/mmlt/sshclient"
	"github.com/mmlt/systemd-operator/internal/kclient"
	"github.com/mmlt/systemd-operator/internal/systemctl"
	"path"
	"sort"
	"strings"
)

// TODO add doc
// assume basename.service and basename.timer match
// assume files are in /usr/lib64/systemd/system
// (symlinked from /usr/lib64/systemd/system/multi-user.target.wants/)

type operator struct {
	sshUser, sshPass, sshKey string
	prefix                   string
	systemDir                string
}

// Actions to reconcile state.
//go:generate stringer -type=action
type action int
const (
	nop action = iota
	create
	update
	delete
)

// New returns an operator instance.
func New(sshUser, sshPass, sshKey, operatorId string, systemDir string) *operator {
	return &operator{
		sshUser:   sshUser,
		sshPass:   sshPass,
		sshKey:    sshKey,
		prefix:    operatorId+"-",
		systemDir: systemDir,
	}
}

// Update
func (op *operator) Update(instr *kclient.Instruction) {
	glog.V(2).Info(instr.String())
	err := op.Reconcile(instr.DesiredState.Address, instr.DesiredState.Units)
	if err != nil {
		glog.Errorf("reconcile %s: %v", instr.DesiredState.Address, err)
	}
}

// Reconcile
func (op *operator) Reconcile(ip string, desiredState map[string]string) error {
	var cl *sshclient.SshClient
	var err error
	if op.sshKey != "" {
		cl, err = sshclient.DailSSHWithKey(ip, op.sshUser, op.sshPass, []byte(op.sshKey))
	} else {
		cl, err = sshclient.DailWithPassword(ip, op.sshUser, op.sshPass)
	}
	if err != nil {
		return fmt.Errorf("dail %s@%s: %v\n", op.sshUser, ip)
	}
	defer cl.Close()

	// Fetch
	// Convert desiredState to cm[prefixed-name]content map
	cm := make(map[string]string, len(desiredState))
	for k, v := range desiredState {
		cm[op.prefix+k] = v
	}
	// Get hashes of local and remote content.
	localHash := getSha1OfMap(cm)
	remoteHash, err := getSha1OfFiles(cl, path.Join(op.systemDir, op.prefix+"*"))
	if err != nil {
		return fmt.Errorf("get sha1: %v", err)
	}

	// Decode
	timer, service := calculateActions(localHash, remoteHash)

	glog.V(2).Info("timer reconcile;", sprintActions(timer))
	glog.V(2).Info("service reconcile;", sprintActions(service))

	// Execute
	cl.SkipOnErr(true)
	op.apply(cl, cm, timer, service)
	err = cl.Err()
	if err != nil {
		return fmt.Errorf("apply: %v", err)
	}

	return nil
}

func (op *operator) apply(cl *sshclient.SshClient, cm map[string]string, timer map[string]action, service map[string]action) {
	for n, a := range timer {
		switch a {
		case create:
			createTimer(cl, op.systemDir, n, cm[n+".timer"], cm[n+".service"])
		case update:
			updateTimer(cl, op.systemDir, n, cm[n+".timer"], cm[n+".service"])
		case delete:
			deleteTimer(cl, op.systemDir, n)
		}
	}
	for n, a := range service {
		switch a {
		case create:
			createService(cl, op.systemDir, n, cm[n+".service"])
		case update:
			updateService(cl, op.systemDir, n, cm[n+".service"])
		case delete:
			deleteService(cl, op.systemDir, n)
		}
	}
}

// CalculateActions determines what actions to perform on timers and services to reconcile local with remote state.
// It returns a timer and service map with key=name of timer/service and value is the action to perform.
func calculateActions(localHash map[string]string, remoteHash map[string]string) (timer map[string]action, service map[string]action) {
	timer = make(map[string]action)
	service = make(map[string]action)

	// what to create or update?
	for k, _ := range localHash {
		if !strings.HasSuffix(k, ".service") {
			continue
		}
		// check if .service has corresponding .timer
		n := strings.TrimSuffix(k, ".service")
		rtHash, hasRT := remoteHash[n+".timer"]
		rsHash, hasRS := remoteHash[n+".service"]
		ltHash, hasLT := localHash[n+".timer"]
		lsHash, _ := localHash[n+".service"]

		if hasLT {
			// it's a timer
			if !(hasRT && hasRS) {
				// remote .timer or .service file(s) are missing
				timer[n] = create
			} else if ltHash != rtHash || lsHash != rsHash {
				// local and remote .timer or .service files are different
				timer[n] = update
			}
		} else {
			// it's a service
			if !hasRS {
				// remote .service file is missing
				service[n] = create
			} else if lsHash != rsHash {
				// local and remote .service files are different
				service[n] = update
			}
		}
	}

	// what to delete?
	for k, _ := range remoteHash {
		if !strings.HasSuffix(k, ".service") {
			continue
		}
		// check if .service has corresponding .timer
		n := strings.TrimSuffix(k, ".service")
		_, hasLT := localHash[n+".timer"]
		_, hasLS := localHash[n+".service"]

		if _, ok := remoteHash[n+".timer"]; ok {
			// it's a timer
			if !(hasLT && hasLS) {
				// local .timer or .service file(s) are missing
				timer[n] = delete
			}
		} else {
			// it's a service
			if !hasLS {
				timer[n] = delete
			}
		}
	}

	return timer, service
}

func createTimer(cl *sshclient.SshClient, dir, name, timerContent, serviceContent string) {
	copyFile(cl, dir, name+".timer", []byte(timerContent))
	copyFile(cl, dir, name+".service", []byte(serviceContent))
	sc := systemctl.New(cl)
	sc.Unit(systemctl.Start, name+".timer")
}

func updateTimer(cl *sshclient.SshClient, dir, name, timerContent, serviceContent string) {
	copyFile(cl, dir, name+".timer", []byte(timerContent))
	copyFile(cl, dir, name+".service", []byte(serviceContent))
	sc := systemctl.New(cl)
	sc.DaemonReload()
}

func deleteTimer(cl *sshclient.SshClient, dir, name string) {
	sc := systemctl.New(cl)
	sc.Unit(systemctl.Stop, name+".timer")
	deleteFile(cl, dir, name+".timer")
	deleteFile(cl, dir, name+".service")
}

func createService(cl *sshclient.SshClient, dir, name, serviceContent string) {
	copyFile(cl, dir, name+".service", []byte(serviceContent))
	sc := systemctl.New(cl)
	sc.DaemonReload()
	sc.UnitFile(systemctl.Enable, name+".service")
	sc.Unit(systemctl.Start, name+".service")
}

func updateService(cl *sshclient.SshClient, dir, name, serviceContent string) {
	copyFile(cl, dir, name+".service", []byte(serviceContent))
	sc := systemctl.New(cl)
	sc.DaemonReload()
}

func deleteService(cl *sshclient.SshClient, dir, name string) {
	sc := systemctl.New(cl)
	sc.Unit(systemctl.Stop, name+".service")
	sc.UnitFile(systemctl.Disable, name+".service")
	deleteFile(cl, dir, name+".service")
}

// CopyFile copies data to a file (664 root root name) on a remote host.
// Assume non root user in sudo group.
func copyFile(cl *sshclient.SshClient, dir string, name string, data []byte) {
	fn := "/var/tmp/" + name // temporary file location
	cl.ScpTo(data, fn, 0644)
	cl.Exec("sudo", "chown", "root:root", fn)
	cl.Exec("sudo", "mv", fn, dir)
}

// DeleteFile from a remote host.
// Assume non root user in sudo group.
func deleteFile(cl *sshclient.SshClient, dir string, name string) {
	cl.Exec("sudo", "rm", path.Join(dir, name))
}

// GetSha1OfFiles returns a map with key=name of file and value=sha1 of file.
func getSha1OfFiles(cl *sshclient.SshClient, pattern string) (map[string]string, error) {
	result := make(map[string]string)

	s, err := cl.Exec("sha1sum", pattern)
	if err != nil {
		if strings.HasSuffix(err.Error(), "Process exited with status 1") {
			// error indicates no matching files
			return nil, nil
		}
		return nil, err
	}
	r := strings.NewReader(s)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 {
			result[path.Base(f[1])] = f[0]
		}
	}

	return result, nil
}


//*** Helpers *****************************************************************

// GetSha1OfMap return a map of [name]sha1 strings.
func getSha1OfMap(content map[string]string) map[string]string {
	result := make(map[string]string, len(content))

	for k,v := range content {
		h := sha1.New()
		h.Write([]byte(v))
		result[k] = fmt.Sprintf("%x", h.Sum(nil))
	}

	return result
}

// SprintActions returns a string of service:action pairs.
func sprintActions(unit map[string]action) string {
	if len(unit) == 0 {
		return " no actions"
	}

	var ks []string
	for k, _ := range unit {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	var r string
	for _, n := range ks {
		r = fmt.Sprintf("%s %s:%d", r, n, unit[n]) //TODO %s unit[n].String()
	}
	return r
}