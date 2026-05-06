ARG UPSTREAM_REPO
ARG UPSTREAM_TAG
ARG GO_VER
FROM ${UPSTREAM_REPO:-uselagoon}/commons:${UPSTREAM_TAG:-latest} AS commons

FROM golang:${GO_VER:-1.25}-alpine3.23 AS golang

RUN apk add --no-cache git
# renovate: datasource=github-releases depName=a8m/envsubst
ENV ENVSUBST_VERSION=v1.4.3
RUN go install github.com/a8m/envsubst/cmd/envsubst@${ENVSUBST_VERSION}

WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum

COPY main.go main.go
COPY cmd/ cmd/
COPY internal/ internal/

ARG BUILD
ARG GO_VER
ARG VERSION 
ENV BUILD=${BUILD} \
    GO_VER=${GO_VER} \
    VERSION=${VERSION}

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Do not force rebuild of up-to-date packages (do not use -a) and use the compiler cache folder
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build \
    -ldflags="-s -w \
    -X github.com/uselagoon/build-deploy-tool/cmd.bdtBuild=${BUILD} \
    -X github.com/uselagoon/build-deploy-tool/cmd.goVersion=${GO_VER} \
    -X github.com/uselagoon/build-deploy-tool/cmd.bdtVersion=${VERSION} \
    -extldflags '-static'" \
    -o /app/build-deploy-tool .

# RUN go mod download
# RUN go build -o /app/build-deploy-tool

FROM docker:29.3.0-alpine3.23

LABEL org.opencontainers.image.authors="The Lagoon Authors" maintainer="The Lagoon Authors"
LABEL org.opencontainers.image.source="https://github.com/uselagoon/build-deploy-tool" repository="https://github.com/uselagoon/build-deploy-tool"

ENV LAGOON=build-deploy-image

COPY --from=golang /go/bin/envsubst /bin/envsubst

ARG LAGOON_VERSION
ENV LAGOON_VERSION=$LAGOON_VERSION

# Copy commons files
COPY --from=commons /lagoon /lagoon
COPY --from=commons /bin/fix-permissions /bin/ep /bin/docker-sleep /bin/
COPY --from=commons /sbin/tini /sbin/
COPY --from=commons /home /home

RUN chmod g+w /etc/passwd \
    && mkdir -p /home

ENV TMPDIR=/tmp \
    TMP=/tmp \
    HOME=/home \
    # When Bash is invoked via `sh` it behaves like the old Bourne Shell and sources a file that is given in `ENV`
    ENV=/home/.bashrc \
    # When Bash is invoked as non-interactive (like `bash -c command`) it sources a file that is given in `BASH_ENV`
    BASH_ENV=/home/.bashrc

# Defining Versions
# renovate: datasource=github-tags depName=kubernetes/kubernetes
ENV KUBECTL_VERSION=v1.35.3
# renovate: datasource=github-releases depName=helm/helm
ENV HELM_VERSION=v3.20.2
# renovate: datasource=github-releases depName=mikefarah/yq
ENV YQ_VERSION=v4.53.2
# renovate: datasource=github-releases depName=uselagoon/lagoon-linter
ENV LAGOON_LINTER_VERSION=v0.8.0

RUN apk upgrade --no-cache \
    && apk add --no-cache openssl curl jq parallel bash git py-pip skopeo \
    && git config --global user.email "lagoon@lagoon.io" && git config --global user.name lagoon \
    && pip install --break-system-packages yq

RUN architecture=$(case $(uname -m) in x86_64 | amd64) echo "amd64" ;; aarch64 | arm64 | armv8) echo "arm64" ;; *) echo "amd64" ;; esac) \
    && curl -Lo /usr/bin/kubectl https://dl.k8s.io/release/$KUBECTL_VERSION/bin/linux/${architecture}/kubectl \
    && chmod +x /usr/bin/kubectl \
    && curl -Lo /usr/bin/yq https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_${architecture} \
    && chmod +x /usr/bin/yq \
    && curl -Lo /tmp/helm.tar.gz https://get.helm.sh/helm-${HELM_VERSION}-linux-${architecture}.tar.gz \
    && mkdir /tmp/helm \
    && tar -xzf /tmp/helm.tar.gz -C /tmp/helm --strip-components=1 \
    && mv /tmp/helm/helm /usr/bin/helm \
    && chmod +x /usr/bin/helm \
    && rm -rf /tmp/helm*

RUN mkdir -p /kubectl-build-deploy/git
RUN mkdir -p /kubectl-build-deploy/lagoon
RUN mkdir -p /kubectl-build-deploy/hooks

WORKDIR /kubectl-build-deploy/git

COPY legacy/docker-entrypoint.sh /lagoon/entrypoints/100-docker-entrypoint.sh
COPY legacy/build-deploy.sh /kubectl-build-deploy/build-deploy.sh
COPY legacy/build-deploy-docker-compose.sh /kubectl-build-deploy/build-deploy-docker-compose.sh

COPY legacy/scripts /kubectl-build-deploy/scripts

ENV DBAAS_OPERATOR_HTTP=dbaas.lagoon.svc:5000
ENV DOCKER_HOST=docker-host.lagoon.svc
ENV LAGOON_FEATURE_FLAG_DEFAULT_DOCUMENTATION_URL=https://docs.lagoon.sh

RUN architecture=$(case $(uname -m) in x86_64 | amd64) echo "amd64" ;; aarch64 | arm64 | armv8) echo "arm64" ;; *) echo "amd64" ;; esac) \
    && curl -sSL https://github.com/uselagoon/lagoon-linter/releases/download/${LAGOON_LINTER_VERSION}/lagoon-linter_${LAGOON_LINTER_VERSION#v}_linux_${architecture}.tar.gz \
    | tar -xz -C /usr/local/bin lagoon-linter

COPY --from=golang /app/build-deploy-tool /usr/local/bin/build-deploy-tool

# enable running unprivileged
RUN fix-permissions /home && fix-permissions /kubectl-build-deploy

ENTRYPOINT ["/sbin/tini", "--", "/lagoon/entrypoints.sh"]
CMD ["/kubectl-build-deploy/build-deploy.sh"]
