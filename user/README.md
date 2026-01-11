# User

Для сборки:
```bash
  brewkit build
```

Для запуска
```bash
  docker compose up --build
```

Вызов API via grpcurl на примере FindUser(запуск из корня проекта):
```shell
grpcurl -plaintext -d '{"userID": "df02c657-fa6d-454f-8273-b2b80b8d78d4"}' \
  -vv -import-path api/server/userpublicapi \
  -proto api/server/userpublicapi/userpublicapi.proto \
  localhost:8081 User.UserPublicAPI/FindUser
```