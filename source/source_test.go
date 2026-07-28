package source

import (
	"errors"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/internal/testutil"
)

func TestLoadConfig(t *testing.T) {

	const baseDir = "testlocation"
	defaultPostLoc := path.Join(baseDir, _postsLoc)
	configLoc := path.Join(baseDir, ConfigName)

	tests := []struct {
		name       string
		configFile fstest.MapFS

		wantUrl string
		wantDir string

		wantErr error
	}{
		{
			name: "Context Should Contain Partial Override Values From File",
			configFile: fstest.MapFS{
				configLoc: &fstest.MapFile{
					Data: []byte("site_url: https://location.com\npost_input_dir: location123/posts"),
				},
			},
			wantUrl: "https://location.com",
			wantDir: "location123/posts",
		},
		{
			name: "Empty Config Should Use Full Default",
			configFile: fstest.MapFS{
				configLoc: &fstest.MapFile{Data: []byte("")},
			},
			wantUrl: _defaultUrl,
			wantDir: defaultPostLoc,
		},
		{
			name:       "Config File Not Found Should Use Full Default",
			configFile: nil,
			wantUrl:    _defaultUrl,
			wantDir:    defaultPostLoc,
		},
		{
			name: "Invalid Configuration Should Return Error",
			configFile: fstest.MapFS{
				configLoc: &fstest.MapFile{Data: []byte("site_url: : : : :")},
			},
			wantErr: errInvalidConfig,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctx, err := LoadSiteContext(baseDir, baseDir, testutil.NewFakeStorage(test.configFile))
			if err != nil {
				if test.wantErr != nil && errors.Is(err, test.wantErr) {
					return
				}
				t.Fatalf("unexpected error state while loading site context test: %v, wantErr: %v", err, test.wantErr)
			}
			testutil.AssertEqual(t, "SiteUrl", ctx.SiteURL, test.wantUrl)
			testutil.AssertEqual(t, "PostInputDir", ctx.PostInputDir, test.wantDir)
		})
	}
}

func TestConfigOverrideCmdSettings(t *testing.T) {

	testRoot := "testRoot"
	configLoc := filepath.Join(testRoot, "arcconfig.yml")

	c := fstest.MapFS{
		configLoc: &fstest.MapFile{Data: []byte("site_output_root: /test/out")},
	}

	store := testutil.NewFakeStorage(c)

	ctx, err := LoadSiteContext(testRoot, "testOut/ignored", store)
	if err != nil {
		t.Fatalf("Unexpected error while test config override")
	}

	testutil.AssertEqual(t, "SiteOutputRoot", ctx.SiteOutputRoot, "/test/out")

}

func TestTaxonomy(t *testing.T) {

	initContext := config.NewContext("test")

	tests := []struct {
		name string

		path     string
		fileName string
		fileData []byte

		allowTaxonomyPaths bool

		wantType    string
		wantOutPath string
	}{
		{
			name:     "Post With Default Type Should Be Added To Manifest",
			path:     "fakepostloc",
			fileName: "fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Hello\n---\n# Header"),

			allowTaxonomyPaths: true,

			wantType:    initContext.DefaultType,
			wantOutPath: initContext.DefaultType + "/basic-post/index.html",
		},
		{
			name:     "Post Valid Override Type Should Be Added To Manifest",
			path:     "fakepostloc",
			fileName: "fakepostloc/2026-06-01-basic_post.md",
			fileData: []byte("---\ntitle: Hello\ntype: other---\n# Header"),

			allowTaxonomyPaths: true,

			wantType:    "other",
			wantOutPath: "other/basic-post/index.html",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctx := CreateDefaultContext("public")
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"
			ctx.AllowTaxonomyPaths = test.allowTaxonomyPaths

			testFile := fstest.MapFS{test.fileName: &fstest.MapFile{Data: test.fileData}}
			ml := NewMetadata(ctx, testutil.NewFakeStorage(testFile))

			result, err := ml.LoadMetadata(test.path)
			if err != nil {
				t.Fatalf("LoadMetadata() execution unexpected error = %v", err)
			}
			if len(result.SiteContentEntities) == 0 {
				t.Fatalf("Expected Metadata for taxonomy test did not load")
			}

			ce := result.SiteContentEntities[0]
			testutil.AssertEqual(t, "wrong type", ce.ContentMetadata.Type, test.wantType)

			_, ok := result.ContentManifest[test.wantType]
			if !ok {
				t.Errorf("type was not applied to content manifest %s %s", test.wantType, ce.FileName)
			}

			if test.allowTaxonomyPaths {
				testutil.AssertEqual(t, "expected type in path", ce.OutputPath, test.wantOutPath)
			}
		})
	}
}

func TestInvalidMetadata(t *testing.T) {

	tests := []struct {
		name     string
		fileData fstest.MapFS
		path     string

		wantErr error
	}{
		{
			name: "Unknown File Extension Should Not Load Or Attempt To Load Anything",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-about.xyz": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
		},
		{
			name: "Draft Post Should Not Load When Set To True",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-draft_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\ndraft: true\n---\n# Header")},
			},
		},
		{
			name: "Post Without Date Prefix Should Not Load With Default Settings",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/noprefix_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantErr: errDatePrefix,
		},
		{
			name: "Unknown Page Location Should Not Load Or Attempt To Load Anything",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"xxxxxx/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctx := CreateDefaultContext("public")
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"

			ml := metadata{ctx: ctx, store: testutil.NewFakeStorage(test.fileData)}
			result, err := ml.LoadMetadata(test.path)
			if err != nil {

				if test.wantErr != nil && errors.Is(err, test.wantErr) {
					return
				}

				t.Fatalf("source loading unexpected error = %v", err)
			}

			entitiesCount := len(result.SiteContentEntities)
			if entitiesCount != 0 {
				t.Errorf("expected 0 entities loaded, got %d", entitiesCount)
			}
		})
	}
}

func TestLoadValidMetadata(t *testing.T) {

	tests := []struct {
		name     string
		fileData fstest.MapFS
		path     string

		wantTitle    string
		wantPath     string
		wantTemplate string
		wantType     string

		allowDateLoad bool
	}{
		{
			name: "Known Post Location Should Load",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-basic_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantTitle:    "Hello",
			wantPath:     "basic-post/index.html",
			wantTemplate: "post.html",
		},
		{
			name: "Known Post Should Load With Irregular Name",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-basic *&^(*&)  post.md": &fstest.MapFile{Data: []byte("---\ntitle: Irregular Name\n---\n# Header")},
			},
			wantTitle:    "Irregular Name",
			wantPath:     "basic-post/index.html",
			wantTemplate: "post.html",
		},
		{
			name: "Post Without Date Prefix With Allow Sort Settings Should Load",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/noprefix_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Without Date Prefix\n---\n# Header")},
			},
			allowDateLoad: true,

			wantTitle:    "Without Date Prefix",
			wantPath:     "noprefix-post/index.html",
			wantTemplate: "post.html",
		},
		{
			name: "Known Page Location Should Load",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"fakepageloc/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			wantTitle:    "About",
			wantPath:     "about/index.html",
			wantTemplate: "page.html",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			ctx := CreateDefaultContext("public")
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"
			ctx.AllowNamelessDateSort = test.allowDateLoad

			ml := metadata{ctx: ctx, store: testutil.NewFakeStorage(test.fileData)}
			result, err := ml.LoadMetadata(test.path)
			if err != nil {
				t.Fatalf("source loading unexpected error = %v", err)
			}

			if len(result.SiteContentEntities) == 0 {
				t.Errorf("metadata should have been loaded for target path: %q", test.path)
				return
			}

			ce := result.SiteContentEntities[0]
			if ce.InputPath == ctx.PostInputDir {
				if ce.ContentMetadata.Type == "" {
					testutil.AssertEqual(t, "type should have defaulted", ce.ContentMetadata.Type, ctx.DefaultType)
				} else {
					testutil.AssertEqual(t, "type", ce.ContentMetadata.Type, test.wantType)
				}
			}
			testutil.AssertEqual(t, "template", ce.ContentMetadata.TemplateId, test.wantTemplate)
			testutil.AssertEqual(t, "title", ce.ContentMetadata.Title, test.wantTitle)
			testutil.AssertEqual(t, "path", ce.OutputPath, test.wantPath)
		})
	}
}

func TestNoPrefixLoadOrder(t *testing.T) {

	files := fstest.MapFS{
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
		"1-post.md", "2-post.md", "3-post.md", "4-post.md", "5-post.md", "6-post.md", "7-post.md", "8-post.md",
	}

	ctx := CreateDefaultContext("root")
	ctx.PostInputDir = "fakepostloc"
	ctx.AllowNamelessDateSort = true

	ml := metadata{ctx: ctx, store: testutil.NewFakeStorage(files)}
	result, err := ml.LoadMetadata("fakepostloc")
	if err != nil {
		t.Fatalf("source loading unexpected error = %v", err)
	}

	if len(result.SiteContentEntities) == 0 {
		t.Errorf("metadata should have been loaded for target path: %q", "fakepostloc")
		return
	}

	for i, f := range expected {
		if result.SiteContentEntities[i].FileName != f {
			t.Fatalf("Wrong order wnt: %s, got: %s", f, result.SiteContentEntities[i].FileName)
		}
	}
}
