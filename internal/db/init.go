package db

import "fmt"

func Init() {
	initAuthDb()
	fmt.Println("[DB]Initailized AuthDb")
	initUserDb()
	fmt.Println("[DB]Initailized UserDb")
	initPostDb()
	fmt.Println("[DB]Initailized PostDb")
}
