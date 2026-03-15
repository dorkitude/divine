package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Developer-only commands",
}

var incrementVersionCmd = &cobra.Command{
	Use:   "increment_version",
	Short: "Increment patch version in the VERSION file",
	RunE:  runIncrementVersion,
}

func init() {
	devCmd.AddCommand(incrementVersionCmd)
	rootCmd.AddCommand(devCmd)
}

func runIncrementVersion(cmd *cobra.Command, args []string) error {
	versionPath, err := findVersionFile()
	if err != nil {
		return err
	}

	raw, err := os.ReadFile(versionPath)
	if err != nil {
		return fmt.Errorf("read version file: %w", err)
	}

	current := strings.TrimSpace(string(raw))
	if current == "" {
		return fmt.Errorf("empty version file: %s", versionPath)
	}

	major, minor, patch, err := parseSemVer(current)
	if err != nil {
		return err
	}

	nextVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
	if err := os.WriteFile(versionPath, []byte(nextVersion+"\n"), 0o644); err != nil {
		return fmt.Errorf("write version file: %w", err)
	}

	fmt.Printf("Version updated: %s -> %s (%s)\n", current, nextVersion, versionPath)
	return nil
}

func findVersionFile() (string, error) {
	searchDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not determine working directory: %w", err)
	}

	for i := 0; i < 6; i++ {
		candidate := filepath.Join(searchDir, "VERSION")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		next := filepath.Dir(searchDir)
		if next == searchDir {
			break
		}
		searchDir = next
	}

	return "", fmt.Errorf("VERSION file not found in working tree (checked up to 6 parent directories)")
}

func parseSemVer(versionText string) (int, int, int, error) {
	parts := strings.Split(versionText, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("expected version format MAJOR.MINOR.PATCH, got %q", versionText)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version %q: %w", parts[0], err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version %q: %w", parts[1], err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version %q: %w", parts[2], err)
	}

	return major, minor, patch, nil
}
