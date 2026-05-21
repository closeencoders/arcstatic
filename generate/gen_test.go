package generate

import (
	"log/slog"
	"testing"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
)

// TODO: also seeing if the test coverage can tell this test is useless or not.
func TestConvertToContent(t *testing.T) {

	ctx := config.NewContext("testloc")

	temp, err := NewTemplater(ctx.ComponentMap, nil)
	if err != nil {
		t.Fatal("learning before refactoring")
	}

	conv := NewConverter(ctx, *NewMarkdown(ctx), *temp)

	rawFile := []byte("blah blah blah")
	metadata := source.ContentEntity{}
	manifest := source.ContentManifest{}

	content, err := conv.ConvertToContent(rawFile, &metadata, manifest)
	if err != nil {
		t.Fatal("learning before refactoring")
	}

	slog.Info("content", "content", content)
	// fmt.Printf("content %v", content)
}
