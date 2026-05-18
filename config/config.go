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

	PostInputDir string `yaml:"post_root_dir"`
	PostOutDir   string
	PageInputDir string

	FullHtmlPath         bool
	GeneratePostMetadata bool
	GenerateSitemapXml   bool
	MakeTableOfContents  bool `yaml:"make_toc"`
}
