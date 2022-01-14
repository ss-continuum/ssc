package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pkg/errors"
	"github.com/ss-continuum/ssc/pkg/connection/directory"
)

const directoryServerPort = 4990

func main() {
	fs := flag.NewFlagSet("ssc-directory", flag.ExitOnError)

	var Port int
	var Debug bool

	fs.IntVar(&Port, "port", directoryServerPort, "server port")
	fs.BoolVar(&Debug, "debug", false, "log network packets")

	root := &ffcli.Command{
		ShortUsage: fmt.Sprintf("%s [-debug] [-port <portnumber>] address", os.Args[0]),
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.Errorf("Unexpected number of args. Expected: 1, got: %d", len(args))
			}

			addr := fmt.Sprintf("%s:%d", args[0], Port)

			log.Printf("Requesting directory at %s\n", addr)
			//list, err := requestDirectoryList(addr, Debug)
			dirConn, err := directory.Dial(addr)
			if err != nil {
				return errors.Wrap(err, "Dial")
			}
			defer dirConn.Close()

			dirConn.Debug = Debug
			if err := dirConn.Login(0); err != nil {
				return errors.Wrap(err, "login")
			}
			list, err := dirConn.Directory(0)
			if err != nil {
				return errors.Wrap(err, "error requesting list")
			}

			for _, entry := range list.Entries {
				fmt.Println("---")
				fmt.Println(entry)
			}
			fmt.Println("---")

			return nil
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
