#!/bin/bash

# Build 옵션 확인
BUILD_FLAG=false
for arg in "$@"; do
    if [ "$arg" = "build" ]; then
        BUILD_FLAG=true
        break
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

if [ "$BUILD_FLAG" = true ]; then
    echo "Building webapp..."
    cd webapp/channels

    # Node.js 의존성 설치
    if [ ! -d "node_modules" ]; then
        echo "Installing Node.js dependencies..."
        npm install
    fi

    # 웹앱 빌드
    echo "Building webapp files..."
    npm run build

    cd $MATTERMOST_ROOT

    echo "Checking for built webapp..."
    if [ -d "webapp/channels/dist" ]; then
        echo "Found built webapp files in dist directory"
        echo "Copying dist files to client directory..."
    else
        echo "Error: Build failed - dist directory not found"
        exit 1
    fi
else
    echo "Skipping build process..."
    if [ ! -d "webapp/channels/dist" ]; then
        echo "Error: dist directory not found. Please run with 'build' parameter first."
        exit 1
    fi
    echo "Using existing build files from dist directory"
fi
echo "Creating client directory..."
rm -rf client
mkdir -p client

echo "Copying dist files to client directory..."
cp -rv webapp/channels/dist/* client/ || {
    echo "Error copying dist files"
    exit 1
}

echo "Copying root.html to client directory..."
cp -v webapp/channels/src/root.html client/ || {
    echo "Error copying root.html"
    exit 1
}

echo "Starting Mattermost server..."
cd $MATTERMOST_ROOT
nohup ./bin/mattermost server --config ./server/config/config.json > mattermost.log 2>&1 &
echo "Waiting for server to start..."
sleep 5

echo "Deploy completed! Server should be running now."
echo "You can check the logs with: tail -f mattermost.log"

# PostgreSQL 서버 상태 출력
$PG_CTL status -D $PGDATA