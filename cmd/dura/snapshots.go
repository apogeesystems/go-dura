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
)

var (
	latest = map[string]*git.Oid{}
)

type CaptureStatus struct {
	DuraBranch string `json:"dura_branch"`
	CommitHash string `json:"commit_hash"`
	BaseHash   string `json:"base_hash"`
}

func (cs *CaptureStatus) Display() (err error) {
	_, err = fmt.Printf("dura: %s, commit_hash: %s, base: %s\n", cs.DuraBranch, cs.CommitHash, cs.BaseHash)
	return
}

func IsRepo(path string) (isRepo bool, err error) {
	isRepo = true
	if _, err = git.OpenRepository(path); err != nil {
		isRepo = false
	}
	return
}

func statusCheck(repo *git.Repository) (ok bool, err error) {
	var (
		statusList *git.StatusList
		count      int
	)
	if statusList, err = repo.StatusList(nil); err != nil {
		return
	}
	if count, err = statusList.EntryCount(); err != nil {
		return
	}
	if count > 0 {
		fmt.Printf("Diffs found: %d\n", count)
		ok = true
	} else {
		ok = false
	}
	return
}

func headPeelToCommit(repo *git.Repository) (obj *git.Commit, err error) {
	var (
		head   *git.Reference
		objObj *git.Object
	)
	if head, err = repo.Head(); err != nil {
		return
	}
	if objObj, err = head.Peel(git.ObjectCommit); err != nil {
		return
	}
	if obj, err = objObj.AsCommit(); err != nil {
		return
	}
	return
}

// TODO finish implementation, look over signature/committer portion
func Capture(path string) (cs *CaptureStatus, err error) {
	var (
		repo            *git.Repository
		head            *git.Commit
		message         = "dura auto-backup"
		statusCheckPass bool
		branchName      string
		branchCommit    *git.Commit
	)
	if repo, err = git.OpenRepository(path); err != nil {
		return
	}

	// Get the repo HEAD, peel to the latest Commit as "head"
	if head, err = headPeelToCommit(repo); err != nil {
		return
	}

	if statusCheckPass, err = statusCheck(repo); err != nil || !statusCheckPass {
		if err == nil {
			err = errors.New("repository did not pass status check, i.e. status list was empty for the repository")
		}
		return
	}

	if head.Id() != nil {
		branchName = fmt.Sprintf("dura/%s", head.Id().String())
		fmt.Printf("HEAD ID: %s\n", head.Id().String())
	} else {
		return nil, errors.New("head.Id() was nil")
	}

	if branchCommit, err = findHead(repo, branchName); err != nil {
		fmt.Println(err)
		if _, err = repo.LookupBranch(branchName, git.BranchLocal); err != nil {
			var branch *git.Branch
			if branch, err = repo.CreateBranch(branchName, head, false); err != nil {
				return
			}
			var name string
			if name, err = branch.Name(); err != nil {
				return
			}
			fmt.Printf("Branch Name: %s\n", name)
			fmt.Printf("Branch target: %s\n", branch.Target().String())
			fmt.Printf("Branch Shorthand: %s\n", branch.Shorthand())
			var commitObj *git.Object
			if commitObj, err = branch.Peel(git.ObjectCommit); err != nil {
				return
			}
			if branchCommit, err = commitObj.AsCommit(); err != nil {
				return
			}
			fmt.Printf("Created branch %s...\n", branchName)
			fmt.Printf("Branch commit ID: %s\n", branchCommit.Id().String())
		}
	}

	var index *git.Index
	if index, err = repo.Index(); err != nil {
		return
	}
	if err = index.AddAll([]string{"*"}, git.IndexAddDefault, nil); err != nil {
		return
	}

	var (
		dirtyDiff *git.Diff
		oldTree   *git.Tree
		diffOpts  git.DiffOptions
		deltas    int
	)
	if branchCommit != nil {
		if oldTree, err = branchCommit.Tree(); err != nil {
			fmt.Printf("Failed to retrieve branchCommit tree: %s\n", err.Error())
			if oldTree, err = head.Tree(); err != nil {
				return
			}
		}
	} else {
		if oldTree, err = head.Tree(); err != nil {
			return
		}
	}

	if diffOpts, err = git.DefaultDiffOptions(); err != nil {
		return
	}
	diffOpts.Flags = git.DiffIncludeUntracked
	diffOpts.Pathspec = []string{"*"}
	//fmt.Printf("diffOpts:\n%+v\n\n", diffOpts)

	if dirtyDiff, err = repo.DiffTreeToIndex(
		oldTree,
		index,
		&diffOpts,
	); err != nil {
		return
	}
	if deltas, err = dirtyDiff.NumDeltas(); err != nil || deltas == 0 {
		if err == nil {
			err = errors.New("dirtyDiff has zero deltas")
		}
		return
	}
	//fmt.Printf("Current has %d deltas from index\n", deltas)

	var (
		treeOid *git.Oid
		tree    *git.Tree
	)
	if treeOid, err = index.WriteTreeTo(repo); err != nil {
		return
	}
	if tree, err = repo.LookupTree(treeOid); err != nil {
		return
	}

	var committer *git.Signature
	if committer, err = repo.DefaultSignature(); err != nil {
		return
	}

	var (
		oid    *git.Oid
		commit = head
	)
	if branchCommit != nil {
		fmt.Println("Assigning branchCommit as parent")

		commit = branchCommit
	} else {
		fmt.Println("Assigning head as parent")
	}
	var ok bool
	if oid, ok = latest[path]; ok {
		fmt.Println("Using latest map")

		if commit, err = repo.LookupCommit(oid); err != nil {
			return
		}
	}

	if oid, err = repo.CreateCommit(
		fmt.Sprintf("refs/head/%s", branchName),
		committer,
		committer,
		message,
		tree,
		commit,
	); err != nil {
		fmt.Println("repo.CreateCommit failed")
		fmt.Println(err)
		return
	}

	latest[path] = oid

	fmt.Println("repo.CreateCommit successful")

	cs = &CaptureStatus{
		DuraBranch: branchName,
		CommitHash: oid.String(),
		BaseHash:   head.Id().String(),
	}

	return
}

func findHead(repo *git.Repository, branchName string) (head *git.Commit, err error) {
	var branch *git.Branch
	if branch, err = repo.LookupBranch(branchName, git.BranchLocal); err != nil {
		return
	} else {
		fmt.Printf("Branch %s found...\n", branchName)
		var headObj *git.Object
		if headObj, err = branch.Peel(git.ObjectCommit); err != nil {
			return
		}
		if head, err = headObj.AsCommit(); err != nil {
			return
		}
		fmt.Printf("branchCommit head commit ID: %s\n", head.Id().String())
	}
	return
}

func getGitAuthor(repo *git.Repository) (author string) {
	if config.Commit.Author != nil {
		author = *config.Commit.Author
		return
	}
	if !config.Commit.ExcludeGitConfig {
		var signature *git.Signature
		if signature, err = repo.DefaultSignature(); err == nil {
			author = signature.Name
			return
		}
	}
	return "dura"
}

func getGitEmail(repo *git.Repository) (email string) {
	if config.Commit.Email != nil {
		email = *config.Commit.Email
		return
	}
	if !config.Commit.ExcludeGitConfig {
		var signature *git.Signature
		if signature, err = repo.DefaultSignature(); err == nil {
			email = signature.Email
			return
		}
	}
	return "dura@github.io"
}
