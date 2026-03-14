package repository

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	DetectionModeGit             = "git"
	DetectionModeOptimusCtxState = "optimusctx-state"
)

var ErrRepositoryNotFound = errors.New("repository root not found")

type Locator struct{}

type RepositoryRoot struct {
	RootPath      string
	DetectionMode string
	Fingerprint   RepositoryFingerprint
}

type RepositoryFingerprint struct {
	RootPath      string
	GitCommonDir  string
	GitHeadRef    string
	GitHeadCommit string
}

func NewLocator() Locator {
	return Locator{}
}

func (Locator) Resolve(startPath string) (RepositoryRoot, error) {
	canonicalStart, err := canonicalPath(startPath)
	if err != nil {
		return RepositoryRoot{}, fmt.Errorf("canonicalize start path: %w", err)
	}

	if gitRoot, err := resolveGitRoot(canonicalStart); err == nil {
		return gitRoot, nil
	}

	if stateRoot, err := resolveOptimusCtxStateRoot(canonicalStart); err == nil {
		return stateRoot, nil
	}

	return RepositoryRoot{}, ErrRepositoryNotFound
}

func ResolveRepositoryRoot(startPath string) (RepositoryRoot, error) {
	return NewLocator().Resolve(startPath)
}

func canonicalPath(path string) (string, error) {
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		return filepath.Clean(absPath), nil
	}

	return filepath.Clean(resolved), nil
}

func resolveGitRoot(startPath string) (RepositoryRoot, error) {
	rootPath, err := runGit(startPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return RepositoryRoot{}, err
	}

	canonicalRoot, err := canonicalPath(rootPath)
	if err != nil {
		return RepositoryRoot{}, err
	}

	gitCommonDir, err := runGit(canonicalRoot, "rev-parse", "--git-common-dir")
	if err == nil && !filepath.IsAbs(gitCommonDir) {
		gitCommonDir = filepath.Join(canonicalRoot, gitCommonDir)
	}
	if gitCommonDir != "" {
		gitCommonDir, _ = canonicalPath(gitCommonDir)
	}

	headRef, err := runGit(canonicalRoot, "symbolic-ref", "-q", "--short", "HEAD")
	if err != nil {
		headRef = ""
	}

	headCommit, err := runGit(canonicalRoot, "rev-parse", "HEAD")
	if err != nil {
		headCommit = ""
	}

	return RepositoryRoot{
		RootPath:      canonicalRoot,
		DetectionMode: DetectionModeGit,
		Fingerprint: RepositoryFingerprint{
			RootPath:      canonicalRoot,
			GitCommonDir:  gitCommonDir,
			GitHeadRef:    headRef,
			GitHeadCommit: headCommit,
		},
	}, nil
}

func resolveOptimusCtxStateRoot(startPath string) (RepositoryRoot, error) {
	current := startPath
	for {
		stateDir := filepath.Join(current, ".optimusctx")
		info, err := os.Stat(stateDir)
		if err == nil && info.IsDir() {
			return RepositoryRoot{
				RootPath:      current,
				DetectionMode: DetectionModeOptimusCtxState,
				Fingerprint: RepositoryFingerprint{
					RootPath: current,
				},
			}, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return RepositoryRoot{}, err
		}

		parent := filepath.Dir(current)
		if parent == current {
			return RepositoryRoot{}, ErrRepositoryNotFound
		}
		current = parent
	}
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
