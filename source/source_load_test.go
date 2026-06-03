package source

import (
	"path"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/closeencoders/arcstatic/internal/testutil"
)

func TestLoadConfig(t *testing.T) {

	t.Parallel()

	const baseDir = "testlocation"
	defaultPostLoc := path.Join(baseDir, _postsLoc)
	configLoc := path.Join(baseDir, _configName)

	tests := []struct {
		name       string
		configFile fstest.MapFS
		wantUrl    string
		wantDir    string
		wantErr    bool
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
			name: "Invalid Configuration Should Return Error",
			configFile: fstest.MapFS{
				configLoc: &fstest.MapFile{Data: []byte("site_url: : : : :")},
			},
			wantErr: true,
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
	}

	var wg sync.WaitGroup

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			wg.Add(1)

			mockStore := testutil.NewFakeStorage(tt.configFile)

			result, err := LoadSiteContext(baseDir, mockStore)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state while loading site context test: %v, wantErr: %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			testutil.AssertEqual(t, "SiteUrl", result.SiteURL, tt.wantUrl)
			testutil.AssertEqual(t, "PostInputDir", result.PostInputDir, tt.wantDir)
		})
	}
}

// TODO: Refactor to break test up into chunks of expected behavior
func TestLoadMetadata(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name         string
		fileData     fstest.MapFS
		path         string
		wantTitle    string
		wantPath     string
		wantTemplate string
		wantType     string
		notLoaded    bool
	}{
		{
			name: "Known Post Location Should Load",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-basic_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantTitle:    "Hello",
			wantPath:     "basic-post",
			wantTemplate: "post.html",
		},
		{
			name: "Known Post Location Should Load With Irregular Name",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-basic *&^(*&)  post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\n---\n# Header")},
			},
			wantTitle:    "Hello",
			wantPath:     "basic-post",
			wantTemplate: "post.html",
		},
		{
			name: "Unknown File Extension Should Not Load Or Attempt To Load Anything",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-about.xyz": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			notLoaded: true,
		},
		{
			name: "Draft Post Should Not Load When Set To True",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-draft_post.md": &fstest.MapFile{Data: []byte("---\ntitle: Hello\ndraft: true\n---\n# Header")},
			},
			notLoaded: true,
		},
		{
			name: "Custom Taxonomy Post Should Use Custom Type",
			path: "fakepostloc",
			fileData: fstest.MapFS{
				"fakepostloc/2026-06-01-other_type_post.md": &fstest.MapFile{
					Data: []byte("---\ntitle: Custom Type\ntype: others\n---\n# Header"),
				},
			},
			wantTitle:    "Custom Type",
			wantPath:     "other-type-post",
			wantTemplate: "post.html",
			wantType:     "others",
		},
		{
			name: "Known Page Location Should Load",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"fakepageloc/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			wantTitle:    "About",
			wantPath:     "about",
			wantTemplate: "page.html",
		},
		{
			name: "Unknown Page Location Should Not Load Or Attempt To Load Anything",
			path: "fakepageloc",
			fileData: fstest.MapFS{
				"xxxxxx/about.html": &fstest.MapFile{Data: []byte("---\ntitle: About\n---\n# Header")},
			},
			notLoaded: true,
		},
	}

	var wg sync.WaitGroup

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			wg.Add(1)

			ctx := createDefaultContext("public")
			ctx.PostInputDir = "fakepostloc"
			ctx.PageInputDir = "fakepageloc"

			fs := testutil.NewFakeStorage(tt.fileData)
			ml := metadata{ctx: ctx, store: fs}

			result, err := ml.LoadMetadata(tt.path)
			if err != nil {
				t.Fatalf("LoadMetadata() execution unexpected error = %v", err)
			}

			entitiesCount := len(result.SiteContentEntities)
			if tt.notLoaded {
				if entitiesCount != 0 {
					t.Fatalf("expected 0 entities loaded, got %d", entitiesCount)
				}
				return
			}
			if entitiesCount == 0 {
				t.Fatalf("metadata should have been loaded for target path: %q", tt.path)
			}

			ce := result.SiteContentEntities[0]
			if ce.InputPath == ctx.PostInputDir {
				if ce.ContentMetadata.Type == "" {
					testutil.AssertEqual(t, "type should have defaulted", ce.ContentMetadata.Type, ctx.DefaultType)
				} else {
					testutil.AssertEqual(t, "type", ce.ContentMetadata.Type, tt.wantType)
				}
			}
			testutil.AssertEqual(t, "template", ce.ContentMetadata.TemplateId, tt.wantTemplate)
			testutil.AssertEqual(t, "title", ce.ContentMetadata.Title, tt.wantTitle)
		})
	}
}
