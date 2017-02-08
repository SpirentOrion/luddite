negotiation [![Build Status](https://travis-ci.org/K-Phoen/negotiation.svg?branch=master)](https://travis-ci.org/K-Phoen/negotiation)
===========

**negotiation** is a standalone library that allows you to implement [content negotiation](http://www.w3.org/Protocols/rfc2616/rfc2616-sec12.html)
in your application, whatever framework you use.
This provides specific functions to negotiate `Accept` and `Accept-Language`
headers, although any kind of header can be parsed.

## Status

This project is **DEPRECATED** and should NOT be used. 

If someone magically appears and wants to maintain this project, I'll gladly give access to this repository.

## Usage

### Language negotiation

```go
package main

import (
  "fmt"
  "github.com/K-Phoen/negotiation"
)

func main() {
  language, err := negotiation.NegotiateLanguage("da, en-gb;q=0.8, en;q=0.7", []string{"es", "fr", "en"})

  if err != nil {
    fmt.Println("Unable to negotiate the language")
  }

  fmt.Println("Negotiated language: ", language.Value) // outputs: "Negotiated language: en"
}
```

### Format negotiation

```go
package main

import (
  "fmt"
  "github.com/K-Phoen/negotiation"
)

func main() {
  negotiation.RegisterFormat("html", []string{"text/html", "application/xhtml+xml"})
  format, err := negotiation.NegotiateAccept("application/rdf+xml;q=0.5,text/html;q=.3", []string{"html", "application/xml"})

  if err != nil {
    fmt.Println("Unable to negotiate the format")
  }

  fmt.Println("Negotiated format: " + format.Name + " (" + format.Value + ")") // outputs: "Negotiated format: html (text/html)"
}
```

## Tests

```bash
$ go test
```

## Credits

  * Author: [Kévin Gomez](https://github.com/K-Phoen/)
  * [William Durand](https://github.com/willdurand/) and it's [Negotiation PHP library](https://github.com/willdurand/Negotiation/),
    from whom I borrowed this library's name and a few ideas.

## License

This library is released under the MIT License. See the bundled LICENSE file for
details.
