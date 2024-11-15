package database

type Groups struct {
	groupID   int `gorm:"primaryKey"`
	groupName string
}
