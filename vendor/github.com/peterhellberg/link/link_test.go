package link

import (
	"fmt"
	"net/http"
	"testing"
)

func TestLinkString(t *testing.T) {
	l := Parse(`<https://example.com/?page=2>; rel="next"; title="foo"`)["next"]

	if got, want := l.String(), "https://example.com/?page=2"; got != want {
		t.Fatalf(`l.String() = %q, want %q`, got, want)
	}
}

func TestParseRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Set("Link", `<https://example.com/?page=2>; rel="next"`)

	g := ParseRequest(req)

	if got, want := len(g), 1; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["next"] == nil {
		t.Fatalf(`g["next"] == nil`)
	}

	if got, want := g["next"].URI, "https://example.com/?page=2"; got != want {
		t.Fatalf(`g["next"].URI = %q, want %q`, got, want)
	}

	if got, want := g["next"].Rel, "next"; got != want {
		t.Fatalf(`g["next"].Rel = %q, want %q`, got, want)
	}

	if got, want := len(ParseRequest(nil)), 0; got != want {
		t.Fatalf(`len(ParseRequest(nil)) = %d, want %d`, got, want)
	}
}

func TestParseResponse(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Link", `<https://example.com/?page=2>; rel="next"`)

	g := ParseResponse(resp)

	if got, want := len(g), 1; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["next"] == nil {
		t.Fatalf(`g["next"] == nil`)
	}

	if got, want := g["next"].URI, "https://example.com/?page=2"; got != want {
		t.Fatalf(`g["next"].URI = %q, want %q`, got, want)
	}

	if got, want := g["next"].Rel, "next"; got != want {
		t.Fatalf(`g["next"].Rel = %q, want %q`, got, want)
	}

	if got, want := len(ParseResponse(nil)), 0; got != want {
		t.Fatalf(`len(ParseResponse(nil)) = %d, want %d`, got, want)
	}
}

func TestParseHeader_single(t *testing.T) {
	h := http.Header{}
	h.Set("Link", `<https://example.com/?page=2>; rel="next"`)

	g := ParseHeader(h)

	if got, want := len(g), 1; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["next"] == nil {
		t.Fatalf(`g["next"] == nil`)
	}

	if got, want := g["next"].URI, "https://example.com/?page=2"; got != want {
		t.Fatalf(`g["next"].URI = %q, want %q`, got, want)
	}

	if got, want := g["next"].Rel, "next"; got != want {
		t.Fatalf(`g["next"].Rel = %q, want %q`, got, want)
	}
}

func TestParseHeader_multiple(t *testing.T) {
	h := http.Header{}
	h.Add("Link", `<https://example.com/?page=2>; rel="next",<https://example.com/?page=34>; rel="last"`)

	g := ParseHeader(h)

	if got, want := len(g), 2; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["next"] == nil {
		t.Fatalf(`g["next"] == nil`)
	}

	if got, want := g["next"].URI, "https://example.com/?page=2"; got != want {
		t.Fatalf(`g["next"].URI = %q, want %q`, got, want)
	}

	if got, want := g["next"].Rel, "next"; got != want {
		t.Fatalf(`g["next"].Rel = %q, want %q`, got, want)
	}

	if g["last"] == nil {
		t.Fatalf(`g["last"] == nil`)
	}

	if got, want := g["last"].URI, "https://example.com/?page=34"; got != want {
		t.Fatalf(`g["last"].URI = %q, want %q`, got, want)
	}

	if got, want := g["last"].Rel, "last"; got != want {
		t.Fatalf(`g["last"].Rel = %q, want %q`, got, want)
	}
}

func TestParseHeader_multiple_headers(t *testing.T) {
	h := http.Header{}
	h.Add("Link", `<https://example.com/?page=2>; rel="next",<https://example.com/?page=34>; rel="last"`)
	h.Add("Link", `<https://example.com/?page=foo>; rel="foo",<https://example.com/?page=bar>; rel="bar"`)

	g := ParseHeader(h)

	if got, want := len(g), 4; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["foo"] == nil {
		t.Fatalf(`g["foo"] == nil`)
	}

	if got, want := g["bar"].URI, "https://example.com/?page=bar"; got != want {
		t.Fatalf(`g["bar"].URI = %q, want %q`, got, want)
	}

	if got, want := g["next"].Rel, "next"; got != want {
		t.Fatalf(`g["next"].Rel = %q, want %q`, got, want)
	}

	if g["last"] == nil {
		t.Fatalf(`g["last"] == nil`)
	}

	if got, want := g["last"].URI, "https://example.com/?page=34"; got != want {
		t.Fatalf(`g["last"].URI = %q, want %q`, got, want)
	}

	if got, want := g["last"].Rel, "last"; got != want {
		t.Fatalf(`g["last"].Rel = %q, want %q`, got, want)
	}
}

func TestParseHeader_extra(t *testing.T) {
	h := http.Header{}
	h.Add("Link", `<https://example.com/?page=2>; rel="next"; title="foo"`)

	g := ParseHeader(h)

	if got, want := len(g), 1; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["next"] == nil {
		t.Fatalf(`g["next"] == nil`)
	}

	if got, want := g["next"].Extra["title"], "foo"; got != want {
		t.Fatalf(`g["next"].Extra["title"] = %q, want %q`, got, want)
	}
}

func TestParseHeader_noLink(t *testing.T) {
	if ParseHeader(http.Header{}) != nil {
		t.Fatalf(`Parse(http.Header{}) != nil`)
	}
}

func TestParseHeader_nilHeader(t *testing.T) {
	if ParseHeader(nil) != nil {
		t.Fatalf(`ParseHeader(nil) != nil`)
	}
}

func TestParse_emptyString(t *testing.T) {
	if Parse("") != nil {
		t.Fatalf(`Parse("") != nil`)
	}
}

func TestParse_valuesWithComma(t *testing.T) {
	g := Parse(`<//www.w3.org/wiki/LinkHeader>; rel="original latest-version",<//www.w3.org/wiki/Special:TimeGate/LinkHeader>; rel="timegate",<//www.w3.org/wiki/Special:TimeMap/LinkHeader>; rel="timemap"; type="application/link-format"; from="Mon, 03 Sep 2007 14:52:48 GMT"; until="Tue, 16 Jun 2015 22:59:23 GMT",<//www.w3.org/wiki/index.php?title=LinkHeader&oldid=10152>; rel="first memento"; datetime="Mon, 03 Sep 2007 14:52:48 GMT",<//www.w3.org/wiki/index.php?title=LinkHeader&oldid=84697>; rel="last memento"; datetime="Tue, 16 Jun 2015 22:59:23 GMT"`)

	if got, want := len(g), 5; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if got, want := g["original latest-version"].URI, "//www.w3.org/wiki/LinkHeader"; got != want {
		t.Fatalf(`g["original latest-version"].URI = %q, want %q`, got, want)
	}

	if got, want := g["last memento"].Extra["datetime"], "Tue 16 Jun 2015 22:59:23 GMT"; got != want {
		t.Fatalf(`g["last memento"].Extra["datetime"] = %q, want %q`, got, want)
	}
}

func TestParse_rfc5988Example1(t *testing.T) {
	g := Parse(`<http://example.com/TheBook/chapter2>; rel="previous"; title="previous chapter"`)

	if got, want := len(g), 1; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["previous"] == nil {
		t.Fatalf(`g["previous"] == nil`)
	}

	if got, want := g["previous"].Extra["title"], "previous chapter"; got != want {
		t.Fatalf(`g["previous"].Extra["title"] = %q, want %q`, got, want)
	}
}

func TestParse_rfc5988Example2(t *testing.T) {
	g := Parse(`</>; rel="http://example.net/foo"`)

	if got, want := len(g), 1; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["http://example.net/foo"] == nil {
		t.Fatalf(`g["http://example.net/foo"] == nil`)
	}

	l := g["http://example.net/foo"]

	if got, want := l.URI, "/"; got != want {
		t.Fatalf(`l.URI = %q, want %q`, got, want)
	}
}

func TestParse_rfc5988Example3(t *testing.T) {
	// Extended notation is not supported yet
	// g := Parse(`</TheBook/chapter2>; rel="previous"; title*=UTF-8'de'letztes%20Kapitel, </TheBook/chapter4>; rel="next"; title*=UTF-8'de'n%c3%a4chstes%20Kapitel`)
}

func TestParse_rfc5988Example4(t *testing.T) {
	// Extension relation types are ignored for now
	g := Parse(`<http://example.org/>; rel="start http://example.net/relation/other"`)

	if got, want := len(g), 1; got != want {
		t.Fatalf(`len(g) = %d, want %d`, got, want)
	}

	if g["start"] == nil {
		t.Fatalf(`g["start"] == nil`)
	}

	if got, want := g["start"].URI, "http://example.org/"; got != want {
		t.Fatalf(`g["start"].URI = %q, want %q`, got, want)
	}
}

func TestParse_fuzzCrashers(t *testing.T) {
	Parse("0")
}

func ExampleParse() {
	l := Parse(`<https://example.com/?page=2>; rel="next"; title="foo"`)["next"]

	fmt.Printf("URI: %q, Rel: %q, Extra: %+v\n", l.URI, l.Rel, l.Extra)
	// Output: URI: "https://example.com/?page=2", Rel: "next", Extra: map[title:foo]
}
