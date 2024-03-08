package main

import (
	"flag"
	"os"
	"time"

	"github.com/robertlestak/clambucket/internal/event"
	"github.com/robertlestak/clambucket/pkg/clambucket"
	log "github.com/sirupsen/logrus"
)

var (
	clambucketFlagset = flag.NewFlagSet("clambucket", flag.ExitOnError)
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func handleEvent(e event.Event) error {
	l := log.WithFields(log.Fields{
		"fn": "handleEvent",
	})
	l.Debug("handling event")
	uri, err := e.GetUri()
	if err != nil {
		return err
	}
	if uri == "" {
		l.Debug("no URI")
		return nil
	}
	if err := clambucket.Scan(uri); err != nil {
		return err
	}
	return nil
}

func main() {
	l := log.WithFields(log.Fields{
		"fn": "main",
	})
	l.Debug("starting clambucket")
	clean := clambucketFlagset.String("clean", "", "clean bucket")
	cleanPrefix := clambucketFlagset.String("clean-prefix", "", "clean prefix")
	quarantine := clambucketFlagset.String("quarantine", "", "quarantine bucket")
	quarantinePrefix := clambucketFlagset.String("quarantine-prefix", "", "quarantine prefix")
	assumeRoleArn := clambucketFlagset.String("role", "", "assume role ARN")
	eventHandler := clambucketFlagset.String("event", "", "event handler")
	eventConfig := clambucketFlagset.String("config", "", "event configuration")
	watch := clambucketFlagset.Bool("watch", false, "watch for events")
	watchDelay := clambucketFlagset.String("watch-delay", "5s", "watch delay")
	clambucketFlagset.Parse(os.Args[1:])
	clambucket.Downstreams = &clambucket.Downstream{
		Clean:            *clean,
		CleanPrefix:      *cleanPrefix,
		Quarantine:       *quarantine,
		QuarantinePrefix: *quarantinePrefix,
	}
	if os.Getenv("CLEAN_BUCKET") != "" {
		clambucket.Downstreams.Clean = os.Getenv("CLEAN_BUCKET")
	}
	if os.Getenv("CLEAN_PREFIX") != "" {
		clambucket.Downstreams.CleanPrefix = os.Getenv("CLEAN_PREFIX")
	}
	if os.Getenv("QUARANTINE_BUCKET") != "" {
		clambucket.Downstreams.Quarantine = os.Getenv("QUARANTINE_BUCKET")
	}
	if os.Getenv("QUARANTINE_PREFIX") != "" {
		clambucket.Downstreams.QuarantinePrefix = os.Getenv("QUARANTINE_PREFIX")
	}
	l.Debug("initializing event handler")
	if *assumeRoleArn == "" {
		*assumeRoleArn = os.Getenv("ASSUME_ROLE_ARN")
	}
	if *assumeRoleArn != "" {
		event.AssumeRoleArn = *assumeRoleArn
	}
	if *eventHandler == "" {
		*eventHandler = os.Getenv("EVENT_HANDLER")
	}
	if os.Getenv("WATCH") == "true" {
		*watch = true
	}
	e := event.New(event.EventHandler(*eventHandler))
	if e == nil {
		l.Fatal("invalid event handler")
	}
	l.Debug("initializing event")
	if *eventConfig == "" {
		*eventConfig = os.Getenv("EVENT_CONFIG")
	}
	if err := e.Init(*eventConfig); err != nil {
		l.Fatal(err)
	}
	if *watch {
		l.Debug("watching for events")
		if os.Getenv("WATCH_DELAY") != "" {
			*watchDelay = os.Getenv("WATCH_DELAY")
		}
		dur, err := time.ParseDuration(*watchDelay)
		if err != nil {
			l.Fatal(err)
		}
		for {
			l.Debug("handling event")
			if err := handleEvent(e); err != nil {
				l.Error(err)
			}
			l.Debug("waiting for event")
			time.Sleep(dur)
		}
	}
	l.Debug("scanning URI")
	if err := handleEvent(e); err != nil {
		l.Fatal(err)
	}
}
