package feature

import "sync"

var onlyonce sync.Once
var protect chan int
var singleton *Feature

func GetSingleton() (r *Feature) {
	onlyonce.Do(func() {
		singleton = New()
	})
	return singleton
}

// ClearCache clears the cached features that were loaded from the object server.
func ClearCache() {
	protect <- 1
	defer func() { <-protect }()
	GetSingleton().ClearCache()
}

// SetGlobalContext sets the default global context to the given context.
func SetGlobalContext(ctx string) {
	protect <- 1
	defer func() { <-protect }()
	GetSingleton().SetGlobalContext(ctx)
}

// Bool returns the boolean value of the feature application state for the
// given tag.
func Bool(tag string) bool {
	protect <- 1
	defer func() { <-protect }()
	return GetSingleton().Bool(tag)
}

// Int returns the integer value of the feature application state for the given
// tag.
func Int(tag string) int {
	protect <- 1
	defer func() { <-protect }()
	return GetSingleton().Int(tag)
}

// Str returns the string value of the feature application state for the given
// tag.
func Str(tag string) string {
	protect <- 1
	defer func() { <-protect }()
	return GetSingleton().Str(tag)
}

func init() {
	protect = make(chan int, 1) // Allocate a buffer channel
}
