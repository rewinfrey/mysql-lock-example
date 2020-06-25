package main

import (
	"context"
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
	afterDuration   = 1 * time.Second
	noDelay         = 0 * time.Second
	numWorkers      = 20
)

func log(id int, context string, val interface{}) {
	fmt.Printf("%d[%s]: %v\n", id, context, val)
}

func newGoRoutine(id int, db *gorm.DB, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		hasLock := make(chan bool)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// _, err := db.DB().ExecContext(ctx, "SET TRANSACTION ISOLATION LEVEL READ COMMITTED")
		// if err != nil {
		// 	log(id, "SetTxIsolationLevel", err)
		// 	return
		// }

		// txOptions := sql.TxOptions{Isolation: 0, ReadOnly: false}
		// tx, err := db.DB().BeginTx(ctx, &txOptions)
		tx, err := db.DB().BeginTx(ctx, nil)
		if err != nil {
			log(id, "BeginTx", err)
			return
		}

		go func(hasLock chan bool) {
			defer close(hasLock)

			_, acquireErr := tx.ExecContext(ctx, "INSERT INTO example (id) VALUES(1)")
			if acquireErr != nil {
				log(id, "acquireErr", acquireErr)
				// hasLock <- false
				return
			}

			hasLock <- true
		}(hasLock)

		select {
		case res := <-hasLock:
			if res {
				log(id, "<-hasLock", "has lock")

				time.Sleep(delay)

			} 

			err = tx.Rollback()
			if err != nil {
				log(id, "Rollback", err)
			}

			return

		case <-time.After(afterDuration):
			log(id, "Timeout", "duration met")
			// err = tx.Rollback()
			// if err!= nil {
			// 	log(id, "Rollback", err)
			// }

			cancel()
			time.Sleep(delay)
			wg.Add(1)
			newGoRoutine(id*10, db, wg)
			return

		case done := <-ctx.Done():
			log(id, "Ctx Done", done)
			// err = tx.Rollback()
			// if err!= nil {
			// 	log(id, "Rollback", err)
			// }
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
		newGoRoutine(i+1, db, &wg)
	}

	wg.Wait()

	fmt.Println("all go routines done")

	return
}
