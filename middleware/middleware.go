package middleware

//import (
//	"firebaseAuth/database/helper"
//	"firebaseAuth/handler"
//	"firebaseAuth/models"
//	"firebaseAuth/utilities"
//	"github.com/dgrijalva/jwt-go"
//	"github.com/sirupsen/logrus"
//	"net/http"
//)
//
//func AuthMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		token := r.Header.Get("token")
//
//		claims := models.Claims{}
//
//		tkn, err1 :=
//
//			jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
//				return handler.JwtKey, nil
//			})
//		if err1 != nil {
//			if err1 == jwt.ErrSignatureInvalid {
//				logrus.Printf("Signature invalid:%v", err1)
//				w.WriteHeader(http.StatusUnauthorized)
//				return
//			}
//
//			logrus.Printf("ParseErr:%v", err1)
//			w.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		if !tkn.Valid {
//			w.WriteHeader(http.StatusUnauthorized)
//			logrus.Printf("token is invalid")
//			return
//		}
//
//		_, err := helper.CheckSession(claims.ID)
//		if err != nil {
//			logrus.Printf("session expired:%v", err)
//			return
//		}
//		userID := claims.ID
//
//		value := models.ContextValues{ID: userID}
//		ctx := context.WithValue(r.Context(), utilities.UserContextKey, value)
//		next.ServeHTTP(w, r.WithContext(ctx))
//	})
//}
