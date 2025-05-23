FROM golang:1.23 as builder
WORKDIR /app

# Копируем файлы проекта
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка бинарного файла приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o telemed ./cmd

FROM alpine:latest
WORKDIR /root/

# Копируем собранное приложение
COPY --from=builder /app/telemed .

# Копируем .env файл
COPY .env .

# Копируем шаблоны
COPY --from=builder /app/templates /root/templates
# Копируем статику
COPY --from=builder /app/templates/static /root/static

COPY --from=builder /app/templates/static /root/static


EXPOSE 8080

CMD ["./telemed"]
