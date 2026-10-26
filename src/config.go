package opennox

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"

	"github.com/opennox/libs/env"
	"github.com/opennox/libs/log"
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	viper.SetConfigName(configName)
	viper.AddConfigPath(wd)
	if sdir := env.AppUserDir(); sdir != "" {
		viper.AddConfigPath(sdir)
		configPath = filepath.Join(sdir, configName+"."+configExt)
	} else {
		configPath = filepath.Join(wd, configName+"."+configExt)
	}
	viper.AddConfigPath(filepath.Dir(os.Args[0]))
	if runtime.GOOS != "windows" {
		if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
			viper.AddConfigPath(filepath.Join(dir, "opennox"))
		}
		if home, err := os.UserHomeDir(); err == nil {
			// Linux Snapcraft installation replaces HOME variable
			if rhome := os.Getenv("SNAP_REAL_HOME"); rhome != "" {
				home = rhome
			}
			viper.AddConfigPath(filepath.Join(home, ".config/opennox"))
		}
	}
}

const (
	configName        = "opennox"
	configExt         = "yml"
	configNoxDataPath = "game.data"
	configNoxSerial   = "game.serial"
)

var (
	configLog      = log.New("config")
	configPath     string
	configDirty    bool
	configReadOnly = isDedicatedServer
	onConfigRead   []func()
)

// Finite owners cancel their binding at shutdown. Process-wide bindings may
// ignore the returned function; vacant slots are reused without retaining owners.
func registerOnConfigRead(fnc func()) func() {
	index := len(onConfigRead)
	for i, old := range onConfigRead {
		if old == nil {
			index = i
			break
		}
	}
	if index == len(onConfigRead) {
		onConfigRead = append(onConfigRead, fnc)
	} else {
		onConfigRead[index] = fnc
	}
	active := true
	return func() {
		if !active {
			return
		}
		active = false
		onConfigRead[index] = nil
		for len(onConfigRead) > 0 && onConfigRead[len(onConfigRead)-1] == nil {
			onConfigRead = onConfigRead[:len(onConfigRead)-1]
		}
	}
}

func writeConfig() error {
	if configReadOnly || env.IsE2E() {
		return nil
	}
	configLog.Printf("writing to %q", configPath)
	if err := viper.WriteConfigAs(configPath); err != nil {
		configLog.Printf("cannot save file: %v", err)
		return err
	}
	return nil
}

func maybeWriteConfig() {
	if !configDirty {
		return
	}
	_ = writeConfig()
}

func writeConfigLater() {
	configDirty = true
}

func configStrPtr(key, env string, def string, ptr *string) {
	viper.SetDefault(key, def)
	if env != "" {
		viper.BindEnv(key, env)
	}
	*ptr = viper.GetString(key)
	registerOnConfigRead(func() {
		*ptr = viper.GetString(key)
	})
}

func configBoolPtr(key, env string, def bool, ptr *bool) func() {
	viper.SetDefault(key, def)
	if env != "" {
		viper.BindEnv(key, env)
	}
	*ptr = viper.GetBool(key)
	return registerOnConfigRead(func() {
		*ptr = viper.GetBool(key)
	})
}

func configHiddenBoolPtr(key, env string, ptr *bool) func() {
	if env != "" {
		viper.BindEnv(key, env)
	}
	*ptr = viper.GetBool(key)
	return registerOnConfigRead(func() {
		*ptr = viper.GetBool(key)
	})
}

func readConfig(path string) error {
	defer func() {
		for _, fnc := range onConfigRead {
			if fnc != nil {
				fnc()
			}
		}
	}()
	if path != "" {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		viper.SetConfigFile(path)
		configPath = path
	}
	err := viper.ReadInConfig()
	if err == nil {
		configPath = viper.ConfigFileUsed()
		configDirty = false
		configLog.Printf("using file: %q", configPath)
		return nil
	}
	if _, ok := err.(viper.ConfigFileNotFoundError); ok || os.IsNotExist(err) {
		writeConfigLater()
		configLog.Println("file not found, using defaults")
		return nil
	}
	// There's a weird behavior in viper that it may try to read config file without extension.
	// It's a problem for us because we have "opennox" binary on Linux, so viper will read the binary itself and fail.
	// So we check this case and ignore it as if the config was not found.
	if filepath.Base(viper.ConfigFileUsed()) == configName {
		writeConfigLater()
		configLog.Println("file not found, using defaults")
		return nil
	}
	return err // error parsing config
}
