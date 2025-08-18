#!/usr/bin/env bash
set -euo pipefail

mkdir -p bin
allowed_archs="amd64 arm arm64 386"
echo "Building binaries for allowed architectures: $allowed_archs"
for var in $(go tool dist list); do
    # skip disallowed archs
    if [[ ! $allowed_archs =~ "$(cut -d '/' -f 2 <<<$var)" ]]; then
        echo "Skipping: $var"
        continue
    fi
    # skip arm for windows
    if [[ "$(cut -d '/' -f 1 <<<$var)" == "windows" && "$(cut -d '/' -f 2 <<<$var)" == "arm" ]]; then
        echo "Skipping: $var (windows/arm)"
        continue
    fi
    
    file_name="jiotv_go-$(cut -d '/' -f 1 <<< $var)-$(cut -d '/' -f 2 <<< $var)"
    case "$(cut -d '/' -f 1 <<< $var)" in
        "windows")
            echo "Building $var"
            CGO_ENABLED=0 GOOS="$(cut -d '/' -f 1 <<< $var)" GOARCH="$(cut -d '/' -f 2 <<< $var)" go build -o bin/"$file_name.exe" -trimpath -ldflags="-s -w" . &
        ;;
        "linux" | "darwin")
            echo "Building $var"
            CGO_ENABLED=0 GOOS="$(cut -d '/' -f 1 <<< $var)" GOARCH="$(cut -d '/' -f 2 <<< $var)" go build -o bin/"$file_name" -trimpath -ldflags="-s -w" . &
        ;;
        "android")
            echo "Building $var"
            case "$(cut -d '/' -f 2 <<<$var)" in
                "arm")
                    CGO_ENABLED=1 GOOS="$(cut -d '/' -f 1 <<<$var)" GOARCH="$(cut -d '/' -f 2 <<<$var)" CC="armv7a-linux-androideabi28-clang" CXX="armv7a-linux-androideabi28-clang++" go build -o bin/"jiotv_go-$(cut -d '/' -f 1 <<<$var)-$(cut -d '/' -f 2 <<<$var)" -trimpath -ldflags="-s -w" .
                ;;
                "arm64")
                    CGO_ENABLED=1 GOOS="$(cut -d '/' -f 1 <<<$var)" GOARCH="$(cut -d '/' -f 2 <<<$var)" CC="aarch64-linux-android32-clang" CXX="aarch64-linux-android32-clang++" go build -o bin/"jiotv_go-$(cut -d '/' -f 1 <<<$var)-$(cut -d '/' -f 2 <<<$var)" -trimpath -ldflags="-s -w" .
                ;;
                "amd64")
                    CGO_ENABLED=1 GOOS="$(cut -d '/' -f 1 <<<$var)" GOARCH="$(cut -d '/' -f 2 <<<$var)" CC="x86_64-linux-android32-clang" CXX="x86_64-linux-android32-clang++" go build -o bin/"jiotv_go-$(cut -d '/' -f 1 <<<$var)-$(cut -d '/' -f 2 <<<$var)" -trimpath -ldflags="-s -w" .
                ;;
                *)
                    echo "Skipping: $var"
                ;;
            esac
        ;;
        *)
            echo "Skipping: $var"
        ;;
    esac
done

# Wait for all background jobs to finish
wait

# Build for android5 arm with CC=armv7a-linux-androideabi21-clang
echo "Building android5 arm"
CGO_ENABLED=1 GOOS=android GOARCH=arm GOARM=7 CC="armv7a-linux-androideabi21-clang" CXX="armv7a-linux-androideabi21-clang++" go build -o bin/jiotv_go-android5-armv7 -trimpath -ldflags="-s -w" .
