#!/usr/bin/env bash

set -o errtrace
set -o errexit

echo "----------------------------------------------------"
echo "------------------ gitfullstory --------------------"
echo "--------------- @author Todd Matthews --------------"
echo "--------------- @date 12.09.2018 -------------------"

# Installed using command: curl -sL https://donfuego.github.io/gitfullstory/scripts/install.sh | bash
# Install file created using: tar -zcvf gitfullstory.tar.gz ./gitfullstory

# Create destination folder
INSTALL_DIR="/usr/local/bin"

# Temporary write directory to store downloaded/unpacked install file
TEMP_DIR="/tmp"

# Location of install
INSTALL_URL="https://donfuego.github.io/gitfullstory/scripts"
INSTALL_FILE="gitfullstory.tar.gz"
INSTALL_BINARY="gitfullstory"

# Our makeshift logger command
log()  { printf "%b\n" "$*"; }

# Our makeshift debug logger
debug(){ [[ ${gitfullstory_debug_flag:-0} -eq 0 ]] || printf "%b\n" "$*" >&2; }

# Our makeshift fail logger
fail() { log "\nERROR: $*\n" >&2 ; exit 1 ; }

# Initialize anything needed during setup - correct version of bash, which, grep and curl
gitfullstory_install_initialize()
{
    log "Checking for install requirements..."
    BASH_MIN_VERSION="3.2.25"
    if
        [[ -n "${BASH_VERSION:-}" &&
        "$(\printf "%b" "${BASH_VERSION:-}\n${BASH_MIN_VERSION}\n" | LC_ALL=C \sort -t"." -k1,1n -k2,2n -k3,3n | \head -n1)" != "${BASH_MIN_VERSION}"
        ]]
    then
        echo "BASH ${BASH_MIN_VERSION} required (you have $BASH_VERSION)"
        exit 1
    fi

    \which which >/dev/null 2>&1 || fail "Could not find 'which' command, make sure it's available first before continuing installation."
    \which grep >/dev/null 2>&1 || fail "Could not find 'grep' command, make sure it's available first before continuing installation."
    \which curl >/dev/null 2>&1 || fail "Could not find 'curl' command, make sure it's available first before continuing installation."
    \which tar >/dev/null 2>&1 || fail "Could not find 'tar' command, make sure it's available first before continuing installation."
}

# Retrieves compressed binary from remote location
gitfullstory_get_package()
{
  _url="${INSTALL_URL}/${INSTALL_FILE}"
  
  log "Downloading ${_url}"
  __gitfullstory_curl -sS ${_url} > ${TEMP_DIR}/${INSTALL_FILE} && log "Download complete" ||
  {
    _return=$?
    case $_return in
      (*)
        log "Could not download '${_url}'. curl returned status '$_return'."
        ;;
    esac
    return $_return
  }
}

# Unpackages the binary
# TODO: Add checksum
gitfullstory_unpack() 
{
    log "Unpacking binary..."
    # Unpack binary
    __gitfullstory_debug_command cd ${TEMP_DIR} && __gitfullstory_debug_command tar xzf ./${INSTALL_FILE} && log "Binary unpacked" ||
    {
        _return=$?
        log "Could not extract gitfullstory binary."
        return $_return
    }
}

# Copies the uncompressed binary from the downloaded folder into a folder on the user's path i.e. /usr/local/bin
gitfullstory_copy()
{
    log "Copying binary to ${INSTALL_DIR}"
    # Unpack binary
    __gitfullstory_debug_command mv ${TEMP_DIR}/${INSTALL_BINARY} ${INSTALL_DIR} && echo "Copy complete" ||
    {
        _return=$?
        log "Could not copy binary to ${INSTALL_DIR}"
        return $_return
    }
}

# Removes any temporary files downloaded/created during install
gitfullstory_cleanup()
{
    # Cleanup
    __gitfullstory_debug_command rm ${TEMP_DIR}/${INSTALL_FILE}
}

# Post installation output instructions
gitfullstory_post_install()
{
    log "Installation Successful!"
    log "Make sure '/usr/local/bin' is on your PATH"
    log "type 'gitfullstory -help' for usage details"
}

# Wrapper for user's curl command
__gitfullstory_curl()
(
  typeset -a __flags
  __flags=( --fail --location --max-redirs 10 )

  [[ "$*" == *"--max-time"* ]] ||
  [[ "$*" == *"--connect-timeout"* ]] ||
    __flags+=( --connect-timeout 30 --retry-delay 2 --retry 3 )

  __gitfullstory_curl_output_control

  unset curl
  __gitfullstory_debug_command \curl "${__flags[@]}" "$@" || return $?
)

# Handle's how the user's curl commands outputs errors
__gitfullstory_curl_output_control()
{
  if
    [[ " $*" == *" -s"* || " $*" == *" --silent"* ]]
  then
    # make sure --show-error is used with --silent
    [[ " $*" == *" -S"* || " $*" == *" -sS"* || " $*" == *" --show-error"* ]] ||
    {
      __flags+=( "--show-error" )
    }
  fi
}

# Used for debugging local commands
__gitfullstory_debug_command()
{
  debug "Running($#): $*"
  "$@" || return $?
  true
}

# Outputs an error string passed into this function
gitfullstory_error()  
{
    printf "ERROR: %b\n" "$*"; 
}

# Main install orchestration method, handles order of execution
gitfullstory_install() 
{
    log "Installing gitfullstory"
    gitfullstory_install_initialize
    gitfullstory_get_package
    gitfullstory_unpack
    gitfullstory_copy
    gitfullstory_cleanup
    gitfullstory_post_install
}

# Kick off the install process
gitfullstory_install