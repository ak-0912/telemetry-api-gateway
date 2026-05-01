FROM mcr.microsoft.com/devcontainers/go:1.25

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends postgresql-client \
    && apt-get clean && rm -rf /var/lib/apt/lists/*
