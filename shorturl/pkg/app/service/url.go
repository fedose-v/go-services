package service

import (
	"context"
	"server/pkg/app/model"
)

type UrlService interface {
	Save(ctx context.Context, shortUrl string, longUrl string) error
}

func NewUrlService(repo model.UrlRepository) UrlService {
	return &urlService{
		repo: repo,
	}
}

type urlService struct {
	repo model.UrlRepository
}

func (s *urlService) Save(ctx context.Context, shortUrl string, longUrl string) error {
	newUrl := model.NewUrl(shortUrl, longUrl)
	return s.repo.Store(ctx, newUrl)
}
