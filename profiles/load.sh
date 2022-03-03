hey -n 200 -c 5 -m POST -d 'http://xxx.ru' http://localhost:8080
#hey -n 10000 -c 5 -m GET  http://localhost:8080/GSAPGATLMO
#hey -n 10000 -c 5 -m GET  http://localhost:8080//user/urls
#hey -n 10000 -c 5 -m POST -d '{"url": "http://test.ru"}' http://localhost:8080/api/shorten
#hey -n 10000 -c 5 -m POST -d '{\"url\": \"http://bench" + strconv.Itoa(i) + ".ru\"}' http://localhost:8080/
#hey -n 10000 -c 5 -m POST -d '[{"correlation_id": "123","original_url": "yandex.ru"},{"correlation_id": "555","original_url": "nnm.ru"}]' http://localhost:8080/api/shorten/batch