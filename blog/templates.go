package blog

import (
	"bytes"
	"html/template"
	"os"
)

const postTemplateHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="../css/style.css">
</head>
<body>
    <article>
        <header>
            <h1>{{.Title}}</h1>
            <time datetime="{{.DateISO}}">{{.DateFormatted}}</time>
        </header>
        <main>{{.Content}}</main>
    </article>
    <nav>
        <a href="index.html">&larr; Back to posts</a>
    </nav>
</body>
</html>
`

const indexTemplateHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Blog</title>
    <link rel="stylesheet" href="../css/style.css">
</head>
<body>
    <h1>Blog</h1>
    <ul>
    {{range .Posts}}
        <li><time datetime="{{.DateISO}}">{{.DateFormatted}}</time> - <a href="{{.Slug}}.html">{{.Title}}</a></li>
    {{end}}
    </ul>
</body>
</html>
`

// postTemplateData holds the data for rendering a post template.
type postTemplateData struct {
	Title         string
	Slug          string
	DateISO       string
	DateFormatted string
	Content       template.HTML
}

// indexTemplateData holds the data for rendering the index template.
type indexTemplateData struct {
	Posts []postTemplateData
}

// LoadTemplate reads a template from a file path. If the file doesn't exist,
// returns the default template. If the file exists but can't be read, returns an error.
func LoadTemplate(path, defaultTemplate string) (string, error) {
	if path == "" {
		return defaultTemplate, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultTemplate, nil
		}
		return "", err
	}

	return string(content), nil
}

// RenderPostTemplate renders a post to HTML using the post template.
// If customTemplatePath is provided and the file exists, it will be used instead of the default.
func RenderPostTemplate(post *Post, customTemplatePath string) (string, error) {
	templateStr, err := LoadTemplate(customTemplatePath, postTemplateHTML)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("post").Parse(templateStr)
	if err != nil {
		return "", err
	}

	data := postTemplateData{
		Title:         post.Title,
		Slug:          post.Slug,
		DateISO:       post.Date.Format("2006-01-02"),
		DateFormatted: post.Date.Format("January 2, 2006"),
		Content:       template.HTML(post.HTML),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderIndexTemplate renders the index page listing all posts.
// If customTemplatePath is provided and the file exists, it will be used instead of the default.
func RenderIndexTemplate(posts []*Post, customTemplatePath string) (string, error) {
	templateStr, err := LoadTemplate(customTemplatePath, indexTemplateHTML)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("index").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var postData []postTemplateData
	for _, post := range posts {
		postData = append(postData, postTemplateData{
			Title:         post.Title,
			Slug:          post.Slug,
			DateISO:       post.Date.Format("2006-01-02"),
			DateFormatted: post.Date.Format("January 2, 2006"),
		})
	}

	data := indexTemplateData{Posts: postData}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
