# Cookie Monster #

This sublibrary (someday perhaps available separately, but for now, not) is
convenience wrapper for handling client state in client (fully in the
cookie). For now, the state is not encrypted at all (although we could), as
I do not really see the point. The idea is not to store secrets using this,
just e.g. state in various views so that the client resumes same state
implicitly when they go back to a view.

## Short form

```
err = monster.Run(r, w, &state)
```

Where
```
r = http.Request
w = http.ResponseWriter
state = arbitrary struct (ideally empty)
```

The struct is examined through reflection and `cm:"field"` entries are
CRUDed based on `"field"` in the `http.Request`. Note that json definitions
must be also supplied for the fields, or otherwise they will not be
imported/exported.
