ARG BASE_IMAGE=m.daocloud.io/docker.io/library/debian:bookworm-slim
FROM ${BASE_IMAGE}

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    fonts-dejavu-core \
    fonts-noto-cjk \
    libreoffice-core \
    libreoffice-writer \
  && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["soffice"]
