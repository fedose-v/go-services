package repo

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"server/pkg/app/model"
	"server/pkg/infrastructure/keyvalue"
)

func NewUrlRepository(rdb *redis.Client) model.UrlRepository {
	return &urlRepository{
		storage: keyvalue.NewStorage[urlSerializable](rdb),
	}
}

type urlSerializable struct {
	LongUrl string `json:"long_url"`
}

type urlRepository struct {
	storage keyvalue.Storage[urlSerializable]
}

func (t *urlRepository) FindByShortUrl(ctx context.Context, shortUrl model.ShortUrl) (model.Url, error) {
	jsonUrl, err := t.storage.Get(ctx, string(shortUrl))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.Url{}, nil
		}
		return model.Url{}, err
	}

	return model.NewUrl(string(shortUrl), jsonUrl.LongUrl), nil
}

func (t *urlRepository) Store(ctx context.Context, url model.Url) error {
	err := t.storage.Set(ctx, string(url.ShortUrl()), urlSerializable{
		LongUrl: string(url.LongUrl()),
	}, 0)

	if errors.Is(err, keyvalue.ErrKeyAlreadyExists) {
		err := t.storage.Delete(ctx, string(url.ShortUrl()))
		if err != nil {
			return err
		}

		return t.storage.Set(ctx, string(url.ShortUrl()), urlSerializable{
			LongUrl: string(url.LongUrl()),
		}, 0)
	}

	return err
}

func (t *urlRepository) Delete(ctx context.Context, shortUrl model.ShortUrl) error {
	return t.storage.Delete(ctx, string(shortUrl))
}
