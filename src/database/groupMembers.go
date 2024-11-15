package database

type GroupMembers struct {
	groupID int `gorm:"primaryKey"`
	userID  int `gorm:"primaryKey"`
	roleID  int
}
