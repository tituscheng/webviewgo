//go:build !darwin && !windows && !linux && !headless

package core

import (
	"fmt"

	"github.com/tituscheng/webviewgo/internal/types"
)

func newNative(opts types.Options) (Platform, error) {
	return nil, fmt.Errorf("core: no native backend available for this platform")
}
