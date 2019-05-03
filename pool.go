//@author Devansh Gupta
//This file contains code Pool implementation
//It has to methods Get and Put
//Get method reuse an idle resource if have else create new with help of factory function
package ppool

import (
	"errors"
	"log"
	"sync"
	"time"
)

//ResourePool, contains attributes containing info about Resource Pool
type ResourePool struct {
	//IdleTime, for how long persist idle resource
	IdleTime time.Duration
	//New, is the constructor for new Resource
	New func() interface{}

	//Total resource online
	totalcount int
	//Total resource ready to be reused
	idlecount int
	//list maintains list of all connection which are idlized ever
	list []*idleresource
	//mutex, provides synchronization
	mutex *sync.Mutex
	//wait , ensures that get method blocks untill a idle resource is being processed
	//it also help to avoid race conditon
	//i.e. run TestPool in pool_test.go
	wait *sync.WaitGroup
}

//New, is the initializer for the ResourcePool
func New() *ResourePool {
	x := new(ResourePool)
	x.mutex = new(sync.Mutex)
	x.wait = new(sync.WaitGroup)
	return x
}

//EmptyDestructer, is the utlity function which returns an empty function i.e. func(){}
func EmptyDestructer() func() {
	return func() {}
}

//Get, Retrives one resource from pool
//Returns resource and boolean value indicates, whether resource is reused or not
func (r *ResourePool) Get() (interface{}, bool) {
	r.wait.Wait() //It blocks get method if any put method is online
	r.mutex.Lock()
	k := r.idlecount
	r.mutex.Unlock()
	if k == 0 { //We have zero idle resource
		r.totalcount++
		log.Println("New Object created")
		return r.New(), false
	}
	return r.getidle(), true
}

func (r *ResourePool) getidle() interface{} {
	for _, v := range r.list {
		if v.state == state_idle {
			log.Println("Waiting for channel")
			v.comeback <- true //This terminates life monitering go routine
			log.Println("Idle object reused")
			p := <-v.resch //Waiting for channel to return the value
			return p
		}
	}
	return 1
}

//IdleResourceCount, returns no of idle resource in pool
func (r ResourePool) IdleResourceCount() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.idlecount
}

//TotalResourceCount, returns no of total live resource in pool
func (r ResourePool) TotalResourceCount() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.totalcount
}

//ActiveResourceCount, returns no. of active resource
func (r ResourePool) ActiveResourceCount() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.totalcount - r.idlecount
}

//idleresource, handles idle resource
type idleresource struct {
	//comeback channel
	comeback chan bool
	res      interface{}
	resch    chan interface{}
	state    resstate
	closef   func()
}

type resstate int8

const (
	state_active = resstate(1)
	state_idle   = resstate(2)
	state_dead   = resstate(3)
)

func newidleresource(res interface{}, f func()) *idleresource {
	x := new(idleresource)
	x.resch = make(chan interface{})
	x.res = res
	x.comeback = make(chan bool)
	x.closef = f

	return x
}

var (
	DeadResource      = errors.New("Dead Resource ")
	IdleAlready       = errors.New("Resource is already in Idle State")
	NoDestroyFunction = errors.New("Destroy function not defined")
)
