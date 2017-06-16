package feature

import (
	"c3/osm"
	"c3/osm/webframework"
	"context"
	"encoding/json"
	"testing"
)

var ctx = context.Background()

func TestNonExistentString(t *testing.T) {
	osm.SetFakeInvokeResult(`null`, nil)
	f := New(ctx)
	got := f.Str("does-not-exist")
	want := ""
	if got != want {
		t.Errorf("f.Str(`does-not-exist`) must return empty string\nwant: %v\n got: %v", want, got)
	}
}

func TestNonExistentBool(t *testing.T) {
	osm.SetFakeInvokeResult(`null`, nil)
	f := New(ctx)
	got := f.Bool("does-not-exist")
	want := false
	if got != want {
		t.Errorf("Bool(`does-not-exist`) must return false\nwant: %v\n got: %v", want, got)
	}
}

func TestNonExistentInt(t *testing.T) {
	osm.SetFakeInvokeResult(`null`, nil)
	f := New(ctx)
	got := f.Int("does-not-exist")
	want := -1
	if got != want {
		t.Errorf("Int(`does-not-exist`) must return -1\nwant: %v\n got: %v", want, got)
	}
}

func TestStates(t *testing.T) {
	// Populate fake features
	featureTags := []string{
		"a-string",
		"a-bool",
		"an-int",
	}
	contexts := []string{
		"context:1;default",
	}
	f := New(ctx)
	f.SetGlobalContext("context:1")
	fakeResponse := webframework.FeatureApplicationSlice{
		&webframework.FeatureApplication{
			FeatureTag: "a-string",
			Context:    "context:1;default",
			Maintype:   webframework.Feature_TYPE_STRING,
			Value:      "Calibri",
		},
		&webframework.FeatureApplication{
			FeatureTag: "a-bool",
			Context:    "context:1;default",
			Maintype:   webframework.Feature_TYPE_BOOLEAN,
			Value:      "true",
		},
		&webframework.FeatureApplication{
			FeatureTag: "an-int",
			Context:    "context:1;default",
			Maintype:   webframework.Feature_TYPE_INTEGER,
			Value:      "42",
		},
	}
	buf, _ := json.Marshal(fakeResponse)
	osm.SetFakeInvokeResult(string(buf), nil)

	f.States(featureTags, contexts)

	strWant := "Calibri"
	strGot := f.Str("a-string")
	if strGot != strWant {
		t.Errorf("want: %v\n got: %v", strWant, strGot)
	}

	boolWant := true
	boolGot := f.Bool("a-bool")
	if boolGot != boolWant {
		t.Errorf("want: %v\n got: %v", boolWant, boolGot)
	}

	intWant := 42
	intGot := f.Int("an-int")
	if intGot != intWant {
		t.Errorf("want: %v\n got: %v", intWant, intGot)
	}
}
