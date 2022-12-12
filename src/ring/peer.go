package ring

import (
	"context"
	"encoding/json"
	"lab1DS/src/peerNet"
	"lab1DS/src/protocol"
	"lab1DS/src/sem"
	"log"
	"net"
)

type PeerList []Peer

/*Sort.Interface methods to aid with sorting.*/
func (a PeerList) Len() int {
	return len(a)
}
func (a PeerList) Less(i, j int) bool {
	return a[i].ID < a[j].ID
}
func (a PeerList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a PeerList) Append(b Peer) {
	a = append(a, b)
}

func (a PeerList) Get(i int) *Peer {
	if i < 0 || i >= a.Len() {
		log.Panic("INDEX OUT OF BOUNDS FOR PEERLIST")
	}
	return &a[i]
}

func (a Peer) UnMarshal(bytes []byte) *Peer {
	json.Unmarshal(bytes, &a)
	return &a
}

func (a Peer) Marshal() []byte {
	ret, _ := json.Marshal(a)
	return ret
}

func (a Peer) Connect() *Peer {
	a.SendSem.Acquire(context.Background(), 1)
	a.Connection = peerNet.ConnectToPeer(a.Ip, a.Port)
	a.SendSem.Release(1)
	return &a
}

func (a Peer) Notify(owner Peer) *PeerList {
	a.Connect()
	a.Send(*protocol.NewMessage().MakeNotify(owner).Marshal())
	res := a.ReadMessage()
	var ret PeerList = nil
	if len(res.Vars) != 0 {
		for i := 0; i < len(res.Vars); i++ {
			ret.Append(*new(Peer).UnMarshal([]byte(res.Vars[i])))
		}
	}
	a.Close()
	return &ret
}

func (a Peer) Send(data []byte) *Peer {
	a.SendSem.Acquire(context.Background(), 1)
	peerNet.SendToPeer(a.Connection, data)
	a.SendSem.Release(1)
	return &a
}

func (a Peer) ReadMessage() protocol.MessageType {
	a.SendSem.Acquire(context.Background(), 1)
	data := peerNet.ListenForData(a.Connection)
	a.SendSem.Release(1)
	message := protocol.MessageType{}
	err := json.Unmarshal(data, &message)

	if err != nil {
		log.Println("ERROR READING FROM PEER: "+string(a.ID), "ERROR : ", err.Error())
	}
	return message
}

func (a Peer) Close() *Peer {
	a.SendSem.Acquire(context.Background(), 1)
	a.Connection.Close()
	a.SendSem.Release(1)
	return &a
}

/*
func (a Peer) Create(id, ip, port string) Peer{
	a.ID = id
	a.Port = port
	a.Ip = ip
	a.SendSem = *sem.NewWeighted(1)
	return a
}
*/

func NewPeer(id, ip, port string) *Peer {
	a := new(Peer)
	a.ID = id
	a.Port = port
	a.Ip = ip
	a.SendSem = *sem.NewWeighted(1)
	return a
}

type Peer struct {
	Ip         string
	Port       string
	ID         string
	Connection net.Conn     `json:"-"`
	SendSem    sem.Weighted `json:"-"`
}
