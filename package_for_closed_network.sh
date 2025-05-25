#!/bin/bash

# Mattermost 폐쇄망 배포용 패키징 스크립트 (안전 버전)
# 작성일: 2025-05-25 (수정됨)

CURR_DIR=$(pwd)
PACKAGE_DIR="mattermost_package"
TARGET_FILE="mattermost_closed_network.tar.gz"
SPLIT_SIZE="1.9G"

echo "[📦] Mattermost 폐쇄망 배포 패키지 생성 시작..."

# 작업 디렉토리 생성
mkdir -p "$PACKAGE_DIR"

# 1. Mattermost 디렉토리 복사
MATTERMOST_DIR="$CURR_DIR/mattermost"
echo "[🔍] Mattermost 디렉토리: $MATTERMOST_DIR"
if [ ! -d "$MATTERMOST_DIR" ]; then
    echo "❌ Error: $MATTERMOST_DIR 디렉토리가 없습니다!"
    exit 1
fi

echo "[📁] Mattermost 파일 복사 중..."
rsync -av --exclude=".git*" --exclude=".vscode" --exclude="*.log" \
    --exclude="tmp" --exclude="logs" --exclude="*.swp" \
    "$MATTERMOST_DIR/" "$PACKAGE_DIR/mattermost/"

# 2. PostgreSQL 복사
for DIR in "postgresql-14.10" "pgsql"; do
    echo "[📁] $DIR 복사 중..."
    if [ ! -d "$CURR_DIR/$DIR" ]; then
        echo "❌ Error: $DIR 디렉토리가 없습니다!"
        exit 1
    fi
    cp -r "$CURR_DIR/$DIR" "$PACKAGE_DIR/"
done

# 3. Go 복사
GO_DIR="$CURR_DIR/go-1.21.5"
echo "[📁] Go 복사 중..."
if [ ! -d "$GO_DIR" ]; then
    echo "❌ Error: $GO_DIR 디렉토리가 없습니다!"
    exit 1
fi
cp -r "$GO_DIR" "$PACKAGE_DIR/"

# 4. pgdata 복사
PGDATA_DIR="$CURR_DIR/pgdata"
echo "[📁] DB 데이터 복사 중..."
if [ ! -d "$PGDATA_DIR" ]; then
    echo "❌ Error: $PGDATA_DIR 디렉토리가 없습니다!"
    exit 1
fi
cp -r "$PGDATA_DIR" "$PACKAGE_DIR/"

# 5. 데스크탑 앱 복사 (선택)
DESKTOP_APP="$CURR_DIR/mattermost-desktop-5.10.0-win-x64.msi"
if [ -f "$DESKTOP_APP" ]; then
    cp "$DESKTOP_APP" "$PACKAGE_DIR/"
else
    echo "⚠️ Warning: 데스크탑 앱이 없어 생략합니다."
fi

# 6. 설치/복원 스크립트 복사
for SCRIPT in "install.sh" "restore_package.sh"; do
    if [ -f "$CURR_DIR/$SCRIPT" ]; then
        cp "$CURR_DIR/$SCRIPT" "$PACKAGE_DIR/"
        chmod +x "$PACKAGE_DIR/$SCRIPT"
    else
        echo "⚠️ Warning: $SCRIPT 없음 (필요 시 수동 복사)"
    fi
done

# 7. 압축 (gzip)
echo "[🗜️] tar.gz 파일 생성 중..."
tar -czf "$TARGET_FILE" "$PACKAGE_DIR"

# 8. 분할 압축
echo "[🔄] ${SPLIT_SIZE} 단위로 분할 중..."
split -b "$SPLIT_SIZE" -d -a 3 "$TARGET_FILE" "${TARGET_FILE}.part"

# 9. tar.gz 원본 삭제 (split 후)
rm "$TARGET_FILE"

# 10. 결과 안내
PART_COUNT=$(ls -1 ${TARGET_FILE}.part* | wc -l)
TOTAL_SIZE=$(du -ch ${TARGET_FILE}.part* | grep total | cut -f1)

echo ""
echo "✅ 패키징 완료!"
echo "📁 총 ${PART_COUNT}개의 파일로 분할되었으며, 총 크기: ${TOTAL_SIZE}"
echo "📝 파일명 형식: ${TARGET_FILE}.part000, part001, ..."
echo ""
echo "📤 폐쇄망으로 다음 파일들을 이동하세요:"
echo "  - ${TARGET_FILE}.part*** (분할 파일 전부)"
echo "  - restore_package.sh"
echo ""
echo "📦 복원 절차:"
echo "  1) cat ${TARGET_FILE}.part* > ${TARGET_FILE}"
echo "  2) tar -xzf ${TARGET_FILE}"
echo "  3) cd $PACKAGE_DIR && ./install.sh"
echo ""
echo "💡 원본 소스 및 파일들은 삭제되지 않았으며 안전하게 보존되어 있습니다."
