# Arcstatic
[![coverage](https://raw.githubusercontent.com/closeencoders/arcstatic/badges/.badges/main/coverage.svg)](https://github.com/closeencoders/arcstatic/actions)
[![Go Status](https://github.com/closeencoders/arcstatic/actions/workflows/go.yml/badge.svg)](https://github.com/closeencoders/arcstatic/actions)

Arcstatic is a simple static site generator practice/learning project. While not intended for a production environment, I do personally use it for my own projects.

(documentation work in progress, this code is not fully released or completed yet)

## Example Commands
```
Usage:
  arcstatic [flags]

Flags:
  -b, --build       builds static site from provided resources
  -h, --help        help for arcstatic
  -i, --in string   override default site context current working directory input location (default CWD)
  -p, --port int    port number to file serve the site, if the serve command is not used, this is ignored (default 8000)
  -s, --serve       serve static site from provided resources, currently only for testing
  -v, --verbose     run verbose with debug logs
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