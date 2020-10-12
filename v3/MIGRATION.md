# Migrating from luddite v2 to v3

* Config changes

  * YAML config is parsed strictly. Any fields that are found in the data that
    do not have corresponding struct members, or mapping keys that are
    duplicates, will result in an error.

  * `ServiceConfig.Debug.StackSize` has been removed. Stacks will always be dumped
    in their entirety.
    
  * Keeping with OpenTracing terminology, `ServiceConfig.Trace.Recorder` has
    been renamed to `ServiceConfig.Trace.Tracer`. Also, the
    `ServiceConfig.Trace.Buffer` parameter has been removed.
  
* Service method receiver changes

  * v3 restores the v1 middleware abstraction. See the type definitions in
    `handler.go`. The signature of `Service.AddHandler` has been updated
    accordingly.

  * `Service.SetRecoveryHandler` has been removed. If your service needs this
    functionality, then register a recovery middleware handler using
    `Service.AddHandler`.
    
* Tracing changes

  * The dependency on `gopkg.in/SpirentOrion/trace.v2` has been replaced with
    OpenTracing. OpenTracing supports a variety of backends, including DataDog.
    
  * v3 comes "batteries included" with JSON and YAML tracers (which write to
    local files) as well as a DataDog tracer (which, by default, writes to
    DataDog's agent on `localhost`).
    
  * Services can register their own tracers using `RegisterTracerKind` before
    calling `NewService`.
  
  * If you're creating your own trace spans in your code, there are changes
    required. You should read the [OpenTracing
    README](https://github.com/opentracing/opentracing-go) and familiarize
    yourself with their data model -- in particular, tags and logs.
  
  * Creating a new span from an `http.Request` context:
  
		parentSpan := opentracing.SpanFromContext(req.Context())
		span := parentSpan.Tracer().StartSpan("operation", opentracing.ChildOf(parentSpan.Context()))
        defer span.Finish()
  
  * Creating a new root span:
  
        span := service.Tracer().StartSpan("operation")
        defer span.Finish()
