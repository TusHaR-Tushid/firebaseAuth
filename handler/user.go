package handler

import (
	"context"
	"database/sql"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebaseAuth/database/helper"
	"firebaseAuth/models"
	"firebaseAuth/utilities"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var userDetails models.UsersLoginDetails

	decoderErr := utilities.Decoder(r, &userDetails)
	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Decoder error:%v", decoderErr)
		return
	}

	userDetails.Email = strings.ToLower(userDetails.Email)

	userCredentials, fetchErr := helper.FetchPasswordAndID(userDetails.Email)
	if fetchErr != nil {
		if fetchErr == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("ERROR: Wrong details"))
			if err != nil {
				return
			}

			logrus.Printf("FetchPasswordAndId: not able to get password or id:%v", fetchErr)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if PasswordErr := bcrypt.CompareHashAndPassword([]byte(userCredentials.Password), []byte(userDetails.Password)); PasswordErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logrus.Printf("password misMatch")
		_, err := w.Write([]byte("ERROR: Wrong password"))
		if err != nil {
			return
		}
		return
	}

	opt := option.WithCredentialsJSON([]byte(os.Getenv("firebase_key")))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logrus.Printf("Login:cannot create firebase application object:%v", err)
		return
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		logrus.Printf("Login:cannot create client client:%v", err)
		return
	}

	uid, err := helper.FetchUID(userDetails.Email)
	if err != nil {
		logrus.Printf("FetchUID: cannot get user uid:%v", err)
		return
	}

	claims := map[string]interface{}{
		"id":    userCredentials.ID,
		"email": userDetails.Email,
	}

	customToken, err := client.CustomTokenWithClaims(context.Background(), uid, claims)
	if err != nil {
		log.Fatalf("error setting custom claims %v\n", err)
		return
	}

	//token, err := client.CustomToken(context.Background(), uid)
	//if err != nil {
	//	log.Fatalf("error minting custom token: %v\n", err)
	//	return
	//}

	err = helper.CreateSession(userCredentials.ID)
	if err != nil {
		logrus.Printf("Login: CreateSession: cannot create session:%v", err)
		return
	}

	userOutboundData := make(map[string]interface{})

	userOutboundData["token"] = customToken

	err = utilities.Encoder(w, userOutboundData)
	if err != nil {
		logrus.Printf("Login: Not able to login:%v", err)
		return
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	var userDetails models.UserDetails

	decoderErr := utilities.Decoder(r, &userDetails)
	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("Register: Decoder error:%v", decoderErr)
		return
	}

	opt := option.WithCredentialsJSON([]byte(os.Getenv("firebase_key")))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logrus.Printf("Register:cannot create firebase application object:%v", err)
		return
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		logrus.Printf("Register:cannot create client client:%v", err)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(userDetails.Email).
		EmailVerified(true).
		PhoneNumber(userDetails.Phone).
		Password(userDetails.Password).
		DisplayName(userDetails.Name).
		Disabled(false)
	u, err := client.CreateUser(context.Background(), params)
	if err != nil {
		log.Fatalf("Register:error creating user at firebase: %v\n", err)
		return
	}

	userDetails.UID = u.UID

	userID, err := helper.Register(userDetails)
	if err != nil {
		err := client.DeleteUser(context.Background(), u.UID)
		if err != nil {
			logrus.Errorf("Register: cannot delete firebase user from firebase:%v", err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("Register: cannot register user:%v", err)
		return
	}

	userOutboundData := make(map[string]int)

	userOutboundData["successfully registered with id:"] = userID

	err = utilities.Encoder(w, userOutboundData)
	if err != nil {
		logrus.Printf("Register: encoding error:%v", err)
		return
	}
}

func filters(r *http.Request) (models.FiltersCheck, error) {
	filtersCheck := models.FiltersCheck{}

	var limit int
	var err error
	var page int
	strLimit := r.URL.Query().Get("limit")
	if strLimit == "" {
		limit = 10
	} else {
		limit, err = strconv.Atoi(strLimit)
		if err != nil {
			logrus.Printf("Limit: cannot get limit:%v", err)
			return filtersCheck, err
		}
	}

	strPage := r.URL.Query().Get("page")
	if strPage == "" {
		page = 0
	} else {
		page, err = strconv.Atoi(strPage)
		if err != nil {
			logrus.Printf("Page: cannot get page:%v", err)
			return filtersCheck, err
		}
	}

	filtersCheck = models.FiltersCheck{
		Page:  page,
		Limit: limit}
	return filtersCheck, nil
}

func SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	var friendRequest models.FriendRequest

	decoderErr := utilities.Decoder(r, &friendRequest)

	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("SendFriendRequest: Decoder error:%v", decoderErr)
		return
	}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("SendFriendRequest:QueryParam for ID:%v", ok)
		return
	}
	err := helper.SendFriendRequest(friendRequest, contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("SendFriendRequest: cannot send request to user:%v", err)
		return
	}
}

func SeeFriendRequests(w http.ResponseWriter, r *http.Request) {
	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("SeeFriendRequests:QueryParam for ID:%v", ok)
		return
	}

	filterCheck, err := filters(r)
	if err != nil {
		logrus.Printf("SeeFriendRequests: filterCheck error:%v", err)
		return
	}

	allRequests, err := helper.SeeFriendRequests(filterCheck, contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("SeeFriendRequests: cannot get all requests:%v", err)
		return
	}

	err = utilities.Encoder(w, allRequests)
	if err != nil {
		logrus.Printf("Register: encoding error:%v", err)
		return
	}
}

func UpdateFriendRequestStatus(w http.ResponseWriter, r *http.Request) {
	var allRequests models.AllRequests

	decoderErr := utilities.Decoder(r, &allRequests)
	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("UpdateFriendRequestStatus: Decoder error:%v", decoderErr)
		return
	}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("UpdateFriendRequestStatus:QueryParam for ID:%v", ok)
		return
	}

	err := helper.UpdateFriendRequest(allRequests, contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("UpdateFriendRequestStatus: cannot accept request:%v", err)
		return
	}

	_, err = w.Write([]byte("request updated successfully"))
	if err != nil {
		return
	}
}

func GetFriendList(w http.ResponseWriter, r *http.Request) {
	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("GetFriendList:QueryParam for ID:%v", ok)
		return
	}

	filterCheck, err := filters(r)
	if err != nil {
		logrus.Printf("GetFriendList: filterCheck error:%v", err)
		return
	}

	friendsList, err := helper.GetFriendList(filterCheck, contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("GetFriendList: cannot get list of friends:%v", err)
		return
	}

	err = utilities.Encoder(w, friendsList)
	if err != nil {
		logrus.Printf("GetFriendList: encoding error:%v", err)
		return
	}
}

func UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	var userDetails models.UserDetails

	decoderErr := utilities.Decoder(r, &userDetails)
	if decoderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Printf("UpdateUserInfo: Decoder error:%v", decoderErr)
		return
	}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("UpdateUserInfo:QueryParam for ID:%v", ok)
		return
	}

	err := helper.UpdateUserInfo(userDetails, contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("UpdateUserInfo: cannot update user:%v", err)
		return
	}

	_, err = w.Write([]byte("user Updated successfully"))
	if err != nil {
		return
	}
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	filterCheck, err := filters(r)
	if err != nil {
		logrus.Printf("GetUsers: filterCheck error:%v", err)
		return
	}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("GetUsers:QueryParam for ID:%v", ok)
		return
	}

	userDetails, err := helper.GetUsers(filterCheck, contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("GetUsers: cannot get users:%v", err)
		return
	}

	err = utilities.Encoder(w, userDetails)
	if err != nil {
		logrus.Printf("GetUsers: encoding error:%v", err)
		return
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("Logout:QueryParam for ID:%v", ok)
		return
	}

	err := helper.Logout(contextValues.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("Logout:unable to logout:%v", err)
		return
	}
}

//if fetchErr == sql.ErrNoRows {
//userID, err := helper.CreateNewUser(userDetails)
//if err != nil {
//w.WriteHeader(http.StatusInternalServerError)
//logrus.Printf("FirebaseLogin: cannot create new user:%v", err)
//return
//}
//userEmailPassword, err := helper.GetEmailPassword(userID)
//if err != nil {
//w.WriteHeader(http.StatusInternalServerError)
//logrus.Printf("FirebaseLogin: cannot get new user:%v", err)
//return
//}
//
//userDetails.Email = userEmailPassword[0].Email
//userCredentials.Password = userEmailPassword[0].Password
//} else

//var JwtKey = []byte("secret_key")
//
//func Login(w http.ResponseWriter, r *http.Request) {
//	var userDetails models.UsersLoginDetails
//	decoderErr := utilities.Decoder(r, &userDetails)
//
//	if decoderErr != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		logrus.Printf("Decoder error:%v", decoderErr)
//		return
//	}
//
//	if userDetails.Email == "" {
//		r.Header.Add("oauthToken", userDetails.FirebaseToken)
//		tokenString, err := FirebaseLogin(userDetails.FirebaseToken)
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			logrus.Printf("Login: OauthLogin: cannot get token from oauth login:%v", err)
//			return
//		}
//		userOutboundData := make(map[string]interface{})
//
//		userOutboundData["token"] = tokenString
//
//		err = utilities.Encoder(w, userOutboundData)
//		if err != nil {
//			logrus.Printf("Login: Not able to login:%v", err)
//			return
//		}
//		return
//	}
//
//	userDetails.Email = strings.ToLower(userDetails.Email)
//
//	userCredentials, fetchErr := helper.FetchPasswordAndID(userDetails.Email)
//
//	if fetchErr != nil {
//		if fetchErr == sql.ErrNoRows {
//			w.WriteHeader(http.StatusBadRequest)
//			_, err := w.Write([]byte("ERROR: Wrong details"))
//			if err != nil {
//				return
//			}
//
//			logrus.Printf("FetchPasswordAndId: not able to get password or id:%v", fetchErr)
//			return
//		}
//		w.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	if PasswordErr := bcrypt.CompareHashAndPassword([]byte(userCredentials.Password), []byte(userDetails.Password)); PasswordErr != nil {
//		w.WriteHeader(http.StatusUnauthorized)
//		logrus.Printf("password misMatch")
//		_, err := w.Write([]byte("ERROR: Wrong password"))
//		if err != nil {
//			return
//		}
//		return
//	}
//
//	expiresAt := time.Now().Add(60 * time.Minute)
//
//	claims := &models.Claims{
//		ID: userCredentials.ID,
//		StandardClaims: jwt.StandardClaims{
//			ExpiresAt: expiresAt.Unix(),
//		},
//	}
//
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	tokenString, err := token.SignedString(JwtKey)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		logrus.Printf("TokenString: cannot create token string:%v", err)
//		return
//	}
//
//	err = helper.CreateSession(claims)
//	if err != nil {
//		w.WriteHeader(http.StatusInternalServerError)
//		logrus.Printf("CreateSession: cannot create session:%v", err)
//		return
//	}
//
//	userOutboundData := make(map[string]interface{})
//
//	userOutboundData["token"] = tokenString
//
//	err = utilities.Encoder(w, userOutboundData)
//	if err != nil {
//		logrus.Printf("Login: Not able to login:%v", err)
//		return
//	}
//}

//func FirebaseLogin(authToken string) (string, error) {
//	opt := option.WithCredentialsJSON([]byte(os.Getenv("firebase_key")))
//	app, err := firebase.NewApp(context.Background(), nil, opt)
//	if err != nil {
//		logrus.Printf("FirebaseLogin:cannot create firebase application object:%v", err)
//		return "", err
//	}
//	client, err := app.Auth(context.Background())
//	if err != nil {
//		logrus.Printf("FirebaseLogin:cannot create client client:%v", err)
//	}
//
//	//header := r.Header.Get(echo.HeaderAuthorization)
//	idToken := strings.TrimSpace(strings.Replace(authToken, "Bearer", "", 1))
//	firebaseToken, err := client.VerifyIDToken(context.Background(), idToken)
//	if err != nil {
//		logrus.Printf("firebaseToken:cannot verify token:%v", err)
//		return "", err
//	}
//	userDetails, err := client.GetUser(context.Background(), firebaseToken.UID)
//	if err != nil {
//		logrus.Printf("firebaseToken: cannot get user details:%v", err)
//		return "", err
//	}
//
//	var userID int
//	ID, err := helper.CheckEmail(userDetails.Email)
//	if err != nil {
//		if err == sql.ErrNoRows {
//			ID, createErr := helper.CreateNewUser(userDetails)
//			if createErr != nil {
//				logrus.Printf("firebaseToken: cannot create new user:%v", createErr)
//				return "", createErr
//			}
//			userID = ID
//		} else {
//			logrus.Printf("firebaseToken: cannot check user google email:%v", err)
//			return "", err
//		}
//	} else {
//		userID = ID
//	}
//
//	expiresAt := time.Now().Add(60 * time.Minute)
//
//	claims := &models.Claims{
//		ID: userID,
//		StandardClaims: jwt.StandardClaims{
//			ExpiresAt: expiresAt.Unix(),
//		},
//	}
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	tokenString, err := token.SignedString(JwtKey)
//	if err != nil {
//		logrus.Printf("firebaseToken:TokenString: cannot create token string:%v", err)
//		return "", err
//	}
//
//	err = helper.CreateSession(claims)
//	if err != nil {
//		logrus.Printf("firebaseToken: CreateSession: cannot create session:%v", err)
//		return "", err
//	}
//	return tokenString, nil
//}
