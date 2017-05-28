// Package feature provides access to the system features tables.
//
// These features are represented by the webframework.osm classes Feature,
// FeatureGroup and FeatureApplication.
//
// Features supported by the system are organized in groups such as
// "Administrator", "Area", "Errand Handling", etc. These are provided by the
// osm classes FeatureGroup and Feature. FeatureApplication represents the
// feature value.
//
// Features are associated with context. Context is a colon-separated pair of
// strings.
//
// Features can be queried via tags. Once loaded, features are cached.
// Subsequent query of the feature value using State, Str, Bool and Int
// will use the cached feature.
package feature

import (
	"c3/osm/webframework"
	"context"
	"log"
	"strings"
)

type Boolean interface {
	Bool(string) bool
}

type Number interface {
	Int(string) bool
}

type String interface {
	Str(string) bool
}

type All interface {
	Bool(string) bool
	Int(string) bool
	Str(string) bool
}

type Feature struct {
	defaultContexts      []string
	defaultGlobalContext string
	featureCache         map[string]*webframework.FeatureApplication
}

func New() *Feature {
	return &Feature{featureCache: map[string]*webframework.FeatureApplication{}}
}

// ClearCache clears the cached features that were loaded from the object server.
func (f *Feature) ClearCache() {
	f.featureCache = make(map[string]*webframework.FeatureApplication)
}

// Preload loads all the feature for the current default contexts.
func (f *Feature) Preload() {
	f.States([]string{}, f.defaultContexts)
}

// SetGlobalContext sets the default global context to the given context.
func (f *Feature) SetGlobalContext(ctx string) {
	f.defaultGlobalContext = ctx
}

func (f *Feature) GetGlobalContext() string {
	return f.defaultGlobalContext
}

// SetDefaultContexts sets the default contexts to the given contexts.
func (f *Feature) SetDefaultContexts(ctxs []string) {
	f.defaultContexts = ctxs
}

func (f *Feature) GetDefaultContexts() []string {
	return f.defaultContexts
}

// SetDefaultContext sets the default context to the given context string. More
// than one context can be specified by separating them with semicolon.
func (f *Feature) SetDefaultContext(ctx string) {
	f.SetDefaultContexts(strings.Split(ctx, ";"))
}

// DefaultContext returns the semicolon-separated default contexts.
func (f *Feature) DefaultContext() string {
	return f.defaultGlobalContext + ";" + webframework.FeatureApplication_DEFAULT_CONTEXT
}

// States implements the Go equivalent of webframework/Core/Features.feh's
// state().
func (f *Feature) States(featureTags []string, contexts []string) (map[string]*webframework.FeatureApplication, error) {
	// TODO rename processList to byPriorityDesc
	var processList webframework.FeatureApplicationSlice
	var err error
	var featureList []string
	// TODO rename list to featureMap
	list := map[string]*webframework.FeatureApplication{}

	contextList := []string{f.DefaultContext()}
	for _, v := range contexts {
		contextList = append(contextList, f.defaultGlobalContext+`;`+v)
	}

	if len(featureTags) == 0 {
		processList, err = webframework.QueryFeatureApplication_fetchInContextContext(context.TODO(), contextList)
		if err != nil {
			return nil, err
		}
	} else {
		for _, tag := range featureTags {
			if fa, exists := f.featureCache[tag]; exists {
				list[tag] = fa
			} else {
				featureList = append(featureList, tag)
			}
		}
		if len(featureList) > 0 {
			processList, err = webframework.QueryFeatureApplication_fetchByFeaturesInContextContext(context.TODO(), featureList, contextList)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, featureApplication := range processList {
		if _, exists := list[featureApplication.FeatureTag]; !exists {
			list[featureApplication.FeatureTag] = featureApplication
		}
	}

	if len(featureTags) == 0 {
		f.featureCache = list
	} else {
		for _, tag := range featureList {
			//fmt.Printf("%8v => %v\n", tag, list[tag].Val())
			if v, exists := list[tag]; exists {
				f.featureCache[tag] = v
			}
		}
	}

	return list, nil
}

func (f *Feature) stateForTagWithDefaultContexts(tag string) (*webframework.FeatureApplication, error) {
	m, err := f.States([]string{tag}, f.defaultContexts)
	if err != nil {
		return nil, err
	}
	for _, v := range m {
		return v, nil
	}
	return nil, nil
}

// State returns the (possibly cached) feature application state for the given
// tag.
func (f *Feature) State(tag string) (*webframework.FeatureApplication, error) {
	if v, exists := f.featureCache[tag]; exists {
		return v, nil
	}
	return f.stateForTagWithDefaultContexts(tag)
}

// Bool returns the boolean value of the feature application state for the
// given tag.
func (f *Feature) Bool(tag string) bool {
	fa, err := f.State(tag)
	if err != nil {
		log.Printf("feature.Bool(`%v`): %v", tag, err)
	}
	return fa != nil && fa.Bool()
}

// Int returns the integer value of the feature application state for the given
// tag.
func (f *Feature) Int(tag string) int {
	fa, err := f.State(tag)
	if err != nil {
		log.Printf("feature.Int(`%v`): %v", tag, err)
	}
	if fa == nil {
		return -1
	}
	return fa.Int()
}

// Str returns the string value of the feature application state for the given
// tag.
func (f *Feature) Str(tag string) string {
	fa, err := f.State(tag)
	if err != nil {
		log.Printf("feature.Str(`%v`): %v", tag, err)
	}
	if fa == nil {
		return ""
	}
	return fa.Str()
}
