//go:build !production

package web

import "io/fs"

// 本地开发由 Vite 提供页面，Go 只提供 API。
func Assets() fs.FS { return nil }
