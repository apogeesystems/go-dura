package dura

type CallState string

const (
	YIELD   CallState = "Yield"
	RECURSE           = "Recurse"
	DONE              = "Done"
)

type GitRepoIter struct {
	configIter map[string]WatchConfig
	subIter    []subIter
}

type subIter struct {
	path  string
	wConf WatchConfig
	fs    bool
}

func NewGitRepoIter(config Config) (iter *GitRepoIter) {
	return &GitRepoIter{
		configIter: config.Repositories,
		subIter:    nil,
	}
}

func (iter *GitRepoIter) GetNext() (state CallState) {
	return
}

func IsValidDirectory(basePath string, childPath string, value WatchConfig) (valid bool, err error) {

	return
}
