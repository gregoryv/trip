// package trip implements a round-trip pattern for http requests
package trip

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Command struct {
	Client   Client
	Request  *http.Request
	Response *http.Response
	IsOk     func(*http.Response) bool
}

// NewCommand returns a command using the http.DefaultClient.
// By default requests that have a status code larger or equal to 400 return an error
func NewCommand(request *http.Request) (cmd *Command) {
	cmd = &Command{
		Client:  http.DefaultClient,
		Request: request,
	}
	cmd.IsOk = func(r *http.Response) bool {
		if r == nil {
			return false
		}
		return r.StatusCode < 400
	}
	return
}

// Run calls the Output method with no model
func (cmd *Command) Run() (statusCode int, err error) {
	return cmd.Output(nil)
}

// Output sends the request and does a status validation against considered status codes.
// Failing to send the request altogether results in a 599
func (cmd *Command) Output(model interface{}) (statusCode int, err error) {
	cmd.Response, err = cmd.Client.Do(cmd.Request)
	if !cmd.IsOk(cmd.Response) {
		if cmd.Response == nil {
			return 590, err
		}
		return cmd.Response.StatusCode, fmt.Errorf("%s", cmd.Response.Status)
	}
	if model != nil {
		// Default parser is json to model
		var body []byte
		body, err = ioutil.ReadAll(cmd.Response.Body)
		if err != nil {
			return 590, err
		}
		err = json.Unmarshal([]byte(body), model)
		if err != nil {
			return 590, err
		}
	}
	return cmd.Response.StatusCode, err
}

// Dump writes request and response information, if body has already been read body=true has no affect.
func (cmd *Command) Dump(w io.Writer, body bool) {
	var dump []byte
	if cmd.Request != nil {
		dump, _ = httputil.DumpRequestOut(cmd.Request, body)
		fmt.Fprintf(w, "%s\n", dump)
	}
	if cmd.Response != nil {
		dump, _ = httputil.DumpResponse(cmd.Response, body)
		fmt.Fprintf(w, "%s\n", dump)
	}
}
