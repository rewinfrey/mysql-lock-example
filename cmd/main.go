package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	config "github.com/rewinfrey/go-example/config"
)

func main() {
	db, err := config.OpenDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	result, err4 := db.DB().ExecContext(context.Background(), "INSERT INTO users (name, namey, age) VALUES (?,?,?)", "KingJames", "JamesKing", 37)
	if err4 != nil {
		panic(err4)
	}

	id, _ := result.LastInsertId()
	fmt.Println("id: " + strconv.FormatInt(id, 10))

	var wg sync.WaitGroup
	wg.Add(1)
	go func(id int64) {
		defer wg.Done()

		fmt.Println("Go func 1")

		db, err := config.OpenDB()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		fmt.Println("Go func 1: Opening transaction")
		tx, err := db.DB().BeginTx(context.Background(), nil)
		if err != nil {
			panic(err)
		}

		fmt.Println("Go func 1: Locking row")

		acquireResult, acquireErr := tx.ExecContext(context.Background(), "SELECT * FROM users WHERE id = ? FOR UPDATE", id)
		if acquireErr != nil {
			panic(acquireErr)
		}
		fmt.Println(acquireResult)

		fmt.Println("Go func 1: Lock acquired")
		fmt.Println("Go func 1: Sleeping with lock")

		time.Sleep(10 * time.Second)

		fmt.Println("Go func 1: Issuing update")

		updateResult, updateErr := tx.ExecContext(context.Background(), "UPDATE users SET namey = ? WHERE id = ?", "GoFunc1Name", id)
		if updateErr != nil {
			panic(updateErr)
		}
		fmt.Println(updateResult)

		fmt.Println("Go func 1: Update issued")

		fmt.Println("Go func 1: Committing transaction")
		if err := tx.Commit(); err != nil {
			panic(err)
		}

		return
	}(id)

	time.Sleep(2 * time.Second)

	wg.Add(1)
	go func(id int64) {
		defer wg.Done()

		fmt.Println("Go func 2")

		db, err := config.OpenDB()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		fmt.Println("Go func 2: Opening transaction")

		tx, err := db.DB().BeginTx(context.Background(), nil)
		if err != nil {
			panic(err)
		}

		fmt.Println("Go func 2: Locking row")

		acquireResult, acquireErr := tx.ExecContext(context.Background(), "SELECT * FROM users WHERE id = ? FOR UPDATE", id)
		if acquireErr != nil {
			panic(acquireErr)
		}
		fmt.Println(acquireResult)

		fmt.Println("Go func 2: Lock acquired")

		fmt.Println("Go func 2: Issuing update")

		updateResult, updateErr := tx.ExecContext(context.Background(), "UPDATE users SET namey = ? WHERE id = ?", "GoFunc2Name", id)
		if updateErr != nil {
			panic(updateErr)
		}
		fmt.Println(updateResult)

		fmt.Println("Go func 2: Update issued")

		fmt.Println("Go func 2: Committing transaction")
		if err := tx.Commit(); err != nil {
			panic(err)
		}

		return
	}(id)

	wg.Wait()

	return
}
