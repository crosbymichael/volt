package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/VoltFramework/volt/api"
	"github.com/VoltFramework/volt/mesoslib"
	"github.com/VoltFramework/volt/mesosproto"
	flag "github.com/dotcloud/docker/pkg/mflag"
)

var (
	port      int
	master    string
	user      string
	ip        string
	debug     bool
	cert, key string
	ca        string

	log             = logrus.New()
	frameworkName   = "volt"
	registerTimeout = 5 * time.Second
)

func init() {
	flag.IntVar(&port, []string{"p", "-port"}, 8080, "Port to listen on for the API")
	flag.StringVar(&master, []string{"m", "-master"}, "localhost:5050", "Master to connect to")
	flag.BoolVar(&debug, []string{"D", "-debug"}, false, "")
	flag.StringVar(&user, []string{"u", "-user"}, "root", "User to execute tasks as")
	flag.StringVar(&ip, []string{"-ip"}, "", "IP address to listen on [default: autodetect]")
	flag.StringVar(&ca, []string{"-ca"}, "", "TLS CA Certificate file")
	flag.StringVar(&cert, []string{"-cert"}, "", "TLS Certificate file")
	flag.StringVar(&key, []string{"-key"}, "", "TLS Key file")

	flag.Parse()
}

func waitForSignals(m *mesoslib.MesosLib) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for sig := range signals {
		log.Debugf("received signal %s unregistering framework\n", sig)

		if err := m.UnRegisterFramework(); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}
}

func getTLSConfig() (*tls.Config, error) {
	if cert == "" {
		return nil, nil
	}

	file, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(file)

	config := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            pool,
	}

	cert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	config.Certificates = []tls.Certificate{cert}

	return config, nil
}

func main() {
	var (
		err    error
		config *tls.Config

		frameworkInfo = &mesosproto.FrameworkInfo{Name: &frameworkName, User: &user}
	)

	if config, err = getTLSConfig(); err != nil {
		log.Fatal(err)
	}

	if debug {
		log.Level = logrus.DebugLevel
	}

	// initialize MesosLib
	m := mesoslib.NewMesosLib(master, log, frameworkInfo, ip)

	// try to register against the master
	if err := m.RegisterFramework(); err != nil {
		log.Fatal(err)
	}

	// wait for the registered event
	select {
	case event := <-m.GetEvent(mesosproto.Event_REGISTERED):
		log.WithFields(logrus.Fields{"FrameworkId": *event.Registered.FrameworkId.Value}).Info("Registration successful.")
	case <-time.After(registerTimeout):
		log.WithField("--ip", ip).Fatal("Registration timed out. --ip must route to this host from the mesos-master.")
	}

	go waitForSignals(m)

	// once we are registered, start the API
	if err := api.NewAPI(m).ListenAndServe(port, config); err != nil {
		log.Fatal(err)
	}
}
