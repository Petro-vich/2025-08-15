package downloader

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("testdata"))
	}))
	defer server.Close()
	tmpfile, err := ioutil.TempFile("", "downloaded_*.txt")
	if err != nil {
		t.Fatalf("TempFile error: %v", err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())
	if err := DownloadFile(server.URL, tmpfile.Name()); err != nil {
		t.Errorf("DownloadFile failed: %v", err)
	}
	data, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(data) != "testdata" {
		t.Errorf("Downloaded data mismatch: got %q", string(data))
	}
}
