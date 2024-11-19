package database

type GroupMembers struct {
	groupID  int    `gorm:"primaryKey"`
	userName string `gorm:"primaryKey"`
	roleID   int
}
