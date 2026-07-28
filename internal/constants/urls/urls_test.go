package urls

import (
	"fmt"
	"testing"
)

func TestEPGURL(t *testing.T) {
	got := fmt.Sprintf(EPGURL, 1, 143)
	want := "https://jiotv.data.cdn.jio.com/apis/v1.3/getepg/get?offset=1&channel_id=143"

	if got != want {
		t.Fatalf("EPGURL formatted as %q, want %q", got, want)
	}
}
