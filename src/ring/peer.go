package ring

import (
	"context"
	"encoding/json"
	"lab1DS/src/peerNet"
	"lab1DS/src/sec"
	"lab1DS/src/sem"
	"log"
	"math/big"
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

func (a *PeerList) Append(b Peer) {
	*a = append(*a, b)

}

func (a PeerList) Get(i int) *Peer {
	if i < 0 || i >= a.Len() {
		log.Panic("INDEX OUT OF BOUNDS FOR PEERLIST")
	}
	return &a[i]
}

func (a *Peer) UnMarshal(bytes []byte) *Peer {
	json.Unmarshal(bytes, &a)
	a.SendSem = *sem.NewWeighted(1)
	return a
}

func (a *Peer) Marshal() []byte {
	ret, _ := json.Marshal(a)
	return ret
}

func (a *Peer) ToJsonString() string {
	return string(a.Marshal())
}

func (a *Peer) Connect() *Peer {
	a.SendSem.Acquire(context.Background(), 1)
	priv := sec.GeneratePrivate()
	pub := priv.PublicKey
	a.Connection = peerNet.ConnectToPeer(a.Ip, a.Port)
	if a.Connection == nil {
		log.Println("ERROR CONNECTING TO CLIENT")
	}
	log.Println("CONNECTION ESTABLISHED: Handshaking..")
	peerNet.SendToPeer(a.Connection, pub.X.Bytes())
	peerNet.SendToPeer(a.Connection, pub.Y.Bytes())
	X := new(big.Int).SetBytes(peerNet.ListenForData(a.Connection))
	Y := new(big.Int).SetBytes(peerNet.ListenForData(a.Connection))
	a.SessionKey = sec.CalculateSessionKey(*priv, X, Y)
	log.Println("Handshake complete")
	a.SendSem.Release(1)
	return a
}

func (a *Peer) Notify(owner Peer) *PeerList {
	a.Connect()
	if a.Connection != nil {
		a.Send(NewMessage().MakeNotify(owner).Marshal())
		res := a.ReadMessage()
		var ret PeerList = nil
		if len(res.Vars) != 0 {
			for i := 0; i < len(res.Vars); i++ {
				ret.Append(*new(Peer).UnMarshal([]byte(res.Vars[i])))
				println("Appended! ", ret.Len())
			}
		}
		a.Close()
		return &ret
	} else {
		return nil
	}
}

func (a *Peer) Send(data []byte) *Peer {
	a.SendSem.Acquire(context.Background(), 1)
	ciphertext := sec.Encrypt(a.SessionKey, data)
	peerNet.SendToPeer(a.Connection, ciphertext)
	a.SendSem.Release(1)
	return a
}

func (a *Peer) ReadMessage() Message {
	a.SendSem.Acquire(context.Background(), 1)
	data := peerNet.ListenForData(a.Connection)
	data = sec.Decrypt(a.SessionKey, data)
	a.SendSem.Release(1)
	message := Message{}
	err := json.Unmarshal(data, &message)

	if err != nil {
		log.Println("ERROR READING FROM PEER: "+string(a.ID), " ERROR : ", err.Error())
	}
	return message
}

func (a *Peer) ReadFile() *[]byte {
	ret := peerNet.ListenForData(a.Connection)
	return &ret
}

func (a *Peer) Close() *Peer {
	a.SendSem.Acquire(context.Background(), 1)
	a.Connection.Close()
	a.SendSem.Release(1)
	return a
}

func FromJsonString(jsonString string) *Peer {
	ret := new(Peer)
	ret.UnMarshal([]byte(jsonString))
	return ret
}

func (a *Peer) Search(term string, owner *Peer) *Peer {
	a.Connect()
	if a.Connection != nil {
		a.Send(new(Message).MakeSearch(term, *owner).Marshal())
		res := a.ReadMessage()
		a.Close()
		dest := FromJsonString(res.Vars[0])

		for dest.ID != res.Owner.ID && dest.ID != owner.ID { //&& dest.ID != owner.ID {
			println("SEARCH DESTINATION: "+dest.ID, " SEARCH RESPONSE OWNER : "+res.Owner.ID+" SEARCH TERM:  "+term)
			dest.Connect()
			dest.Send(new(Message).MakeSearch(term, *owner).Marshal())
			res = dest.ReadMessage()
			dest.Close()
			dest = FromJsonString(res.Vars[0])
		}
		return dest
	}
	return nil
}

func (a *Peer) SetConn(conn net.Conn) {
	a.SendSem = *sem.NewWeighted(1)
	a.Connection = conn
}

func (a *Peer) Int64() *int64 {
	if a.ID != "" {
		bigId, succ := new(big.Int).SetString(a.ID, 16)
		if !succ {
			log.Println("Error parsing iteger from ID: " + a.ID)
			return nil
		} else {
			ret := bigId.Int64()
			return &ret
		}
	}
	log.Println("Warning: attempting to parse int64 from null ID")
	return nil
}

func FromNetwork(conn net.Conn) *Peer {
	peer := new(Peer)
	X := new(big.Int).SetBytes(peerNet.ListenForData(conn))
	Y := new(big.Int).SetBytes(peerNet.ListenForData(conn))
	priv := sec.GeneratePrivate()
	peer.SessionKey = sec.CalculateSessionKey(*priv, X, Y)
	pub := priv.PublicKey
	peer.Send(pub.X.Bytes())
	peer.Send(pub.Y.Bytes())
	peer.Connection = conn
	return peer
}

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
	SessionKey []byte       `json:"-"`
	Connection net.Conn     `json:"-"`
	SendSem    sem.Weighted `json:"-"`
}
