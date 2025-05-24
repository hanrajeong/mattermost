// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import classNames from 'classnames';
import throttle from 'lodash/throttle';
import React, {useState, useEffect, useRef, useCallback} from 'react';
import type {FocusEvent} from 'react';
import {FormattedMessage, useIntl} from 'react-intl';
import {useSelector, useDispatch} from 'react-redux';
import {useLocation, useHistory, Route} from 'react-router-dom';

import type {ServerError} from '@mattermost/types/errors';
import type {UserProfile} from '@mattermost/types/users';

import {getTeamInviteInfo} from 'mattermost-redux/actions/teams';
import {createUser, loadMe} from 'mattermost-redux/actions/users';
import {Client4} from 'mattermost-redux/client';
import {getConfig, getLicense, getPasswordConfig} from 'mattermost-redux/selectors/entities/general';
import {getIsOnboardingFlowEnabled} from 'mattermost-redux/selectors/entities/preferences';
import {getCurrentUserId} from 'mattermost-redux/selectors/entities/users';
import {isEmail, isEmployeeId} from 'mattermost-redux/utils/helpers';

import {redirectUserToDefaultTeam} from 'actions/global_actions';
import {removeGlobalItem, setGlobalItem} from 'actions/storage';
import {addUserToTeamFromInvite} from 'actions/team_actions';
import {trackEvent} from 'actions/telemetry_actions.jsx';
import {loginById} from 'actions/views/login';
import {getGlobalItem} from 'selectors/storage';

import AlertBanner from 'components/alert_banner';
import type {ModeType, AlertBannerProps} from 'components/alert_banner';
import useCWSAvailabilityCheck, {CSWAvailabilityCheckTypes} from 'components/common/hooks/useCWSAvailabilityCheck';
import DesktopAuthToken from 'components/desktop_auth_token';
import ExternalLink from 'components/external_link';
import ExternalLoginButton from 'components/external_login_button/external_login_button';
import type {ExternalLoginButtonType} from 'components/external_login_button/external_login_button';
import AlternateLinkLayout from 'components/header_footer_route/content_layouts/alternate_link';
import ColumnLayout from 'components/header_footer_route/content_layouts/column';
import type {CustomizeHeaderType} from 'components/header_footer_route/header_footer_route';
import LoadingScreen from 'components/loading_screen';
import Markdown from 'components/markdown';
import SaveButton from 'components/save_button';
import EntraIdIcon from 'components/widgets/icons/entra_id_icon';
import LockIcon from 'components/widgets/icons/lock_icon';
import LoginGitlabIcon from 'components/widgets/icons/login_gitlab_icon';
import LoginGoogleIcon from 'components/widgets/icons/login_google_icon';
import LoginOpenIDIcon from 'components/widgets/icons/login_openid_icon';
import CheckInput from 'components/widgets/inputs/check';
import Input, {SIZE} from 'components/widgets/inputs/input/input';
import type {CustomMessageInputType} from 'components/widgets/inputs/input/input';
import PasswordInput from 'components/widgets/inputs/password_input/password_input';

import {Constants, HostedCustomerLinks, ItemStatus, ValidationErrors} from 'utils/constants';
import {isValidPassword} from 'utils/password';
import {isDesktopApp} from 'utils/user_agent';
import {isValidUsername, getRoleFromTrackFlow, getMediumFromTrackFlow} from 'utils/utils';

import type {GlobalState} from 'types/store';

import './signup.scss';

const MOBILE_SCREEN_WIDTH = 1200;

type SignupProps = {
    onCustomizeHeader?: CustomizeHeaderType;
}

const Signup = ({onCustomizeHeader}: SignupProps) => {
    const intl = useIntl();
    const {formatMessage} = intl;
    const dispatch = useDispatch();
    const history = useHistory();
    const {search} = useLocation();

    const params = new URLSearchParams(search);
    const token = params.get('t') ?? '';
    const inviteId = params.get('id') ?? '';
    const data = params.get('d');
    const parsedData: Record<string, string> = data ? JSON.parse(data) : {};
    const {email: parsedEmail, name: parsedTeamName, reminder_interval: reminderInterval} = parsedData;

    const config = useSelector(getConfig);
    const {
        EnableOpenServer,
        EnableUserCreation,
        NoAccounts,
        EnableSignUpWithEmail,
        EnableSignUpWithGitLab,
        EnableSignUpWithGoogle,
        EnableSignUpWithOffice365,
        EnableSignUpWithOpenId,
        EnableLdap,
        EnableSaml,
        SamlLoginButtonText,
        LdapLoginFieldName,
        SiteName,
        CustomDescriptionText,
        GitLabButtonText,
        GitLabButtonColor,
        OpenIdButtonText,
        OpenIdButtonColor,
        EnableCustomBrand,
        CustomBrandText,
        TermsOfServiceLink,
        PrivacyPolicyLink,
    } = config;
    const {IsLicensed} = useSelector(getLicense);
    const loggedIn = Boolean(useSelector(getCurrentUserId));
    const onboardingFlowEnabled = useSelector(getIsOnboardingFlowEnabled);
    const usedBefore = useSelector((state: GlobalState) => (!inviteId && !loggedIn && token ? getGlobalItem(state, token, null) : undefined));

    const emailInput = useRef<HTMLInputElement>(null);
    const nameInput = useRef<HTMLInputElement>(null);
    const passwordInput = useRef<HTMLInputElement>(null);

    const isLicensed = IsLicensed === 'true';
    const enableOpenServer = EnableOpenServer === 'true';
    const enableUserCreation = EnableUserCreation === 'true';
    const noAccounts = NoAccounts === 'true';
    const enableSignUpWithEmail = enableUserCreation && EnableSignUpWithEmail === 'true';
    const enableSignUpWithGitLab = enableUserCreation && EnableSignUpWithGitLab === 'true';
    const enableSignUpWithGoogle = enableUserCreation && EnableSignUpWithGoogle === 'true';
    const enableSignUpWithOffice365 = enableUserCreation && EnableSignUpWithOffice365 === 'true';
    const enableSignUpWithOpenId = enableUserCreation && EnableSignUpWithOpenId === 'true';
    const enableLDAP = EnableLdap === 'true';
    const enableSAML = EnableSaml === 'true';
    const enableCustomBrand = EnableCustomBrand === 'true';

    const noOpenServer = !inviteId && !token && !enableOpenServer && !noAccounts && !enableUserCreation;

    const [email, setEmail] = useState(parsedEmail ?? '');
    const [name, setName] = useState('');
    const [password, setPassword] = useState('');
    const [loading, setLoading] = useState(Boolean(inviteId));
    const [isWaiting, setIsWaiting] = useState(false);
    const [emailError, setEmailError] = useState('');
    const [nameError, setNameError] = useState('');
    const [passwordError, setPasswordError] = useState('');
    const [brandImageError, setBrandImageError] = useState(false);
    const [serverError, setServerError] = useState('');
    const [teamName, setTeamName] = useState(parsedTeamName ?? '');
    const [alertBanner, setAlertBanner] = useState<AlertBannerProps | null>(null);
    const [isMobileView, setIsMobileView] = useState(false);
    const [submitClicked, setSubmitClicked] = useState(false);

    const cwsAvailability = useCWSAvailabilityCheck();

    const enableExternalSignup = enableSignUpWithGitLab || enableSignUpWithOffice365 || enableSignUpWithGoogle || enableSignUpWithOpenId || enableLDAP || enableSAML;
    const hasError = Boolean(emailError || nameError || passwordError || serverError || alertBanner);
    const canSubmit = Boolean(email && name && password) && !hasError && !loading;
    const passwordConfig = useSelector(getPasswordConfig);
    const {error: passwordInfo} = isValidPassword('', passwordConfig, intl);

    const [desktopLoginLink, setDesktopLoginLink] = useState('');

    const getExternalSignupOptions = () => {
        const externalLoginOptions: ExternalLoginButtonType[] = [];

        if (!enableExternalSignup) {
            return externalLoginOptions;
        }

        if (enableSignUpWithGitLab) {
            const url = `${Client4.getOAuthRoute()}/gitlab/signup${search}`;
            externalLoginOptions.push({
                id: 'gitlab',
                url,
                icon: <LoginGitlabIcon/>,
                label: GitLabButtonText || formatMessage({id: 'login.gitlab', defaultMessage: 'GitLab'}),
                style: {color: GitLabButtonColor, borderColor: GitLabButtonColor},
                onClick: desktopExternalAuth(url),
            });
        }

        if (isLicensed && enableSignUpWithGoogle) {
            const url = `${Client4.getOAuthRoute()}/google/signup${search}`;
            externalLoginOptions.push({
                id: 'google',
                url,
                icon: <LoginGoogleIcon/>,
                label: formatMessage({id: 'login.google', defaultMessage: 'Google'}),
                onClick: desktopExternalAuth(url),
            });
        }

        if (isLicensed && enableSignUpWithOffice365) {
            const url = `${Client4.getOAuthRoute()}/office365/signup${search}`;
            externalLoginOptions.push({
                id: 'office365',
                url,
                icon: <EntraIdIcon/>,
                label: formatMessage({id: 'login.office365', defaultMessage: 'Entra ID'}),
                onClick: desktopExternalAuth(url),
            });
        }

        if (isLicensed && enableSignUpWithOpenId) {
            const url = `${Client4.getOAuthRoute()}/openid/signup${search}`;
            externalLoginOptions.push({
                id: 'openid',
                url,
                icon: <LoginOpenIDIcon/>,
                label: OpenIdButtonText || formatMessage({id: 'login.openid', defaultMessage: 'Open ID'}),
                style: {color: OpenIdButtonColor, borderColor: OpenIdButtonColor},
                onClick: desktopExternalAuth(url),
            });
        }

        if (isLicensed && enableLDAP) {
            const newSearchParam = new URLSearchParams(search);
            newSearchParam.set('extra', Constants.CREATE_LDAP);

            externalLoginOptions.push({
                id: 'ldap',
                url: `${Client4.getUrl()}/login?${newSearchParam.toString()}`,
                icon: <LockIcon/>,
                label: LdapLoginFieldName || formatMessage({id: 'signup.ldap', defaultMessage: 'AD/LDAP Credentials'}),
                onClick: () => {},
            });
        }

        if (isLicensed && enableSAML) {
            const newSearchParam = new URLSearchParams(search);
            newSearchParam.set('action', 'signup');

            const url = `${Client4.getUrl()}/login/sso/saml?${newSearchParam.toString()}`;
            externalLoginOptions.push({
                id: 'saml',
                url,
                icon: <LockIcon/>,
                label: SamlLoginButtonText || formatMessage({id: 'login.saml', defaultMessage: 'SAML'}),
                onClick: desktopExternalAuth(url),
            });
        }

        return externalLoginOptions;
    };

    const handleHeaderBackButtonOnClick = useCallback(() => {
        if (!noAccounts) {
            trackEvent('signup_email', 'click_back');
        }

        history.goBack();
    }, [noAccounts, history]);

    const handleInvalidInvite = ({
        // eslint-disable-next-line @typescript-eslint/naming-convention
        server_error_id,
        message,
    }: {server_error_id: string; message: string}) => {
        let errorMessage;

        if (server_error_id === 'store.sql_user.save.max_accounts.app_error' ||
            server_error_id === 'api.team.add_user_to_team_from_invite.guest.app_error') {
            errorMessage = message;
        }

        setServerError(errorMessage || formatMessage({id: 'signup_user_completed.invalid_invite.title', defaultMessage: 'This invite link is invalid'}));
        setLoading(false);
    };

    const handleAddUserToTeamFromInvite = async (token: string, inviteId: string) => {
        const {data: team, error} = await dispatch(addUserToTeamFromInvite(token, inviteId));

        if (team) {
            history.push('/' + team.name + `/channels/${Constants.DEFAULT_CHANNEL}`);
        } else if (error) {
            handleInvalidInvite(error);
        }
    };

    const getInviteInfo = async (inviteId: string) => {
        const {data, error} = await dispatch(getTeamInviteInfo(inviteId));

        if (data) {
            setServerError('');
            setTeamName(data.name);
        } else if (error) {
            handleInvalidInvite(error);
        }

        setLoading(false);
    };

    const getAlternateLink = useCallback(() => (
        <AlternateLinkLayout
            className='signup-body-alternate-link'
            alternateMessage={formatMessage({
                id: 'signup_user_completed.haveAccount',
                defaultMessage: 'Already have an account?',
            })}
            alternateLinkPath='/login'
            alternateLinkLabel={formatMessage({
                id: 'signup_user_completed.signIn',
                defaultMessage: 'Log in',
            })}
        />
    ), []);

    const onWindowResize = throttle(() => {
        setIsMobileView(window.innerWidth < MOBILE_SCREEN_WIDTH);
    }, 100);

    const desktopExternalAuth = (href: string) => {
        return (event: React.MouseEvent) => {
            if (isDesktopApp()) {
                event.preventDefault();

                setDesktopLoginLink(href);
                history.push(`/signup_user_complete/desktop${search}`);
            }
        };
    };

    useEffect(() => {
        dispatch(removeGlobalItem('team'));
        trackEvent('signup', 'signup_user_01_welcome', {...getRoleFromTrackFlow(), ...getMediumFromTrackFlow()});

        onWindowResize();

        window.addEventListener('resize', onWindowResize);

        if (search) {
            if ((inviteId || token) && loggedIn) {
                handleAddUserToTeamFromInvite(token, inviteId);
            } else if (inviteId) {
                getInviteInfo(inviteId);
            } else if (loggedIn) {
                if (onboardingFlowEnabled) {
                    // need info about whether admin or not,
                    // and whether admin has already completed
                    // first tiem onboarding. Instead of fetching and orchestrating that here,
                    // let the default root component handle it.
                    history.push('/');
                } else {
                    redirectUserToDefaultTeam();
                }
            }
        }

        return () => {
            window.removeEventListener('resize', onWindowResize);
        };
    }, []);

    useEffect(() => {
        document.title = formatMessage(
            {
                id: 'signup.title',
                defaultMessage: 'Create Account | {siteName}',
            },
            {siteName: SiteName || 'Mattermost'},
        );
    }, [formatMessage, SiteName]);

    useEffect(() => {
        if (onCustomizeHeader) {
            onCustomizeHeader({
                onBackButtonClick: handleHeaderBackButtonOnClick,
                alternateLink: isMobileView ? getAlternateLink() : undefined,
            });
        }
    }, [onCustomizeHeader, handleHeaderBackButtonOnClick, isMobileView, getAlternateLink, search]);

    useEffect(() => {
        if (submitClicked) {
            if (emailError && emailInput.current) {
                emailInput.current.focus();
            } else if (nameError && nameInput.current) {
                nameInput.current.focus();
            } else if (passwordError && passwordInput.current) {
                passwordInput.current.focus();
            }
            setSubmitClicked(false);
        }
    }, [emailError, nameError, passwordError, submitClicked]);

    if (loading) {
        return (<LoadingScreen/>);
    }

    const handleBrandImageError = () => {
        setBrandImageError(true);
    };

    const getCardTitle = () => {
        if (CustomDescriptionText) {
            return CustomDescriptionText;
        }

        if (!enableSignUpWithEmail && enableExternalSignup) {
            return formatMessage({id: 'signup_user_completed.cardtitle.external', defaultMessage: 'Create your account with one of the following:'});
        }

        return formatMessage({id: 'signup_user_completed.cardtitle', defaultMessage: 'Create your account'});
    };

    const getMessageSubtitle = () => {
        if (enableCustomBrand) {
            return CustomBrandText ? (
                <div className='signup-body-custom-branding-markdown'>
                    <Markdown
                        message={CustomBrandText}
                        options={{mentionHighlight: false}}
                    />
                </div>
            ) : null;
        }

        return (
            <p className='signup-body-message-subtitle'>
                {formatMessage({
                    id: 'signup_user_completed.subtitle',
                    defaultMessage: 'Create your Mattermost account to start collaborating with your team',
                })}
            </p>
        );
    };

    const handleEmailOnChange = ({target: {value: email}}: React.ChangeEvent<HTMLInputElement>) => {
        setEmail(email);
        dismissAlert();

        if (emailError) {
            setEmailError('');
        }
    };

    const handleNameOnChange = ({target: {value: name}}: React.ChangeEvent<HTMLInputElement>) => {
        setName(name);
        dismissAlert();

        if (nameError) {
            setNameError('');
        }
    };

    const handlePasswordInputOnChange = ({target: {value: password}}: React.ChangeEvent<HTMLInputElement>) => {
        setPassword(password);
        dismissAlert();

        if (passwordError) {
            setPasswordError('');
        }
    };

    const handleSignupSuccess = async (user: UserProfile, data: UserProfile) => {
        trackEvent('signup', 'signup_user_02_complete', getRoleFromTrackFlow());

        if (reminderInterval) {
            trackEvent('signup', `signup_from_reminder_${reminderInterval}`, {user: user.id});
        }

        const redirectTo = (new URLSearchParams(search)).get('redirect_to');

        const {error} = await dispatch(loginById(data.id, user.password));

        if (error) {
            if (error.server_error_id === 'api.user.login.not_verified.app_error') {
                let verifyUrl = '/should_verify_email?email=' + encodeURIComponent(user.email);

                if (teamName) {
                    verifyUrl += '&teamname=' + encodeURIComponent(teamName);
                }

                if (redirectTo) {
                    verifyUrl += '&redirect_to=' + redirectTo;
                }

                history.push(verifyUrl);
            } else {
                setServerError(error.message);
                setIsWaiting(false);
            }

            return;
        }

        await postSignupSuccess();
    };

    const postSignupSuccess = async () => {
        const redirectTo = (new URLSearchParams(search)).get('redirect_to');

        await dispatch(loadMe());

        if (token) {
            setGlobalItem(token, JSON.stringify({usedBefore: true}));
        }

        if (redirectTo) {
            history.push(redirectTo);
        } else if (onboardingFlowEnabled) {
            // need info about whether admin or not,
            // and whether admin has already completed
            // first tiem onboarding. Instead of fetching and orchestrating that here,
            // let the default root component handle it.
            history.push('/');
        } else {
            redirectUserToDefaultTeam();
        }
    };

    function sendSignUpTelemetryEvents(telemetryId: string, props?: any) {
        trackEvent('signup', telemetryId, props);
    }

    type TelemetryErrorList = {errors: Array<{field: string; rule: string}>; success: boolean};

    const isUserValid = () => {
        let isValid = true;
        const telemetryEvents: TelemetryErrorList = {errors: [], success: true};

        // Username(사번) 검사
        const providedUsername = nameInput.current?.value.trim();
        if (!providedUsername) {
            setNameError(formatMessage({id: 'signup_user_completed.required', defaultMessage: 'This field is required'}));
            telemetryEvents.errors.push({field: 'username', rule: 'not_provided'});
            isValid = false;
        } else if (!isEmployeeId(providedUsername)) {
            setNameError(formatMessage({id: 'signup_user_completed.validEmployeeId', defaultMessage: 'Please enter a valid employee ID (must start with K, mi, or mt, followed by numbers, total 7 characters)'}));
            telemetryEvents.errors.push({field: 'username', rule: 'invalid_employee_id'});
            isValid = false;
        }

        // Email(이름) 검사 - 단순히 비어있지 않은지만 확인
        const providedEmail = emailInput.current?.value.trim();
        if (!providedEmail) {
            setEmailError(formatMessage({id: 'signup_user_completed.required', defaultMessage: 'This field is required'}));
            telemetryEvents.errors.push({field: 'email', rule: 'not_provided'});
            isValid = false;
        }

        // 비밀번호 검사
        const providedPassword = passwordInput.current?.value ?? '';
        const {error, telemetryErrorIds} = isValidPassword(providedPassword, passwordConfig, intl);
        if (error) {
            setPasswordError(error as string);
            telemetryEvents.errors = [...telemetryEvents.errors, ...telemetryErrorIds];
            isValid = false;
        }

        if (telemetryEvents.errors.length) {
            telemetryEvents.success = false;
        }

        sendSignUpTelemetryEvents('validate_user', telemetryEvents);
        return isValid;
    };

    const dismissAlert = () => {
        setAlertBanner(null);
    };

    const handleSubmit = async (e: React.MouseEvent | React.KeyboardEvent) => {
        e.preventDefault();
        sendSignUpTelemetryEvents('click_create_account', getRoleFromTrackFlow());
        setIsWaiting(true);
        setSubmitClicked(true);

        if (isUserValid()) {
            setNameError('');
            setEmailError('');
            setPasswordError('');
            setServerError('');
            setIsWaiting(true);

            const employeeId = nameInput.current?.value.trim();
            const fullName = emailInput.current?.value.trim();

            const user = {
                email: `${employeeId}@kbfg.com`,  // 필수 필드이므로 임시 이메일 생성
                username: employeeId.toLowerCase(),  // 사번을 username으로 저장
                password: passwordInput.current?.value,
                first_name: fullName,  // 전체 이름을 first_name에 저장
                last_name: '',  // 사용하지 않음
                props: {
                    nickname_disabled: 'true',  // 닉네임 변경 비활성화
                },
                allow_marketing: true,
            };

            let redirect_to;
            if (inviteId) {
                redirect_to = `/signup_user_complete/?id=${inviteId}`;
            }

            const {data, error} = await dispatch(createUser(user, token, inviteId, redirect_to));

            if (error) {
                setAlertBanner({
                    mode: 'danger' as ModeType,
                    title: (error as ServerError).message,
                    onDismiss: dismissAlert,
                });
                setIsWaiting(false);

                // Special case for accessibility to show the error message when the username is already taken
                if (error.server_error_id === 'app.user.save.username_exists.app_error') {
                    setNameError(error.message);
                    setSubmitClicked(true);
                }
                return;
            }

            await handleSignupSuccess(user, data!);
        } else {
            setIsWaiting(false);
        }
    };

    const handleReturnButtonOnClick = () => history.replace('/');

    const getNewsletterCheck = () => {
        return (
            <div className='newsletter'>
                <span className='interested'>
                    {formatMessage({id: 'newsletter_optin.title', defaultMessage: 'Interested in receiving Mattermost security, product, promotions, and company updates updates via newsletter?'})}
                </span>
                <span className='link'>
                    {formatMessage(
                        {id: 'newsletter_optin.desc', defaultMessage: 'Sign up at <a>{link}</a>.'},
                        {
                            link: HostedCustomerLinks.SECURITY_UPDATES,
                            a: (chunks: React.ReactNode | React.ReactNodeArray) => (
                                <ExternalLink
                                    location='signup'
                                    href={HostedCustomerLinks.SECURITY_UPDATES}
                                >
                                    {chunks}
                                </ExternalLink>
                            ),
                        },
                    )}
                </span>
            </div>
        );
    };

    const handleOnBlur = (e: FocusEvent<HTMLInputElement | HTMLTextAreaElement>, inputId: string) => {
        const text = e.target.value;
        if (!text) {
            return;
        }
        sendSignUpTelemetryEvents(`typed_input_${inputId}`);
    };

    const getContent = () => {
        if (!enableSignUpWithEmail && !enableExternalSignup) {
            return (
                <ColumnLayout
                    title={formatMessage({id: 'login.noMethods.title', defaultMessage: 'This server doesn’t have any sign-in methods enabled'})}
                    message={formatMessage({id: 'login.noMethods.subtitle', defaultMessage: 'Please contact your System Administrator to resolve this.'})}
                />
            );
        }

        if (!isWaiting && (noOpenServer || serverError || usedBefore)) {
            const titleColumn = noOpenServer ? (
                formatMessage({id: 'signup_user_completed.no_open_server.title', defaultMessage: 'This server doesn’t allow open signups'})
            ) : (
                serverError ||
                formatMessage({id: 'signup_user_completed.invalid_invite.title', defaultMessage: 'This invite link is invalid'})
            );

            return (
                <ColumnLayout
                    title={titleColumn}
                    message={formatMessage({id: 'signup_user_completed.invalid_invite.message', defaultMessage: 'Please speak with your Administrator to receive an invitation.'})}
                    extraContent={(
                        <div className='signup-body-content-button-container'>
                            <button
                                className='signup-body-content-button-return'
                                onClick={handleReturnButtonOnClick}
                            >
                                {formatMessage({id: 'signup_user_completed.return', defaultMessage: 'Return to log in'})}
                            </button>
                        </div>
                    )}
                />
            );
        }

        if (desktopLoginLink) {
            return (
                <Route
                    path={'/signup_user_complete/desktop'}
                    render={() => (
                        <DesktopAuthToken
                            href={desktopLoginLink}
                            onLogin={postSignupSuccess}
                        />
                    )}
                />
            );
        }

        let emailCustomLabelForInput: CustomMessageInputType = parsedEmail ? {
            type: ItemStatus.INFO,
            value: formatMessage(
                {
                    id: 'signup_user_completed.emailIs',
                    defaultMessage: "You'll use this address to sign in to {siteName}.",
                },
                {siteName: SiteName},
            ),
        } : null;

        // error will have preference over info message
        if (emailError) {
            emailCustomLabelForInput = {type: ItemStatus.ERROR, value: emailError};
        }

        return (
            <>
                <div
                    className={classNames(
                        'signup-body-message',
                        {
                            'custom-branding': enableCustomBrand,
                            'with-brand-image': enableCustomBrand && !brandImageError,
                            'with-alternate-link': !isMobileView,
                        },
                    )}
                >
                    {enableCustomBrand && !brandImageError ? (
                        <img
                            className={classNames('signup-body-custom-branding-image')}
                            alt='brand image'
                            src={Client4.getBrandImageUrl('0')}
                            onError={handleBrandImageError}
                        />
                    ) : (
                        <h1 className='signup-body-message-title'>
                            {formatMessage({id: 'signup_user_completed.title', defaultMessage: 'Let’s get started'})}
                        </h1>
                    )}
                    {getMessageSubtitle()}
                </div>
                <div className='signup-body-action'>
                    {!isMobileView && getAlternateLink()}
                    <div className={classNames('signup-body-card', {'custom-branding': enableCustomBrand, 'with-error': hasError})}>
                        <div
                            className='signup-body-card-content'
                        >
                            <p className='signup-body-card-title'>
                                {getCardTitle()}
                            </p>
                            {enableCustomBrand && getMessageSubtitle()}
                            {alertBanner && (
                                <AlertBanner
                                    className='login-body-card-banner'
                                    mode={alertBanner.mode}
                                    title={alertBanner.title}
                                    onDismiss={alertBanner.onDismiss}
                                />
                            )}
                            {enableSignUpWithEmail && (
                                <form className='signup-body-card-form'>
                                    <Input
                                        data-testid='signup-body-card-form-email-input'
                                        ref={emailInput}
                                        name='email'
                                        className='signup-body-card-form-email-input'
                                        type='text'
                                        inputSize={SIZE.LARGE}
                                        value={email}
                                        onChange={handleEmailOnChange}
                                        placeholder='Full Name'
                                        aria-label='Full Name'
                                        disabled={isWaiting || Boolean(parsedEmail)}
                                        autoFocus={true}
                                        customMessage={emailCustomLabelForInput}
                                        onBlur={(e) => handleOnBlur(e, 'email')}
                                    />
                                    <Input
                                        data-testid='signup-body-card-form-name-input'
                                        ref={nameInput}
                                        name='name'
                                        className='signup-body-card-form-name-input'
                                        type='text'
                                        inputSize={SIZE.LARGE}
                                        value={name}
                                        onChange={handleNameOnChange}
                                        placeholder='Employee ID'
                                        aria-label='Employee ID'
                                        disabled={isWaiting}
                                        autoFocus={Boolean(parsedEmail)}
                                        customMessage={
                                            nameError ? {type: ItemStatus.ERROR, value: nameError} : {
                                                type: ItemStatus.INFO,
                                                value: 'Please enter your employee ID (must start with K, mi, or mt, followed by numbers, total 7 characters). Case insensitive.',
                                            }
                                        }
                                        onBlur={(e) => handleOnBlur(e, 'username')}
                                    />
                                    <PasswordInput
                                        data-testid='signup-body-card-form-password-input'
                                        ref={passwordInput}
                                        className='signup-body-card-form-password-input'
                                        value={password}
                                        inputSize={SIZE.LARGE}
                                        onChange={handlePasswordInputOnChange}
                                        disabled={isWaiting}
                                        createMode={true}
                                        info={passwordInfo as string}
                                        error={passwordError}
                                        onBlur={(e) => handleOnBlur(e, 'password')}
                                    />
                                    <p className='signup-body-card-form-contact-info' style={{marginTop: '20px', color: 'rgba(63, 67, 80, 0.75)', fontSize: '12px', textAlign: 'center'}}>
                                        사용/가입 관련 문의가 있으시면 KB국민카드 정보개발부 공통개발팀 정한라 계장에게 문의 부탁드립니다.(4623)
                                    </p>
                                    <SaveButton
                                        extraClasses='signup-body-card-form-button-submit large'
                                        saving={isWaiting}
                                        disabled={!canSubmit}
                                        onClick={handleSubmit}
                                        defaultMessage={formatMessage({id: 'signup_user_completed.create', defaultMessage: 'Create account'})}
                                        savingMessage={formatMessage({id: 'signup_user_completed.saving', defaultMessage: 'Creating account…'})}
                                    />
                                </form>
                            )}
                            {enableSignUpWithEmail && enableExternalSignup && (
                                <div className='signup-body-card-form-divider'>
                                    <span className='signup-body-card-form-divider-label'>
                                        {formatMessage({id: 'signup_user_completed.or', defaultMessage: 'or create an account with'})}
                                    </span>
                                </div>
                            )}
                            {enableExternalSignup && (
                                <div className={classNames('signup-body-card-form-login-options', {column: !enableSignUpWithEmail})}>
                                    {getExternalSignupOptions().map((option) => (
                                        <ExternalLoginButton
                                            key={option.id}
                                            direction={enableSignUpWithEmail ? undefined : 'column'}
                                            {...option}
                                        />
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </>
        );
    };

    return (
        <div className='signup-body'>
            <div className='signup-body-content'>
                {getContent()}
            </div>
        </div>
    );
};

export default Signup;
