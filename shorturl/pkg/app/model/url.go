package model

import (
	"context"
)

type ShortUrl string
type LongUrl string

type Url struct {
	shortUrl ShortUrl
	longUrl  LongUrl
}

type UrlRepository interface {
	UrlReadRepository
	Store(ctx context.Context, url Url) error
	Delete(ctx context.Context, shortUrl ShortUrl) error
}

type UrlReadRepository interface {
	FindByShortUrl(ctx context.Context, shortUrl ShortUrl) (Url, error)
}

func NewUrl(
	shortUrl string,
	longUrl string,
) Url {
	return Url{
		shortUrl: ShortUrl(shortUrl),
		longUrl:  LongUrl(longUrl),
	}
}

func (t *Url) ShortUrl() ShortUrl {
	return t.shortUrl
}

func (t *Url) LongUrl() LongUrl {
	return t.longUrl
}
