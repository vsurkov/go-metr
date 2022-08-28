
#rest-receiver
docker build -t surkovvs/go-metr-receiver --build-arg PRODUCT_NAME=rest-receiver --build-arg PRODUCT_VER=0.0.48 -f ./build/Dockerfile .
#clickhouse-sender
docker build -t surkovvs/go-metr-sender --build-arg PRODUCT_NAME=clickhouse-sender --build-arg PRODUCT_VER=0.0.48 -f ./build/Dockerfile .