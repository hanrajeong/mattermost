#!/bin/bash

# Mattermost 폐쇄망 배포용 패키징 스크립트
# 작성일: 2025-05-25

CURR_DIR=$(pwd)
PACKAGE_DIR="mattermost_package"
TARGET_FILE="mattermost_closed_network.tar.gz"

echo "Mattermost 폐쇄망 배포 패키지 생성 시작..."

# 작업 디렉토리 생성
mkdir -p $PACKAGE_DIR

# Mattermost 서버 파일 복사 (.git 등 불필요한 파일 제외)
# node_modules는 폐쇄망에서 빌드/수정이 필요할 경우를 대비해 포함
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

# 데스크톱 앱 복사
echo "Mattermost 데스크톱 앱 복사 중..."
cp ./mattermost-desktop-5.10.0-win-x64.msi $PACKAGE_DIR/

# 설치 스크립트 생성
cat > $PACKAGE_DIR/install.sh << 'EOF'
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
if [ ! -d "$INSTALL_DIR/pgdata" ]; then
    echo "PostgreSQL 데이터 디렉토리 초기화 중..."
    mkdir -p $INSTALL_DIR/pgdata
    $INSTALL_DIR/pgsql/bin/initdb -D $INSTALL_DIR/pgdata
    
    # PostgreSQL 설정 파일 수정
    echo "PostgreSQL 설정 수정 중..."
    echo "listen_addresses = '*'" >> $INSTALL_DIR/pgdata/postgresql.conf
    echo "host all all 0.0.0.0/0 md5" >> $INSTALL_DIR/pgdata/pg_hba.conf
fi

# PostgreSQL 시작
echo "PostgreSQL 시작 중..."
$INSTALL_DIR/pgsql/bin/pg_ctl -D $INSTALL_DIR/pgdata start

# Mattermost 데이터베이스 생성
echo "Mattermost 데이터베이스 생성 중..."
$INSTALL_DIR/pgsql/bin/createuser -s mattermost || true
$INSTALL_DIR/pgsql/bin/createdb -O mattermost mattermost_db || true

# Mattermost 설정 파일 업데이트
echo "Mattermost 설정 파일 업데이트 중..."
CONFIG_FILE="$INSTALL_DIR/mattermost/config/config.json"

# 백업 생성
cp $CONFIG_FILE ${CONFIG_FILE}.bak

# 데이터베이스 설정 업데이트
sed -i 's/"DriverName": ".*"/"DriverName": "postgres"/g' $CONFIG_FILE
sed -i 's/"DataSource": ".*"/"DataSource": "postgres:\/\/mattermost:mattermost@localhost:5432\/mattermost_db?sslmode=disable\&connect_timeout=10"/g' $CONFIG_FILE

# 폰트 디렉토리 생성
echo "폰트 디렉토리 생성 중..."
mkdir -p $INSTALL_DIR/mattermost/fonts

# Mattermost 서버 시작
echo "Mattermost 서버 시작 중..."
cd $INSTALL_DIR/mattermost
./bin/mattermost &

echo "설치가 완료되었습니다!"
echo "Mattermost 서버는 http://localhost:8065 에서 접속할 수 있습니다."
echo "데스크톱 앱은 Windows 환경에서 mattermost-desktop-5.10.0-win-x64.msi 파일을 설치하세요."
EOF

# 설치 스크립트 실행 권한 부여
chmod +x $PACKAGE_DIR/install.sh

# 패키지 압축
echo "패키지 압축 중..."
tar -czf $TARGET_FILE $PACKAGE_DIR

echo "패키징 완료! $TARGET_FILE 파일이 생성되었습니다."
echo "이 파일을 폐쇄망 환경으로 이동한 후, 압축을 풀고 install.sh 스크립트를 실행하세요."
