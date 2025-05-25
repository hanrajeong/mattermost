// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import classNames from 'classnames';
import React, {memo, forwardRef} from 'react';
import type {HTMLAttributes, RefObject, SyntheticEvent} from 'react';
import {useIntl} from 'react-intl';

import {Client4} from 'mattermost-redux/client';

import BotDefaultIcon from 'images/bot_default_icon.png';
import DefaultProfileImage from './profile.jpg';

import './avatar.scss';

export type TAvatarSizeToken = 'xxs' | 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'xl-custom-GM' | 'xl-custom-DM' | 'xxl';

export const getAvatarWidth = (size: TAvatarSizeToken) => {
    switch (size) {
    case 'xxs':
        return 16;
    case 'xs':
        return 20;
    case 'sm':
        return 24;
    case 'md':
        return 32;
    case 'lg':
        return 36;
    case 'xl':
        return 50;
    case 'xl-custom-GM':
        return 72;
    case 'xl-custom-DM':
        return 96;
    case 'xxl':
        return 128;
    }
    return 0;
};

type Props = {
    url?: string;
    username?: string;
    size?: TAvatarSizeToken;
    text?: string;
};

type Attrs = HTMLAttributes<HTMLElement>;

const isURLForUser = (url: string) => url && url.startsWith(Client4.getUsersRoute());

// 사용자 정의 기본 프로필 이미지를 반환하는 함수
const getDefaultProfileImageUrl = () => {
    return DefaultProfileImage;
};

const Avatar = forwardRef<HTMLElement, Props & Attrs>(({
    url,
    username,
    size = 'md',
    text,
    ...attrs
}, ref) => {
    const {formatMessage} = useIntl();

    const classes = classNames(`Avatar Avatar-${size}`, attrs.className);

    if (text) {
        return (
            <div
                {...attrs}
                ref={ref as RefObject<HTMLDivElement>}
                className={classNames(classes, 'Avatar-plain')}
                data-content={text}
            />
        );
    }

    function handleOnError(e: SyntheticEvent<HTMLImageElement, Event>) {
        // 사용자 프로필 이미지인 경우 기본 프로필 이미지로 대체
        let fallbackSrc;
        
        if (url && isURLForUser(url)) {
            // 사용자 프로필 이미지인 경우 커스텀 프로필 이미지로 대체
            fallbackSrc = getDefaultProfileImageUrl();
        } else {
            // 봇이나 다른 유형의 이미지인 경우 기본 봇 아이콘 사용
            fallbackSrc = BotDefaultIcon;
        }

        if (e.currentTarget.src !== fallbackSrc) {
            e.currentTarget.src = fallbackSrc;
        }
    }

    // URL이 없거나 비어 있는 경우 커스텀 기본 이미지 사용
    const imageSrc = url || getDefaultProfileImageUrl();

    return (
        <img
            {...attrs}
            ref={ref as RefObject<HTMLImageElement>}
            className={classes}
            alt={formatMessage({id: 'avatar.alt', defaultMessage: '{username} profile image'}, {
                username: username || 'user',
            })}
            src={imageSrc}
            loading='lazy'
            onError={handleOnError}
        />
    );
});

Avatar.displayName = 'Avatar';

export default memo(Avatar);
