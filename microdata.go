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
	"golang.org/x/net/html/charset"
)

type Microdata struct {
	Items []*Item `json:"items"`
}

// addItem adds the item to the items list.
func (m *Microdata) addItem(item *Item) {
	m.Items = append(m.Items, item)
}

type Item struct {
	Types      []string    `json:"type"`
	Properties PropertyMap `json:"properties"`
	Id         string      `json:"id,omitempty"`
}

type ValueList []interface{}

type PropertyMap map[string]ValueList

// addString adds the key, value pair to the properties map. It appends to any
// existing properties associated with key.
func (i *Item) addString(key, value string) {
	i.Properties[key] = append(i.Properties[key], value)
}

// addItem adds the key, value pair to the properties map. It appends to any
// existing properties associated with key.
func (i *Item) addItem(key string, value *Item) {
	i.Properties[key] = append(i.Properties[key], value)
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

// Parse parses the HTML document available in the given reader and returns the
// microdata. The given baseURL is used to complete incomplete URLs in src and
// href attributes. The given contentType is used convert the content of r to
// UTF-8.
func Parse(r io.Reader, baseURL string, contentType string) (Microdata, error) {
	data := Microdata{}
	u, err := url.Parse(baseURL)

	r, err = charset.NewReader(r, contentType)
	if err != nil {
		return data, err
	}

	doc, err := html.Parse(r)
	if err != nil {
		return data, err
	}

	item := NewItem()
	var f func(*html.Node, *Item) error
	f = func(n *html.Node, item *Item) error {
		switch n.Type {
		case html.ElementNode:
			var itemscope bool
			var itemtype, itemprop []string
			var value, itemid string

			for _, attr := range n.Attr {
				switch attr.Key {
				case "itemscope":
					itemscope = true
				case "itemtype":
					itemtype = strings.Split(attr.Val, " ")
				case "itemprop":
					itemprop = strings.Split(attr.Val, " ")
				case "src", "href":
					refURL, _ := u.Parse(attr.Val)
					value = refURL.String()
				case "content", "data", "datetime":
					value = attr.Val
				case "itemid":
					itemid = attr.Val
				}
			}

			// New Item
			if itemscope && len(itemtype) > 0 {
				i := NewItem()
				for _, v := range itemtype {
					i.addType(v)
				}
				switch {
				case len(itemid) > 0:
					i.Id = itemid
					data.addItem(i)
				case len(itemprop) == 0:
					data.addItem(i)
				case len(itemprop) > 0:
					// Might not be a valid spec
					for _, key := range itemprop {
						item.addItem(key, i)
					}
				}
				item = i
				break
			}

			// New Property
			if len(itemprop) > 0 {
				for _, key := range itemprop {
					if len(value) < 1 {
						var buf bytes.Buffer
						var f func(*html.Node)
						f = func(n *html.Node) {
							if n.Type == html.TextNode {
								buf.WriteString(n.Data)
							}
							for c := n.FirstChild; c != nil; c = c.NextSibling {
								f(c)
							}
						}
						f(n)
						if buf.Len() > 0 {
							value = buf.String()
						}
					}
					item.addString(key, value)
				}
				break
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := f(c, item); err != nil {
				return err
			}
		}
		return nil
	}
	if err := f(doc, item); err != nil {
		return data, err
	}

	return data, nil
}

// ParseURL parses the HTML document available at the given URL and returns the
// microdata.
func ParseURL(urlStr string) (Microdata, error) {
	var data Microdata

	resp, err := http.DefaultClient.Get(urlStr)
	if err != nil {
		return data, err
	}
	contentType := resp.Header.Get("Content-Type")

	return Parse(resp.Body, urlStr, contentType)
}
