#!/bin/bash

# 분할된 파일을 하나로 합치는 스크립트
echo "분할된 파일들을 합치는 중..."
cat mattermost_closed_network.tar.part* > mattermost_closed_network.tar

# 압축 풀기
echo "파일 압축 해제 중..."
tar -xf mattermost_closed_network.tar

# 원본 파일 삭제 (선택사항)
read -p "합쳐진 tar 파일을 삭제하시겠습니까? (y/n): " answer
if [ "$answer" = "y" ]; then
    rm mattermost_closed_network.tar
    echo "tar 파일을 삭제했습니다."
fi

echo -e "\n복원이 완료되었습니다. 이제 mattermost_package 디렉토리에서 install.sh를 실행하세요:"
echo "cd mattermost_package"
echo "./install.sh"
