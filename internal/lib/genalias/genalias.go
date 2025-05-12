package genalias

import (
	"context"
	"fmt"
	"github.com/kxddry/url-shortener/internal/lib/random"
	"time"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func GenerateAlias(length int, urlGetter URLGetter) (string, error) {
	const op = "lib.genalias.GenerateAlias"
	alias := random.NewRandomString(length)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, err := urlGetter.GetURL(alias); err == nil; _, err = urlGetter.GetURL(alias) {
		if ctx.Err() != nil {
			return "", fmt.Errorf("%s: %s", op, "couldn't generate alias: time ran out")
		}
		alias = random.NewRandomString(length)
	}
	return alias, nil
}
