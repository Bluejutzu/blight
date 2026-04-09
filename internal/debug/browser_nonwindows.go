//go:build !windows

package debug

import "github.com/pkg/browser"

func openBrowser(url string) {
	browser.OpenURL(url)
}
