package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
)

type postUsecase struct {
	postRepo  repo.PostRepository
	topicRepo repo.TopicRepository
}

func NewPostUsecase(postRepo repo.PostRepository, topicRepo repo.TopicRepository) *postUsecase {
	return &postUsecase{postRepo: postRepo, topicRepo: topicRepo}
}

func (u *postUsecase) Create(ctx context.Context, post entity.Post) (int64, error) {
	if err := u.checkTopic(ctx, post.TopicID); err != nil {
		return 0, err
	}

	id, err := u.postRepo.Create(ctx, post)
	if err != nil {
		return 0, fmt.Errorf("ForumService - PostUsecase - Create - postRepo.Create(): %w", err)
	}

	return id, nil
}

func (u *postUsecase) GetByID(ctx context.Context, id int64) (*entity.Post, error) {
	post, err := u.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ForumService - PostUsecase - GetByID - postRepo.GetByID(): %w", ErrPostNotFound)
		}
		return nil, fmt.Errorf("ForumService - PostUsecase - GetByID - postRepo.GetByID(): %w", err)
	}
	return post, nil
}

func (u *postUsecase) GetByTopic(ctx context.Context, postID int64) ([]entity.Post, error) {
	posts, err := u.postRepo.GetByTopic(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("ForumService - PostUsecase - GetByTopic - postRepo.GetByTopic(): %w", err)
	}
	return posts, nil
}

func (u *postUsecase) Update(ctx context.Context, postID int64, userID int64, role string, content string) error {
	if err := u.checkAccess(ctx, postID, userID, role); err != nil {
		return err
	}

	if err := u.postRepo.Update(ctx, postID, content); err != nil {
		return fmt.Errorf("ForumService - PostUsecase - Update - postRepo.Update(): %w", err)
	}
	return nil
}

func (u *postUsecase) Delete(ctx context.Context, postID int64, userID int64, role string) error {
	if err := u.checkAccess(ctx, postID, userID, role); err != nil {
		return err
	}

	if err := u.postRepo.Delete(ctx, postID); err != nil {
		return fmt.Errorf("ForumService - PostUsecase - Delete - postRepo.delete(): %w", err)
	}
	return nil
}

func (u *postUsecase) checkTopic(ctx context.Context, topicID int64) error {
	if _, err := u.topicRepo.GetByID(ctx, topicID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("ForumService - PostUsecase - checkTopic - topicRepo.GetByID(): %w", ErrTopicNotFound)
		}
		return fmt.Errorf("ForumService - PostUsecase - checkTopic - topicRepo.GetByID(): %w", err)
	}

	return nil
}

func (u *postUsecase) checkAccess(ctx context.Context, postID int64, userID int64, role string) error {
	post, err := u.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("ForumService - PostUsecase - checkAccess - postRepo.GetByID(): %w", ErrPostNotFound)
		}
		return fmt.Errorf("ForumService - PostUsecase - checkAccess  - postRepo.GetByID(): %w", err)
	}

	if post.AuthorID != userID && role != "admin" {
		return fmt.Errorf("ForumService - PostUsecase - checkAccess  - postRepo.Update(): %w", ErrForbidden)
	}

	return nil
}
