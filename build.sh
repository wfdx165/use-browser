#!/bin/bash

# use-browser Build Script
# Cross-platform build automation with automatic version increment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="use-browser"
DIST_DIR="dist"
CMD_PATH="./cmd/browser"

# Platforms to build
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to get the latest git tag
get_latest_tag() {
    local latest_tag
    latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "$latest_tag"
}

# Function to increment version (patch level)
increment_version() {
    local version=$1
    # Remove 'v' prefix if present
    version=${version#v}
    
    # Split version into parts
    local major minor patch
    IFS='.' read -r major minor patch <<< "$version"
    
    # Increment patch version
    patch=$((patch + 1))
    
    echo "${major}.${minor}.${patch}"
}

# Function to get git commit hash
get_commit_hash() {
    git rev-parse --short HEAD 2>/dev/null || echo "unknown"
}

# Function to get build date
get_build_date() {
    date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown"
}

# Function to clean build artifacts
clean() {
    print_info "Cleaning build artifacts..."
    rm -rf "$DIST_DIR"
    print_success "Clean complete."
}

# Function to build for a specific platform
build_platform() {
    local platform=$1
    local version=$2
    local commit=$3
    local date=$4
    
    local goos=${platform%/*}
    local goarch=${platform#*/}
    local dir_name="${BINARY_NAME}-v${version}-${goos}-${goarch}"
    
    # Windows needs .exe extension
    local binary_name="use-browser"
    if [ "$goos" = "windows" ]; then
        binary_name="use-browser.exe"
    fi
    
    print_info "Building for ${platform}..."
    
    # Create dist directory if not exists
    mkdir -p "$DIST_DIR"
    
    # Create platform directory
    local platform_dir="$DIST_DIR/$dir_name"
    mkdir -p "$platform_dir"
    
    # Build binary
    local ldflags="-X github.com/wfdx165/use-browser/pkg/version.Version=${version} -X github.com/wfdx165/use-browser/pkg/version.Commit=${commit} -X github.com/wfdx165/use-browser/pkg/version.Date=${date}"
    
    GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -ldflags "$ldflags" -o "$platform_dir/$binary_name" "$CMD_PATH"
    
    # Copy README and LICENSE if they exist
    if [ -f "README.md" ]; then
        cp "README.md" "$platform_dir/"
    fi
    
    # Check for LICENSE (various possible names)
    for license_file in LICENSE LICENSE.md COPYING; do
        if [ -f "$license_file" ]; then
            cp "$license_file" "$platform_dir/"
            break
        fi
    done
    
    # Create archive based on platform
    if [ "$goos" = "windows" ]; then
        # Windows: ZIP format
        (cd "$DIST_DIR" && zip -q -r "${dir_name}.zip" "$dir_name")
        print_success "Created: $DIST_DIR/${dir_name}.zip"
    else
        # Linux/macOS: tar.gz format
        (cd "$DIST_DIR" && tar -czf "${dir_name}.tar.gz" "$dir_name")
        print_success "Created: $DIST_DIR/${dir_name}.tar.gz"
    fi
    
    # Clean up the directory
    rm -rf "$platform_dir"
}

# Function to generate checksums
generate_checksums() {
    print_info "Generating checksums..."
    cd "$DIST_DIR"
    
    # Combine all archive files
    local archives="*.tar.gz *.zip"
    
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum *.tar.gz *.zip 2>/dev/null > checksums.txt
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 *.tar.gz *.zip 2>/dev/null > checksums.txt
    else
        print_warning "Neither sha256sum nor shasum found. Skipping checksum generation."
        cd ..
        return
    fi
    
    print_success "Checksums written to $DIST_DIR/checksums.txt"
    cd ..
}

# Function to show help
show_help() {
    cat << EOF
use-browser Build Script

Usage:
    $0                      Build all platforms
    $0 <platform>           Build specific platform
    $0 --tag                Build all and create git tag
    $0 --clean              Clean build artifacts
    $0 --version            Show version info
    $0 --help               Show this help message

Platforms:
    linux-amd64
    linux-arm64
    darwin-amd64
    darwin-arm64
    windows-amd64

Examples:
    $0                      # Build all platforms
    $0 darwin-arm64         # Build only for macOS Apple Silicon
    $0 --tag                # Build all and create vX.X.X tag
    $0 --clean              # Clean dist/ directory

EOF
}

# Function to show version info
show_version() {
    local latest_tag
    latest_tag=$(get_latest_tag)
    local next_version
    next_version=$(increment_version "$latest_tag")
    local commit
    commit=$(get_commit_hash)
    local date
    date=$(get_build_date)
    
    echo "Current Git Tag:    $latest_tag"
    echo "Next Build Version: v$next_version"
    echo "Git Commit:         $commit"
    echo "Build Date:         $date"
}

# Function to create git tag
create_tag() {
    local version=$1
    local tag_name="v${version}"
    
    print_info "Creating git tag: $tag_name"
    git tag "$tag_name"
    print_success "Tag created: $tag_name"
    print_info "To push tag: git push origin $tag_name"
}

# Main build function
build_all() {
    local create_git_tag=false
    
    # Check if --tag flag is passed
    for arg in "$@"; do
        if [ "$arg" = "--tag" ]; then
            create_git_tag=true
        fi
    done
    
    # Get version info
    local latest_tag
    latest_tag=$(get_latest_tag)
    local version
    version=$(increment_version "$latest_tag")
    local commit
    commit=$(get_commit_hash)
    local date
    date=$(get_build_date)
    
    print_info "Building $BINARY_NAME v$version..."
    print_info "Commit: $commit"
    print_info "Date: $date"
    
    # Create dist directory
    mkdir -p "$DIST_DIR"
    
    # Build all platforms
    for platform in "${PLATFORMS[@]}"; do
        build_platform "$platform" "$version" "$commit" "$date"
    done
    
    # Generate checksums
    generate_checksums
    
    print_success "Build complete. Artifacts in $DIST_DIR/"
    
    # List artifacts
    echo ""
    echo "Build artifacts:"
    ls -lh "$DIST_DIR/"
    
    # Create git tag if requested
    if [ "$create_git_tag" = true ]; then
        echo ""
        create_tag "$version"
    fi
}

# Function to build single platform
build_single() {
    local platform=$1
    
    # Validate platform
    local valid=false
    for p in "${PLATFORMS[@]}"; do
        if [ "$p" = "$platform" ]; then
            valid=true
            break
        fi
    done
    
    if [ "$valid" = false ]; then
        print_error "Invalid platform: $platform"
        echo "Valid platforms:"
        for p in "${PLATFORMS[@]}"; do
            echo "  $p"
        done
        exit 1
    fi
    
    # Get version info
    local latest_tag
    latest_tag=$(get_latest_tag)
    local version
    version=$(increment_version "$latest_tag")
    local commit
    commit=$(get_commit_hash)
    local date
    date=$(get_build_date)
    
    print_info "Building $BINARY_NAME v$version for $platform..."
    
    # Create dist directory
    mkdir -p "$DIST_DIR"
    
    # Build
    build_platform "$platform" "$version" "$commit" "$date"
    
    print_success "Build complete."
}

# Main script logic
case "${1:-}" in
    --help|-h)
        show_help
        exit 0
        ;;
    --version|-v)
        show_version
        exit 0
        ;;
    --clean|-c)
        clean
        exit 0
        ;;
    --tag)
        build_all "$@"
        exit 0
        ;;
    "")
        build_all "$@"
        exit 0
        ;;
    *)
        # Check if it's a platform or invalid argument
        if [[ "$1" == *"/"* ]] || [[ "$1" == *"-"* ]]; then
            # Convert linux-amd64 to linux/amd64 format
            platform="${1//-//}"
            build_single "$platform"
        else
            print_error "Unknown option: $1"
            echo "Run '$0 --help' for usage information."
            exit 1
        fi
        ;;
esac
