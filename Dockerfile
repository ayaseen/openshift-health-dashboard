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

