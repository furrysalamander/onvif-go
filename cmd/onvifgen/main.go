// onvifgen is the ONVIF WSDL/XSD → Go code generator.
//
// During M0 this binary is a deliberate no-op: it exits 0 so that
// `go generate ./...` (and the CI "generator no-drift" check) succeed while
// the real generator is built in M1/M2.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "onvifgen: M0 no-op (real generator arrives in M1)")
}
