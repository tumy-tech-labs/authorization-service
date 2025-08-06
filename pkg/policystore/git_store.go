package policystore

import (
    "fmt"
    "os/exec"
    "strings"
)

// GitStore clones and tracks a policy repository.
type GitStore struct {
    repoURL   string
    branch    string
    path      string
    commitSHA string
}

// CloneRepo clones the given repository branch to the local path and returns a store.
func CloneRepo(url, branch, path string) (*GitStore, error) {
    if branch == "" {
        branch = "main"
    }
    cmd := exec.Command("git", "clone", "--depth", "1", "--branch", branch, url, path)
    if out, err := cmd.CombinedOutput(); err != nil {
        return nil, fmt.Errorf("git clone: %v: %s", err, string(out))
    }
    gs := &GitStore{repoURL: url, branch: branch, path: path}
    if err := gs.updateCommit(); err != nil {
        return nil, err
    }
    return gs, nil
}

// PullLatest fetches and resets to the latest commit on the tracked branch.
func (g *GitStore) PullLatest() error {
    fetch := exec.Command("git", "-C", g.path, "fetch", "origin", g.branch)
    if out, err := fetch.CombinedOutput(); err != nil {
        return fmt.Errorf("git fetch: %v: %s", err, string(out))
    }
    reset := exec.Command("git", "-C", g.path, "reset", "--hard", "origin/"+g.branch)
    if out, err := reset.CombinedOutput(); err != nil {
        return fmt.Errorf("git reset: %v: %s", err, string(out))
    }
    return g.updateCommit()
}

// CommitSHA returns the current repository revision.
func (g *GitStore) CommitSHA() string {
    return g.commitSHA
}

func (g *GitStore) updateCommit() error {
    rev := exec.Command("git", "-C", g.path, "rev-parse", "HEAD")
    out, err := rev.Output()
    if err != nil {
        return fmt.Errorf("git rev-parse: %w", err)
    }
    g.commitSHA = strings.TrimSpace(string(out))
    return nil
}

