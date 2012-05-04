package main

import (
	"testing"
)

func TestScreenshot(t *testing.T) {
	pool = NewWebkitPool(1)
	pool.Screenshot("http://www.snowdevil.ca")
}
