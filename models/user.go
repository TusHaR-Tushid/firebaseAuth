package models

import "github.com/dgrijalva/jwt-go"

type UsersLoginDetails struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
	//FirebaseToken string `json:"firebaseToken"`
}

type UserCredentials struct {
	ID       int    `json:"id"`
	Password string `json:"password"`
}

type UserEmailPassword struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type Claims struct {
	ID int `json:"id"`
	jwt.StandardClaims
}

type UserDetails struct {
	ID       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
	Phone    string `json:"phone" db:"phone"`
	Age      int    `json:"age" db:"age"`
	Gender   string `json:"gender" db:"gender"`
	UID      string `json:"UID" db:"user_uid"`
	Status   string `json:"status" db:"status"`
}

type ContextValues struct {
	ID int `json:"id"`
}

type FriendRequest struct {
	RequestTo int    `json:"requestTo"`
	Status    string `json:"status"`
}

type AllRequests struct {
	ID          int    `json:"id" db:"id"`
	RequestFrom int    `json:"requestFrom" db:"request_from"`
	Status      string `json:"status" db:"status"`
}

type RequestList struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId" db:"user_id"`
	Name   string `json:"name" db:"user_name"`
}

type FriendList struct {
	UserID int    `json:"userId" db:"user_id"`
	Name   string `json:"name" db:"user_name"`
}

type FiltersCheck struct {
	Limit int
	Page  int
}
