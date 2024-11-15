package database

type Users struct {
	userID   int    `gorm:"primaryKey"`
	userName string `gorm:"unique;not null"`
	password string
}
