#!/usr/bin/env bash
OS_LIST=(darwin windows linux)
ARCH_LIST=(amd64 386)
PROJECT_NAME=${PWD##*/}
BUILD_FOLDER="dist"

# create temporary build folder
TMP=$(mktemp -d 2>/dev/null || mktemp -d -t 'mytmpdir')

function compile_go_binary() {
    local os=${1}
    local arch=${2}

    local output_filename="$PROJECT_NAME-$os-$arch"
    if [ "$os" = "windows" ]; then
        output_filename="${output_filename}.exe"
    fi

    GOOS=$os GOARCH=$arch go build -o "$TMP/$output_filename"
}

function prepare_static_folder() {
    (
        cd photoboxrecents-web
        npm install && npm run build
    )

    mkdir "$TMP/photoboxrecents-web"
    cp -rf photoboxrecents-web/dist "$TMP/photoboxrecents-web"
}

for os in "${OS_LIST[@]}"; do
    for arch in "${ARCH_LIST[@]}"; do
        echo "Compile $os-$arch"
        compile_go_binary $os $arch
    done
done

prepare_static_folder

DATE=$(date +"%Y%m%d-%H%M%S")
(cd $TMP && zip -r - .) > "$BUILD_FOLDER/$PROJECT_NAME-$DATE.zip"
rm -rf $TMP