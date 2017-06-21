package feature

import (
	"c3/osm"
	"c3/osm/webframework"
	"context"
)

type gofeature struct {
	*Feature
	ch chan struct{}
}

// threaded feature is thread aware feature.
func NewTF() Featurer {
	return &gofeature{
		Feature: New(osm.NewContext(context.Background(), "objsrv")),
		ch:      make(chan struct{}, 1),
	}
}

// ClearCache clears the cached features that were loaded from the object server.
func (t *gofeature) ClearCache() {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	t.Feature.ClearCache()
}

func (t *gofeature) Preload() {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	t.Feature.Preload()
}

func (t *gofeature) SetDefaultContexts(ctxs []string) {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	t.Feature.SetDefaultContexts(ctxs)
}

// SetDefaultContext sets the default context to the given context string. More
// than one context can be specified by separating them with semicolon.
func (t *gofeature) SetDefaultContext(ctx string) {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	t.Feature.SetDefaultContext(ctx)
}

// SetGlobalContext sets the default global context to the given context.
func (t *gofeature) SetGlobalContext(ctx string) {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	t.Feature.SetGlobalContext(ctx)
}

func (t *gofeature) DefaultContext() string {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	return t.Feature.DefaultContext()
}

func (t *gofeature) States(featureTags, contexts []string) (map[string]*webframework.FeatureApplication, error) {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	return t.Feature.States(featureTags, contexts)
}

func (t *gofeature) State(tag string) (*webframework.FeatureApplication, error) {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	return t.Feature.State(tag)
}

// Bool returns the boolean value of the feature application state for the
// given tag.
func (t *gofeature) Bool(tag string) bool {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	return t.Feature.Bool(tag)
}

// Int returns the integer value of the feature application state for the given
// tag.
func (t *gofeature) Int(tag string) int {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	return t.Feature.Int(tag)
}

// Str returns the string value of the feature application state for the given
// tag.
func (t *gofeature) Str(tag string) string {
	t.ch <- struct{}{}
	defer func() { <-t.ch }()
	return t.Feature.Str(tag)
}
