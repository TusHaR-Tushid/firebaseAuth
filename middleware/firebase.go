package middleware

import (
	"context"
	"database/sql"
	firebase "firebase.google.com/go"
	"firebaseAuth/database/helper"
	"firebaseAuth/models"
	"firebaseAuth/utilities"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"net/http"
	"os"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firebaseToken := r.Header.Get("token")
		opt := option.WithCredentialsJSON([]byte(os.Getenv("firebase_key")))
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			logrus.Printf("FirebaseLogin:cannot create firebase application object:%v", err)
			return
		}
		client, err := app.Auth(context.Background())
		if err != nil {
			logrus.Printf("FirebaseLogin:cannot create client:%v", err)
		}

		//header := r.Header.Get(echo.HeaderAuthorization)
		token, err := client.VerifyIDToken(context.Background(), firebaseToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Printf("Auth: cannot virfy token:%v", err)
			return
		}

		userDetails, err := client.GetUser(context.Background(), token.UID)
		if err != nil {
			logrus.Printf("firebaseToken: cannot get user details:%v", err)
			return
		}

		userIDAndPassword, err := helper.FetchPasswordAndID(userDetails.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Printf("FetchPasswordAndID: cannot get user id:%v", err)
			return
		}

		_, err = helper.CheckSession(userIDAndPassword.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				_, EncoderErr := w.Write([]byte("ERROR: session expired"))
				if EncoderErr != nil {
					return
				}
				logrus.Printf("session expired:%v", err)
				//w.WriteHeader(http.StatusUnauthorized)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Printf("CheckSession: unable to check session:%v", err)
				return
			}
		}

		value := models.ContextValues{ID: userIDAndPassword.ID}
		ctx := context.WithValue(r.Context(), utilities.UserContextKey, value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
