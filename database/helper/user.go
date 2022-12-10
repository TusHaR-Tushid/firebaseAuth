package helper

import (
	"firebaseAuth/database"
	"firebaseAuth/models"
	"firebaseAuth/utilities"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func FetchPasswordAndID(userMail string) (models.UserCredentials, error) {
	// language=SQL
	SQL := `SELECT users.id,password
            FROM   users
            WHERE  email=$1 
            AND    archived_at IS NULL `

	var userCredentials models.UserCredentials

	err := database.FirebaseDB.Get(&userCredentials, SQL, userMail)
	if err != nil {
		logrus.Printf("FetchPasswordAndID: Not able to fetch password or ID : %v", err)
		return userCredentials, err
	}
	return userCredentials, nil
}

func CheckEmail(email string) (int, error) {
	// language=SQL
	SQL := `SELECT id
            FROM   users
            WHERE  email = $1`

	var userID int

	err := database.FirebaseDB.Get(&userID, SQL, email)
	if err != nil {
		logrus.Printf("CheckEmail: cannot get userID from email:%v", err)
		return userID, err
	}
	return userID, nil
}

func CheckSession(userID int) (int, error) {
	SQL := `SELECT id
           FROM    sessions
           WHERE   expires_at IS NULL
           AND     user_id=$1
           ORDER BY id DESC 
           LIMIT 1`
	var sessionID int

	err := database.FirebaseDB.Get(&sessionID, SQL, userID)
	if err != nil {
		logrus.Printf("CheckSession: session expired:%v", err)
		return sessionID, err
	}
	return sessionID, nil
}

func GetEmailPassword(userID int) ([]models.UserEmailPassword, error) {
	SQL := `SELECT email,
                   password
            FROM   users
            WHERE id=$1`

	userEmailPassword := make([]models.UserEmailPassword, 0)

	err := database.FirebaseDB.Select(&userEmailPassword, SQL, userID)
	if err != nil {
		logrus.Printf("GetEmailPassword: cannot get email or password:%v", err)
		return userEmailPassword, err
	}
	return userEmailPassword, nil
}

func FetchUID(email string) (string, error) {
	SQL := `SELECT user_uid
            FROM   users
            WHERE  email = $1`

	var uid string

	err := database.FirebaseDB.Get(&uid, SQL, email)
	if err != nil {
		logrus.Printf("FetchUID: cannot get uid:%V", err)
		return uid, err
	}
	return uid, nil
}

func Register(userDetails models.UserDetails) (int, error) {
	// language=SQL
	SQL := `INSERT INTO users(name, email, password, phone_no, age, gender, user_uid) 
                   VALUES ($1, $2, $3, $4, $5, $6, $7)
                   RETURNING id`
	var userID int

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userDetails.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Printf("Register: Not able to hash password:%v", err)
		return userID, err
	}

	err = database.FirebaseDB.Get(&userID, SQL, userDetails.Name, userDetails.Email, hashPassword, userDetails.Phone, userDetails.Age, userDetails.Gender, userDetails.UID)
	if err != nil {
		logrus.Printf("Register: cannot register user:%v", err)
		return userID, err
	}
	return userID, nil
}

func CreateSession(userID int) error {
	SQL := `INSERT INTO sessions(user_id)
            VALUES   ($1)
            `
	_, err := database.FirebaseDB.Exec(SQL, userID)
	if err != nil {
		logrus.Printf("CreateSession: cannot create user session:%v", err)
		return err
	}
	return nil
}

func SendFriendRequest(friendRequest models.FriendRequest, userID int) error {
	SQL := `INSERT INTO friend_request(request_from, request_to) 
                   VALUES ($1, $2)
                   `
	_, err := database.FirebaseDB.Exec(SQL, userID, friendRequest.RequestTo)
	if err != nil {
		logrus.Printf("SendFriendRequest: cannot send request to user:%v", err)
		return err
	}
	return nil
}

func SeeFriendRequests(filterCheck models.FiltersCheck, userID int) ([]models.RequestList, error) {
	SQL := `SELECT fr.id as id,
                   u.name as user_name,
                   u.id as  user_id
       			   
	        FROM friend_request fr
	             JOIN users u on u.id = fr.request_from
	             
            WHERE fr.request_to = $1
            AND fr.archived_at IS NULL 
            AND u.archived_at IS NULL 
            LIMIT $2 OFFSET $3
            `

	allRequests := make([]models.RequestList, 0)

	err := database.FirebaseDB.Select(&allRequests, SQL, userID, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)
	if err != nil {
		logrus.Printf("SeeFriendRequests: cannot get all requests:%v", err)
		return allRequests, err
	}
	return allRequests, nil
}

func UpdateFriendRequest(allRequest models.AllRequests, userID int) error {
	SQL := `UPDATE friend_request 
            SET status = $1
            WHERE request_to = $2
            AND  request_from = $3
            AND  status != $4  
            AND  archived_at IS NULL 
            `

	var status string
	if allRequest.Status == "accepted" {
		status = utilities.Accepted
	} else if allRequest.Status == "rejected" {
		status = utilities.Rejected
	}
	_, err := database.FirebaseDB.Exec(SQL, status, userID, allRequest.RequestFrom, utilities.Rejected)
	if err != nil {
		logrus.Printf("UpdateFriendRequest: unable to accept request:%v", err)
		return err
	}
	return nil
}

func UpdateUserInfo(userDetails models.UserDetails, userID int) error {
	SQL := `UPDATE users
            SET    name = $1,
                   email = $2,
                   password = $3,
                   phone_no = $4,
                   age = $5,
                   gender = $6
            WHERE id = $7
            AND archived_at IS NULL `

	_, err := database.FirebaseDB.Exec(SQL, userDetails.Name, userDetails.Email, userDetails.Password, userDetails.Phone, userDetails.Age, userDetails.Gender, userID)
	if err != nil {
		logrus.Printf("UpdateUserInfo: cannot update user:%v", err)
		return err
	}
	return nil
}

func GetFriendList(filterCheck models.FiltersCheck, userID int) ([]models.FriendList, error) {
	SQL := `SELECT u.id   as user_id,
       			   u.name as user_name
            FROM   friend_request fr
                   JOIN users u on u.id = fr.request_from
            WHERE request_to = $1
            AND   status = $2
            AND   fr.archived_at IS NULL 
            AND   u.archived_at  IS NULL 
            LIMIT $3 OFFSET $4
            `

	friendList := make([]models.FriendList, 0)

	err := database.FirebaseDB.Select(&friendList, SQL, userID, utilities.Accepted, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)
	if err != nil {
		logrus.Printf("GetFriendList: cannot get friend list:%v", err)
		return friendList, err
	}
	return friendList, nil
}

func GetUsers(filterCheck models.FiltersCheck, userID int) ([]models.UserDetails, error) {
	SQL := `SELECT users.id,
                  name,
                  email,
                  password,
                  phone_no as phone,
                  age,
                  gender
           FROM   users JOIN friend_request fr on users.id = fr.request_from
           WHERE users.archived_at IS NULL
           AND   fr.archived_at IS NULL 
           AND   (users.id <> fr.request_from AND fr.request_to = $2 AND status = $1)
           AND   users.id != $2
           LIMIT $3 OFFSET $4`

	userDetails := make([]models.UserDetails, 0)

	err := database.FirebaseDB.Select(&userDetails, SQL, utilities.Accepted, userID, filterCheck.Limit, filterCheck.Limit*filterCheck.Page)
	if err != nil {
		logrus.Printf("GetUsers: cannot get users:%v", err)
		return userDetails, err
	}
	userList := make([]models.UserDetails, 0)
	for i, _ := range userDetails {
		for j, _ := range userList {
			if userDetails[i].Email == userList[j].Email {
				if userDetails[i].Status == "user" {
					userList[j].Status = "friend"
				}
			} else {

			}
		}
	}
	return userDetails, nil
}

func Logout(userID int) error {
	SQL := `UPDATE sessions
			SET expires_at=now()
			WHERE  id IN(
    			SELECT id
    			FROM sessions
    			WHERE user_id = 5
    			ORDER BY id DESC
    			LIMIT 1
			)`

	_, err := database.FirebaseDB.Exec(SQL, userID)
	if err != nil {
		logrus.Printf("Logout: cannot do logout:%v", err)
		return err
	}
	return nil
}

//func CreateNewUser(userDetails models.UsersLoginDetails) (int, error) {
//
//	var userID int
//	// language=SQL
//	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userDetails.Password), bcrypt.DefaultCost)
//	if err != nil {
//		logrus.Printf("CreateUser: Not able to hash password:%v", err)
//		return userID, err
//	}
//	userName := strings.TrimSpace(strings.Replace(userDetails.Email, "@gmail.com", "", 1))
//
//	SQL := `INSERT INTO users(name, email, password)
//                   VALUES ($1, $2, $3)
//                   RETURNING id`
//
//	userDetails.Email = strings.ToLower(userDetails.Email)
//
//	err = database.FirebaseDB.Get(&userID, SQL, userName, userDetails.Email, hashPassword)
//	if err != nil {
//		logrus.Printf("CreateNewUser: cannot create new user:%v", err)
//		return userID, err
//	}
//	return userID, nil
//}
