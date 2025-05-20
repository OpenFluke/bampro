#!/bin/bash

# Exit on any error
set -e

# Function to detect Linux distribution
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
    else
        echo "Error: Cannot detect distribution. /etc/os-release not found."
        exit 1
    fi
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Docker and Docker Compose on Ubuntu
install_ubuntu() {
    # Update package index
    sudo apt-get update -y

    # Install prerequisites
    sudo apt-get install -y ca-certificates curl gnupg lsb-release

    # Remove old Docker installations, if any
    sudo apt-get remove -y docker docker-engine docker.io containerd runc || true

    # Add Docker's official GPG key
    if [ ! -f /etc/apt/keyrings/docker.gpg ]; then
        sudo mkdir -p /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        sudo chmod a+r /etc/apt/keyrings/docker.gpg
    fi

    # Set up the Docker repository
    if [ ! -f /etc/apt/sources.list.d/docker.list ]; then
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
        $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    fi

    # Update package index and install Docker
    sudo apt-get update -y
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    # Start and enable Docker service
    sudo systemctl enable docker.service
    sudo systemctl enable containerd.service
    sudo systemctl start docker.service
}

# Function to install Docker and Docker Compose on Fedora
install_fedora() {
    # Remove conflicting Docker installations, including Fedora's moby-engine and docker-compose
    sudo dnf remove -y docker docker-client docker-client-latest docker-common docker-latest docker-latest-logrotate docker-logrotate docker-selinux docker-engine-selinux docker-engine moby-engine docker-cli docker-compose moby-filesystem moby-engine-nano docker-compose-switch docker-buildx || true

    # Install dnf-plugins-core if not present
    if ! rpm -q dnf-plugins-core >/dev/null 2>&1; then
        sudo dnf -y install dnf-plugins-core
    fi

    # Add Docker repository
    if ! sudo dnf repolist --enabled | grep -q docker-ce; then
        if ! sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo 2>/dev/null; then
            echo "Warning: dnf config-manager --add-repo failed. Adding repo manually."
            sudo bash -c 'cat > /etc/yum.repos.d/docker-ce.repo <<EOF
[docker-ce-stable]
name=Docker CE Stable - \$basearch
baseurl=https://download.docker.com/linux/fedora/\$releasever/\$basearch/stable
enabled=1
gpgcheck=1
gpgkey=https://download.docker.com/linux/fedora/gpg
EOF'
        fi
    fi

    # Install Docker and Compose with --allowerasing to resolve conflicts
    sudo dnf -y install --allowerasing docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    # Start and enable Docker service
    sudo systemctl enable docker.service
    sudo systemctl enable containerd.service
    sudo systemctl start docker.service
}

# Function to configure Docker for non-root user
configure_docker() {
    # Add user to docker group
    if ! getent group docker >/dev/null; then
        sudo groupadd docker
    fi
    if ! groups $USER | grep -q docker; then
        sudo usermod -aG docker $USER
        echo "Added $USER to docker group. Please log out and back in to apply, or run 'newgrp docker' in this session."
    fi
}

# Main script
echo "Detecting Linux distribution..."
detect_distro

echo "Detected distribution: $DISTRO"

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo "This script must be run as root or with sudo."
    exit 1
fi

# Install based on distribution
case "$DISTRO" in
    ubuntu)
        install_ubuntu
        ;;
    fedora)
        install_fedora
        ;;
    *)
        echo "Error: Unsupported distribution: $DISTRO. This script supports Ubuntu and Fedora only."
        exit 1
        ;;
esac

# Configure Docker for non-root access
configure_docker

# Verify installations
if command_exists docker; then
    echo "Docker installed successfully: $(docker --version)"
else
    echo "Error: Docker installation failed."
    exit 1
fi

if command_exists docker-compose; then
    echo "Docker Compose installed successfully: $(docker compose version)"
else
    echo "Error: Docker Compose installation failed."
    exit 1
fi

# Verify Docker daemon is running
if ! systemctl is-active --quiet docker; then
    echo "Error: Docker daemon is not running."
    exit 1
fi

echo "Installation and configuration complete!"
echo "You can now use 'docker compose' (note: not 'docker-compose') for Docker Compose V2."