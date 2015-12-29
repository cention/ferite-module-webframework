package feature

import (
	"sync"

	// cention
	wf "c3/osm/webframework"
)

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

func Preload() {
	protect <- 1
	defer func() { <-protect }()
	getSingleton().Preload()
}

func SetDefaultContexts(ctxs []string) {
	protect <- 1
	defer func() { <-protect }()
	getSingleton().SetDefaultContexts(ctxs)
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

func DefaultContext() string {
	protect <- 1
	defer func() { <-protect }()
	return getSingleton().DefaultContext()
}

func States(featureTags, contexts []string) (map[string]*wf.FeatureApplication, error) {
	protect <- 1
	defer func() { <-protect }()
	return getSingleton().States(featureTags, contexts)
}

func State(tag string) (*wf.FeatureApplication, error) {
	protect <- 1
	defer func() { <-protect }()
	return getSingleton().State(tag)
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
	Preload()
	SetDefaultContexts([]string)
	SetDefaultContext(string)
	SetGlobalContext(string)
	DefaultContext() string
	States([]string, []string) (map[string]*wf.FeatureApplication, error)
	State(string) (*wf.FeatureApplication, error)
	Bool(string) bool
	Int(string) int
	Str(string) string
}

type globalFeature struct{}

func (g globalFeature) ClearCache() {
	ClearCache()
}

func (g globalFeature) Preload() {
	Preload()
}

func (g globalFeature) SetDefaultContexts(ctxs []string) {
	SetDefaultContexts(ctxs)
}

func (g globalFeature) SetDefaultContext(ctx string) {
	SetDefaultContext(ctx)
}

func (g globalFeature) SetGlobalContext(ctx string) {
	SetGlobalContext(ctx)
}

func (g globalFeature) DefaultContext() string {
	return DefaultContext()
}

func (g globalFeature) States(featureTags, contexts []string) (map[string]*wf.FeatureApplication, error) {
	return States(featureTags, contexts)
}

func (g globalFeature) State(tag string) (*wf.FeatureApplication, error) {
	return State(tag)
}

func (g globalFeature) Bool(tag string) bool {
	return Bool(tag)
}

func (g globalFeature) Int(tag string) int {
	return Int(tag)
}

func (g globalFeature) Str(tag string) string {
	return Str(tag)
}
