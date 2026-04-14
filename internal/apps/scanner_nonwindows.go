//go:build !windows

package apps

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type AppEntry struct {
	Name    string
	Path    string
	LnkPath string
	IsLnk   bool
}

type Scanner struct {
	mu   sync.RWMutex
	apps []AppEntry
}

func NewScanner() *Scanner {
	s := &Scanner{}
	s.Scan()
	return s
}

func (s *Scanner) Scan() {
	var found []AppEntry

	found = append(found, scanDesktopApps()...)
	found = append(found, scanPathAppsNonWindows()...)
	found = deduplicate(found)

	s.mu.Lock()
	s.apps = found
	s.mu.Unlock()
}

func (s *Scanner) Apps() []AppEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]AppEntry, len(s.apps))
	copy(result, s.apps)
	return result
}

func (s *Scanner) Names() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, len(s.apps))
	for i, app := range s.apps {
		names[i] = app.Name
	}
	return names
}

func scanDesktopApps() []AppEntry {
	var dirs []string
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		dirs = []string{"/Applications", filepath.Join(home, "Applications")}
	case "linux":
		dirs = []string{"/usr/share/applications", filepath.Join(home, ".local", "share", "applications")}
	}

	var results []AppEntry
	for _, root := range dirs {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			lower := strings.ToLower(name)
			path := filepath.Join(root, name)

			if runtime.GOOS == "darwin" && strings.HasSuffix(lower, ".app") {
				results = append(results, AppEntry{Name: strings.TrimSuffix(name, ".app"), Path: path})
			}
			if runtime.GOOS == "linux" && strings.HasSuffix(lower, ".desktop") {
				appName := desktopAppName(path)
				if appName == "" {
					appName = strings.TrimSuffix(name, ".desktop")
				}
				results = append(results, AppEntry{Name: appName, Path: path})
			}
		}
	}
	return results
}

func scanPathAppsNonWindows() []AppEntry {
	var results []AppEntry
	seen := make(map[string]bool)

	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			full := filepath.Join(dir, name)
			if seen[strings.ToLower(full)] {
				continue
			}
			if !isExecutable(full) {
				continue
			}
			results = append(results, AppEntry{Name: name, Path: full})
			seen[strings.ToLower(full)] = true
		}
	}

	return results
}

func isExecutable(path string) bool {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		info, err := os.Stat(path)
		if err != nil {
			return false
		}
		return info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
	}
	_, err := exec.LookPath(path)
	return err == nil
}

func deduplicate(apps []AppEntry) []AppEntry {
	seen := make(map[string]bool)
	var result []AppEntry
	for _, app := range apps {
		key := strings.ToLower(app.Name)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, app)
	}
	return result
}

// desktopAppName reads the Name= field from a .desktop file and returns it.
// Locale-specific variants (Name[xx]=) are ignored; only the unlocalized Name= line is used.
// Returns an empty string if the file cannot be read or no Name= entry is found.
func desktopAppName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip locale variants like Name[de]= or Name[zh_CN]=
		if strings.HasPrefix(line, "Name=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Name="))
		}
	}
	return ""
}
