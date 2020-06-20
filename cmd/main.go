package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	config "github.com/rewinfrey/go-example/config"
)

const (
	delay           = 2 * time.Second
	timeoutDuration = 1 * time.Second
	noDelay         = 0 * time.Second
	numWorkers      = 5
)

func log(id int, context string, val interface{}) {
	fmt.Printf("%d[%s]: %v\n", id, context, val)
}

func newGoRoutine(id int, db *gorm.DB, wg *sync.WaitGroup) {
	go func() {
		ctx := context.Background()
		defer wg.Done()

		txOptions := sql.TxOptions{Isolation: 2, ReadOnly: false}

		tx, err := db.DB().BeginTx(ctx, &txOptions)
		if err != nil {
			log(id, "BeginTx", err)
			return
		}

		hasLock := make(chan bool)

		go func(hasLock chan bool) {
			ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
			defer cancel()

			_, acquireErr := tx.ExecContext(ctx, "INSERT INTO example (id) VALUES(1)")
			if acquireErr != nil {
				log(id, "acquireErr", acquireErr)
				hasLock <- false
				return
			}

			hasLock <- true
		}(hasLock)

		select {
		case res := <-hasLock:
			if res {
				log(id, "<-hasLock", "has lock")

				time.Sleep(delay)

				_, err = tx.ExecContext(ctx, "DELETE FROM example WHERE id = 1")
				if err != nil {
					log(id, "Delete", err)
					return
				}

				err = tx.Rollback()
				if err != nil {
					log(id, "Rollback", err)
					return
				}
				return
			}

			wg.Add(1)
			newGoRoutine(id*100, db, wg)
		}

		return
	}()
}

func main() {
	db, err := config.OpenDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	var wg sync.WaitGroup

	fmt.Println("kicking off go routines")
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		newGoRoutine(i, db, &wg)
	}

	wg.Wait()

	fmt.Println("all go routines done")

	return
}
