FROM golang:1.24

# Set global environment variables
ENV GOPATH=/go
ENV PATH=/usr/local/go/bin:$GOPATH/bin:$PATH
ENV GOBIN=$GOPATH/bin
ENV SHELL=/bin/bash

# Install additional OS packages
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    sqlite3 \
    postgresql-client \
    git \
    curl \
    bash \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Create a non-root user
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && apt-get update \
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    && chsh -s /bin/bash $USERNAME

# Create workspace directory and set permissions
RUN mkdir -p /workspace/tmp && \
    chown -R $USERNAME:$USERNAME /workspace && \
    mkdir -p "$GOPATH" && \
    chown -R $USERNAME:$USERNAME "$GOPATH"

# Add environment variables to shell profile
RUN echo "export GOPATH=/go" >> /home/$USERNAME/.bashrc && \
    echo "export GOBIN=/go/bin" >> /home/$USERNAME/.bashrc && \
    echo "export PATH=/usr/local/go/bin:$GOPATH/bin:$PATH" >> /home/$USERNAME/.bashrc && \
    echo "export SHELL=/bin/bash" >> /home/$USERNAME/.bashrc

# Switch to non-root user
USER $USERNAME

# Install Air from source
RUN git clone https://github.com/cosmtrek/air.git && \
    cd air && \
    go build -o /go/bin/air . && \
    cd .. && \
    rm -rf air

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Set working directory
WORKDIR /workspace 