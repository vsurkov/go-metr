#Заметки на полях для сборки и запуска

#Сборка бэкенда
docker build -t go-metr-receiver --build-arg PRODUCT_NAME=rest-receiver -f build/Dockerfile .
#Сборка обработчиков
docker build -t go-metr-sender --build-arg PRODUCT_NAME=clickhouse-sender -f build/Dockerfile .