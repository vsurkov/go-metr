#Заметки на полях для сборки и запуска

#Сборка бэкенда
docker build -t surkovvs/go-metr-receiver --build-arg PRODUCT_NAME=rest-receiver PRODUCT_VER=0.0.3 -f build/Dockerfile .
#docker push surkovvs/go-metr-rest-receiver
#Сборка обработчиков
docker build -t surkovvs/go-metr-sender --build-arg PRODUCT_NAME=clickhouse-sender -f build/Dockerfile .
#docker push surkovvs/go-metr-sender