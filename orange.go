// Package orange encodes and decodes HTTP range headers.
package orange

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const Version = "0.0.2"

// Accept sets the HTTP Accept-Ranges response header.
func Accept(w http.ResponseWriter, props ...string) {
	w.Header().Set("Accept-Ranges", strings.Join(props, ", "))
}

// Next sets the HTTP Next-Range response header.
func Next(w http.ResponseWriter, r *Range) {
	w.Header().Set("Next-Range", r.String())
}

// From parses the HTTP Range request header.
// It returns an empty Range struct if no header is provided.
func From(r *http.Request) (*Range, error) {
	s := r.Header.Get("Range")
	if s == "" {
		s = r.Header.Get("X-Range")
	}
	if s == "" {
		return &Range{}, nil
	}
	return ParseString(s)
}

// ParseString parses a range header into a Range struct.
// It returns InvalidFormat if the parser fails.
func ParseString(s string) (*Range, error) {
	return parseRange(s)
}

var (
	InvalidFormat = errors.New("range header invalid format")
)

// Range struct represents a parsed range header.
type Range struct {
	Sort           string
	Start          string
	StartExclusive bool
	End            string
	Limit          int
	Desc           bool
}

// String encodes a Range struct into the HTTP header format.
func (r *Range) String() string {
	var buf bytes.Buffer
	buf.WriteString(url.QueryEscape(r.Sort))
	buf.WriteRune(' ')
	if len(r.Start) > 0 {
		if r.StartExclusive {
			buf.WriteRune('~')
		}
		buf.WriteString(url.QueryEscape(r.Start))
	}
	buf.WriteString("..")
	if len(r.End) > 0 {
		buf.WriteString(url.QueryEscape(r.End))
	}
	buf.WriteRune(';')
	opts := []string{}
	if r.Limit > 0 {
		opts = append(opts, "max="+strconv.Itoa(r.Limit))
	}
	if r.Desc {
		opts = append(opts, "order=desc")
	}
	if len(opts) > 0 {
		buf.WriteRune(' ')
		buf.WriteString(strings.Join(opts, ","))
		buf.WriteRune(';')
	}
	return buf.String()
}

// Next takes the sort property of the last element in the current
// results to generate the next Range start.
func (r *Range) Next(prev string) *Range {
	return &Range{
		Sort:           r.Sort,
		Start:          prev,
		StartExclusive: true,
		End:            r.End,
		Limit:          r.Limit,
		Desc:           r.Desc,
	}
}

func parseRange(s string) (*Range, error) {
	r := &Range{}
	scanner := bufio.NewReader(strings.NewReader(s))
	prop, err := scanner.ReadString(';')
	if err != nil {
		return nil, InvalidFormat
	}
	prop = strings.TrimSuffix(prop, ";")
	p, err := parseProperty(prop)
	if err != nil {
		return nil, err
	}
	r.Sort, err = url.QueryUnescape(p.sort)
	if err != nil {
		return nil, err
	}
	if len(p.start) > 0 {
		if p.start[0] == '~' {
			r.Start, err = url.QueryUnescape(p.start[1:])
			if err != nil {
				return nil, err
			}
			r.StartExclusive = true
		} else {
			r.Start, err = url.QueryUnescape(p.start)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(p.end) > 0 {
		r.End, err = url.QueryUnescape(p.end)
		if err != nil {
			return nil, err
		}
	}
	opts, err := scanner.ReadString(';')
	if err != nil {
		return r, nil
	}
	opts = strings.TrimSuffix(opts, ";")
	o, err := parseOptions(opts)
	if err != nil {
		return nil, err
	}
	if o.max != "" {
		max, err := strconv.Atoi(o.max)
		if err != nil {
			return nil, InvalidFormat
		}
		r.Limit = max
	}
	if o.order != "" {
		switch o.order {
		case "desc":
			r.Desc = true
		case "asc":
			r.Desc = false
		default:
			return nil, InvalidFormat
		}
	}
	return r, nil
}

type property struct {
	sort  string
	start string
	end   string
}

func parseProperty(s string) (*property, error) {
	r := bufio.NewScanner(strings.NewReader(s))
	r.Split(bufio.ScanWords)
	if ok := r.Scan(); !ok {
		return nil, InvalidFormat
	}
	p := &property{
		sort: r.Text(),
	}
	if ok := r.Scan(); !ok {
		return p, nil
	}
	vs := strings.Split(r.Text(), "..")
	if len(vs) > 2 {
		return nil, InvalidFormat
	}
	p.start = vs[0]
	if len(vs) == 2 {
		p.end = vs[1]
	}
	return p, nil
}

type options struct {
	max   string
	order string
}

func parseOptions(s string) (*options, error) {
	o := &options{}
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return o, nil
	}
	opts := strings.Split(s, ",")
	for _, opt := range opts {
		kv := strings.Split(strings.TrimSpace(opt), "=")
		if len(kv) != 2 {
			return nil, InvalidFormat
		}
		k, v := kv[0], kv[1]
		switch k {
		case "max":
			o.max = v
		case "order":
			o.order = v
		default:
			return nil, InvalidFormat
		}
	}
	return o, nil
}
