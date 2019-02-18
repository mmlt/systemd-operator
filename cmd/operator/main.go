package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/mmlt/systemd-operator/internal/kclient"
	"github.com/mmlt/systemd-operator/internal/operator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"io/ioutil"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"sync"
	"time"
)

var (
	// Version as set during build.
	Version string

	k8sApi = flag.String("k8s-api", "",
		`URL of Kubernetes API server or "" when running in-cluster`)

	sshUser = flag.String("ssh-user", "",
		`SSH user name`)

	sshPass = flag.String("ssh-pass", "",
		`SSH user password or pass-phrase when shh-file is set`)
	sshFile = flag.String("ssh-file", "",
		`File containing the user SSH key`)

	operatorId = flag.String("id", "nto",
		`String to identify service and timer entries created by this operator. Check README before changing!`)

	promAddrs = flag.String("prom-addrs", ":9102",
		`The Prometheus endpoint address.`)
)

func init() {
	// Create Prometheus counters for number of glog'd info, warning and error lines.
	logged_errors := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Subsystem: "unit_operator", //TODO make const (also in kclient)
			Name:      "logged_errors",
			Help:      "Number of logged errors.",
		},
		func() float64 {
			return float64(glog.Stats.Error.Lines())
		})

	prometheus.MustRegister(logged_errors)

	logged_warnings := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Subsystem: "unit_operator",
			Name:      "logged_warnings",
			Help:      "Number of logged warnings.",
		},
		func() float64 {
			return float64(glog.Stats.Warning.Lines())
		})

	prometheus.MustRegister(logged_warnings)

	logged_info := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Subsystem: "unit_operator",
			Name:      "logged_info",
			Help:      "Number of logged info.",
		},
		func() float64 {
			return float64(glog.Stats.Info.Lines())
		})

	prometheus.MustRegister(logged_info)
}

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	flag.Parse() // glog needs flag otherwise it will Prefix 'ERROR: logging before flag.Parse:' to each message.

	s := fmt.Sprintf("Start unit_operator %s", Version) //TODO global rename node-timer-operator to unit_operator (make it const)
	pflag.VisitAll(func(flag *pflag.Flag) {
		s = fmt.Sprintf("%s %s=%q", s, flag.Name, flag.Value)
	})
	glog.Info(s)

	// Validate cli flags
	if *operatorId == "" {
		glog.Fatal("operator-id invalid: ", *operatorId)
	}

	// Start components
	config, err := clientcmd.BuildConfigFromFlags(*k8sApi, "")
	if err != nil {
		glog.Fatal("k8s config err: ", err)
	}

	kubeClient := kubernetes.NewForConfigOrDie(config)
	sharedInformers := informers.NewSharedInformerFactory(kubeClient, 15*time.Minute)

	// Create client that talks to the API server.
	c := kclient.New(kubeClient, sharedInformers, *operatorId)

	// Create backend to modify systemd units.
	// TODO op := operator.New(*sshUser, *sshPass, *sshFile, *operatorId, "/etc/systemd/system/")
	//// /usr/lib64/systemd/system/ is read-only
	op := operator.New(*sshUser,
		readFileOrReturnArg(*sshPass),
		readFileOrReturnArg(*sshFile),
		*operatorId,
		"/etc/systemd/system/")

	// Wire the components.
	c.OnChange(op.Update) //TODO rename to c.OnInstruction(b.Execute)

	// Start the instances.
	stop := make(chan struct{})

	sharedInformers.Start(stop)
	wg := &sync.WaitGroup{} // GO routines should add themselves
	go c.Run(stop, wg)

	// Start prometheus endpoint
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(*promAddrs, nil)
	if err != http.ErrServerClosed {
		glog.Error(err)
	}

	glog.Info("Shutting down.")
	close(stop)
	wg.Wait()
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
