package urlvalues_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nahojer/urlvalues"
)

func ExampleUnmarshal() {
	req, err := http.NewRequest(http.MethodGet, "http://localhost?since=now-3m&directors=Quentin%20Tarantino&directors=Christopher%20Nolan", nil)
	if err != nil {
		panic(err)
	}

	var params struct {
		Since          time.Time `urlvalue:"since,default:now-3m,layout=2006-01-02"`
		Until          time.Time `urlvalue:"until,default:now,layout=2006-01-02"`
		Genres         []string  `urlvalue:"genres,default:action;drama"`
		Directors      []string  `urlvalue:"directors"`
		NameFilter     *string   `urlvalue:"name_filter"`
		OscarNominated *bool     `urlvalue:"oscar_nominated,default:false"`
	}
	if err := urlvalues.Unmarshal(req.URL.Query(), &params); err != nil {
		panic(err)
	}

	fmt.Printf("Genres: %s\n", strings.Join(params.Genres, ", "))
	fmt.Printf("Directors: %s\n", strings.Join(params.Directors, ", "))
	fmt.Printf("Name filter: %v\n", params.NameFilter)
	fmt.Printf("Oscar nominated: %v\n", *params.OscarNominated)
	// Output:
	// Genres: action, drama
	// Directors: Quentin Tarantino, Christopher Nolan
	// Name filter: <nil>
	// Oscar nominated: false
}

func ExampleUnmarshal_parseError() {
	data := url.Values{"meaning_of_life": {"What do I know?"}}
	var theBiggestQuestion struct {
		FortyTwo int `urlvalue:"meaning_of_life"`
	}
	if err := urlvalues.Unmarshal(data, &theBiggestQuestion); err != nil {
		var parseErr *urlvalues.ParseError
		switch {
		case errors.As(err, &parseErr):
			fmt.Println(parseErr)
			fmt.Printf("Field name: %s\n", parseErr.FieldName)
			fmt.Printf("Key: %s\n", parseErr.Key)
		default:
			panic("never reached")
		}
	}
	// Output:
	// error parsing value of meaning_of_life: strconv.ParseInt: parsing "What do I know?": invalid syntax
	// Field name: FortyTwo
	// Key: meaning_of_life
}
