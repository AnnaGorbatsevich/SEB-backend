# Примеры запросов от SEB

1. Положение курсора
```json
{
  "event": "cursor_position",
  "timestamp": "2026-04-23T13:41:40.431074Z",
  "sessionId": "ef198e96-a038-461d-a244-33df6fd21d71",
  "customUserId": "email_test@mail.ru",
  "data": {
    "x": 854,
    "y": 612,
    "ts": "2026-04-23T13:41:40.431074Z"
  }
}
```
2. Обычное нажатие клавиши 

```json
{
  "event": "key_press",
  "timestamp": "2026-04-23T13:41:40.431074Z",
  "sessionId": "ef198e96-a038-461d-a244-33df6fd21d71",
  "customUserId": "email_test@mail.ru",
  "data": {
    "keyCode": 65,
    "keyName": "A",
    "modifiers": [],
    "isCombo": false,
    "ts": "2026-04-23T13:41:40.431074Z"
  }
}
```
3. Комбинация клавиш (Ctrl+C)

```json
{
  "event": "key_press",
  "timestamp": "2026-04-23T13:41:40.431074Z",
  "sessionId": "ef198e96-a038-461d-a244-33df6fd21d71",
  "customUserId": "email_test@mail.ru",
  "data": {
    "keyCode": 67,
    "keyName": "C",
    "modifiers": ["Ctrl"],
    "isCombo": true,
    "ts": "2026-04-23T13:41:40.431074Z"
  }
}
```
4. Логи

```json
{
  "event": "log",
  "timestamp": "2026-04-23T13:41:40.431074Z",
  "sessionId": "ef198e96-a038-461d-a244-33df6fd21d71",
  "customUserId": "email_test@mail.ru",
  "data": {
    "level": "INFO",
    "message": "[ShellResponsibility] The user entered the correct quit password, the application will now terminate.",
    "ts": "2026-04-23T13:41:40.431074Z"
  }
}
```




 ```
 http://localhost:5000/telemetry -Method POST -ContentType "application/json" -Body '{"event":"key_press","timestamp":"2026-04-16T14:00:07.123Z","sessionId": "ef198e96-a038-461d-a244-33df6fd21d71",
 "data":{"keyCode":65,"keyName":"A","modifiers":[],"isCombo":false,"ts":"2026-04-16T14:53:07.123Z"}}'
 ```