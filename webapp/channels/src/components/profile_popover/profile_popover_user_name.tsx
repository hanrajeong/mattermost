// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {useSelector} from 'react-redux';

import type {GlobalState} from '@mattermost/types/store';
import type {UserProfile} from '@mattermost/types/users';

import {getUser} from 'mattermost-redux/selectors/entities/users';

type Props = {
    hasFullName: boolean;
    username: string;
}

const ProfilePopoverUserName = ({
    hasFullName,
    username,
}: Props) => {
    // 사용자 정보를 가져와서 실제 이름이 있는지 확인
    const user = useSelector((state: GlobalState) => getUser(state, username));
    
    // 실제 이름이 있으면 표시, 없으면 사번 표시
    const displayName = user && user.first_name && user.last_name ? 
        `${user.first_name} ${user.last_name}` : 
        `@${username}`;
    
    return (
        <p
            id='userPopoverUsername'
            className={
                hasFullName ? 'user-profile-popover__non-heading' : 'user-profile-popover__heading'
            }
            title={username}
        >
            {displayName}
        </p>
    );
};

export default ProfilePopoverUserName;
