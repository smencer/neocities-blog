package site

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"neocities/api"
)

const siteConfigFile = ".neocities.json"

// SiteConfig holds local site configuration.
type SiteConfig struct {
	Sitename string   `json:"sitename,omitempty"`
	Ignore   []string `json:"ignore,omitempty"`
}

// DefaultIgnore contains default patterns to ignore during sync.
var DefaultIgnore = []string{
	".git",
	".gitignore",
	".neocities.json",
	"node_modules",
	".DS_Store",
	"Thumbs.db",
	"posts",
	"drafts",
}

// LocalFile represents a local file with its metadata.
type LocalFile struct {
	Path     string
	FullPath string
	SHA1Hash string
	Size     int64
}

// Init scaffolds a new site in the given directory.
func Init(dir string) error {
	// Create directory structure
	dirs := []string{
		dir,
		filepath.Join(dir, "css"),
		filepath.Join(dir, "js"),
		filepath.Join(dir, "images"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	// Create index.html
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My Neocities Site</title>
    <link rel="stylesheet" href="css/style.css">
</head>
<body>
    <header>
        <h1>Welcome to My Site</h1>
    </header>
    <main>
        <p>This is your new Neocities site. Edit this file to get started!</p>
    </main>
    <footer>
        <p>Hosted on <a href="https://neocities.org">Neocities</a></p>
    </footer>
    <script src="js/main.js"></script>
</body>
</html>
`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(indexHTML), 0644); err != nil {
		return err
	}

	// Create style.css
	styleCSS := `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, sans-serif;
    line-height: 1.6;
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
}

header {
    margin-bottom: 2rem;
}

main {
    margin-bottom: 2rem;
}

footer {
    color: #666;
    font-size: 0.9rem;
}

a {
    color: #0066cc;
}
`
	if err := os.WriteFile(filepath.Join(dir, "css", "style.css"), []byte(styleCSS), 0644); err != nil {
		return err
	}

	// Create main.js
	mainJS := `// Your JavaScript code here
console.log('Site loaded!');
`
	if err := os.WriteFile(filepath.Join(dir, "js", "main.js"), []byte(mainJS), 0644); err != nil {
		return err
	}

	// Create .neocities.json
	cfg := SiteConfig{
		Ignore: DefaultIgnore,
	}
	cfgData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, siteConfigFile), cfgData, 0644); err != nil {
		return err
	}

	return nil
}

// LoadSiteConfig loads the site configuration from the given directory.
func LoadSiteConfig(dir string) (*SiteConfig, error) {
	path := filepath.Join(dir, siteConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &SiteConfig{Ignore: DefaultIgnore}, nil
		}
		return nil, err
	}

	var cfg SiteConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Merge with defaults
	if len(cfg.Ignore) == 0 {
		cfg.Ignore = DefaultIgnore
	}

	return &cfg, nil
}

// shouldIgnore checks if a path should be ignored based on patterns.
func shouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Check if any component of the path matches the pattern
		parts := strings.Split(path, string(filepath.Separator))
		for _, part := range parts {
			if matched, _ := filepath.Match(pattern, part); matched {
				return true
			}
		}
	}
	return false
}

// computeSHA1 computes the SHA1 hash of a file.
func computeSHA1(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ListLocalFiles walks the directory and returns all files with their hashes.
func ListLocalFiles(dir string, ignorePatterns []string) ([]LocalFile, error) {
	var files []LocalFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Skip if it matches ignore patterns
		if shouldIgnore(relPath, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Compute hash
		hash, err := computeSHA1(path)
		if err != nil {
			return err
		}

		files = append(files, LocalFile{
			Path:     relPath,
			FullPath: path,
			SHA1Hash: hash,
			Size:     info.Size(),
		})

		return nil
	})

	return files, err
}

// SyncPlan represents the actions needed to sync local to remote.
type SyncPlan struct {
	Upload []LocalFile // Files to upload (new or changed)
	Delete []string    // Files to delete from remote
}

// ComputeSyncPlan compares local files with remote and determines what needs to be synced.
func ComputeSyncPlan(localFiles []LocalFile, remoteFiles []api.FileInfo, deleteRemoved bool) *SyncPlan {
	plan := &SyncPlan{}

	// Build map of remote files by path
	remoteMap := make(map[string]api.FileInfo)
	for _, rf := range remoteFiles {
		if !rf.IsDir {
			remoteMap[rf.Path] = rf
		}
	}

	// Build map of local files by path
	localMap := make(map[string]LocalFile)
	for _, lf := range localFiles {
		localMap[lf.Path] = lf
	}

	// Check each local file
	for _, lf := range localFiles {
		rf, exists := remoteMap[lf.Path]
		if !exists {
			// File doesn't exist remotely, needs upload
			plan.Upload = append(plan.Upload, lf)
		} else if rf.SHA1Hash != lf.SHA1Hash {
			// File exists but hash is different, needs upload
			plan.Upload = append(plan.Upload, lf)
		}
	}

	// Check for files to delete (exist remotely but not locally)
	if deleteRemoved {
		for path, rf := range remoteMap {
			if _, exists := localMap[path]; !exists && !rf.IsDir {
				// Don't delete index.html
				if path != "index.html" {
					plan.Delete = append(plan.Delete, path)
				}
			}
		}
	}

	return plan
}
