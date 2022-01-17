package dura

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

var (
	dbViper   *viper.Viper
	cacheHome string
	cacheFile string
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

	// Search config in home directory with name ".go-dura" (without extension).
	dbViper.AddConfigPath(cacheHome)
	dbViper.SetConfigType("json")
	dbViper.SetConfigName("runtime.db")
	cacheFile = strings.TrimRight(cacheHome, "/") + "/runtime.db"

}

type RuntimeLock struct {
	Pid *uint32
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
	return rl.LoadFile(cacheFile)
}

func (rl *RuntimeLock) LoadFile(filepath string) (err error) {
	var file *os.File
	if file, err = os.Open(filepath); err != nil {
		return
	}
	return dbViper.ReadConfig(file)
}

func (rl *RuntimeLock) Save() (err error) {
	return dbViper.WriteConfig()
}

func (rl *RuntimeLock) CreateDir(path string) (err error) {
	return os.MkdirAll(path, 0755)
}

func (rl *RuntimeLock) SaveToPath(path string) (err error) {
	return dbViper.WriteConfigAs(path)
}
