package dura

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

var (
	configFile string
	configHome string
	config     Config
	err        error
)

func GetConfig() (config *Config) {
	return config
}

func InitConfig() {
	config = Config{
		Viper: viper.GetViper(),
		Commit: CommitConfig{
			ExcludeGitConfig: false,
		},
		Repositories: map[string]WatchConfig{},
	}

	// Find home directory.
	if configHome, err = os.UserHomeDir(); err != nil {
		log.Fatalln(err)
	}

	config.SetEnvPrefix("dura")

	if tmp := os.Getenv("DURA_CONFIG_HOME"); tmp != "" {
		configHome = tmp
	}

	// Search config in home directory with name ".go-dura" (without extension).
	config.AddConfigPath(configHome)
	config.SetConfigType("yaml")
	config.SetConfigName(".go-dura")

	config.SetDefault("commit", map[string]interface{}{
		"author":             nil,
		"email":              nil,
		"exclude_git_config": false,
	})
	config.SetDefault("repositories", map[string]WatchConfig{})

	config.AutomaticEnv() // read in environment variables that match

	config.OnConfigChange(func(e fsnotify.Event) {
		readInConfig()
	})
	config.WatchConfig()

	// If a config file is found, read it in.
	readInConfig()
}

func readInConfig() {
	if err = config.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file: ", config.ConfigFileUsed())
	}
	config.Unmarshal(&config)
}

type WatchConfig struct {
	Include  []string
	Exclude  []string
	MaxDepth int
}

func NewWatchConfig() (wc *WatchConfig) {
	return &WatchConfig{
		Include:  []string{},
		Exclude:  []string{},
		MaxDepth: 255,
	}
}

type Config struct {
	*viper.Viper
	Commit       CommitConfig           `mapstructure:"commit"`
	Repositories map[string]WatchConfig `mapstructure:"repos,omitempty"`
}

type CommitConfig struct {
	Author           *string `mapstructure:"author,omitempty"`
	Email            *string `mapstructure:"email,omitempty"`
	ExcludeGitConfig bool    `mapstructure:"exclude_git_config"`
}

func (c *Config) Empty() {
	c.Commit.ExcludeGitConfig = false
	c.Commit.Author = nil
	c.Commit.Email = nil
	c.Repositories = map[string]WatchConfig{}
}

func (c *Config) DefaultPath() (path string) {
	return strings.TrimRight(c.GetDuraConfigHome(), "/") + "/.go-dura.yaml"
}

func (c *Config) GetDuraConfigHome() (path string) {
	return configHome
}

func (c *Config) Load() {
	config.ReadInConfig()
}

func (c *Config) LoadFile(filepath string) (err error) {
	var file *os.File
	if file, err = os.Open(filepath); err != nil {
		return
	}
	return config.ReadConfig(file)
}

func (c *Config) Save() (err error) {
	return config.WriteConfig()
}

func (c *Config) CreateDir(path string) (err error) {
	return os.MkdirAll(path, 0755)
}

func (c *Config) SaveToPath(filename string) (err error) {
	return config.WriteConfigAs(filename)
}

func (c *Config) SetWatch(path string, cfg WatchConfig) (err error) {
	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(path); err != nil {
		return
	}
	if !fileInfo.IsDir() {
		return errors.New(fmt.Sprintf("path '%s' is not a directory", path))
	}
	if c.Repositories != nil {
		var ok bool
		if _, ok = c.Repositories[path]; !ok {
			c.Repositories[path] = cfg
			if err = c.Save(); err != nil {
				return
			}
			fmt.Printf("Now watching %s\n", path)
		} else {
			fmt.Printf("%s already being watched\n", path)
		}
	} else {
		return errors.New("no repositories set in config")
	}
	return
}

func (c *Config) SetUnwatch(path string) (err error) {
	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(path); err != nil {
		return
	}
	if !fileInfo.IsDir() {
		return errors.New(fmt.Sprintf("path '%s' is not a directory", path))
	}
	if c.Repositories != nil {
		var ok bool
		if _, ok = c.Repositories[path]; ok {
			delete(c.Repositories, path)
			if err = c.Save(); err != nil {
				return
			}
			fmt.Printf("Stopped watching %s\n", path)
		} else {
			fmt.Printf("%s is not being watched\n", path)
		}
	} else {
		return errors.New("no repositories set in config")
	}
	return
}

func (c *Config) GitRepos() (repos map[string]WatchConfig) {
	return c.Repositories
}
