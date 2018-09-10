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

package runc

import "testing"

func TestParseVersion(t *testing.T) {
	const data = `runc version 1.0.0-rc3
commit: 17f3e2a07439a024e54566774d597df9177ee216
spec: 1.0.0-rc5-dev`

	v, err := parseVersion([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if v.Runc != "1.0.0-rc3" {
		t.Errorf("expected runc version 1.0.0-rc3 but received %s", v.Runc)
	}
	if v.Commit != "17f3e2a07439a024e54566774d597df9177ee216" {
		t.Errorf("expected commit 17f3e2a07439a024e54566774d597df9177ee216 but received %s", v.Commit)
	}
	if v.Spec != "1.0.0-rc5-dev" {
		t.Errorf("expected spec version 1.0.0-rc5-dev but received %s", v.Spec)
	}
}
