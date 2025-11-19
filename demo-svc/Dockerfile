FROM golang:1.24-alpine AS build-stage

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /flagger-test /app/main.go

FROM gcr.io/distroless/base-debian11 AS release-stage
WORKDIR /
COPY --from=build-stage /flagger-test /flagger-test

USER nonroot:nonroot
EXPOSE 8080
CMD [ "/flagger-test" ]
