# Luddite Service Framework Package, Version 2

Luddite is a [golang][golang] package that provides a micro-framework for
RESTful web services. It is built around extensible, pluggable middleware layers
and includes a flexible resource abstraction that makes it easy to implement
services that comply with the [Orion REST API Standards][apistds].

[golang]: http://golang.org/
[apistds]: https://github.com/SpirentOrion/orion-api/blob/master/doc/api/api-standards.md

To run the example service:

    $ make all
    $ cd example
    $ ./example -c config.yaml

## Request Handling

The basic request handling built into `luddite` combines CORS, tracing, logging,
metrics, profiling, and recovery actions.

Tracing generates a unique request id and optionally records traces to a file or
persistent backend. The framework currently uses `v2` of the
[trace](https://github.com/SpirentOrion/trace/tree/v2) package.

Logging is based on [logrus](https://github.com/sirupsen/logrus). A service log
is established for general use. An access log is maintained separately. Both use
structured JSON logging.

[Prometheus](https://prometheus.io/) metrics provide basic request/response
stats. By default, the metrics endpoint is served on `/metrics`.

The standard [net/http/pprof](https://golang.org/pkg/net/http/pprof/) profiling
handlers may be optionally enabled. These are served on `/debug/pprof`.

Recovery handles panics that occur in resource handlers and optionally includes
stack traces in `500` responses.

## Request Middleware

Currently, `luddite` registers two middleware handlers for each service:

* Negotiation: Performs JSON (default) and XML content negotiation
  based on HTTP requests' `Accept` headers.

* Version: Performs API version selection and enforces the service's min/max
  supported version constraints.  Makes the selected API version available
  to resource handlers as part of the request [context][context].

[context]: http://blog.golang.org/context

Implementations are free to register their own additional middleware handlers in
addition to these two.

## Resource Abstraction

Generally, each resource falls into one of two categories.

* Collection: Supports `GET`, `POST`, `PUT`, and `DELETE`.
* Singleton: Supports `GET` and `PUT`.

The framework defines several interfaces that establish its resource
abstraction. For collection-style resources:

* `CollectionLister` returns all elements in response to `GET /resource`.
* `CollectionCounter` returns a count of its elements in response to `GET /resource/all/count`.
* `CollectionGetter` returns a specific element in response to `GET /resource/:id`.
* `CollectionCreator` creates a new element in response to `POST /resource`.
* `CollectionUpdater` updates a specific element in response to `PUT /resource/:id`.
* `CollectionDeleter` deletes a specific element in response to `DELETE /resource/:id`.
  It may also optionally delete the entire collection in response to `DELETE /resource`
* `CollectionActioner` executes an action in response to `POST /resource/:id/:action`.

And for singleton-style resources:

* `SingletonGetter` returns a response to `GET /resource`.
* `SingletonUpdater` is updated in response to `PUT /resource`.
* `SingletonActioner` executes an action in response to `POST /resource/:action`.

Routes are automatically created for resource handler types that implement these
interfaces. However, since `luddite` is a framework, implementations retain
substantial flexibility to register their own routes if these are not
sufficient.

## Resource Versioning

The framework allows implementations to support multiple API versions
simultaneously. In addition to API version selection via middleware, the
framework also allows for version-specific resource registration.

Typically, implementations define a separate resource handler type for each API
version. The routes for each type are registered in a version-specific router.
Since route lookup occurs after version negotiation, each router is free to
handle requests without further consideration of API version.
