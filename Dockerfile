FROM golang:1.25 AS build
WORKDIR /src

# templ is a build-time codegen dependency.
RUN go install github.com/a-h/templ/cmd/templ@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN /go/bin/templ generate
RUN CGO_ENABLED=0 go build -o /bin/web ./cmd/web

FROM gcr.io/distroless/static-debian12
COPY --from=build /bin/web /web
EXPOSE 8080
ENTRYPOINT ["/web"]
