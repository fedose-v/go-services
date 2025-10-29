package query

import (
	"context"

	"server/pkg/app/model"
)

type UrlQueryService interface {
	GetLongUrlByShortUrl(ctx context.Context, shortUrl string) (string, error)
}

func NewUrlQueryService(repo model.UrlReadRepository) UrlQueryService {
	return &urlQueryService{
		repo: repo,
	}
}

type urlQueryService struct {
	repo model.UrlReadRepository
}

func (t *urlQueryService) GetLongUrlByShortUrl(ctx context.Context, shortUrl string) (string, error) {
	url, err := t.repo.FindByShortUrl(ctx, model.ShortUrl(shortUrl))
	if err != nil {
		return "", err
	}

	return string(url.LongUrl()), nil
}
