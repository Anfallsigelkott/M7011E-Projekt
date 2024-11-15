package main

import (
	"log"
	"m7011e-projekt/src/database"
	"m7011e-projekt/src/routes"
	"os"
)

func main() {
	os.Setenv("secretkey", "shouldberandomized")
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatalln("Got an error", err)
	}
	_ = db

	//external.InitHTTPClient()
	routes.Router(db)
}
