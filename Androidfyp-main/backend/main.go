package main

import (
	"log"
	"net/http"
	"os"

	db "wazzafak_back/internal/database"
	"wazzafak_back/internal/handler"
	"wazzafak_back/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize your ID generator
	db.InitIDGenerator()

	// Load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, assuming environment variables are set externally")
	}

	// Connect to database
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Check that essential env variables are set
	emailFrom := os.Getenv("EMAIL_FROM")
	apiKey := os.Getenv("BREVO_API_KEY")

	if emailFrom == "" || apiKey == "" {
		log.Fatal("EMAIL_FROM and BREVO_API_KEY must be set")
	}

	log.Println("Database connected, ready to go!")

	// Create router
	r := chi.NewRouter()

	// Public routes
	r.Post("/login", handler.LoginHandler)
	r.Get("/health", handler.HealthCheckHandler)

	// Email verification & registration routes (public)
	r.Post("/send-verification-code", handler.SendVerificationCodeHandler)
	r.Post("/verify-email-code", handler.VerifyEmailCodeHandler)
	r.Post("/complete-registration", handler.CompleteRegistrationHandler)

	// Password reset routes (public)
	r.Post("/send-password-reset", handler.SendPasswordResetCodeHandler)
	r.Post("/reset-password", handler.ResetPasswordHandler)

	// Protected routes (require auth)
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware) // Auth middleware applied here
		r.Get("/users/me/posts", handler.GetMyPostsHandler)

		// User routes
		r.Get("/users/me", handler.GetUserProfile)                      // Get authenticated user's profile
		r.Get("/users/me/followers", handler.GetMyFollowers)            // Get authenticated user's followers
		r.Get("/users/me/following", handler.GetMyFollowing)            // Get authenticated user's following
		r.Get("/users/id/{userID}", handler.GetUserByIDHandler)         // Get user by ID
		r.Get("/users/id/{userID}/photo", handler.GetUserPhotoByID)     // Get user photo by ID
		r.Get("/users/id/{userID}/followers", handler.GetFollowersByID) // Get user's followers by ID
		r.Get("/users/id/{userID}/following", handler.GetFollowingByID) // Get user's following by ID
		r.Put("/users/name", handler.UpdateUserName)
		r.Put("/users/photo", handler.UpdatePhoto)
		r.Delete("/users/photo", handler.DeletePhoto)
		r.Get("/posts/{postID}/likes/users", handler.GetPostLikesHandler)

		// User profile by username (public info)
		r.Get("/users/{username}", handler.GetUserByUsername)
		r.Get("/users/{username}/posts", handler.GetUserPosts)
		r.Get("/users/{username}/followers", handler.GetFollowers)
		r.Get("/users/{username}/following", handler.GetFollowing)

		// Post routes
		r.Post("/posts", handler.CreatePost)
		r.Get("/posts/{postID}", handler.GetPost)
		r.Delete("/posts/{postID}", handler.DeletePost)
		r.Get("/posts/all", handler.GetAllPostsHandler)
		r.Get("/posts/{postID}/comments", handler.GetCommentsForPostHandler)
		// Follow/unfollow
		r.Post("/follow/{userID}", handler.FollowHandler)
		r.Delete("/unfollow/{userID}", handler.UnfollowHandler)

		// Post interactions
		r.Post("/posts/{postID}/like", handler.LikePostHandler)
		r.Post("/posts/{postID}/unlike", handler.UnlikePostHandler)
		r.Post("/posts/{postID}/comment", handler.CommentOnPostHandler)
		r.Delete("/posts/{postID}/comments/{commentID}", handler.DeleteCommentFromPostHandler)

		// Feed
		r.Get("/users/feed", handler.GetFeedHandler)
		// Notification routes
		r.Get("/notifications", handler.GetNotificationsHandler)
		r.Get("/notifications/unread-count", handler.GetUnreadCountHandler)
		r.Put("/notifications/{notificationID}/read", handler.MarkNotificationReadHandler)
		r.Put("/notifications/mark-all-read", handler.MarkAllNotificationsReadHandler)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
