# urlvalues

Copyright 2022, Johan Rönkkö

Go package for unmarshalling URL values (typically query parameters and form values) 
into struct values. Below illustrates how one might use this package to unmarshall and validate 
query paramaters in a HTTP handler.

```
func (h Handler) ListNicknames(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Filter   []string  `urlvalue:"filter"`
		Until    time.Time `urlvalue:"until,default:now,layout:RFC822"`
		Since    time.Time `urlvalue:"since,default:now-3m+15d-1y,layout:2006-01-02""`
		IsActive *bool     `urlvalue:"is_active"`
	}
	if err := urlvalues.Unmarshal(r.URL.Query(), &params); err != nil {
		var parseErr *urlvalues.ParseError
		switch {
			case errors.As(err, &parseErr):
				http.Error(w, parseErr.Error(), http.StatusBadRequest)
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	// ...
}
```

## Install 
Install by running 

```go 
go get github.com/nahojer/urlvalues
```

Or just copy the souce code into your project.

## Documentation

All of the documentation can be found on the go.dev website. **[TODO: link to documentation]**

## Licensing

```
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## Credits

A special thanks to [Ardan Labs](www.ardanlabs.com) whose [conf](https://github.com/ardanlabs/conf) 
project inspired the field extracting and processing implementation of this project.
