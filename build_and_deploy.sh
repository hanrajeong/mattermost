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

echo "Building webapp..."
cd $MATTERMOST_ROOT/webapp
npm run build

echo "Copying dist files to client directory..."
cd $MATTERMOST_ROOT
# client 디렉토리 생성 (없는 경우)
mkdir -p client
# 기존 파일 백업 (옵션)
if [ -d "client" ] && [ "$(ls -A client)" ]; then
    mv client client_backup_$(date +%Y%m%d_%H%M%S)
fi
# dist 파일들을 client 디렉토리로 직접 복사
cp -rv $MATTERMOST_ROOT/webapp/channels/dist/* $MATTERMOST_ROOT/client/

echo "Starting Mattermost server..."
cd $MATTERMOST_ROOT
nohup ./bin/mattermost server --config ./server/config/config.json > mattermost.log 2>&1 &
echo "Waiting for server to start..."
sleep 5

echo "Deploy completed! Server should be running now."
echo "You can check the logs with: tail -f mattermost.log"

# PostgreSQL 서버 상태 출력
$PG_CTL status -D $PGDATA