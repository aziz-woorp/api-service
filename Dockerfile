ARG GOLANG_VERSION=1.23

FROM golang:${GOLANG_VERSION} AS builder

ENV GOPRIVATE=gitlab.tamaratech.co/tamara-backend/shared-kernel/go \
  CGO_ENABLED=0

WORKDIR /src

RUN mkdir -p -m 0700 ~/.ssh && \
  ssh-keyscan gitlab.tamaratech.co >> ~/.ssh/known_hosts && \
  git config --global url.ssh://git@gitlab.tamaratech.co/.insteadOf https://gitlab.tamaratech.co/

RUN --mount=type=ssh \
  --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,source=go.sum,target=go.sum \
  --mount=type=bind,source=go.mod,target=go.mod \
  ssh -q -T git@gitlab.tamaratech.co 2>&1 | go mod download

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=bind,target=. \
  go build -ldflags "-s -w" -o /bin/api ./cmd/api

FROM gcr.io/distroless/static-debian11

COPY --from=builder /bin/api /api
COPY /translation /translation

CMD ["/api"]
