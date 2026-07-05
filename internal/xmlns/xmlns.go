// Package xmlns is the namespace URI registry for ONVIF and its dependencies.
// Placeholder during M0.
package xmlns

// Namespace constants used across the ONVIF stack and the support schemas it
// builds on (WS-Security, WS-Addressing, WS-Notification, WSRF, SOAP).
const (
	ONVIFSchema     = "http://www.onvif.org/ver10/schema"
	DeviceMgmt      = "http://www.onvif.org/ver10/device/wsdl"
	Media           = "http://www.onvif.org/ver10/media/wsdl"
	Media2          = "http://www.onvif.org/ver20/media/wsdl"
	PTZ             = "http://www.onvif.org/ver20/ptz/wsdl"
	Events          = "http://www.onvif.org/ver10/events/wsdl"
	Imaging         = "http://www.onvif.org/ver20/imaging/wsdl"
	WSAddressing    = "http://www.w3.org/2005/08/addressing"
	WSSecurity      = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	WSUtility       = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
	WSNotification  = "http://docs.oasis-open.org/wsn/b-2"
	WSNotificationT = "http://docs.oasis-open.org/wsn/t-1"
	WSRF            = "http://docs.oasis-open.org/wsrf/rw-2"
	SOAPEnvelope    = "http://www.w3.org/2003/05/soap-envelope"
)
