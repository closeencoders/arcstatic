package generate

import (
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

// TODO: seeing if the test coverage can tell this test is useless or not.
func TestConvertToContent(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name     string
		ce       source.ContentEntity
		manifest source.Manifest
		rawFile  []byte
	}{
		{
			name:    "Invalid data should not generate anything",
			rawFile: []byte("blah blah blah"),
		},
	}

	for _, tt := range tests {

		ctx := config.NewContext("testloc")
		temp, err := NewTemplater(ctx.ComponentMap, nil)
		if err != nil {
			t.Fatal("failed to load templater, invalid test configuration")
		}

		conv := NewConverter(ctx, *NewMarkdown(ctx), *temp)
		gen := NewGenerator(ctx, *conv, FakeStorage{fstest.MapFS{}})
		metadata := source.SiteMetadata{
			SiteContentEntities: []*source.ContentEntity{
				&tt.ce,
			},
		}

		err = gen.Generate(metadata)
		if err != nil {
			t.Fatal("learning before refactoring")
		}
	}

}
