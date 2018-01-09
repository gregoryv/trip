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

const BadResponse = 590

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Command struct {
	Client   Client
	Request  *http.Request
	Response *http.Response
	// IsOk should return false if the response is considered wrong, according
	// to status and headers. The body should not be parsed.
	IsOk func(*http.Response) bool
	// Parse should convert the response into the given model. By default json.Unmarshal is
	// used. Parse should close the reader when done or on error.
	Parse func(io.ReadCloser, interface{}) error

	lastError error
}

// NewCommand returns a command using the http.DefaultClient.
// By default requests that have a status code larger or equal to 400 return an error
func NewCommand(request *http.Request) (cmd *Command) {
	cmd = &Command{
		Client:  http.DefaultClient,
		Request: request,
	}
	// A response that is not ok will result in a BadResponse in the execution chain
	cmd.IsOk = func(r *http.Response) bool {
		// r is never nil, that is checked for us
		return r.StatusCode < 400
	}
	cmd.Parse = parseJson
	return
}

func parseJson(body io.ReadCloser, model interface{}) (err error) {
	defer body.Close()
	// Default parser is json to model
	var buf []byte
	buf, _ = ioutil.ReadAll(body) // Ignore the error
	err = json.Unmarshal(buf, model)
	return
}

// Run, calls the Output method with no model
func (cmd *Command) Run() (statusCode int, err error) {
	return cmd.Output(nil)
}

// Output sends the request and does a status validation against considered status codes.
// Failing to send the request altogether results in a 590. Parsing errors result in 591
func (cmd *Command) Output(model interface{}) (statusCode int, err error) {
	defer func() { cmd.lastError = err }()
	cmd.Response, err = cmd.Client.Do(cmd.Request)
	if err != nil {
		return BadResponse, err
	}
	if !cmd.IsOk(cmd.Response) {
		return cmd.Response.StatusCode, fmt.Errorf("%s", cmd.Response.Status)
	}
	if model != nil {
		err = cmd.Parse(cmd.Response.Body, model)
		if err != nil {
			return 591, err
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

func (cmd *Command) Error() string {
	return cmd.lastError.Error()
}
