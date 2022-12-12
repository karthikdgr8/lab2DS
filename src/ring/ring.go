package ring

import (
	"context"
	"lab1DS/src/peerNet"
	"lab1DS/src/protocol"
	"lab1DS/src/sem"
	"log"
	"math/big"
	"sort"
)

var ID_MAX int64 = 1 << 32

type Ring struct {
	neighbors     PeerList
	successors    PeerList
	fingerTable   PeerList
	owner         Peer
	MAX_NEIGHBORS int
	FINGERS_SIZE  int
	modifySem     sem.Weighted
}

func NewRing(ownId, ownIp, ownPort string, fingerSize int, maxNeighbors int) *Ring {
	ret := new(Ring)
	ret.FINGERS_SIZE = fingerSize
	ret.MAX_NEIGHBORS = maxNeighbors
	ret.owner = *NewPeer(ownId, ownIp, ownPort)
	return ret
}

func (a Ring) GetOwner() Peer {
	return a.owner
}

func (a Ring) AddNeighbour(peer Peer) {
	log.Println("Adding neighbour: " + peer.ID)
	a.modifySem.Acquire(context.Background(), 1)
	neigh := a.neighbors
	ownId := a.owner.ID

	neigh.Append(peer)
	sort.Sort(neigh)
	index := 0

	for i := 0; i < neigh.Len(); i++ {
		if neigh[i].ID < ownId {
			index++
		} else {
			break
		}
	}

	if neigh.Len() > a.MAX_NEIGHBORS*2 {
		removal := (index + a.MAX_NEIGHBORS) % neigh.Len() // Should be the diametrical opposite of index
		a.RemoveNeighbor(removal)
	}
	a.modifySem.Release(1)
}

func (a Ring) RemoveNeighbor(index int) {
	a.neighbors = append(a.neighbors[:index], a.neighbors[index+1:]...)
}

func (a Ring) Search(term string) *Peer {
	if a.neighbors.Len() == 0 { // Know no peers, cant help further than self.
		return &a.owner
	}
	neigh := a.neighbors
	succIndex := 0
	for i := 0; i < neigh.Len(); i++ { // Locate ourselves in the neighbour list.
		if neigh[i].ID > a.owner.ID {
			succIndex++
		} else {
			break
		}
	}
	if succIndex == neigh.Len()-1 {
		succIndex = 0
	}
	if neigh.Len() < a.MAX_NEIGHBORS*2 { // We have complete knowledge of the ring.
		for i := 0; i < neigh.Len(); i++ {
			if neigh[i].ID > term {
				return &neigh[i]
			}
		}
		//If we end up here we have found no larger ids, hence the smallest known is the successor.
		if a.owner.ID < neigh[0].ID {
			return &a.owner
		}
		return &neigh[0]

		/*
			if succIndex- a.MAX_NEIGHBORS == 0 { // centered acyclic neighbour list
				// Look through neighbours from index and down or up.
				if term < a.owner.ID{
					if term < neigh[succIndex- 1].ID && term < neigh[succIndex-2].ID{
						return a.FingerSearch(term)
					} else if term < neigh[succIndex-1].ID{
						return &neigh[succIndex-1]
					} else{
						return &a.owner
					}
				} else{ // look up
					if term > neigh[succIndex].ID {
						return a.FingerSearch(term)
					} else{
						return &neigh[succIndex]
					}
				}

			}
		*/
	} else { //If we don't have complete knowledge. We check if we, or our immediate successor is the successor of the
		// searchTerm. If not. we return the result of a fingerSearch.
		preIndex := succIndex - 1%neigh.Len()
		if term < a.owner.ID { // look down
			if term < neigh[preIndex].ID {
				return a.FingerSearch(term)
			} else {
				return &a.owner
			}
		} else { // look up
			if term > neigh[succIndex].ID {
				return a.FingerSearch(term)
			} else {
				return &neigh[succIndex]
			}
		}

	}
}

func (a Ring) FingerSearch(term string) *Peer {
	a.modifySem.Acquire(context.Background(), 1)
	fingerCopy := a.fingerTable
	a.modifySem.Release(1)
	sort.Sort(fingerCopy)
	if fingerCopy.Len() < 0 {
		return nil
	}
	for i := 0; i < fingerCopy.Len()-1; i++ {
		if term >= fingerCopy[i].ID && term < fingerCopy[i+1].ID {
			return &fingerCopy[i]
		}
	}
	return &fingerCopy[fingerCopy.Len()-1]
}

func (a Ring) FixFingers() {
	a.modifySem.Acquire(context.Background(), 1)
	ownerId, err := new(big.Int).SetString(a.owner.ID, 16)
	if !err {
		log.Println("ERROR Parsing id whilst building fingerTable")
	}
	var slider int64 = 1

	for i := 0; i < a.FINGERS_SIZE; i++ {
		searchId := new(big.Int).SetInt64((ownerId.Int64() + slider<<i) % ID_MAX).Text(16)
		closest := a.Search(searchId)
		searchMessage := protocol.NewMessage().MakeSearch(searchId, a.owner)
		a.fingerTable[i] = peerNet.Search(*searchMessage.Marshal(), *closest)

	}
	a.modifySem.Release(1)
}

/*

func (a Ring) Search(term string) *Peer {
	pre := a.predecessors
	succ := a.successors
	if term <= a.owner.ID {
		if term == a.owner.ID {
			return &a.owner
		} else {
			if pre.Len() <= 0 {
				log.Println("Got search request for smaller id, but have no predecessors.")
				return nil
			} else {
				log.Println("WARNING: Got ID - search smaller than self")
				return &a.predecessors[0]
			}
		}
	} else{

		if succ.Len() == 0{
			return nil
		}
		if term <= succ[succ.Len() -1].ID{
			var ret *Peer
			for i := 0; i < succ.Len(); i++{
				if succ[i].ID > term{
					ret = &succ[i]
				} else{break}
			}
			return ret
		} else{
			finger := a.fingerTable
			if term <= finger[finger.Len() -1].ID{
				var ret *Peer
				for i :=


			} else{return nil}

		}
	}
}

/*
func fingerSearch(id string) *Peer {
	var foundPeer *Peer = nil
	searchTerm, _ := new(big.Int).SetString(id, 16)
	i := 0
	for ; i < fingerTable.Len(); i++ {
		tmp, noProbs := new(big.Int).SetString(fingerTable[i].ID, 16)
		if noProbs && tmp.Cmp(searchTerm) > 0 {
			foundPeer = &fingerTable[i]
		} else if !noProbs {
			log.Println("PROBS!")
		}
	}
	return foundPeer
}


*/
/*Typedef needed to aid in sorting.*/
