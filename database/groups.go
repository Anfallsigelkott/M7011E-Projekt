package database

type Groups struct {
	GroupID   int `gorm:"primaryKey"`
	GroupName string
}
