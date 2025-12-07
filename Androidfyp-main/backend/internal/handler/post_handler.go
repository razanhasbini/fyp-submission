package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"wazzafak_back/internal/middleware"
	"wazzafak_back/internal/service"

	"github.com/go-chi/chi/v5"
)

// Extended PostResponse for full Android UI support
type PostResponse struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	PhotoURL      string `json:"photo_url"`
	Content       string `json:"content"`
	LikesCount    int    `json:"likes_count"`
	CommentsCount int    `json:"comments_count"`
	IsLiked       bool   `json:"is_liked"`
	IsFollowing   bool   `json:"is_following"`
	CreatedAt     string `json:"created_at"`
	IsOwner       bool   `json:"is_owner"` // ✅ Add this line
}

// ============ Create Post ============
func CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: user ID not found in token", http.StatusUnauthorized)
		return
	}

	var input struct {
		PhotoURL string `json:"photo_url"`
		Content  string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid input"})
		return
	}

	err := service.CreatePost(userID, input.PhotoURL, input.Content)
	if err != nil {
		if err == service.ErrInvalidPostInput {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create post"})
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Post created successfully"})
}

// ============ Delete Post ============
func DeletePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID not found in token"})
		return
	}

	postIDStr := chi.URLParam(r, "postID")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid post ID"})
		return
	}

	err = service.DeletePost(postID, userID)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		case service.ErrUnauthorized:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "You can only delete your own posts"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete post"})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Post deleted successfully"})
}

// ============ Get Feed (Following) ============
func GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: user ID not found in token", http.StatusUnauthorized)
		return
	}

	posts, err := service.GetUserFeed(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve feed"})
		return
	}

	var response []PostResponse
	for _, post := range posts {
		likesCount, _ := service.GetLikesCount(post.ID)
		commentsCount, _ := service.GetCommentsCount(post.ID)
		isLiked, _ := service.HasUserLiked(userID, post.ID)
		isFollowing, _ := service.IsFollowing(userID, post.UserID)

		response = append(response, PostResponse{
			ID:            strconv.FormatUint(post.ID, 10),
			UserID:        strconv.FormatUint(post.UserID, 10),
			PhotoURL:      post.PhotoURL,
			Content:       post.Content,
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
			IsLiked:       isLiked,
			IsFollowing:   isFollowing,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ============ Get User's Own Posts ============
func GetUserPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := chi.URLParam(r, "username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Username is required"})
		return
	}

	posts, err := service.GetPostsByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve posts"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

// ============ Get Single Post ============
func GetPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postID := chi.URLParam(r, "postID")
	if postID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Post ID is required"})
		return
	}

	id, err := strconv.ParseUint(postID, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid post ID"})
		return
	}

	post, err := service.GetPostByID(id)
	if err != nil {
		if err.Error() == "post not found" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Post not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve post"})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

// ============ Get All Posts (For You Feed) ============
func GetAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: user ID not found in token", http.StatusUnauthorized)
		return
	}

	posts, err := service.GetAllPosts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve all posts"})
		return
	}

	var response []PostResponse
	for _, post := range posts {
		likesCount, _ := service.GetLikesCount(post.ID)
		commentsCount, _ := service.GetCommentsCount(post.ID)
		isLiked, _ := service.HasUserLiked(userID, post.ID)
		isFollowing, _ := service.IsFollowing(userID, post.UserID)

		response = append(response, PostResponse{
			ID:            strconv.FormatUint(post.ID, 10),
			UserID:        strconv.FormatUint(post.UserID, 10),
			PhotoURL:      post.PhotoURL,
			Content:       post.Content,
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
			IsLiked:       isLiked,
			IsFollowing:   isFollowing,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			IsOwner:       post.UserID == userID, // ✅ added
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ============ Get My Posts (JWT authenticated) ============
func GetMyPostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID not found in token"})
		return
	}

	posts, err := service.GetMyPosts(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve posts"})
		return
	}

	var response []PostResponse
	for _, post := range posts {
		likesCount, _ := service.GetLikesCount(post.ID)
		commentsCount, _ := service.GetCommentsCount(post.ID)

		response = append(response, PostResponse{
			ID:            strconv.FormatUint(post.ID, 10),
			UserID:        strconv.FormatUint(post.UserID, 10),
			PhotoURL:      post.PhotoURL,
			Content:       post.Content,
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
			IsLiked:       false,
			IsFollowing:   false,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			IsOwner:       true, // always true because JWT user
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
