FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o dist/main main.go

FROM busybox

WORKDIR /app

COPY --from=builder /app/dist/main ./main
COPY ./resources ./resources

CMD [ "/app/main" ]
