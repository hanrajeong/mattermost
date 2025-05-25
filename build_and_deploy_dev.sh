#!/bin/bash

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

echo "Building webapp for development..."
cd webapp/channels

# 기존 빌드 파일 정리
echo "Cleaning previous build..."
rm -rf dist

# Node.js 의존성 설치
if [ ! -d "node_modules" ]; then
    echo "Installing Node.js dependencies..."
    npm install
fi

# 웹앱 개발 모드 빌드 (더 빠름)
echo "Building webapp files in development mode..."
NODE_ENV=development npm run build

if [ ! -d "dist" ]; then
    echo "Error: Build failed - dist directory not found"
    exit 1
fi

echo "Development build completed successfully"
cd $MATTERMOST_ROOT

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

# PostgreSQL 서버 상태 출력
$PG_CTL status -D $PGDATA
