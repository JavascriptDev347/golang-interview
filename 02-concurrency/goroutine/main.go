package main

import (
	"fmt"

	waitgrouprace "github.com/JavascriptDev347/golang-interview/02-concurrency/goroutine/waitgroup_race"
	waitgroupsync "github.com/JavascriptDev347/golang-interview/02-concurrency/goroutine/waitgroup_sync"
)

func main() {
	fmt.Println("02-Concurrency")
	waitgroupsync.WaitGroupSync()
	waitgrouprace.WaitGroupRace()
}
