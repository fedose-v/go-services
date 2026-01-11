# Inventory

Для сборки:
```bash
  brewkit build
```

Для запуска
```bash
  docker compose up --build
```

Вызов API via grpcurl на примере FindProduct (запуск из корня проекта):
```shell
grpcurl -plaintext -d '{"inventoryID": "df02c657-fa6d-454f-8273-b2b80b8d78d4"}' \
  -vv -import-path api/server/inventorypublicapi \
  -proto api/server/inventorypublicapi/inventorypublicapi.proto \
  localhost:8081 Inventory.InventoryPublicAPI/FindInventory
```