package config

type SiteContext struct {
	ComponentMap map[string][]byte
	TemplateMap  map[string][]byte

	SiteUrl  string `yaml:"site_url"`
	SiteRoot string
	Base     string

	DefaultType string `yaml:"default_type"`

	SiteInputRoot string

	FrontmatterToken []byte `yaml:"frontmatter_token"`

	PostInputDir string `yaml:"post_input_dir"`
	PostOutDir   string `yaml:"post_output_dir"`
	PageInputDir string

	// TODO: Isolation to individual content preferences
	FullHtmlPath        bool
	MakePostMetadata    bool `yaml:"make_post_data"`
	MakeSitemapXml      bool `yaml:"make_sitemap"`
	MakeTableOfContents bool `yaml:"make_toc"`
}

// TODO: enforce access patterns based on when data is mutated.
func NewContext(root string) *SiteContext {
	return &SiteContext{
		ComponentMap: make(map[string][]byte),
		TemplateMap:  make(map[string][]byte),

		SiteInputRoot: root,
		SiteRoot:      root,
		PostOutDir:    root,
		Base:          "/",

		FrontmatterToken: []byte("---"),

		FullHtmlPath:        false,
		MakeSitemapXml:      true,
		MakePostMetadata:    true,
		MakeTableOfContents: false,
	}
}
