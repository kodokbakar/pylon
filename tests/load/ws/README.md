# WebSocket Load Testing

Pylon sends chat messages through the API Gateway WebSocket endpoint:

```text
GET /ws
```

`clank-cli` v0.2.0 does not support WebSocket load testing. Use a WebSocket-capable tool such as k6 or websocat for this scenario.

Suggested scenarios:

* Open authenticated WebSocket connections.
* Join a room.
* Send message envelopes.
* Measure message fan-out latency.
* Track dropped messages and connection errors.

Example message envelope:

```json
{
  "type": "message",
  "room_id": "ROOM_ID",
  "content": "load test message",
  "msg_type": "text"
}
```