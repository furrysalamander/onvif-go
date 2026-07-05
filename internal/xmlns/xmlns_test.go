package xmlns

import "testing"

func TestNamespaces_NotEmpty(t *testing.T) {
	cases := []struct{ name, val string }{
		{"ONVIFSchema", ONVIFSchema},
		{"DeviceMgmt", DeviceMgmt},
		{"Events", Events},
		{"WSAddressing", WSAddressing},
		{"WSNotification", WSNotification},
		{"SOAPEnvelope", SOAPEnvelope},
	}
	for _, c := range cases {
		if c.val == "" {
			t.Fatalf("namespace %s is empty", c.name)
		}
	}
}
