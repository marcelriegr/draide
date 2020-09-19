package gittools

import (
	"github.com/go-git/go-git/v5"
)

// RepoDetails contains repository info
type RepoDetails struct {
	Branch     string
	CommitHash string
}

// GetRepoDetails return repository info
func GetRepoDetails(path string) (*RepoDetails, error) {
	details := RepoDetails{}

	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}

	headRef, err := repo.Head()
	if err != nil {
		return nil, err
	}

	details.CommitHash = headRef.Hash().String()
	details.Branch = headRef.Name().Short()

	return &details, nil
}
