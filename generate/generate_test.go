package generate

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
	"github.com/closeencoders/arcstatic/storage"
)

type fakeStorage struct {
	testFiles fstest.MapFS
	state     *fakeState
}

func newFakeStorage(testFiles fstest.MapFS) fakeStorage {
	return fakeStorage{testFiles, &fakeState{writtenFiles: []string{}}}
}

type fakeState struct {
	writtenFiles []string
	mu           sync.Mutex
}

var _ storage.Storage = fakeStorage{}

func (fs fakeStorage) Write(name string, data []byte, perm int) error {
	slog.Warn("logging", "name", name)
	fs.state.writtenFiles = append(fs.state.writtenFiles, name)
	return nil
}

func (fs fakeStorage) Open(name string) (fs.File, error) {
	return fs.testFiles.Open(name)
}

func (fs fakeStorage) Mkdir(perm int, path ...string) (string, error) {
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

			store := newFakeStorage(fstest.MapFS{test.fileName: &fstest.MapFile{Data: test.fileData}})
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

func createTestData(ctx *config.SiteContext, store fakeStorage) (*converter, *source.SiteMetadata, error) {

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
