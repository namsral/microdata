// Copyright 2015 Lars Wiegman. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

/*

Microdata is a package for the Go programming language to extract HTML
Microdata from HTML5 documents. It depends on the
golang.org/x/net/html HTML5-compliant parser.

Usage:
	var result microdata.Result
	result, _ = microdata.ParseURL("http://example.com/blogposting")
	b, _ := json.MarshalIndent(result, "", "  ")
	os.Stdout.Write(b)

*/

package microdata

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var (
	ErrInvalidPropertyType Error = errors.New("invalid property type")
)

type Error error

type Result struct {
	Items []*Item `json:"items"`
}

type Item struct {
	Type       []string               `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Id         string                 `json:"id,omitempty"`
}

// addProperty adds the key, property to the Item. It appends to any existing
// properties associated with the key.
func (item *Item) addProperty(key string, property interface{}) error {
	switch property.(type) {
	case *Item:
		if a, ok := item.Properties[key]; ok {
			item.Properties[key] = append((a).([]*Item), (property).(*Item))
		} else {
			item.Properties[key] = []*Item{(property).(*Item)}
		}
	case string:
		if a, ok := item.Properties[key]; ok {
			item.Properties[key] = append((a).([]string), (property).(string))
		} else {
			item.Properties[key] = []string{(property).(string)}
		}
	default:
		return ErrInvalidPropertyType
	}
	return nil
}

// NewItem returns a new Item with the given itemtype(s).
func NewItem(itemtype []string) *Item {
	props := make(map[string]interface{})
	return &Item{
		Type:       itemtype,
		Properties: props,
	}
}

// Parse parses the HTML document available in the given reader and returns the
// result. The given baseURL is used to complete incomplete URLs in src and href
// attributes. The given contentType is used convert the content of r to UTF-8.
func Parse(r io.Reader, baseURL string, contentType string) (Result, error) {
	result := Result{}
	u, err := url.Parse(baseURL)

	r, err = charset.NewReader(r, contentType)
	if err != nil {
		return result, err
	}

	doc, err := html.Parse(r)
	if err != nil {
		return result, err
	}

	var item *Item
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
				i := NewItem(itemtype)
				switch {
				case len(itemid) > 0:
					i.Id = itemid
					result.Items = append(result.Items, i)
				case len(itemprop) == 0:
					result.Items = append(result.Items, i)
				case len(itemprop) > 0:
					// Might not be a valid spec
					for _, key := range itemprop {
						if err := item.addProperty(key, i); err != nil {
							return err
						}
					}
				}
				item = i
				break
			}

			// New Property
			if item != nil && len(itemprop) > 0 {
				for _, key := range itemprop {
					if len(value) < 1 {
						var s string
						var f func(*html.Node)
						f = func(n *html.Node) {
							if n.Type == html.TextNode {
								s += n.Data
							}
							for c := n.FirstChild; c != nil; c = c.NextSibling {
								f(c)
							}
						}
						f(n)
						if len(s) > 0 {
							value = s
						}
					}
					if err := item.addProperty(key, value); err != nil {
						return err
					}
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
		return result, err
	}

	return result, nil
}

// ParseURL parses the HTML document available at the given URL and returns the
// result.
func ParseURL(urlStr string) (Result, error) {
	var result Result

	resp, err := http.DefaultClient.Get(urlStr)
	if err != nil {
		return result, err
	}
	contentType := resp.Header.Get("Content-Type")

	return Parse(resp.Body, urlStr, contentType)
}
