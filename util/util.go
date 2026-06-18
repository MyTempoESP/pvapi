package util

import (
	"fmt"
	"maps"

	"linhazero.com/pvapi/env"
)

// Sybau 🥀😭
func Sybau(vals ...any) {}

type Show[R any] interface {
	ToResource() R
}

func ToResourceCollection[M any, R any](
	collection []M,
	toResource func(M) R,
) []R {
	resources := make([]R, len(collection))
	for i, item := range collection {
		resources[i] = toResource(item)
	}
	return resources
}

func ResourceCollection[R any, M Show[R]](collection []M) []R {
	return ToResourceCollection(collection, M.ToResource)
}

func Merge(args ...Args) (merge Args) {
	merge = make(Args)
	for _, a := range args {
		maps.Copy(merge, a)
	}
	return
}

func Coalesce[T any](nullable *T, fallback T) T {
	if nullable != nil {
		return *nullable
	}
	return fallback
}

func GetDomainURL() (url string, err error) {
	domain, err := env.DomainName()
	if err != nil {
		return
	}

	protocol := "http"
	if env.HTTPSEnabled {
		protocol = "https"
	}

	url = fmt.Sprintf(
		"%s://%s",
		protocol, domain,
	)
	return
}

func GetImageURL(base string) (url string, err error) {
	domain, err := GetDomainURL()
	if err != nil {
		return
	}

	url = fmt.Sprintf(
		"%s/public/uploads/%s",
		domain, base,
	)
	return
}
