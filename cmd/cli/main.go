package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/mmlt/systemd-operator/internal/operator"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	// Version as set during build.
	Version string

	sshUser = flag.String("ssh-user", "",
		`SSH user name`)

	sshPass = flag.String("ssh-pass", "",
		`SSH user password or pass-phrase when shh-file is set`)
	sshFile = flag.String("ssh-file", "",
		`File containing the user SSH key`)

	operatorId = flag.String("id", "nto",
		`String to identify service and timer entries created. Check README before changing!`)

	host = flag.String("host", "",
		`The remote host IP address.`)

	statePath = flag.String("state", "",
		`The yaml file containing desired state.`)
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	flag.Parse() // glog needs flag otherwise it will Prefix 'ERROR: logging before flag.Parse:' to each message.

	s := fmt.Sprintf("CLI %s", Version)
	pflag.VisitAll(func(flag *pflag.Flag) {
		s = fmt.Sprintf("%s %s=%q", s, flag.Name, flag.Value)
	})
	glog.Info(s)

	// /usr/lib64/systemd/system/ is read-only
	op := operator.New(*sshUser,
		readFileOrReturnArg(*sshPass),
		readFileOrReturnArg(*sshFile),
		*operatorId,
		"/etc/systemd/system/")

	// Read yaml
	configYaml, err := ioutil.ReadFile(*statePath)
	if err != nil {
		glog.Exit(err)
	}

	// Read desired state file.
	desiredState := make(map[string]string)
	err = yaml.Unmarshal(configYaml, desiredState)
	if err != nil {
		glog.Exit("parsing yaml config", *statePath, err)
	}

	// Reconcile
	err = op.Reconcile(*host, desiredState)
	if err != nil {
		glog.Error(err)
	}

	glog.Info("CLI completed")
}

// ReadFileOrReturnArg tries to read a file and return its content.
// If that fails the argument is returned.
func readFileOrReturnArg(pathOrPassword string) string {
	b, err := ioutil.ReadFile(pathOrPassword)
	if err != nil {
		return pathOrPassword
	}

	return string(b)
}
