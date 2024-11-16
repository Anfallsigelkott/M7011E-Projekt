package database

type Posts struct {
	postID   int `gorm:"primaryKey"`
	replyID  int `gorm:"default:null"`
	posterID int `gorm:"foreignKey:UserID"`
	content  string
}
