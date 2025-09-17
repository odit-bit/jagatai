package telebot

import (
	"time"

	"github.com/odit-bit/jagatai/api"
)

type ChatCache struct {
	m map[int64]*StoredChat
}

func NewCache() *ChatCache {
	return &ChatCache{m: map[int64]*StoredChat{}}
}

func (cc *ChatCache) Get(ID int64) *StoredChat {
	sc, ok := cc.m[ID]
	if !ok {
		return &StoredChat{
			id:      ID,
			updated: time.Now(),
			store:   cc.m,
		}
	}
	return sc
}

func (cc *ChatCache) Clear(id int64) error {
	delete(cc.m, id)
	return nil
}

func (cc *ChatCache) CountMessages(id int64) int {
	sc := cc.Get(id)
	return len(sc.messages)
}

// func (cc *ChatCache) NotifC(ctx context.Context) <-chan int64 {
// 	idC := make(chan int64, 1)

// 	go func() {
// 		ticker := time.NewTicker(5 * time.Second)
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-ticker.C:

// 		}
// 	}()
// }

type StoredChat struct {
	id       int64
	messages []*api.Message
	updated  time.Time
	store    map[int64]*StoredChat
}

func (sc *StoredChat) Add(msg api.Message) {
	sc.messages = append(sc.messages, &msg)
	sc.updated = time.Now()
}

// Return slice of messages, it is save for modified
func (sc *StoredChat) Messages() []*api.Message {
	return sc.messages
}

func (sc *StoredChat) Save() {
	sc.store[sc.id] = sc
}

func (sc *StoredChat) Len() int {
	return len(sc.store)
}
