# FLB output plugin for Telegram

FluentBit output plugin for Telegram

## FluentBit OUTPUT Params

| param name  | param value             | description                   | mandatory |
|-------------|-------------------------|-------------------------------|-----------|
| Name        | telegram                | fixed                         | yes       |
| api_token   | YOUT_TELEGRAM_API_TOKEN | telegram api token            | yes       |
| room_ids    | ROOM_IDs                | comma seperated room ids      | yes       |
| message_key | message                 | key for message to send (WIP) | no        |

## Build and Run

Build docker image `flb-tg` which is Telegram enabled fluent-bit image:

```bash
docker build --tag=flb-tg:latest .
```

Run example:

```bash
docker run \
  -it --rm \
  -e TG_API_TOKEN="YOUR_TELEGRAM_API_TOKEN"
  -e TG_ROOM_IDS="ROOM_ID1, ROOM_ID2"
  flb-tg:latest
```

## Reference

- <https://docs.fluentbit.io/manual/>
