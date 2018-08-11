/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package vcs

var defaultGitIgnores = []string{
	// Self
	".thrap",
	"secrets.*",
	// OS X
	".Trash",
	".DS_Store",
	// Misc
	"*.log",
	"*.test",
	".env",
	// Bins
	"*.exe",
	"*.dll",
	"*.so",
	"*.dylib",
	// IDEs
	".vscode/",
}

// DefaultGitIgnores returns the default set to use in a .gitignore file
func DefaultGitIgnores() []string {
	return defaultGitIgnores
}
