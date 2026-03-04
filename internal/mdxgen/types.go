package mdxgen

// PackageData contains the minimum data needed to render a package page.
type PackageData struct {
	Path        string
	Name        string
	ModulePath  string
	Version     string
	Synopsis    string
	Description string
	Readme      string
	DocSource   []byte
}

// Summary reports generator results.
type Summary struct {
	Generated int
	Skipped   int
	Failed    int
}

