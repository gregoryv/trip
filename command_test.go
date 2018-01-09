package trip_test

import (
	"fmt"
	"github.com/gregoryv/trip"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCommand_Dump(t *testing.T) {
	request, _ := http.NewRequest("GET", "http://example.com/", nil)
	fullCmd := trip.NewCommand(request)
	fullCmd.Run()

	data := []struct {
		cmd *trip.Command
	}{
		{trip.NewCommand(nil)},
		{trip.NewCommand(request)},
		{fullCmd},
	}

	for _, d := range data {
		d.cmd.Dump(os.Stdout, false)
	}
}

func TestCommand_Output(t *testing.T) {
	data := []struct {
		body, expName string
		statusCode    int
		ok            bool
	}{
		{``, "", 200, false},
		{`{"Name":"trip"}`, "trip", 200, true},
		{`{"Name":"trip"}`, "trip", 404, false},
		{`"Name":"trip"}`, "", 200, false}, // broken json
	}

	for _, d := range data {
		// A service responding to our requests
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if d.statusCode != 200 {
				w.WriteHeader(d.statusCode)
				return
			}
			fmt.Fprintln(w, d.body)
		}))
		defer ts.Close()
		// Send request
		request, _ := http.NewRequest("GET", ts.URL, nil)
		cmd := trip.NewCommand(request)
		// Model to store response in
		var model struct{ Name string }
		_, err := cmd.Output(&model)
		// Verify
		if d.ok && model.Name != d.expName {
			t.Errorf("Output(model) should be ok for %q", d.body)
		}
		if d.ok && err != nil {
			t.Errorf("Output(model) expected to work for %q, got %s", d.body, err)
		}
		if !d.ok && err == nil {
			t.Errorf("Output(model) expected to fail for %q", d.body)
		}
	}
}

func TestCommand_Run(t *testing.T) {
	data := []struct {
		url string
		ok  bool
	}{
		{"http://badhost", false},
		{"http://localhost:1234", false},
		{"http://example.com", true},
	}

	for _, d := range data {
		request, err := http.NewRequest("GET", d.url, nil)
		fatal(t, err)
		cmd := trip.NewCommand(request)
		_, err = cmd.Run()
		if d.ok && err != nil {
			t.Errorf("GET %q expected to be ok: %s", d.url, err)
		}
		if !d.ok && err == nil {
			t.Errorf("GET %q expected fail", d.url)
		}
		if !d.ok && cmd.Error() == "" {
			t.Errorf("Error() should not return empty string")
		}
	}
}

func TestCommand_Run_brokenClient(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	fatal(t, err)
	cmd := trip.NewCommand(request)
	cmd.Client = &BrokenClient{}
	_, err = cmd.Run()
	if err == nil {
		t.Error("Run() should fail")
	}
}

func fatal(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatalf("%s", err)
	}
}

type BrokenClient struct{}

func (c *BrokenClient) Do(r *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("broken")
}
