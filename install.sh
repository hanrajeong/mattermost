#!/bin/bash

# Mattermost 폐쇄망 설치 스크립트
# 작성일: 2025-05-25

INSTALL_DIR="/root/fshome/sctadm"
CURR_DIR=$(pwd)

echo "Mattermost 폐쇄망 설치를 시작합니다..."

# 디렉토리 생성
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# 파일 복사
echo "파일 복사 중..."
cp -r $CURR_DIR/mattermost ./
cp -r $CURR_DIR/postgresql-14.10 ./
cp -r $CURR_DIR/pgsql ./
cp -r $CURR_DIR/go-1.21.5 ./
cp $CURR_DIR/mattermost-desktop-5.10.0-win-x64.msi ./

# 환경 변수 설정
echo "환경 변수 설정 중..."
export PATH=$INSTALL_DIR/go-1.21.5/bin:$PATH
export GOROOT=$INSTALL_DIR/go-1.21.5
export PATH=$INSTALL_DIR/pgsql/bin:$PATH

# PostgreSQL 데이터 디렉토리 설정
echo "PostgreSQL 데이터 설정 중..."
if [ -d "$CURR_DIR/pgdata" ]; then
    # 기존 pgdata가 포함되어 있는 경우 (사용자 계정, 메시지 등 보존)
    if [ -d "$INSTALL_DIR/pgdata" ]; then
        echo "기존 pgdata 백업 중..."
        mv $INSTALL_DIR/pgdata $INSTALL_DIR/pgdata.bak.$(date +%Y%m%d%H%M%S)
    fi
    cp -r $CURR_DIR/pgdata $INSTALL_DIR/
    
    # 설정 파일 확인 및 수정
    if ! grep -q "listen_addresses = '*'" $INSTALL_DIR/pgdata/postgresql.conf; then
        echo "listen_addresses = '*'" >> $INSTALL_DIR/pgdata/postgresql.conf
    fi
    if ! grep -q "host all all 0.0.0.0/0 md5" $INSTALL_DIR/pgdata/pg_hba.conf; then
        echo "host all all 0.0.0.0/0 md5" >> $INSTALL_DIR/pgdata/pg_hba.conf
    fi
else
    echo "Error: pgdata 디렉토리가 패키지에 포함되어 있지 않습니다!"
    exit 1
fi

# PostgreSQL 시작
echo "PostgreSQL 시작 중..."
$INSTALL_DIR/pgsql/bin/pg_ctl -D $INSTALL_DIR/pgdata start

# Mattermost 서버 시작
echo "Mattermost 서버 시작 중..."
cd $INSTALL_DIR/mattermost
./bin/mattermost &

echo "설치가 완료되었습니다!"
echo "Mattermost 서버는 http://localhost:8065 에서 접속할 수 있습니다."
echo "데스크톱 앱은 Windows 환경에서 mattermost-desktop-5.10.0-win-x64.msi 파일을 설치하세요."
