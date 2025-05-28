#!/bin/bash

# Build 옵션 확인
BUILD_FLAG=false
SERVER_FLAG=false
REDHAT_COMPAT=false

for arg in "$@"; do
    if [ "$arg" = "build" ]; then
        BUILD_FLAG=true
    elif [ "$arg" = "server" ]; then
        SERVER_FLAG=true
    elif [ "$arg" = "redhat" ]; then
        REDHAT_COMPAT=true
    fi
done

# PostgreSQL 설정
PG_CTL="/home/hanra/fshome/sctadm/pgsql/bin/pg_ctl"
PGDATA="/home/hanra/fshome/sctadm/pgdata"

# PostgreSQL 서버 상태 확인 및 시작
echo "Checking PostgreSQL server status..."
if ! $PG_CTL status -D $PGDATA > /dev/null 2>&1; then
    echo "Starting PostgreSQL server..."
    $PG_CTL start -D $PGDATA -l $PGDATA/logfile
    sleep 3
fi

# Mattermost 루트 디렉토리 설정
MATTERMOST_ROOT="/home/hanra/fshome/sctadm/mattermost"
cd $MATTERMOST_ROOT

echo "Stopping Mattermost server..."
# 실행 중인 프로세스 ID 찾기
MATTERMOST_PID=$(pgrep mattermost)
if [ ! -z "$MATTERMOST_PID" ]; then
    echo "Killing Mattermost process (PID: $MATTERMOST_PID)..."
    kill $MATTERMOST_PID
    # 프로세스가 완전히 종료될 때까지 대기
    while kill -0 $MATTERMOST_PID 2>/dev/null; do
        sleep 1
    done
    echo "Mattermost process stopped"
else
    echo "No Mattermost process found running"
fi

if [ "$SERVER_FLAG" = true ]; then
    echo "Building server..."
    cd server

    # 서버 빌드 (RedHat 8.1 호환성 모드)
    if [ "$REDHAT_COMPAT" = true ]; then
        echo "Building server with RedHat 8.1 compatibility (static linking)..."
        # 정적 링킹으로 빌드 (glibc 의존성 제거)
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ../bin/mattermost ./cmd/mattermost/main.go
    else
        echo "Building server normally..."
        make build-linux
    fi

    # 빌드 결과 확인
    if [ -f "../bin/mattermost" ]; then
        echo "Server build completed successfully"
        # 빌드된 바이너리 정보 출력
        file ../bin/mattermost
        echo "Dynamic dependencies:"
        ldd ../bin/mattermost || echo "Not a dynamic executable (statically linked)"
    else
        echo "Error: Server build failed"
        exit 1
    fi

    cd $MATTERMOST_ROOT
fi

if [ "$BUILD_FLAG" = true ]; then
    echo "Building webapp..."
    cd webapp/channels

    # 기존 빌드 파일 정리
    echo "Cleaning previous build..."
    rm -rf dist

    # Node.js 의존성 설치
    if [ ! -d "node_modules" ]; then
        echo "Installing Node.js dependencies..."
        npm install
    fi

    # 웹앱 빌드
    echo "Building webapp files..."
    NODE_ENV=production npm run build

    if [ ! -d "dist" ]; then
        echo "Error: Build failed - dist directory not found"
        exit 1
    fi

    echo "Build completed successfully"
    cd $MATTERMOST_ROOT
else
    echo "Skipping build process..."
    if [ ! -d "webapp/channels/dist" ]; then
        echo "Error: dist directory not found. Please run with 'build' parameter first."
        exit 1
    fi
    echo "Using existing build files from dist directory"
fi

# client 디렉토리 준비
echo "Preparing client directory..."
rm -rf client
mkdir -p client

# dist 파일 복사
echo "Copying dist files to client directory..."
if cp -r webapp/channels/dist/* client/ 2>/dev/null; then
    echo "Files copied successfully"
else
    echo "Error: Failed to copy dist files"
    exit 1
fi

echo "Starting Mattermost server..."
cd $MATTERMOST_ROOT
nohup ./bin/mattermost server --config ./server/config/config.json > mattermost.log 2>&1 &
echo "Waiting for server to start..."
sleep 5

echo "Deploy completed! Server should be running now."
echo "You can check the logs with: tail -f mattermost.log"
echo ""
echo "Build options:"
echo "  ./build_and_deploy.sh build      - Build webapp"
echo "  ./build_and_deploy.sh server     - Build server only"
echo "  ./build_and_deploy.sh redhat     - Build with RedHat 8.1 compatibility"
echo "  ./build_and_deploy.sh server redhat - Build server with RedHat 8.1 compatibility"

# PostgreSQL 서버 상태 출력
$PG_CTL status -D $PGDATA