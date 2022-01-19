/*
Copyright © 2022 Dane Nelson <apogeesystemsllc@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package dura

import (
	"errors"
	"fmt"
	git "github.com/libgit2/git2go/v33"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

type CaptureStatus struct {
	DuraBranch string `json:"dura_branch"`
	CommitHash string `json:"commit_hash"`
	BaseHash   string `json:"base_hash"`
}

func (cs *CaptureStatus) Display() {
	log.Trace().Msg("entering Display")
	log.Trace().Dict("cs", zerolog.Dict().Str("DuraBranch", cs.DuraBranch).Str("CommitHash", cs.CommitHash).Str("BaseHash", cs.BaseHash)).Msg("calling Display for capture status (cs)")
	log.Info().Msgf("dura: %s\ncommit_hash: %s\nbase: %s", cs.DuraBranch, cs.CommitHash, cs.BaseHash)
	log.Trace().Msg("leaving Display")
	return
}

func IsRepo(path string) (isRepo bool, err error) {
	log.Trace().Msg("entering IsRepo")
	logger := log.With().Str("path", path).Logger()
	logger.Trace().Msgf("checking if '%s' is a git repository", path)
	isRepo = true
	if _, err = git.OpenRepository(path); err != nil {
		logger.Error().Err(err).Msgf("error encountered while opening repository '%s", path)
		isRepo = false
	}
	logger.Debug().Msg("path is not a git repository")
	log.Trace().Msg("leaving IsRepo")
	return
}

func statusCheck(repo *git.Repository) (ok bool, err error) {
	log.Trace().Msg("entering statusCheck")
	logger := log.With().Str("repo", repo.Path()).Logger()
	logger.Trace().Msgf("checking status list of repository '%s'", repo.Path())
	var (
		statusList *git.StatusList
		count      int
	)
	logger.Trace().Interface("opt", nil).Msg("calling repo.StatusList(opt)")
	if statusList, err = repo.StatusList(nil); err != nil {
		logger.Error().Err(err).Msg("error encountered while calling repo.StatusList")
		return
	}
	logger.Debug().Msg("repository status list is non-nil")
	logger.Trace().Msg("calling statusList.EntryCount()")
	if count, err = statusList.EntryCount(); err != nil {
		logger.Error().Err(err).Msg("error encountered while calling statusList.EntryCount()")
		return
	}
	logger.Debug().Msg("call to statusList.EntryCount() was successful")
	logger.Trace().Int("count", count).Msgf("statusList has %d entries", count)
	if count > 0 {
		logger.Debug().Msg("statusList is not empty (PASS)")
		ok = true
	} else {
		logger.Debug().Msg("statusList is empty (FAIL)")
		ok = false
	}
	logger.Debug().Bool("pass", ok).Msgf("repository passes: %t", ok)
	log.Trace().Msg("leaving statusCheck")
	return
}

func headPeelToCommit(repo *git.Repository) (obj *git.Commit, err error) {
	log.Trace().Msg("entereing headPeelToCommit")
	logger := log.With().Str("repo", repo.Path()).Logger()
	logger.Trace().Msgf("peeling head of repository '%s'", repo.Path())
	var (
		head   *git.Reference
		objObj *git.Object
	)
	logger.Trace().Msg("calling repo.Head()")
	if head, err = repo.Head(); err != nil {
		log.Error().Err(err).Str("repo", repo.Path()).Msg("error encountered while calling repo.Head()")
		return
	}
	logger.Debug().Msg("head successfully retrieved")
	logger.Trace().Msg("calling head.Peel(git.ObjectCommit)")
	if objObj, err = head.Peel(git.ObjectCommit); err != nil {
		log.Error().Err(err).Str("repo", repo.Path()).Msg("error encountered while calling head.Peel(git.ObjectCommit)")
		return
	}
	logger.Debug().Str("oid", objObj.Id().String()).Msg("commit object found for head")
	logger.Trace().Msg("casting abstract git.Object to git.Commit (calling objObj.AsCommit()")
	if obj, err = objObj.AsCommit(); err != nil {
		log.Error().Err(err).Str("repo", repo.Path()).Msg("error encountered while calling objObj.AsCommit()")
		return
	}
	logger.Debug().Str("commit", obj.Id().String()).Msg("head commit successfully retrieved")
	log.Trace().Msg("leaving headPeelToCommit")
	return
}

func Capture(path string) (cs *CaptureStatus, err error) {
	log.Trace().Msg("entering Capture")
	logger := log.With().Str("path", path).Logger()
	var (
		repo            *git.Repository
		head            *git.Commit
		message         = "dura auto-backup"
		statusCheckPass bool
		branchName      string
		branchCommit    *git.Commit
	)
	logger.Trace().Msgf("calling git.OpenRepository for path '%s'", path)
	if repo, err = git.OpenRepository(path); err != nil {
		log.Error().Err(err).Str("path", path).Msgf("error encountered while attempting to open git repository at '%s'", path)
		return
	}
	logger.Debug().Msgf("opened repository at '%s'", path)

	// Get the repo HEAD, peel to the latest Commit as "head"
	logger.Trace().Msg("calling headPeelToCommit")
	if head, err = headPeelToCommit(repo); err != nil {
		logger.Error().Err(err).Msg("error encountered while calling headPeelToCommit")
		return
	}
	logger.Debug().Str("commit", head.Id().String()).Msg("successfully retrieved repository head commit")

	logger.Trace().Msg("executing statusCheck")
	if statusCheckPass, err = statusCheck(repo); err != nil || !statusCheckPass {
		if err == nil {
			err = errors.New("repository status list is empty")
		}
		logger.Error().Err(err).Msg("error encountered while executing statusCheck")
		return
	}
	logger.Debug().Msg("repository passed status check")

	if head.Id() != nil {
		log.Trace().Msg("setting Dura branch name")
		branchName = fmt.Sprintf("dura/%s", head.Id().String())
		logger = logger.With().Str("branch", branchName).Logger()
		logger.Debug().Msg("Dura branch name set")
	} else {
		err = errors.New("head.Id() was nil")
		logger.Error().Err(err).Msg("head ID was nil, without it the Dura branch can be named")
		return
	}

	logger.Trace().Msg("calling findHead")
	if branchCommit, err = findHead(repo, branchName); err != nil {
		logger.Error().Err(err).Msgf("could not find head for branch %s, branch may not yet exist", branchName)
		logger.Trace().Msg("checking if branch exists")
		if _, err = repo.LookupBranch(branchName, git.BranchLocal); err != nil {
			logger.Error().Err(err).Msgf("error encountered while looking for branch %s", branchName)
			logger.Debug().Msgf("branch %s does not exist", branchName)
			var branch *git.Branch
			logger.Trace().Str("target", head.Id().String()).Msgf("creating branch %s", branchName)
			if branch, err = repo.CreateBranch(branchName, head, false); err != nil {
				logger.Error().Err(err).Msgf("error encountered while creating branch %s", branchName)
				return
			}
			logger.Debug().Msgf("successfully created branch %s", branchName)
			var commitObj *git.Object
			logger.Trace().Msgf("find (peel to) latest commit for branch %s")
			if commitObj, err = branch.Peel(git.ObjectCommit); err != nil {
				logger.Error().Err(err).Msg("error encountered while peeling branch to latest commit")
				return
			}
			logger.Debug().Str("oid", commitObj.Id().String()).Msgf("peel to latest commit for branch %s was successful")
			logger.Trace().Msg("casting commit object to git.Commit (commitObj.AsCommit())")
			if branchCommit, err = commitObj.AsCommit(); err != nil {
				logger.Error().Err(err).Msgf("error occured while casting %s to commit", commitObj.Id().String())
				return
			}
			logger.Debug().Str("commit", branchCommit.Id().String()).Msgf("successfully retrieved latest commit for branch %s", branchName)
		}
	}

	var index *git.Index
	logger.Trace().Msg("retrieving repository index")
	if index, err = repo.Index(); err != nil {
		logger.Error().Err(err).Msgf("error encountered while retrieving repository (%s) index", repo.Path())
		return
	}
	logger.Debug().Msgf("successfully retrieved repository (%s) index", repo.Path())
	logger.Trace().Msgf("add wildcard to repository (%s) index pathspec (*)", repo.Path())
	if err = index.AddAll([]string{"*"}, git.IndexAddDefault, nil); err != nil {
		logger.Error().Err(err).Msgf("error encountered while adding wildcard to repository (%s) index", repo.Path())
		return
	}
	logger.Trace().Msgf("successfully added wildcard to repository (%s) index pathsepc", repo.Path())

	var (
		dirtyDiff *git.Diff
		oldTree   *git.Tree
		diffOpts  git.DiffOptions
		deltas    int
	)
	if branchCommit != nil {
		logger.Debug().Msg("branchCommit is non-nil")
		logger.Trace().Msg("retrieve branchCommit tree")
		if oldTree, err = branchCommit.Tree(); err != nil {
			logger.Error().Err(err).Msg("error encountered while retrieving branchCommit tree")
			logger.Trace().Msg("retrieve tree of repository head commit")
			if oldTree, err = head.Tree(); err != nil {
				logger.Error().Err(err).Msg("error encountered while retrieving tree of the repository head commit")
				return
			}
			logger.Debug().Msg("successfully retrieved tree of the repository head commit")
		}
		logger.Debug().Msg("successfully retrieved branchCommit tree")
	} else {
		logger.Debug().Msg("branchCommit is nil")
		logger.Trace().Msg("retrieve tree of repository head commit")
		if oldTree, err = head.Tree(); err != nil {
			logger.Error().Err(err).Msg("error encountered while retrieving tree of the repository head commit")
			return
		}
		logger.Debug().Msg("successfully retrieved tree of the repository head commit")
	}

	logger.Trace().Msg("setting diff options")
	if diffOpts, err = git.DefaultDiffOptions(); err != nil {
		logger.Error().Err(err).Msg("error encountered while attempting to set diff options")
		return
	}
	logger.Trace().Msg("setting diff options flags to git.DiffIncludeUntracked")
	diffOpts.Flags = git.DiffIncludeUntracked
	logger.Trace().Msg("setting diff options pathspec to wildcard (*)")
	diffOpts.Pathspec = []string{"*"}
	logger.Debug().Msg("diff options set")

	logger.Trace().Str("oldTree", oldTree.Id().String()).Str("index", index.Path()).Msg("calling repo.DiffTreeToIndex")
	if dirtyDiff, err = repo.DiffTreeToIndex(
		oldTree,
		index,
		&diffOpts,
	); err != nil {
		logger.Error().Err(err).Str("oldTree", oldTree.Id().String()).Msg("error encountered while getting diff of tree-to-index")
		return
	}
	logger.Debug().Msg("successfully retrieved diffs")
	logger.Trace().Msg("get number of deltas in diff")
	if deltas, err = dirtyDiff.NumDeltas(); err != nil || deltas == 0 {
		if err == nil {
			err = errors.New("no differences detected")
		}
		logger.Error().Int("deltas", deltas).Err(err).Msg("error encountered while retrieving deltas")
		return
	}
	logger.Debug().Msg("deltas found")

	var (
		treeOid *git.Oid
		tree    *git.Tree
	)
	logger.Trace().Msgf("write index (%s) to repository %s", index.Path(), repo.Path())
	if treeOid, err = index.WriteTreeTo(repo); err != nil {
		logger.Error().Err(err).Msgf("error encountered attempting to write index (%s) to repository %s", index.Path(), repo.Path())
		return
	}
	logger.Debug().Str("tree", treeOid.String()).Msg("successfully wrote index to repository")
	logger.Trace().Msg("lookup tree in repository")
	if tree, err = repo.LookupTree(treeOid); err != nil {
		logger.Error().Err(err).Msgf("error encountered while looking for tree %s in repository %s", treeOid.String(), repo.Path())
		return
	}
	logger.Debug().Msgf("found tree %s in repository %s", treeOid.String(), repo.Path())

	logger.Trace().Msg("set commit signature")
	var committer = &git.Signature{
		Name:  getGitAuthor(repo),
		Email: getGitEmail(repo),
		When:  time.Time{},
	}
	logger.Debug().Dict("committer", zerolog.Dict().Str("Name", committer.Name).Str("Email", committer.Email).Time("When", committer.When)).Msg("commit signature set")

	var (
		oid    *git.Oid
		commit = head
	)
	logger.Debug().Str("parent", commit.Id().String()).Msgf("set commit parent to head commit (%s)", commit.Id().String())
	if branchCommit != nil {
		logger.Trace().Msg("branchCommit is non-nil, setting commit parent to branchCommit")
		commit = branchCommit
		logger.Debug().Str("parent", commit.Id().String()).Msgf("set commit parent to branchCommit (%s)", commit.Id().String())
	}

	logger.Trace().Msg("create commit")
	logger = logger.With().Str("ref", fmt.Sprintf("refs/heads/%s", branchName)).Logger()
	if oid, err = repo.CreateCommit(
		fmt.Sprintf("refs/heads/%s", branchName),
		committer,
		committer,
		message,
		tree,
		commit,
	); err != nil {
		logger.Error().Err(err).Msg("error encountered while creating commit")
		return
	}
	logger.Debug().Msgf("successfully created commit (ref: %s)", fmt.Sprintf("refs/heads/%s", branchName))

	logger.Trace().Msg("create capture status")
	cs = &CaptureStatus{
		DuraBranch: branchName,
		CommitHash: oid.String(),
		BaseHash:   head.Id().String(),
	}
	logger.Debug().Dict("cs", zerolog.Dict().Str("DuraBranch", cs.DuraBranch).Str("CommitHash", cs.CommitHash).Str("BaseHash", cs.BaseHash)).Msg("capture status created")

	log.Trace().Msg("leaving Capture")
	return
}

func findHead(repo *git.Repository, branchName string) (head *git.Commit, err error) {
	log.Trace().Msg("entered findHead")
	logger := log.With().Str("repo", repo.Path()).Logger()
	var branch *git.Branch
	logger.Trace().Msgf("looking for branch %s in repository %s", branchName, repo.Path())
	if branch, err = repo.LookupBranch(branchName, git.BranchLocal); err != nil {
		logger.Error().Err(err).Msgf("error encountered while searching for branch %s in repository %s", branchName, repo.Path())
		return
	} else {
		logger.Debug().Msgf("found branch %s in repository %s", branchName, repo.Path())
		var headObj *git.Object
		logger.Trace().Msgf("peel to latest commit on branch %s", branchName)
		if headObj, err = branch.Peel(git.ObjectCommit); err != nil {
			logger.Error().Err(err).Msgf("error encountered while peeling to latest commit on branch %s in repository %s", branchName, repo.Path())
			return
		}
		logger.Debug().Msgf("successfully retrieved latest branch (%s) commit", branchName)
		logger.Trace().Msg("cast commit object to git.Commit")
		if head, err = headObj.AsCommit(); err != nil {
			logger.Error().Err(err).Msg("error encountered while attempting to cast object to git.Commit")
			return
		}
		logger.Debug().Msgf("successfully retrieved head commit (%s) for branch %s in repository %s", head.Id().String(), branchName, repo.Path())
	}
	log.Trace().Msg("leaving findHead")
	return
}

func getGitAuthor(repo *git.Repository) (author string) {
	log.Trace().Msg("entered getGitAuthor")
	logger := log.With().Str("repo", repo.Path()).Logger()
	if config.Commit.Author != nil {
		author = *config.Commit.Author
		logger.Debug().Str("author", author).Msgf("found author set in config (%s)", author)
		return
	}
	if !config.Commit.ExcludeGitConfig {
		var signature *git.Signature
		logger.Trace().Msg("retrieving default signature for repository")
		if signature, err = repo.DefaultSignature(); err == nil {
			author = signature.Name
			logger.Debug().Str("author", author).Msgf("found default signature author for repository (%s)", author)
			return
		} else {
			logger.Error().Err(err).Msgf("error encountered while attempting to retrieve default signature for repository %s", repo.Path())
		}
	}
	author = "dura"
	logger.Debug().Str("author", author).Msgf("using default author (%s)", author)
	log.Trace().Msg("leaving getGitAuthor")
	return
}

func getGitEmail(repo *git.Repository) (email string) {
	log.Trace().Msg("entered getGitEmail")
	logger := log.With().Str("repo", repo.Path()).Logger()
	if config.Commit.Email != nil {
		email = *config.Commit.Email
		logger.Debug().Str("email", email).Msgf("found email set in config (%s)", email)
		return
	}
	if !config.Commit.ExcludeGitConfig {
		var signature *git.Signature
		logger.Trace().Msg("retrieving default signature for repository")
		if signature, err = repo.DefaultSignature(); err == nil {
			email = signature.Email
			logger.Debug().Str("email", email).Msgf("found default signature email for repository (%s)", email)
			return
		} else {
			logger.Error().Err(err).Msgf("error encountered while attempting to retrieve default signature for repository %s", repo.Path())
		}
	}
	email = "dura@github.io"
	logger.Debug().Str("email", email).Msgf("using default email (%s)", email)
	log.Trace().Msg("leaving getGitEmail")
	return
}
