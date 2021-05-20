package inspect

import (
	"github.com/bytepowered/flux"
	"os"
	"strings"
)

// EnvQueryHandler 查询环境变量
func EnvQueryHandler(webex flux.ServerWebContext) error {
	osenv := os.Environ()
	envs := make(map[string]string, len(osenv))
	for _, e := range osenv {
		pair := strings.SplitN(e, "=", 2)
		envs[pair[0]] = pair[1]
	}
	// Resolve key
	key := webex.FormVar("key")
	if "" == key {
		return WriteResponse(webex, flux.StatusOK, os.Getenv(key))
	} else {
		return WriteResponse(webex, flux.StatusOK, envs)
	}

}
