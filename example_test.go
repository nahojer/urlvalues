package urlvalues_test

import (
	"fmt"
	"net/http"
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
		Since     time.Time `urlvalue:"since,default:now-3m,layout=2006-01-02"`
		Until     time.Time `urlvalue:"until,default:now,layout=2006-01-02"`
		Genres    []string  `urlvalue:"genres,default:action;drama"`
		Directors []string  `urlvalue:"directors"`
	}
	if err := urlvalues.Unmarshal(req.URL.Query(), &params); err != nil {
		panic(err)
	}

	fmt.Printf("Genres: %s\n", strings.Join(params.Genres, ", "))
	fmt.Printf("Directors: %s\n", strings.Join(params.Directors, ", "))
	// Output:
	// Genres: action, drama
	// Directors: Quentin Tarantino, Christopher Nolan
}
