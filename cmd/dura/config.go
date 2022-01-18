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
	configFile      string
	configHome      string
	config          Config
	err             error
	configType      = "toml"
	configName      = ".go-dura"
	DefSleepSeconds = 5
)

func GetConfig() (config *Config) {
	return config
}

func InitConfig() {
	config = Config{
		//Viper: viper.GetViper(),
		Commit: CommitConfig{
			ExcludeGitConfig: false,
		},
		Repositories: map[string]WatchConfig{},
	}

	// Find home directory.
	if configHome, err = os.UserHomeDir(); err != nil {
		log.Fatalln(err)
	}

	viper.SetEnvPrefix("dura")

	if tmp := os.Getenv("DURA_CONFIG_HOME"); tmp != "" {
		configHome = tmp
	}

	// Search config in home directory with name ".go-dura" (without extension).
	viper.AddConfigPath(configHome)
	viper.SetConfigType(configType)
	viper.SetConfigName(configName)

	viper.SetDefault("commit", map[string]interface{}{
		"author":             nil,
		"email":              nil,
		"exclude_git_config": false,
	})
	viper.SetDefault("dura.sleep_seconds", DefSleepSeconds)

	viper.AutomaticEnv() // read in environment variables that match

	viper.OnConfigChange(func(e fsnotify.Event) {
		readInConfig()
	})
	viper.WatchConfig()

	// If a config file is found, read it in.
	readInConfig()
}

func readInConfig() (err error) {
	if err = viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Config file loaded: ", viper.ConfigFileUsed())
	} else {
		log.Fatal(err)
	}
	//fmt.Printf("%+v", viper.Get("repos"))
	if err = viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%+v\n\n", config)
	//fmt.Printf("%s\n%s\n%+v\n\n", *config.Commit.Author, *config.Commit.Email, config.Commit.ExcludeGitConfig)
	//fmt.Printf("%+v\n\n", config.Repositories)
	return
}

type WatchConfig struct {
	Include  []string `toml:"include" mapstructure:"include,omitempty"`
	Exclude  []string `toml:"exclude" mapstructure:"exclude,omitempty"`
	MaxDepth int      `toml:"max_depth" mapstructure:"max_depth,omitempty"`
}

func NewWatchConfig() (wc *WatchConfig) {
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
	c.Dura.SleepSeconds = 5
	c.Commit.ExcludeGitConfig = false
	c.Commit.Author = nil
	c.Commit.Email = nil
	c.Repositories = map[string]WatchConfig{}
}

func (c *Config) DefaultPath() (path string) {
	return strings.TrimRight(c.GetDuraConfigHome(), "/") + fmt.Sprintf("/%s.%s", configName, configType)
}

func (c *Config) GetDuraConfigHome() (path string) {
	return configHome
}

func (c *Config) Load() {
	viper.ReadInConfig()
}

func (c *Config) LoadFile(filepath string) (err error) {
	var file *os.File
	if file, err = os.Open(filepath); err != nil {
		return
	}
	return viper.ReadConfig(file)
}

func (c *Config) Save() (err error) {
	return viper.WriteConfig()
}

func (c *Config) CreateDir(path string) (err error) {
	return os.MkdirAll(path, 0755)
}

func (c *Config) SaveToPath(filename string) (err error) {
	return viper.WriteConfigAs(filename)
}

func (c *Config) SetWatch(path string, cfg WatchConfig) (err error) {
	var (
		fileInfo os.FileInfo
		isRepo   bool
	)
	if fileInfo, err = os.Stat(path); err != nil {
		return
	}
	if !fileInfo.IsDir() {
		return errors.New(fmt.Sprintf("path '%s' is not a directory", path))
	}
	if isRepo, err = IsRepo(path); err != nil || !isRepo {
		if err == nil {
			err = errors.New(fmt.Sprintf("%s is not a repo", path))
		}
		return
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
