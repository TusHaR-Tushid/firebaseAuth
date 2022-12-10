package server

import (
	"firebaseAuth/handler"
	"firebaseAuth/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Server struct {
	chi.Router
}

func SetupRoutes() *Server {
	router := chi.NewRouter()
	router.Route("/", func(home chi.Router) {
		home.Post("/register", handler.Register)
		home.Post("/login", handler.Login)
		home.Route("/user", func(user chi.Router) {
			user.Use(middleware.Auth)
			user.Get("/friends", handler.GetFriendList)
			user.Put("/", handler.UpdateUserInfo)
			user.Get("/", handler.GetUsers)
			user.Put("/logout", handler.Logout)
			user.Route("/friend-request", func(request chi.Router) {
				request.Post("/", handler.SendFriendRequest)
				request.Get("/", handler.SeeFriendRequests)
				request.Put("/", handler.UpdateFriendRequestStatus)
			})
		})
	})
	return &Server{router}
}

func (svc *Server) Run(port string) error {
	return http.ListenAndServe(port, svc)
}
