package package_json

import (
	"github.com/cloudfoundry/libbuildpack"
)

type PackageJSON struct {
	Engines Engines `json:"engines"`
}

type Engines struct {
	Node string `json:"node"`
}

type logger interface {
	Info(format string, args ...interface{})
}

const (
	DefaultNodeVersion = "6"
	PackageJson        = "package.json"
)

func GetNodeFromPackageJSON(pkgJSONPath string, logger logger) (string, error) {
	var p PackageJSON

	if err := libbuildpack.NewJSON().Load(pkgJSONPath, &p); err != nil {
		return "", err
	}

	if p.Engines.Node != "" {
		logger.Info("engines.node (%s): %s", PackageJson, p.Engines.Node)
		return p.Engines.Node, nil
	} else {
		logger.Info("engines.node (%s): unspecified", PackageJson)
	}

	return DefaultNodeVersion, nil
}
