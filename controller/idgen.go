package main

import (
	"crypto/rand"
	"fmt"
)

// newJobID generates a short random hex ID. Avoids pulling in an external
// uuid package — the sandbox has no module-proxy network access, and a
// random 8-byte hex string is more than unique enough for a course-project
// testbed of tens of nodes.
func newJobID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("job-%x", b)
}
