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
	"strings"
	"testing"
)

func TestParseItemScope(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := len(data.Items[0].Properties)
	expected := 1
	if result != expected {
		t.Errorf("Result should have been \"%d\", but it was \"%d\"", expected, result)
	}
}

func TestParseItemType(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Types[0]
	expected := "http://example.com/Person"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseItemRef(t *testing.T) {
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

func TestParseItemProp(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["name"][0].(string)
	expected := "Penelope"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseItemId(t *testing.T) {
	html := `
		<ul itemscope itemtype="http://example.com/Book" itemid="urn:isbn:978-0141196404">
			<li itemprop="title">The Black Cloud</li>
			<li itemprop="author">Fred Hoyle</li>
		</ul>`

	data := ParseData(html, t)

	result := data.Items[0].ID
	expected := "urn:isbn:978-0141196404"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
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
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
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
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
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
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
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
		t.Errorf("Result should have been \"%d\", but it was \"%d\"", expected, result)
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
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseHTML(t *testing.T) {
	buf := bytes.NewBufferString(gallerySnippet)
	u, _ := url.Parse("http://blog.example.com/progress-report")
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
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestNestedItems(t *testing.T) {
	html := `
		<div>
			<div itemscope itemtype="http://example.com/Person">
				<p>My name is <span itemprop="name">Penelope</span>.</p>
				<p>I am <date itemprop="age" value="22">22 years old.</span>.</p>
				<div itemscope itemtype="http://example.com/Breadcrumb">
					<a itemprop="url" href="http://example.com/users/1"><span itemprop="title">profile</span></a>
				</div>
			</div>
		</div>`

	data := ParseData(html, t)

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://example.com/Person"],"properties":{"age":["22 years old.."],"name":["Penelope"]}},{"type":["http://example.com/Breadcrumb"],"properties":{"title":["profile"],"url":["http://example.com/users/1"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
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

func TestParseW3CBookSnippet(t *testing.T) {
	buf := bytes.NewBufferString(bookSnippet)
	u, _ := url.Parse("")
	data, err := ParseHTML(buf, "charset=utf-8", u)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://vocab.example.net/book"],"properties":{"author":["Peter F. Hamilton"],"pubdate":["1996-01-26"],"title":["The Reality Dysfunction"]},"id":"urn:isbn:0-330-34032-8"}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseW3CGalleySnippet(t *testing.T) {
	buf := bytes.NewBufferString(gallerySnippet)
	u, _ := url.Parse("")
	data, err := ParseHTML(buf, "charset=utf-8", u)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://n.whatwg.org/work"],"properties":{"license":["http://www.opensource.org/licenses/mit-license.php"],"title":["The house I found."],"work":["/images/house.jpeg"]}},{"type":["http://n.whatwg.org/work"],"properties":{"license":["http://www.opensource.org/licenses/mit-license.php"],"title":["The mailbox."],"work":["/images/mailbox.jpeg"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseW3CBlogSnippet(t *testing.T) {
	buf := bytes.NewBufferString(blogSnippet)
	u, _ := url.Parse("http://blog.example.com/progress-report")
	data, err := ParseHTML(buf, "charset=utf-8", u)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://schema.org/BlogPosting"],"properties":{"comment":[{"type":["http://schema.org/UserComments"],"properties":{"commentTime":["2013-08-29"],"creator":[{"type":["http://schema.org/Person"],"properties":{"name":["Greg"]}}],"url":["http://blog.example.com/progress-report#c1"]}},{"type":["http://schema.org/UserComments"],"properties":{"commentTime":["2013-08-29"],"creator":[{"type":["http://schema.org/Person"],"properties":{"name":["Charlotte"]}}],"url":["http://blog.example.com/progress-report#c2"]}}],"datePublished":["2013-08-29"],"headline":["Progress report"],"url":["http://blog.example.com/progress-report?comments=0"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func BenchmarkParser(b *testing.B) {
	buf := bytes.NewBufferString(blogSnippet)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u, _ := url.Parse("http://blog.example.com/progress-report")
		_, err := ParseHTML(buf, "utf-8", u)
		if err != nil && err != io.EOF {
			b.Error(err)
		}
	}
}

// This HTML snippet is taken from the W3C Working Group website at http://www.w3.org/TR/microdata.
var bookSnippet string = `
<dl itemscope
    itemtype="http://vocab.example.net/book"
    itemid="urn:isbn:0-330-34032-8">
 <dt>Title</td>
 <dd itemprop="title">The Reality Dysfunction</dd>
 <dt>Author</dt>
 <dd itemprop="author">Peter F. Hamilton</dd>
 <dt>Publication date</dt>
 <dd><time itemprop="pubdate" datetime="1996-01-26">26 January 1996</time></dd>
</dl>`

// This HTML snippet is taken from the W3C Working Group website at http://www.w3.org/TR/microdata.
var gallerySnippet string = `
<!DOCTYPE HTML>
<html>
 <head>
  <title>Photo gallery</title>
 </head>
 <body>
  <h1>My photos</h1>
  <figure itemscope itemtype="http://n.whatwg.org/work" itemref="licenses">
   <img itemprop="work" src="images/house.jpeg" alt="A white house, boarded up, sits in a forest.">
   <figcaption itemprop="title">The house I found.</figcaption>
  </figure>
  <figure itemscope itemtype="http://n.whatwg.org/work" itemref="licenses">
   <img itemprop="work" src="images/mailbox.jpeg" alt="Outside the house is a mailbox. It has a leaflet inside.">
   <figcaption itemprop="title">The mailbox.</figcaption>
  </figure>
  <footer>
   <p id="licenses">All images licensed under the <a itemprop="license"
   href="http://www.opensource.org/licenses/mit-license.php">MIT
   license</a>.</p>
  </footer>
 </body>
</html>`

// This HTML document is taken from the W3C Working Group website at http://www.w3.org/TR/microdata.
var blogSnippet string = `
<!DOCTYPE HTML>
<title>My Blog</title>
<article itemscope itemtype="http://schema.org/BlogPosting">
 <header>
  <h1 itemprop="headline">Progress report</h1>
  <p><time itemprop="datePublished" datetime="2013-08-29">today</time></p>
  <link itemprop="url" href="?comments=0">
 </header>
 <p>All in all, he's doing well with his swim lessons. The biggest thing was he had trouble
 putting his head in, but we got it down.</p>
 <section>
  <h1>Comments</h1>
  <article itemprop="comment" itemscope itemtype="http://schema.org/UserComments" id="c1">
   <link itemprop="url" href="#c1">
   <footer>
    <p>Posted by: <span itemprop="creator" itemscope itemtype="http://schema.org/Person">
     <span itemprop="name">Greg</span>
    </span></p>
    <p><time itemprop="commentTime" datetime="2013-08-29">15 minutes ago</time></p>
   </footer>
   <p>Ha!</p>
  </article>
  <article itemprop="comment" itemscope itemtype="http://schema.org/UserComments" id="c2">
   <link itemprop="url" href="#c2">
   <footer>
    <p>Posted by: <span itemprop="creator" itemscope itemtype="http://schema.org/Person">
     <span itemprop="name">Charlotte</span>
    </span></p>
    <p><time itemprop="commentTime" datetime="2013-08-29">5 minutes ago</time></p>
   </footer>
   <p>When you say "we got it down"...</p>
  </article>
 </section>
</article>`
