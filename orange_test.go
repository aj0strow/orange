package orange

import (
	"net/http"
	"reflect"
	"testing"
)

func TestFromXRange(t *testing.T) {
	req, err := http.NewRequest("GET", "/items", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Range", "name ..;")
	r, err := From(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.Sort != "name" {
		t.Fatalf("wrong sort name: %s", r.Sort)
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []string{
		`name ..;`,
		`name ~Tal+Schwartz..; max=200;`,
		`name ..jack; max=20,order=desc;`,
	}
	for i, test := range tests {
		r, err := ParseString(test)
		if err != nil {
			t.Fatalf("error (%d) %s", i, err)
		}
		if test != r.String() {
			t.Errorf("bad round trip (%d) %s", i, test)
		}
	}
}

func TestRangeString(t *testing.T) {
	tests := []struct {
		Range  *Range
		String string
	}{
		{&Range{Sort: "name"}, "name ..;"},
		{&Range{Sort: "name", Start: "g"}, "name g..;"},
		{&Range{Sort: "name", Desc: true}, "name ..; order=desc;"},
		{&Range{Sort: "name", Limit: 100}, "name ..; max=100;"},
		{&Range{Sort: "name", Desc: true, Limit: 5}, "name ..; max=5,order=desc;"},
		{&Range{Sort: "name", End: "z"}, "name ..z;"},
		{&Range{Sort: "name", Start: "a", StartExclusive: true}, "name ~a..;"},
	}
	for i, test := range tests {
		if test.Range.String() != test.String {
			t.Errorf("wrong output (%d): %s %s %#v", i, test.String, test.Range, test.Range)
		}
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		Input string
		Range *Range
	}{
		{"name;", &Range{Sort: "name"}},
		{"name ..;", &Range{Sort: "name"}},
		{"name; order=desc;", &Range{Sort: "name", Desc: true}},
		{
			"name ~meredith..; max=10;",
			&Range{Sort: "name", Start: "meredith", StartExclusive: true, Limit: 10},
		},
		{
			"created_at 2016-09-21T17:12:21.649Z..2016-09-21T17:12:21.649Z;",
			&Range{Sort: "created_at", Start: "2016-09-21T17:12:21.649Z", End: "2016-09-21T17:12:21.649Z"},
		},
		{"id; max=5, order=asc;", &Range{Sort: "id", Limit: 5}},
		{"name Gary+Schwartz..;", &Range{Sort: "name", Start: "Gary Schwartz"}},
	}
	for i, test := range tests {
		r, err := ParseString(test.Input)
		if err != nil {
			t.Fatalf("error (%d) %s", i, err)
		}
		if !reflect.DeepEqual(test.Range, r) {
			t.Errorf("wrong output (%d): \n%s\n%#v", i, test.Input, r)
		}
	}
}

func TestParseProperty(t *testing.T) {
	tests := []struct {
		Input string
		Prop  *property
	}{
		{"name", &property{sort: "name"}},
		{"name ..", &property{sort: "name"}},
		{"name choi..", &property{sort: "name", start: "choi"}},
		{"name ~choi..", &property{sort: "name", start: "~choi"}},
		{"name ..john", &property{sort: "name", end: "john"}},
		{"name choi..john", &property{sort: "name", start: "choi", end: "john"}},
	}
	for i, test := range tests {
		prop, err := parseProperty(test.Input)
		if err != nil {
			t.Fatalf("error (%d) %s", i, err)
		}
		if !reflect.DeepEqual(test.Prop, prop) {
			t.Errorf("wrong output (%d): \n%s\n%#v", i, test.Input, prop)
		}
	}
}

func TestParseOptions(t *testing.T) {
	tests := []struct {
		Input string
		Opts  *options
	}{
		{"", &options{}},
		{"   ", &options{}},
		{"max=5", &options{max: "5"}},
		{"order=desc,max=5", &options{max: "5", order: "desc"}},
		{"order=asc, max=5", &options{max: "5", order: "asc"}},
	}
	for i, test := range tests {
		opts, err := parseOptions(test.Input)
		if err != nil {
			t.Fatalf("error (%d) %s", i, err)
		}
		if !reflect.DeepEqual(opts, test.Opts) {
			t.Errorf("wrong output (%d): \n%s\n%#v", i, test.Input, opts)
		}
	}
}
