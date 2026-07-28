package main

import (
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/internal/testutil"
	"github.com/closeencoders/arcstatic/source"
)

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

			fileName: "fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Hello\n---\ntest"),

			count: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			root := "rootpath"
			ctx := source.CreateDefaultContext(root)
			ctx.PostInputDir = "fakepostloc"
			ctx.MakePostMetadata = false
			ctx.MakeSitemapXML = false

			fileMap := fstest.MapFS{test.fileName: &fstest.MapFile{Data: test.fileData}}
			store := testutil.NewFakeStorage(fileMap)
			err := executeBuild(ctx, store)
			if err != nil {
				t.Fatalf("learning before refactoring %v", err)
			}

			if len(store.State().WrittenFiles) != test.count {
				t.Errorf("the correct number of files was not written to fake storage for test wnt: %d got: %d", test.count, len(store.State().WrittenFiles))
			}
		})
	}
}

func TestNoPrefixWriteOrder(t *testing.T) {

	fileMap := fstest.MapFS{
		"fakepostloc/8-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 1\ndate: 1995-01-01\n---\n# Header")},
		"fakepostloc/2-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 2\ndate: 2023-10-10\n---\n# Header")},
		"fakepostloc/5-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 5\ndate: 2020-10-10\n---\n# Header")},
		"fakepostloc/3-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 3\ndate: 2022-10-11\n---\n# Header")},
		"fakepostloc/4-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 4\ndate: 2021-11-10\n---\n# Header")},
		"fakepostloc/1-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 1\ndate: 2024-10-10\n---\n# Header")},
		"fakepostloc/7-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 1\ndate: 2020-06-10\n---\n# Header")},
		"fakepostloc/6-post.md": &fstest.MapFile{Data: []byte("---\ntitle: 1\ndate: 2020-09-10\n---\n# Header")},
	}

	expected := []string{
		"rootpath/1-post/index.html",
		"rootpath/2-post/index.html",
		"rootpath/3-post/index.html",
		"rootpath/4-post/index.html",
		"rootpath/5-post/index.html",
		"rootpath/6-post/index.html",
		"rootpath/7-post/index.html",
		"rootpath/8-post/index.html",
	}

	root := "rootpath"
	ctx := source.CreateDefaultContext(root)
	ctx.PostInputDir = "fakepostloc"
	ctx.MakePostMetadata = false
	ctx.MakeSitemapXML = false

	// invokes a sort function mid exec by the provided frontmatter date instead of the file name.
	ctx.AllowNamelessDateSort = true

	store := testutil.NewFakeStorage(fileMap)
	err := executeBuild(ctx, store)
	if err != nil {
		t.Fatalf("build test failed with: %v", err)
	}

	if len(expected) != len(store.State().WrittenFiles) {
		t.Fatalf("the correct number of files was not written to fake storage for test wnt: %d got: %d", len(expected), len(store.State().WrittenFiles))
	}

	for i, name := range expected {
		if store.State().WrittenFiles[i] != name {
			t.Fatalf("Wrong order wnt: %s, got: %s", name, store.State().WrittenFiles[i])
		}
	}
}
