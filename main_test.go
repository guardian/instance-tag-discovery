package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

func TestTagsFromImds(t *testing.T) {
	metadataStubURL := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tags := []string{"App", "Stack", "Stage", "gu:cdk:version"}

		if r.URL.Path == "/latest/api/token" {
			fmt.Fprintln(w, "12345")
			return
		}

		switch {
		case r.URL.Path == "/latest/api/token":
			fmt.Fprintln(w, "12345")
		case r.URL.Path == "/latest/meta-data/tags/instance":
			fmt.Fprint(w, strings.Join(tags, "\n"))
		case strings.HasPrefix(r.URL.Path, "/latest/meta-data/tags/instance/"):
			tag := strings.TrimPrefix(r.URL.Path, "/latest/meta-data/tags/instance/")
			fmt.Fprint(w, tag+"-VALUE")
		default: // individual tag
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Unknown request path: %s", r.URL.Path)
		}

	}))
	defer metadataStubURL.Close()

	client := imds.New(imds.Options{Endpoint: metadataStubURL.URL})
	got, err := tagsFromIMDS(client)
	if err != nil {
		t.Fatalf("retrieving tags failed: %v", err)
	}

	want := map[string]string{
		"App":            "App-VALUE",
		"Stack":          "Stack-VALUE",
		"Stage":          "Stage-VALUE",
		"gu:cdk:version": "gu:cdk:version-VALUE",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tags did not match; got %v, want %v", got, want)
	}
}
