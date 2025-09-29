FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd cmd/wsproxy && go build -o /out/wsproxy

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/wsproxy /app/wsproxy
EXPOSE 8080
ENV ADDR=:8080
USER 65532:65532
ENTRYPOINT ["/app/wsproxy"]


