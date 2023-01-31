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
	"strings"
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
	//log.Println("CONNECTION ESTABLISHED: Handshaking..")
	peerNet.SendToPeer(a.Connection, pub.X.Bytes())
	peerNet.SendToPeer(a.Connection, pub.Y.Bytes())
	X := new(big.Int).SetBytes(peerNet.ListenForData(a.Connection))
	Y := new(big.Int).SetBytes(peerNet.ListenForData(a.Connection))
	a.SessionKey = sec.CalculateSessionKey(*priv, X, Y)
	//log.Println("Handshake complete")
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
				//println("Appended! ", ret.Len())
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
	//log.Println("Sending: ", string(data))
	peerNet.SendToPeer(a.Connection, ciphertext)
	a.SendSem.Release(1)
	return a
}

func (a *Peer) ReadMessage() *Message {
	if a == nil {
		log.Println("Error, cannot read from nil peer")
		return nil
	}
	a.SendSem.Acquire(context.Background(), 1)
	//log.Println("Reading message from peer:")
	data := peerNet.ListenForData(a.Connection)
	//log.Println("Decrypting..")
	if data == nil {
		a.SendSem.Release(1)
		return nil
	}
	data = sec.Decrypt(a.SessionKey, data)
	//log.Println("READ: ", string(data))
	a.SendSem.Release(1)
	message := Message{}
	err := json.Unmarshal(data, &message)

	if err != nil {
		log.Println("ERROR READING FROM PEER: "+string(a.ID), " ERROR : ", err.Error())
	}
	return &message
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

/*
This function, not to be confused with Ring.ClosestKnown, searches the network for the successor of the term argument
It is called on the entrypoint peer. For efficiency this peer should be deducted by a previous call to Ring.ClosestKnown
The returned Peer, is freshly created from the response given from the closest found Peer.
*/
func (a *Peer) Search(term string, owner *Peer) *Peer {
	a.Connect()
	if a.Connection != nil {
		a.Send(new(Message).MakeSearch(term, *owner).Marshal())
		res := a.ReadMessage()
		a.Close()
		dest := FromJsonString(res.Vars[0])

		for dest.ID != res.Owner.ID && dest.ID != owner.ID {
			//println("SEARCH DESTINATION: "+dest.ID, " SEARCH RESPONSE OWNER : "+res.Owner.ID+" SEARCH TERM:  "+term)
			dest.Connect()
			dest.Send(new(Message).MakeSearch(term, *owner).Marshal())
			res = dest.ReadMessage()
			dest.Close()
			dest = FromJsonString(res.Vars[0])
		}
		if dest.ID == res.Owner.ID { // Self is given as reply, we need to change the address
			print("Node has given itself as return\n")
			dest.Ip = a.Ip
			println("RETURN ADDRESS: ", dest.Ip, ":", dest.Port)
		}

		return dest
	}
	return nil
}

/*
Set the connection of a peer.
*/
func (a *Peer) SetConn(conn net.Conn) {
	a.SendSem = *sem.NewWeighted(1)
	a.Connection = conn
}

/*
Convert ID to int64, used in sorting neighbours.
*/
func (a *Peer) Int64() *int64 {
	if a.ID != "" {
		bigId, succ := new(big.Int).SetString(a.ID, 16)
		if !succ {
			log.Println("Error parsing integer from ID: " + a.ID)
			return nil
		} else {
			ret := bigId.Int64()
			return &ret
		}
	}
	log.Println("Warning: attempting to parse int64 from null ID")
	return nil
}

/*
FromNetwork is a function to build a peer from an encrypted connection over the network.
The first handshaking is performed, and then data relevant to further connections with the peer is
transmitted.
*/
func FromNetwork(conn net.Conn) *Peer {
	peer := new(Peer)
	X := new(big.Int).SetBytes(peerNet.ListenForData(conn))
	Y := new(big.Int).SetBytes(peerNet.ListenForData(conn))
	if X.Int64() == 0 || Y.Int64() == 0 {
		log.Println("Error reading node keys, discarding connection.")
		return nil
	}
	priv := sec.GeneratePrivate()
	peer.SessionKey = sec.CalculateSessionKey(*priv, X, Y)
	//print("SessionKey: " + string(peer.SessionKey))
	pub := priv.PublicKey
	peerNet.SendToPeer(conn, pub.X.Bytes())
	peerNet.SendToPeer(conn, pub.Y.Bytes())
	//println("WRITTEN KEYS")
	peer.Connection = conn
	address := strings.Split(conn.RemoteAddr().String(), ":")
	peer.Ip = address[0]
	println("New peer address: ", peer.Ip, ":", peer.Port)
	peer.SendSem = *sem.NewWeighted(1)
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
	SessionKey []byte       `json:"-"` //Don't marshal
	Connection net.Conn     `json:"-"` //-|-
	SendSem    sem.Weighted `json:"-"` //-|-
}
