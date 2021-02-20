package main

import (
	"log"
	prowxylow "proxy/proxylow/forwarder"
	"time"
)

func main() {
	err := prowxylow.ForwarderRun("udp", "192.3.249.197:5667", "10.0.0.11/24", "223.6.6.6")
	if err != nil {
		log.Fatal(err)
	}
	for {
		time.Sleep(time.Millisecond * 2000)
		log.Println("alive")
	}
}
