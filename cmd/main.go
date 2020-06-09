package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	config "github.com/rewinfrey/go-example/config"
)

const (
	defaultDelay    = 3 * time.Second
	pauseDelay      = 1 * time.Second
	timeoutDuration = 1 * time.Second
	noDelay         = 0 * time.Second
)

func log(name string, msg string) {
	fmt.Println(name + ": " + msg)
}

func newGoRoutine(name string, id int64, delay time.Duration, wg *sync.WaitGroup) {
	go func(name string, id int64, wg *sync.WaitGroup) {
		defer wg.Done()

		db, err := config.OpenDB()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		tx, err := db.DB().BeginTx(context.Background(), nil)
		if err != nil {
			panic(err)
		}

		hasLock := make(chan bool)

		go func(hasLock chan bool, id int64) {
			_, acquireErr := tx.ExecContext(context.Background(), "SELECT * FROM users WHERE id = ? FOR UPDATE", id)
			if acquireErr != nil {
				panic(acquireErr)
			}

			hasLock <- true
		}(hasLock, id)

		log(name, "waiting for lock")
		select {
		case <-hasLock:
			log(name, "has lock")
		case <-time.After(timeoutDuration):
			log(name, "lock timed out")
			log(name, "no update")
			return
		}

		_, updateErr := tx.ExecContext(context.Background(), "UPDATE users SET namey = ? WHERE id = ?", "GoFunc1Name", id)
		if updateErr != nil {
			panic(updateErr)
		}

		time.Sleep(delay)

		if err := tx.Commit(); err != nil {
			panic(err)
		}

		log(name, "update successful")
		return
	}(name, id, wg)
}

func main() {
	db, err := config.OpenDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	tx, beginTxErr := db.DB().BeginTx(context.Background(), nil)
	if beginTxErr != nil {
		panic(beginTxErr)
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			if err == sql.ErrTxDone {
				fmt.Println("already commited")
			} else {
				fmt.Println(err)
				panic(err)
			}
		}
	}()

	result, err4 := db.DB().ExecContext(context.Background(), "INSERT INTO users (name, namey, age) VALUES (?,?,?)", "KingJames", "JamesKing", 37)
	if err4 != nil {
		panic(err4)
	}

	id, _ := result.LastInsertId()
	fmt.Println("id: " + strconv.FormatInt(id, 10))

	var wg sync.WaitGroup
	wg.Add(1)
	newGoRoutine("1", id, defaultDelay, &wg)

	time.Sleep(pauseDelay)

	wg.Add(1)
	newGoRoutine("2", id, defaultDelay, &wg)

	time.Sleep(pauseDelay)

	wg.Add(1)
	newGoRoutine("3", id, noDelay, &wg)

	wg.Wait()

	log("main", "rolling back tx")
	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		panic(rollbackErr)
	}

	return
}
