# go-musthave-shortener-tpl
Шаблон репозитория для практического трек "Веб-разработка на Go"

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` - адрес вашего репозитория на Github без префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона выполните следующую команды:

```
git remote add -m main template https://github.com/yandex-praktikum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

затем добавьте полученые изменения в свой репозиторий.

# Запуск программы
```shell
go run cmd/shortener/main.go -a ':8080' -d 'postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable'
go run cmd/shortener/main.go  -s 'ssl'
# start gRPC
./cmd/shortener/grpc/main -d 'postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable'

```
# Награзука
```shell
hey -n 10000 -c 5 -m POST -d 'http://xxx.ru' http://localhost:8080 
hey -n 10000 -c 5 -m GET  http://localhost:8080/GSAPGATLMO
hey -n 10000 -c 5 -m GET  http://localhost:8080//user/urls
hey -n 10000 -c 5 -m POST -d '{"url": "http://test.ru"}' http://localhost:8080/api/shorten
hey -n 10000 -c 5 -m POST -d '{\"url\": \"http://bench" + strconv.Itoa(i) + ".ru\"}' http://localhost:8080/
hey -n 10000 -c 5 -m POST -d '[{"correlation_id": "123","original_url": "yandex.ru"},{"correlation_id": "555","original_url": "nnm.ru"}]' http://localhost:8080/api/shorten/batch

```

# SSL Server
```shell
# Key considerations for algorithm "RSA" ≥ 2048-bit
openssl genrsa -out server.key 2048

# Key considerations for algorithm "ECDSA" ≥ secp384r1
# List ECDSA the supported curves (openssl ecparam -list_curves)
openssl ecparam -genkey -name secp384r1 -out server.key
```