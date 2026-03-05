package install

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/config"
)

const (
	skillsAutoCheckEnvVar  = "ASC_SKILLS_AUTO_CHECK"
	skillsCheckInterval    = 24 * time.Hour
	skillsCheckTimeout     = 8 * time.Second
	skillsCheckedAtLayout  = time.RFC3339
	skillsUpdateMessageFmt = "skills updates may be available. Run 'npx skills update' to refresh installed skills."
)

var (
	loadConfigForSkillsCheck       = config.Load
	persistSkillsCheckedAtForCheck = defaultPersistSkillsCheckedAt
	nowForSkillsCheck              = time.Now
	runSkillsCheckCommand          = defaultRunSkillsCheckCommand
	progressEnabledForCheck        = shared.ProgressEnabled
	lookupSkillsCheckCLI           = exec.LookPath
)

// MaybeCheckForSkillUpdates checks for skill updates once per interval and prints
// a non-blocking stderr notice when updates appear available.
func MaybeCheckForSkillUpdates(ctx context.Context) {
	if !skillsAutoCheckEnabled(strings.TrimSpace(os.Getenv(skillsAutoCheckEnvVar))) {
		return
	}
	if os.Getenv("CI") != "" {
		return
	}
	if !progressEnabledForCheck() {
		return
	}

	cfg, err := loadConfigForSkillsCheck()
	if err != nil {
		// Keep command execution unaffected when config is absent or unreadable.
		return
	}
	if cfg == nil {
		return
	}

	now := nowForSkillsCheck().UTC()
	if !shouldRunSkillsCheck(now, cfg.SkillsCheckedAt) {
		return
	}

	checkCtx, cancel := context.WithTimeout(ctx, skillsCheckTimeout)
	defer cancel()

	output, runErr := runSkillsCheckCommand(checkCtx)

	// Persist the check timestamp regardless of check result to avoid repeated
	// invocation overhead during failures.
	_ = persistSkillsCheckedAtForCheck(now.Format(skillsCheckedAtLayout))

	if runErr != nil {
		return
	}
	if !skillsOutputHasUpdates(output) {
		return
	}

	fmt.Fprintln(os.Stderr, skillsUpdateMessageFmt)
}

func shouldRunSkillsCheck(now time.Time, lastCheckedAt string) bool {
	lastCheckedAt = strings.TrimSpace(lastCheckedAt)
	if lastCheckedAt == "" {
		return true
	}

	lastChecked, err := time.Parse(skillsCheckedAtLayout, lastCheckedAt)
	if err != nil {
		return true
	}
	return now.Sub(lastChecked.UTC()) >= skillsCheckInterval
}

func skillsAutoCheckEnabled(value string) bool {
	if value == "" {
		return true
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return true
	}
}

func skillsOutputHasUpdates(output string) bool {
	normalized := strings.ToLower(strings.TrimSpace(output))
	if normalized == "" {
		return false
	}

	switch {
	case strings.Contains(normalized, "all skills are up to date"):
		return false
	case strings.Contains(normalized, "no updates available"), strings.Contains(normalized, "no update available"):
		return false
	case strings.Contains(normalized, "update available"):
		return true
	case strings.Contains(normalized, "updates available"):
		return true
	default:
		return false
	}
}

func defaultRunSkillsCheckCommand(ctx context.Context) (string, error) {
	skillsPath, err := lookupSkillsCheckCLI("skills")
	if err != nil {
		return "", nil
	}
	if shouldSkipProjectLocalSkillsBinary(skillsPath) {
		return "", nil
	}

	cmd := exec.CommandContext(ctx, skillsPath, "check")
	// Avoid resolving project-local node_modules in the current repository.
	cmd.Dir = skillsCheckWorkingDirectory()
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined

	if err := cmd.Run(); err != nil {
		return combined.String(), err
	}
	return combined.String(), nil
}

func skillsCheckWorkingDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.TrimSpace(homeDir) != "" {
		return homeDir
	}
	return os.TempDir()
}

func shouldSkipProjectLocalSkillsBinary(binaryPath string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	repoRoot := cwd
	if root, ok := detectRepoRoot(cwd); ok {
		repoRoot = root
	}

	resolvedBinary := resolvePathForComparison(binaryPath)
	resolvedRoot := resolvePathForComparison(repoRoot)
	return isPathWithin(resolvedBinary, resolvedRoot)
}

func detectRepoRoot(start string) (string, bool) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func resolvePathForComparison(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		return resolved
	}
	return absPath
}

func isPathWithin(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)))
}

func defaultPersistSkillsCheckedAt(timestamp string) error {
	path, err := config.Path()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	if doc == nil {
		doc = map[string]json.RawMessage{}
	}

	encoded, err := json.Marshal(strings.TrimSpace(timestamp))
	if err != nil {
		return err
	}
	doc["skills_checked_at"] = encoded

	updated, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, updated, 0o600)
}
