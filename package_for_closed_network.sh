#!/bin/bash

# Mattermost 폐쇄망 배포용 패키징 스크립트
# 작성일: 2025-05-25

CURR_DIR=$(pwd)
PACKAGE_DIR="mattermost_package"
TARGET_FILE="mattermost_closed_network.tar"
SPLIT_SIZE="200m"  # 이메일 첨부 제한 (200MB)

echo "Mattermost 폐쇄망 배포 패키지 생성 시작..."

# 작업 디렉토리 생성
mkdir -p $PACKAGE_DIR

# Mattermost 서버 파일 복사 (.git 등 불필요한 파일 제외)
echo "Mattermost 서버 파일 복사 중..."
rsync -av --exclude=".git" --exclude=".gitlab" \
    --exclude=".github" --exclude=".vscode" --exclude="*.log" \
    --exclude="tmp" --exclude="logs" --exclude="*.swp" \
    ./mattermost/ $PACKAGE_DIR/mattermost/

# PostgreSQL 복사
echo "PostgreSQL 파일 복사 중..."
cp -r ./postgresql-14.10 $PACKAGE_DIR/
cp -r ./pgsql $PACKAGE_DIR/

# Go 복사
echo "Go 환경 복사 중..."
cp -r ./go-1.21.5 $PACKAGE_DIR/

# 데이터베이스 내용(사용자 계정, 메시지 등) 복사
echo "데이터베이스 내용 복사 중..."
cp -r ./pgdata $PACKAGE_DIR/

# 데스크톱 앱 복사
echo "Mattermost 데스크톱 앱 복사 중..."
cp ./mattermost-desktop-5.10.0-win-x64.msi $PACKAGE_DIR/

# 기존 설치 및 복원 스크립트 복사
echo "설치 및 복원 스크립트 복사 중..."
cp ./install.sh $PACKAGE_DIR/
chmod +x $PACKAGE_DIR/install.sh

# 패키지 압축 - 이메일 첨부용 분할 압축 (200MB 단위)
echo "패키지 압축 중..."

# 먼저 단일 파일로 압축
tar -cf $TARGET_FILE $PACKAGE_DIR

# 분할 압축 실행
echo "파일을 ${SPLIT_SIZE} 단위로 분할하여 압축하는 중..."
split -b $SPLIT_SIZE $TARGET_FILE "${TARGET_FILE}.part"

# 원본 tar 파일 삭제
rm $TARGET_FILE

# 분할된 파일 목록 출력
echo -e "\n분할 압축 파일 목록:"
ls -lh ${TARGET_FILE}.part*

echo -e "\n패키징 완료! 파일이 여러 개의 부분으로 분할되어 생성되었습니다."
echo "이 파일들(${TARGET_FILE}.part*)과 restore_package.sh 스크립트를 함께 폐쇄망 환경으로 이동하세요."
echo "폐쇄망에서 restore_package.sh를 실행하여 파일을 복원한 후 install.sh를 실행하세요."