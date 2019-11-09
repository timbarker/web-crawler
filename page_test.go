package main

import (
	"errors"
	"net/url"
	"testing"
)

func TestPageString(t *testing.T) {
	type fields struct {
		location *url.URL
		links    []*url.URL
		err      error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"No Links", fields{URL("https://bbc.co.uk"), []*url.URL{}, nil}, "Page{location:https://bbc.co.uk, links: [], err: <nil>}"},
		{"One Link", fields{URL("https://bbc.co.uk"), []*url.URL{URL("/one")}, nil}, "Page{location:https://bbc.co.uk, links: [/one], err: <nil>}"},
		{"Many Links", fields{URL("https://bbc.co.uk"), []*url.URL{URL("/one"), URL("/two")}, nil}, "Page{location:https://bbc.co.uk, links: [/one /two], err: <nil>}"},
		{"Error", fields{URL("https://bbc.co.uk"), []*url.URL{}, errors.New("some error")}, "Page{location:https://bbc.co.uk, links: [], err: some error}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{
				location: tt.fields.location,
				links:    tt.fields.links,
				err:      tt.fields.err,
			}
			if got := page.String(); got != tt.want {
				t.Errorf("Page.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
