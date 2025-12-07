package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"wazzafak_back/internal/middleware"
	"wazzafak_back/internal/service"

	"github.com/go-chi/chi/v5"
)

// FollowHandler - logged-in user wants to follow another user
func FollowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// followerID from JWT (you - the person doing the following)
	followerID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized: user ID not found in token"})
		return
	}

	// followingID from path param (the person you want to follow)
	followingIDStr := chi.URLParam(r, "userID")
	followingID, err := strconv.ParseUint(followingIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Prevent self-follow
	if followerID == followingID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "You cannot follow yourself"})
		return
	}

	// Call service: YOU (followerID) follow THEM (followingID)
	err = service.FollowUser(followerID, followingID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Follow failed"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Followed successfully"})
}

// UnfollowHandler - logged-in user wants to unfollow another user
func UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// followerID from JWT (you - the person doing the unfollowing)
	followerID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized: user ID not found in token"})
		return
	}

	// followingID from path param (the person you want to unfollow)
	followingIDStr := chi.URLParam(r, "userID")
	followingID, err := strconv.ParseUint(followingIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Prevent self-unfollow
	if followerID == followingID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "You cannot unfollow yourself"})
		return
	}

	// Call service: YOU (followerID) unfollow THEM (followingID)
	err = service.UnfollowUser(followerID, followingID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to unfollow"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Unfollowed successfully"})
}

// GetFollowers returns all followers of a user by username
func GetFollowers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := chi.URLParam(r, "username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Username is required"})
		return
	}

	followers, err := service.GetFollowersByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve followers"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(followers)
}

// GetFollowing returns all users that a user is following by username
func GetFollowing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := chi.URLParam(r, "username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Username is required"})
		return
	}

	following, err := service.GetFollowingByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve following"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(following)
}

// GetMyFollowers returns followers of the authenticated user (from JWT)
func GetMyFollowers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized: user ID not found in token"})
		return
	}

	followers, err := service.GetFollowersByUserID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve followers"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(followers)
}

// GetMyFollowing returns users that the authenticated user is following (from JWT)
func GetMyFollowing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized: user ID not found in token"})
		return
	}

	following, err := service.GetFollowingByUserID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve following"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(following)
}

// GetFollowersByID returns all followers of a user by user ID
func GetFollowersByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDStr := chi.URLParam(r, "userID")
	if userIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID is required"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid user ID"})
		return
	}

	followers, err := service.GetFollowersByUserID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve followers"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(followers)
}

// GetFollowingByID returns all users that a user is following by user ID
func GetFollowingByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDStr := chi.URLParam(r, "userID")
	if userIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID is required"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid user ID"})
		return
	}

	following, err := service.GetFollowingByUserID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve following"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(following)
}
