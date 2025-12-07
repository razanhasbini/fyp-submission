package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"wazzafak_back/internal/middleware"
	"wazzafak_back/internal/service"

	"github.com/go-chi/chi/v5"
)

// ✅ Like a post
func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postIDStr := chi.URLParam(r, "postID")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid post ID"})
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := service.LikePost(userID, postID); err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Post liked successfully"})
}

// ✅ Unlike a post
func UnlikePostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postIDStr := chi.URLParam(r, "postID")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid post ID"})
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := service.UnlikePost(userID, postID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Post unliked successfully"})
}

// ✅ Get users who liked a post
func GetPostLikesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postIDStr := chi.URLParam(r, "postID")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid post ID"})
		return
	}

	likes, err := service.GetUsersWhoLikedPost(postID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch likes"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(likes)
}
