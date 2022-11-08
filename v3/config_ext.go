package luddite

import "io"

// ServiceConfigExt holds custom extensions to a service's config.
type ServiceConfigExt struct {
	ServiceLogWriter io.Writer
	AccessLogWriter  io.Writer
}
