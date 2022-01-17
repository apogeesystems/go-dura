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
	git "github.com/libgit2/git2go"
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

	if head, err = headPeelToCommit(repo); err != nil {
		return
	}

	if statusCheckPass, err = statusCheck(repo); err != nil || !statusCheckPass {
		return
	}

	if head.Id() != nil {
		branchName = fmt.Sprintf("dura/%s", head.Id().String())
	} else {
		return nil, errors.New("head.Id() was nil")
	}
	branchCommit, _ = findHead(repo, branchName)

	if _, err = repo.LookupBranch(branchName, git.BranchLocal); err != nil {
		if _, err = repo.CreateBranch(branchName, head, false); err != nil {
			return
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
	if oldTree, err = head.AsTree(); err != nil {
		if branchCommit != nil {
			if oldTree, err = branchCommit.AsTree(); err != nil {
				return
			}
		}
		return
	}
	if diffOpts, err = git.DefaultDiffOptions(); err != nil {
		return
	}
	diffOpts.Flags = git.DiffIncludeUntracked

	if dirtyDiff, err = repo.DiffTreeToIndex(
		oldTree,
		index,
		&diffOpts,
	); err != nil {
		return
	}
	if deltas, err = dirtyDiff.NumDeltas(); err != nil || deltas == 0 {
		return
	}

	var (
		treeOid *git.Oid
		tree    *git.Tree
	)
	if treeOid, err = index.WriteTree(); err != nil {
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
		commit = branchCommit
	}
	if oid, err = repo.CreateCommit(
		fmt.Sprintf("refs/head/%s", branchName),
		committer,
		committer,
		message,
		tree,
		commit,
	); err != nil {
		return
	}

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
		var headObj *git.Object
		if headObj, err = branch.Peel(git.ObjectCommit); err != nil {
			return
		}
		if head, err = headObj.AsCommit(); err != nil {
			return
		}
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
