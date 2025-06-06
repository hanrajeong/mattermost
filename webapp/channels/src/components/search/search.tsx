// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useEffect, useState, useRef, useCallback} from 'react';
import type {ChangeEvent, FormEvent} from 'react';
import {useIntl, FormattedMessage} from 'react-intl';
import {useSelector} from 'react-redux';
import debounce from 'lodash/debounce';
import {trackEvent} from 'actions/telemetry_actions';

import {getCurrentChannelNameForSearchShortcut} from 'mattermost-redux/selectors/entities/channels';

import HeaderIconWrapper from 'components/channel_header/components/header_icon_wrapper';
import SearchBar from 'components/search_bar/search_bar';
import SearchHint from 'components/search_hint/search_hint';
import SearchResults from 'components/search_results';
import type Provider from 'components/suggestion/provider';
import SearchChannelProvider from 'components/suggestion/search_channel_provider';
import SearchDateProvider from 'components/suggestion/search_date_provider';
import SearchUserProvider from 'components/suggestion/search_user_provider';
import SearchIcon from 'components/widgets/icons/search_icon';
import Popover from 'components/widgets/popover';

import Constants, {searchHintOptions, RHSStates, searchFilesHintOptions} from 'utils/constants';
import * as Keyboard from 'utils/keyboard';
import {isServerVersionGreaterThanOrEqualTo} from 'utils/server_version';
import {isDesktopApp, getDesktopVersion, isMacApp} from 'utils/user_agent';

import type {SearchType} from 'types/store/rhs';

import type {Props, SearchFilterType} from './types';

declare namespace JSX {
    interface IntrinsicElements {
        div: React.DetailedHTMLProps<React.HTMLAttributes<HTMLDivElement>, HTMLDivElement>;
        span: React.DetailedHTMLProps<React.HTMLAttributes<HTMLSpanElement>, HTMLSpanElement>;
        h4: React.DetailedHTMLProps<React.HTMLAttributes<HTMLHeadingElement>, HTMLHeadingElement>;
        button: React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement>;
        a: React.DetailedHTMLProps<React.AnchorHTMLAttributes<HTMLAnchorElement>, HTMLAnchorElement>;
        i: React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>;
        ul: React.DetailedHTMLProps<React.HTMLAttributes<HTMLUListElement>, HTMLUListElement>;
        li: React.DetailedHTMLProps<React.LiHTMLAttributes<HTMLLIElement>, HTMLLIElement>;
        form: React.DetailedHTMLProps<React.FormHTMLAttributes<HTMLFormElement>, HTMLFormElement>;
        input: React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement>;
    }
}

interface SearchHintOption {
    searchTerm: string;
    message: {
        id: string;
        defaultMessage: string;
    };
}

// 검색 필터 옵션을 항상 표시하도록 수정된 함수
const determineVisibleSearchHintOptions = (searchTerms: string, searchType: SearchType): SearchHintOption[] => {
    // 검색 타입에 따라 적절한 옵션 선택
    let options = searchHintOptions;
    if (searchType === 'files') {
        options = searchFilesHintOptions;
    }
    
    // 검색어에 상관없이 항상 모든 필터 옵션 표시
    return options;
};

const Search = ({
    actions: {
        autocompleteChannelsForSearch,
        autocompleteUsersInTeam,
        closeRightHandSide,
        filterFilesSearchByExt,
        getMoreFilesForSearch,
        getMorePostsForSearch,
        openRHSSearch,
        setRhsExpanded,
        showChannelFiles,
        showSearchResults,
        updateRhsState,
        updateSearchTeam,
        updateSearchTerms,
        updateSearchTermsForShortcut,
        updateSearchType,
    },
    crossTeamSearchEnabled,
    hideMobileSearchBarInRHS,
    isChannelFiles,
    isFlaggedPosts,
    isMentionSearch,
    isMobileView,
    isPinnedPosts,
    isRhsExpanded,
    isSearchingTerm,
    searchTeam,
    searchTerms,
    searchType,
    searchVisible,
    channelDisplayName,
    children,
    currentChannel,
    enableFindShortcut,
    getFocus,
    hideSearchBar,
    isSideBarRight,
    isSideBarRightOpen,
}: Props): React.ReactElement => {
    const intl = useIntl();
    const currentChannelName = useSelector(getCurrentChannelNameForSearchShortcut);

    // generate intial component state and setters
    const [focused, setFocused] = useState<boolean>(false);
    const [dropdownFocused, setDropdownFocused] = useState<boolean>(false);
    const [keepInputFocused, setKeepInputFocused] = useState<boolean>(false);
    const [indexChangedViaKeyPress, setIndexChangedViaKeyPress] = useState<boolean>(false);
    const [highlightedSearchHintIndex, setHighlightedSearchHintIndex] = useState<number>(-1);
    const [visibleSearchHintOptions, setVisibleSearchHintOptions] = useState<SearchHintOption[]>(
        determineVisibleSearchHintOptions(searchTerms, searchType),
    );
    const [searchFilterType, setSearchFilterType] = useState<SearchFilterType>('all');

    const suggestionProviders = useRef<Provider[]>([
        new SearchDateProvider(),
        new SearchChannelProvider(autocompleteChannelsForSearch),
        new SearchUserProvider(autocompleteUsersInTeam),
    ]);

    const isDesktop = isDesktopApp() && isServerVersionGreaterThanOrEqualTo(getDesktopVersion(), '4.7.0');
    useEffect(() => {
        if (!enableFindShortcut) {
            return undefined;
        }

        const handleKeyDown = (e: KeyboardEvent) => {
            if (Keyboard.cmdOrCtrlPressed(e) && Keyboard.isKeyPressed(e, Constants.KeyCodes.F)) {
                if (!isDesktop && !e.shiftKey) {
                    return;
                }

                // Special case for Mac Desktop xApp where Ctrl+Cmd+F triggers full screen view
                if (isMacApp() && e.ctrlKey) {
                    return;
                }

                e.preventDefault();
                if (hideSearchBar) {
                    openRHSSearch();
                    setKeepInputFocused(true);
                }
                if (currentChannelName) {
                    updateSearchTermsForShortcut();
                }
                handleFocus();
            }
        };

        document.addEventListener('keydown', handleKeyDown);
        return () => {
            document.removeEventListener('keydown', handleKeyDown);
        };
    }, [hideSearchBar, currentChannelName]);

    useEffect((): void => {
        if (isMobileView && isSideBarRight) {
            handleFocus();
        }
    }, [isMobileView, isSideBarRight]);

    useEffect(() => {
        // 항상 모든 필터 옵션을 표시하도록 변경
        const options = searchType === 'files' ? searchFilesHintOptions : searchHintOptions;
        setVisibleSearchHintOptions(options);
    }, [searchType]);

    useEffect((): void => {
        if (!isMobileView && focused && keepInputFocused) {
            handleBlur();
        }
    }, [isMobileView, searchTerms]);

    const getMorePostsForSearchCallback = useCallback(() => {
        let team = searchTeam;
        if (isMentionSearch) {
            team = '';
        }
        getMorePostsForSearch(team);
    }, [searchTeam, isMentionSearch, getMorePostsForSearch]);

    const getMoreFilesForSearchCallback = useCallback(() => {
        let team = searchTeam;
        if (isMentionSearch) {
            team = '';
        }
        getMoreFilesForSearch(team);
    }, [searchTeam, isMentionSearch, getMoreFilesForSearch]);

    // handle cloding of rhs-flyout
    const handleClose = (): void => closeRightHandSide();

    // focus the search input
    const handleFocus = (): void => setFocused(true);

    // release focus from the search input or unset `keepInputFocused` value
    // `keepInputFocused` is used to keep the search input focused when a
    // user selects a suggestion from `SearchHint` with a click
    const handleBlur = (): void => {
        // add time out so that the pinned and member buttons are clickable
        // when focus is released from the search box.
        setTimeout((): void => {
            if (keepInputFocused) {
                setKeepInputFocused(false);
            } else {
                setFocused(false);
            }
        }, 0);
        updateHighlightedSearchHint(0, false);
    };

    const handleDropdownBlur = () => setDropdownFocused(false);

    const handleDropdownFocus = () => setDropdownFocused(true);

    const handleSearchHintSelection = (): void => {
        if (focused) {
            setKeepInputFocused(true);
        } else {
            setFocused(true);
        }
    };

    // 검색어에 새 필터 추가 함수 - 여러 필터를 동시에 적용할 수 있도록 수정
    const handleAddSearchTerm = (term: string): void => {
        const pretextArray = searchTerms.split(/\s+/g);
        const pretext = pretextArray[pretextArray.length - 1];
        const prefix = (pretext.startsWith('@') && term === 'From:') ? '' : ' ';
        
        // 검색어에 새 필터 추가
        const newSearchTerms = searchTerms + prefix + term + ' ';
        updateSearchTerms(newSearchTerms);
        
        // 필터 추가 후 바로 검색 실행
        showSearchResults(false);
        
        // 검색창에 포커스 유지
        setFocused(true);
        setKeepInputFocused(true);
    };

    const handleUpdateSearchTeamFromResult = async (teamId: string) => {
        updateSearchTeam(teamId);
        const newTerms = searchTerms.
            replace(/\bin:[^\s]*/gi, '').replace(/\s{2,}/g, ' ').
            replace(/\bfrom:[^\s]*/gi, '').replace(/\s{2,}/g, ' ');

        if (newTerms.trim() !== searchTerms.trim()) {
            updateSearchTerms(newTerms);
        }

        handleSearch().then(() => {
            setKeepInputFocused(false);
            setFocused(false);
        });
    };

    const handleUpdateSearchTerms = (terms: string): void => {
        updateSearchTerms(terms);
        updateHighlightedSearchHint();
    };

    const handleOnSearchTypeSelected = (searchType || searchTerms) ? undefined : (value: SearchType) => {
        updateSearchType(value);
        if (!searchType) {
            setDropdownFocused(false);
        }
        setFocused(true);
    };

    // 검색 실행 함수 - 검색 버튼 클릭이나 엔터 키 입력 시 호출됨
    const handleSearch = async (): Promise<void> => {
        updateHighlightedSearchHint(0, false);

        // clear the search filter
        if (searchType) {
            updateSearchType('');
        }

        if (!searchTerms.trim()) {
            return;
        }

        try {
            await showSearchResults(false);
        } finally {
            handleSearchOnSuccess();
        }
    };

    // 검색어 입력 시 실시간으로 검색 결과 업데이트를 위한 디바운스 함수
    const debouncedSearch = useCallback(
        debounce((term: string) => {
            if (term.trim().length > 0) {
                // 검색 결과를 표시하고 RHS를 열어줌
                openRHSSearch();
                showSearchResults(isMentionSearch);
                const teamId = searchTeam || currentChannel?.team_id;
                updateSearchTeam(teamId);
                trackEvent('ui', 'ui_rhs_search');
            }
        }, 300), // 디바운스 시간을 300ms로 줄여 더 빠르게 반응하도록 함
        [openRHSSearch, showSearchResults, isMentionSearch, searchTeam, currentChannel, updateSearchTeam]
    );

    const handleChange = (e: ChangeEvent<HTMLInputElement>): void => {
        const term = e.target.value;
        updateSearchTerms(term);
        
        // 실시간 검색 구현 - 타이핑할 때마다 검색 결과 업데이트
        if (term.trim()) {
            // 검색어가 있을 때만 검색 실행
            debouncedSearch();
        }

        // update the visible search hint options based on the current search terms
        setVisibleSearchHintOptions(determineVisibleSearchHintOptions(term, searchType));

        const caps = e.target.value.length > 0 && e.target.value.charAt(0) === e.target.value.charAt(0).toUpperCase();

        if (!ignoreChannelMentionCapsRef.current && caps && term.toLowerCase().indexOf('from:') !== 0 && term.toLowerCase().indexOf('in:') !== 0) {
            updateAutocompleteChannelsForSearch(term, true);
            ignoreChannelMentionCapsRef.current = true;
        } else if (ignoreChannelMentionCapsRef.current && caps && term.toLowerCase().indexOf('from:') !== 0 && term.toLowerCase().indexOf('in:') !== 0) {
            updateAutocompleteChannelsForSearch(term, true);
        } else if (term.toLowerCase().indexOf('from:') === 0) {
            updateAutocompleteUsersForSearch(term.substring(5), '');  // strip off from: to get user started
        }
    };

    const handleEnterKey = (e: ChangeEvent<HTMLInputElement>): void => {
        e.preventDefault();

        if (indexChangedViaKeyPress) {
            setKeepInputFocused(true);
            if (!searchType && !searchTerms) {
                updateSearchType(highlightedSearchHintIndex === 0 ? 'messages' : 'files');
                setHighlightedSearchHintIndex(-1);
            } else {

                handleAddSearchTerm(visibleSearchHintOptions[highlightedSearchHintIndex].searchTerm);
            }
            return;
        }

        if (isMentionSearch) {
            updateRhsState(RHSStates.SEARCH);
        }

        handleSearch().then(() => {
            setKeepInputFocused(false);
            setFocused(false);
        });
    };

    const handleSubmit = (e: FormEvent<HTMLFormElement>): void => {
        e.preventDefault();

        handleSearch().then(() => {
            setKeepInputFocused(false);
            setFocused(false);
        });
    };

    const handleSearchOnSuccess = (): void => {
        if (isMobileView) {
            handleClear();
        }
    };

    const handleClear = (): void => {
        if (isMentionSearch) {
            setFocused(false);
            updateRhsState(RHSStates.SEARCH);
        }
        updateSearchTerms('');
        updateSearchTeam(null);
        updateSearchType('');
    };

    const handleShrink = (): void => {
        setRhsExpanded(false);
    };

    const handleSetSearchFilter = (filterType: SearchFilterType): void => {
        // 필터 타입이 같으면 변경하지 않음
        if (searchFilterType === filterType) {
            return;
        }

        setSearchFilterType(filterType);
        updateRhsState(RHSStates.SEARCH);
        
        // 필터 변경 후 바로 검색 결과 업데이트
        showSearchResults(false);
    };

    const setHoverHintIndex = (_highlightedSearchHintIndex: number): void => {
        setHighlightedSearchHintIndex(_highlightedSearchHintIndex);
        setIndexChangedViaKeyPress(false);
    };

    const searchButtonClick = (e: React.MouseEvent) => {
        e.preventDefault();

        openRHSSearch();
    };

    // 검색 필터 초기화 함수
    const handleClearFilters = (): void => {
        // 검색어에서 모든 필터 제거
        const cleanedTerms = searchTerms.replace(/(From:|In:|On:|Before:|After:|Ext:)\s*\S+/gi, '').trim();
        updateSearchTerms(cleanedTerms);
        
        // 필터 제거 후 바로 검색 실행
        if (cleanedTerms) {
            showSearchResults(false);
        }
        
        // 검색창에 포커스 유지
        setFocused(true);
        setKeepInputFocused(true);
    };

    // 검색 힌트 팝오버 렌더링 함수 - 항상 표시되도록 수정 및 위치 변경
    const renderHintPopover = (): JSX.Element => {
        // 사용된 필터 개수 계산
        let filtersUsed = 0;
        const filterRegex = /(From:|In:|On:|Before:|After:|Ext:)\s*\S+/gi;
        const matches = searchTerms.match(filterRegex) || [];
        filtersUsed = matches.length;

        if (visibleSearchHintOptions.length === 0 || isMentionSearch) {
            return <></>;
        }

        // 항상 표시되도록 helpClass 수정
        const helpClass = 'search-help-popover visible';

        return (
            <Popover
                id={`${isSideBarRight ? 'sbr-' : ''}searchbar-help-popup`}
                placement='bottom-end' // 위치를 검색창 하단 오른쪽으로 변경
                className={helpClass}
                style={{marginTop: '4px', width: '280px', position: 'absolute', right: 0}} // 스타일 조정
            >
                <div className="search-filters-header" style={{display: 'flex', justifyContent: 'space-between', padding: '8px 16px', borderBottom: '1px solid rgba(0, 0, 0, 0.1)'}}>
                    <h4 className="search-filters-title" style={{margin: 0, fontSize: '14px', fontWeight: 600}}>
                        <FormattedMessage
                            id="search_bar.filters.title"
                            defaultMessage="Filter your search"
                        />
                    </h4>
                    {filtersUsed > 0 && (
                        <button 
                            type="button"
                            className="search-filters-clear-button" 
                            onClick={handleClearFilters}
                            aria-label="Clear all filters"
                            style={{background: 'none', border: 'none', color: 'var(--button-bg)', cursor: 'pointer', fontSize: '13px', padding: '0 8px'}}
                        >
                            <FormattedMessage
                                id="search_bar.filters.clear"
                                defaultMessage="Clear filters"
                            />
                        </button>
                    )}
                </div>
                <SearchHint
                    options={visibleSearchHintOptions}
                    withTitle={false} // 제목을 위에서 직접 추가했으므로 false로 설정
                    onOptionSelected={handleAddSearchTerm}
                    onMouseDown={handleSearchHintSelection}
                    highlightedIndex={highlightedSearchHintIndex}
                    onOptionHover={setHoverHintIndex}
                    onSearchTypeSelected={handleOnSearchTypeSelected}
                    onElementBlur={handleDropdownBlur}
                    onElementFocus={handleDropdownFocus}
                    searchType={searchType}
                />
            </Popover>
        );
    };

    const renderSearchBar = (): JSX.Element => (
        <>
            <div className='sidebar-collapse__container'>
                <div
                    id={isSideBarRight ? 'sbrSidebarCollapse' : 'sidebarCollapse'}
                    className='sidebar-collapse'
                    onClick={handleClose}
                >
                    <span
                        className='fa fa-2x fa-angle-left'
                        title={intl.formatMessage({id: 'generic_icons.back', defaultMessage: 'Back Icon'})}
                    />
                </div>
            </div>
            <SearchBar
                updateHighlightedSearchHint={(delta: number, keyPress: boolean) => updateHighlightedSearchHint(delta, keyPress)}
                handleEnterKey={handleEnterKey}
                handleClear={handleClear}
                handleChange={handleChange}
                handleSubmit={handleSubmit}
                handleFocus={handleFocus}
                handleBlur={handleBlur}
                keepFocused={keepInputFocused}
                setKeepFocused={setKeepInputFocused}
                isFocused={focused}
                suggestionProviders={suggestionProviders.current}
                isSideBarRight={isSideBarRight}
                isSearchingTerm={isSearchingTerm}
                getFocus={getFocus}
                searchTerms={searchTerms}
                searchType={searchType}
                clearSearchType={() => updateSearchType('')}
            >
                {!isMobileView && (
                    <div className="search-filters-container" style={{position: 'absolute', right: 0, top: '100%', zIndex: 1000}}>
                        {renderHintPopover()}
                    </div>
                )}
            </SearchBar>
        </>
    );

    // when inserted in RHSSearchNav component, just return SearchBar
    if (!isSideBarRight) {
        if (hideSearchBar) {
            return (
                <HeaderIconWrapper
                    buttonId={'channelHeaderSearchButton'}
                    onClick={searchButtonClick}
                    tooltip={intl.formatMessage({id: 'channel_header.search', defaultMessage: 'Search'})}
                >
                    <SearchIcon
                        className='icon icon--standard'
                        aria-hidden='true'
                    />
                    {null /* 필요한 children 속성 추가 */}
                </HeaderIconWrapper>
            );
        }

        return (
            <div
                id='searchbarContainer'
                className={'search-bar-container--global'}
            >
                <div className='sidebar-right__table'>
                    {renderSearchBar()}
                </div>
            </div>
        );
    }

    return (
        <div className='sidebar--right__content'>
            {!hideMobileSearchBarInRHS && (
                <div className='search-bar__container channel-header alt'>
                    <div className='sidebar-right__table'>
                        {renderSearchBar()}
                    </div>
                </div>
            )}
            {searchVisible ? (
                <SearchResults
                    isMentionSearch={isMentionSearch}
                    isFlaggedPosts={isFlaggedPosts}
                    isPinnedPosts={isPinnedPosts}
                    isChannelFiles={isChannelFiles}
                    shrink={handleShrink}
                    channelDisplayName={channelDisplayName}
                    isOpened={isSideBarRightOpen}
                    updateSearchTerms={handleAddSearchTerm}
                    updateSearchTeam={handleUpdateSearchTeamFromResult}
                    handleSearchHintSelection={handleSearchHintSelection}
                    isSideBarExpanded={isRhsExpanded}
                    getMorePostsForSearch={getMorePostsForSearchCallback}
                    getMoreFilesForSearch={getMoreFilesForSearchCallback}
                    setSearchFilterType={handleSetSearchFilter}
                    searchFilterType={searchFilterType}
                    setSearchType={(value: SearchType) => updateSearchType(value)}
                    searchType={searchType || 'messages'}
                    crossTeamSearchEnabled={crossTeamSearchEnabled}
                />
            ) : children}
        </div>
    );
};

const defaultProps: Partial<Props> = {
    searchTerms: '',
    channelDisplayName: '',
    isSideBarRight: false,
    hideMobileSearchBarInRHS: false,
    getFocus: () => {},
};

Search.defaultProps = defaultProps;

export default React.memo(Search);
