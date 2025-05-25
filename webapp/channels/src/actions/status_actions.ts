// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {UserProfile} from '@mattermost/types/users';

import {addUserIdsForStatusFetchingPoll} from 'mattermost-redux/actions/status_profile_polling';
import {getStatusesByIds} from 'mattermost-redux/actions/users';
import {getCurrentChannelId} from 'mattermost-redux/selectors/entities/channels';
import {getIsUserStatusesConfigEnabled} from 'mattermost-redux/selectors/entities/common';
import {getPostsInCurrentChannel} from 'mattermost-redux/selectors/entities/posts';
import {getDirectShowPreferences} from 'mattermost-redux/selectors/entities/preferences';
import {getCurrentUserId} from 'mattermost-redux/selectors/entities/users';

import {loadCustomEmojisForCustomStatusesByUserIds} from 'actions/emoji_actions';

import type {ActionFunc} from 'types/store';

/**
 * Adds the following users to the status pool for fetching their statuses:
 * - All users of current channel with recent posts.
 * - All users who have DMs open with the current user.
 * - The current user.
 */
export function addVisibleUsersInCurrentChannelAndSelfToStatusPoll(): ActionFunc<boolean> {
    return (dispatch, getState) => {
        const state = getState();
        const currentUserId = getCurrentUserId(state);
        const currentChannelId = getCurrentChannelId(state);
        const postsInChannel = getPostsInCurrentChannel(state);
        const numberOfPostsVisibleInCurrentChannel = state.views.channel.postVisibility[currentChannelId] || 0;

        const userIdsToFetchStatusFor = new Set<string>();

        // 현재 채널에서 최근에 게시한 사용자들의 상태 가져오기
        if (postsInChannel && numberOfPostsVisibleInCurrentChannel > 0) {
            // 더 많은 게시물을 포함하도록 개선 (100개까지)
            const posts = postsInChannel.slice(0, Math.max(numberOfPostsVisibleInCurrentChannel, 100));
            for (const post of posts) {
                if (post.user_id && post.user_id !== currentUserId) {
                    userIdsToFetchStatusFor.add(post.user_id);
                }
            }
        }

        // DM 채널의 사용자들 상태 가져오기
        const directShowPreferences = getDirectShowPreferences(state);
        for (const directShowPreference of directShowPreferences) {
            if (directShowPreference.value === 'true') {
                // DM의 상대방 사용자 ID
                userIdsToFetchStatusFor.add(directShowPreference.name);
            }
        }

        // 현재 사용자도 상태 가져오기 목록에 추가
        userIdsToFetchStatusFor.add(currentUserId);

        // 모든 사용자 ID를 배열로 변환
        const userIdsForStatus = Array.from(userIdsToFetchStatusFor);
        if (userIdsForStatus.length > 0) {
            // 상태 업데이트를 위해 폴링 시작
            dispatch(addUserIdsForStatusFetchingPoll(userIdsForStatus));
            
            // 즉시 상태 가져오기 (폴링과 별도로 즉시 업데이트)
            dispatch(getStatusesByIds(userIdsForStatus));
        }

        return {data: true};
    };
}

export function loadStatusesForProfilesList(users: UserProfile[] | null): ActionFunc<boolean> {
    return (dispatch) => {
        if (users == null) {
            return {data: false};
        }

        const statusesToLoad = [];
        for (let i = 0; i < users.length; i++) {
            statusesToLoad.push(users[i].id);
        }

        dispatch(loadStatusesByIds(statusesToLoad));

        return {data: true};
    };
}

export function loadStatusesForProfilesMap(users: Record<string, UserProfile> | UserProfile[] | null): ActionFunc {
    return (dispatch) => {
        if (users == null) {
            return {data: false};
        }

        const statusesToLoad = [];
        for (const userId in users) {
            if (Object.hasOwn(users, userId)) {
                statusesToLoad.push(userId);
            }
        }

        dispatch(loadStatusesByIds(statusesToLoad));

        return {data: true};
    };
}

export function loadStatusesByIds(userIds: string[]): ActionFunc {
    return (dispatch, getState) => {
        const state = getState();
        const enabledUserStatuses = getIsUserStatusesConfigEnabled(state);

        if (userIds.length === 0 || !enabledUserStatuses) {
            return {data: false};
        }

        dispatch(getStatusesByIds(userIds));
        dispatch(loadCustomEmojisForCustomStatusesByUserIds(userIds));
        return {data: true};
    };
}

export function loadProfilesMissingStatus(users: UserProfile[]): ActionFunc {
    return (dispatch, getState) => {
        const state = getState();
        const enabledUserStatuses = getIsUserStatusesConfigEnabled(state);

        const statuses = state.entities.users.statuses;

        const missingStatusByIds = users.
            filter((user) => !statuses[user.id]).
            map((user) => user.id);

        if (missingStatusByIds.length === 0 || !enabledUserStatuses) {
            return {data: false};
        }

        dispatch(getStatusesByIds(missingStatusByIds));
        dispatch(loadCustomEmojisForCustomStatusesByUserIds(missingStatusByIds));
        return {data: true};
    };
}
