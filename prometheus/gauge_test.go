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

package prometheus

import (
	"flag"
	"math"
	"math/rand"
	"sync"
	"testing"
	"testing/quick"
)

func ExampleGauge() {
	delOps := NewGauge(GaugeOpts{
		Namespace: "our_company",
		Subsystem: "blob_storage",
		Name:      "deletes",
		Help:      "How many delete operations we have conducted against our blob storage system.",
	})
	MustRegister(delOps)

	delOps.Set(900) // That's all, folks!
}

func ExampleGaugeVec() {
	binaryVersion := flag.String("binary_version", "debug", "Version of the binary: debug, canary, production.")
	flag.Parse()

	delOps := NewGaugeVec(
		GaugeOpts{
			Namespace:   "our_company",
			Subsystem:   "blob_storage",
			Name:        "deletes",
			Help:        "How many delete operations we have conducted against our blob storage system, partitioned by data corpus and qos.",
			ConstLabels: Labels{"binary_version": *binaryVersion},
		},
		[]string{
			// What is the body of data being deleted?
			"corpus",
			// How urgently do we need to delete the data?
			"qos",
		},
	)
	MustRegister(delOps)

	// Set a sample value using compact (but order-sensitive!) WithLabelValues().
	delOps.WithLabelValues("profile-pictures", "immediate").Set(4)
	// Set a sample value with a map using WithLabels. More verbose, but
	// order doesn't matter anymore.
	delOps.With(Labels{"qos": "lazy", "corpus": "cat-memes"}).Set(1)
}

func listenGaugeStream(vals, result chan float64, done chan struct{}) {
	var sum float64
outer:
	for {
		select {
		case <-done:
			close(vals)
			for v := range vals {
				sum += v
			}
			break outer
		case v := <-vals:
			sum += v
		}
	}
	result <- sum
	close(result)
}

func TestGaugeConcurrency(t *testing.T) {
	it := func(n uint32) bool {
		mutations := int(n % 10000)
		concLevel := int((n % 15) + 1)

		start := &sync.WaitGroup{}
		start.Add(1)
		end := &sync.WaitGroup{}
		end.Add(concLevel)

		sStream := make(chan float64, mutations*concLevel)
		result := make(chan float64)
		done := make(chan struct{})

		go listenGaugeStream(sStream, result, done)
		go func() {
			end.Wait()
			close(done)
		}()

		gge := NewGauge(GaugeOpts{
			Name: "test_gauge",
			Help: "no help can be found here",
		})
		for i := 0; i < concLevel; i++ {
			vals := make([]float64, 0, mutations)
			for j := 0; j < mutations; j++ {
				vals = append(vals, rand.Float64()-0.5)
			}

			go func(vals []float64) {
				start.Wait()
				for _, v := range vals {
					sStream <- v
					gge.Add(v)
				}
				end.Done()
			}(vals)
		}
		start.Done()

		if expected, got := <-result, gge.(*value).val; math.Abs(expected-got) > 0.000001 {
			t.Fatalf("expected approx. %f, got %f", expected, got)
			return false
		}
		return true
	}

	if err := quick.Check(it, nil); err != nil {
		t.Fatal(err)
	}
}

func TestGaugeVecConcurrency(t *testing.T) {
	it := func(n uint32) bool {
		mutations := int(n % 10000)
		concLevel := int((n % 15) + 1)
		vecLength := int((n % 5) + 1)

		start := &sync.WaitGroup{}
		start.Add(1)
		end := &sync.WaitGroup{}
		end.Add(concLevel)

		sStreams := make([]chan float64, vecLength)
		results := make([]chan float64, vecLength)
		done := make(chan struct{})

		for i, _ := range sStreams {
			sStreams[i] = make(chan float64, mutations*concLevel)
			results[i] = make(chan float64)
			go listenGaugeStream(sStreams[i], results[i], done)
		}

		go func() {
			end.Wait()
			close(done)
		}()

		gge := NewGaugeVec(
			GaugeOpts{
				Name: "test_gauge",
				Help: "no help can be found here",
			},
			[]string{"label"},
		)
		for i := 0; i < concLevel; i++ {
			vals := make([]float64, 0, mutations)
			for j := 0; j < mutations; j++ {
				vals = append(vals, rand.Float64()-0.5)
			}

			go func(vals []float64) {
				start.Wait()
				for _, v := range vals {
					i := rand.Intn(vecLength)
					sStreams[i] <- v
					gge.WithLabelValues(string('A' + i)).Add(v)
				}
				end.Done()
			}(vals)
		}
		start.Done()

		for i, _ := range sStreams {
			if expected, got := <-results[i], gge.WithLabelValues(string('A'+i)).(*value).val; math.Abs(expected-got) > 0.000001 {
				t.Fatalf("expected approx. %f, got %f", expected, got)
				return false
			}
		}
		return true
	}

	if err := quick.Check(it, nil); err != nil {
		t.Fatal(err)
	}
}
