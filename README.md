# Luddite service framework package

Luddite is a [golang][golang] package that provides a micro-framework
for RESTful web services.  It is built around extensible, pluggable
middleware layers and includes a standard `Resource` interface that
makes it easy to implement services that comply with the
[Orion REST API Standards][apistds].

[golang]: http://golang.org/
[apistds]: https://github.com/SpirentOrion/orion-api/blob/master/doc/api/api-standards.md

To run the example service:

    $ make all
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

Currently, the framework registers several middleware handlers for
each service:

* Bottom: Combines CORS, tracing, logging, metrics, and recovery
  actions. Tracing generates a unique request id and optionally
  records traces to a persistent backend.  Logging logs
  requests/responses in a structured JSON format.  Metrics
  increments basic request/response stats.  Recovery handles panics
  that occur in HTTP method handlers and optionally includes stack
  traces in 500 responses.  Also makes the `Service` instance,
  request id, and response headers available to resource handlers as
  part of the request [context][context].
* Negotiation: Performs JSON (default) and XML content negotiation
  based on HTTP requests' `Accept` headers.
* Version: Performs API version selection and enforces the service's min/max
  supported version constraints.  Makes the selected API version available
  to resource handlers as part of the request [context][context].
[context]: http://blog.golang.org/context

## TODO

Need to document these:
* Metrics
* Schema file serving
* TLS requirements
* Min/max versioning
