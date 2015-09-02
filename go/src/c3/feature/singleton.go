package feature

import "sync"

var onlyonce sync.Once
var protect chan int
var singleton *Feature

// Use this function to get the interface Featurer which support both singleton
// and also user declared feature object.
func GetSingleton() Featurer {
	return &globalFeature{}
}

// Never expose this func in order to directly access global Feature singleton
// object. As it can break the thread safety of the singleton access. Use
// the GetSingleton() to get interface Featurer which support both singleton
// object and also user declared object.
func getSingleton() (r *Feature) {
	onlyonce.Do(func() {
		singleton = New()
	})
	return singleton
}

// ClearCache clears the cached features that were loaded from the object server.
func ClearCache() {
	protect <- 1
	defer func() { <-protect }()
	getSingleton().ClearCache()
}

// SetDefaultContext sets the default context to the given context string. More
// than one context can be specified by separating them with semicolon.
func SetDefaultContext(ctx string) {
	protect <- 1
	defer func() { <-protect }()
	getSingleton().SetDefaultContext(ctx)
}

// SetGlobalContext sets the default global context to the given context.
func SetGlobalContext(ctx string) {
	protect <- 1
	defer func() { <-protect }()
	getSingleton().SetGlobalContext(ctx)
}

// Bool returns the boolean value of the feature application state for the
// given tag.
func Bool(tag string) bool {
	protect <- 1
	defer func() { <-protect }()
	return getSingleton().Bool(tag)
}

// Int returns the integer value of the feature application state for the given
// tag.
func Int(tag string) int {
	protect <- 1
	defer func() { <-protect }()
	return getSingleton().Int(tag)
}

// Str returns the string value of the feature application state for the given
// tag.
func Str(tag string) string {
	protect <- 1
	defer func() { <-protect }()
	return getSingleton().Str(tag)
}

func init() {
	protect = make(chan int, 1) // Allocate a buffer channel
}

type Featurer interface {
	ClearCache()
	SetDefaultContext(ctx string)
	SetGlobalContext(ctx string)
	Bool(tag string) bool
	Int(tag string) int
	Str(tag string) string
}

type globalFeature struct {
}

func (g *globalFeature) ClearCache() {
	ClearCache()
}

func (g *globalFeature) SetDefaultContext(ctx string) {
	SetDefaultContext(ctx)
}

func (g *globalFeature) SetGlobalContext(ctx string) {
	SetGlobalContext(ctx)
}

func (g *globalFeature) Bool(tag string) bool {
	return Bool(tag)
}

func (g *globalFeature) Int(tag string) int {
	return Int(tag)
}

func (g *globalFeature) Str(tag string) string {
	return Str(tag)
}
