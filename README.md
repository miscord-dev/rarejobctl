# rarejobctl

レアジョブの講師予約をCLI上から行うためのツールです。

rarejobctl is a tool to automate tutor reservations for the "Rarejob" via CLI. Rarejob is an online English tutorial service for Japanese.

## 使い方

### CLI

`127.0.0.1:4444`で動いているSeleniumサーバを使用し、2022/12/27 9:30開始のレッスンを予約する場合

```
$ rarejobctl \
        -selenium-host 127.0.0.1 \
        -selenium-port 4444 \
        -selenium-browser-name firefox \
        -year 2022 \
        -month 12 \
        -day 27 \
        -time "9:30"
```

### Docker

2022/12/27 9:30開始のレッスンを予約する場合

```
docker run -it ghcr.io/musaprg/rarejobctl-standalone \
        rarejobctl \
                -year 2022 \
                -month 12 \
                -day 27 \
                -time "9:30"
```

2022/12/27 9:30~10:00開始のレッスンを予約する場合

```
docker run -it ghcr.io/musaprg/rarejobctl-standalone \
        rarejobctl \
                -year 2022 \
                -month 12 \
                -day 27 \
                -time "9:30" \
                -margin 30
```
