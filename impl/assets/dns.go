package assets

import (
	_ "embed"
	"path/filepath"
)

//go:embed dnsshim.ts
var Shim string

//go:embed dnsprobe.ts
var Probe string

const DNSDirName = ".kcpdns"

var DNSProbeEnv = []string{"KCP_SERVICE_DOMAIN", "KCP_DNS_TABLE", "KCP_TOKENS", "KCP_SERVER"}

func DNSSet(runsDir string) *Set {
	return &Set{
		Dir: filepath.Join(runsDir, DNSDirName),
		Files: map[string][]byte{
			ShimName:  []byte(Shim),
			ProbeName: []byte(Probe),
		},
	}
}

const (
	ShimName = "shim.ts"

	ProbeName = "probe.ts"
)
