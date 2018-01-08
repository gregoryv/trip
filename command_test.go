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
	}{
		{`{"Name":"trip"}`, "trip"},
		{`"Name":"trip"}`, ""}, // broken json
	}

	for _, d := range data {
		// A service responding to our requests
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, d.body)
		}))
		defer ts.Close()
		// Send request
		request, _ := http.NewRequest("GET", ts.URL, nil)
		cmd := trip.NewCommand(request)
		// Model to store response in		
		var model struct{ Name string }
		cmd.Output(&model)
		// Verify
		if model.Name != d.expName {
			t.Errorf("Output(model) should unmarshal the json response")
		}
	}
}

func TestCommand_Run(t *testing.T) {
	data := []struct {
		url           string
		expStatusCode int
	}{
		{"http://badhost", http.StatusServiceUnavailable},		
		{"http://localhost:1234", 590},
		{"http://example.com", http.StatusOK},
	}

	for _, d := range data {
		request, err := http.NewRequest("GET", d.url, nil)
		fatal(t, err)
		cmd := trip.NewCommand(request)
		statusCode, err := cmd.Run()
		if d.expStatusCode != statusCode {
			t.Errorf("GET %q expected to return statusCode %v, got %v: %s", d.url, d.expStatusCode, statusCode, err)
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
