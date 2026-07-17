package generate

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
)

type TemplateRenderer struct {
	template *template.Template
}

type Renderer interface {
	Render(data any, templateSrc string) ([]byte, error)
	Load(components map[string][]byte, customFuncs map[string]any) error
}

var _ Renderer = &TemplateRenderer{}

func (t *TemplateRenderer) Load(components map[string][]byte, customFuncs map[string]any) error {

	funcs := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"limit":    limitFunc,
	}
	for k, v := range customFuncs {
		funcs[k] = v
	}

	t.template = template.New("base").Funcs(funcs)
	for name, content := range components {
		if _, err := t.template.New(name).Parse(string(content)); err != nil {
			return fmt.Errorf("component %q parse error: %w", name, err)
		}
	}

	return nil
}

func (t *TemplateRenderer) Render(data any, templateSrc string) ([]byte, error) {

	tmpl, err := t.template.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone template: %w", err)
	}

	tmpl, err = tmpl.Parse(templateSrc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func limitFunc(limit int, data any) any {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		l := v.Len()
		if l > limit {
			l = limit
		}
		return v.Slice(0, l).Interface()
	default:
		return data
	}
}
