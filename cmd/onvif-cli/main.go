package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/furrysalamander/onvif-go/internal/wsdiscovery"
	"github.com/furrysalamander/onvif-go/onvif/devicemgmt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: onvif-cli <discover|info> [args...]\n")
		fmt.Fprintf(os.Stderr, "  discover [types]        discover ONVIF devices\n")
		fmt.Fprintf(os.Stderr, "  info <endpoint> <user> <pass>  get device info\n")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "discover":
		types := ""
		if len(os.Args) > 2 {
			types = os.Args[2]
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		devices, err := discover(ctx, types)
		if err != nil {
			fmt.Fprintf(os.Stderr, "discover: %v\n", err)
			os.Exit(1)
		}
		for _, d := range devices {
			fmt.Printf("Endpoint: %s\n", d.Endpoint)
			fmt.Printf("  Address: %s\n", d.Address)
			fmt.Printf("  Types: %v\n", d.Types)
			fmt.Printf("  XAddrs: %v\n", d.XAddrs)
			fmt.Println()
		}

	case "info":
		if len(os.Args) < 5 {
			fmt.Fprintf(os.Stderr, "usage: onvif-cli info <endpoint> <username> <password>\n")
			os.Exit(2)
		}
		endpoint := os.Args[2]
		username := os.Args[3]
		password := os.Args[4]

		c := devicemgmt.NewClient(endpoint, username, password)
		info, err := c.GetDeviceInformation()
		if err != nil {
			fmt.Fprintf(os.Stderr, "GetDeviceInformation: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Manufacturer:    %s\n", info.Manufacturer)
		fmt.Printf("Model:           %s\n", info.Model)
		fmt.Printf("FirmwareVersion: %s\n", info.FirmwareVersion)
		fmt.Printf("SerialNumber:    %s\n", info.SerialNumber)
		fmt.Printf("HardwareId:      %s\n", info.HardwareId)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(2)
	}
}

func discover(ctx context.Context, types string) ([]wsdiscovery.DeviceInfo, error) {
	if types != "" {
		return wsdiscovery.DiscoverWithTypes(ctx, types)
	}
	return wsdiscovery.Discover(ctx)
}
