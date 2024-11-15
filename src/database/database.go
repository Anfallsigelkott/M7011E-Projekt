package database

import (
	"fmt"
	"log"
	"os"
	"sync"

	mysqlErr "github.com/go-sql-driver/mysql" // only for error checking
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Forum_db struct {
	db           *gorm.DB
	templateLock *sync.RWMutex
	roleLock     *sync.RWMutex
}

/*
	Creates a Bsight_db using the following environment variables

MYSQL_USER:       user to log in as
MYSQL_PASSWORD:   password of the user
DB_HOST:          ip of db
MYSQL_PORT:       port of the database
MYSQL_DATABASE:   name of the database
*/
func InitDatabase() (Forum_db, error) {
	// formats the dsn in a weird internet way
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
	gormConfig := gorm.Config{PrepareStmt: true, Logger: logger.Discard}
	// opens the connection to the database. second arg is configurations
	db, err := gorm.Open(mysql.Open(dsn), &gormConfig)

	switch err.(type) {
	case *mysqlErr.MySQLError:
		// creating the database will only work the tables that the MYSQL_USER has access to

		db_name := os.Getenv("MYSQL_DATABASE")
		// log.Printf("Couldn't find database %s, creating", db_name)

		query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", db_name)
		dsn2 := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
			os.Getenv("MYSQL_USER"),
			os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("MYSQL_PORT"),
		)

		db, err = gorm.Open(mysql.Open(dsn2), &gormConfig)
		err = db.Exec(query).Error
		if err != nil {
			log.Fatalf("Creating database %s failed due to %v\n", db_name, err)
		}
		db, err = gorm.Open(mysql.Open(dsn), &gormConfig)
		break
	case nil:
		break
	default:
		log.Println("Error establishing connection:", err)
		return Forum_db{}, err
	}

	forum_db := Forum_db{
		db:           db,
		templateLock: &sync.RWMutex{},
		roleLock:     &sync.RWMutex{},
	}
	err = forum_db.autoMigration()
	if err != nil {
		return forum_db, err
	}

	return forum_db, err
}

// creates/updates the tables according the structs
func (self *Forum_db) autoMigration() error {
	// automigrate won't delete old columns
	return self.db.AutoMigrate()
}
