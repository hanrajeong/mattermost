// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {useSelector} from 'react-redux';

import type {Post} from '@mattermost/types/posts';
import {getUser} from 'mattermost-redux/selectors/entities/users';
import type {GlobalState} from '@mattermost/types/store';

import ProfilePicture from 'components/profile_picture';

type Props = {
    post: Post;
    onClick: (e: React.MouseEvent<HTMLDivElement>) => void;
}

const ReplyPreview: React.FC<Props> = ({post, onClick}) => {
    const user = useSelector((state: GlobalState) => getUser(state, post.user_id));
    
    // 메시지가 너무 길면 자르기
    const truncatedMessage = post.message.length > 100 ? 
        `${post.message.substring(0, 100)}...` : 
        post.message;
    
    return (
        <div className='ReplyPreview' onClick={onClick}>
            <div className='ReplyPreview__avatar'>
                <ProfilePicture
                    src={user?.profile_image_url}
                    userId={post.user_id}
                    size='xs'
                />
            </div>
            <div className='ReplyPreview__content'>
                <div className='ReplyPreview__username'>
                    {user?.username || 'Unknown User'}
                </div>
                <div className='ReplyPreview__message'>
                    {truncatedMessage}
                </div>
            </div>
        </div>
    );
};

export default ReplyPreview;
