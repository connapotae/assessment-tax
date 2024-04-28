FROM golang:1.21.9-alpine3.19 as build
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o ./out/assessment-tax .

FROM alpine:3.19
WORKDIR /src
COPY --from=build /app/out/assessment-tax .
COPY --from=build /app/.env .
EXPOSE 8080
CMD ["/src/assessment-tax"]