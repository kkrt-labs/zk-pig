package spf13

import (
	"fmt"

	"github.com/kkrt-labs/kakarot-controller/pkg/common"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Flag interface {
	Add(v *viper.Viper, f *pflag.FlagSet)
}

func AddFlag(v *viper.Viper, f *pflag.FlagSet, flag Flag) {
	flag.Add(v, f)
}

type StringFlag struct {
	ViperKey     string
	Name         string
	Shorthand    string
	Env          string
	Description  string
	DefaultValue *string
}

func (flag *StringFlag) Add(v *viper.Viper, f *pflag.FlagSet) {
	if flag.Name != "" {
		if flag.Shorthand != "" {
			f.StringP(flag.Name, flag.Shorthand, common.Val(flag.DefaultValue), FlagDesc(flag.Description, flag.Env))
		} else {
			f.String(flag.Name, common.Val(flag.DefaultValue), FlagDesc(flag.Description, flag.Env))
		}
		_ = v.BindPFlag(flag.ViperKey, f.Lookup(flag.Name))
	}
	if flag.Env != "" {
		_ = v.BindEnv(flag.ViperKey, flag.Env)
	}

	if flag.DefaultValue != nil {
		v.SetDefault(flag.ViperKey, *flag.DefaultValue)
	}
}

type StringArrayFlag struct {
	ViperKey     string
	Name         string
	Shorthand    string
	Env          string
	Description  string
	DefaultValue []string
}

func (flag *StringArrayFlag) Add(v *viper.Viper, f *pflag.FlagSet) {
	if flag.Name != "" {
		if flag.Shorthand != "" {
			f.StringArrayP(flag.Name, flag.Shorthand, flag.DefaultValue, FlagDesc(flag.Description, flag.Env))
		} else {
			f.StringArray(flag.Name, flag.DefaultValue, FlagDesc(flag.Description, flag.Env))
		}
		_ = v.BindPFlag(flag.ViperKey, f.Lookup(flag.Name))
	}
	if flag.Env != "" {
		_ = v.BindEnv(flag.ViperKey, flag.Env)
	}
	if len(flag.DefaultValue) > 0 {
		v.SetDefault(flag.ViperKey, flag.DefaultValue)
	}
}

type BoolFlag struct {
	ViperKey     string
	Name         string
	Shorthand    string
	Env          string
	Description  string
	DefaultValue *bool
}

func (flag *BoolFlag) Add(v *viper.Viper, f *pflag.FlagSet) {
	if flag.Name != "" {
		if flag.Shorthand != "" {
			f.BoolP(flag.Name, flag.Shorthand, common.Val(flag.DefaultValue), FlagDesc(flag.Description, flag.Env))
		} else {
			f.Bool(flag.Name, common.Val(flag.DefaultValue), FlagDesc(flag.Description, flag.Env))
		}
		_ = v.BindPFlag(flag.ViperKey, f.Lookup(flag.Name))
	}
	if flag.Env != "" {
		_ = v.BindEnv(flag.ViperKey, flag.Env)
	}
	v.SetDefault(flag.ViperKey, flag.DefaultValue)
}

func FlagDesc(desc, envVar string) string {
	if envVar != "" {
		desc = fmt.Sprintf("%v [env: %v]", desc, envVar)
	}

	return desc
}
