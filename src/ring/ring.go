package ring

import (
	"context"
	"lab1DS/src/sem"
	"log"
	"math/big"
	"sort"
)

var ID_MAX int64 = 1 << 32

type Ring struct {
	neighbors     PeerList
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
	ret.modifySem = *sem.NewWeighted(1)
	return ret
}

func (a *Ring) GetOwner() Peer {
	return a.owner
}

func (a *Ring) GetNeighbors() PeerList {
	a.modifySem.Acquire(context.Background(), 1)
	ret := a.neighbors // Supposedly hard copy
	a.modifySem.Release(1)
	return ret
}

func (a *Ring) GetPredecessor() *Peer {
	index := 0
	for i := 0; i < a.neighbors.Len(); i++ {
		if *a.neighbors[i].Int64() < *a.owner.Int64() {
			index++
		} else {
			break
		}
	}
	return &a.neighbors[index-1%a.neighbors.Len()]
}

func (a *Ring) GetSuccessor() *Peer {
	index := 0
	for i := 0; i < a.neighbors.Len(); i++ {
		if *a.neighbors[i].Int64() < *a.owner.Int64() {
			index++
		} else {
			break
		}
	}
	if index == a.neighbors.Len()-1 {
		return &a.neighbors[0]
	} else {
		return &a.neighbors[index]
	}
}

// AddNeighbour
// The function takes a peer, and adds it, if and only if there is sufficient room in the neighbours list,
// Or, the peer is closer to any other peer in the list. If this is the case the new node will replace the furthest
// previously known peer.
func (a *Ring) AddNeighbour(peer Peer) {
	for i := 0; i < a.neighbors.Len(); i++ {
		if a.neighbors.Get(i).ID == peer.ID {
			a.RemoveNeighbor(i)
		}
	}
	if peer.ID == a.owner.ID {
		return
	}
	log.Println("Adding neighbour: " + peer.ID)
	ownId := a.owner.Int64()

	a.neighbors = append(a.neighbors, peer)
	sort.Sort(a.neighbors)
	index := 0

	for i := 0; i < a.neighbors.Len(); i++ {
		if *a.neighbors[i].Int64() < *ownId {
			index++
		} else {
			break
		}
	}

	if a.neighbors.Len() > a.MAX_NEIGHBORS*2 {
		removal := (index + a.MAX_NEIGHBORS - 1) % a.neighbors.Len() // Should be the diametrical opposite of index
		log.Println("Removing index: ", removal, "With id: "+a.neighbors.Get(removal).ID)
		a.RemoveNeighbor(removal)
	}
}

func (a *Ring) RemoveNeighbor(index int) {
	println("Removing neighbour with index: ", index)
	a.neighbors = append(a.neighbors[:index], a.neighbors[index+1:]...)
}

// ClosestKnown Not to be confused with the Peer.search() function this function conducts an internal search on the ring
// itself, and returns the closest found peer either by looking ats its immidiete neighbors, */
func (a *Ring) ClosestKnown(term string) *Peer {
	tmp, _ := new(big.Int).SetString(term, 16)
	internalTerm := tmp.Int64()
	if internalTerm < *a.owner.Int64() {
		return &a.owner
	}
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
			if *neigh[i].Int64() > internalTerm {
				return &neigh[i]
			}
		}
		if *a.owner.Int64() < *neigh[0].Int64() {
			return &a.owner
		}
		//If we end up here we have found no larger ids, hence the smallest known is the successor.
		return &neigh[0]
	} else { //If we don't have complete knowledge. We check if we, or our immediate successor is the successor of the
		// searchTerm. If not. we return the result of a fingerSearch.
		preIndex := succIndex - 1%neigh.Len()
		if internalTerm < *a.owner.Int64() { // look down
			if internalTerm < *neigh[preIndex].Int64() {
				return a.FingerSearch(term)
			} else {
				return &a.owner
			}
		} else { // look up
			if internalTerm > *neigh[succIndex].Int64() {
				return a.FingerSearch(term)
			} else {
				return &neigh[succIndex]
			}
		}
	}
}

// FingerSearch  This function conducts a search in the fingertable, and
// returns the closest found node to the term, which is not its successor. */
func (a *Ring) FingerSearch(term string) *Peer {
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

func (a *Ring) Stabilize() {
	a.modifySem.Acquire(context.Background(), 1)

	for i := 0; i < a.neighbors.Len(); i++ {
		if a.neighbors.Get(i).Connect() == nil {
			a.RemoveNeighbor(i)
		} else {
			a.neighbors.Get(i).Close()
		}
	}
	a.modifySem.Release(1)
}

// FixFingers  Function refreshes the fingers for node a. This is done by conducting a search call for the ids
//
//	that would populate a full fingertable./*
func (a *Ring) FixFingers() {
	log.Println("MAINTENENCE: fixing fingers")
	a.modifySem.Acquire(context.Background(), 1)
	ownerId, err := new(big.Int).SetString(a.owner.ID, 16)
	if !err {
		log.Println("ERROR Parsing id whilst building fingerTable")
		return
	}
	var slider int64 = 1

	for i := 0; i < a.FINGERS_SIZE; i++ {
		searchId := new(big.Int).SetInt64((ownerId.Int64() + slider<<i) % ID_MAX).Text(16)
		closest := *a.ClosestKnown(searchId)
		if closest.ID == a.owner.ID {
			break
		}
		//searchMessage := protocol.NewMessage().MakeSearch(searchId, a.owner)
		if i >= a.fingerTable.Len() {
			tmp := closest.Search(searchId, &a.owner)
			if tmp != nil {
				a.fingerTable = append(a.fingerTable, *tmp)
			}

		} else {
			tmp := closest.Search(searchId, &a.owner)
			if tmp != nil {
				a.fingerTable[i] = *tmp
			}
		}

	}
	a.modifySem.Release(1)
}

// MaintainNeighbors This function checks to make sure the closest neighbour is still online. In case its disconnected
// The next predecessor is checked instead./*
func (a *Ring) MaintainNeighbors() {
	if a.neighbors.Len() == 0 {
		return
	}
	ownerId := a.owner.ID
	index := 0

	for i := 0; i < a.neighbors.Len(); i++ {
		if a.neighbors.Get(i).ID < ownerId {
			index++
		} else {
			break
		}
	}

	var resList *PeerList

	for i := 1; i <= a.neighbors.Len(); i++ {
		peer := a.neighbors.Get(index - i%a.neighbors.Len())
		resList = peer.Notify(a.owner)
		if resList == nil {
			a.RemoveNeighbor(index - i%a.neighbors.Len())
			i--
		} else {
			break
		}
		if i == resList.Len() {
			log.Println("WARNING: ran out of neighbors")
			a.AddNeighbour(*a.ClosestKnown(a.owner.ID))
		}
	}

	for i := 0; i < resList.Len(); i++ {
		a.AddNeighbour(*resList.Get(i))
	}
}
