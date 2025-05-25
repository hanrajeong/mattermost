#!/bin/bash

# 시스템 설정 검사 스크립트 - 서버 마이그레이션용
# 작성일: 2025-05-25

# 변수 설정
HOSTNAME=$(hostname)
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
OUTPUT_FILE="system_config_${HOSTNAME}_${TIMESTAMP}.txt"

# 헤더 출력
echo "====================================================" | tee -a "$OUTPUT_FILE"
echo "시스템 설정 검사 보고서" | tee -a "$OUTPUT_FILE"
echo "서버 이름: $HOSTNAME" | tee -a "$OUTPUT_FILE"
echo "실행 일시: $(date)" | tee -a "$OUTPUT_FILE"
echo "====================================================" | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# 실행 중인 커널 정보
echo "## 커널 정보" | tee -a "$OUTPUT_FILE"
echo "* 커널 버전:" | tee -a "$OUTPUT_FILE"
uname -a | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# CPU, 메모리 정보
echo "## 하드웨어 정보" | tee -a "$OUTPUT_FILE"
echo "* CPU 정보:" | tee -a "$OUTPUT_FILE"
grep "model name" /proc/cpuinfo | head -1 | tee -a "$OUTPUT_FILE"
echo "* CPU 코어 수:" | tee -a "$OUTPUT_FILE"
grep -c "processor" /proc/cpuinfo | tee -a "$OUTPUT_FILE"
echo "* 메모리 정보:" | tee -a "$OUTPUT_FILE"
free -h | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# OS 버전
echo "## OS 버전 정보" | tee -a "$OUTPUT_FILE"
if [ -f /etc/os-release ]; then
    cat /etc/os-release | tee -a "$OUTPUT_FILE"
elif [ -f /etc/redhat-release ]; then
    cat /etc/redhat-release | tee -a "$OUTPUT_FILE"
else
    echo "OS 버전 파일을 찾을 수 없습니다." | tee -a "$OUTPUT_FILE"
fi
echo "" | tee -a "$OUTPUT_FILE"

# 프로세스 관련 설정
echo "## 프로세스 설정" | tee -a "$OUTPUT_FILE"
echo "* PID 최대값:" | tee -a "$OUTPUT_FILE"
cat /proc/sys/kernel/pid_max | tee -a "$OUTPUT_FILE"
echo "* 최대 사용자 프로세스 수:" | tee -a "$OUTPUT_FILE"
ulimit -u | tee -a "$OUTPUT_FILE"
echo "* 최대 열린 파일 수:" | tee -a "$OUTPUT_FILE"
ulimit -n | tee -a "$OUTPUT_FILE"
echo "* 시스템 전체 최대 열린 파일 수:" | tee -a "$OUTPUT_FILE"
cat /proc/sys/fs/file-max | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# 시스템 한계 설정
echo "## 시스템 한계 설정 (/etc/security/limits.conf)" | tee -a "$OUTPUT_FILE"
if [ -f /etc/security/limits.conf ]; then
    grep -v "^#" /etc/security/limits.conf | grep -v "^$" | tee -a "$OUTPUT_FILE"
else
    echo "limits.conf 파일을 찾을 수 없습니다." | tee -a "$OUTPUT_FILE"
fi
echo "" | tee -a "$OUTPUT_FILE"

# 시스템 설정 (sysctl)
echo "## 시스템 설정 (sysctl)" | tee -a "$OUTPUT_FILE"
echo "* 네트워크 설정:" | tee -a "$OUTPUT_FILE"
sysctl -a | grep net.ipv4 | grep -E 'tcp_max_syn_backlog|tcp_syncookies|tcp_fin_timeout|tcp_tw_reuse|ip_local_port_range|rmem_max|wmem_max' | tee -a "$OUTPUT_FILE"
echo "* 커널 설정:" | tee -a "$OUTPUT_FILE"
sysctl -a | grep -E 'kernel.shmmax|kernel.shmall|kernel.shmmni|kernel.sem|kernel.msgmni|kernel.msgmax' | tee -a "$OUTPUT_FILE"
echo "* 파일 시스템 설정:" | tee -a "$OUTPUT_FILE"
sysctl -a | grep -E 'fs.file-max|fs.aio-max-nr' | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# 디스크 정보
echo "## 디스크 정보" | tee -a "$OUTPUT_FILE"
echo "* 디스크 사용량:" | tee -a "$OUTPUT_FILE"
df -h | tee -a "$OUTPUT_FILE"
echo "* 파일 시스템 타입:" | tee -a "$OUTPUT_FILE"
mount | grep -E '/ |/boot|/home|/tmp|/var' | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# 네트워크 설정
echo "## 네트워크 설정" | tee -a "$OUTPUT_FILE"
echo "* 네트워크 인터페이스:" | tee -a "$OUTPUT_FILE"
ip addr | tee -a "$OUTPUT_FILE"
echo "* 라우팅 테이블:" | tee -a "$OUTPUT_FILE"
ip route | tee -a "$OUTPUT_FILE"
echo "* 네트워크 포트 상태:" | tee -a "$OUTPUT_FILE"
ss -tuln | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# Java 관련 설정 (JDK 버전 등)
echo "## Java 환경 설정" | tee -a "$OUTPUT_FILE"
echo "* Java 버전:" | tee -a "$OUTPUT_FILE"
if command -v java &> /dev/null; then
    java -version 2>&1 | tee -a "$OUTPUT_FILE"
else
    echo "Java가 설치되어 있지 않습니다." | tee -a "$OUTPUT_FILE"
fi
echo "* JAVA_HOME:" | tee -a "$OUTPUT_FILE"
echo $JAVA_HOME | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# WebLogic 관련 설정
echo "## WebLogic 설정" | tee -a "$OUTPUT_FILE"
echo "* WebLogic 설치 경로:" | tee -a "$OUTPUT_FILE"
if [ -n "$WL_HOME" ]; then
    echo "$WL_HOME" | tee -a "$OUTPUT_FILE"
    echo "* WebLogic 버전:" | tee -a "$OUTPUT_FILE"
    if [ -f "$WL_HOME/server/lib/weblogic.jar" ]; then
        java -cp "$WL_HOME/server/lib/weblogic.jar" weblogic.version 2>&1 | head -2 | tee -a "$OUTPUT_FILE"
    else
        echo "WebLogic JAR 파일을 찾을 수 없습니다." | tee -a "$OUTPUT_FILE"
    fi
else
    echo "WebLogic 환경 변수가 설정되지 않았습니다." | tee -a "$OUTPUT_FILE"
fi
echo "" | tee -a "$OUTPUT_FILE"

# 보안 설정
echo "## 보안 설정" | tee -a "$OUTPUT_FILE"
echo "* SELinux 상태:" | tee -a "$OUTPUT_FILE"
if command -v getenforce &> /dev/null; then
    getenforce | tee -a "$OUTPUT_FILE"
else
    echo "SELinux 명령어를 찾을 수 없습니다." | tee -a "$OUTPUT_FILE"
fi
echo "* 방화벽 상태:" | tee -a "$OUTPUT_FILE"
if command -v systemctl &> /dev/null; then
    systemctl status firewalld | grep Active | tee -a "$OUTPUT_FILE"
elif command -v service &> /dev/null; then
    service iptables status | grep -i "running\|active" | tee -a "$OUTPUT_FILE"
else
    echo "방화벽 상태를 확인할 수 없습니다." | tee -a "$OUTPUT_FILE"
fi
echo "" | tee -a "$OUTPUT_FILE"

# 시스템 서비스
echo "## 시스템 서비스" | tee -a "$OUTPUT_FILE"
echo "* 실행 중인 주요 서비스:" | tee -a "$OUTPUT_FILE"
if command -v systemctl &> /dev/null; then
    systemctl list-units --type=service --state=running | grep -E 'weblogic|oracle|postgres|mysql|nginx|apache|httpd|tomcat' | tee -a "$OUTPUT_FILE"
elif command -v service &> /dev/null; then
    service --status-all | grep -E 'weblogic|oracle|postgres|mysql|nginx|apache|httpd|tomcat' | tee -a "$OUTPUT_FILE"
else
    echo "서비스 상태를 확인할 수 없습니다." | tee -a "$OUTPUT_FILE"
fi
echo "" | tee -a "$OUTPUT_FILE"

# 타임존 설정
echo "## 시간 설정" | tee -a "$OUTPUT_FILE"
echo "* 타임존:" | tee -a "$OUTPUT_FILE"
timedatectl 2>/dev/null || (ls -l /etc/localtime && date) | tee -a "$OUTPUT_FILE"
echo "* NTP 설정:" | tee -a "$OUTPUT_FILE"
if [ -f /etc/ntp.conf ]; then
    grep "^server" /etc/ntp.conf | tee -a "$OUTPUT_FILE"
elif [ -f /etc/chrony.conf ]; then
    grep "^server" /etc/chrony.conf | tee -a "$OUTPUT_FILE"
else
    echo "NTP 설정 파일을 찾을 수 없습니다." | tee -a "$OUTPUT_FILE"
fi
echo "" | tee -a "$OUTPUT_FILE"

# 사용자 환경 변수
echo "## 시스템 환경 변수" | tee -a "$OUTPUT_FILE"
echo "* PATH:" | tee -a "$OUTPUT_FILE"
echo $PATH | tee -a "$OUTPUT_FILE"
echo "* LD_LIBRARY_PATH:" | tee -a "$OUTPUT_FILE"
echo $LD_LIBRARY_PATH | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

echo "====================================================" | tee -a "$OUTPUT_FILE"
echo "검사 완료!" | tee -a "$OUTPUT_FILE"
echo "결과 파일: $OUTPUT_FILE" | tee -a "$OUTPUT_FILE"
echo "====================================================" | tee -a "$OUTPUT_FILE"

# 권한 설정
chmod 644 "$OUTPUT_FILE"

echo "시스템 설정 검사가 완료되었습니다."
echo "결과 파일은 $OUTPUT_FILE 에 저장되었습니다."
echo "이 파일을 AS-IS(현재) 및 TO-BE(새 서버) 시스템에서 각각 실행한 후 비교하세요."
echo "diff를 사용하여 비교할 수 있습니다: diff AS-IS_파일.txt TO-BE_파일.txt"
