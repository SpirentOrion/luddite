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

  * Creating a new span from an `http.Request` context:
  
		parentSpan := opentracing.SpanFromContext(req.Context())
		span := parentSpan.Tracer().StartSpan("operation", opentracing.ChildOf(parentSpan.Context()))
  
  * Creating a new root span:
  
        span := service.Tracer().StartSpan("operation")

