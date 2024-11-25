package database

type Users struct {
	//UserID   int    `gorm:"primaryKey"`
	UserName string `gorm:"primaryKey"`
	Password string
	IsAdmin  bool
}
