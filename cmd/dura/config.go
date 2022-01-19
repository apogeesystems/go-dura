package dura

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	configFile      string
	configHome      string
	config          Config
	err             error
	configType             = "toml"
	configName             = ".go-dura"
	DefSleepSeconds        = 5
	fileMode        uint32 = 0644
)

func GetConfig() (config *Config) {
	log.Trace().Msg("returning config")
	return config
}

func InitConfig() {
	config = Config{
		Commit: CommitConfig{
			ExcludeGitConfig: false,
		},
		Repositories: map[string]WatchConfig{},
	}

	// Find home directory.
	log.Trace().Msg("retrieve user's home directory")
	if configHome, err = os.UserHomeDir(); err != nil {
		log.Fatal().Err(err).Msg("error encountered while attempting to retrieve user's home directory")
	}

	viper.SetEnvPrefix("dura")
	log.Debug().Msg("set viper environment prefix to DURA")

	log.Trace().Msg("attempt to retrieve DURA_CONFIG_HOME environment variable value")
	if tmp := os.Getenv("DURA_CONFIG_HOME"); tmp != "" {
		configHome = tmp
		log.Debug().Msg("config directory set via DURA_CONFIG_HOME environment variable")
	}
	log.Debug().Msgf("config directory: %s", configHome)

	// Search config in home directory with name ".go-dura" (without extension).
	viper.AddConfigPath(configHome)
	log.Debug().Msgf("viper config path set to %s", configHome)
	viper.SetConfigType(configType)
	log.Debug().Msgf("viper config type set to %s", configType)
	viper.SetConfigName(configName)
	log.Debug().Msgf("viper config file name set to %s", configName)

	viper.SetDefault("commit", map[string]interface{}{
		"author":             nil,
		"email":              nil,
		"exclude_git_config": false,
	})
	log.Debug().Msg("viper default commit structure set")
	viper.SetDefault("dura.sleep_seconds", DefSleepSeconds)
	log.Debug().Msgf("Dura sleep seconds set to %d seconds", DefSleepSeconds)

	viper.AutomaticEnv() // read in environment variables that match
	log.Debug().Msg("viper automatic environment setup called")

	log.Trace().Msg("setting callback for OnConfigChange")
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debug().Msg("configuration change detected in config file")
		log.Trace().Msg("calling readInConfig()")
		readInConfig()
	})
	viper.WatchConfig()
	log.Debug().Msg("viper set to watch configuration file for changes")

	// If a config file is found, read it in.
	log.Debug().Msg("calling readInConfig()")
	readInConfig()
	log.Info().Msg("configuration initialization complete")
}

func readInConfig() (err error) {
	log.Trace().Msg("entered readInConfig")
	if err = viper.ReadInConfig(); err == nil {
		log.Info().Msgf("Loaded configuration: %s", viper.ConfigFileUsed())
	} else {
		log.Fatal().Err(err).Msg("error encountered while attempting to read in configuration")
	}
	if err = viper.Unmarshal(&config); err != nil {
		log.Fatal().Err(err).Msg("error encountered while unmarshalling viper in config structure")
	}
	log.Trace().Msg("leaving readInConfig")
	return
}

type WatchConfig struct {
	Include  []string `toml:"include" mapstructure:"include,omitempty"`
	Exclude  []string `toml:"exclude" mapstructure:"exclude,omitempty"`
	MaxDepth int      `toml:"max_depth" mapstructure:"max_depth,omitempty"`
}

func NewWatchConfig() (wc *WatchConfig) {
	log.Trace().Msg("returning new watch configuration")
	return &WatchConfig{
		Include:  []string{},
		Exclude:  []string{},
		MaxDepth: 255,
	}
}

type Config struct {
	Dura         DuraConfig             `toml:"dura"`
	Commit       CommitConfig           `toml:"commit"`
	Repositories map[string]WatchConfig `toml:"repos" mapstructure:"repos"`
}

type DuraConfig struct {
	SleepSeconds int `toml:"sleep_seconds" mapstructure:"sleep_seconds"`
}

type CommitConfig struct {
	Author           *string `toml:"author"`
	Email            *string `toml:"email"`
	ExcludeGitConfig bool    `toml:"exclude_git_config"`
}

func (c *Config) Empty() {
	log.Trace().Msg("entered Empty")
	log.Trace().Msg("emptying configuration")
	c.Dura.SleepSeconds = 5
	c.Commit.ExcludeGitConfig = false
	c.Commit.Author = nil
	c.Commit.Email = nil
	c.Repositories = map[string]WatchConfig{}
	log.Trace().Msg("emptied configuration")
	log.Trace().Msgf("leaving Empty")
}

func (c *Config) DefaultPath() (path string) {
	log.Trace().Msgf("returning configuration default path: %s", strings.TrimRight(c.GetDuraConfigHome(), "/")+fmt.Sprintf("/%s.%s", configName, configType))
	return strings.TrimRight(c.GetDuraConfigHome(), "/") + fmt.Sprintf("/%s.%s", configName, configType)
}

func (c *Config) GetDuraConfigHome() (path string) {
	log.Trace().Msgf("returning configuration home directory: %s", configHome)
	return configHome
}

func (c *Config) Load() (err error) {
	log.Trace().Msg("entered Load")
	log.Trace().Msg("calling viper.ReadInConfig()")
	err = viper.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Msg("error encountered while reading in configuration")
	}
	log.Debug().Msg("viper successfully loaded configuration")
	log.Trace().Msgf("leaving Load")
	return
}

func (c *Config) LoadFile(filepath string) (err error) {
	log.Trace().Msg("entered LoadFile")
	var file *os.File
	log.Trace().Msgf("opening %s", filepath)
	if file, err = os.Open(filepath); err != nil {
		log.Error().Err(err).Msgf("error encountered attempting to open file %s", filepath)
		return
	}
	if err = viper.ReadConfig(file); err != nil {
		log.Error().Err(err).Msgf("error encountered while reading configuration file (%s)", filepath)
		return
	}
	log.Trace().Msg("leaving LoadFile")
	return
}

func (c *Config) Save() (err error) {
	log.Trace().Msg("entered Save")
	if err = viper.WriteConfig(); err != nil {
		log.Error().Err(err).Msg("error encountered attempting to save configuration")
		return
	}
	log.Debug().Msg("successfully saved configuration")
	log.Trace().Msg("leaving Save")
	return
}

func (c *Config) CreateDir(path string) (err error) {
	log.Trace().Msg("entered CreateDir")
	if err = os.MkdirAll(path, os.FileMode(fileMode)); err != nil {
		log.Error().Err(err).Uint32("mode", fileMode).Msgf("error encountered attempting to create directory %s", path)
		return
	}
	log.Debug().Uint32("mode", fileMode).Msgf("successfully created directory %s", path)
	log.Trace().Msg("leaving CreateDir")
	return
}

func (c *Config) SaveToPath(filename string) (err error) {
	log.Trace().Msg("entered SaveToPath")
	if err = viper.WriteConfigAs(filename); err != nil {
		log.Error().Err(err).Msgf("error encountered attempting to save configuration to %s", filename)
		return
	}
	log.Debug().Msgf("successfully saved configuration to %s", filename)
	log.Trace().Msg("leaving SaveToPath")
	return
}

func (c *Config) SetWatch(path string, cfg WatchConfig) (err error) {
	log.Trace().Msg("entered SetWatch")
	var (
		fileInfo os.FileInfo
		isRepo   bool
	)
	log.Trace().Msgf("retrieve file information for path %s", path)
	if fileInfo, err = os.Stat(path); err != nil {
		log.Error().Err(err).Msgf("error encountered attempting to retrieve file information for path %s", path)
		return
	}
	log.Trace().Msg("checking that path provided is a directory")
	if !fileInfo.IsDir() {
		err = errors.New(fmt.Sprintf("path '%s' is not a directory", path))
		log.Error().Err(err).Msg("path given is not a directory")
		return
	}
	log.Debug().Msgf("path 5s is a directory, can proceed", path)
	log.Trace().Msg("check if directory (path)  is a git repository")
	if isRepo, err = IsRepo(path); err != nil || !isRepo {
		if err == nil {
			err = errors.New(fmt.Sprintf("%s is not a repo", path))
		}
		log.Error().Err(err).Msgf("error encountered testing directory %s", path)
		return
	}
	log.Debug().Msgf("directory %s is a git repository, can proceed", path)
	log.Trace().Msg("check that repository hasn't already been added")
	if c.Repositories != nil {
		log.Trace().Msg("repositories exist in configuration (i.e. non-nil)")
		var ok bool
		log.Trace().Msg("check if path already exists in the configuration")
		if _, ok = c.Repositories[path]; !ok {
			log.Trace().Msg("add path to repositories watched")
			c.Repositories[path] = cfg
			log.Trace().Msg("save configuration after repository update")
			if err = c.Save(); err != nil {
				log.Error().Err(err).Msg("error encountered attempting to save configuration after repository update")
				return
			}
			log.Info().Msgf("now watching %s", path)
		} else {
			log.Warn().Msgf("%s is already being watched")
		}
	} else {
		log.Trace().Msg("initialize configuration repositories map")
		c.Repositories = map[string]WatchConfig{
			path: cfg,
		}
		log.Trace().Msg("save configuration after repository update")
		if err = c.Save(); err != nil {
			log.Error().Err(err).Msg("error encountered attempting to save configuration after repository update")
			return
		}
		log.Info().Msgf("now watching %s", path)
	}
	log.Trace().Msg("leaving SetWatch")
	return
}

func (c *Config) SetUnwatch(path string) (err error) {
	log.Trace().Msg("entered SetUnwatch")
	var fileInfo os.FileInfo
	log.Trace().Msgf("retrieve file information for path %s", path)
	if fileInfo, err = os.Stat(path); err != nil {
		log.Error().Err(err).Msgf("error encountered attempting to retrieve file information for path %s", path)
		return
	}
	log.Trace().Msg("checking that path provided is a directory")
	if !fileInfo.IsDir() {
		err = errors.New(fmt.Sprintf("path '%s' is not a directory", path))
		log.Error().Err(err).Msg("path given is not a directory")
		return
	}
	log.Debug().Msgf("path 5s is a directory, can proceed", path)
	if c.Repositories != nil {
		log.Trace().Msg("repositories exist in configuration (i.e. non-nil)")
		var ok bool
		log.Trace().Msg("check if path already exists in the configuration")
		if _, ok = c.Repositories[path]; ok {
			log.Trace().Msg("remove repository from configuration")
			delete(c.Repositories, path)
			log.Debug().Msg("successfully removed repository from configuration")
			log.Trace().Msg("saving configuration")
			if err = c.Save(); err != nil {
				log.Error().Err(err).Msg("error encountered attempting to save configuration after repository update")
				return
			}
			log.Debug().Msg("successfully saved configuration after repository update")
			log.Info().Msgf("stopped watching %s", path)
		} else {
			log.Warn().Msgf("%s is not being watched, nothing to do", path)
		}
	} else {
		log.Trace().Msg("repositories map in configuration is nil")
		c.Repositories = map[string]WatchConfig{}
		log.Debug().Msg("initialized repositories map in configuration")
		log.Trace().Msg("saving configuration")
		if err = c.Save(); err != nil {
			log.Error().Err(err).Msg("error encountered attempting to save configuration after repositories initialization")
			return
		}
		log.Debug().Msg("successfully saved configuration")
		log.Warn().Msgf("%s is not being watched, nothing to do", path)
	}
	log.Trace().Msg("leaving SetUnwatch")
	return
}

func (c *Config) GitRepos() (repos map[string]WatchConfig) {
	log.Trace().Msg("returning repositories from configuration")
	return c.Repositories
}
