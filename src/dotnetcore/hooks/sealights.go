package hooks

import (
	"github.com/Sealights/libbuildpack-sealights"
	"github.com/cloudfoundry/libbuildpack"
)

func init() {
	libbuildpack.AddHook(sealights.NewHook())
}
