package ring

import (
	"encoding/json"
	"log"
)

type Message struct {
	Action string
	Owner  Peer
	Vars   []string
}

func NewMessage() *Message {
	return new(Message)
}

func (a *Message) MakeSearch(searchTerm string, owner Peer) *Message {
	a.Action = "search"
	a.Owner = owner
	a.Vars = append(a.Vars, searchTerm)
	return a
}

func (a *Message) MakeNotify(owner Peer) *Message {
	a.Owner = owner
	a.Action = "notify"
	return a
}

func (a *Message) MakeResponse(owner Peer) *Message {
	a.Owner = owner
	a.Action = "response"
	return a
}

func (a *Message) Marshal() []byte {
	ret, err := json.Marshal(&a)
	if err != nil {
		log.Println("ERROR Marshalling message", err.Error())
	}
	return ret
}

func Unmarshal(data []byte) *Message {
	ret := new(Message)
	err := json.Unmarshal(data, &ret)

	if err != nil {
		log.Println("ERROR UNMARSHALLING MESSAGE: ", err)
	}
	return ret
}
