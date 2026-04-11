package enumerators

type PluginInfo struct {
	Name    string
	Path    string
	Version string
}

type ThemeInfo struct {
	Name    string
	Path    string
	Version string
}

type UserInfo struct {
	Username string
	ID       int
	Email    string
	Method   string
}

type VersionInfo struct {
	Software string
	Version  string
	Source   string
}

type DiscoveryResult struct {
	Target           string
	WordPressVersion string
	Plugins          []PluginInfo
	Themes           []ThemeInfo
	Users            []UserInfo
	Versions         []VersionInfo
	Errors           []string
}
