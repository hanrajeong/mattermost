// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {ServerError} from '@mattermost/types/errors';
import type {ServerLimits} from '@mattermost/types/limits';

import {LimitsTypes} from 'mattermost-redux/action_types';
import {logError} from 'mattermost-redux/actions/errors';
import {forceLogoutIfNecessary} from 'mattermost-redux/actions/helpers';
import {Client4} from 'mattermost-redux/client';
import {getCurrentUserRoles} from 'mattermost-redux/selectors/entities/users';
import type {ActionFuncAsync} from 'mattermost-redux/types/actions';
import {isAdmin} from 'mattermost-redux/utils/user_utils';

export function getServerLimits(): ActionFuncAsync<ServerLimits> {
    return async (dispatch, getState) => {
        const roles = getCurrentUserRoles(getState());
        const amIAdmin = isAdmin(roles);
        
        // 관리자가 아닌 경우 기본 값을 반환
        if (!amIAdmin) {
            return {
                data: {
                    activeUserCount: 0,
                    maxUsersLimit: 0,
                },
            };
        }
        
        // 엔터프라이즈 버전인지 확인 (서버 정보에서 확인 가능)
        const state = getState();
        const isEnterpriseReady = state?.entities?.general?.license?.IsLicensed === 'true';
        
        // 엔터프라이즈 버전이 아닌 경우 API 호출 안함
        if (!isEnterpriseReady) {
            return {
                data: {
                    activeUserCount: 0,
                    maxUsersLimit: 0,
                },
            };
        }

        let response;
        try {
            response = await Client4.getServerLimits();
        } catch (err) {
            forceLogoutIfNecessary(err, dispatch, getState);
            // 404 오류는 로그에 기록하지 않음 (엔터프라이즈 버전이 아닌 경우 발생하는 오류)
            if (!(err as ServerError).status_code || (err as ServerError).status_code !== 404) {
                dispatch(logError(err));
            }
            return {
                data: {
                    activeUserCount: 0,
                    maxUsersLimit: 0,
                },
            };
        }

        const data: ServerLimits = {
            activeUserCount: response?.data?.activeUserCount ?? 0,
            maxUsersLimit: response?.data?.maxUsersLimit ?? 0,
        };

        dispatch({type: LimitsTypes.RECIEVED_APP_LIMITS, data});

        return {data};
    };
}

