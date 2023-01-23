# FLB output plugin for Telegram

FluentBit output plugin for Telegram

## FluentBit OUTPUT Params

| param name  | param value             | description                   | mandatory |
|-------------|-------------------------|-------------------------------|-----------|
| Name        | telegram                | fixed                         | yes       |
| api_token   | YOUT_TELEGRAM_API_TOKEN | telegram api token            | yes       |
| room_id     | ROOM_IDs                | comma seperated room ids      | yes       |
| message_key | message                 | key for message to send (WIP) | no        |

## Build the plugin

```bash
make
```

## Example configuration

```ini
￼[SERVICE]
    plugins_file /path/to/out_telegram.so

​￼[INPUT]
    Name dummy

​￼[OUTPUT]
    Name        telegram
    api_token   YOUR_TELEGRAM_API_TOKEN
    room_ids    ROOM_ID # can set multiple rooms like; ROOM_ID1,ROOM_ID2
```

## Run FluentBit with External plugin

```bash
fluent-bit -e /path/to/out_telegram.so -i cpu -o gstdout
```

## Reference

- <https://docs.fluentbit.io/manual/>
