#!/bin/bash
set -e

# Define variables
DASHBOARD_DIR="dashboard-temp"
OUTPUT_DIR="../web/static"
RECHARTS_VERSION="2.7.2"
REACT_VERSION="18.2.0"

# Define variables
REGISTRY="quay-quay-registry.apps.ocp.rhlab.dev"
NAMESPACE="ayaseen"
IMAGE_NAME="operator"
TAG="v0.1.1"
IMAGE="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}:${TAG}"
GO_VERSION="1.24.2" # Updated Go version

echo "=== Building OpenShift Health Dashboard ==="

# Navigate to the dashboard directory
cd $DASHBOARD_DIR

# Check if required files exist
if [ ! -f "src/index.js" ]; then
  echo "ERROR: Missing src/index.js file. Please check that all required files are present."
  exit 1
fi

if [ ! -f "public/index.html" ]; then
  echo "ERROR: Missing public/index.html file. Please check that all required files are present."
  exit 1
fi

# Clean previous node_modules to avoid conflicts
echo "Cleaning node_modules..."
rm -rf node_modules
rm -f package-lock.json


# Clear output directory to avoid stale files
echo "Clearing output directory..."
rm -rf ../web/static

# Create the output directory
mkdir -p ../web/static

# Ensure public directory has basic files
mkdir -p public


# Check and add Webpack dependencies if needed
if ! grep -q '"webpack":' package.json; then
  npm install --save-dev webpack@5.88.2 webpack-cli@5.1.4 webpack-dev-server@4.15.1
  npm install --save-dev html-webpack-plugin@5.5.3 mini-css-extract-plugin@2.7.6
  npm install --save-dev babel-loader@9.1.3 @babel/core@7.22.10 @babel/preset-env@7.22.10 @babel/preset-react@7.22.5
  npm install --save-dev css-loader@6.8.1 style-loader@3.3.3

fi

# Install dependencies with clean install
npm install -D tailwindcss postcss autoprefixer
echo "Installing dependencies..."
npm ci || npm install --no-audit --no-fund

# Add required dependencies if missing
if ! grep -q "recharts" package.json; then
  echo "Adding recharts dependency..."
  npm install --save recharts@$RECHARTS_VERSION
fi

if ! grep -q "react " package.json; then
  echo "Adding React dependencies..."
  npm install --save react@$REACT_VERSION react-dom@$REACT_VERSION
fi

# Ensure Tailwind CSS is properly installed and configured
echo "Setting up Tailwind CSS..."
npx tailwindcss init -p

# Force rebuild of CSS
echo "Rebuilding CSS..."
npx tailwindcss -i ./src/index.css -o ./src/tailwind.css
cp ./src/tailwind.css ./src/index.css


# Build the dashboard
echo "Building the dashboard..."
npm run build


# Copy the built files to the output directory
echo "Copying built files to $OUTPUT_DIR..."
cp -r build/* ../web/static/

# Verify index.html was created
if [ -f "../web/static/index.html" ]; then
  echo "Dashboard build completed successfully!"
  echo "Files have been copied to ../web/static/"
else
  echo "ERROR: Failed to create index.html in output directory."
  exit 1
fi

# Return to the original directory
cd -

echo "=== Dashboard Build Complete ==="




echo "=== Building OpenShift Health Operator Image ==="
echo "Target image: ${IMAGE}"

# Detect architecture
ARCH=$(uname -m)
echo "Host architecture: ${ARCH}"


# Ensure the static directory has the dashboard file
if [ ! -f "web/static/index.html" ]; then
    echo "ERROR: The dashboard file (web/static/index.html) is missing."
    echo "Please create the web/static/index.html file before running this script."
    exit 1
fi

# Build the Go binary with correct architecture
echo "Building Go binary for Linux/amd64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/manager cmd/manager/main.go

# Build the static file server from the separate directory
echo "Building static file server..."
cd staticserver
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../bin/serve staticserver.go
cd ..

# Verify binary architecture
file bin/manager
file bin/serve
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

# Copy the binaries
COPY bin/manager /usr/local/bin/manager
COPY bin/serve /usr/local/bin/serve

# Copy web assets for dashboard
COPY web/static/ /web/static/

# Set permissions for OpenShift random UID compatibility
RUN chmod -R g+rwX /web/static && \
    chmod +x /usr/local/bin/manager /usr/local/bin/serve && \
    chgrp -R 0 /usr/local/bin/manager /usr/local/bin/serve && \
    chmod -R g=u /usr/local/bin/manager /usr/local/bin/serve

# Create startup script with proper permissions
RUN echo '#!/bin/bash' > /usr/local/bin/start.sh && \
    echo 'echo "Starting static file server on port $STATIC_PORT"' >> /usr/local/bin/start.sh && \
    echo '/usr/local/bin/serve &' >> /usr/local/bin/start.sh && \
    echo 'echo "Starting manager service on port 8082"' >> /usr/local/bin/start.sh && \
    echo 'exec /usr/local/bin/manager' >> /usr/local/bin/start.sh && \
    chmod +x /usr/local/bin/start.sh && \
    chgrp 0 /usr/local/bin/start.sh && \
    chmod g=u /usr/local/bin/start.sh

# Expose ports
EXPOSE 8082
EXPOSE 8084

# Set environment variables
ENV STATIC_DIR=/web/static \
    STATIC_PORT=8084 \
    API_TARGET=http://localhost:8082

# Switch to non-root user
USER ${USER_ID}

# Use the startup script
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