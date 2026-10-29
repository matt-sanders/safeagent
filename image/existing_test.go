package image_test

import (
	"reflect"
	"testing"

	"safeagent/docker/dockertest"
	"safeagent/image"
)

func TestListExistingParsesSafeagentRefsAndListForFilters(t *testing.T) {
	d := dockertest.New()
	d.AddImage("safeagent/s1:p1")
	d.AddImage("safeagent/s1:p2")
	d.AddImage("safeagent/s2:p1")
	d.AddImage("safeagent/latest")
	d.AddImage("other/s1:p1")

	imgs, err := image.ListExisting(d)
	if err != nil {
		t.Fatal(err)
	}
	want := []image.Existing{
		{Ref: "safeagent/s1:p1", StackID: "s1", ProfileID: "p1"},
		{Ref: "safeagent/s1:p2", StackID: "s1", ProfileID: "p2"},
		{Ref: "safeagent/s2:p1", StackID: "s2", ProfileID: "p1"},
	}
	if !reflect.DeepEqual(imgs, want) {
		t.Fatalf("ListExisting = %+v", imgs)
	}
	if got, _ := image.ListForStack(d, "s1"); !reflect.DeepEqual(got, []string{"safeagent/s1:p1", "safeagent/s1:p2"}) {
		t.Errorf("ListForStack = %v", got)
	}
	if got, _ := image.ListForProfile(d, "p1"); !reflect.DeepEqual(got, []string{"safeagent/s1:p1", "safeagent/s2:p1"}) {
		t.Errorf("ListForProfile = %v", got)
	}
}
