package inspect

import (
	"github.com/bytepowered/flux"
	"github.com/spf13/viper"
)

func ConfigurationQueryHandler(webex flux.ServerWebContext) error {
	global := viper.GetViper().AllSettings()
	config := global
	// Namespaces
	ns := webex.FormVar("ns")
	switch ns {
	case "all", "":
		config = global
	default:
		config = flux.NewConfiguration(ns).ToStringMap()
	}
	return send(webex, flux.StatusOK, config)
}
