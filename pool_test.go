//@author Devansh Gupta

package ppool

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestPool(t *testing.T) {
	x := new(ResourePool)
	x.IdleTime, _ = time.ParseDuration("5s")
	x.New = contructor
	x.Get() //new connection
	a, _ := x.Get()
	b, _ := a.(*sql.DB)
	x.Get() //new connection
	x.Put(b, func() { b.Close() })
	x.Get() //idle connection
	x.Get()
	d, _ := time.ParseDuration("10s")
	<-time.After(d)
}

func contructor() interface{} {
	d, _ := getDatabaseConnection()
	return d
}

//getDatabaseConnection returns a new database client to dgefault application database
func getDatabaseConnection() (*sql.DB, error) {

	s := fmt.Sprint(os.Getenv("MYSQL_DEV_USER"), ":", os.Getenv("MYSQL_DEV_PASSWORD"), "@tcp(", os.Getenv("MYSQL_DEV_HOST"), ")/pubg")
	return sql.Open("mysql", s)
}
