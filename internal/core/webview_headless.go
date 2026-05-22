//go:build headless

package core

import (
	"github.com/tituscheng/webviewgo/internal/types"
)

func newNative(opts types.Options) (Platform, error) {
	return newHeadless(opts)
}
