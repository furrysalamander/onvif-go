package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/furrysalamander/onvif-go/onvif"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: onvif-cli <discover|info> [args...]\n")
		fmt.Fprintf(os.Stderr, "  discover               discover ONVIF devices\n")
		fmt.Fprintf(os.Stderr, "  info <endpoint> <user> <pass>  get device info\n")
		os.Exit(2)
	}

	var code int
	switch os.Args[1] {
	case "discover":
		code = cmdDiscover()
	case "info":
		code = cmdInfo()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		code = 2
	}
	os.Exit(code)
}

func cmdDiscover() int {
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
	for _, d := range devices {
		fmt.Printf("Endpoint: %s\n", d.Info.Endpoint)
		fmt.Printf("  Address: %s\n", d.Info.Address)
		fmt.Printf("  Service URL: %s\n", d.XAddr)
		fmt.Printf("  Types: %v\n", d.Info.Types)
		fmt.Println()
	}
	return 0
}

func cmdInfo() int {
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "usage: onvif-cli info <endpoint> <username> <password>\n")
		return 2
	}
	endpoint := os.Args[2]
	username := os.Args[3]
	password := os.Args[4]

	d := onvif.NewDevice(endpoint, username, password)
	ctx := context.Background()
	info, err := d.GetInfo(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetDeviceInformation: %v\n", err)
		return 1
	}
	fmt.Printf("Manufacturer:    %s\n", info.Manufacturer)
	fmt.Printf("Model:           %s\n", info.Model)
	fmt.Printf("FirmwareVersion: %s\n", info.FirmwareVersion)
	fmt.Printf("SerialNumber:    %s\n", info.SerialNumber)
	fmt.Printf("HardwareID:      %s\n", info.HardwareID)
	return 0
}
