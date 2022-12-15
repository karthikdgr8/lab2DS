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

func (a *Message) MakePut(fileName string, file []byte, owner Peer) *Message {
	a.Action = "put"
	a.Owner = owner
	a.Vars = append(a.Vars, fileName)
	a.Vars = append(a.Vars, string(file))
	return a
}

func (a *Message) MakeGet(fileName string, owner Peer) *Message {
	a.Action = "get"
	a.Owner = owner
	a.Vars = append(a.Vars, fileName)
	return a
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
