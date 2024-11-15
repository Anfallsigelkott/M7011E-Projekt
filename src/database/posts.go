package database

type Posts struct {
	postID  int `gorm:"primaryKey"`
	replyID int `gorm:"default:null"`
	content string
}
