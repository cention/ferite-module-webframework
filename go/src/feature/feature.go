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
// Subsequent query of the feature value using State, Str, Bool and Int using
// either uses the cached feature.
package feature

import (
	"log"
	wf "osm/webframework"
	"strings"
)

var DefaultContexts []string
var DefaultGlobalContext = ""
var featureCache = make(map[string]*wf.FeatureApplication)

// ClearCache clears the cached features that were loaded from the object server.
func ClearCache() {
	featureCache = make(map[string]*wf.FeatureApplication)
}

// Preload loads all the feature for the current default contexts.
func Preload() {
	States([]string{}, DefaultContexts)
}

// SetGlobalContext sets the default global context to the given context.
func SetGlobalContext(ctx string) {
	DefaultGlobalContext = ctx
}

// SetDefaultContexts sets the default contexts to the given contexts.
func SetDefaultContexts(ctxs []string) {
	DefaultContexts = ctxs
}

// SetDefaultContext sets the default context to the given context string. More
// than one context can be specified by separating them with semicolon.
func SetDefaultContext(ctx string) {
	SetDefaultContexts(strings.Split(ctx, ";"))
}

// DefaultContext returns the semicolon-separated default contexts.
func DefaultContext() string {
	return DefaultGlobalContext + ";" + wf.FeatureApplication_DEFAULT_CONTEXT
}

// States implements the Go equivalent of webframework/Core/Features.feh's
// state().
func States(featureTags []string, contexts []string) (map[string]*wf.FeatureApplication, error) {
	// TODO rename processList to byPriorityDesc
	var processList wf.FeatureApplicationSlice
	var err error
	var featureList []string
	// TODO rename list to featureMap
	list := map[string]*wf.FeatureApplication{}

	contextList := []string{DefaultContext()}
	for _, v := range contexts {
		contextList = append(contextList, DefaultGlobalContext+`;`+v)
	}

	if len(featureTags) == 0 {
		processList, err = wf.QueryFeatureApplicationFetchInContext(contextList)
		if err != nil {
			return nil, err
		}
	} else {
		for _, tag := range featureTags {
			if fa, exists := featureCache[tag]; exists {
				list[tag] = fa
			} else {
				featureList = append(featureList, tag)
			}
		}
		if len(featureList) > 0 {
			processList, err = wf.QueryFeatureApplicationFetchByFeaturesInContext(featureList, contextList)
			if err != nil {
				// FIXME don't panic
				panic(err)
			}
		}
	}

	for _, featureApplication := range processList {
		if _, exists := list[featureApplication.FeatureTag]; !exists {
			list[featureApplication.FeatureTag] = featureApplication
		}
	}

	if len(featureTags) == 0 {
		featureCache = list
	} else {
		for _, tag := range featureList {
			//fmt.Printf("%8v => %v\n", tag, list[tag].Val())
			if v, exists := list[tag]; exists {
				featureCache[tag] = v
			}
		}
	}

	return list, nil
}

func stateForTagWithDefaultContexts(tag string) (*wf.FeatureApplication, error) {
	m, err := States([]string{tag}, DefaultContexts)
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
func State(tag string) (*wf.FeatureApplication, error) {
	if v, exists := featureCache[tag]; exists {
		return v, nil
	}
	fa, err := stateForTagWithDefaultContexts(tag)
	if err != nil {
		return nil, err
	}
	if fa != nil {
		return fa, nil
	}
	return nil, nil
}

// Bool returns the boolean value of the feature application state for the
// given tag.
func Bool(tag string) bool {
	fa, err := State(tag)
	if err != nil {
		log.Printf("feature.Bool(`%v`): %v", tag, err)
	}
	return fa != nil && fa.Bool()
}

// Int returns the integer value of the feature application state for the given
// tag.
func Int(tag string) int {
	fa, err := State(tag)
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
func Str(tag string) string {
	fa, err := State(tag)
	if err != nil {
		log.Printf("feature.Str(`%v`): %v", tag, err)
	}
	if fa == nil {
		return ""
	}
	return fa.Str()
}
