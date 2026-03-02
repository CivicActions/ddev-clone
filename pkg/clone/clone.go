package clone

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/civicactions/ddev-clone/pkg/code"
	"github.com/civicactions/ddev-clone/pkg/ddev"
	"github.com/civicactions/ddev-clone/pkg/docker"
	"github.com/civicactions/ddev-clone/pkg/volume"
	"github.com/go-git/go-git/v5"
)

// cloneNameRegex validates clone names: alphanumeric + hyphens, no leading/trailing hyphens.
var cloneNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)

// Create creates a new DDEV project clone with code isolation and volume duplication.
func Create(opts CreateOptions) error {
	// 1. Derive clone project name and path
	cloneProjectName := getCloneProjectName(opts.SourceProjectName, opts.CloneName)
	clonePath := filepath.Join(filepath.Dir(opts.SourcePath), cloneProjectName)

	fmt.Printf("Creating clone '%s' of project '%s'...\n", opts.CloneName, opts.SourceProjectName)

	// 2. Validate clone name
	if !cloneNameRegex.MatchString(opts.CloneName) {
		return fmt.Errorf("invalid clone name %q: must contain only alphanumeric characters and hyphens", opts.CloneName)
	}

	// 3. Check for existing DDEV project with same name
	if _, err := ddev.Describe(cloneProjectName); err == nil {
		return fmt.Errorf("a project named '%s' already exists. Use 'ddev clone remove %s' to remove it first", cloneProjectName, opts.CloneName)
	}

	// 4. Check for existing directory
	if _, err := os.Stat(clonePath); err == nil {
		return fmt.Errorf("directory %s already exists. Remove it or choose a different clone name", clonePath)
	}

	// 5. Determine code strategy
	strategy, err := resolveCodeStrategy(opts.SourcePath, opts.CodeStrategyName)
	if err != nil {
		return fmt.Errorf("failed to determine code strategy: %w", err)
	}
	fmt.Printf("  Using code strategy: %s\n", strategy.Name())

	// 6. Check for dirty working directory
	if dirty, files, err := strategy.IsDirty(opts.SourcePath); err == nil && dirty {
		fmt.Printf("  Warning: source has uncommitted changes (%d files modified)\n", len(files))
		for _, f := range files {
			fmt.Printf("    %s\n", f)
		}
	}

	// 7. Create code copy
	branch := opts.Branch
	if branch == "" {
		branch = opts.CloneName
	}
	fmt.Printf("  Creating %s at %s...\n", strategy.Name(), clonePath)
	if err := strategy.Create(opts.SourcePath, clonePath, branch); err != nil {
		return fmt.Errorf("failed to create code copy: %w", err)
	}

	// Track what was created for rollback
	var createdVolumes []string
	codeCreated := true

	// Rollback function for error cleanup
	rollback := func() {
		ctx := context.Background()
		for _, vol := range createdVolumes {
			_ = docker.RemoveVolume(ctx, vol)
		}
		if codeCreated {
			_ = strategy.Remove(opts.SourcePath, clonePath, true)
		}
	}

	// 7b. Copy .ddev/ directory from source if not present in clone.
	// This is necessary for worktree clones where .ddev/ is untracked/gitignored
	// and therefore not included in the worktree checkout.
	cloneDdevDir := filepath.Join(clonePath, ".ddev")
	cloneConfigYAML := filepath.Join(cloneDdevDir, "config.yaml")
	if _, err := os.Stat(cloneConfigYAML); os.IsNotExist(err) {
		sourceDdevDir := filepath.Join(opts.SourcePath, ".ddev")
		if _, err := os.Stat(sourceDdevDir); err == nil {
			fmt.Println("  Copying .ddev/ directory from source...")
			if err := copyDdevDir(sourceDdevDir, cloneDdevDir); err != nil {
				rollback()
				return fmt.Errorf("failed to copy .ddev directory: %w", err)
			}
		}
	}

	// 8. Write .ddev/config.local.yaml with clone project name and auto-assigned ports
	// to avoid port conflicts with the source or other clones.
	// Uses yaml.v3 Node API to preserve any existing comments and structure.
	if err := os.MkdirAll(cloneDdevDir, 0755); err != nil {
		rollback()
		return fmt.Errorf("failed to create .ddev directory: %w", err)
	}

	// Allocate unique free ports to avoid conflicts with source or other clones.
	freePorts, err := getFreePorts(4)
	if err != nil {
		rollback()
		return fmt.Errorf("failed to allocate free ports: %w", err)
	}

	configPath := filepath.Join(cloneDdevDir, "config.local.yaml")
	fmt.Printf("  Writing .ddev/config.local.yaml with project name '%s'...\n", cloneProjectName)
	if err := writeConfigLocalYAML(configPath, configLocalOverrides{
		Name:              cloneProjectName,
		HostDBPort:        freePorts[0],
		HostHTTPSPort:     freePorts[1],
		HostWebserverPort: freePorts[2],
		HostMailpitPort:   freePorts[3],
	}); err != nil {
		rollback()
		return fmt.Errorf("failed to write config.local.yaml: %w", err)
	}

	// 9. Get source project details for volume enumeration
	desc, err := ddev.Describe(opts.SourceProjectName)
	if err != nil {
		rollback()
		return fmt.Errorf("failed to describe source project: %w", err)
	}

	// 10. Determine volumes to clone
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	volumePairs := getVolumesToClone(ctx, opts.SourceProjectName, cloneProjectName, desc.DatabaseType, desc.MutagenEnabled)

	// 11. Select volume cloner
	volCloner := opts.VolumeCloner
	if volCloner == nil {
		volCloner = volume.NewTarCopyCloner()
	}

	// 12. Stop source DB for consistent copy
	sourceRunning, _ := ddev.IsRunning(opts.SourceProjectName)
	if sourceRunning {
		fmt.Println("  Stopping database container for consistent copy...")
		if err := docker.StopDBService(opts.SourceProjectName); err != nil {
			rollback()
			return fmt.Errorf("failed to stop source database: %w", err)
		}
		// Ensure DB is restarted regardless of outcome
		defer func() {
			fmt.Println("  Resuming database container...")
			if startErr := docker.StartDBService(opts.SourceProjectName); startErr != nil {
				fmt.Printf("  Warning: failed to restart source database: %v\n", startErr)
			}
		}()
	}

	// 13. Clone each volume
	for _, vp := range volumePairs {
		fmt.Printf("  Cloning volume %s -> %s...\n", vp.Source, vp.Target)
		if err := volCloner.CloneVolume(ctx, vp.Source, vp.Target); err != nil {
			rollback()
			return fmt.Errorf("failed to clone volume %s: %w", vp.Source, err)
		}
		createdVolumes = append(createdVolumes, vp.Target)
	}

	// 14. Start clone unless --no-start
	if !opts.NoStart {
		fmt.Printf("  Starting clone project %s...\n", cloneProjectName)
		if err := ddev.StartProjectByPath(clonePath); err != nil {
			rollback()
			return fmt.Errorf("failed to start clone project: %w", err)
		}
	}

	fmt.Printf("Successfully created clone '%s'\n", opts.CloneName)
	if !opts.NoStart {
		fmt.Printf("Clone URL: https://%s.ddev.site\n", cloneProjectName)
	}

	return nil
}

// List returns all clones discovered for the given source project.
func List(sourceProjectName, sourcePath string) ([]CloneInfo, error) {
	return DiscoverClones(sourceProjectName, sourcePath)
}

// Remove removes a clone and cleans up all its resources.
func Remove(opts RemoveOptions) error {
	cloneProjectName := getCloneProjectName(opts.SourceProjectName, opts.CloneName)

	fmt.Printf("Removing clone '%s' of project '%s'...\n", opts.CloneName, opts.SourceProjectName)

	// Discover the clone to get its metadata
	clones, err := DiscoverClones(opts.SourceProjectName, opts.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to discover clones: %w", err)
	}

	var targetClone *CloneInfo
	for i, c := range clones {
		if c.CloneName == opts.CloneName {
			targetClone = &clones[i]
			break
		}
	}

	// Determine strategy
	var strategy code.CodeStrategy
	clonePath := ""

	if targetClone != nil {
		clonePath = targetClone.ClonePath
		if targetClone.CodeStrategyType == "worktree" {
			strategy = code.NewGitWorktreeStrategy()
		} else {
			strategy = code.NewDirectoryCopyStrategy()
		}
	} else {
		// Clone not found in discovery, but try to delete DDEV project anyway
		fmt.Printf("  Warning: clone '%s' not found in discovery, attempting cleanup...\n", opts.CloneName)
	}

	// Check dirty state if clone path exists
	if strategy != nil && clonePath != "" {
		if _, err := os.Stat(clonePath); err == nil {
			dirty, files, dirtyErr := strategy.IsDirty(clonePath)
			if dirtyErr == nil && dirty && !opts.Force {
				if opts.ConfirmFunc != nil {
					if !opts.ConfirmFunc(opts.CloneName, files) {
						return nil // User cancelled
					}
				} else {
					return fmt.Errorf("clone '%s' has uncommitted changes. Use --force to skip confirmation", opts.CloneName)
				}
			}
		} else {
			fmt.Println("  Warning: clone directory not found, cleaning up remaining resources...")
		}
	}

	// Delete DDEV project
	fmt.Printf("  Deleting DDEV project %s...\n", cloneProjectName)
	if err := ddev.DeleteProject(cloneProjectName); err != nil {
		// Don't fail entirely — try code removal too
		fmt.Printf("  Warning: failed to delete DDEV project: %v\n", err)
	}

	// Remove code copy
	if strategy != nil && clonePath != "" {
		if _, err := os.Stat(clonePath); err == nil {
			stratName := "code copy"
			if targetClone != nil {
				stratName = targetClone.CodeStrategyType
			}
			fmt.Printf("  Removing %s...\n", stratName)
			if err := strategy.Remove(opts.SourcePath, clonePath, opts.Force); err != nil {
				return fmt.Errorf("failed to remove code copy: %w", err)
			}
		}
	}

	fmt.Printf("Successfully removed clone '%s'\n", opts.CloneName)
	return nil
}

// Prune removes stale clone references whose directories no longer exist.
func Prune(opts PruneOptions) ([]string, error) {
	clones, err := DiscoverClones(opts.SourceProjectName, opts.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to discover clones: %w", err)
	}

	var pruned []string
	for _, c := range clones {
		if _, err := os.Stat(c.ClonePath); err == nil {
			continue // Directory exists, not stale
		}

		if opts.DryRun {
			fmt.Printf("Would remove stale clone '%s' (directory %s no longer exists)\n", c.CloneName, c.ClonePath)
		} else {
			fmt.Printf("Pruning stale clone '%s'...\n", c.CloneName)
			fmt.Printf("  Deleting DDEV project %s...\n", c.ProjectName)
			if err := ddev.DeleteProject(c.ProjectName); err != nil {
				fmt.Printf("  Warning: failed to delete DDEV project: %v\n", err)
			}
		}
		pruned = append(pruned, c.CloneName)
	}

	// Prune git worktree references if source is a git repo
	if !opts.DryRun {
		if _, err := git.PlainOpen(opts.SourcePath); err == nil {
			fmt.Println("  Pruning git worktree references...")
			worktreeStrategy := code.NewGitWorktreeStrategy()
			// Use a dummy remove just to trigger prune
			_ = pruneGitWorktrees(opts.SourcePath, worktreeStrategy)
		}
	}

	return pruned, nil
}

// pruneGitWorktrees runs git worktree prune on the source repo.
func pruneGitWorktrees(sourcePath string, _ *code.GitWorktreeStrategy) error {
	cmd := exec.Command("git", "-C", sourcePath, "worktree", "prune")
	_, err := cmd.CombinedOutput()
	return err
}

// resolveCodeStrategy determines the code strategy based on user preference and git availability.
func resolveCodeStrategy(sourcePath, strategyName string) (code.CodeStrategy, error) {
	switch strings.ToLower(strategyName) {
	case "copy":
		return code.NewDirectoryCopyStrategy(), nil
	case "worktree", "":
		// Try to detect git repo
		_, err := git.PlainOpen(sourcePath)
		if err != nil {
			if strategyName == "worktree" {
				return nil, fmt.Errorf("source path %q is not a git repository", sourcePath)
			}
			// Auto-detect: fall back to copy
			fmt.Println("  No git repository detected. Using code strategy: copy")
			return code.NewDirectoryCopyStrategy(), nil
		}
		return code.NewGitWorktreeStrategy(), nil
	default:
		return nil, fmt.Errorf("unknown code strategy: %q (valid: worktree, copy)", strategyName)
	}
}

// getVolumesToClone builds the list of volume pairs to clone.
func getVolumesToClone(ctx context.Context, sourceProject, cloneProject, dbType string, mutagenEnabled bool) []VolumePair {
	var pairs []VolumePair

	// DB volume (required — always clone)
	switch dbType {
	case "mariadb", "mysql":
		pairs = append(pairs, VolumePair{
			Source: sourceProject + "-mariadb",
			Target: cloneProject + "-mariadb",
		})
	case "postgres":
		pairs = append(pairs, VolumePair{
			Source: sourceProject + "-postgres",
			Target: cloneProject + "-postgres",
		})
	}

	// Optional volumes — clone if they exist
	optionals := []VolumePair{
		{Source: "ddev-" + sourceProject + "-snapshots", Target: "ddev-" + cloneProject + "-snapshots"},
		{Source: sourceProject + "-ddev-config", Target: cloneProject + "-ddev-config"},
	}
	if mutagenEnabled {
		optionals = append(optionals, VolumePair{
			Source: sourceProject + "_project_mutagen",
			Target: cloneProject + "_project_mutagen",
		})
	}
	for _, vp := range optionals {
		if docker.VolumeExists(ctx, vp.Source) {
			pairs = append(pairs, vp)
		}
	}

	// Custom service volumes via Docker compose label
	composeProject := docker.GetComposeProjectName(sourceProject)
	cloneCompose := docker.GetComposeProjectName(cloneProject)
	customVols, err := docker.ListProjectVolumes(ctx, composeProject)
	if err == nil {
		knownSources := make(map[string]bool)
		for _, p := range pairs {
			knownSources[p.Source] = true
		}
		for _, vol := range customVols {
			if knownSources[vol] {
				continue
			}
			// Map source volume name to clone volume name
			cloneVol := strings.Replace(vol, composeProject, cloneCompose, 1)
			if cloneVol == vol {
				// If no match on compose project prefix, try source project name
				cloneVol = strings.Replace(vol, sourceProject, cloneProject, 1)
			}
			pairs = append(pairs, VolumePair{Source: vol, Target: cloneVol})
		}
	}

	return pairs
}

// copyDdevDir copies the .ddev/ directory from source to clone, skipping
// runtime-generated files that are project-specific (docker-compose files,
// image build caches, etc.). This is needed for worktree clones where .ddev/
// is typically untracked/gitignored.
func copyDdevDir(srcDdevDir, dstDdevDir string) error {
	// Files/dirs to skip — these are generated per-project by DDEV
	skipPrefixes := map[string]bool{
		".ddev-docker-compose": true,
		".dbimageBuild":        true,
		".webimageBuild":       true,
		".sshimageBuild":       true,
		".importdb":            true,
		".downloads":           true,
		"db_snapshots":         true,
		"config.local.yaml":    true,
		"config.local.yml":     true,
		"sequelpro.spf":        true,
		"traefik":              true,
	}

	entries, err := os.ReadDir(srcDdevDir)
	if err != nil {
		return fmt.Errorf("failed to read source .ddev directory: %w", err)
	}

	if err := os.MkdirAll(dstDdevDir, 0755); err != nil {
		return fmt.Errorf("failed to create clone .ddev directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip runtime/generated files
		if skipPrefixes[name] {
			continue
		}
		// Also skip any dotfile starting with .ddev-docker-compose
		if strings.HasPrefix(name, ".ddev-docker-compose") {
			continue
		}
		// Skip hidden download dirs
		if strings.HasPrefix(name, ".") && strings.Contains(name, "downloads") {
			continue
		}

		src := filepath.Join(srcDdevDir, name)
		dst := filepath.Join(dstDdevDir, name)

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.IsDir() {
			// Recursive copy using os commands for simplicity and permission preservation
			cmd := exec.Command("cp", "-a", src, dst)
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to copy %s: %s: %w", name, string(out), err)
			}
		} else {
			// Copy single file
			data, err := os.ReadFile(src)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", name, err)
			}
			if err := os.WriteFile(dst, data, info.Mode()); err != nil {
				return fmt.Errorf("failed to write %s: %w", name, err)
			}
		}
	}

	return nil
}
