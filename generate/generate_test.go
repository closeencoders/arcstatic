package generate

import (
	"bytes"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
	"github.com/closeencoders/arcstatic/storage"
)

type FakeStorage struct {
	TestFiles fstest.MapFS
}

var _ storage.Storage = FakeStorage{}

func (fs FakeStorage) Write(name string, data []byte, perm int) error {
	return nil
}

func (fs FakeStorage) Open(name string) (fs.File, error) {
	return fs.TestFiles.Open(name)
}

func (fs FakeStorage) Mkdir(perm int, path ...string) (string, error) {
	return "", nil
}

func TestContentConversion(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		fileData []byte

		want []byte
	}{
		{
			name: "Should Generate Basic Text",

			fileName: "rootpath/fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Text\n---\ntest"),

			want: []byte("<p>test</p>\n"),
		},
		{
			name: "Should Generate Basic Markdown",

			fileName: "rootpath/fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Markdown\n---\n# One\n## Two\n### Three\nTesting"),

			want: []byte("<h1 id=\"one\">One</h1>\n<h2 id=\"two\">Two</h2>\n<h3 id=\"three\">Three</h3>\n<p>Testing</p>\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			root := "rootpath"
			ctx := source.CreateDefaultContext(root)
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"

			store := FakeStorage{fstest.MapFS{
				test.fileName: &fstest.MapFile{Data: test.fileData},
			}}

			conv, sm, err := createTestData(ctx, store)
			if err != nil {
				t.Fatal("test failed")
			}

			content, err := conv.ToContent(test.fileData, sm.SiteContentEntities[0], sm.ContentManifest)
			if err != nil {
				t.Fatal("learning before refactoring")
			}
			if !bytes.Equal(test.want, content) {
				t.Errorf("invalid conversion. Strings: \nwnt: %s\ngot: %s \nBytes: \nwnt: %x\ngot: %x ", []byte(test.want), []byte(content), test.want, content)
			}
			if len(content) == 0 {
				t.Error("no conversion")
			}
		})
	}
}

func TestGenerator(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		fileData []byte
		count    int
	}{
		{
			name:  "Should Not Generate Anything Without Files",
			count: 0,
		},
		{
			name: "Should Generate Basic Content",

			fileName: "rootpath/fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Hello\n---\ntest"),

			count: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			root := "rootpath"
			ctx := source.CreateDefaultContext(root)
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"

			store := FakeStorage{fstest.MapFS{test.fileName: &fstest.MapFile{Data: test.fileData}}}
			conv, sm, err := createTestData(ctx, store)
			if err != nil {
				t.Fatal("test failed")
			}

			gen := NewGenerator(ctx, *conv, store)
			count, err := gen.Generate(sm)
			if err != nil {
				t.Fatal("learning before refactoring")
			}
			if count != test.count {
				t.Error("generator did not process expected")
			}
		})
	}
}

func createTestData(ctx *config.SiteContext, store FakeStorage) (*converter, *source.SiteMetadata, error) {

	templ, err := NewTemplater(ctx.ComponentMap, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load templater for test %w", err)
	}

	conv := NewConverter(ctx, *NewMarkdown(ctx), *templ)
	ml := source.NewMetadata(ctx, store)
	sm, err := ml.LoadMetadata(ctx.SiteRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load metadata for test: %w", err)
	}

	return conv, sm, nil
}
