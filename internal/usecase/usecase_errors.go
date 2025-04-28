package usecase

import "errors"

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrTopicNotFound    = errors.New("topic not found")
	ErrPostNotFound     = errors.New("post not found")
	ErrForbidden        = errors.New("forbidden")
)
