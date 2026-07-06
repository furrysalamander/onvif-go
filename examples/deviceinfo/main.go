package main

import (
	"fmt"
	"os"

	"github.com/furrysalamander/onvif-go/onvif/devicemgmt"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "usage: deviceinfo <endpoint> <username> <password>\n")
		os.Exit(2)
	}
	endpoint := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]

	c := devicemgmt.NewClient(endpoint, username, password)

	info, err := c.GetDeviceInformation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetDeviceInformation: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Device Information:")
	fmt.Printf("  Manufacturer:    %s\n", info.Manufacturer)
	fmt.Printf("  Model:           %s\n", info.Model)
	fmt.Printf("  FirmwareVersion: %s\n", info.FirmwareVersion)
	fmt.Printf("  SerialNumber:    %s\n", info.SerialNumber)
	fmt.Printf("  HardwareId:      %s\n", info.HardwareId)

	svcs, err := c.GetServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetServices: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\nServices:")
	for _, svc := range svcs.Service {
		fmt.Printf("  %s", svc.Namespace)
		if svc.XAddr != "" {
			fmt.Printf(" -> %s", svc.XAddr)
		}
		fmt.Println()
	}
}
