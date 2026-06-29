// Package dedup 提供文章 URL 去重功能。
package dedup

import (
	"crypto/sha256"
	"fmt"
)

// URLHash 计算文章 URL 的 SHA256 哈希值。
func URLHash(url string) string {
	h := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", h)
}
