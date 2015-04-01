Microdata
=========

Microdata is a package for the Go programming language to extract [HTML Microdata][0] from HTML5 documents. It depends on the [golang.org/x/net/html][1] HTML5-compliant parser.

__HTML Microdata__ is a markup specification often used in combination with the [schema collection][3] to make it easier for search engines to identify and understand content on web pages. One of the most common schema is the rating you see when you google for something. Other schemas are persons, places, events, products, etc.


Command Line
------------

The command line utility returns a JSON object containing any items in the given HTML document.

__Install:__

```sh
$ go install github.com/namsral/microdata/microdata
```

__Run:__

```sh
$ microdata http://example.com/blogposting
{
  "items": [
    {
      "type": [
        "http://schema.org/BlogPosting"
      ],
      "properties": {
        "comment": [
          {
            "type": [
              "http://schema.org/UserComments"
            ],
            "properties": {
...
```


Library
-------

```go
package main

import (
	"encoding/json"
	"os"

	"github.com/namsral/microdata"
)

func main() {
	var result microdata.Result
	result, _ = microdata.ParseURL("http://example.com/blogposting")
	b, _ := json.MarshalIndent(result, "", "  ")
	os.Stdout.Write(b)
}
```

For documentation see [godoc.org/github.com/namsral/microdata][2].

[0]: http://www.w3.org/TR/microdata
[1]: https://golang.org/x/net/html
[2]: http://godoc.org/github.com/namsral/microdata
[3]: http://schema.org

License
-------

Copyright (c) 2015 Lars Wiegman. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
