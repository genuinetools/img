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
