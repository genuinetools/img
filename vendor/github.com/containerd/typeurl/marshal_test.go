/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package typeurl

import (
	"fmt"
	"reflect"
	"testing"
)

type TestType struct {
	ID string
}

func init() {
	Register(&TestType{}, "typeurl.Type")
}

func TestMarshalEvent(t *testing.T) {
	for _, testcase := range []struct {
		event interface{}
		url   string
	}{
		{
			event: &TestType{ID: "Test"},
			url:   "typeurl.Type",
		},
	} {
		t.Run(fmt.Sprintf("%T", testcase.event), func(t *testing.T) {
			a, err := MarshalAny(testcase.event)
			if err != nil {
				t.Fatal(err)
			}
			if a.TypeUrl != testcase.url {
				t.Fatalf("unexpected url: %q != %q", a.TypeUrl, testcase.url)
			}

			v, err := UnmarshalAny(a)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(v, testcase.event) {
				t.Fatalf("round trip failed %v != %v", v, testcase.event)
			}
		})
	}
}

func BenchmarkMarshalEvent(b *testing.B) {
	ev := &TestType{}
	expected, err := MarshalAny(ev)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		a, err := MarshalAny(ev)
		if err != nil {
			b.Fatal(err)
		}
		if a.TypeUrl != expected.TypeUrl {
			b.Fatalf("incorrect type url: %v != %v", a, expected)
		}
	}
}
