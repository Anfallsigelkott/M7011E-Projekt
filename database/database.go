package database

import (
	"fmt"
	"log"
	"os"
	"sync"

	"strconv"

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
	gormConfig := gorm.Config{PrepareStmt: true, Logger: logger.Default}
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
		fmt.Println("failed after automigrate")
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

func (self *Forum_db) CreateNewUser(uName string, pw string, admin bool) error {
	tmp := Users{
		UserName: uName,
		Password: pw,
		IsAdmin:  admin,
	}
	err := self.db.Create(&tmp).Error
	return err
}

func (self *Forum_db) CreateNewGroup(gName string) error {
	tmp := Groups{
		GroupName: gName,
	}
	err := self.db.Create(&tmp).Error
	return err
}

func (self *Forum_db) AddUserToGroup(user string, group int, role int) error {
	tmp := GroupMembers{
		UserName: user,
		GroupID:  group,
		RoleID:   role,
	}
	err := self.db.Create(&tmp).Error
	return err
}

func (self *Forum_db) CreatePostEntry(poster string, group int, postContent string, reply int) error {
	tmp := Posts{
		PosterName:    poster,
		PostedGroupID: group,
		Content:       postContent,
		ReplyID:       reply,
	}
	err := self.db.Create(&tmp).Error
	return err
}

// ------------- Entry updaters ------------- //

func (self *Forum_db) UpdateUserRole(user string, group int, newRole int) error {
	err := self.db.Where(&GroupMembers{UserName: user, GroupID: group}).Updates(&GroupMembers{RoleID: newRole}).Error
	return err
}

func (self *Forum_db) UpdatePostContent(post int, newContent string) error {
	err := self.db.Where(&Posts{PostID: post}).Updates(&Posts{Content: newContent}).Error
	return err
} // This also serves to remove posts since a post being 'removed' is equivalent to the content being removed (to maintain reply logic)

func (self *Forum_db) UpdateUsername(oldUsername string, newUsername string) error {
	err := self.db.Where(&Users{UserName: oldUsername}).Updates(&Users{UserName: newUsername}).Error
	return err
}

// ------------- Entry removers ------------- //

func (self *Forum_db) RemoveUserFromGroup(user string, group int) error {
	err := self.db.Delete(&GroupMembers{UserName: user, GroupID: group}).Error
	return err
}

func (self *Forum_db) RemoveGroup(group int) error {
	tx := self.db.Begin()
	err := tx.Delete(&Groups{GroupID: group}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Delete(&GroupMembers{GroupID: group}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Delete(&Posts{PostedGroupID: group}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit().Error
	return err
} // Deletes the group itself and the associated entires in groupMembers and posts

// ------------- Entry getters ------------- //

func (self *Forum_db) GetGroups() ([]Groups, error) {
	var res []Groups
	err := self.db.Find(&res).Error
	return res, err
}

//func (self *Forum_db) GetUserByID(user int) (Users, error) {
//	var res Users
//	err := self.db.Find(&res, user).Error
//	return res, err
//}

func (self *Forum_db) GetUserByUsername(user string) (Users, error) {
	var res Users
	err := self.db.Where(Users{UserName: user}).Find(&res).Error
	return res, err
}

func (self *Forum_db) GetJoinedGroups(user string) ([]Groups, error) {
	var res []Groups
	var pairs []GroupMembers
	var groupIDs []int
	// Find all group-member pairs for this user
	err := self.db.Select("group_id").Where(GroupMembers{UserName: user}).Find(&pairs).Error
	if err != nil {
		fmt.Printf("errored while placing data in tmp")
		return nil, err
	}
	for _, pair := range pairs {
		groupIDs = append(groupIDs, pair.GroupID)
	}
	// Find and return those groups
	err = self.db.Find(&res, groupIDs).Error
	return res, err
}

func (self *Forum_db) GetUsersInGroup(group int) ([]string, error) {
	//var res []Users
	var pairs []GroupMembers
	var userNames []string
	// Find all member-group pairs for this group
	err := self.db.Where(GroupMembers{GroupID: group}).Find(&pairs).Error
	if err != nil {
		fmt.Printf("errored while placing data in tmp")
		return nil, err
	}
	for _, pair := range pairs {
		userNames = append(userNames, pair.UserName, strconv.Itoa(pair.RoleID))
	}
	// Find and return those users
	//err = self.db.Find(&res, userNames).Error
	return userNames, err
}

func (self *Forum_db) UserTableIsEmpty() (bool, error) {
	var count int64
	err := self.db.Model(Users{}).Count(&count).Error
	return (count < 1), err
}

func (self *Forum_db) GetPostsInGroup(group int) ([]Posts, error) {
	var res []Posts
	err := self.db.Find(&res, Posts{PostedGroupID: group}).Error
	return res, err
}

func (self *Forum_db) GetRoleInGroup(group int, user string) (int, error) {
	var res GroupMembers
	err := self.db.Select("role_id").Where(GroupMembers{GroupID: group, UserName: user}).Find(&res).Error
	return res.RoleID, err
}

func (self *Forum_db) MatchUserToPost(user string, post int) (Posts, error) {
	var res Posts
	err := self.db.Where(Posts{PosterName: user, PostID: post}).Find(&res).Error
	return res, err
}
