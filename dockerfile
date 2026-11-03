# 빌드: make docker (docker build -t basicproject_go .)
# 설정 / 비밀값은 이미지에 넣지 않는다. 실행할 때 환경변수(--env-file .env, docker compose env_file)로 준다.
FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server .

FROM alpine:3.22

RUN apk --no-cache add ca-certificates tzdata \
    && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=builder /out/server ./server
RUN mkdir -p webdata && chown app:app webdata

USER app

ENV APP_MODE=production \
    PORT=8007

EXPOSE 8007
VOLUME ["/app/webdata"]

CMD ["./server"]
