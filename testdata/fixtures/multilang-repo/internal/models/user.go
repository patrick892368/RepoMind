package models

type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"uniqueIndex"`
    Orders []Order
}

type Order struct {
    ID     uint `gorm:"primaryKey"`
    UserID uint
    User   User
}
