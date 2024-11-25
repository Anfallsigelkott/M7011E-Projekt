package database

type GroupMembers struct {
	GroupID  int    `gorm:"primaryKey"`
	UserName string `gorm:"primaryKey"`
	RoleID   int
}
