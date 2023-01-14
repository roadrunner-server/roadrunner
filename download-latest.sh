#!/bin/sh

# This script can optionally use a GitHub token to increase your request limit (for example, if using this script in a CI).
# To use a GitHub token, pass it through the GITHUB_PAT environment variable.

# GLOBALS

# Colors
RED='\033[31m'
GREEN='\033[32m'
DEFAULT='\033[0m'

# Project name
PNAME='roadrunner'

# GitHub API address
GITHUB_API='https://api.github.com/repos/roadrunner-server/roadrunner/releases'
# GitHub Release address
GITHUB_REL='https://github.com/roadrunner-server/roadrunner/releases/download'

# FUNCTIONS

# Gets the version of the latest stable version of RoadRunner by setting the $latest variable.
# Returns 0 in case of success, 1 otherwise.
get_latest() {
  # temp_file is needed because the grep would start before the download is over
  temp_file=$(mktemp -q /tmp/$PNAME.XXXXXXXXX)
  latest_release="$GITHUB_API/latest"

  if ! temp_file=$(mktemp -q /tmp/$PNAME.XXXXXXXXX); then
    echo "$0: Can't create temp file."
    fetch_release_failure_usage
    exit 1
  fi

  if [ -z "$GITHUB_PAT" ]; then
    curl -s "$latest_release" >"$temp_file" || return 1
  else
    curl -H "Authorization: token $GITHUB_PAT" -s "$latest_release" >"$temp_file" || return 1
  fi

  latest="$(grep <"$temp_file" '"tag_name":' | cut -d ':' -f2 | tr -d '"' | tr -d ',' | tr -d ' ' | tr -d 'v')"
  latestV="$(grep <"$temp_file" '"tag_name":' | cut -d ':' -f2 | tr -d '"' | tr -d ',' | tr -d ' ')"

  rm -f "$temp_file"
  return 0
}

# 0 -> not alpine
# 1 -> alpine
isAlpine() {
  # shellcheck disable=SC2143
  if [ "$(grep <"/etc/os-release" "NAME=" | grep -ic "Alpine")" ]; then
    return 1
  fi

  return 0
}

# Gets the OS by setting the $os variable.
# Returns 0 in case of success, 1 otherwise.
get_os() {
  os_name=$(uname -s)
  case "$os_name" in
  # ---
  'Darwin')
    os='darwin'
    ;;

    # ---
  'Linux')
    os='linux'
    if isAlpine; then
      os="unknown-musl"
    fi
    ;;

    # ---
  'MINGW'*)
    os='windows'
    ;;

    # ---
  *)
    return 1
    ;;
  esac
  return 0
}

# Gets the architecture by setting the $arch variable.
# Returns 0 in case of success, 1 otherwise.
get_arch() {
  architecture=$(uname -m)

  # case 1
  case "$architecture" in
  'x86_64' | 'amd64')
    arch='amd64'
    ;;

    # case 2
  'arm64')
    arch='arm64'
    ;;

  # all other
  *)
    return 1
    ;;
  esac

  return 0
}

get_compress() {
  os_name=$(uname -s)
  case "$os_name" in
  'Darwin')
    compress='tar.gz'
    ;;
  'Linux')
    compress='tar.gz'
    if isAlpine; then
      compress="zip"
    fi
    ;;
  'MINGW'*)
    compress='zip'
    ;;
  *)
    return 1
    ;;
  esac
  return 0
}

not_available_failure_usage() {
  printf "$RED%s\n$DEFAULT" 'ERROR: RoadRunner binary is not available for your OS distribution or your architecture yet.'
  echo ''
  echo 'However, you can easily compile the binary from the source files.'
  echo 'Follow the steps at the page ("Source" tab): TODO'
}

fetch_release_failure_usage() {
  echo ''
  printf "$RED%s\n$DEFAULT" 'ERROR: Impossible to get the latest stable version of RoadRunner.'
  echo 'Please let us know about this issue: https://github.com/roadrunner-server/roadrunner/issues/new/choose'
  echo ''
  echo 'In the meantime, you can manually download the appropriate binary from the GitHub release assets here: https://github.com/roadrunner-server/roadrunner/releases/latest'
}

fill_release_variables() {
  # Fill $latest variable.
  if ! get_latest; then
    fetch_release_failure_usage
    exit 1
  fi
  if [ "$latest" = '' ]; then
    fetch_release_failure_usage
    exit 1
  fi
  # Fill $os variable.
  if ! get_os; then
    not_available_failure_usage
    exit 1
  fi
  # Fill $arch variable.
  if ! get_arch; then
    not_available_failure_usage
    exit 1
  fi

  # Fill $compress variable
  if ! get_compress; then
    not_available_failure_usage
    exit 1
  fi
}

download_binary() {
  fill_release_variables
  echo "Downloading RoadRunner binary $latest for $os, architecture $arch..."
  release_file="$PNAME-$latest-$os-$arch.$compress"

  if ! curl --fail -OL "$GITHUB_REL/$latestV/$release_file"; then
    fetch_release_failure_usage
    exit 1
  fi

  printf "$GREEN%s\n$DEFAULT" "RoadRunner $latest archive successfully downloaded as $release_file"
}

download_binary
