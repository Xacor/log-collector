# log-collector
Позволяет принимать логи с сервисов по HTTP и централизовано записывать в Yandex Cloud Logging.

## Как запустить
1. Создать сервисный аккаунт с ролью log.writer
2. Загрузить приватный ключ, удостовериться что он в PEM формате.
3. В директорию config добавить файл config.json следующего формата:
    ```json
    {
        "key_id": "<id ключа сервисного аккаунта>",
        "service_account_id": "<id сервисного аккаунта>",
        "key_file": "<путь до файла приватного ключа в формате PEM>",
        "log_group_id": "<id группы логирования>",
        "address": "<ip:port http сервера>"
    }
    ```
4. go run cmd/main.go
5. Сервер ожидает получить POST запрос на `/log/` c JSON в теле запроса
    Возможный формат JSON.
    ```json
    {
        "stream_name":"abc-zxc",
        "agent":"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/113.0",
        "code":401,
        "ip":"127.0.0.1",
        "latency":969,
        "level":"info",
        "method":"POST",
        "msg":"",
        "path":"/auth/login",
        "time":"2023-05-21T10:17:32+03:00"
    }
    ```
    Важно иметь поля stream_name, time, level, msg. Без них будет работать, но смысла будет мало.
    