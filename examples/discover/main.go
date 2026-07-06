package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/furrysalamander/onvif-go/onvif"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devices, err := onvif.Discover(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover: %v\n", err)
		return 1
	}

	if len(devices) == 0 {
		fmt.Println("No ONVIF devices found")
		return 0
	}

	fmt.Printf("Found %d device(s):\n\n", len(devices))
	for i, d := range devices {
		fmt.Printf("Device %d:\n", i+1)
		fmt.Printf("  Endpoint:   %s\n", d.Info.Endpoint)
		fmt.Printf("  Address:    %s\n", d.Info.Address)
		fmt.Printf("  Service URL: %s\n", d.XAddr)
		fmt.Printf("  Types:      %v\n", d.Info.Types)
		fmt.Println()
	}
	return 0
}
