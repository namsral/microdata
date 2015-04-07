// Copyright 2015 Lars Wiegman. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package microdata

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestItemScope(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := len(data.Items[0].Properties)
	expected := 1
	if result != expected {
		t.Errorf("Result should have been \"%d\", but it was \"%d\"", result, expected)
	}
}

func TestItemType(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Types[0]
	expected := "http://example.com/Person"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", result, expected)
	}
}

func TestItemRef(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Movie" itemref="properties">
			<p><span itemprop="name">Rear Window</span> is a movie from 1954.</p>
		</div>
		<ul id="properties">
			<li itemprop="genre">Thriller</li>
			<li itemprop="description">A homebound photographer spies on his neighbours.</li>
		</ul>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"genre", "Thriller"},
		{"description", "A homebound photographer spies on his neighbours."},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestItemProp(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["name"][0].(string)
	expected := "Penelope"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", result, expected)
	}
}

func TestParseHref(t *testing.T) {
	html := `
		<html itemscope itemtype="http://example.com/Person">
			<head>
				<link itemprop="linkTest" href="http://example.com/cde">
			<head>
			<div>
				<a itemprop="aTest" href="http://example.com/abc" /></a>
				<area itemprop="areaTest" href="http://example.com/bcd" />
			</div>
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"aTest", "http://example.com/abc"},
		{"areaTest", "http://example.com/bcd"},
		{"linkTest", "http://example.com/cde"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseSrc(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Videocast">
			<audio itemprop="audioTest" src="http://example.com/abc" />
			<embed itemprop="embedTest" src="http://example.com/bcd" />
			<iframe itemprop="iframeTest" src="http://example.com/cde"></iframe>
			<img itemprop="imgTest" src="http://example.com/def" />
			<source itemprop="sourceTest" src="http://example.com/efg" />
			<track itemprop="trackTest" src="http://example.com/fgh" />
			<video itemprop="videoTest" src="http://example.com/ghi" />
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"audioTest", "http://example.com/abc"},
		{"embedTest", "http://example.com/bcd"},
		{"iframeTest", "http://example.com/cde"},
		{"imgTest", "http://example.com/def"},
		{"sourceTest", "http://example.com/efg"},
		{"trackTest", "http://example.com/fgh"},
		{"videoTest", "http://example.com/ghi"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseMetaContent(t *testing.T) {
	html := `
		<html itemscope itemtype="http://example.com/Person">
			<meta itemprop="length" content="1.70" />
		</html>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["length"][0].(string)
	expected := "1.70"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", result, expected)
	}
}

func TestParseValue(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Container">
			<data itemprop="capacity" value="80">80 liters</data>
			<meter itemprop="volume" min="0" max="100" value="25">25%</meter>
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"capacity", "80"},
		{"volume", "25"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseDatetime(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<time itemprop="birthDate" datetime="1993-10-02">22 years</time>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["birthDate"][0].(string)
	expected := "1993-10-02"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", result, expected)
	}
}

func TestParseText(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Product">
			<span itemprop="price">3.95</span>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["price"][0].(string)
	expected := "3.95"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", result, expected)
	}
}

func TestParseMultiItemTypes(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Park http://example.com/Zoo">
			<span itemprop="name">ZooParc Overloon</span>
		</div>`

	data := ParseData(html, t)

	result := len(data.Items[0].Types)
	expected := 2
	if result != expected {
		t.Errorf("Result should have been \"%d\", but it was \"%d\"", result, expected)
	}
}

func TestJSON(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
			<p>I am <date itemprop="age" value="22">22 years old.</span>.</p>
		</div>`

	data := ParseData(html, t)

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://example.com/Person"],"properties":{"age":["22 years old.."],"name":["Penelope"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", result, expected)
	}
}

func TestParseHTML(t *testing.T) {
	buf := newTestBuffer()
	u, _ := url.Parse("http://example.com")
	contentType := "charset=utf-8"

	_, result := ParseHTML(buf, contentType, u)
	if result != nil {
		t.Errorf("Result should have been nil, but it was \"%s\"", result)
	}
}

func TestParseURL(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer ts.Close()

	data, err := ParseURL(ts.URL)
	if err != nil {
		t.Error(err)
	}

	result := data.Items[0].Properties["name"][0].(string)
	expected := "Penelope"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", result, expected)
	}
}

func BenchmarkParser(b *testing.B) {
	buf := newTestBuffer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u, _ := url.Parse("http://example.com")
		_, err := ParseHTML(buf, "utf-8", u)
		if err != nil && err != io.EOF {
			b.Error(err)
		}
	}
}

func ParseData(html string, t *testing.T) *Microdata {
	r := strings.NewReader(html)
	u, _ := url.Parse("http://example.com")

	p, err := newParser(r, "utf-8", u)
	if err != nil {
		t.Error(err)
	}

	data, err := p.parse()
	if err != nil {
		t.Error(err)
	}
	return data
}

func newTestBuffer() *bytes.Buffer {
	f, err := os.Open("./testdata/blogposting.html")
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(f)
	if err != nil {
		panic(err)
	}
	f.Close()
	return buf
}
