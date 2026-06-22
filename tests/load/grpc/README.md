# gRPC / Connect Load Testing

Pylon internal services use connect-go over HTTP.

`clank-cli` is used for API Gateway HTTP load testing only. For direct gRPC or Connect service load testing, use a protocol-aware tool such as ghz or a small Go benchmark client.

Suggested targets:

- Chat Service `SendMessage`
- Chat Service `GetMessages`
- Room Service `CreateRoom`
- Room Service `ListRooms`
- Presence Service `SetOnline`
- Notification Service `GetNotifications`