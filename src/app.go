package app

import (
	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"

	"ruffnote"
)

func init() {
	if _, err := ruffnote.RegisterService(); err != nil {
		panic(err.Error())
	}
	endpoints.HandleHTTP()
}
