package units

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	tcases := []struct {
		inp      Bytes
		fmt, out string
	}{
		{0, "%d", "0"},
		{0, "%#d", "0B"},
		{123 * B, "%d", "123"},
		{123 * B, "%#d", "123B"},
		{123 * B, "%#4d", " 123B"},
		{123 * B, "%#04d", "0123B"},
		{12 * B, "%#4d", "  12B"},
		{1230 * B, "%d", "1230"},

		{12 * B, "%.2f", "12B"},
		{123 * B, "%5.2f", "  123B"},
		{1230 * B, "%.2f", "1.23kB"},
		{1230 * B, "%5.2f", " 1.23kB"},
		{1230 * B, "%.1f", "1.2kB"},
		{1230 * B, "%.3f", "1.230kB"},
		{1234 * B, "%#.2f", "1.21KiB"},
		{1234 * B, "%#5.2f", " 1.21KiB"},
		{1234 * B, "%#.1f", "1.2KiB"},
		{1234 * B, "%#.3f", "1.205KiB"},
		{1234 * B, "%+#.1f", "+1.2KiB"},
		{-1234 * B, "%#.3f", "-1.205KiB"},

		{2 * MiB, "%#.3f", "2.000MiB"},
		{3 * GiB, "%#.3f", "3.000GiB"},
		{4 * TiB, "%#.3f", "4.000TiB"},
		{5 * PiB, "%#.3f", "5.000PiB"},

		{2 * MB, "%.3f", "2.000MB"},
		{3 * GB, "%.3f", "3.000GB"},
		{4 * TB, "%.3f", "4.000TB"},
		{5 * PB, "%.3f", "5.000PB"},

		{1234 * B, "%#.4g", "1.205KiB"},
		{1234 * B, "%#.3g", "1.21KiB"},
		{1200 * B, "%.3g", "1.2kB"},
		{1200 * B, "%5.3g", "  1.2kB"},

		{KiB * KiB, "%#g", "1MiB"},
		{KiB * MiB, "%#g", "1GiB"},
		{MiB * MiB, "%#g", "1TiB"},
		{KB * KB, "%g", "1MB"},
		{KB * MB, "%g", "1GB"},
		{MB * MB, "%g", "1TB"},
		{EiB, "%.6g", "1.15292EB"},

		{12 * B, "%v", "12B"},
		{123 * B, "%v", "123B"},
		{1234 * B, "%v", "1.234kB"},
		{1200 * B, "%v", "1.2kB"},
		{1234 * B, "%#v", "bytes(1234)"},
		{1200 * B, "%#v", "bytes(1200)"},
	}

	for _, tcase := range tcases {
		actual := fmt.Sprintf(tcase.fmt, tcase.inp)
		assert.Equal(t, tcase.out, actual)
	}
}
