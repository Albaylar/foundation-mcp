#!/usr/bin/env bash
# Really compiles the catalog's components against the generated token file, so a
# component or a token name that does not compile can never reach an app.
# Usage: scripts/compile-catalog.sh swiftui|flutter   (from the repository root)
set -euo pipefail
platform="${1:?usage: compile-catalog.sh swiftui|flutter}"
root="$(pwd)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

go build -o "$work/foundation" ./cmd/foundation

case "$platform" in
  swiftui)
    "$work/foundation" --catalog ./catalog --export swiftui > "$work/FoundationTokens.swift"
    cp catalog/components/*/swiftui/*.swift "$work/"
    cd "$work"
    xcrun swiftc -typecheck -sdk "$(xcrun --sdk iphonesimulator --show-sdk-path)" \
      -target arm64-apple-ios17.0-simulator -swift-version 6 ./*.swift
    ;;
  flutter)
    mkdir -p "$work/app/lib/tokens" "$work/app/lib/components"
    "$work/foundation" --catalog "$root/catalog" --export flutter > "$work/app/lib/tokens/foundation_tokens.dart"
    cp catalog/components/*/flutter/*.dart "$work/app/lib/components/"
    cd "$work/app"
    cat > pubspec.yaml <<'EOF'
name: catalog_check
publish_to: none
environment:
  sdk: ">=3.5.0 <4.0.0"
dependencies:
  flutter:
    sdk: flutter
EOF
    flutter pub get
    flutter analyze --no-pub --no-fatal-infos
    ;;
  *)
    echo "unknown platform: $platform" >&2
    exit 2
    ;;
esac
echo "catalog compiles: $platform"
