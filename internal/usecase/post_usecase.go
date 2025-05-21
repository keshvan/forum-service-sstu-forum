package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/keshvan/forum-service-sstu-forum/internal/client"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
	"github.com/rs/zerolog"
)

type postUsecase struct {
	postRepo   repo.PostRepository
	topicRepo  repo.TopicRepository
	userClient client.UserClient
	log        *zerolog.Logger
}

const (
	createPostOp = "PostUsecase.Create"
	getByTopicOp = "PostUsecase.GetByTopic"
	deletePostOp = "PostUsecase.Delete"
	updatePostOp = "PostUsecase.Update"
)

func NewPostUsecase(postRepo repo.PostRepository, topicRepo repo.TopicRepository, userClient client.UserClient, log *zerolog.Logger) PostUsecase {
	return &postUsecase{postRepo: postRepo, topicRepo: topicRepo, userClient: userClient, log: log}
}

func (u *postUsecase) Create(ctx context.Context, post entity.Post) (int64, error) {
	if err := u.checkTopic(ctx, post.TopicID); err != nil {
		u.log.Error().Err(err).Str("op", createPostOp).Int64("topic_id", post.TopicID).Msg("Topic not found")
		return 0, err
	}

	id, err := u.postRepo.Create(ctx, post)
	if err != nil {
		u.log.Error().Err(err).Str("op", createPostOp).Any("post", post).Msg("Failed to create post in repository")
		return 0, fmt.Errorf("ForumService - PostUsecase - Create - postRepo.Create(): %w", err)
	}

	u.log.Info().Str("op", createPostOp).Any("post", post).Msg("Post successfully created")
	return id, nil
}

/*func (u *postUsecase) GetByID(ctx context.Context, id int64) (*entity.Post, error) {
	post, err := u.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u.log.Error().Err(err).Str("op", get).Any("post", post).Msg("Failed to create post in repository")
			return nil, fmt.Errorf("ForumService - PostUsecase - GetByID - postRepo.GetByID(): %w", ErrPostNotFound)
		}
		return nil, fmt.Errorf("ForumService - PostUsecase - GetByID - postRepo.GetByID(): %w", err)
	}
	return post, nil
}*/

func (u *postUsecase) GetByTopic(ctx context.Context, topicID int64) ([]entity.Post, error) {
	if err := u.checkTopic(ctx, topicID); err != nil {
		u.log.Error().Err(err).Str("op", getByTopicOp).Int64("topic_id", topicID).Msg("Topic not found")
		return nil, err
	}
	posts, err := u.postRepo.GetByTopic(ctx, topicID)
	if err != nil {
		u.log.Error().Err(err).Str("op", getByTopicOp).Int64("topic_id", topicID).Msg("Failed to get posts")
		return nil, fmt.Errorf("ForumService - PostUsecase - GetByTopic - postRepo.GetByTopic(): %w", err)
	}

	var authorIDs []int64
	authorIDSet := make(map[int64]bool)
	for i := range posts {
		if posts[i].AuthorID != nil {
			if _, exists := authorIDSet[*posts[i].AuthorID]; !exists {
				authorIDs = append(authorIDs, *posts[i].AuthorID)
				authorIDSet[*posts[i].AuthorID] = true
			}
		}
	}

	usernames, err := u.userClient.GetUsernames(ctx, authorIDs)
	if err != nil {
		return nil, fmt.Errorf("ForumService - TopicUsecase  - GetByCategory - userClient.GetUsernames(): %w", err)
	}

	for i := range posts {
		if posts[i].AuthorID == nil {
			posts[i].Username = "Удаленный пользователь"
			continue
		}

		if username, exists := usernames[*posts[i].AuthorID]; exists {
			posts[i].Username = username
		} else {
			posts[i].Username = "Удаленный пользователь"
		}
	}

	u.log.Info().Str("op", getByTopicOp).Int64("topic_id", topicID).Msg("Posts by topic succesfully taken")
	return posts, nil
}

func (u *postUsecase) Update(ctx context.Context, postID int64, userID int64, role string, content string) error {
	if err := u.checkAccess(ctx, postID, userID, role); err != nil {
		u.log.Warn().Err(err).Str("op", updatePostOp).Int64("post_id", postID).Int64("user_id", userID).Msg("Access denied")
		return err
	}

	if err := u.postRepo.Update(ctx, postID, content); err != nil {
		u.log.Error().Err(err).Str("op", updatePostOp).Int64("post_id", postID).Int64("user_id", userID).Msg("Failed to update post in repository")
		return fmt.Errorf("ForumService - PostUsecase - Update - postRepo.Update(): %w", err)
	}

	u.log.Info().Str("op", updatePostOp).Int64("post_id", postID).Msg("Post updated successfully")
	return nil
}

func (u *postUsecase) Delete(ctx context.Context, postID int64, userID int64, role string) error {
	fmt.Printf("USER_ID: %d ,  POST_ID: %d , ROLE: %s", userID, postID, role)
	if err := u.checkAccess(ctx, postID, userID, role); err != nil {
		u.log.Warn().Err(err).Str("op", deletePostOp).Int64("post_id", postID).Int64("user_id", userID).Msg("Access denied")
		return err
	}

	if err := u.postRepo.Delete(ctx, postID); err != nil {
		u.log.Error().Err(err).Str("op", deletePostOp).Int64("post_id", postID).Int64("user_id", userID).Msg("Failed to delete post in repository")
		return fmt.Errorf("ForumService - PostUsecase - Delete - postRepo.delete(): %w", err)
	}

	u.log.Info().Str("op", updatePostOp).Int64("post_id", postID).Msg("Post deleted successfully")
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

	if role == "admin" {
		return nil
	}

	if post.AuthorID == nil || (*post.AuthorID != userID) {
		return fmt.Errorf("ForumService - PostUsecase - checkAccess  - postRepo.Update(): %w", ErrForbidden)
	}

	return nil
}
