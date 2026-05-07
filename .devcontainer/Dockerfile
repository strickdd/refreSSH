FROM mcr.microsoft.com/devcontainers/base:bookworm

# Install any additional OS packages needed for Go development on Linux
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    build-essential \
    curl \
    git \
    && apt-get clean && rm -rf /var/lib/apt/lists/*
