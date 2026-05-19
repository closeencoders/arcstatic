package source

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadsiteCtx(t *testing.T) {
	const location = "/testlocation"
	defaultPostLoc := filepath.Join(location, _postsLoc)

	tests := []struct {
		name    string
		yaml    string
		wantUrl string
		wantDir string
		wantErr bool
	}{
		{
			name:    "Full Override",
			yaml:    "site_url: https://blah.com\npost_input_dir: /blah/posts",
			wantUrl: "https://blah.com",
			wantDir: "/blah/posts",
		},
		{
			name:    "Partial Override",
			yaml:    "site_url: https://blah2.com",
			wantUrl: "https://blah2.com",
			wantDir: defaultPostLoc,
		},
		{
			name:    "Full Default",
			yaml:    "",
			wantUrl: _defaultUrl,
			wantDir: defaultPostLoc,
		},
		{
			name:    "Malformed YAML",
			yaml:    "site_url: : : : :",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.yaml)
			result, err := loadSiteCtx(location, r)

			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr is %v, but got error: %v", tt.wantErr, err)
			}
			if tt.wantErr {
				return
			}
			if result.SiteUrl != tt.wantUrl {
				t.Errorf("SiteUrl mismatch\n got:  %s\n want: %s", result.SiteUrl, tt.wantUrl)
			}
			if result.PostInputDir != tt.wantDir {
				t.Errorf("PostInputDir mismatch\n got:  %s\n want: %s", result.PostInputDir, tt.wantDir)
			}
		})
	}
}

type testFileData struct {
	fileName  string
	rootPath  string
	fileBytes []byte
}

func TestGetContentEntity(t *testing.T) {

	ctx := newDefaultContext("public")
	ctx.SiteRoot = "public"
	ctx.Base = "/"
	ctx.PostInputDir = "/fakepostloc"

	ml := Metadata{ctx: ctx}

	tests := []struct {
		name         string
		fileData     testFileData
		wantTitle    string
		wantPath     string
		wantBody     string
		wantTemplate string
	}{
		{
			name:      "Basic Markdown Default Post",
			fileData:  testFileData{fileName: "my-post.md", rootPath: "/fakepostloc", fileBytes: []byte("---\ntitle: Hello\n---\n# Header")},
			wantTitle: "Hello",
			wantPath:  "my-post", // Or your specific pretty URL logic
			// wantBody:     "<h1>Header</h1>\n", // Depends on your markdown engine
			wantTemplate: "post.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ml.getContentMetadata(tt.fileData.fileBytes, tt.fileData.fileName, tt.fileData.rootPath)
			if err != nil {
				t.Fatalf("getContentEntity() error = %v", err)
			}
			if result.ContentMetadata.TemplateId != tt.wantTemplate {
				t.Errorf("Template = %v, want %v", result.ContentMetadata.TemplateId, tt.wantTemplate)
			}
			if result.ContentMetadata.Title != tt.wantTitle {
				t.Errorf("Title = %v, want %v", result.ContentMetadata.Title, tt.wantTitle)
			}
		})
	}
}
