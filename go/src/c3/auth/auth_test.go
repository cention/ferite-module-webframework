package auth

import (
	"fmt"
	"testing"
)

func TestRemoveSpace(t *testing.T) {
	for i, tt := range []struct {
		username string
		want     string
	}{
		{
			username: "bob",
			want:     "bob",
		},
		{
			username: "bob@foo",
			want:     "bob",
		},
		{
			username: "bob@foo@bar",
			want:     "bob@foo",
		},
		{
			username: "bob@",
			want:     "bob",
		},
		{
			username: "@",
			want:     "",
		},
	} {
		got := removeSpace(tt.username)
		if got != tt.want {
			t.Errorf("%d removeSpace(%s)\nwant %s\n got %s", i, tt.username, tt.want, got)
		} else {
			fmt.Printf("Okay removeSpace(%q) => %q\n", tt.username, got)
		}
	}
}
