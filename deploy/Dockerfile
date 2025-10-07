FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd cmd/network-debugger && go build -o /out/network-debugger

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/network-debugger /app/network-debugger
EXPOSE 8080
ENV ADDR=:8080
USER 65532:65532
ENTRYPOINT ["/app/network-debugger"]


