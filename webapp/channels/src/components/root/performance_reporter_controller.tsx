// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {useEffect, useRef} from 'react';
import {useStore} from 'react-redux';

import {Client4} from 'mattermost-redux/client';

import DesktopAppAPI from 'utils/desktop_api';
import PerformanceReporter from 'utils/performance_telemetry/reporter';

export default function PerformanceReporterController() {
    const store = useStore();

    const reporter = useRef<PerformanceReporter>();

    useEffect(() => {
        // 404 오류를 방지하기 위해 성능 보고서 전송을 비활성화합니다.
        // 이 기능이 필요한 경우 아래 코드를 주석 해제하세요.
        /*
        reporter.current = new PerformanceReporter(Client4, store, DesktopAppAPI);
        reporter.current.observe();
        */

        // There's no way to clean up web-vitals, so continue to assume that this component won't ever be unmounted
        return () => {
            // eslint-disable-next-line no-console
            console.error('PerformanceReporterController - Component unmounted or store changed');
        };
    }, [store]);

    return null;
}
