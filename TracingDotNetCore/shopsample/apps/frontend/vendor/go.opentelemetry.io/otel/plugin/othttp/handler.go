// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package othttp

import (
	"io"
	"net/http"

	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/propagation"
	"go.opentelemetry.io/otel/api/trace"
)

var _ http.Handler = &Handler{}

// Attribute keys that the Handler can add to a span.
const (
	HostKey       = core.Key("http.host")        // the http host (http.Request.Host)
	MethodKey     = core.Key("http.method")      // the http method (http.Request.Method)
	PathKey       = core.Key("http.path")        // the http path (http.Request.URL.Path)
	URLKey        = core.Key("http.url")         // the http url (http.Request.URL.String())
	UserAgentKey  = core.Key("http.user_agent")  // the http user agent (http.Request.UserAgent())
	RouteKey      = core.Key("http.route")       // the http route (ex: /users/:id)
	StatusCodeKey = core.Key("http.status_code") // if set, the http status
	ReadBytesKey  = core.Key("http.read_bytes")  // if anything was read from the request body, the total number of bytes read
	ReadErrorKey  = core.Key("http.read_error")  // If an error occurred while reading a request, the string of the error (io.EOF is not recorded)
	WroteBytesKey = core.Key("http.wrote_bytes") // if anything was written to the response writer, the total number of bytes written
	WriteErrorKey = core.Key("http.write_error") // if an error occurred while writing a reply, the string of the error (io.EOF is not recorded)
)

// Handler is http middleware that corresponds to the http.Handler interface and
// is designed to wrap a http.Mux (or equivalent), while individual routes on
// the mux are wrapped with WithRouteTag. A Handler will add various attributes
// to the span using the core.Keys defined in this package.
type Handler struct {
	operation string
	handler   http.Handler

	tracer           trace.Tracer
	props            propagation.Propagators
	spanStartOptions []trace.StartOption
	public           bool
	readEvent        bool
	writeEvent       bool
}

// Option function used for setting *optional* Handler properties
type Option func(*Handler)

// WithTracer configures the Handler with a specific tracer. If this option
// isn't specified then the global tracer is used.
func WithTracer(tracer trace.Tracer) Option {
	return func(h *Handler) {
		h.tracer = tracer
	}
}

// WithPublicEndpoint configures the Handler to link the span with an incoming
// span context. If this option is not provided, then the association is a child
// association instead of a link.
func WithPublicEndpoint() Option {
	return func(h *Handler) {
		h.public = true
	}
}

// WithPropagators configures the Handler with specific propagators. If this
// option isn't specified then Propagators with
// go.opentelemetry.io/otel/api/trace.DefaultHTTPPropagator are used.
func WithPropagators(ps propagation.Propagators) Option {
	return func(h *Handler) {
		h.props = ps
	}
}

// WithSpanOptions configures the Handler with an additional set of
// trace.StartOptions, which are applied to each new span.
func WithSpanOptions(opts ...trace.StartOption) Option {
	return func(h *Handler) {
		h.spanStartOptions = opts
	}
}

type event int

// Different types of events that can be recorded, see WithMessageEvents
const (
	ReadEvents event = iota
	WriteEvents
)

// WithMessageEvents configures the Handler to record the specified events
// (span.AddEvent) on spans. By default only summary attributes are added at the
// end of the request.
//
// Valid events are:
//     * ReadEvents: Record the number of bytes read after every http.Request.Body.Read
//       using the ReadBytesKey
//     * WriteEvents: Record the number of bytes written after every http.ResponeWriter.Write
//       using the WriteBytesKey
func WithMessageEvents(events ...event) Option {
	return func(h *Handler) {
		for _, e := range events {
			switch e {
			case ReadEvents:
				h.readEvent = true
			case WriteEvents:
				h.writeEvent = true
			}
		}
	}
}

// NewHandler wraps the passed handler, functioning like middleware, in a span
// named after the operation and with any provided HandlerOptions.
func NewHandler(handler http.Handler, operation string, opts ...Option) http.Handler {
	h := Handler{handler: handler, operation: operation}
	tcPropagator := trace.DefaultHTTPPropagator()
	ccPropagator := correlation.DefaultHTTPPropagator()
	props := propagation.New(
		propagation.WithInjectors(tcPropagator, ccPropagator),
		propagation.WithExtractors(tcPropagator, ccPropagator),
	)
	defaultOpts := []Option{
		WithTracer(global.TraceProvider().Tracer("go.opentelemetry.io/plugin/othttp")),
		WithPropagators(props),
		WithSpanOptions(trace.WithSpanKind(trace.SpanKindServer)),
	}

	for _, opt := range append(defaultOpts, opts...) {
		opt(&h)
	}
	return &h
}

// ServeHTTP serves HTTP requests (http.Handler)
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts := append([]trace.StartOption{}, h.spanStartOptions...) // start with the configured options

	// TODO: do something with the correlation context
	ctx := propagation.ExtractHTTP(r.Context(), h.props, r.Header)

	// not a valid span context, so no link / parent relationship to establish
	if sc := trace.RemoteSpanContextFromContext(ctx); sc.IsValid() {
		var opt trace.StartOption
		if h.public {
			// If the endpoint is a public endpoint, it should start a new trace
			// and incoming remote sctx should be added as a link.
			opt = trace.LinkedTo(sc)
			opts = append(opts, opt)
		} else { // not a private endpoint, so assume child relationship
			ctx = trace.ContextWithRemoteSpanContext(ctx, sc)
		}
	}

	ctx, span := h.tracer.Start(ctx, h.operation, opts...)
	defer span.End()

	readRecordFunc := func(int64) {}
	if h.readEvent {
		readRecordFunc = func(n int64) {
			span.AddEvent(ctx, "read", ReadBytesKey.Int64(n))
		}
	}
	bw := bodyWrapper{ReadCloser: r.Body, record: readRecordFunc}
	r.Body = &bw

	writeRecordFunc := func(int64) {}
	if h.writeEvent {
		writeRecordFunc = func(n int64) {
			span.AddEvent(ctx, "write", WroteBytesKey.Int64(n))
		}
	}

	rww := &respWriterWrapper{ResponseWriter: w, record: writeRecordFunc, ctx: ctx, props: h.props}

	// Setup basic span attributes before calling handler.ServeHTTP so that they
	// are available to be mutated by the handler if needed.
	span.SetAttributes(
		HostKey.String(r.Host),
		MethodKey.String(r.Method),
		PathKey.String(r.URL.Path),
		URLKey.String(r.URL.String()),
		UserAgentKey.String(r.UserAgent()),
	)

	h.handler.ServeHTTP(rww, r.WithContext(ctx))

	setAfterServeAttributes(span, bw.read, rww.written, int64(rww.statusCode), bw.err, rww.err)
}

func setAfterServeAttributes(span trace.Span, read, wrote, statusCode int64, rerr, werr error) {
	kv := make([]core.KeyValue, 0, 5)
	// TODO: Consider adding an event after each read and write, possibly as an
	// option (defaulting to off), so as to not create needlessly verbose spans.
	if read > 0 {
		kv = append(kv, ReadBytesKey.Int64(read))
	}
	if rerr != nil && rerr != io.EOF {
		kv = append(kv, ReadErrorKey.String(rerr.Error()))
	}
	if wrote > 0 {
		kv = append(kv, WroteBytesKey.Int64(wrote))
	}
	if statusCode > 0 {
		kv = append(kv, StatusCodeKey.Int64(statusCode))
	}
	if werr != nil && werr != io.EOF {
		kv = append(kv, WriteErrorKey.String(werr.Error()))
	}
	span.SetAttributes(kv...)
}

// WithRouteTag annotates a span with the provided route name using the
// RouteKey Tag.
func WithRouteTag(route string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		//TODO: Why doesn't tag.Upsert work?
		span.SetAttributes(RouteKey.String(route))
		h.ServeHTTP(w, r.WithContext(trace.ContextWithSpan(r.Context(), span)))
	})
}
