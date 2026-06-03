package generate

import (
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/internal/testutil"
	"github.com/closeencoders/arcstatic/source"
)

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
		gen := NewGenerator(ctx, *conv, testutil.NewFakeStorage(fstest.MapFS{}))
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
