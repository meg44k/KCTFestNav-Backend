# ビルド
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd

# 実行(シェルも無い小さいイメージ)
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
# Cloud Run は PORT を渡してくる。手元で動かすとき用の既定
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
