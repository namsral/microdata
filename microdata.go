// Copyright 2015 Lars Wiegman. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

/*

	Package flag implements command-line flag parsing.
	Package microdata implements a HTML microdata parser. It depends on the
	golang.org/x/net/html HTML5-compliant parser.

	Usage:

	Pass a reader, baseURL and contentType to the Parse function.
		data, err := microdata.Parse(reader, baseURL, contentType)
		items := data.Items

	Pass an URL to the ParseURL function.
		data, _ := microdata.ParseURL("http://example.com/blogposting")
		items := data.Items
*/

package microdata

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/net/html/charset"
)

type Microdata struct {
	Items []*Item `json:"items"`
}

// addItem adds the item to the items list.
func (m *Microdata) addItem(item *Item) {
	m.Items = append(m.Items, item)
}

type ValueList []interface{}

type PropertyMap map[string]ValueList

type Item struct {
	Types      []string    `json:"type"`
	Properties PropertyMap `json:"properties"`
	Id         string      `json:"id,omitempty"`
}

// addString adds the property, value pair to the properties map. It appends to any
// existing property.
func (i *Item) addString(property, value string) {
	i.Properties[property] = append(i.Properties[property], value)
}

// addItem adds the property, value pair to the properties map. It appends to any
// existing property.
func (i *Item) addItem(property string, value *Item) {
	i.Properties[property] = append(i.Properties[property], value)
}

// addType adds the value to the types list.
func (i *Item) addType(value string) {
	i.Types = append(i.Types, value)
}

// NewItem returns a new Item.
func NewItem() *Item {
	return &Item{
		Types:      make([]string, 0),
		Properties: make(PropertyMap, 0),
	}
}

type parser struct {
	tree            *html.Node
	data            *Microdata
	baseURL         *url.URL
	identifiedNodes map[string]*html.Node
}

// parse returns the microdata from the parser's node tree.
func (p *parser) parse() (*Microdata, error) {

	toplevelNodes := []*html.Node{}

	walkNodes(p.tree, func(n *html.Node) {
		if _, ok := getAttr("itemscope", n); ok {
			if _, ok := getAttr("itemtype", n); ok {
				toplevelNodes = append(toplevelNodes, n)
			}
		}
		if id, ok := getAttr("id", n); ok {
			p.identifiedNodes[id] = n
		}
	})

	for _, node := range toplevelNodes {
		item := NewItem()
		p.data.Items = append(p.data.Items, item)
		if s, ok := getAttr("itemtype", node); ok {
			for _, itemtype := range strings.Split(s, " ") {
				if len(itemtype) > 0 {
					item.addType(itemtype)
				}
			}

			if s, ok := getAttr("itemid", node); ok {
				if u, err := p.baseURL.Parse(s); err == nil {
					item.Id = u.String()
				}
			}
		}

		if s, ok := getAttr("itemref", node); ok {
			for _, itemref := range strings.Split(s, " ") {
				if len(itemref) > 0 {
					if n, ok := p.identifiedNodes[itemref]; ok {
						p.readItem(item, n)
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			p.readItem(item, c)
		}
	}
	return p.data, nil
}

// readItem parses the given node and populates the given item with the result.
func (p *parser) readItem(item *Item, node *html.Node) {
	if itemprops, ok := getAttr("itemprop", node); ok {
		if _, ok := getAttr("itemscope", node); ok {
			subItem := NewItem()

			if itemrefs, ok := getAttr("itemref", node); ok {
				for _, itemref := range strings.Split(itemrefs, " ") {
					if len(itemref) > 0 {
						if n, ok := p.identifiedNodes[itemref]; ok {
							p.readItem(subItem, n)
						}
					}
				}
			}

			for c := node.FirstChild; c != nil; c = c.NextSibling {
				p.readItem(subItem, c)
			}

			for _, propName := range strings.Split(itemprops, " ") {
				if len(propName) > 0 {
					item.addItem(propName, subItem)
				}
			}
			return
		} else {
			var propValue string

			switch node.DataAtom {
			case atom.Meta:
				if value, ok := getAttr("content", node); ok {
					propValue = value
				}
			case atom.Audio, atom.Embed, atom.Iframe, atom.Img, atom.Source, atom.Track, atom.Video:
				if value, ok := getAttr("src", node); ok {
					if u, err := p.baseURL.Parse(value); err == nil {
						propValue = u.String()
					}
				}
			case atom.A, atom.Area, atom.Link:
				if value, ok := getAttr("href", node); ok {
					if u, err := p.baseURL.Parse(value); err == nil {
						propValue = u.String()
					}
				}
			case atom.Data, atom.Meter:
				if value, ok := getAttr("value", node); ok {
					propValue = value
				}
			case atom.Time:
				if value, ok := getAttr("datetime", node); ok {
					propValue = value
				}
			default:
				var buf bytes.Buffer
				walkNodes(node, func(n *html.Node) {
					if n.Type == html.TextNode {
						buf.WriteString(n.Data)
					}
				})
				propValue = buf.String()
			}

			if len(propValue) > 0 {
				for _, propName := range strings.Split(itemprops, " ") {
					if len(propName) > 0 {
						item.addString(propName, propValue)
					}
				}
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		p.readItem(item, c)
	}
}

// newParser returns a parser that converts the content of r to UTF-8 based on the content type of r.
func newParser(r io.Reader, contentType string, baseURL *url.URL) (*parser, error) {
	r, err := charset.NewReader(r, contentType)
	if err != nil {
		return nil, err
	}

	tree, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return &parser{
		tree:            tree,
		data:            &Microdata{},
		baseURL:         baseURL,
		identifiedNodes: make(map[string]*html.Node),
	}, nil
}

// getAttr returns the value associated with the given attribute from the given node.
func getAttr(attribute string, node *html.Node) (string, bool) {
	for _, attr := range node.Attr {
		if attribute == attr.Key {
			return attr.Val, true
		}
	}
	return "", false
}

// walkNodes traverses the node tree executing the given functions.
func walkNodes(n *html.Node, f func(*html.Node)) {
	if n != nil {
		f(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkNodes(c, f)
		}
	}
}

// ParseHTML parses the HTML document available in the given reader and returns
// the microdata. The given url is used to resolve the URLs in the
// attributes. The given contentType is used convert the content of r to UTF-8.
func ParseHTML(r io.Reader, contentType string, u *url.URL) (*Microdata, error) {
	p, err := newParser(r, contentType, u)
	if err != nil {
		return nil, err
	}
	return p.parse()
}

// ParseURL parses the HTML document available at the given URL and returns the
// microdata.
func ParseURL(urlStr string) (*Microdata, error) {
	var data *Microdata

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Get(urlStr)
	if err != nil {
		return data, err
	}

	contentType := resp.Header.Get("Content-Type")

	p, err := newParser(resp.Body, contentType, u)
	if err != nil {
		return nil, err
	}

	return p.parse()
}
