package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func main() {
	// parse flags
	var (
		configPath = pflag.StringP("config-path", "c", "$HOME/.fwd.yaml", "path to config file")
	)
	pflag.Parse()

	// if os.Geteuid() != 0 {
	// 	fmt.Println("fwd must run as root")
	// 	os.Exit(1)
	// }

	log := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.TextFormatter{},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	cidr, targets, err := readConfig(log, *configPath)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go watchSignal(log, cancel)

	f := &fwd{
		log:     log,
		kubectl: kubectlExec{},
		cidr:    cidr,
		targets: targets,
	}
	if err := f.run(ctx); err != nil {
		log.Fatal(err)
	}
}

func watchSignal(log logrus.FieldLogger, cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-ch
	log.Infof("caught [%s], exiting...", sig)
	cancel()
}
