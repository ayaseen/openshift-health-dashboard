#!/bin/bash
# Cleanup script for OpenShift Health Dashboard

echo "🧹 Cleaning up unnecessary files and directories..."

# Create a backup first
echo "📦 Creating backup folder..."
mkdir -p backup

# Backup important files just in case
echo "💾 Backing up important files..."
cp -r go.mod go.sum README.md backup/

# Remove operator-specific directories
echo "🗑️ Removing operator-specific directories..."
rm -rf cmd
rm -rf config
rm -rf deploy

# Remove duplicate/unnecessary directories
echo "🗑️ Removing duplicate directories..."
rm -rf dashboard-temp
rm -rf staticserver
rm -rf pkg

# Remove old binaries
echo "🗑️ Removing old binaries..."
rm -rf bin/manager
rm -rf bin/serve

# Remove old web directory (now moved to app/web)
echo "🗑️ Removing old web directory..."
rm -rf web

# Remove old deployment scripts
echo "🗑️ Removing operator deployment scripts..."
rm -f deploy-operator.sh

# Make sure our main directories exist
echo "📁 Creating necessary directories..."
mkdir -p app/web/static
mkdir -p bin

echo "✅ Cleanup completed!"
echo "📁 Project structure now contains only necessary components."
echo "🔍 Make sure to run build-image.sh to build the new dashboard!"