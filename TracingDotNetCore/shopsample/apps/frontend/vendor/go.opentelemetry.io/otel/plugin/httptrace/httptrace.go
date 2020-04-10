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

package httptrace

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/key"
	"go.opentelemetry.io/otel/api/propagation"
	"go.opentelemetry.io/otel/api/trace"
)

var (
	HostKey = key.New("http.host")
	URLKey  = key.New("http.url")

	tcPropagator = trace.DefaultHTTPPropagator()
	ccPropagator = correlation.DefaultHTTPPropagator()
	propagators  = propagation.New(
		propagation.WithInjectors(tcPropagator, ccPropagator),
		propagation.WithExtractors(tcPropagator, ccPropagator),
	)
)

// Returns the Attributes, Context Entries, and SpanContext that were encoded by Inject.
func Extract(ctx context.Context, req *http.Request) ([]core.KeyValue, []core.KeyValue, core.SpanContext) {
	ctx = propagation.ExtractHTTP(ctx, propagators, req.Header)

	attrs := []core.KeyValue{
		URLKey.String(req.URL.String()),
		// Etc.
	}

	var correlationCtxKVs []core.KeyValue
	correlation.MapFromContext(ctx).Foreach(func(kv core.KeyValue) bool {
		correlationCtxKVs = append(correlationCtxKVs, kv)
		return true
	})

	return attrs, correlationCtxKVs, trace.RemoteSpanContextFromContext(ctx)
}

func Inject(ctx context.Context, req *http.Request) {
	propagation.InjectHTTP(ctx, propagators, req.Header)
}
