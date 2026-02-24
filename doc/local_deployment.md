# Инструкция для локального развертывания компонентов postgis,redis,mino

## <span style="color: green"> env </span>

Необходимо в конфигурации проекта, в разделе envfile указать путь к .example.env

## <span style="color: green"> postgis </span>


Для запуска postgis в контейнере, выполните следующую команду:
linux/windows
```sh
docker run -d --name postgis -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_DB=postgres -p 5432:5432 postgis/postgis
```
MacOS M1/M2/M3
```sh
docker run -d --name postgis --platform linux/amd64 -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_DB=postgres -p 5432:5432 postgis/postgis
```




----------------------------------------------
## <span style="color: green"> redis </span>

```shell
docker run -d -e RDB_ADDR=localhost:6379 -e RDB_PASSWORD= -e RDB_DB=0 -p 6379:6379 redis:latest
```


----------------------------------------------------
## <span style="color: green"> minio  </span> - check docker-compose.yaml (in minio dir)

docker compose up -d

docker compose down


---------------------------------------------------------





