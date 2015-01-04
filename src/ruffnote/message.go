package ruffnote

import (
	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
)

// * parseTag parses "endpoints" field tag into endpointsTag struct.
//   <https://github.com/GoogleCloudPlatform/go-endpoints/blob/master/endpoints/apiconfig.go#L643>
//
//       type MyMessage struct {
//           SomeField int `endpoints:"req,min=0,max=100,desc="Int field"`
//           WithDefault string `endpoints:"d=Hello gopher"`
//       }
//
//   - req, required (boolean)
//   - d=val, default value
//   - min=val, min value
//   - max=val, max value
//   - desc=val, description

type ItemResponseMessage struct {
	Id   string `json:"id" endpoints:"req"`
	Name string `json:"name" endpoints:"req"`
}

type ItemsRequestMessage struct {
	Limit int `json:"limit" endpoints:"d=10,min=1,max=100"`
}

type ItemsResponseMessage struct {
	Items []*ItemResponseMessage `json:"items"`
}

type AvailabileRequestMessage struct {
	Name string `json:"name" endpoints:"req"`
}

type AvailabileResponseMessage struct {
	Name   string `json:"name" endpoints:"req"`
	Status bool   `json:"status" endpoints:"req"`
}

type CreateRequestMessage struct {
	Name string `json:"name" endpoints:"req"`
}

func (message CreateRequestMessage) Validate() (err error) {
	if message.Name == "" {
		err = endpoints.NewBadRequestError("require name")
	}
	return
}

type ReNameRequestMessage struct {
	Id   string `json:"id" endpoints:"req"`
	Name string `json:"name" endpoints:"req"`
}

type DeleteRequestMessage struct {
	Id string `json:"id" endpoints:"req"`
}

type DeleteResponseMessage struct {
}
