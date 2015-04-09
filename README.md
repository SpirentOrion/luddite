# Luddite service framework package

Luddite is a [golang][golang] package that provides a micro-framework
for RESTful web services.  It builds on top of [Negroni's][negroni]
middleware concept and makes it easy to implement services that comply
with the [Orion REST API Standards][apistds].

[golang]: http://golang.org/
[negroni]: https://github.com/codegangsta/negroni
[apistds]: https://github.com/SpirentOrion/orion-docs/blob/master/api/api-standards.md

To run the example service:

    $ make
    $ cd example
    $ ./example -c config.yaml

## Resources

Two types of resources are provided:

* Singleton: Supports `GET` and `PUT`.
* Collection: Supports `GET`, `POST`, `PUT`, and `DELETE`.

Resources may also implement `POST` actions and be made read-only.
Since `luddite` is a micro-framework, implementations retain
substantial flexibility.

## Middleware

Currently, the framework registers several middleware handlers for each service:

* Recovery: recovers from panics that occur in HTTP method handlers
  and optionally includes stack traces in 500 responses
* Trace: generates a unique request id and optionally records traces
  to a persistent backend
* Logging: optionally logs requests and responses
* Negotiation: performs JSON (default) and XML content negotiation

# TODO

* Additional middleware handlers
  * JWT decode and validation
* Metrics / statsd integration
