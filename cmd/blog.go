package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"neocities/blog"
)

// RunBlog handles the blog command and its subcommands.
func RunBlog(args []string) {
	if len(args) < 1 {
		blogUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "init":
		runBlogInit(subargs)
	case "build":
		runBlogBuild(subargs)
	case "new":
		runBlogNew(subargs)
	case "help", "-h", "--help":
		blogUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown blog subcommand: %s\n\n", subcommand)
		blogUsage()
		os.Exit(1)
	}
}

func blogUsage() {
	fmt.Println(`Usage: neocities blog <subcommand> [options] [arguments]

Subcommands:
  init           Create blog directory structure (posts/ and drafts/)
  build          Generate HTML from markdown posts
  new <title>    Create a new blog post with today's date

Options:
  -dir             Source directory (default: current directory)
  -out             Output directory for generated HTML (default: blog/)
  -draft           Create new post as draft (for 'new' subcommand)
  -post-template   Custom HTML template for posts
  -index-template  Custom HTML template for index page

Template Variables (for post-template.html):
  {{.Title}}         Post title
  {{.Slug}}          URL-safe slug
  {{.DateISO}}       Date in ISO format (2006-01-02)
  {{.DateFormatted}} Date in readable format (January 2, 2006)
  {{.Content}}       Rendered HTML content from markdown

Template Variables (for index-template.html):
  {{.Posts}}         Array of posts (use {{range .Posts}}...{{end}})
    {{.Title}}       Post title
    {{.Slug}}        URL-safe slug
    {{.DateISO}}     Date in ISO format
    {{.DateFormatted}} Date in readable format

Examples:
  neocities blog init                    Create posts/ and drafts/ directories
  neocities blog new "Hello World"       Create a new post
  neocities blog new -draft "WIP"        Create a new draft post
  neocities blog build                   Build all posts to blog/
  neocities blog build -out public/blog  Build to custom output directory
  neocities blog build -post-template my-post.html  Use custom post template

By default, the build command looks for post-template.html and index-template.html
in the source directory. If found, they will be used automatically.`)
}

func runBlogInit(args []string) {
	fs := flag.NewFlagSet("blog init", flag.ExitOnError)
	dir := fs.String("dir", ".", "Base directory for blog structure")
	fs.Parse(args)

	if err := blog.Init(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	postsDir := filepath.Join(*dir, "posts")
	draftsDir := filepath.Join(*dir, "drafts")

	fmt.Println("Created blog directory structure:")
	fmt.Printf("  %s/  (published posts)\n", postsDir)
	fmt.Printf("  %s/ (unpublished drafts)\n", draftsDir)
	fmt.Println("\nNext steps:")
	fmt.Println("  neocities blog new \"My First Post\"")
	fmt.Println("  neocities blog build")
}

func runBlogBuild(args []string) {
	fs := flag.NewFlagSet("blog build", flag.ExitOnError)
	dir := fs.String("dir", ".", "Source directory containing posts/")
	out := fs.String("out", "blog", "Output directory for generated HTML")
	postTemplate := fs.String("post-template", "", "Custom post template file")
	indexTemplate := fs.String("index-template", "", "Custom index template file")
	fs.Parse(args)

	postsDir := filepath.Join(*dir, "posts")

	// Discover templates if not explicitly specified
	postTemplatePath := *postTemplate
	indexTemplatePath := *indexTemplate
	if postTemplatePath == "" || indexTemplatePath == "" {
		discoveredPost, discoveredIndex := blog.DiscoverTemplates(*dir)
		if postTemplatePath == "" {
			postTemplatePath = discoveredPost
		}
		if indexTemplatePath == "" {
			indexTemplatePath = discoveredIndex
		}
	}

	cfg := blog.BuildConfig{
		PostsDir:          postsDir,
		OutDir:            *out,
		PostTemplatePath:  postTemplatePath,
		IndexTemplatePath: indexTemplatePath,
	}

	if err := blog.BuildWithConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runBlogNew(args []string) {
	fs := flag.NewFlagSet("blog new", flag.ExitOnError)
	dir := fs.String("dir", ".", "Base directory containing posts/")
	draft := fs.Bool("draft", false, "Create as draft in drafts/ folder")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: title is required")
		fmt.Fprintln(os.Stderr, "Usage: neocities blog new <title>")
		os.Exit(1)
	}

	title := strings.Join(fs.Args(), " ")
	postsDir := filepath.Join(*dir, "posts")

	path, err := blog.NewPost(postsDir, title, *draft)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created: %s\n", path)
	fmt.Println("\nNext steps:")
	fmt.Printf("  Edit %s\n", path)
	fmt.Println("  neocities blog build")
}
