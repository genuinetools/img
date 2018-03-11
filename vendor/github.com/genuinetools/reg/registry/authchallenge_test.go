package registry

import (
	"reflect"
	"strings"
	"testing"
)

type authServiceMock struct {
	service string
	realm   string
	scope   []string
}

type challengeTestCase struct {
	header      string
	errorString string
	value       authServiceMock
}

func (asm authServiceMock) equalTo(v *authService) bool {
	if asm.service != v.Service {
		return false
	}
	if reflect.DeepEqual(asm.scope, v.Scope) {
		return false
	}
	if asm.realm != v.Realm.String() {
		return false
	}
	return true
}

func TestParseChallenge(t *testing.T) {
	challengeHeaderCases := []challengeTestCase{
		{
			header: `Bearer realm="https://foobar.com/api/v1/token",service=foobar.com,scope=""`,
			value: authServiceMock{
				service: "foobar.com",
				realm:   "https://foobar.com/api/v1/token",
			},
		},
	}

	for _, tc := range challengeHeaderCases {
		val, err := parseChallenge(tc.header)
		if err != nil && !strings.Contains(err.Error(), tc.errorString) {
			t.Fatalf("expected error to contain %v,  got %s", tc.errorString, err)
		}
		if !tc.value.equalTo(val) {
			t.Fatalf("got %v, expected %v", val, tc.value)
		}

	}
}
