package model

type User struct {
	ID       uint   `gorm:"column:id;primaryKey" json:"id"`
	Username string `gorm:"column:username;unique" json:"username"`
	Password string `gorm:"column:password;not null" json:"-"`
}

type UserUpdate struct {
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
}

type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Trust    bool   `json:"trust"`
}

type UserChangePassword struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type UserChangeUsername struct {
	NewUsername string `json:"new_username" binding:"required"`
}
