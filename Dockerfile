######## Build stage ########
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o idle-svc .

######## Release stage ########
FROM gcr.io/distroless/static:nonroot
COPY --from=build /src/idle-svc /usr/bin/idle-svc
USER nonroot:nonroot
ENTRYPOINT ["/usr/bin/idle-svc"] 