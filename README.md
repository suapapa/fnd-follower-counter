# fnd-follower-counter

SNS follower/subscriber counter displayed on a TM1638 8-digit FND module on Raspberry Pi.
Displays two configured SNS accounts side by side (left 4 digits / right 4 digits), abbreviating counts above `9999` (`k` / `m`) and lighting the top LEDs Green on success or Red on error.

![image](photo/fnd-twitch-counter.jpg)

## Configuration

Create a YAML file (see [config.example.yaml](config.example.yaml)):

```yaml
interval: 60s
left:
  youtube: "@vibetinker"
right:
  insta: "homin.crc"
```

Supported providers for `left` and `right`:
- `youtube`: channel handle (e.g. `@vibetinker`) or `UC...` channel ID (requires `YOUTUBE_API_KEY`)
- `insta` (or `instagram`): Business/Creator username (requires `INSTAGRAM_ACCESS_TOKEN` and `INSTAGRAM_IG_USER_ID` for Meta Graph API Business Discovery)
- `twitch`: channel login name (requires `TWITCH_CLIENT_ID` and `TWITCH_CLIENT_SECRET`)

## Build and Run

Build on host for Raspberry Pi:

    GOOS=linux GOARCH=arm go build -o fnd-follower-counter .

Run on target (defaults to `config.yaml` when `-c` is omitted):

    ./fnd-follower-counter -c config.yaml

Run in console ASCII-art mode without TM1638 hardware:

    ./fnd-follower-counter -n -c config.yaml

## Install systemd Service

Update environment variables in `fnd-follower-counter.service` and install:

    $ sudo cp fnd-follower-counter.service /lib/systemd/system/fnd-follower-counter.service
    $ sudo systemctl enable fnd-follower-counter
