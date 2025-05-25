#!/bin/bash

# Mattermost 폐쇄망 배포용 패키징 스크립트
# 작성일: 2025-05-25

CURR_DIR=$(pwd)
PACKAGE_DIR="mattermost_package"
TARGET_FILE="mattermost_closed_network.tar"
SPLIT_SIZE="2G"  # 이메일 첨부 제한 (2GB)

echo "Mattermost 폐쇄망 배포 패키지 생성 시작..."

# 작업 디렉토리 생성
mkdir -p $PACKAGE_DIR

# 현재 디렉토리 구조 확인
echo "현재 디렉토리 구조:"
ls -la

# Mattermost 서버 파일 복사 (.git 등 불필요한 파일 제외)
echo "Mattermost 서버 파일 복사 중..."

# Mattermost 절대 경로 사용
MATTERMOST_DIR="$CURR_DIR/mattermost"
echo "Mattermost 디렉토리: $MATTERMOST_DIR"

# 디렉토리 존재 확인
if [ ! -d "$MATTERMOST_DIR" ]; then
    echo "Error: $MATTERMOST_DIR 디렉토리가 존재하지 않습니다!"
    echo "현재 위치: $(pwd)"
    exit 1
fi

rsync -av --exclude=".git" --exclude=".gitlab" \
    --exclude=".github" --exclude=".vscode" --exclude="*.log" \
    --exclude="tmp" --exclude="logs" --exclude="*.swp" \
    "$MATTERMOST_DIR/" "$PACKAGE_DIR/mattermost/"

# PostgreSQL 복사
echo "PostgreSQL 파일 복사 중..."
PG_DIR="$CURR_DIR/postgresql-14.10"
PGSQL_DIR="$CURR_DIR/pgsql"

if [ ! -d "$PG_DIR" ]; then
    echo "Error: $PG_DIR 디렉토리가 존재하지 않습니다!"
    exit 1
fi

if [ ! -d "$PGSQL_DIR" ]; then
    echo "Error: $PGSQL_DIR 디렉토리가 존재하지 않습니다!"
    exit 1
fi

cp -r "$PG_DIR" "$PACKAGE_DIR/"
cp -r "$PGSQL_DIR" "$PACKAGE_DIR/"

# Go 복사
echo "Go 환경 복사 중..."
GO_DIR="$CURR_DIR/go-1.21.5"

if [ ! -d "$GO_DIR" ]; then
    echo "Error: $GO_DIR 디렉토리가 존재하지 않습니다!"
    exit 1
fi

cp -r "$GO_DIR" "$PACKAGE_DIR/"

# 데이터베이스 내용(사용자 계정, 메시지 등) 복사
echo "데이터베이스 내용 복사 중..."
PGDATA_DIR="$CURR_DIR/pgdata"

if [ ! -d "$PGDATA_DIR" ]; then
    echo "Error: $PGDATA_DIR 디렉토리가 존재하지 않습니다!"
    exit 1
fi

cp -r "$PGDATA_DIR" "$PACKAGE_DIR/"

# 데스크톱 앱 복사
echo "Mattermost 데스크톱 앱 복사 중..."
DESKTOP_APP="$CURR_DIR/mattermost-desktop-5.10.0-win-x64.msi"

if [ ! -f "$DESKTOP_APP" ]; then
    echo "Warning: $DESKTOP_APP 파일이 존재하지 않습니다!"
    echo "데스크톱 앱은 포함되지 않습니다."
else
    cp "$DESKTOP_APP" "$PACKAGE_DIR/"
fi

# 기존 설치 및 복원 스크립트 복사
echo "설치 및 복원 스크립트 복사 중..."
INSTALL_SCRIPT="$CURR_DIR/install.sh"
RESTORE_SCRIPT="$CURR_DIR/restore_package.sh"

if [ ! -f "$INSTALL_SCRIPT" ]; then
    echo "Error: $INSTALL_SCRIPT 파일이 존재하지 않습니다!"
    exit 1
fi

cp "$INSTALL_SCRIPT" "$PACKAGE_DIR/"
chmod +x "$PACKAGE_DIR/install.sh"

# 복원 스크립트도 패키지에 포함
if [ -f "$RESTORE_SCRIPT" ]; then
    cp "$RESTORE_SCRIPT" "$PACKAGE_DIR/"
    chmod +x "$PACKAGE_DIR/restore_package.sh"
    echo "복원 스크립트도 패키지에 포함되었습니다."
fi

# 패키지 압축 - 이메일 첨부용 분할 압축 (200MB 단위)
echo "패키지 압축 중..."

# 먼저 단일 파일로 압축
tar -cf $TARGET_FILE $PACKAGE_DIR

# 분할 압축 실행
echo "파일을 ${SPLIT_SIZE} 단위로 분할하여 압축하는 중..."
split -b $SPLIT_SIZE $TARGET_FILE "${TARGET_FILE}.part"

# 원본 tar 파일 삭제
rm $TARGET_FILE

# 분할된 파일 개수와 용량 표시
echo -e "\
분할 압축 파일 정보:"
FILE_COUNT=$(ls -1 ${TARGET_FILE}.part* | wc -l)
TOTAL_SIZE=$(du -ch ${TARGET_FILE}.part* | grep total | cut -f1)
echo "총 $FILE_COUNT개의 파일로 분할되었습니다. 총 크기: $TOTAL_SIZE"

# 처음과 마지막 파일만 보여주기
echo "처음 파일:"
ls -lh $(ls -1 ${TARGET_FILE}.part* | head -n 1)
echo "마지막 파일:"
ls -lh $(ls -1 ${TARGET_FILE}.part* | tail -n 1)

echo -e "\
패키징 완료! 파일이 총 $FILE_COUNT개의 부분으로 분할되어 생성되었습니다."

# 파일명 규칙 설명
echo "파일명 규칙: ${TARGET_FILE}.part?? - 숫자 순서로 자동 생성됩니다 (aa, ab, ac, ...)"
echo "예시: ${TARGET_FILE}.partaa, ${TARGET_FILE}.partab, ..."

echo "이 파일들(${TARGET_FILE}.part*)과 restore_package.sh 스크립트를 함께 폐쇄망 환경으로 이동하세요."
echo "폐쇄망에서 restore_package.sh를 실행하여 파일을 복원한 후 install.sh를 실행하세요."