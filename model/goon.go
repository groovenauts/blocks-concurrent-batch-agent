package model

import (
	"golang.org/x/net/context"

	"github.com/mjibson/goon"
)

var ModelNameToKindMap = map[string]string{
	"Organization": "Organizations",
	"Auth":         "Auths",
}

func GoonFromContext(c context.Context) *goon.Goon {
	r := GoonFromContext(c)
	baseResolver := r.KindNameResolver
	r.KindNameResolver = func(src interface{}) string {
		base := baseResolver(src)
		mapped := ModelNameToKindMap[base]
		if mapped != "" {
			return mapped
		}
		return base
	}
	return r
}
