package generate

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
)

type Templater struct {
	template *template.Template
}

func NewTemplater(components map[string][]byte, customFuncs map[string]any) (*Templater, error) {

	funcs := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"limit":    limitFunc,
	}
	for k, v := range customFuncs {
		funcs[k] = v
	}

	temp := template.New("base").Funcs(funcs)
	for name, content := range components {
		if _, err := temp.New(name).Parse(string(content)); err != nil {
			return nil, fmt.Errorf("component %q parse error: %w", name, err)
		}
	}

	return &Templater{template: temp}, nil
}

func (t *Templater) Render(data any, templateSrc string) ([]byte, error) {

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
