package inspect

import (
	"github.com/bytepowered/flux"
)

func ConfigurationQueryHandler(webex flux.ServerWebContext) error {
	root := flux.NewRootConfiguration()
	// Namespaces
	ns := webex.FormVar("namespace")
	switch ns {
	case "all", "":
		// root = root
	default:
		root = flux.NewConfiguration(ns)
	}
	// Resolve key
	key := webex.FormVar("key")
	if "" == key {
		return send(webex, flux.StatusOK, root.ToStringMap())
	} else {
		return send(webex, flux.StatusOK, map[string]interface{}{
			"namespace": ns,
			"key":       key,
			"value":     root.Get(key),
		})
	}
}
