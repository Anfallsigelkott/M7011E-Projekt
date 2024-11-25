package database

type Posts struct {
	PostID        int `gorm:"primaryKey"`
	ReplyID       int `gorm:"default:0"`
	PosterName    string
	PostedGroupID int
	Content       string
}
