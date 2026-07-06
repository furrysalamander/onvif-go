package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/furrysalamander/onvif-go/onvif/client"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devices, err := client.Discover(ctx)
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
		fmt.Printf("  Endpoint: %s\n", d.Endpoint)
		fmt.Printf("  Address:  %s\n", d.Address)
		fmt.Printf("  Types:    %v\n", d.Types)
		fmt.Printf("  XAddrs:   %v\n", d.XAddrs)
		fmt.Println()
	}
	return 0
}
