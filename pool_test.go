//@author Devansh Gupta

package ppool

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestPool(t *testing.T) {
	x := New()
	x.IdleTime, _ = time.ParseDuration("5s")
	x.New = contructor
	a, _ := x.Get()
	b := a.(*sql.DB)
	t.Log(x.IdleResourceCount(), x.TotalResourceCount())
	x.Put(b, func() { b.Close() })
	t.Log(x.IdleResourceCount(), x.TotalResourceCount())
	e, err := x.Get()
	t.Log(err)
	c := e.(*sql.DB)

	log.Println(x.Put(c))
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
