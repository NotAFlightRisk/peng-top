FROM --platform=$BUILDPLATFORM node:25-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS server
ARG TARGETOS TARGETARCH TARGETVARIANT VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o /peng-top .

FROM scratch
COPY --from=server /peng-top /peng-top
ENV PENGTOP_HOST=0.0.0.0 PENGTOP_HOST_ROOT=/host
EXPOSE 7340
HEALTHCHECK --interval=30s --timeout=5s CMD ["/peng-top", "--healthcheck"]
ENTRYPOINT ["/peng-top"]
