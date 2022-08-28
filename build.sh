#Заметки на полях для сборки и запуска

#Сборка бэкенда
docker build -t surkovvs/go-metr-receiver --build-arg PRODUCT_NAME=rest-receiver --build-arg PRODUCT_VER=0.0.4 -f build/Dockerfile .
#docker tag c74cb6f510ca surkovvs/go-metr-receiver:0.0.3
#docker push surkovvs/go-metr-receiver:0.0.3
#Сборка обработчиков
docker build -t surkovvs/go-metr-sender --build-arg PRODUCT_NAME=clickhouse-sender --build-arg PRODUCT_VER=0.0.4  -f build/Dockerfile .
#docker tag c74cb6f510ca surkovvs/go-metr-sender:0.0.3
#docker push surkovvs/go-metr-sender:0.0.3