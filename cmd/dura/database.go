package dura

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	dbViper      *viper.Viper
	cacheHome    string
	cacheFile    string
	runtimeLock  RuntimeLock
	dbConfigType = "json"
	dbConfigName = "runtime"
)

func InitDatabase() {
	log.Trace().Msg("entered InitDatabase")
	log.Trace().Msg("create new viper instance")
	dbViper = viper.New()
	log.Debug().Msg("new viper instance set to dbViper")
	// Find home directory.
	log.Trace().Msg("retrieve user's home directory")
	if cacheHome, err = os.UserHomeDir(); err != nil {
		log.Fatal().Err(err).Msg("error encountered while retrieving user's home directory")
	}

	dbViper.SetEnvPrefix("dura")
	log.Debug().Msg("dbViper environment variable prefix set to DURA")

	log.Trace().Msg("attempt to retrieve cache directory from DURA_CACHE_HOME environment variable value")
	if tmp := os.Getenv("DURA_CACHE_HOME"); tmp != "" {
		cacheHome = tmp
		log.Debug().Msgf("cache directory set via environment variable (DURA_CACHE_HOME) to %s", cacheHome)
	} else {
		cacheHome = strings.TrimRight(cacheHome, "/") + "/.cache/dura"
		log.Debug().Msgf("cache directory set to %s", cacheHome)
	}

	// Search config in home directory with name ".go-dura" (without extension).
	dbViper.AddConfigPath(cacheHome)

	dbViper.SetConfigType(dbConfigType)
	log.Debug().Msgf("runtime database config type set to %s", dbConfigType)
	dbViper.SetConfigName(dbConfigName)
	log.Debug().Msgf("runtime database config name set to %s", dbConfigName)
	cacheFile = strings.TrimRight(cacheHome, "/") + "/runtime"
	dbViper.Set("pid", nil)
	log.Debug().Msg("set configuration PID property to nil")

	log.Trace().Msg("reading in runtime database")
	readInRuntime(nil)
	log.Info().Msg("runtime database initialization complete")
	log.Trace().Msg("leaving InitDatabase")
}

func readInRuntime(path *string) (err error) {
	log.Trace().Msg("entered readInRuntime")
	if path == nil {
		log.Debug().Msg("reading runtime database from preset location")
		if err = dbViper.ReadInConfig(); err != nil {
			log.Error().Err(err).Msg("error encountered while attempting to read in runtime database")
			return
		}
		log.Debug().Msg("successfully loaded runtime database")
	} else {
		log.Debug().Msgf("reading runtime database from file %s", *path)
		var file *os.File
		log.Trace().Msgf("open file for reading %s", path)
		if file, err = os.Open(*path); err != nil {
			log.Error().Err(err).Msgf("error encountered attempting to open path %s", *path)
			return
		}
		log.Debug().Msgf("successfully opened file %s", path)
		log.Trace().Msgf("reading in runtime database from file %s", path)
		if err = dbViper.ReadConfig(file); err != nil {
			log.Error().Err(err).Msgf("error encountered attempting to read runtime database from path %s", path)
			return
		}
		log.Debug().Msgf("successfully read in runtime database from file %s", path)
	}
	log.Trace().Msg("unmarshalling dbViper into runtimeLock structure")
	if err = dbViper.Unmarshal(&runtimeLock); err != nil {
		log.Fatal().Err(err).Msg("error encountered attempting to unmarshal runtime database to runtimeLock structure")
		return
	}
	log.Info().Msgf("successfully loaded runtime database")
	log.Trace().Msg("leaving readInRuntime")
	return
}

type RuntimeLock struct {
	Pid *uint32 `json:"pid,omitempty" mapstructure:"pid,omitempty"`
}

func (rl *RuntimeLock) Empty() {
	log.Trace().Msg("entered Empty")
	log.Trace().Msg("emptying runtimeLock")
	rl.Pid = nil
	log.Trace().Msg("emptied runtiemLock")
	log.Trace().Msg("leaving Empty")
}

func (rl *RuntimeLock) DefaultPath() (path string) {
	log.Debug().Msgf("returning default runtime database path %s", cacheFile)
	return cacheFile
}

func (rl *RuntimeLock) GetDuraCacheHome() (path string) {
	log.Debug().Msgf("returning the cache directory %s", cacheHome)
	return cacheHome
}

func (rl *RuntimeLock) Load() (err error) {
	log.Trace().Msgf("entered Load")
	if err = readInRuntime(nil); err != nil {
		log.Error().Err(err).Msg("error encountered attempting to read in runtime database")
		return
	}
	log.Debug().Msg("successfully loaded runtime database")
	log.Trace().Msg("leaving Load")
	return
}

func (rl *RuntimeLock) LoadFile(filepath string) (err error) {
	log.Trace().Msg("entered LoadFile")
	if err = readInRuntime(&filepath); err != nil {
		log.Error().Err(err).Msgf("error encountered attempting to read in runtime database from %s", filepath)
		return
	}
	log.Debug().Msg("successfully loaded runtime database")
	log.Trace().Msg("leaving LoadFile")
	return
}

func (rl *RuntimeLock) Save() (err error) {
	log.Trace().Msg("entered Save")
	log.Trace().Msgf("setting runtime database PID: %v", rl.Pid)
	dbViper.Set("pid", rl.Pid)
	if err = dbViper.WriteConfig(); err != nil {
		log.Error().Err(err).Msg("error encountered attempting to save runtime database")
		return
	}
	log.Debug().Msg("successfully saved configuration after setting PID")
	log.Trace().Msg("leaving Save")
	return
}

func (rl *RuntimeLock) CreateDir(path string) (err error) {
	log.Trace().Msg("entered CreateDir")
	if err = os.MkdirAll(path, os.FileMode(fileMode)); err != nil {
		log.Error().Err(err).Uint32("mode", fileMode).Msgf("error encountered attempting to create directory %s", path)
		return
	}
	log.Debug().Uint32("mode", fileMode).Msgf("successfully created directory %s", path)
	log.Trace().Msg("leaving CreateDir")
	return
}

func (rl *RuntimeLock) SaveToPath(path string) (err error) {
	log.Trace().Msg("entered SaveToPath")
	if err = dbViper.WriteConfigAs(path); err != nil {
		log.Error().Err(err).Msgf("error encountered attempting to save runtime database to file %s", path)
		return
	}
	log.Debug().Msgf("successfully saved runtime database to file %s", path)
	log.Trace().Msg("leaving SaveToPath")
	return
}
