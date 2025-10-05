#!/bin/bash
# Version management script for NCloud Server Controller

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
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

# Function to get current version from git tags
get_current_version() {
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [[ -z "$latest_tag" ]]; then
        echo "0.0.0"
    else
        echo "${latest_tag#v}"  # Remove 'v' prefix
    fi
}

# Function to get next version based on type
get_next_version() {
    local current_version=$1
    local version_type=$2
    
    # Split version into parts
    IFS='.' read -ra VERSION_PARTS <<< "$current_version"
    local major=${VERSION_PARTS[0]}
    local minor=${VERSION_PARTS[1]}
    local patch=${VERSION_PARTS[2]}
    
    case $version_type in
        "major")
            echo "$((major + 1)).0.0"
            ;;
        "minor")
            echo "$major.$((minor + 1)).0"
            ;;
        "patch")
            echo "$major.$minor.$((patch + 1))"
            ;;
        *)
            echo "$major.$minor.$((patch + 1))"
            ;;
    esac
}

# Function to update version in files
update_version_files() {
    local new_version=$1
    
    print_status "Updating version to $new_version in files..."
    
    # Update Makefile
    sed -i.bak "s/VERSION ?= .*/VERSION ?= $new_version/" Makefile
    rm -f Makefile.bak
    
    # Update Helm Chart
    sed -i.bak "s/^version: .*/version: $new_version/" helm/ncloud-server-controller/Chart.yaml
    sed -i.bak "s/^appVersion: .*/appVersion: \"$new_version\"/" helm/ncloud-server-controller/Chart.yaml
    rm -f helm/ncloud-server-controller/Chart.yaml.bak
    
    # Update go.mod if needed (for module version)
    if [[ -f "go.mod" ]]; then
        # This is optional - go.mod doesn't typically need version updates
        print_status "go.mod version management is handled by Go modules"
    fi
    
    print_success "Version updated in all files"
}

# Function to create git tag
create_git_tag() {
    local version=$1
    local tag="v$version"
    
    print_status "Creating git tag: $tag"
    
    # Check if tag already exists
    if git tag -l | grep -q "^$tag$"; then
        print_error "Tag $tag already exists!"
        exit 1
    fi
    
    # Create and push tag
    git tag -a "$tag" -m "Release $tag"
    git push origin "$tag"
    
    print_success "Tag $tag created and pushed"
}

# Function to show version information
show_version_info() {
    local current_version=$(get_current_version)
    local next_patch=$(get_next_version "$current_version" "patch")
    local next_minor=$(get_next_version "$current_version" "minor")
    local next_major=$(get_next_version "$current_version" "major")
    
    echo "📦 Version Information"
    echo "===================="
    echo "Current version: $current_version"
    echo ""
    echo "Next versions:"
    echo "  Patch (bug fixes): $next_patch"
    echo "  Minor (new features): $next_minor"
    echo "  Major (breaking changes): $next_major"
    echo ""
    echo "Usage:"
    echo "  $0 patch    - Create patch release ($next_patch)"
    echo "  $0 minor    - Create minor release ($next_minor)"
    echo "  $0 major    - Create major release ($next_major)"
    echo "  $0 info     - Show this information"
}

# Main script logic
main() {
    local command=${1:-"info"}
    
    case $command in
        "patch"|"minor"|"major")
            local current_version=$(get_current_version)
            local new_version=$(get_next_version "$current_version" "$command")
            
            print_status "Creating $command release: $current_version -> $new_version"
            
            # Check if working directory is clean
            if [[ -n $(git status --porcelain) ]]; then
                print_error "Working directory is not clean. Please commit or stash changes first."
                exit 1
            fi
            
            # Update version files
            update_version_files "$new_version"
            
            # Commit version changes
            git add Makefile helm/ncloud-server-controller/Chart.yaml
            git commit -m "chore: bump version to $new_version"
            
            # Create and push tag
            create_git_tag "$new_version"
            
            print_success "Release $new_version created successfully!"
            print_status "GitHub Actions will now build and deploy the new version"
            ;;
        "info")
            show_version_info
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            show_version_info
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
