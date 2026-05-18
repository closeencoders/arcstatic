# arcstatic

Arcstatic is a simple static site generator practice/learning project. While not intended for a production environement, I do personally use it for my own projects.

(documentation work in progress, this code is not fully released or completed yet)

## Example Commands

```
Usage of arcstatic:
  -build
        Builds static site from provided resources. Defaults to current worked directory.
  -in string
        Static site context input location. Defaults to current worked directory.
  -port int
        location to file serve. If the serve command is not used, this is ignored (default 8000)
  -serve
        Serve static site from provided resources for testing. Defaults to current worked directory.
  -verbose
        run verbose
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
post_root_dir: /override/input/post/location
site_url: https://yourdomain.com

make_toc: true

# type --> category --> tag
default_type: blog
```