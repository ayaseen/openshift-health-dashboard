#!/bin/bash
set -e

# Define variables
DASHBOARD_DIR="app/dashboard"
OUTPUT_DIR="../web/static"
RECHARTS_VERSION="2.7.2"
REACT_VERSION="18.2.0"

# Define image variables
REGISTRY="quay-quay-registry.apps.ocp.rhlab.dev"
NAMESPACE="ayaseen"
IMAGE_NAME="dashboard"
TAG="v0.1.1"
IMAGE="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}:${TAG}"
GO_VERSION="1.24.2" # Updated Go version

echo "=== Building OpenShift Health Dashboard ==="

# Force remove previous builds to ensure clean state
rm -rf $DASHBOARD_DIR/build
rm -rf $DASHBOARD_DIR/node_modules
rm -rf app/web/static/*

# Navigate to the dashboard directory

cd app/dashboard

# Install fresh dependencies

echo "Installing dependencies..."
npm install --save lucide-react
npm install --no-audit --no-fund

# Force rebuild of CSS
echo "Rebuilding CSS..."
npx tailwindcss -i ./src/index.css -o ./src/tailwind.css

# Build the dashboard with production settings
echo "Building the dashboard with production configuration..."
NODE_ENV=production npm run build

# Verify build directory exists and contains index.html
if [ ! -d "build" ] || [ ! -f "build/index.html" ]; then
  echo "ERROR: Build failed - build directory or index.html missing"
  exit 1
fi

# Clear web/static directory and copy new files
echo "Copying built files to web/static..."
mkdir -p ../web/static
cp -r build/* ../web/static/

# Verify copied files
echo "Files in web/static:"
ls -la ../web/static
echo "JS files in web/static:"
ls -la ../web/static/js

# Return to the original directory
cd ../..

echo "=== Dashboard Build Complete ==="

podman system prune --all --force


echo "=== Building OpenShift Health Dashboard Image ==="
echo "Target image: ${IMAGE}"

# Detect architecture
ARCH=$(uname -m)
echo "Host architecture: ${ARCH}"


# Ensure the static directory has the dashboard file
if [ ! -f "app/web/static/index.html" ]; then
    echo "ERROR: The dashboard file (app/web/static/index.html) is missing."
    echo "Please create the app/web/static/index.html file before running this script."
    exit 1
fi

# Build the Go binary with correct architecture
echo "Building Go binary for Linux/amd64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/manager app/server/main.go



# Verify binary architecture
file bin/manager
echo "Binaries built successfully"

# Create Dockerfile with Ruby 3.0
echo "Creating Dockerfile..."
cat > Dockerfile << 'INNEREOF'
FROM registry.access.redhat.com/ubi8:latest

# Set non-root user ID for OpenShift compatibility
# Use a fixed UID/GID that works with OpenShift's random UID assignment
ENV USER_ID=1001 \
    GROUP_ID=0

# Labels required by OpenShift
LABEL name="openshift-health-operator" \
      maintainer="ayaseen@redhat.com" \
      vendor="Red Hat" \
      version="v0.1.2" \
      release="1" \
      summary="OpenShift Health Check Operator" \
      description="Provides comprehensive health checks for OpenShift clusters"

# Set up Go and Ruby in a single layer to reduce image size
RUN yum module reset -y ruby && \
    yum module enable -y ruby:3.0 && \
    yum install -y ruby ruby-devel \
        gcc gcc-c++ make zlib-devel \
        redhat-rpm-config tar gzip diffutils curl procps-ng && \
    curl -LO https://go.dev/dl/go1.24.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz && \
    rm go1.24.2.linux-amd64.tar.gz && \
    gem install asciidoctor --no-document && \
    gem install asciidoctor-pdf --no-document && \
    yum clean all && \
    rm -rf /var/cache/yum

# Set Go environment variables
ENV PATH=$PATH:/usr/local/go/bin \
    GOROOT=/usr/local/go

# Create directories with proper permissions
RUN mkdir -p /tmp/health-reports /tmp/health-checks /web/static && \
    chgrp -R 0 /tmp/health-reports /tmp/health-checks /web/static && \
    chmod -R g=u /tmp/health-reports /tmp/health-checks /web/static

# Configure working directory
WORKDIR /opt/app-root/src

# Copy the binary
COPY bin/manager /usr/local/bin/manager

# Clear and recreate web directory to ensure fresh files
RUN rm -rf /web/static && mkdir -p /web/static

# Copy web assets for dashboard
COPY app/web/static/ /web/static/

# Set permissions for OpenShift random UID compatibility
RUN chmod -R g+rwX /web/static && \
    chmod +x /usr/local/bin/manager && \
    chgrp -R 0 /usr/local/bin/manager && \
    chmod -R g=u /usr/local/bin/manager

# Create startup script with proper permissions
RUN echo '#!/bin/bash' > /usr/local/bin/start.sh && \
    echo 'echo "Starting OpenShift Health Dashboard on port $PORT"' >> /usr/local/bin/start.sh && \
    echo 'exec /usr/local/bin/manager' >> /usr/local/bin/start.sh && \
    chmod +x /usr/local/bin/start.sh && \
    chgrp 0 /usr/local/bin/start.sh && \
    chmod g=u /usr/local/bin/start.sh

# Expose port
EXPOSE 8082

# Set environment variables for the dashboard
ENV STATIC_DIR=/web/static \
    PORT=8080 \
    DEBUG=false

# Switch to non-root user
USER ${USER_ID}

# Start the server
ENTRYPOINT ["/usr/local/bin/start.sh"]

INNEREOF

# Build the container image
echo "Building container image..."
if [ "$ARCH" = "arm64" ]; then
    # For Apple Silicon Macs, explicitly target amd64
    podman build --platform=linux/amd64 -t "${IMAGE}" .
else
    # For Intel Macs
    podman build -t "${IMAGE}" .
fi

# Login to registry
echo "Logging in to registry ${REGISTRY}..."
podman login ${REGISTRY}

# Push the image
echo "Pushing image ${IMAGE}..."
podman push "${IMAGE}"

echo "=== Image build and push completed successfully ==="
echo "Image: ${IMAGE}"