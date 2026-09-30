# neocities

A command-line utility for managing [Neocities](https://neocities.org) sites. Upload files, sync local directories, and manage your site from the terminal.

## Installation

```bash
go install neocities@latest
```

Or build from source:

```bash
git clone <repo>
cd neocities
go build -o neocities
```

## Authentication

The utility supports two authentication methods:

1. **Environment variable** (takes precedence):
   ```bash
   export NEOCITIES_API_KEY=your-api-key
   ```

2. **Config file** (`~/.config/neocities/config.json`):
   ```bash
   neocities key
   ```
   This prompts for your username and password, retrieves your API key, and saves it locally.

## Commands

### init

Scaffold a new site with a standard directory structure.

```bash
neocities init mysite
```

Creates:
```
mysite/
├── index.html
├── css/style.css
├── js/main.js
├── images/
└── .neocities.json
```

The `.neocities.json` file configures ignore patterns for sync.

### sync

Synchronize a local directory to your Neocities site. Compares files using SHA1 hashes and only uploads changed or new files.

```bash
neocities sync                  # Sync current directory
neocities sync -dir ./public    # Sync specific directory
neocities sync -delete          # Also delete remote files not present locally
neocities sync -dry-run         # Preview changes without uploading
```

### upload

Upload specific files to your site.

```bash
neocities upload index.html
neocities upload css/style.css js/main.js
neocities upload -path blog/ post.html    # Upload to subdirectory
```

### delete

Delete files from your site. Prompts for confirmation unless `-f` is used.

```bash
neocities delete old-page.html
neocities delete -f temp.html debug.js
```

Note: `index.html` cannot be deleted.

### list

List files on your Neocities site.

```bash
neocities list              # Simple listing
neocities list -l           # Detailed listing with size and date
neocities list images/      # List specific directory
```

### info

Get information about a Neocities site.

```bash
neocities info              # Your site (requires auth)
neocities info kyledrake    # Any public site
```

### key

Retrieve and manage your API key.

```bash
neocities key         # Retrieve and save API key
neocities key -show   # Display current API key
```

### blog

Manage a markdown-based blog. Write posts in markdown, generate HTML, and sync to your site.

```bash
neocities blog init                    # Create posts/ and drafts/ directories
neocities blog new "My First Post"     # Create a new post with today's date
neocities blog new -draft "WIP"        # Create a draft (not published)
neocities blog build                   # Generate HTML to blog/
neocities blog build -out public/blog  # Generate to custom directory
```

**Post filename format:** `YYYY-MM-DD-slug-title.md`

Example: `2024-01-15-hello-world.md` creates a post dated January 15, 2024 with the title "Hello World".

**Directory structure after `blog init`:**
```
site/
├── posts/           # Published markdown files
├── drafts/          # Unpublished posts (ignored by build)
└── blog/            # Generated HTML output
    ├── index.html
    └── hello-world.html
```

The `posts/` and `drafts/` directories are automatically excluded from sync, so only the generated HTML in `blog/` is uploaded.

## Configuration

### .neocities.json

Place this file in your site root to configure sync behavior:

```json
{
  "sitename": "mysite",
  "ignore": [
    ".git",
    ".gitignore",
    ".neocities.json",
    "node_modules",
    ".DS_Store",
    "Thumbs.db",
    "posts",
    "drafts"
  ]
}
```

Files matching ignore patterns are excluded from sync operations.

## Typical Workflow

```bash
# First time setup
neocities key
neocities init mysite
cd mysite

# Edit your site...

# Deploy
neocities sync

# Check status
neocities info
neocities list -l
```

## Blogging Workflow

```bash
# Set up blog in existing site
cd mysite
neocities blog init

# Create a new post
neocities blog new "Hello World"
# Edit posts/2024-01-15-hello-world.md

# Build and deploy
neocities blog build
neocities sync
```

## API Rate Limiting

Neocities recommends limiting recurring site updates to one per minute. The sync command uploads all changed files in a single request to minimize API calls.

## License

MIT
