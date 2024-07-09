package main

import (
	"github.com/multisig-labs/tartarus/database"
	"github.com/multisig-labs/tartarus/node"
)

func main() {
	db, err := database.ConnectSQLite("")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		n, err := node.Generate()
		if err != nil {
			panic(err)
		}
		result := db.Create(&n)
		if result.Error != nil {
			panic(result.Error)
		}
	}
}
