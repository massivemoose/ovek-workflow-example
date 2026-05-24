# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/ovek-workflow-example-digest ./cmd/digest

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/massivemoose/ovek-workflow-example"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/ovek-workflow-example-digest /ovek-workflow-example-digest

USER 65532:65532

ENTRYPOINT ["/ovek-workflow-example-digest"]
