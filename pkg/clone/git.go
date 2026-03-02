package clone

import "github.com/go-git/go-git/v5"

// openRepo opens a git repository at the given path.
func openRepo(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}
