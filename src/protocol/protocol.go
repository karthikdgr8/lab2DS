package protocol

import (
	"encoding/json"
	"lab1DS/src/ring"
	"log"
)

type MessageType struct {
	Action string
	Owner  ring.Peer
	Vars   []string
}

func NewMessage() *MessageType {
	return new(MessageType)
}

func (a MessageType) MakeSearch(searchTerm string, owner ring.Peer) *MessageType {
	a.Action = "search"
	a.Owner = owner
	a.Vars[0] = searchTerm
	return &a
}

func (a MessageType) MakeNotify(owner ring.Peer) *MessageType {
	a.Owner = owner
	a.Action = "notify"
	return &a
}

func (a MessageType) Marshal() *[]byte {
	ret, err := json.Marshal(&a)
	if err != nil {
		log.Println("ERROR Marshalling message", err.Error())
	}
	return &ret
}

func Unmarshal(data []byte) *MessageType {
	ret := new(MessageType)
	err := json.Unmarshal(data, &ret)

	if err != nil {
		log.Println("ERROR UNMARSHALLING MESSAGE: ", err)
	}
	return ret
}
