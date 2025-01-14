package log

import (
	"fmt"

	"github.com/kkrt-labs/kakarot-controller/pkg/common"
	"github.com/kkrt-labs/kakarot-controller/pkg/spf13"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	logLevelFlag = &spf13.StringFlag{
		ViperKey:     "log.level",
		Name:         "log-level",
		Env:          "LOG_LEVEL",
		Description:  fmt.Sprintf("Log level (one of %q)", levelsStr),
		DefaultValue: common.Ptr(levelsStr[InfoLevel]),
	}
	formatFlag = &spf13.StringFlag{
		ViperKey:     "log.format",
		Name:         "log-format",
		Env:          "LOG_FORMAT",
		Description:  fmt.Sprintf("Log formatter (one of %q)", formatsStr),
		DefaultValue: common.Ptr(formatsStr[TextFormat]),
	}
)

func AddFlags(v *viper.Viper, f *pflag.FlagSet) {
	logLevelFlag.Add(v, f)
	formatFlag.Add(v, f)
}
