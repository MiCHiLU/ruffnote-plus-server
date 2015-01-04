package ruffnote

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/user"
	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
	"github.com/mjibson/goon"
)

const (
	clientId = "" //TODO
	debug    = true
)

var (
	scopes    = []string{endpoints.EmailScope}
	clientIds = []string{clientId, endpoints.APIExplorerClientID}
	audiences = []string{clientId}
)

func RegisterService() (rpcService *endpoints.RPCService, err error) {
	api := &RuffnoteApi{}
	rpcService, err = endpoints.RegisterService(api,
		"ruffnote_plus", "v1", "ruffnote+", true)
	if err != nil {
		return
	}

	info := rpcService.MethodByName("Items").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"item.list", "GET", "items", "List items."
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences

	info = rpcService.MethodByName("Availabile").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"item.availabile", "GET", "item", ""
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences

	info = rpcService.MethodByName("Create").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"item.create", "POST", "item", ""
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences

	info = rpcService.MethodByName("ReName").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"item.rename", "PUT", "item", ""
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences

	info = rpcService.MethodByName("Delete").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"item.delete", "DELETE", "item", ""
	info.Scopes, info.ClientIds, info.Audiences = scopes, clientIds, audiences

	return
}

func makeEndpointsError(c endpoints.Context, err error) error {
	switch err.(type) {
	case *endpoints.APIError:
	default:
		if debug {
			c.Debugf("%s: %s", fmt.Sprintln(runtime.Caller(1)), err.Error())
		}
		message := err.Error()
		if appengine.IsCapabilityDisabled(err) {
			err = endpoints.NewInternalServerError(message)
			c.Errorf(message)
		} else if appengine.IsOverQuota(err) {
			err = endpoints.NewInternalServerError(message)
			c.Errorf(message)
		} else if appengine.IsTimeoutError(err) {
			err = endpoints.NewInternalServerError(message)
			c.Errorf(message)
		} else if strings.Contains(message, "500") {
			err = endpoints.InternalServerError
			c.Errorf(message)
		} else if strings.Contains(message, "400") {
			err = endpoints.NewBadRequestError(message)
		} else if strings.Contains(message, "401") {
			err = endpoints.NewUnauthorizedError(message)
		} else if strings.Contains(message, "403") {
			err = endpoints.NewForbiddenError(message)
		} else if strings.Contains(message, "404") {
			err = endpoints.NewNotFoundError(message)
		} else if strings.Contains(message, "409") {
			err = endpoints.NewConflictError(message)
		} else {
			err = endpoints.BadRequestError
			c.Errorf(message)
		}
	}
	return err
}

type RuffnoteApi struct {
}

func (ruffnoteApi *RuffnoteApi) Items(r *http.Request,
	req *ItemsRequestMessage, resp *ItemsResponseMessage) (err error) {

	c := endpoints.NewContext(r)
	u, err := getCurrentUser(c)
	if err != nil {
		err = makeEndpointsError(c, err)
		return
	}

	items, err := getItemsByUser(c, u, req.Limit)
	if err != nil {
		err = makeEndpointsError(c, err)
		return
	}
	resp.Items = make([]*ItemResponseMessage, len(items))
	for i, item := range items {
		resp.Items[i] = item.toMessage(nil)
	}
	return
}

func (ruffnoteApi *RuffnoteApi) Availabile(r *http.Request,
	req *AvailabileRequestMessage, resp *AvailabileResponseMessage) (err error) {

	c := endpoints.NewContext(r)
	_, err = getCurrentUser(c)
	if err != nil {
		err = makeEndpointsError(c, err)
		return
	}

	resp.Name = req.Name
	_, err = GetItemByName(c, req.Name)
	if err == datastore.ErrNoSuchEntity {
		err = nil
		resp.Status = true
		return
	}
	return
}

func (ruffnoteApi *RuffnoteApi) Create(r *http.Request,
	req *CreateRequestMessage, resp *ItemResponseMessage) (err error) {

	c := endpoints.NewContext(r)
	u, err := getCurrentUser(c)
	if err != nil {
		err = makeEndpointsError(c, err)
		return
	}

	err = req.Validate()
	if err != nil {
		err = makeEndpointsError(c, err)
		return
	}

	item, err := createItem(c, u, "id", req.Name)
	if err != nil {
		err = makeEndpointsError(c, err)
		return
	}
	item.toMessage(resp)
	return
}

func (ruffnoteApi *RuffnoteApi) ReName(r *http.Request,
	req *ReNameRequestMessage, resp *ItemResponseMessage) (err error) {

	c := endpoints.NewContext(r)
	u, err := getCurrentUser(c)
	if err != nil {
		err = makeEndpointsError(c, err)
		return err
	}

	//TODO check id
	id := req.Id
	item, err := createItem(c, u, id, req.Name)
	if err != nil {
		err = makeEndpointsError(c, err)
		return
	}
	item.toMessage(resp)

	items, err := getItemsByUser(c, u, 100)
	if err != nil {
		c.Errorf("%v", err)
		err = makeEndpointsError(c, err)
		return
	}
	deleteItemsKey := make([]*datastore.Key, 0, len(items))
	for i, item := range items {
		if item.Id() != id {
			continue
		}
		if item.Name() == req.Name {
			continue
		}
		deleteItemsKey[i] = item.key
	}
	g := goon.FromContext(c)
	err = g.DeleteMulti(deleteItemsKey)
	if err != nil {
		c.Warningf("%v", err)
		err = makeEndpointsError(c, err)
	}
	return
}

func (ruffnoteApi *RuffnoteApi) Delete(r *http.Request,
	req *DeleteRequestMessage, resp *DeleteResponseMessage) (err error) {

	c := endpoints.NewContext(r)
	u, err := getCurrentUser(c)
	if err != nil {
		err = makeEndpointsError(c, err)
		return
	}

	items, err := getItemsByUser(c, u, 100)
	if err != nil {
		c.Errorf("%v", err)
		err = makeEndpointsError(c, err)
		return
	}
	deleteItemsKey := make([]*datastore.Key, 0, len(items))
	for i, item := range items {
		if item.Id() != req.Id {
			continue
		}
		deleteItemsKey[i] = item.key
	}
	g := goon.FromContext(c)
	err = g.DeleteMulti(deleteItemsKey)
	if err != nil {
		c.Warningf("%v", err)
		err = makeEndpointsError(c, err)
	}
	return
}

func getCurrentUser(c endpoints.Context) (u *user.User, err error) {
	u, err = endpoints.CurrentUser(c, scopes, audiences, clientIds)
	if err != nil {
		c.Infof("%v", err)
		err = endpoints.UnauthorizedError
		return
	}
	if u == nil {
		err = endpoints.UnauthorizedError
		return
	}
	return
}

func createItem(c endpoints.Context, u *user.User, id string, name string) (item *Item, err error) {
	item = newItem(id, name, u)

	g := goon.FromContext(c)
	err = g.RunInTransaction(func(tg *goon.Goon) (err error) {
		_, err = GetItemByName(c, name)
		if err != datastore.ErrNoSuchEntity {
			err = endpoints.ConflictError
			return
		}
		err = item.put(c)
		if err != nil {
			c.Errorf("%v", err)
			err = makeEndpointsError(c, err)
		}
		return
	}, nil)

	if err != datastore.ErrConcurrentTransaction {
		c.Errorf("Transaction failed: %v", err)
	}
	return
}
