# Cookie Monster (aka cm)

This sublibrary (someday perhaps available separately, but for now, not) is
convenience wrapper for handling client state in server (fully in the
cookie).

For now, the state is not encrypted at all (although we could), as I do not
really see the point. The idea is not to store secrets using this, just
e.g. state in various views so that the client resumes same state
implicitly when they go back to a view.


## Short form

Define state as normal struct, e.g.

```
type State struct {
  Number int `json:"n" cm:"number"`
}
```

and then call

```
err = cm.Run(r, w, &state)
```

Where
- `r`: `http.Request`
- `w`: `http.ResponseWriter`
- `state`: arbitrary struct with cm + json tags (ideally empty but this will clear most of the state)

The struct is examined through reflection and `cm:"field"` entries are
CRUDed based on `"field"` in the `http.Request.Form`. They are persisted
automagically to cookie derived from the struct name, and with the `json`
tagged names (preferably short ones, as cookie space is limited). If json
definitions are not supplied, they will not be imported/exported.

## Supported datatypes

Currently the code supports only following datatypes:

- bool
- int*
- string
- uint*

Other types, if tagged, cause an error.

## What does it do?

The above cm.Run call will:

- read existing json cookie with name `cm.<module>.State` and unmarshal
  JSON content from it to the State struct
- reads query parameter 'number', and if it is set, update `state` instance
- (.. same with other fields too, if any ..)
- if any of the fields were updated, set the updated cookie for the
  response

## TODO

- Perhaps make the configuration also accept Path? For now, cookies are
  scoped to Path "/" so e.g. global state can be used to convey them; this
  may not be optimal always.

- Document this better?
