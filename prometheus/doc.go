// Copyright 2014 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package prometheus provides embeddable metric primitives for servers and
// standardized exposition of telemetry through a web services interface.
//
// All exported functions and methods are safe to be used concurrently unless
// specified otherwise.
//
// To expose metrics registered with the default registry, you have to register
// prometheus.Handler with your http server. The usual endpoint is "/metrics".
//
//     http.Handle("/metrics", prometheus.Handler())
//
// See the various examples for more details.
//
// TODO: Add a proper tutorial, basic use and advanced use.
package prometheus
