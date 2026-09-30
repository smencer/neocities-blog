package blog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Post represents a blog post.
type Post struct {
	Title    string    // Derived from filename (slugified title)
	Slug     string    // URL-safe identifier
	Date     time.Time // Parsed from filename prefix
	Content  string    // Raw markdown content
	HTML     string    // Rendered HTML
	FilePath string    // Source markdown file
	IsDraft  bool      // In drafts/ folder
}

// filenameRegex matches the expected filename format: YYYY-MM-DD-slug.md
var filenameRegex = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})-(.+)\.md$`)

// ParseFilename extracts date and slug from a filename.
// Expected format: YYYY-MM-DD-slug-title-here.md
func ParseFilename(name string) (time.Time, string, error) {
	matches := filenameRegex.FindStringSubmatch(name)
	if matches == nil {
		return time.Time{}, "", fmt.Errorf("invalid filename format: %s (expected YYYY-MM-DD-slug.md)", name)
	}

	year := matches[1]
	month := matches[2]
	day := matches[3]
	slug := matches[4]

	dateStr := fmt.Sprintf("%s-%s-%s", year, month, day)
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid date in filename: %s", dateStr)
	}

	return date, slug, nil
}

// SlugToTitle converts a slug to a title.
// Example: "hello-world" -> "Hello World"
func SlugToTitle(slug string) string {
	words := strings.Split(slug, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// TitleToSlug converts a title to a URL-safe slug.
// Example: "Hello World!" -> "hello-world"
func TitleToSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

// ReadPost loads and parses a markdown file into a Post.
func ReadPost(path string, isDraft bool) (*Post, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(path)
	date, slug, err := ParseFilename(name)
	if err != nil {
		return nil, err
	}

	return &Post{
		Title:    SlugToTitle(slug),
		Slug:     slug,
		Date:     date,
		Content:  string(content),
		FilePath: path,
		IsDraft:  isDraft,
	}, nil
}

// LoadPosts loads all posts from a directory, sorted by date descending.
func LoadPosts(postsDir string) ([]*Post, error) {
	var posts []*Post

	entries, err := os.ReadDir(postsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return posts, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(postsDir, entry.Name())
		post, err := ReadPost(path, false)
		if err != nil {
			return nil, fmt.Errorf("error reading post %s: %w", entry.Name(), err)
		}
		posts = append(posts, post)
	}

	// Sort by date descending (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}

// RenderPost converts the markdown content to HTML.
func RenderPost(post *Post) error {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	doc := p.Parse([]byte(post.Content))
	post.HTML = string(markdown.Render(doc, renderer))
	return nil
}

// GenerateIndex creates the index.html content listing all posts.
func GenerateIndex(posts []*Post, customTemplatePath string) (string, error) {
	return RenderIndexTemplate(posts, customTemplatePath)
}

// BuildConfig holds configuration for the build process.
type BuildConfig struct {
	PostsDir          string
	OutDir            string
	PostTemplatePath  string // Optional custom template for posts
	IndexTemplatePath string // Optional custom template for index
}

// DiscoverTemplates checks for conventional template files in the given directory.
// Returns paths to post-template.html and index-template.html if they exist.
func DiscoverTemplates(baseDir string) (postPath, indexPath string) {
	postFile := filepath.Join(baseDir, "post-template.html")
	if _, err := os.Stat(postFile); err == nil {
		postPath = postFile
	}

	indexFile := filepath.Join(baseDir, "index-template.html")
	if _, err := os.Stat(indexFile); err == nil {
		indexPath = indexFile
	}

	return postPath, indexPath
}

// Build performs the full build pipeline: load posts, render, and write output.
// This is a convenience wrapper around BuildWithConfig with default settings.
func Build(postsDir, outDir string) error {
	return BuildWithConfig(BuildConfig{
		PostsDir: postsDir,
		OutDir:   outDir,
	})
}

// BuildWithConfig performs the full build pipeline with custom configuration.
func BuildWithConfig(cfg BuildConfig) error {
	// Load all posts
	posts, err := LoadPosts(cfg.PostsDir)
	if err != nil {
		return fmt.Errorf("loading posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found in", cfg.PostsDir)
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Render each post and write to file
	for _, post := range posts {
		if err := RenderPost(post); err != nil {
			return fmt.Errorf("rendering post %s: %w", post.Slug, err)
		}

		html, err := RenderPostTemplate(post, cfg.PostTemplatePath)
		if err != nil {
			return fmt.Errorf("rendering template for %s: %w", post.Slug, err)
		}

		outPath := filepath.Join(cfg.OutDir, post.Slug+".html")
		if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
		fmt.Printf("Generated: %s\n", outPath)
	}

	// Generate index
	indexHTML, err := GenerateIndex(posts, cfg.IndexTemplatePath)
	if err != nil {
		return fmt.Errorf("generating index: %w", err)
	}

	indexPath := filepath.Join(cfg.OutDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(indexHTML), 0644); err != nil {
		return fmt.Errorf("writing index: %w", err)
	}
	fmt.Printf("Generated: %s\n", indexPath)

	fmt.Printf("\nBuilt %d posts to %s\n", len(posts), cfg.OutDir)
	return nil
}

// Init creates the blog directory structure (posts/ and drafts/).
func Init(baseDir string) error {
	postsDir := filepath.Join(baseDir, "posts")
	draftsDir := filepath.Join(baseDir, "drafts")

	if err := os.MkdirAll(postsDir, 0755); err != nil {
		return fmt.Errorf("creating posts directory: %w", err)
	}

	if err := os.MkdirAll(draftsDir, 0755); err != nil {
		return fmt.Errorf("creating drafts directory: %w", err)
	}

	return nil
}

// NewPost creates a new blog post file with the given title.
func NewPost(postsDir, title string, draft bool) (string, error) {
	slug := TitleToSlug(title)
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", date, slug)

	targetDir := postsDir
	if draft {
		targetDir = filepath.Join(filepath.Dir(postsDir), "drafts")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("creating directory: %w", err)
	}

	path := filepath.Join(targetDir, filename)

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("file already exists: %s", path)
	}

	// Create the post with a simple header
	content := fmt.Sprintf("# %s\n\nWrite your post here.\n", title)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return path, nil
}
