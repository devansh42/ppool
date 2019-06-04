//@author Devansh Gupta
package ppool

import (
	"log"
	"time"

	"github.com/devansh42/dsal"
)

//Put, puts a reusable resource in pool
func (r *ResourePool) Put(a interface{}, closefunc ...func()) error {

	t := dsal.NewStackList()

	var err error
	var found bool
	//Check for previous existence
	for r.list.Length() > 0 {
		iv, _ := r.list.Pop()
		v := iv.(*idleresource)
		t.Push(v)

		if v.res == a { //As Pointer are comparable
			//Yeah! We found one
			found = true //Found one resource
			switch v.state {
			case state_dead:
				log.Println(DeadResource)
				err = DeadResource
			case state_idle:
				log.Println(IdleAlready)
				err = IdleAlready
			default:
				r.wait.Add(1)
				go idleRes(r, v) //Make this ideal again
				err = nil
			}
			break
		}
	}
	dsal.DropStack(r.list, t)
	if found {
		return err //Found concering error
	}

	//So we don't know this object
	if len(closefunc) == 0 {
		log.Println(NoDestroyFunction)
		return NoDestroyFunction
	}

	x := newidleresource(a, closefunc[0])
	r.list.Push(x)
	r.wait.Add(1)
	go idleRes(r, x)
	return nil
}

//idleRes, is the go routine which tracks life of an resource
func idleRes(r *ResourePool, i *idleresource) {
	log.Println("Resource idlized")

	i.state = state_idle
	r.mutex.Lock()
	r.idlecount++
	r.mutex.Unlock()
	r.wait.Done() //This ensures that get/put are properly synchronized
	select {
	case <-i.comeback:
		r.mutex.Lock()
		r.idlecount--
		r.mutex.Unlock()
		i.state = state_active
		i.resch <- i.res //This ensures proper synchronzation between
		return           //Ends goroutine
	case <-time.After(r.IdleTime):

		i.state = state_dead //Indicate that resource is dead
		r.mutex.Lock()
		r.idlecount--
		r.totalcount--
		r.mutex.Unlock()
		i.closef() //This might close the resource

		log.Println("Object is destroyed")
		return
	}
}
