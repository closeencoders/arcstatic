# Arcstatic
[![coverage](https://raw.githubusercontent.com/closeencoders/arcstatic/badges/.badges/main/coverage.svg)](https://github.com/closeencoders/arcstatic/actions) [![Go Status](https://github.com/closeencoders/arcstatic/actions/workflows/go.yml/badge.svg)](https://github.com/closeencoders/arcstatic/actions)

Arcstatic is a simple static site generator practice/learning project. While not intended for a production environment, I do personally use it for my own projects.

(documentation work in progress, this code is not fully released or completed yet)

## Example Commands
```
Usage:
  arcstatic [flags]

Flags:
  -b, --build        builds static site from provided resources
  -h, --help         help for arcstatic
  -i, --in string    override default site context current working directory input location, defaults to current location
  -o, --out string   override default site context current working directory out location, defaults to current location
  -p, --port int     port number to file serve the site, if the serve command is not used, this is ignored (default 8000)
  -s, --serve        serve static site from provided resources, currently only for testing
  -v, --verbose      run verbose with debug logs, this will attempt to override any config file settings
```

### Build

The build command takes a directory of Markdown and HTML files and generates static HTML pages using Go templates and Goldmark. I don't have any way to add themes, override plugins, or Goldmark configuration (Yet).

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

Default Localhost Port 8080
```
arcstatic --in ./rendered/website --serve
arcstatic -i ./rendered/website -s
```

Localhost Port Override 4000:
```
arcstatic --in ./rendered/website --port 4000 --serve
arcstatic -i ./rendered/website -p 4000 -s
```

## Example Feed
```html
<div ... >
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

## Example Config

Currently just a config.yml file in the root of the project.

```yaml
post_input_dir: /override/input/post/location
site_url: https://yourdomain.com

make_toc: true
```