package main

import (
	"fmt"
	"log"
)

func main() {
	conf, err := readConfig()
	if err != nil {
		conf.Usage()
		log.Fatal(err)
	}

	if conf.Addr == "" {
		fmt.Println("Error: No address specified")
		conf.Usage()
		return
	}

	if conf.V1 == true && conf.V2 == true {
		log.Fatal("Cannot specify both v1 and v2")
		return
	}
	if conf.V1 == false && conf.V2 == false {
		conf.V1 = true
	}

	log.Println("SSC Ping")

	if conf.V1 {
		resp, err := PingV1(conf.Addr, conf.Port, conf.Debug)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(resp)
	} else if conf.V2 {
		resp, err := PingV2(conf.Addr, conf.Port, conf.Debug, PingGlobalSummary|PingArenaSummary)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(resp)
	}
}
