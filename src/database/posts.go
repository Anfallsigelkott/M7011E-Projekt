package database

type Posts struct {
	postID        int    `gorm:"primaryKey"`
	replyID       int    `gorm:"default:null"`
	posterName    string `gorm:"foreignKey:UserID"`
	postedGroupID int    `gorm:"foreignKey:groupID"`
	content       string
}
