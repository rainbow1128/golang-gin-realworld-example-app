package users

import (
    "golang.org/x/crypto/bcrypt"
    "golang-gin-starter-kit/common"
    "errors"
    "github.com/jinzhu/gorm"
)

type UserModel struct {
    ID           uint        `gorm:"primary_key"`
    Username     string      `gorm:"column:username"`
    Email        string      `gorm:"column:email;unique_index"`
    Bio          string      `gorm:"column:bio;size:1024"`
    Image        *string     `gorm:"column:image"`
    PasswordHash string      `gorm:"column:password;not null"`
}

type FollowModel struct {
    gorm.Model
    Following    UserModel
    FollowingID  uint
    FollowedBy   UserModel
    FollowedByID uint
}

func (u *UserModel) setPassword(password string) error {
    if len(password) == 0 {
        return errors.New("password should not be empty!")
    }
    bytePassword := []byte(password)
    passwordHash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.PasswordHash = string(passwordHash)
    return nil
}

func (u *UserModel) checkPassword(password string) error {
    bytePassword := []byte(password)
    byteHashedPassword := []byte(u.PasswordHash)
    return bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)
}

func FindOneUser(condition interface{}) (UserModel, error) {
    db := common.GetDB()
    var model UserModel
    err := db.Where(condition).First(&model).Error
    return model, err
}

func SaveOne(data interface{}) error {
    db := common.GetDB()
    err := db.Save(data).Error
    return err
}

func (model *UserModel) Update(data interface{}) error {
    db := common.GetDB()
    err := db.Model(model).Update(data).Error
    return err
}

func DeleteFollowModel(condition interface{}) error {
    db := common.GetDB()
    err := db.Where(condition).Delete(FollowModel{}).Error
    return err
}

func (u UserModel) following(v UserModel) error {
    db := common.GetDB()
    var follow FollowModel
    err := db.FirstOrCreate(&follow, &FollowModel{
        FollowingID:  v.ID,
        FollowedByID: u.ID,
    }).Error
    return err
}

func (u UserModel) isFollowing(v UserModel) bool {
    db := common.GetDB()
    var follow FollowModel
    db.Where(FollowModel{
        FollowingID:  v.ID,
        FollowedByID: u.ID,
    }).First(&follow)
    return follow.ID != 0
}

func (u UserModel) unFollowing(v UserModel) error {
    db := common.GetDB()
    err := db.Where(FollowModel{
        FollowingID:  v.ID,
        FollowedByID: u.ID,
    }).Delete(FollowModel{}).Error
    return err
}

func (u UserModel) GetFollowings() []UserModel {
    db := common.GetDB()
    tx := db.Begin()
    var follows []FollowModel
    var followings []UserModel
    tx.Where(FollowModel{
        FollowedByID: u.ID,
    }).Find(&follows)
    for _, follow := range follows{
        var userModel UserModel
        tx.Model(&follow).Related(&userModel,"Following")
        followings = append(followings, userModel)
    }
    tx.Commit()
    return followings
}