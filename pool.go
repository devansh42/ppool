//@author Devansh Gupta
//This file contains code Pool implementation

package ppool

import (
	"log"
	"sync"
	"time"
)

type ResourePool struct {
	//IdleTime, for how long persist idle resource
	IdleTime time.Duration
	//New, is the constructor for new Resource
	New func() interface{}

	//Total resource online
	totalcount int
	//Total resource ready to be reused
	idlecount int
	//list maintains array of idle connection
	list []*idleresource
}

//SetNew, sets constructor for new resources
func (r *ResourePool) SetNew(x func() interface{}) {
	r.New = x
}

//Get, Retrives one resource from pool
//Returns resource and boolean value indicates, whether resource is reused or not
func (r *ResourePool) Get() (interface{}, bool) {
	if r.idlecount == 0 { //We have zero idle resource
		r.totalcount++
		log.Println("New Object created")
		return r.New(), false
	}
	return r.getidle(), true
}

func (r *ResourePool) getidle() interface{} {
	for _, v := range r.list {
		if !v.isdead {
			log.Println("Waiting for channel")
			v.comeback <- true //This terminates life monitering go routine
			log.Println("Idle object reused")
			return v.res
		}
	}
	return 1
}

//Put, puts a reusable resource in pool
func (r *ResourePool) Put(a interface{}, closefunc func()) {
	x := newidleresource(a)
	r.list = append(r.list, x)
	r.idlecount++ //Increase idle count
	log.Println("Idle connection accepted")
	go func(i *idleresource, b interface{}, closef func()) {
		m := new(sync.Mutex) //To avoid race condition

		select {
		case <-i.comeback:
			m.Lock()
			r.idlecount--
			m.Unlock()
			return //Ends goroutine
		case <-time.After(r.IdleTime):
			i.isdead = true //Indicate that resource is dead
			closef()        //This might close the resource
			m.Lock()
			r.totalcount--
			r.idlecount--
			m.Unlock()
			log.Println("Object is destroyed")
			return
		}
	}(x, a, closefunc)
}

//idleresource, handles idle resource
type idleresource struct {
	//comeback channel
	comeback chan bool
	res      interface{}

	isdead bool
}

func newidleresource(res interface{}) *idleresource {
	x := new(idleresource)
	x.comeback = make(chan bool)
	return x
}
