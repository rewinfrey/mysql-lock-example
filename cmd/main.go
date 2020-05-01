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

		fmt.Println("Go func 1: Acquiring lock")

		lockName := "lock_for_user_" + strconv.FormatInt(id, 10)

		acquireResult, acquireErr := db.DB().ExecContext(context.Background(), "SELECT GET_LOCK('"+lockName+"', 10)")
		if acquireErr != nil {
			panic(acquireErr)
		}
		fmt.Println(acquireResult)

		fmt.Println("Go func 1: Lock acquired")

		updateResult, updateErr := db.DB().ExecContext(context.Background(), "UPDATE users SET namey = ? WHERE id = ?", "GoFunc1Name", id)
		if updateErr != nil {
			panic(updateErr)
		}
		fmt.Println(updateResult)

		fmt.Println("Go func 1: Update issued")

		fmt.Println("Go func 1: Awake, releasing lock")

		releaseResult, releaseErr := db.DB().ExecContext(context.Background(), "SELECT RELEASE_LOCK('"+lockName+"')")
		if releaseErr != nil {
			panic(releaseErr)
		}
		fmt.Println(releaseResult)

		fmt.Println("Go func 1: Lock released")

		return
	}(id)

	time.Sleep(3 * time.Second)

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

		fmt.Println("Go func 2: Acquiring lock")

		lockName := "lock_for_user_" + strconv.FormatInt(id, 10)

		acquireResult, acquireErr := db.DB().ExecContext(context.Background(), "SELECT GET_LOCK('"+lockName+"', 10)")
		if acquireErr != nil {
			panic(acquireErr)
		}
		fmt.Println(acquireResult)

		fmt.Println("Go func 2: Lock acquired")

		updateResult, updateErr := db.DB().ExecContext(context.Background(), "UPDATE users SET namey = ? WHERE id = ?", "GoFunc2Name", id)
		if updateErr != nil {
			panic(updateErr)
		}
		fmt.Println(updateResult)

		fmt.Println("Go func 2: Update issued")

		releaseResult, releaseErr := db.DB().ExecContext(context.Background(), "SELECT RELEASE_LOCK('"+lockName+"')")
		if releaseErr != nil {
			panic(releaseErr)
		}
		fmt.Println(releaseResult)

		fmt.Println("Go func 2: Lock released")

		return
	}(id)

	wg.Wait()

	return
}
