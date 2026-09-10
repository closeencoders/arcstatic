# Arcstatic

<table>
  <tr>
    <td><a href="https://github.com/closeencoders/arcstatic/actions"><img src="https://raw.githubusercontent.com/closeencoders/arcstatic/badges/.badges/main/coverage.svg" alt="coverage"></a></td>
    <td><a href="https://github.com/closeencoders/arcstatic/actions"><img src="https://github.com/closeencoders/arcstatic/actions/workflows/go.yml/badge.svg" alt="Go Status"></a></td>
  </tr>
</table>

Arcstatic is a simple static site generator practice/learning project. While not intended for a production environment, I do personally use it for my own projects.

**(Documentation work in progress, this code is not fully released or completed yet)**

## Example Commands

```
Usage:
  arcstatic [flags]

Flags:
  -b, --build        builds static site from provided resources
  -h, --help         help for arcstatic
  -i, --in string    override default site context current working directory input location, defaults to current location
  -o, --out string   override default site context current working directory out location, defaults to input location
  -p, --port int     port number to file serve the site, if the serve command is not used, this is ignored (default 8000)
  -s, --serve        serve static site from provided resources, currently only for testing
  -v, --verbose      run verbose with debug logs, this will attempt to override any config file settings
```

### Build

The build command converts Markdown and HTML files into static pages using Go templates and Goldmark. Themes, plugin overrides, and custom Goldmark configuration are not yet supported, and assets are copied only when explicitly configured to avoid unnecessary duplication.

Default Location:
```
arcstatic --build
arcstatic -b
```

Specified Location:
```
arcstatic --in ./raw/website --build
arcstatic -i ./raw/website -b
```

### Serve

This currently is ONLY meant for testing locally.

Default Localhost Port 8000
```
arcstatic --in ./rendered/website --serve
arcstatic -i ./rendered/website -s
```

Localhost Port Override 4000:
```
arcstatic --in ./rendered/website --port 4000 --serve
arcstatic -i ./rendered/website -p 4000 -s
```

## Example Configuration

Create an `arcconfig.yml` file in the root directory of your project:

```yaml
# Overrides the default post content directory.
# This avoids having to specify -i /your/location on the command line and allows content to be stored separately from the project.
post_input_dir: /override/input/post/location

# The base URL of your site.
site_url: https://yourdomain.com

# Makes table-of-contents data available for content with detectable headings.
make_toc: true

# Assigns a default type to all rendered content.
# This value can be used for querying and looping.
default_type: blog
```

## Example Feed

The engine uses Goldmark and Go’s built-in templating to loop over entities while remaining simple and broadly compatible.

```html
<div ... >
<!-- Represents the default_type override in example configuration -->
{{range .blog}}
<article>
    <a ... href="{{.Url}}">
        <div>
            {{if .Image}}
            <img src="{{.Image}}" alt="{{.Title}}">
            {{else}}
            <img src="..." ... alt="{{.Title}}">
            {{end}}
        </div>
        <h2 {{.Title}}</h2>
        <div>
            <p>
                {{.Date.Format "Jan 02, 2006"}}
                {{if .Categories}} | {{range $index, $cat := .Categories}}{{if $index}}, {{end}}{{$cat}}{{end}}{{end}}
            </p>
        </div>
        <p>{{if .Description}}{{.Description}}{{end}}</p>
    </a>
</article>
{{end}}
</div>
```

## Example Page Template

```html
<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">

    <!-- Multiline Conditions -->
    {{ if .Metadata.Image }}
    <link rel="preload" as="image" href="{{ .Metadata.Image }}" fetchpriority="high">
    {{ end }}

    <!-- Pull directly from the metadata of the page frontmatter -->
    <title>{{.Metadata.Title}}</title>
    <meta name="description" content="{{.Metadata.MetaDescription}}">

    <meta property="og:title" content="{{.Metadata.Title}}">
    <meta property="og:description" content="{{.Metadata.MetaDescription}}">
    <meta property="og:url" content="{{.CanonicalURL}}">
    <meta property="og:image" content="{{ if .Metadata.Image }}{{ .Metadata.Image }}{{ else }}/assets/img/logo.svg{{ end }}">
    <meta property="og:image:alt" content="{{.Metadata.Title}}">

    <meta name="twitter:title" content="{{.Metadata.Title}}">
    <meta name="twitter:description" content="{{.Metadata.MetaDescription}}">
</head>
<body>
    <!-- Explicit use of specific template by file name -->
    {{template "header.html" .}}
    <main class="wrap">
        {{.Body | safeHTML}}
    </main>
    {{template "footer.html" .}}
</body>
</html>
```
