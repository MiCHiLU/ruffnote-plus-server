package ruffnote

import (
	"fmt"
	"strings"

	"appengine"
	"appengine/datastore"
	"appengine/user"
	"github.com/mjibson/goon"
)

const (
	ITEM_KIND = "Item"
)

type Item struct {
	key     *datastore.Key
	keyname string `datastore:"-" goon:"id"`

	User string `datastore:"user"`
	id   string `datastore:"-"`
	name string `datastore:"-"`
}

func (s *Item) parseKey() {
	if s.key != nil {
		s.keyname = s.key.StringID()
	}
	if s.keyname != "" {
		splits := strings.SplitN(s.keyname, "/", 2)
		if len(splits) == 2 {
			s.name, s.id = splits[0], splits[1]
		}
	}
	return
}

func (s *Item) Id() string {
	if s.id == "" {
		s.parseKey()
	}
	return s.id
}

func (s *Item) Name() string {
	if s.name == "" {
		s.parseKey()
	}
	return s.name
}

func (s *Item) toMessage(msg *ItemResponseMessage) *ItemResponseMessage {
	if msg == nil {
		msg = &ItemResponseMessage{}
	}
	msg.Id = s.Id()
	msg.Name = s.Name()
	return msg
}

func (s *Item) put(c appengine.Context) (err error) {
	if s.keyname == "" {
		s.keyname = fmt.Sprintf("%s/%s", s.Name(), s.Id())
	}
	g := goon.FromContext(c)
	_, err = g.Put(s)
	return
}

func newItem(id string, name string, u *user.User) *Item {
	return &Item{
		id:   id,
		name: name,
		User: userId(u),
	}
}

func GetItemByName(c appengine.Context, name string) (item *Item, err error) {
	start := datastore.NewKey(c, ITEM_KIND, fmt.Sprintf("%s/", name), 0, nil)
	end := datastore.NewKey(c, ITEM_KIND, fmt.Sprintf("%s/\uffff", name), 0, nil)
	q := datastore.NewQuery(ITEM_KIND).
		Filter("__key__ >=", start).
		Filter("__key__ <=", end).
		Limit(1).
		KeysOnly()
	g := goon.FromContext(c)
	keys, err := g.GetAll(q, nil)
	if err != nil {
		return
	}
	if len(keys) == 0 {
		err = datastore.ErrNoSuchEntity
		return
	}
	item = new(Item)
	item.key = keys[0]
	return
}

func getItemsByUser(c appengine.Context, u *user.User, limit int) (items []*Item, err error) {
	q := datastore.NewQuery(ITEM_KIND).
		Filter("user =", userId(u)).
		Limit(limit).
		KeysOnly()
	g := goon.FromContext(c)
	keys, err := g.GetAll(q, nil)
	if err != nil {
		return
	}
	items = make([]*Item, 0, len(keys))
	for i, key := range keys {
		items = append(items, &Item{})
		items[i].key = key
	}
	return
}

func userId(u *user.User) string {
	return u.String()
}
