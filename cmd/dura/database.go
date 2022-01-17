package dura

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

var (
	dbViper     *viper.Viper
	cacheHome   string
	cacheFile   string
	runtimeLock RuntimeLock
)

func InitDatabase() {
	dbViper = viper.New()
	// Find home directory.
	if cacheHome, err = os.UserHomeDir(); err != nil {
		log.Fatalln(err)
	}

	dbViper.SetEnvPrefix("dura")

	if tmp := os.Getenv("DURA_CACHE_HOME"); tmp != "" {
		cacheHome = tmp
	} else {
		cacheHome = strings.TrimRight(cacheHome, "/") + "/.cache/dura"
	}
	fmt.Printf("Cache home: '%s'\n", cacheHome)

	// Search config in home directory with name ".go-dura" (without extension).
	dbViper.AddConfigPath(cacheHome)
	dbViper.SetConfigType("json")
	dbViper.SetConfigName("runtime")
	cacheFile = strings.TrimRight(cacheHome, "/") + "/runtime"
	dbViper.Set("pid", nil)

	readInRuntime(nil)
}

func readInRuntime(path *string) (err error) {
	if path == nil {
		if err = dbViper.ReadInConfig(); err == nil {
			//fmt.Fprintln(os.Stderr, "Runtime file loaded: ", viper.ConfigFileUsed())
		} else {
			return
		}
	} else {
		var file *os.File
		if file, err = os.Open(*path); err != nil {
			return
		}
		if err = dbViper.ReadConfig(file); err != nil {
			return
		}
	}
	//fmt.Printf("%+v", viper.Get("repos"))
	if err = dbViper.Unmarshal(&runtimeLock); err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%+v\n\n", config)
	//fmt.Printf("%s\n%s\n%+v\n\n", *config.Commit.Author, *config.Commit.Email, config.Commit.ExcludeGitConfig)
	//fmt.Printf("%+v\n\n", config.Repositories)
	return
}

type RuntimeLock struct {
	Pid *uint32 `json:"pid,omitempty" mapstructure:"pid,omitempty"`
}

func (rl *RuntimeLock) Empty() {
	rl.Pid = nil
}

func (rl *RuntimeLock) DefaultPath() (path string) {
	return cacheFile
}

func (rl *RuntimeLock) GetDuraCacheHome() (path string) {
	return cacheHome
}

func (rl *RuntimeLock) Load() (err error) {
	//fmt.Printf("Loading runtime cache file: '%s'\n", cacheFile)
	return readInRuntime(nil)
}

func (rl *RuntimeLock) LoadFile(filepath string) (err error) {
	return readInRuntime(&filepath)
}

func (rl *RuntimeLock) Save() (err error) {
	dbViper.Set("pid", rl.Pid)
	return dbViper.WriteConfig()
}

func (rl *RuntimeLock) CreateDir(path string) (err error) {
	return os.MkdirAll(path, 0755)
}

func (rl *RuntimeLock) SaveToPath(path string) (err error) {
	return dbViper.WriteConfigAs(path)
}
