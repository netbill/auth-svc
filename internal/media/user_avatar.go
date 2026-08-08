package media

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/errx"
	"github.com/netbill/auth-svc/internal/models"
	"github.com/netbill/awsx"
)

func (s *Uploader) CreateUserAvatarUploadMediaLinks(
	ctx context.Context,
	userID uuid.UUID,
) (models.UploadMediaLink, error) {
	key := CreateTempUserAvatarKey(userID)

	uploadURL, getURL, err := s.s3.PresignPut(
		ctx,
		key,
		s.config.LinkTTL,
	)
	if err != nil {
		return models.UploadMediaLink{}, fmt.Errorf("presigning put for user avatar: %w", err)
	}

	return models.UploadMediaLink{
		Key:        key,
		PreloadUrl: getURL,
		UploadURL:  uploadURL,
	}, nil
}

func (s *Uploader) UpdateUserAvatar(
	ctx context.Context,
	userID uuid.UUID,
	key string,
) (string, error) {
	err := validateTempUserAvatarKey(userID, key)
	if err != nil {
		return "", err
	}

	out, err := s.s3.GetObjectRange(ctx, key, 64*1024)
	switch {
	case errors.Is(err, awsx.ErrNotFound):
		return "", errx.ErrorUserUploadedAvatarInvalid.Raise(
			fmt.Errorf("user avatar not found for key: %s", key),
		)
	case err != nil:
		return "", fmt.Errorf("get object range for user avatar: %w", err)
	}
	defer out.Body.Close()

	if err = s.config.UserAvatar.Validate(out); err != nil {
		return "", errx.ErrorUserUploadedAvatarInvalid.Raise(
			fmt.Errorf("validating user avatar: %w", err),
		)
	}

	finalKey := CreateUserAvatarKey(userID)

	if err = s.s3.CopyObject(ctx, key, finalKey); err != nil {
		return "", fmt.Errorf("copying object for user avatar: %w", err)
	}

	return finalKey, nil
}

func (s *Uploader) DeleteUploadUserAvatar(
	ctx context.Context,
	userID uuid.UUID,
	key string,
) error {
	if err := validateTempUserAvatarKey(userID, key); err != nil {
		return err
	}

	if err := s.s3.DeleteObject(ctx, key); err != nil {
		return fmt.Errorf("deleting temp user avatar object: %w", err)
	}

	return nil
}

func (s *Uploader) DeleteUserAvatar(
	ctx context.Context,
	userID uuid.UUID,
	key string,
) error {
	if err := validateFinalUserAvatarKey(userID, key); err != nil {
		return err
	}

	if err := s.s3.DeleteObject(ctx, key); err != nil {
		return fmt.Errorf("deleting user avatar object: %w", err)
	}

	return nil
}

var tempUserAvatarKeyRe = regexp.MustCompile(
	`^user/avatar/([0-9a-fA-F-]{36})/temp/([0-9a-fA-F-]{36})$`,
)

func CreateTempUserAvatarKey(userID uuid.UUID) string {
	return fmt.Sprintf("user/avatar/%s/temp/%s", userID, uuid.New().String())
}

func validateTempUserAvatarKey(userID uuid.UUID, key string) error {
	matches := tempUserAvatarKeyRe.FindStringSubmatch(key)
	if matches == nil {
		return errx.ErrorUserUploadedAvatarInvalid.Raise(fmt.Errorf("key %s does not match temp user avatar key pattern", key))
	}

	if matches[1] != userID.String() {
		return errx.ErrorUserUploadedAvatarInvalid.Raise(fmt.Errorf("key %s does not belong to user %s", key, userID))
	}

	return nil
}

var finalUserAvatarKeyRe = regexp.MustCompile(
	`^user/avatar/([0-9a-fA-F-]{36})/([0-9a-fA-F-]{36})$`,
)

func CreateUserAvatarKey(userID uuid.UUID) string {
	return fmt.Sprintf("user/avatar/%s/%s", userID, uuid.New().String())
}

func validateFinalUserAvatarKey(userID uuid.UUID, key string) error {
	matches := finalUserAvatarKeyRe.FindStringSubmatch(key)
	if matches == nil {
		return errx.ErrorUserUploadedAvatarInvalid.Raise(fmt.Errorf("key %s does not match final user avatar key pattern", key))
	}

	if matches[1] != userID.String() {
		return errx.ErrorUserUploadedAvatarInvalid.Raise(fmt.Errorf("key %s does not belong to user %s", key, userID))
	}

	return nil
}
