package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var cachePath *string = flag.String("cache", "/tmp/img-cache", "path for cache")

type cacheEntry struct {
	exists   bool
	filepath string
	stat     os.FileInfo
	file     *os.File
}

func (c *cacheEntry) OpenFile() (*os.File, error) {
	if c.file == nil {
		file, err := os.Open(c.filepath)
		if err != nil {
			return nil, err
		}
		c.file = file
	}

	return c.file, nil
}

func init() {
	os.MkdirAll(*cachePath, os.ModeDir|os.ModePerm)
}

func CacheStore(key string, content []byte) error {
	file, err := os.Create(CacheFilename(key))
	if err != nil {
		fmt.Printf("could not create cache: %s", err)
		return err
	}
	defer file.Close()

	file.Write(content)
	return nil
}

func CacheFilename(key string) string {
	return filepath.Join(*cachePath, sha1hash(key))
}

func CacheLookup(key string) *cacheEntry {
	path := CacheFilename(key)

	stat, err := os.Stat(path)
	if err != nil {
		return nil
	}

	return &cacheEntry{
		filepath: path,
		stat:     stat,
	}
}

func sha1hash(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}
