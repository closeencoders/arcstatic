package config

type SiteContext struct {
	// ComponentMap holds raw template partials or blocks keyed by their identifier name.
	ComponentMap map[string][]byte
	// TemplateMap holds raw layout templates keyed by their identifier name.
	TemplateMap map[string][]byte

	// SiteURL is the absolute base URL (e.g., "https://example.com").
	SiteURL  string `yaml:"site_url"`
	SiteRoot string
	Base     string

	Copy string `yaml:"copy"`

	DefaultType string `yaml:"default_type"`

	SiteOutputRoot string `yaml:"site_output_root"`

	// FrontmatterToken marks the boundary of configuration blocks in source files (e.g., "---" or "+++").
	FrontmatterToken string `yaml:"frontmatter_token"`
	// PostInputDir is the absolute path to the raw posts folder, decoupled from output locations.
	PostInputDir string `yaml:"post_input_dir"`
	// PostOutputDir is the path to write compiled posts, relative to the SiteRoot.
	PostOutputDir string `yaml:"post_output_dir"`
	// TODO:
	PageInputDir string

	// If set to true, will use a full path to the html file (e.g. /blah/content.html)
	FullHtmlPaths bool `yaml:"full_html_paths"`

	MaxDescriptionLen int `yaml:"max_description_len"`

	// TODO: Isolation to individual content preferences
	MakePostMetadata    bool `yaml:"make_post_metadata"`
	MakeSitemapXML      bool `yaml:"make_sitemap"`
	MakeTableOfContents bool `yaml:"make_toc"`

	JsonLog  bool   `yaml:"json_log"`
	LogLevel string `yaml:"log_level"`

	// By default, content files must be prefixed with a valid date
	AllowNamelessDateSort bool `yaml:"allow_nameless_date_sort"`

	// TODO: Will eventually support the archival paradigm so you can have urls the automatically have date and other misc data.
	KeepDateUrl bool `yaml:"keep_date_url"`
	// Will allow a defined type on content to dictate it url path. So a type of "posts" can end up in a path of "/posts/xxx"
	AllowTaxonomyPaths bool `yaml:"allow_taxonomy_paths"`

	// rules that the sitemap.xml will exclude. (e.g. /posts) will exclude any rendered paths that contain "/posts"
	SitemapExclusions       []string `yaml:"sitemap_exclusions"`
	SitemapAllowFileLastMod bool     `yaml:"sitemap_allow_file_lastmod"`

	AllowManifest bool
}

// TODO: enforce access patterns based on when data is mutated.
func NewContext(root string) *SiteContext {
	return &SiteContext{
		ComponentMap: make(map[string][]byte),
		TemplateMap:  make(map[string][]byte),

		SiteRoot: root,

		PostOutputDir: "/",
		Base:          "/",

		FrontmatterToken: "---",

		MaxDescriptionLen: 156,

		DefaultType: "All",

		MakeSitemapXML:          true,
		MakePostMetadata:        true,
		SitemapAllowFileLastMod: true,

		AllowManifest: true,
	}
}
