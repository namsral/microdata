Microdata
=========

Microdata is a package for the Go programming language to extract [HTML Microdata][0] from HTML5 documents. It depends on the [golang.org/x/net/html][1] HTML5-compliant parser.

__HTML Microdata__ is a markup specification often used in combination with the [schema collection][3] to make it easier for search engines to identify and understand content on web pages. One of the most common schema is the rating you see when you google for something. Other schemas are persons, places, events, products, etc.


Installation
------------

Single binaries for Linux, macOS and Windows are available on the [release page](https://github.com/namsral/microdata/releases).

Or build from source:

```sh
$ go get -u github.com/namsral/microdata/cmd/microdata
```


Usage
-----

Parse an URL:

```sh
$ microdata https://www.gog.com/game/...
{
  "items": [
    {
      "type": [
        "http://schema.org/Product"
      ],
      "properties": {
        "additionalProperty": [
          {
            "type": [
              "http://schema.org/PropertyValue"
            ],
{
...
```


Parse HTML from the stdin:

```
$ cat saved.html |microdata
```


Format the output with a Go template to return the "price" property:

```sh
$ microdata -format '{{with index .Items 0}}{{with index .Properties "offers" 0}}{{with index .Properties "price" 0 }}{{ . }}{{end}}{{end}}{{end}}' https://www.gog.com/game/...
8.99
```


Features
--------

- Windows/BSD/Linux supported
- Format output with Go templates
- Parse from Stdin


Contribution
------------

Bug reports and feature requests are welcome. Follow GiHub's guide to [using-pull-requests](https://help.github.com/articles/using-pull-requests/)


Go Package
----------

```go
package main

import (
	"encoding/json"
	"os"

	"github.com/namsral/microdata"
)

func main() {
	var data microdata.Microdata
	data, _ = microdata.ParseURL("http://example.com/blogposting")
	b, _ := json.MarshalIndent(data, "", "  ")
	os.Stdout.Write(b)
}
```

For documentation see [godoc.org/github.com/namsral/microdata][2].

[0]: http://www.w3.org/TR/microdata
[1]: https://golang.org/x/net/html
[2]: http://godoc.org/github.com/namsral/microdata
[3]: http://schema.org
