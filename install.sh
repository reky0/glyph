#!/usr/bin/env bash
# glyph installer — Linux & macOS
#
# Interactive usage:
#   ./install.sh
#
# Non-interactive / scripted:
#   ./install.sh --all
#   ./install.sh pin ask
#   curl -fsSL https://github.com/reky0/glyph/releases/latest/download/install.sh | bash -s -- --all
#
# Environment overrides:
#   GLYPH_VERSION=v0.1.0   install a specific version (default: latest)
#   GLYPH_INSTALL_DIR=...  override install location  (default: ~/.local/bin)

set -euo pipefail

REPO="reky0/glyph"
TOOLS=(pin ask diff stand)
declare -A DESC=(
    [pin]="clipboard for URLs, commands, paths and notes"
    [ask]="ask an AI with automatic directory context"
    [diff]="explain a git diff in plain English"
    [stand]="generate a standup from recent git commits"
)

# ── colours ──────────────────────────────────────────────────────────────────
if [ -t 1 ]; then
    BOLD='\033[1m'; DIM='\033[2m'; RESET='\033[0m'
    GREEN='\033[32m'; YELLOW='\033[33m'; RED='\033[31m'; CYAN='\033[36m'
else
    BOLD=''; DIM=''; RESET=''; GREEN=''; YELLOW=''; RED=''; CYAN=''
fi

info()    { printf "  ${CYAN}→${RESET}  %s\n" "$*"; }
success() { printf "  ${GREEN}✓${RESET}  %s\n" "$*"; }
warn()    { printf "  ${YELLOW}⚠${RESET}  %s\n" "$*"; }
die()     { printf "  ${RED}✗${RESET}  %s\n" "$*" >&2; exit 1; }

# ── helpers ───────────────────────────────────────────────────────────────────
detect_platform() {
    local os arch
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    case "$os" in
        linux)  ;;
        darwin) ;;
        *) die "Unsupported OS: $os" ;;
    esac
    case "$arch" in
        x86_64)        arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) die "Unsupported architecture: $arch" ;;
    esac
    echo "${os}-${arch}"
}

fetch() {
    local url="$1" dest="$2"
    if command -v curl &>/dev/null; then
        curl -fsSL --progress-bar "$url" -o "$dest"
    elif command -v wget &>/dev/null; then
        wget -q --show-progress "$url" -O "$dest"
    else
        die "Neither curl nor wget found. Install one and retry."
    fi
}

latest_version() {
    local tag
    if command -v curl &>/dev/null; then
        tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
            | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
    else
        tag=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" \
            | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
    fi
    [ -n "$tag" ] || die "Could not fetch latest version from GitHub."
    echo "$tag"
}

# ── banner ────────────────────────────────────────────────────────────────────
print_banner() {
    printf "\n"
    printf "  ${BOLD}glyph installer${RESET}\n"
    printf "  ${DIM}Small CLI tools for your terminal memory${RESET}\n"
    printf "\n"
}

# ── interactive tool selection ─────────────────────────────────────────────────
select_tools() {
    printf "  ${BOLD}Available tools:${RESET}\n\n"
    for i in "${!TOOLS[@]}"; do
        local t="${TOOLS[$i]}"
        printf "    ${CYAN}%d${RESET}  %-7s  ${DIM}%s${RESET}\n" \
            "$((i+1))" "$t" "${DESC[$t]}"
    done
    printf "\n"
    printf "  Enter numbers to install (e.g. ${BOLD}1 3${RESET}), or ${BOLD}a${RESET} for all:\n"
    printf "  > "

    local input
    read -r input

    local selected=()
    if [[ "$input" == "a" || "$input" == "all" ]]; then
        selected=("${TOOLS[@]}")
    else
        for token in $input; do
            if [[ "$token" =~ ^[1-4]$ ]]; then
                selected+=("${TOOLS[$((token-1))]}")
            else
                warn "Ignoring unknown selection: $token"
            fi
        done
    fi
    [ "${#selected[@]}" -gt 0 ] || die "No tools selected."
    echo "${selected[@]}"
}

# ── install a single tool ──────────────────────────────────────────────────────
install_tool() {
    local tool="$1" version="$2" platform="$3" install_dir="$4"
    local url="https://github.com/${REPO}/releases/download/${version}/glyph-${tool}-${platform}.tar.gz"
    local tmpdir
    tmpdir=$(mktemp -d)
    trap "rm -rf '$tmpdir'" RETURN

    info "Downloading ${BOLD}${tool}${RESET} ${version} (${platform})..."
    fetch "$url" "${tmpdir}/archive.tar.gz"
    tar -xzf "${tmpdir}/archive.tar.gz" -C "$tmpdir"

    mkdir -p "$install_dir"
    mv "${tmpdir}/${tool}" "${install_dir}/${tool}"
    chmod +x "${install_dir}/${tool}"
    success "Installed ${BOLD}${tool}${RESET}  →  ${install_dir}/${tool}"
}

# ── PATH check ────────────────────────────────────────────────────────────────
check_path() {
    local install_dir="$1"
    if echo ":${PATH}:" | grep -q ":${install_dir}:"; then
        return
    fi
    printf "\n"
    warn "${install_dir} is not in your PATH."
    printf "\n"
    printf "  Add it with:\n"
    printf "    ${BOLD}export PATH=\"%s:\$PATH\"${RESET}\n" "$install_dir"
    printf "\n"

    # Offer to append to shell profiles that exist.
    local profiles=()
    for f in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
        [ -f "$f" ] && profiles+=("$f")
    done
    if [ "${#profiles[@]}" -gt 0 ] && [ -t 0 ]; then
        printf "  Append to shell profile(s) automatically? [Y/n] "
        local ans
        read -r ans
        if [[ "$ans" == "" || "$ans" =~ ^[Yy] ]]; then
            for f in "${profiles[@]}"; do
                printf '\nexport PATH="%s:$PATH"\n' "$install_dir" >> "$f"
                success "Updated $f"
            done
            printf "\n"
            info "Restart your shell or run:  ${BOLD}source ~/.bashrc${RESET}"
        fi
    fi
}

# ── diff warning ──────────────────────────────────────────────────────────────
warn_diff_shadow() {
    local install_dir="$1"
    if command -v diff &>/dev/null; then
        local existing
        existing=$(command -v diff)
        if [[ "$existing" != "${install_dir}/diff" ]]; then
            printf "\n"
            warn "The ${BOLD}diff${RESET} binary will shadow the system ${DIM}${existing}${RESET}."
            warn "If that causes issues, rename it:  ${BOLD}mv ${install_dir}/diff ${install_dir}/gdiff${RESET}"
        fi
    fi
}

# ── main ──────────────────────────────────────────────────────────────────────
main() {
    print_banner

    local platform version install_dir
    platform=$(detect_platform)
    version="${GLYPH_VERSION:-$(latest_version)}"
    install_dir="${GLYPH_INSTALL_DIR:-$HOME/.local/bin}"

    printf "  Version  : ${BOLD}%s${RESET}\n" "$version"
    printf "  Platform : ${BOLD}%s${RESET}\n" "$platform"
    printf "  Install  : ${BOLD}%s${RESET}\n" "$install_dir"
    printf "\n"

    # Parse arguments.
    local selected=()
    if [ "$#" -eq 0 ]; then
        if [ -t 0 ]; then
            read -ra selected <<< "$(select_tools)"
        else
            # stdin is a pipe (curl | bash) — install all.
            warn "Non-interactive mode: installing all tools."
            selected=("${TOOLS[@]}")
        fi
    elif [ "$1" = "--all" ]; then
        selected=("${TOOLS[@]}")
    else
        for arg in "$@"; do
            local valid=0
            for t in "${TOOLS[@]}"; do
                [ "$arg" = "$t" ] && { selected+=("$arg"); valid=1; break; }
            done
            [ "$valid" -eq 1 ] || warn "Unknown tool '${arg}', skipping."
        done
        [ "${#selected[@]}" -gt 0 ] || die "No valid tools specified. Available: ${TOOLS[*]}"
    fi

    printf "\n"
    for tool in "${selected[@]}"; do
        install_tool "$tool" "$version" "$platform" "$install_dir"
    done

    # Warn about diff shadow after install.
    for t in "${selected[@]}"; do
        [ "$t" = "diff" ] && { warn_diff_shadow "$install_dir"; break; }
    done

    check_path "$install_dir"

    printf "\n"
    success "${BOLD}Done!${RESET} Run any tool with ${BOLD}--help${RESET} to get started."
    printf "\n"
}

main "$@"
