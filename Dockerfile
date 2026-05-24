# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/ovek-workflow-example-app ./cmd/app

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/massivemoose/ovek-workflow-example"

ENV PORT=8080

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/ovek-workflow-example-app /ovek-workflow-example-app

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/ovek-workflow-example-app"]
