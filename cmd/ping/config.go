package main

import (
	"flag"
	"os"

	"github.com/peterbourgon/ff/v3"
	"github.com/pkg/errors"
)

type Config struct {
	Addr  string
	Port  int
	Debug bool
	V1    bool
	V2    bool

	fs *flag.FlagSet
}

func (c Config) Usage() {
	c.fs.Usage()
}

func readConfig() (Config, error) {
	var c Config

	c.fs = flag.NewFlagSet("ssc-ping", flag.ExitOnError)

	c.fs.StringVar(&c.Addr, "addr", "", "server address")
	c.fs.IntVar(&c.Port, "port", 5001, "server port")
	c.fs.BoolVar(&c.Debug, "debug", false, "log network packets")
	c.fs.BoolVar(&c.V1, "1", false, "use ping v1 (default)")
	c.fs.BoolVar(&c.V2, "2", false, "use ping v2")

	if err := ff.Parse(c.fs, os.Args[1:]); err != nil {
		return Config{}, errors.Wrap(err, "ff.Parse")
	}

	return c, nil
}
