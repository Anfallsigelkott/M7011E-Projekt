package database

type Users struct {
	UserID   int    `gorm:"primaryKey"`
	UserName string `gorm:"unique;not null"`
	Password string
}
