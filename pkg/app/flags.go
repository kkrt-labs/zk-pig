package app

import (
	"github.com/kkrt-labs/go-utils/common"
	"github.com/kkrt-labs/go-utils/spf13"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	mainEntrypointFlag = &spf13.StringFlag{
		ViperKey:     "app.main.entrypoint.address",
		Name:         "main-ep-addr",
		Env:          "MAIN_ENTRYPOINT_ADDR",
		Description:  "Main entrypoint address",
		DefaultValue: common.Ptr(":8080"),
	}
	mainKeepAliveFlag = &spf13.StringFlag{
		ViperKey:     "app.main.entrypoint.keep-alive",
		Name:         "main-ep-keep-alive",
		Env:          "MAIN_ENTRYPOINT_KEEP_ALIVE",
		Description:  "Main entrypoint keep alive",
		DefaultValue: common.Ptr("0"),
	}
	mainReadTimeoutFlag = &spf13.StringFlag{
		ViperKey:     "app.main.read-timeout",
		Name:         "main-read-timeout",
		Env:          "MAIN_READ_TIMEOUT",
		Description:  "Main read timeout",
		DefaultValue: common.Ptr("30s"),
	}
	mainReadHeaderTimeoutFlag = &spf13.StringFlag{
		ViperKey:     "app.main.read-header-timeout",
		Name:         "main-read-header-timeout",
		Env:          "MAIN_READ_HEADER_TIMEOUT",
		Description:  "Main read header timeout",
		DefaultValue: common.Ptr("30s"),
	}
	mainWriteTimeoutFlag = &spf13.StringFlag{
		ViperKey:     "app.main.write-timeout",
		Name:         "main-write-timeout",
		Env:          "MAIN_WRITE_TIMEOUT",
		Description:  "Main write timeout",
		DefaultValue: common.Ptr("30s"),
	}
	mainIdleTimeoutFlag = &spf13.StringFlag{
		ViperKey:     "app.main.idle-timeout",
		Name:         "main-idle-timeout",
		Env:          "MAIN_IDLE_TIMEOUT",
		Description:  "Main idle timeout",
		DefaultValue: common.Ptr("30s"),
	}
	healthzEntrypointFlag = &spf13.StringFlag{
		ViperKey:     "app.healthz.entrypoint.address",
		Name:         "healthz-ep-addr",
		Env:          "HEALTHZ_ENTRYPOINT_ADDR",
		Description:  "Healthz entrypoint address",
		DefaultValue: common.Ptr(":8081"),
	}
	healthzKeepAliveFlag = &spf13.StringFlag{
		ViperKey:     "app.healthz.entrypoint.keep-alive",
		Name:         "healthz-ep-keep-alive",
		Env:          "HEALTHZ_ENTRYPOINT_KEEP_ALIVE",
		Description:  "Healthz entrypoint keep alive",
		DefaultValue: common.Ptr("0"),
	}
	healthzReadTimeoutFlag = &spf13.StringFlag{
		ViperKey:     "app.healthz.read-timeout",
		Name:         "healthz-read-timeout",
		Env:          "HEALTHZ_READ_TIMEOUT",
		Description:  "Healthz read timeout",
		DefaultValue: common.Ptr("30s"),
	}
	healthzReadHeaderTimeoutFlag = &spf13.StringFlag{
		ViperKey:     "app.healthz.read-header-timeout",
		Name:         "healthz-read-header-timeout",
		Env:          "HEALTHZ_READ_HEADER_TIMEOUT",
		Description:  "Healthz read header timeout",
		DefaultValue: common.Ptr("30s"),
	}
	healthzWriteTimeoutFlag = &spf13.StringFlag{
		ViperKey:     "app.healthz.write-timeout",
		Name:         "healthz-write-timeout",
		Env:          "HEALTHZ_WRITE_TIMEOUT",
		Description:  "Healthz write timeout",
		DefaultValue: common.Ptr("30s"),
	}
	healthzIdleTimeoutFlag = &spf13.StringFlag{
		ViperKey:     "app.healthz.idle-timeout",
		Name:         "healthz-idle-timeout",
		Env:          "HEALTHZ_IDLE_TIMEOUT",
		Description:  "Healthz idle timeout",
		DefaultValue: common.Ptr("30s"),
	}
)

func AddFlags(v *viper.Viper, f *pflag.FlagSet) {
	mainEntrypointFlag.Add(v, f)
	mainKeepAliveFlag.Add(v, f)
	mainReadTimeoutFlag.Add(v, f)
	mainReadHeaderTimeoutFlag.Add(v, f)
	mainWriteTimeoutFlag.Add(v, f)
	mainIdleTimeoutFlag.Add(v, f)
	healthzEntrypointFlag.Add(v, f)
	healthzKeepAliveFlag.Add(v, f)
	healthzReadTimeoutFlag.Add(v, f)
	healthzReadHeaderTimeoutFlag.Add(v, f)
	healthzWriteTimeoutFlag.Add(v, f)
	healthzIdleTimeoutFlag.Add(v, f)
}
