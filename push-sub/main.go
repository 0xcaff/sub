package main

import (
	"bytes"
	"flag"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/0xcaff/sub"
	log "github.com/sirupsen/logrus"
)

var (
	confPath = flag.String(
		"config",
		"",
		"The location of the configuration file.",
	)

	verbosity = flag.Int(
		"verbosity",
		4, // InfoLevel
		"The log level. Higher is more detailed.",
	)
)

func main() {
	flag.Parse()

	subscriptions := map[string]*sub.Sub{}

	// set log level
	log.SetLevel(log.Level(*verbosity))

	// default config path
	if len(*confPath) == 0 {
		*confPath = filepath.Join(ConfigBaseDir, "push-sub", "config.toml")
	}

	// open config file
	conf, err := GetConfig(*confPath)
	if err != nil {
		log.Error(err)
		return
	}

	// The muxer which will be delegating all requests to the callback server.
	mux := &http.ServeMux{}

	// register all listeners
	for name, subscription := range conf.Subscriptions {
		fields := log.Fields{"name": name}

		s := sub.New()
		s.Topic = &subscription.Topic.URL

		// discover hub if needed
		if subscription.Hub == nil {
			log.WithFields(fields).Info("discovering hub")

			err = s.Discover()
			if err != nil {
				log.WithFields(fields).Error(err)
				return
			}

			// ensure secure hub
			if !subscription.AllowInsecure {
				s.Hub.Scheme = "https"
			}

			log.WithFields(fields).Info("discovered hub: ", s.Hub)
		} else {
			s.Hub = &subscription.Hub.URL
		}

		// register at random endpoint. We are using a random, hard to guess
		// endpoint for security.
		endpoint := "/" + string(sub.RandAsciiBytes(99))

		// record callback url
		s.Callback = conf.BasePath.ResolveReference(
			MustParseUrl(endpoint))

		if log.GetLevel() >= log.InfoLevel {
			fields["endpoint"] = s.Callback.String()
		}

		// setup listeners
		s.OnMessage = func(_ *http.Request, body []byte) {
			fields := log.Fields{"name": name}
			log.WithFields(fields).Info("received message")
			log.WithFields(fields).Debug("message: ", string(body))

			// create stdin
			stdin := bytes.NewReader(body)

			// create cmd
			cmd := exec.Command(subscription.Command[0], subscription.Command[1:]...)
			cmd.Stdin = stdin

			// launch cmd
			go func() {
				log.WithFields(log.Fields{
					"name": name,
					"args": subscription.Command,
				}).Info("running")
				log.WithFields(fields).Info("running")
				output, err := cmd.CombinedOutput()
				log.Debug("Output: ", string(output))
				if err != nil {
					log.WithFields(fields).Error(err)
				}
			}()
		}

		s.OnError = func(err error) {
			// we only warn because these are non-critical errors
			log.WithFields(fields).Warning(err)
		}

		s.OnRenewLease = func(s *sub.Sub) {
			log.WithFields(fields).Info("renewing")

			// only try to renew once
			err := s.Subscribe()
			if err != nil {
				log.Error(err)
			}
		}

		// register the callback
		log.WithFields(fields).Info("registered")
		mux.Handle(endpoint, s)

		subscriptions[name] = s
	}

	server := &http.Server{
		Addr:    conf.Address,
		Handler: mux,
	}

	// start listening
	log.Info("listening on " + server.Addr)
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Error(err)
		return
	}

	// unsub on shutdown
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt)
	isShuttingDown := false

	go func() {
		<-shutdownSignal
		isShuttingDown = true

		log.Info("shutting down")
		var wg sync.WaitGroup
		for name, subscription := range subscriptions {
			wg.Add(1)
			go func() {
				fields := log.Fields{"name": name}
				log.WithFields(fields).Info("unsubscribing")

				err := subscription.Unsubscribe()
				if err != nil {
					log.WithFields(fields).Error(err)
				}

				log.WithFields(fields).Info("unsubscribed")
				wg.Done()
			}()
		}

		wg.Wait()
		os.Exit(0)
	}()

	go func() {
		// try subscribing to all subscriptions
		for name, subscription := range subscriptions {
			if isShuttingDown {
				break
			}

			fields := log.Fields{"name": name}

			err := subscription.Subscribe()
			log.WithFields(fields).Info("subscribing")

			if err != nil {
				log.WithFields(fields).Error(err)

				// turn off server
				server.Shutdown(nil)
			} else {
				log.WithFields(fields).Info("subscribed")
			}
		}
	}()

	err = server.Serve(KeepAliveListener{
		KeepAlive:   3 * time.Minute,
		TCPListener: listener.(*net.TCPListener),
	})
	if err != nil {
		log.Error(err)
		return
	}
}
