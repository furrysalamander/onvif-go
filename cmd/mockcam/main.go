package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/furrysalamander/onvif-go/internal/mockcam"
)

func main() {
	mc := mockcam.New()
	fmt.Fprintf(os.Stderr, "mockcam: starting ONVIF mock camera server\n")

	go func() {
		if err := mc.Listen(); err != nil {
			log.Fatalf("mockcam: %v", err)
		}
	}()

	fmt.Printf("Device service: %s\n", mc.Addr)
	fmt.Printf("Device info: %s %s\n", mc.DeviceInfo.Manufacturer, mc.DeviceInfo.Model)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Fprintf(os.Stderr, "\nmockcam: shutting down\n")
}
