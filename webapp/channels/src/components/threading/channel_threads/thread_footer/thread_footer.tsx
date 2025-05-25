// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {memo, useCallback, useEffect, useMemo, useState} from 'react';
import {FormattedMessage} from 'react-intl';
import {useDispatch, useSelector} from 'react-redux';

import type {Post, PostList} from '@mattermost/types/posts';
import type {UserThread} from '@mattermost/types/threads';
import {threadIsSynthetic} from '@mattermost/types/threads';

import {setThreadFollow, getThread as fetchThread} from 'mattermost-redux/actions/threads';
import {getPostThread} from 'mattermost-redux/actions/posts';
import {Posts} from 'mattermost-redux/constants';
import {getPost, getPostsInThread} from 'mattermost-redux/selectors/entities/posts';
import {getCurrentTeamId} from 'mattermost-redux/selectors/entities/teams';
import {makeGetThreadOrSynthetic} from 'mattermost-redux/selectors/entities/threads';
import {getCurrentUserId, getUser} from 'mattermost-redux/selectors/entities/users';

import {trackEvent} from 'actions/telemetry_actions';
import {selectPost} from 'actions/views/rhs';

import Button from 'components/threading/common/button';
import FollowButton from 'components/threading/common/follow_button';
import {THREADING_TIME} from 'components/threading/common/options';
import Timestamp from 'components/timestamp';
import Avatars from 'components/widgets/users/avatars';
import WithTooltip from 'components/with_tooltip';
import ProfilePicture from 'components/profile_picture';

import type {GlobalState} from 'types/store';

import ReplyPreview from './reply_preview';

import './thread_footer.scss';

type Props = {
    threadId: UserThread['id'];
    replyClick?: React.EventHandler<React.MouseEvent>;
};

function ThreadFooter({
    threadId,
    replyClick,
}: Props) {
    const dispatch = useDispatch();
    const currentTeamId = useSelector(getCurrentTeamId);
    const currentUserId = useSelector(getCurrentUserId);
    const post = useSelector((state: GlobalState) => getPost(state, threadId));
    const getThreadOrSynthetic = useMemo(makeGetThreadOrSynthetic, [post.id]);
    const thread = useSelector((state: GlobalState) => getThreadOrSynthetic(state, post));
    
    // 댓글 목록 상태 관리
    const [replyPosts, setReplyPosts] = useState<Post[]>([]);
    // 기본적으로 댓글을 표시하도록 설정
    const [showReplies, setShowReplies] = useState(true);
    
    // 스레드의 댓글 가져오기
    useEffect(() => {
        if (thread.reply_count > 0) {
            dispatch(getPostThread(threadId, true)).then((result) => {
                if (result.data) {
                    const posts = Object.values(result.data.posts)
                        .filter((p: Post) => p.id !== threadId) // 원본 포스트 제외
                        .sort((a: Post, b: Post) => b.create_at - a.create_at); // 최신순 정렬(내림차순)
                    setReplyPosts(posts);
                }
            });
        }
    }, [threadId, thread.reply_count]);
    
    useEffect(() => {
        if (threadIsSynthetic(thread) && thread.is_following && thread.reply_count > 0) {
            dispatch(fetchThread(currentUserId, currentTeamId, threadId));
        }
    }, []);

    const {
        participants,
        reply_count: totalReplies = 0,
        last_reply_at: lastReplyAt,
        is_following: isFollowing = false,
        post: {
            channel_id: channelId,
        },
    } = thread;

    const participantIds = useMemo(() => (participants || []).map(({id}) => id).reverse(), [participants]);

    const handleReply = useCallback((e) => {
        if (replyClick) {
            replyClick(e);
            return;
        }

        trackEvent('crt', 'replied_using_footer');
        e.stopPropagation();
        dispatch(selectPost({id: threadId, channel_id: channelId} as Post));
    }, [replyClick, threadId, channelId]);
    
    // 댓글 표시 토글
    const toggleReplies = useCallback((e) => {
        e.stopPropagation();
        setShowReplies(!showReplies);
    }, [showReplies]);

    const handleFollowing = useCallback((e) => {
        e.stopPropagation();
        dispatch(setThreadFollow(currentUserId, currentTeamId, threadId, !isFollowing));
    }, [isFollowing]);

    if (post.delete_at > 0 || post.state === Posts.POST_DELETED) {
        return null;
    }

    return (
        <div className='ThreadFooter'>
            {!isFollowing || threadIsSynthetic(thread) || !thread.unread_replies ? (
                <div className='indicator'/>
            ) : (
                <WithTooltip
                    title={
                        <FormattedMessage
                            id='threading.numNewMessages'
                            defaultMessage='{newReplies, plural, =0 {no unread messages} =1 {one unread message} other {# unread messages}}'
                            values={{newReplies: thread.unread_replies}}
                        />
                    }
                >
                    <div
                        className='indicator'
                        tabIndex={0}
                    >
                        <div className='dot-unreads'/>
                    </div>
                </WithTooltip>
            )}

            {participantIds && participantIds.length > 0 ? (
                <Avatars
                    userIds={participantIds}
                    size='sm'
                />
            ) : null}

            {thread.reply_count > 0 && (
                <div className='ReplySection'>
                    <Button
                        onClick={toggleReplies}
                        className='ReplyButton separated'
                        prepend={
                            <span className='icon'>
                                <i className='icon-reply-outline'/>
                            </span>
                        }
                    >
                        <FormattedMessage
                            id='threading.numReplies'
                            defaultMessage='{totalReplies, plural, =0 {Reply} =1 {# reply} other {# replies}}'
                            values={{totalReplies}}
                        />
                    </Button>
                    
                    {showReplies && replyPosts.length > 0 && (
                        <div className='ReplyPreviewContainer'>
                            <div className='LatestRepliesHeader'>
                                <FormattedMessage
                                    id='threading.latestReplies'
                                    defaultMessage='최신 댓글'
                                />
                            </div>
                            {replyPosts.slice(0, 3).map((replyPost) => (
                                <ReplyPreview 
                                    key={replyPost.id}
                                    post={replyPost}
                                    onClick={handleReply}
                                />
                            ))}
                            {replyPosts.length > 3 && (
                                <div className='MoreReplies' onClick={handleReply}>
                                    <FormattedMessage
                                        id='threading.moreReplies'
                                        defaultMessage='+ {count} more'
                                        values={{count: replyPosts.length - 3}}
                                    />
                                </div>
                            )}
                        </div>
                    )}
                </div>
            )}

            <FollowButton
                isFollowing={isFollowing}
                className='separated'
                onClick={handleFollowing}
            />

            {Boolean(lastReplyAt) && (
                <Timestamp
                    value={lastReplyAt}
                    {...THREADING_TIME}
                >
                    {({formatted}) => (
                        <span className='Timestamp separated alt-visible'>
                            <FormattedMessage
                                id='threading.footer.lastReplyAt'
                                defaultMessage='Last reply {formatted}'
                                values={{formatted}}
                            />
                        </span>
                    )}
                </Timestamp>
            )}
        </div>
    );
}

export default memo(ThreadFooter);
