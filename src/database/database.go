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
	tx := self.db.Begin()
	err := tx.Delete(&Groups{groupID: group}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Delete(&GroupMembers{groupID: group}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Delete(&Posts{postedGroupID: group}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit().Error
	return err
} // Deletes the group itself and the associated entires in groupMembers and posts

// ------------- Entry getters ------------- //

func (self *Forum_db) getGroups() ([]Groups, error) {
	var res []Groups
	err := self.db.Find(&res).Error
	return res, err
}

func (self *Forum_db) getJoinedGroups(user int) ([]Groups, error) {
	var res []Groups
	var tmp []int
	// Find all group-member pairs for this user
	err := self.db.Select("groupID").Where(GroupMembers{userID: user}).Find(&tmp).Error
	if err != nil {
		return nil, err
	}
	// Find and return those groups
	err = self.db.Find(&res, tmp).Error
	return res, err
}

func (self *Forum_db) getUsersInGroup(group int) ([]Users, error) {
	var res []Users
	var tmp []int
	// Find all member-group pairs for this group
	err := self.db.Select("userID").Where(GroupMembers{groupID: group}).Find(&tmp).Error
	if err != nil {
		return nil, err
	}
	// Find and return those users
	err = self.db.Find(&res, tmp).Error
	return res, err
}

func (self *Forum_db) getPostsInGroup(group int) ([]Posts, error) {
	var res []Posts
	err := self.db.Find(&res, Posts{postedGroupID: group}).Error
	return res, err
}

func (self *Forum_db) getRoleInGroup(group int, user int) (int, error) {
	var res int
	err := self.db.Select("roleID").Where(GroupMembers{groupID: group, userID: user}).Find(&res).Error
	return res, err
}
