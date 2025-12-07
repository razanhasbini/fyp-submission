package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"wazzafak_back/internal/middleware"
	"wazzafak_back/internal/service"

	"github.com/go-chi/chi/v5"
)

type CommentRequest struct {
	Content string `json:"content"`
}

// POST /posts/{postID}/comments
func CommentOnPostHandler(w http.ResponseWriter, r *http.Request) {
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
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID not found"})
		return
	}

	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON body"})
		return
	}

	comment, err := service.AddComment(userID, postID, req.Content)
	if err != nil {
		switch err {
		case service.ErrEmptyComment:
			w.WriteHeader(http.StatusBadRequest)
		case service.ErrPostNotFound:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// DELETE /posts/{postID}/comments/{commentID}
func DeleteCommentFromPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postIDStr := chi.URLParam(r, "postID")
	commentIDStr := chi.URLParam(r, "commentID")

	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	commentID, err2 := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil || err2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid ID"})
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID not found"})
		return
	}

	err = service.DeleteCommentFromPost(postID, commentID, userID)
	if err != nil {
		switch err {
		case service.ErrUnauthorized:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "You can only delete your own comments"})
		case service.ErrCommentNotFound:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Comment not found"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete comment"})
		}
		return
	}

	json.NewEncoder(w).Encode(SuccessResponse{Message: "Comment deleted successfully"})
}

// GET /posts/{postID}/comments
func GetCommentsForPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postIDStr := chi.URLParam(r, "postID")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found", http.StatusUnauthorized)
		return
	}

	comments, err := service.GetCommentsByPost(postID, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}

// PATCH /comments/{commentID}
func UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	commentIDStr := chi.URLParam(r, "commentID")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid comment ID"})
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID not found"})
		return
	}

	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON body"})
		return
	}

	comment, err := service.UpdateComment(commentID, userID, req.Content)
	if err != nil {
		switch err {
		case service.ErrEmptyComment:
			w.WriteHeader(http.StatusBadRequest)
		case service.ErrCommentNotFound:
			w.WriteHeader(http.StatusNotFound)
		case service.ErrUnauthorized:
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(comment)
}
