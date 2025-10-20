<h1 align="center">DelayedNotifier — отложенные уведомления через очереди</h1>

> - Позволяет положить уведомление в очередь, удалить его и посмотреть его статус
> - Рассылает уведомления в срок по разным каналам (доступны Email в тестовом режиме и Telegram) 
---

## Быстрый старт 

```
git clone https://github.com/child6yo/wbtech-l3-delayed-notifyer

docker compose up 
```

- в dev-сборке уже будет содержаться Redis, RabbitMQ с UI менеджером и MailHog в качестве локального SMTP сервера.
- также, если вы хотите использовать телеграм-бота для уведомлений, потребуется переименовать `.env.example` -> `.env` и указать ключ своего телеграм бота.
- практически вся система конфигурируема через config/config.yml

## API

### POST /notify

#### Request
```
curl -X POST 'localhost:8080/notify' \
--header 'Content-Type: application/json' \
--data-raw '{
    "notification": "Test notification!",
    "delay_seconds": 2,
    "channels": {
        "email_channel": {
            "email": "test@mail.com"
        },
        "tg_channel": {
            "chat_id": "chat_id"
        }
    }
}'
```
- значения в channels - опциональны

#### Response
*201 Created*
```
{
    "uid": "some uuid"
}
```
*400 Bad Request/500 Internal Server Error*
```
    "error": "some error"
```

### GET /notify/{id}

#### Request
```
curl -X GET 'localhost:8080/notify/some-uuid'
```

#### Response
*200 OK*
```
{
    "status": "notification status"
}
```

Возможные статусы:
- "scheduled" - уведомление запланированно.
- "sending" - уведомление отправляется.
- "sent - уведомление отправлено.
- "failed" - ошибка отправки уведомления.

*500 Internal Server Error*
```
    "error": "failed to get notification"
```

### DELETE /notify/{id}

#### Request
```
curl -X DELETE 'localhost:8080/notify/some-uuid'
```

#### Response
*200 OK*
```
{
    "message": "notification deleted"
}
```

*500 Internal Server Error*
```
    "error": "failed to delete notification"
```

## Архитектура

<div align="center">
<img src="etc\assets\schema.png" width="400" style="flex-shrink: 0;">
</div>

    При получении нового отложенного уведомления, 
    оно отправляется на хранение (internal/usecase/notification.go) 
    в Redis (internal/infrastructure/repository), там, помимо него самого, отдельно создается его статус, 
    а также его айди в отсортированном множестве. 

    Отсортированное множество мониторится горутиной Poller (internal/infrastructure/poller). 
    Она каждый тик запускает выдачу n уведомлений, готовых отправке. 
    Статусы уведомлений обновляются и они отправляются в очередь RabbitMQ (internal/infrastructure/messaging).

    Отдельная горутина Consumer (часть messaging) перенаправляет 
    все сообщения из очереди в отдельный канал. 
    Обработкой канала занимается отдельный контроллер (internal/controller/consumer).

    Полученные уведомления отправляются по всем, укаказанным каналам (internal/usecase/sender.go). 
    В случае неудачи отправки, происходит еще несколько попыток с экспоненциальной задержкой.