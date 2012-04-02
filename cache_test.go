package main

import (
	"testing"
	"testing/quick"
)

func TestCacheFilename(t *testing.T) {
	quick.CheckEqual(CacheFilename("blob"), "", nil)

}
