#!/bin/bash

# 시스템 마이그레이션 종합 검증 스크립트
# 작성일: 2025-05-25 (표준 확장판)

HOSTNAME=$(hostname)
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
OUTPUT_FILE="system_config_${HOSTNAME}_${TIMESTAMP}.txt"

# 시작 헤더
cat <<-EOF | tee "$OUTPUT_FILE"
====================================================
시스템 마이그레이션 검증 보고서
서버 이름: $HOSTNAME
실행 일시: $(date)
====================================================

EOF

# 커널 및 부트로더
{
  echo "## 커널 & 부트로더"
  uname -a
  echo "* /proc/cmdline: $(cat /proc/cmdline)"
  echo "## /etc/default/grub"
  grep -vE '^#' /etc/default/grub 2>/dev/null || echo "grub 파일 없음"
  echo
} | tee -a "$OUTPUT_FILE"

# 하드웨어
{
  echo "## 하드웨어 정보"
  lscpu | grep -E 'Model name|CPU\(s\)|Thread|Socket|NUMA'
  free -h
  echo
} | tee -a "$OUTPUT_FILE"

# OS 및 패키지
{
  echo "## OS & 패키지 정보"
  cat /etc/os-release 2>/dev/null
  cat /etc/redhat-release 2>/dev/null
  echo "## 설치된 패키지 목록"
  rpm -qa | sort
  echo "## Yum 레포 목록"
  ls /etc/yum.repos.d/ | tee /dev/null && grep -H '^\[' /etc/yum.repos.d/*.repo
  echo
} | tee -a "$OUTPUT_FILE"

# 프로세스 & 리소스
{
  echo "## 프로세스 & 리소스 제한"
  echo "* pid_max: $(cat /proc/sys/kernel/pid_max)"
  echo "* max user processes: $(ulimit -u)"
  echo "* open files per proc: $(ulimit -n)"
  echo "* fs.file-max: $(cat /proc/sys/fs/file-max)"
  echo "* vm.swappiness: $(cat /proc/sys/vm/swappiness)"
  echo "* overcommit_memory: $(cat /proc/sys/vm/overcommit_memory)"
  echo "* overcommit_ratio: $(cat /proc/sys/vm/overcommit_ratio)"
  echo
} | tee -a "$OUTPUT_FILE"

# limits.conf
{
  echo "## /etc/security/limits.conf"
  grep -vE '^#|^$' /etc/security/limits.conf
  echo
} | tee -a "$OUTPUT_FILE"

# sysctl 파라미터
{
  echo "## sysctl 파라미터"
  sysctl -a 2>/dev/null | grep -E 'fs\.|net\.|kernel\.|vm\.' | sort
  echo
} | tee -a "$OUTPUT_FILE"

# I/O 스케줄러 & Transparent HugePages
{
  echo "## 디스크 I/O 스케줄러"
  for d in /sys/block/sd*; do
    echo "$(basename $d): $(cat $d/queue/scheduler)"
  done
  echo "## Transparent HugePages"
  cat /sys/kernel/mm/transparent_hugepage/enabled 2>/dev/null
  cat /sys/kernel/mm/transparent_hugepage/defrag 2>/dev/null
  echo
} | tee -a "$OUTPUT_FILE"

# 파일시스템 & 마운트
{
  echo "## 디스크 & 마운트"
  df -h
  lsblk
  mount | grep -E '^/dev'
  echo "## /etc/fstab"
  grep -vE '^#|^$' /etc/fstab
  echo
} | tee -a "$OUTPUT_FILE"

# 네트워크
{
  echo "## 네트워크 설정"
  ip addr
  ip route
  ss -tuln
  echo "## DNS & nsswitch"
  cat /etc/resolv.conf
  cat /etc/nsswitch.conf | grep hosts
  echo
} | tee -a "$OUTPUT_FILE"

# SSH & 보안
{
  echo "## SSH 설정"
  grep -vE '^#|^$' /etc/ssh/sshd_config
  echo "## 사용자 계정 & 그룹"
  awk -F: '$3>=1000{print $1" "$3" "$4}' /etc/passwd
  getent group
  echo "## sudoers"
  grep -H -vE '^#|^$' /etc/sudoers /etc/sudoers.d/* 2>/dev/null
  echo "## authorized_keys"
  for u in $(awk -F: '$3>=1000{print $1}' /etc/passwd); do
    echo "--- $u ---"
    [ -f /home/$u/.ssh/authorized_keys ] && cat /home/$u/.ssh/authorized_keys
  done
  echo
} | tee -a "$OUTPUT_FILE"

# 서비스 & 데몬
{
  echo "## 실행 중 서비스"
  systemctl list-units --type=service --state=running | grep -E 'weblogic|oracle|postgres|mysql|nginx|apache|httpd|tomcat|cron'
  echo "## cron jobs"
  for u in root $(awk -F: '$3>=1000{print $1}' /etc/passwd); do
    echo "--- $u cron ---"
    crontab -u $u -l 2>/dev/null
  done
  echo "## audit rules"
  auditctl -l 2>/dev/null
  echo
} | tee -a "$OUTPUT_FILE"

# Java & WebLogic
{
  echo "## Java 설정"
  command -v java &>/dev/null && java -version 2>&1
  echo "JAVA_HOME=$JAVA_HOME"
  echo "## WebLogic 설정"
  if [ -n "$WL_HOME" ] && [ -f "$WL_HOME/server/lib/weblogic.jar" ]; then
    java -cp "$WL_HOME/server/lib/weblogic.jar" weblogic.version | head -2
  fi
  echo
} | tee -a "$OUTPUT_FILE"

# 보안 & 로그
{
  echo "## SELinux & 방화벽"
  getenforce 2>/dev/null
  systemctl is-active firewalld 2>/dev/null
  iptables -L | head -n 10
  echo "## 로그로테이트"
  grep -H '^rotate' /etc/logrotate.conf /etc/logrotate.d/*
  echo
} | tee -a "$OUTPUT_FILE"

# 시간 & NTP
{
  echo "## 시간 설정"
  timedatectl
  echo "localtime link → $(readlink -f /etc/localtime)"
  echo "NTP server:"
  grep '^server' /etc/ntp.conf /etc/chrony.conf 2>/dev/null
  echo
} | tee -a "$OUTPUT_FILE"

# 커널 모듈 & 드라이버
{
  echo "## 커널 모듈"
  lsmod | head -n 20
  echo "## PCI 장치"
  lspci | head -n 20
  echo "## 네트워크 드라이버 & 설정"
  for iface in $(ls /sys/class/net); do
    ethtool -i $iface | tee -a "$OUTPUT_FILE"
  done
  echo
} | tee -a "$OUTPUT_FILE"

# 바이너리 무결성
{
  echo "## 주요 바이너리 해시"
  [ -x "$(command -v java)" ] && sha256sum "$(command -v java)"
  [ -f "$WL_HOME/server/lib/weblogic.jar" ] && sha256sum "$WL_HOME/server/lib/weblogic.jar"
  echo
} | tee -a "$OUTPUT_FILE"

# 패키지 버전 비교
{
  echo "## 주요 패키지 버전"
  rpm -qa | grep -E 'glibc|kernel|weblogic|jdk|openssl|postgres' | sort
  echo
} | tee -a "$OUTPUT_FILE"

# 마무리
echo "====================================================" | tee -a "$OUTPUT_FILE"
echo "검증 완료: $OUTPUT_FILE" | tee -a "$OUTPUT_FILE"
echo "AS-IS/TO-BE 각각 실행 후 diff로 비교하세요." | tee -a "$OUTPUT_FILE"

chmod 644 "$OUTPUT_FILE"
