package feature

import (
	"encoding/json"
	"osm"
	"osm/webframework"
	"testing"
)

func TestNonExistentString(t *testing.T) {
	osm.SetFakeInvokeResult(`null`, nil)
	got := Str("does-not-exist")
	want := ""
	if got != want {
		t.Errorf("Str(`does-not-exist`) must return empty string\nwant: %v\n got: %v", want, got)
	}
}

func TestNonExistentBool(t *testing.T) {
	osm.SetFakeInvokeResult(`null`, nil)
	got := Bool("does-not-exist")
	want := false
	if got != want {
		t.Errorf("Bool(`does-not-exist`) must return false\nwant: %v\n got: %v", want, got)
	}
}

func TestNonExistentInt(t *testing.T) {
	osm.SetFakeInvokeResult(`null`, nil)
	got := Int("does-not-exist")
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
	SetGlobalContext("context:1")
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

	States(featureTags, contexts)

	strWant := "Calibri"
	strGot := Str("a-string")
	if strGot != strWant {
		t.Errorf("want: %v\n got: %v", strWant, strGot)
	}

	boolWant := true
	boolGot := Bool("a-bool")
	if boolGot != boolWant {
		t.Errorf("want: %v\n got: %v", boolWant, boolGot)
	}

	intWant := 42
	intGot := Int("an-int")
	if intGot != intWant {
		t.Errorf("want: %v\n got: %v", intWant, intGot)
	}
}
