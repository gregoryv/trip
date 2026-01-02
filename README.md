ARCHIVED! 

[trip](https://godoc.org/github.com/gregoryv/trip) - implements a round-trip pattern for http requests

## Quick start

    go get github.com/gregoryv/trip

## The round-trip pattern

A round-trip pattern is basically

1. Prepare trip
2. Execute
3. Optionally parse response

Other descriptions that fit this pattern would be remote procedure
call(RPC), which http requests are of sorts. With this package the
steps inbetween are abstracted and you get to write more go idomatic
code. I designed this package in a way that resembles `os.exec`

Prepare trip

	request := http.NewRequest("GET", "/", nil)
	cmd := trip.NewCommand(request)

Do the trip

    statusCode, err := cmd.Run()
	// or if you want the response parsed
	err := cmd.Output(&model)

## When to use

When you talk to remote services and need to only vary parts of the
flow, ie.  an API has changed and requires a new parameter, then you
only have to modify the part that builds your request. Hopefully it's
easier to maintain a backwards compatible client for a constantly
changing remote service.
