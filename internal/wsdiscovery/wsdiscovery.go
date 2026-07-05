// Package wsdiscovery implements WS-Discovery over UDP multicast
// (Probe / ProbeMatch / Hello / Bye) for ONVIF device discovery. Placeholder
// during M0; real implementation arrives in M4.
package wsdiscovery

// MulticastAddr and Port follow the WS-Discovery / ONVIF defaults.
const (
	MulticastAddr = "239.255.255.250"
	Port          = 3702
)
