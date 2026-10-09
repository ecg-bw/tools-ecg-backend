package swagger

import (
	_ "embed"
)

//go:embed openapi.yaml
var openAPISpec []byte

func GetOpenAPISpec() []byte {
	return openAPISpec
}