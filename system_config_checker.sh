#!/bin/bash

# 시스템 마이그레이션 종합 검증 스크립트
# 작성일: 2025-05-25 (표준 확장판 + 동적 탐색 + 권한/경로 검증 기능 추가)

HOSTNAME=$(hostname)
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
OUTPUT_FILE="system_config_${HOSTNAME}_${TIMESTAMP}.txt"

# 사전 안내
cat <<-EOF | tee "$OUTPUT_FILE"
====================================================
시스템 마이그레이션 검증 보고서
서버 이름: $HOSTNAME
실행 일시: $(date)
스크립트 실행 경로: $(pwd)
====================================================

EOF

# 권한 체크
if [ "$EUID" -ne 0 ]; then
  echo "[⚠️] root 권한이 아닙니다. 일부 정보 수집에 실패할 수 있습니다." | tee -a "$OUTPUT_FILE"
  echo | tee -a "$OUTPUT_FILE"
fi

# 커널 & 부트로더
{
  echo "## 커널 & 부트로더"
  uname -a 2>&1
  echo "* /proc/cmdline: $(cat /proc/cmdline 2>&1)"
  echo "## /etc/default/grub"
  grep -vE '^#' /etc/default/grub 2>&1 || echo "grub 파일 없음"
  echo
} | tee -a "$OUTPUT_FILE"

# 하드웨어 정보
{
  echo "## 하드웨어 정보"
  lscpu 2>&1 | grep -E 'Model name|CPU\(s\)|Thread|Socket|NUMA'
  free -h 2>&1
  echo
} | tee -a "$OUTPUT_FILE"

# OS & 패키지 정보
{
  echo "## OS & 패키지 정보"
  cat /etc/os-release 2>&1
  cat /etc/redhat-release 2>&1
  echo "## 설치된 패키지 목록"
  rpm -qa 2>&1 | sort
  echo "## Yum 레포 목록"
  ls /etc/yum.repos.d/ 2>&1 && grep -H '^\[' /etc/yum.repos.d/*.repo 2>&1
  echo
} | tee -a "$OUTPUT_FILE"

# 프로세스 & 자원 제한
{
  echo "## 프로세스 & 자원 제한"
  echo "* pid_max: $(cat /proc/sys/kernel/pid_max 2>&1)"
  echo "* max user procs: $(ulimit -u 2>&1)"
  echo "* open files per proc: $(ulimit -n 2>&1)"
  echo "* fs.file-max: $(cat /proc/sys/fs/file-max 2>&1)"
  echo "* vm.swappiness: $(cat /proc/sys/vm/swappiness 2>&1)"
  echo "* overcommit_memory: $(cat /proc/sys/vm/overcommit_memory 2>&1)"
  echo "* overcommit_ratio: $(cat /proc/sys/vm/overcommit_ratio 2>&1)"
  echo
} | tee -a "$OUTPUT_FILE"

# /etc/security/limits.conf
{
  echo "## /etc/security/limits.conf"
  grep -vE '^#|^$' /etc/security/limits.conf 2>&1
  echo
} | tee -a "$OUTPUT_FILE"

# sysctl 파라미터
{
  echo "## sysctl 파라미터"
  sysctl -a 2>&1 | grep -E 'fs\.|net\.|kernel\.|vm\.' | sort
  echo
} | tee -a "$OUTPUT_FILE"

# I/O 스케줄러 & THP
{
  echo "## 디스크 I/O 스케줄러"
  for d in /sys/block/sd*; do
    echo "$(basename $d): $(cat $d/queue/scheduler 2>&1)"
  done
  echo "## Transparent HugePages"
  cat /sys/kernel/mm/transparent_hugepage/enabled 2>&1
  cat /sys/kernel/mm/transparent_hugepage/defrag 2>&1
  echo
} | tee -a "$OUTPUT_FILE"

# 디스크 & 마운트
{
  echo "## 디스크 & 마운트"
  df -h 2>&1
  lsblk 2>&1
  mount | grep -E '^/dev' 2>&1
  echo "## /etc/fstab"
  grep -vE '^#|^$' /etc/fstab 2>&1
  echo
} | tee -a "$OUTPUT_FILE"

# 네트워크 설정
{
  echo "## 네트워크 설정"
  ip addr 2>&1
  ip route 2>&1
  ss -tuln 2>&1
  echo "## DNS & NSS"
  cat /etc/resolv.conf 2>&1
  grep hosts /etc/nsswitch.conf 2>&1
  echo
} | tee -a "$OUTPUT_FILE"

# /fsutil 솔루션 정보
{
  echo "## /fsutil 설치 솔루션 정보"
  if [ -d "/fsutil" ]; then
    echo "### 솔루션 목록"; ls -1 /fsutil 2>&1
    echo "### 설정 파일 요약";
    find /fsutil -maxdepth 2 -type f \( -name '*.conf' -o -name '*.xml' -o -name '*.properties' \) 2>/dev/null | while read cfg; do
      echo "--- $cfg ---";
      head -n 50 "$cfg" 2>&1;
    done
  else
    echo "/fsutil 디렉토리가 없습니다."
  fi
  echo
} | tee -a "$OUTPUT_FILE"

# 사용자 및 그룹
{
  echo "## 사용자 및 그룹 정보"
  echo "--- /etc/passwd ---"; grep -vE '^#' /etc/passwd 2>&1
  echo "--- /etc/group ---"; grep -vE '^#' /etc/group 2>&1
  echo
} | tee -a "$OUTPUT_FILE"

# 호스트 파일 & 접근
{
  echo "## 호스트 & 접근 제어"
  cat /etc/hosts 2>&1
  cat /etc/hosts.allow 2>/dev/null
  cat /etc/hosts.deny 2>/dev/null
  echo
} | tee -a "$OUTPUT_FILE"

# SSH & SELinux
{
  echo "## SSH & SELinux"
  grep -vE '^#|^$' /etc/ssh/sshd_config 2>&1
  getsebool -a 2>/dev/null
  echo
} | tee -a "$OUTPUT_FILE"

# 서비스 & cron & audit
{
  echo "## 서비스 & cron & audit"
  systemctl list-units --type=service --state=running 2>&1 | grep -E 'weblogic|cron|nginx|apache|tomcat|postgres|mysql'
  for u in root $(awk -F: '$3>=1000{print $1}' /etc/passwd); do
    echo "--- $u cron ---"; crontab -u $u -l 2>&1
  done
  auditctl -l 2>/dev/null
  echo
} | tee -a "$OUTPUT_FILE"

# Java 동적 경로
{
  echo "## Java 설정"
  JAVA_CMD=$(command -v java 2>/dev/null)
  if [ -n "$JAVA_CMD" ]; then
    echo "java: $JAVA_CMD";
    java -version 2>&1;
    echo "추정 JAVA_HOME: $(dirname $(dirname $JAVA_CMD))";
  else
    echo "Java 미설치";
  fi
  echo
} | tee -a "$OUTPUT_FILE"

# WebLogic 동적 탐색
{
  echo "## WebLogic 설정"
  WL_HOME_PATH=${WL_HOME:-}
  [ -z "$WL_HOME_PATH" ] && WL_HOME_PATH=$(find / -type d -name 'wlserver*' -path '*/oracle*' -prune 2>/dev/null | head -n1)
  if [ -n "$WL_HOME_PATH" ]; then
    echo "WL_HOME: $WL_HOME_PATH";
    java -cp "$WL_HOME_PATH/server/lib/weblogic.jar" weblogic.version 2>&1 | head -2;
    echo "### WL libraries";
    find "$WL_HOME_PATH/server/lib" -maxdepth 1 -type f -name '*.jar' | sort | tee -a "$OUTPUT_FILE";
  else
    echo "WebLogic 미탐색";
  fi
  echo
} | tee -a "$OUTPUT_FILE"

# 방화벽 & 로그로테이트
{
  echo "## 보안 & 로테이트"
  systemctl is-active firewalld 2>/dev/null
  iptables -L 2>&1 | head -n10
  grep -H '^rotate' /etc/logrotate.conf /etc/logrotate.d/* 2>/dev/null
  echo
} | tee -a "$OUTPUT_FILE"

# 시간 & NTP
{
  echo "## 시간 & NTP"
  timedatectl 2>&1
  echo "localtime: $(readlink -f /etc/localtime)";
  grep '^server' /etc/ntp.conf /etc/chrony.conf 2>/dev/null
  echo
} | tee -a "$OUTPUT_FILE"

# 커널 모듈 & 드라이버
{
  echo "## 모듈 & 네트워크 드라이버"
  lsmod 2>&1 | head -n20
  lspci 2>&1 | head -n20
  for iface in $(ls /sys/class/net); do
    echo "$iface:"; ethtool -i $iface 2>&1;
  done
  echo
} | tee -a "$OUTPUT_FILE"

# 해시 & 패키지 버전
{
  echo "## 무결성 & 패키지 버전"
  [ -n "$JAVA_CMD" ] && sha256sum "$JAVA_CMD" 2>&1
  [ -f "$WL_HOME_PATH/server/lib/weblogic.jar" ] && sha256sum "$WL_HOME_PATH/server/lib/weblogic.jar" 2>&1
  rpm -qa | grep -E 'glibc|kernel|weblogic|jdk|openssl|postgres' | sort 2>&1
  echo
} | tee -a "$OUTPUT_FILE"

# 완료 메시지
echo "====================================================" | tee -a "$OUTPUT_FILE"
echo "검증 완료: $OUTPUT_FILE" | tee -a "$OUTPUT_FILE"
echo "AS-IS/TO-BE 실행 후 diff 비교" | tee -a "$OUTPUT_FILE"

chmod 644 "$OUTPUT_FILE"
