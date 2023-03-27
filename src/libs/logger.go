package libs

import "github.com/hashicorp/go-hclog"

func GetLogger(name string) hclog.Logger {

	return hclog.New(&hclog.LoggerOptions{Name: name, Level: hclog.LevelFromString("DEBUG")})

}
