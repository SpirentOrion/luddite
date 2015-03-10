# Luddite service framework package

Luddite is a [golang][golang] package that provides a micro-framework
for RESTful web services.  It builds on top of [Negroni's][negroni]
middleware concept and makes it easy to implement services that comply
with the [Orion REST API Standards][apistds].

[golang]: http://golang.org/
[negroni]: https://github.com/codegangsta/negroni
[apistds]: https://github.com/SpirentOrion/orion-docs/blob/master/api/api-standards.md

To run the example service:

    $ cd example
    $ go build
    $ ./example

## Resources

Two types of resources are provided:

* Singleton: Supports `GET` and `PUT`.
* Collection: Supports `GET`, `POST`, `PUT`, and `DELETE`.

Resources may also be made read-only.  Since `luddite` is a
micro-framework, implementations retain substantial flexibility.

## Media Types

Content negotiation for `application/json` and `application/xml` media
types is built-in.  JSON is the default when clients do not include an
`Accept` header.

## Middleware

Currently, the framework registers several middleware handlers for each service:

* Content negotiation
* Logging
* Panic recovery

# TODO

* Configuration file
* Additional middleware handlers
  * JWT decode and validation
  * Request id generation and distributed tracing
* Metrics / statsd integration
