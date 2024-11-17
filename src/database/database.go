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
	return self.db.AutoMigrate(
		&GroupMembers{},
		&Groups{},
		&Posts{},
		&Users{},
	)
}

// ------------------------- DATABASE FUNCTIONS ------------------------- //

// ------------- Entry creators ------------- //

func (self *Forum_db) createNewUser(uName string, pw string) error {
	tmp := Users{
		userName: uName,
		password: pw,
	}
	err := self.db.Create(&tmp).Error
	return err
}

func (self *Forum_db) createNewGroup(gName string) error {
	tmp := Groups{
		groupName: gName,
	}
	err := self.db.Create(&tmp).Error
	return err
}

func (self *Forum_db) addUserToGroup(user int, group int, role int) error {
	tmp := GroupMembers{
		userID:  user,
		groupID: group,
		roleID:  role,
	}
	err := self.db.Create(&tmp).Error
	return err
}

func (self *Forum_db) createPostEntry(poster int, postContent string, reply int) error {
	tmp := Posts{
		posterID: poster,
		content:  postContent,
		replyID:  reply,
	}
	err := self.db.Create(&tmp).Error
	return err
}

// ------------- Entry updaters ------------- //

func (self *Forum_db) updateUserRole(user int, group int, newRole int) error {
	err := self.db.Where(&GroupMembers{userID: user, groupID: group}).Update("roleID", newRole).Error
	return err
}

func (self *Forum_db) updatePostContent(post int, newContent string) error {
	err := self.db.Where(&Posts{postID: post}).Update("content", newContent).Error
	return err
} // This also serves to remove posts since a post being 'removed' is equivalent to the content being removed (to maintain reply logic)

func (self *Forum_db) updateUsername(user int, newUsername string) error {
	err := self.db.Where(&Users{userID: user}).Update("userName", newUsername).Error
	return err
}

// ------------- Entry removers ------------- //

func (self *Forum_db) removeUserFromGroup(user int, group int) error {
	err := self.db.Delete(&GroupMembers{userID: user, groupID: group}).Error
	return err
}

func (self *Forum_db) removeGroup(group int) error {
	err := self.db.Delete(&Groups{groupID: group}).Error
	return err
} // Deletes the group itself

func (grp *Groups) BeforeDelete(forum *Forum_db) error {
	err := forum.db.Delete(&GroupMembers{groupID: grp.groupID}).Error
	if err != nil {
		fmt.Print("groupMembers delete err: %s", err)
		return err
	}
	err = forum.db.Delete(&Posts{postedGroupID: grp.groupID}).Error
	if err != nil {
		fmt.Print("posts delete err: %s", err)
		return err
	}
	return nil
} // Deletes associated groupMembers entries for the now-obsolete groups and posts associated with it

// ------------- Entry getters ------------- //
