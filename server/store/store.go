//go:generate go run layer_generators/main.go

// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/product"
)

type StoreResult struct {
	Data any

	// NErr a temporary field used by the new code for the AppError migration. This will later become Err when the entire store is migrated.
	NErr error
}

type Store interface {
	Team() TeamStore
	Channel() ChannelStore
	Post() PostStore
	RetentionPolicy() RetentionPolicyStore
	Thread() ThreadStore
	User() UserStore
	Bot() BotStore
	Audit() AuditStore
	ClusterDiscovery() ClusterDiscoveryStore
	RemoteCluster() RemoteClusterStore
	Compliance() ComplianceStore
	Session() SessionStore
	OAuth() OAuthStore
	System() SystemStore
	Webhook() WebhookStore
	Command() CommandStore
	CommandWebhook() CommandWebhookStore
	Preference() PreferenceStore
	License() LicenseStore
	Token() TokenStore
	Emoji() EmojiStore
	Status() StatusStore
	FileInfo() FileInfoStore
	UploadSession() UploadSessionStore
	Reaction() ReactionStore
	Role() RoleStore
	Scheme() SchemeStore
	Job() JobStore
	UserAccessToken() UserAccessTokenStore
	ChannelMemberHistory() ChannelMemberHistoryStore
	Plugin() PluginStore
	TermsOfService() TermsOfServiceStore
	ProductNotices() ProductNoticesStore
	Group() GroupStore
	UserTermsOfService() UserTermsOfServiceStore
	LinkMetadata() LinkMetadataStore
	SharedChannel() SharedChannelStore
	Draft() DraftStore
	MarkSystemRanUnitTests()
	Close()
	LockToMaster()
	UnlockFromMaster()
	DropAllTables()
	RecycleDBConnections(d time.Duration)
	GetDBSchemaVersion() (int, error)
	GetAppliedMigrations() ([]model.AppliedMigration, error)
	GetDbVersion(numerical bool) (string, error)
	// GetInternalMasterDB allows access to the raw master DB
	// handle for the multi-product architecture.
	GetInternalMasterDB() *sql.DB
	// GetInternalReplicaDBs allows access to the raw replica DB
	// handles for the multi-product architecture.
	GetInternalReplicaDB() *sql.DB
	GetInternalReplicaDBs() []*sql.DB
	TotalMasterDbConnections() int
	TotalReadDbConnections() int
	TotalSearchDbConnections() int
	ReplicaLagTime() error
	ReplicaLagAbs() error
	CheckIntegrity() <-chan model.IntegrityCheckResult
	SetContext(context context.Context)
	Context() context.Context
	NotifyAdmin() NotifyAdminStore
	PostPriority() PostPriorityStore
	PostAcknowledgement() PostAcknowledgementStore
	TrueUpReview() TrueUpReviewStore
}

type RetentionPolicyStore interface {
	Save(policy *model.RetentionPolicyWithTeamAndChannelIDs) (*model.RetentionPolicyWithTeamAndChannelCounts, error)
	Patch(patch *model.RetentionPolicyWithTeamAndChannelIDs) (*model.RetentionPolicyWithTeamAndChannelCounts, error)
	Get(id string) (*model.RetentionPolicyWithTeamAndChannelCounts, error)
	GetAll(offset, limit int) ([]*model.RetentionPolicyWithTeamAndChannelCounts, error)
	GetCount() (int64, error)
	Delete(id string) error
	GetChannels(policyId string, offset, limit int) (model.ChannelListWithTeamData, error)
	GetChannelsCount(policyId string) (int64, error)
	AddChannels(policyId string, channelIds []string) error
	RemoveChannels(policyId string, channelIds []string) error
	GetTeams(policyId string, offset, limit int) ([]*model.Team, error)
	GetTeamsCount(policyId string) (int64, error)
	AddTeams(policyId string, teamIds []string) error
	RemoveTeams(policyId string, teamIds []string) error
	DeleteOrphanedRows(limit int) (int64, error)
	GetTeamPoliciesForUser(userID string, offset, limit int) ([]*model.RetentionPolicyForTeam, error)
	GetTeamPoliciesCountForUser(userID string) (int64, error)
	GetChannelPoliciesForUser(userID string, offset, limit int) ([]*model.RetentionPolicyForChannel, error)
	GetChannelPoliciesCountForUser(userID string) (int64, error)
}

type TeamStore interface {
	Save(team *model.Team) (*model.Team, error)
	Update(team *model.Team) (*model.Team, error)
	Get(id string) (*model.Team, error)
	GetMany(ids []string) ([]*model.Team, error)
	GetByName(name string) (*model.Team, error)
	GetByNames(name []string) ([]*model.Team, error)
	SearchAll(opts *model.TeamSearch) ([]*model.Team, error)
	SearchAllPaged(opts *model.TeamSearch) ([]*model.Team, int64, error)
	SearchOpen(opts *model.TeamSearch) ([]*model.Team, error)
	SearchPrivate(opts *model.TeamSearch) ([]*model.Team, error)
	GetAll() ([]*model.Team, error)
	GetAllPage(offset int, limit int, opts *model.TeamSearch) ([]*model.Team, error)
	GetAllPrivateTeamListing() ([]*model.Team, error)
	GetAllTeamListing() ([]*model.Team, error)
	GetTeamsByUserId(userID string) ([]*model.Team, error)
	GetByInviteId(inviteID string) (*model.Team, error)
	GetByEmptyInviteID() ([]*model.Team, error)
	PermanentDelete(teamID string) error
	AnalyticsTeamCount(opts *model.TeamSearch) (int64, error)
	SaveMultipleMembers(members []*model.TeamMember, maxUsersPerTeam int) ([]*model.TeamMember, error)
	SaveMember(member *model.TeamMember, maxUsersPerTeam int) (*model.TeamMember, error)
	UpdateMember(member *model.TeamMember) (*model.TeamMember, error)
	UpdateMultipleMembers(members []*model.TeamMember) ([]*model.TeamMember, error)
	GetMember(ctx context.Context, teamID string, userID string) (*model.TeamMember, error)
	GetMembers(teamID string, offset int, limit int, teamMembersGetOptions *model.TeamMembersGetOptions) ([]*model.TeamMember, error)
	GetMembersByIds(teamID string, userIds []string, restrictions *model.ViewUsersRestrictions) ([]*model.TeamMember, error)
	GetTotalMemberCount(teamID string, restrictions *model.ViewUsersRestrictions) (int64, error)
	GetActiveMemberCount(teamID string, restrictions *model.ViewUsersRestrictions) (int64, error)
	GetTeamsForUser(ctx context.Context, userID, excludeTeamID string, includeDeleted bool) ([]*model.TeamMember, error)
	GetTeamsForUserWithPagination(userID string, page, perPage int) ([]*model.TeamMember, error)
	GetChannelUnreadsForAllTeams(excludeTeamID, userID string) ([]*model.ChannelUnread, error)
	GetChannelUnreadsForTeam(teamID, userID string) ([]*model.ChannelUnread, error)
	RemoveMember(teamID string, userID string) error
	RemoveMembers(teamID string, userIds []string) error
	RemoveAllMembersByTeam(teamID string) error
	RemoveAllMembersByUser(userID string) error
	UpdateLastTeamIconUpdate(teamID string, curTime int64) error
	GetTeamsByScheme(schemeID string, offset int, limit int) ([]*model.Team, error)
	MigrateTeamMembers(fromTeamID string, fromUserID string) (map[string]string, error)
	ResetAllTeamSchemes() error
	ClearAllCustomRoleAssignments() error
	AnalyticsGetTeamCountForScheme(schemeID string) (int64, error)
	GetAllForExportAfter(limit int, afterID string) ([]*model.TeamForExport, error)
	GetTeamMembersForExport(userID string) ([]*model.TeamMemberForExport, error)
	UserBelongsToTeams(userID string, teamIds []string) (bool, error)
	GetUserTeamIds(userID string, allowFromCache bool) ([]string, error)
	InvalidateAllTeamIdsForUser(userID string)
	ClearCaches()

	// UpdateMembersRole sets all of the given team members to admins and all of the other members of the team to
	// non-admin members.
	UpdateMembersRole(teamID string, userIDs []string) error

	// GroupSyncedTeamCount returns the count of non-deleted group-constrained teams.
	GroupSyncedTeamCount() (int64, error)

	// GetCommonTeamIDsForTwoUsers returns the intersection of all the teams to which the specified
	// users belong.
	GetCommonTeamIDsForTwoUsers(userID, otherUserID string) ([]string, error)

	GetNewTeamMembersSince(teamID string, since int64, offset int, limit int) (*model.NewTeamMembersList, int64, error)
}

type ChannelStore interface {
	Save(channel *model.Channel, maxChannelsPerTeam int64) (*model.Channel, error)
	CreateDirectChannel(userID *model.User, otherUserID *model.User, channelOptions ...model.ChannelOption) (*model.Channel, error)
	SaveDirectChannel(channel *model.Channel, member1 *model.ChannelMember, member2 *model.ChannelMember) (*model.Channel, error)
	Update(channel *model.Channel) (*model.Channel, error)
	UpdateSidebarChannelCategoryOnMove(channel *model.Channel, newTeamID string) error
	ClearSidebarOnTeamLeave(userID, teamID string) error
	Get(id string, allowFromCache bool) (*model.Channel, error)
	GetMany(ids []string, allowFromCache bool) (model.ChannelList, error)
	InvalidateChannel(id string)
	InvalidateChannelByName(teamID, name string)
	Delete(channelID string, timestamp int64) error
	Restore(channelID string, timestamp int64) error
	SetDeleteAt(channelID string, deleteAt int64, updateAt int64) error
	PermanentDelete(channelID string) error
	PermanentDeleteByTeam(teamID string) error
	GetByName(team_id string, name string, allowFromCache bool) (*model.Channel, error)
	GetByNames(team_id string, names []string, allowFromCache bool) ([]*model.Channel, error)
	GetByNameIncludeDeleted(team_id string, name string, allowFromCache bool) (*model.Channel, error)
	GetDeletedByName(team_id string, name string) (*model.Channel, error)
	GetDeleted(team_id string, offset int, limit int, userID string) (model.ChannelList, error)
	GetChannels(teamID, userID string, opts *model.ChannelSearchOpts) (model.ChannelList, error)
	GetChannelsWithCursor(teamId string, userId string, opts *model.ChannelSearchOpts, afterChannelID string) (model.ChannelList, error)
	GetChannelsByUser(userID string, includeDeleted bool, lastDeleteAt, pageSize int, fromChannelID string) (model.ChannelList, error)
	GetAllChannelMembersById(id string) ([]string, error)
	GetAllChannels(page, perPage int, opts ChannelSearchOpts) (model.ChannelListWithTeamData, error)
	GetAllChannelsCount(opts ChannelSearchOpts) (int64, error)
	GetMoreChannels(teamID string, userID string, offset int, limit int) (model.ChannelList, error)
	GetPrivateChannelsForTeam(teamID string, offset int, limit int) (model.ChannelList, error)
	GetPublicChannelsForTeam(teamID string, offset int, limit int) (model.ChannelList, error)
	GetPublicChannelsByIdsForTeam(teamID string, channelIds []string) (model.ChannelList, error)
	GetChannelCounts(teamID string, userID string) (*model.ChannelCounts, error)
	GetTeamChannels(teamID string) (model.ChannelList, error)
	GetAll(teamID string) ([]*model.Channel, error)
	GetChannelsByIds(channelIds []string, includeDeleted bool) ([]*model.Channel, error)
	GetChannelsWithTeamDataByIds(channelIds []string, includeDeleted bool) ([]*model.ChannelWithTeamData, error)
	GetForPost(postID string) (*model.Channel, error)
	SaveMultipleMembers(members []*model.ChannelMember) ([]*model.ChannelMember, error)
	SaveMember(member *model.ChannelMember) (*model.ChannelMember, error)
	UpdateMember(member *model.ChannelMember) (*model.ChannelMember, error)
	UpdateMultipleMembers(members []*model.ChannelMember) ([]*model.ChannelMember, error)
	// UpdateMemberNotifyProps patches the notifyProps field with the given props map.
	// It replaces existing fields and creates new ones which don't exist.
	UpdateMemberNotifyProps(channelID, userID string, props map[string]string) (*model.ChannelMember, error)
	GetMembers(channelID string, offset, limit int) (model.ChannelMembers, error)
	GetMember(ctx context.Context, channelID string, userID string) (*model.ChannelMember, error)
	GetChannelMembersTimezones(channelID string) ([]model.StringMap, error)
	GetAllChannelMembersForUser(userID string, allowFromCache bool, includeDeleted bool) (map[string]string, error)
	InvalidateAllChannelMembersForUser(userID string)
	IsUserInChannelUseCache(userID string, channelID string) bool
	GetAllChannelMembersNotifyPropsForChannel(channelID string, allowFromCache bool) (map[string]model.StringMap, error)
	InvalidateCacheForChannelMembersNotifyProps(channelID string)
	GetMemberForPost(postID string, userID string) (*model.ChannelMember, error)
	InvalidateMemberCount(channelID string)
	GetMemberCountFromCache(channelID string) int64
	GetFileCount(channelID string) (int64, error)
	GetMemberCount(channelID string, allowFromCache bool) (int64, error)
	GetMemberCountsByGroup(ctx context.Context, channelID string, includeTimezones bool) ([]*model.ChannelMemberCountByGroup, error)
	InvalidatePinnedPostCount(channelID string)
	GetPinnedPostCount(channelID string, allowFromCache bool) (int64, error)
	InvalidateGuestCount(channelID string)
	GetGuestCount(channelID string, allowFromCache bool) (int64, error)
	GetPinnedPosts(channelID string) (*model.PostList, error)
	RemoveMember(channelID string, userID string) error
	RemoveMembers(channelID string, userIds []string) error
	PermanentDeleteMembersByUser(userID string) error
	PermanentDeleteMembersByChannel(channelID string) error
	UpdateLastViewedAt(channelIds []string, userID string) (map[string]int64, error)
	UpdateLastViewedAtPost(unreadPost *model.Post, userID string, mentionCount, mentionCountRoot, urgentMentionCount int, setUnreadCountRoot bool) (*model.ChannelUnreadAt, error)
	CountPostsAfter(channelID string, timestamp int64, userID string) (int, int, error)
	CountUrgentPostsAfter(channelID string, timestamp int64, userID string) (int, error)
	IncrementMentionCount(channelID string, userIDs []string, isRoot, isUrgent bool) error
	AnalyticsTypeCount(teamID string, channelType model.ChannelType) (int64, error)
	GetMembersForUser(teamID string, userID string) (model.ChannelMembers, error)
	GetTeamMembersForChannel(channelID string) ([]string, error)
	GetMembersForUserWithPagination(userID string, page, perPage int) (model.ChannelMembersWithTeamData, error)
	GetMembersForUserWithCursor(userID, teamID string, opts *ChannelMemberGraphQLSearchOpts) (model.ChannelMembers, error)
	Autocomplete(userID, term string, includeDeleted, isGuest bool) (model.ChannelListWithTeamData, error)
	AutocompleteInTeam(teamID, userID, term string, includeDeleted, isGuest bool) (model.ChannelList, error)
	AutocompleteInTeamForSearch(teamID string, userID string, term string, includeDeleted bool) (model.ChannelList, error)
	SearchAllChannels(term string, opts ChannelSearchOpts) (model.ChannelListWithTeamData, int64, error)
	SearchInTeam(teamID string, term string, includeDeleted bool) (model.ChannelList, error)
	SearchArchivedInTeam(teamID string, term string, userID string) (model.ChannelList, error)
	SearchForUserInTeam(userID string, teamID string, term string, includeDeleted bool) (model.ChannelList, error)
	SearchMore(userID string, teamID string, term string) (model.ChannelList, error)
	SearchGroupChannels(userID, term string) (model.ChannelList, error)
	GetMembersByIds(channelID string, userIds []string) (model.ChannelMembers, error)
	GetMembersByChannelIds(channelIds []string, userID string) (model.ChannelMembers, error)
	GetMembersInfoByChannelIds(channelIDs []string) (map[string][]*model.User, error)
	AnalyticsDeletedTypeCount(teamID string, channelType model.ChannelType) (int64, error)
	GetChannelUnread(channelID, userID string) (*model.ChannelUnread, error)
	ClearCaches()
	ClearMembersForUserCache()
	GetChannelsByScheme(schemeID string, offset int, limit int) (model.ChannelList, error)
	MigrateChannelMembers(fromChannelID string, fromUserID string) (map[string]string, error)
	ResetAllChannelSchemes() error
	ClearAllCustomRoleAssignments() error
	CreateInitialSidebarCategories(userID string, opts *SidebarCategorySearchOpts) (*model.OrderedSidebarCategories, error)
	GetSidebarCategoriesForTeamForUser(userID, teamID string) (*model.OrderedSidebarCategories, error)
	GetSidebarCategories(userID string, opts *SidebarCategorySearchOpts) (*model.OrderedSidebarCategories, error)
	GetSidebarCategory(categoryID string) (*model.SidebarCategoryWithChannels, error)
	GetSidebarCategoryOrder(userID, teamID string) ([]string, error)
	CreateSidebarCategory(userID, teamID string, newCategory *model.SidebarCategoryWithChannels) (*model.SidebarCategoryWithChannels, error)
	UpdateSidebarCategoryOrder(userID, teamID string, categoryOrder []string) error
	UpdateSidebarCategories(userID, teamID string, categories []*model.SidebarCategoryWithChannels) ([]*model.SidebarCategoryWithChannels, []*model.SidebarCategoryWithChannels, error)
	UpdateSidebarChannelsByPreferences(preferences model.Preferences) error
	DeleteSidebarChannelsByPreferences(preferences model.Preferences) error
	DeleteSidebarCategory(categoryID string) error
	GetAllChannelsForExportAfter(limit int, afterID string) ([]*model.ChannelForExport, error)
	GetAllDirectChannelsForExportAfter(limit int, afterID string) ([]*model.DirectChannelForExport, error)
	GetChannelMembersForExport(userID string, teamID string) ([]*model.ChannelMemberForExport, error)
	RemoveAllDeactivatedMembers(channelID string) error
	GetChannelsBatchForIndexing(startTime int64, startChannelID string, limit int) ([]*model.Channel, error)
	UserBelongsToChannels(userID string, channelIds []string) (bool, error)

	// UpdateMembersRole sets all of the given team members to admins and all of the other members of the team to
	// non-admin members.
	UpdateMembersRole(channelID string, userIDs []string) error

	// GroupSyncedChannelCount returns the count of non-deleted group-constrained channels.
	GroupSyncedChannelCount() (int64, error)

	SetShared(channelId string, shared bool) error
	// GetTeamForChannel returns the team for a given channelID.
	GetTeamForChannel(channelID string) (*model.Team, error)

	// Insights - channels
	GetTopChannelsForTeamSince(teamID string, userID string, since int64, offset int, limit int) (*model.TopChannelList, error)
	GetTopChannelsForUserSince(userID string, teamID string, since int64, offset int, limit int) (*model.TopChannelList, error)
	PostCountsByDuration(channelIDs []string, sinceUnixMillis int64, userID *string, duration model.PostCountGrouping, groupingLocation *time.Location) ([]*model.DurationPostCount, error)

	// Insights - inactive channels
	GetTopInactiveChannelsForTeamSince(teamID string, userID string, since int64, offset int, limit int) (*model.TopInactiveChannelList, error)
	GetTopInactiveChannelsForUserSince(teamID string, userID string, since int64, offset int, limit int) (*model.TopInactiveChannelList, error)
}

type ChannelMemberHistoryStore interface {
	LogJoinEvent(userID string, channelID string, joinTime int64) error
	LogLeaveEvent(userID string, channelID string, leaveTime int64) error
	GetUsersInChannelDuring(startTime int64, endTime int64, channelID string) ([]*model.ChannelMemberHistoryResult, error)
	PermanentDeleteBatchForRetentionPolicies(now, globalPolicyEndTime, limit int64, cursor model.RetentionPolicyCursor) (int64, model.RetentionPolicyCursor, error)
	DeleteOrphanedRows(limit int) (deleted int64, err error)
	PermanentDeleteBatch(endTime int64, limit int64) (int64, error)
	GetChannelsLeftSince(userID string, since int64) ([]string, error)
}
type ThreadStore interface {
	GetThreadFollowers(threadID string, fetchOnlyActive bool) ([]string, error)

	Get(id string) (*model.Thread, error)
	GetTotalUnreadThreads(userId, teamID string, opts model.GetUserThreadsOpts) (int64, error)
	GetTotalThreads(userId, teamID string, opts model.GetUserThreadsOpts) (int64, error)
	GetTotalUnreadMentions(userId, teamID string, opts model.GetUserThreadsOpts) (int64, error)
	GetTotalUnreadUrgentMentions(userId, teamID string, opts model.GetUserThreadsOpts) (int64, error)
	GetThreadsForUser(userId, teamID string, opts model.GetUserThreadsOpts) ([]*model.ThreadResponse, error)
	GetThreadForUser(threadMembership *model.ThreadMembership, extended, postPriorityIsEnabled bool) (*model.ThreadResponse, error)
	GetTeamsUnreadForUser(userID string, teamIDs []string, includeUrgentMentionCount bool) (map[string]*model.TeamUnread, error)

	MarkAllAsRead(userID string, threadIds []string) error
	MarkAllAsReadByTeam(userID, teamID string) error
	MarkAllAsReadByChannels(userID string, channelIDs []string) error
	MarkAsRead(userID, threadID string, timestamp int64) error

	UpdateMembership(membership *model.ThreadMembership) (*model.ThreadMembership, error)
	GetMembershipsForUser(userId, teamID string) ([]*model.ThreadMembership, error)
	GetMembershipForUser(userId, postID string) (*model.ThreadMembership, error)
	DeleteMembershipForUser(userId, postID string) error
	MaintainMembership(userID, postID string, opts ThreadMembershipOpts) (*model.ThreadMembership, error)
	PermanentDeleteBatchForRetentionPolicies(now, globalPolicyEndTime, limit int64, cursor model.RetentionPolicyCursor) (int64, model.RetentionPolicyCursor, error)
	PermanentDeleteBatchThreadMembershipsForRetentionPolicies(now, globalPolicyEndTime, limit int64, cursor model.RetentionPolicyCursor) (int64, model.RetentionPolicyCursor, error)
	DeleteOrphanedRows(limit int) (deleted int64, err error)
	GetThreadUnreadReplyCount(threadMembership *model.ThreadMembership) (int64, error)
	DeleteMembershipsForChannel(userID, channelID string) error

	// Insights - threads
	GetTopThreadsForTeamSince(teamID string, userID string, since int64, offset int, limit int) (*model.TopThreadList, error)
	GetTopThreadsForUserSince(teamID string, userID string, since int64, offset int, limit int) (*model.TopThreadList, error)
}

type PostStore interface {
	SaveMultiple(posts []*model.Post) ([]*model.Post, int, error)
	Save(post *model.Post) (*model.Post, error)
	Update(newPost *model.Post, oldPost *model.Post) (*model.Post, error)
	Get(ctx context.Context, id string, opts model.GetPostsOptions, userID string, sanitizeOptions map[string]bool) (*model.PostList, error)
	GetSingle(id string, inclDeleted bool) (*model.Post, error)
	Delete(postID string, timestamp int64, deleteByID string) error
	PermanentDeleteByUser(userID string) error
	PermanentDeleteByChannel(channelID string) error
	GetPosts(options model.GetPostsOptions, allowFromCache bool, sanitizeOptions map[string]bool) (*model.PostList, error)
	GetFlaggedPosts(userID string, offset int, limit int) (*model.PostList, error)
	// @openTracingParams userID, teamID, offset, limit
	GetFlaggedPostsForTeam(userID, teamID string, offset int, limit int) (*model.PostList, error)
	GetFlaggedPostsForChannel(userID, channelID string, offset int, limit int) (*model.PostList, error)
	GetPostsBefore(options model.GetPostsOptions, sanitizeOptions map[string]bool) (*model.PostList, error)
	GetPostsAfter(options model.GetPostsOptions, sanitizeOptions map[string]bool) (*model.PostList, error)
	GetPostsSince(options model.GetPostsSinceOptions, allowFromCache bool, sanitizeOptions map[string]bool) (*model.PostList, error)
	GetPostsByThread(threadID string, since int64) ([]*model.Post, error)
	GetPostAfterTime(channelID string, timestamp int64, collapsedThreads bool) (*model.Post, error)
	GetPostIdAfterTime(channelID string, timestamp int64, collapsedThreads bool) (string, error)
	GetPostIdBeforeTime(channelID string, timestamp int64, collapsedThreads bool) (string, error)
	GetEtag(channelID string, allowFromCache bool, collapsedThreads bool) string
	Search(teamID string, userID string, params *model.SearchParams) (*model.PostList, error)
	AnalyticsUserCountsWithPostsByDay(teamID string) (model.AnalyticsRows, error)
	AnalyticsPostCountsByDay(options *model.AnalyticsPostCountsOptions) (model.AnalyticsRows, error)
	AnalyticsPostCount(options *model.PostCountOptions) (int64, error)
	ClearCaches()
	InvalidateLastPostTimeCache(channelID string)
	GetLastPostRowCreateAt() (int64, error)
	GetPostsCreatedAt(channelID string, timestamp int64) ([]*model.Post, error)
	Overwrite(post *model.Post) (*model.Post, error)
	OverwriteMultiple(posts []*model.Post) ([]*model.Post, int, error)
	GetPostsByIds(postIds []string) ([]*model.Post, error)
	GetEditHistoryForPost(postId string) ([]*model.Post, error)
	GetPostsBatchForIndexing(startTime int64, startPostID string, limit int) ([]*model.PostForIndexing, error)
	PermanentDeleteBatchForRetentionPolicies(now, globalPolicyEndTime, limit int64, cursor model.RetentionPolicyCursor) (int64, model.RetentionPolicyCursor, error)
	DeleteOrphanedRows(limit int) (deleted int64, err error)
	PermanentDeleteBatch(endTime int64, limit int64) (int64, error)
	GetOldest() (*model.Post, error)
	GetMaxPostSize() int
	GetParentsForExportAfter(limit int, afterID string) ([]*model.PostForExport, error)
	GetRepliesForExport(parentID string) ([]*model.ReplyForExport, error)
	GetDirectPostParentsForExportAfter(limit int, afterID string) ([]*model.DirectPostForExport, error)
	SearchPostsForUser(paramsList []*model.SearchParams, userID, teamID string, page, perPage int) (*model.PostSearchResults, error)
	GetRecentSearchesForUser(userID string) ([]*model.SearchParams, error)
	LogRecentSearch(userID string, searchQuery []byte, createAt int64) error
	GetOldestEntityCreationTime() (int64, error)
	HasAutoResponsePostByUserSince(options model.GetPostsSinceOptions, userId string) (bool, error)
	GetPostsSinceForSync(options model.GetPostsSinceForSyncOptions, cursor model.GetPostsSinceForSyncCursor, limit int) ([]*model.Post, model.GetPostsSinceForSyncCursor, error)
	SetPostReminder(reminder *model.PostReminder) error
	GetPostReminders(now int64) ([]*model.PostReminder, error)
	GetPostReminderMetadata(postID string) (*PostReminderMetadata, error)
	// GetNthRecentPostTime returns the CreateAt time of the nth most recent post.
	GetNthRecentPostTime(n int64) (int64, error)

	// Insights - top DMs
	GetTopDMsForUserSince(userID string, since int64, offset int, limit int) (*model.TopDMList, error)
}

type UserStore interface {
	Save(user *model.User) (*model.User, error)
	Update(user *model.User, allowRoleUpdate bool) (*model.UserUpdate, error)
	UpdateNotifyProps(userID string, props map[string]string) error
	UpdateLastPictureUpdate(userID string) error
	ResetLastPictureUpdate(userID string) error
	UpdatePassword(userID, newPassword string) error
	UpdateUpdateAt(userID string) (int64, error)
	UpdateAuthData(userID string, service string, authData *string, email string, resetMfa bool) (string, error)
	ResetAuthDataToEmailForUsers(service string, userIDs []string, includeDeleted bool, dryRun bool) (int, error)
	UpdateMfaSecret(userID, secret string) error
	UpdateMfaActive(userID string, active bool) error
	Get(ctx context.Context, id string) (*model.User, error)
	GetMany(ctx context.Context, ids []string) ([]*model.User, error)
	GetAll() ([]*model.User, error)
	ClearCaches()
	InvalidateProfilesInChannelCacheByUser(userID string)
	InvalidateProfilesInChannelCache(channelID string)
	GetProfilesInChannel(options *model.UserGetOptions) ([]*model.User, error)
	GetProfilesInChannelByStatus(options *model.UserGetOptions) ([]*model.User, error)
	GetProfilesInChannelByAdmin(options *model.UserGetOptions) ([]*model.User, error)
	GetAllProfilesInChannel(ctx context.Context, channelID string, allowFromCache bool) (map[string]*model.User, error)
	GetProfilesNotInChannel(teamID string, channelId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetProfilesWithoutTeam(options *model.UserGetOptions) ([]*model.User, error)
	GetProfilesByUsernames(usernames []string, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetAllProfiles(options *model.UserGetOptions) ([]*model.User, error)
	GetProfiles(options *model.UserGetOptions) ([]*model.User, error)
	GetProfileByIds(ctx context.Context, userIds []string, options *UserGetByIdsOpts, allowFromCache bool) ([]*model.User, error)
	GetProfileByGroupChannelIdsForUser(userID string, channelIds []string) (map[string][]*model.User, error)
	InvalidateProfileCacheForUser(userID string)
	GetByEmail(email string) (*model.User, error)
	GetByAuth(authData *string, authService string) (*model.User, error)
	GetAllUsingAuthService(authService string) ([]*model.User, error)
	GetAllNotInAuthService(authServices []string) ([]*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetForLogin(loginID string, allowSignInWithUsername, allowSignInWithEmail bool) (*model.User, error)
	VerifyEmail(userID, email string) (string, error)
	GetEtagForAllProfiles() string
	GetEtagForProfiles(teamID string) string
	UpdateFailedPasswordAttempts(userID string, attempts int) error
	GetSystemAdminProfiles() (map[string]*model.User, error)
	PermanentDelete(userID string) error
	AnalyticsActiveCount(timestamp int64, options model.UserCountOptions) (int64, error)
	AnalyticsActiveCountForPeriod(startTime int64, endTime int64, options model.UserCountOptions) (int64, error)
	GetUnreadCount(userID string, isCRTEnabled bool) (int64, error)
	GetUnreadCountForChannel(userID string, channelID string) (int64, error)
	GetAnyUnreadPostCountForChannel(userID string, channelID string) (int64, error)
	GetRecentlyActiveUsersForTeam(teamID string, offset, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetNewUsersForTeam(teamID string, offset, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	Search(teamID string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchNotInTeam(notInTeamID string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchInChannel(channelID string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchNotInChannel(teamID string, channelID string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchWithoutTeam(term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchInGroup(groupID string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchNotInGroup(groupID string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	AnalyticsGetInactiveUsersCount() (int64, error)
	AnalyticsGetExternalUsers(hostDomain string) (bool, error)
	AnalyticsGetSystemAdminCount() (int64, error)
	AnalyticsGetGuestCount() (int64, error)
	GetProfilesNotInTeam(teamID string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetEtagForProfilesNotInTeam(teamID string) string
	ClearAllCustomRoleAssignments() error
	InferSystemInstallDate() (int64, error)
	GetAllAfter(limit int, afterID string) ([]*model.User, error)
	GetUsersBatchForIndexing(startTime int64, startFileID string, limit int) ([]*model.UserForIndexing, error)
	Count(options model.UserCountOptions) (int64, error)
	GetTeamGroupUsers(teamID string) ([]*model.User, error)
	GetChannelGroupUsers(channelID string) ([]*model.User, error)
	PromoteGuestToUser(userID string) error
	DemoteUserToGuest(userID string) (*model.User, error)
	DeactivateGuests() ([]string, error)
	AutocompleteUsersInChannel(teamID, channelID, term string, options *model.UserSearchOptions) (*model.UserAutocompleteInChannel, error)
	GetKnownUsers(userID string) ([]string, error)
	IsEmpty(excludeBots bool) (bool, error)
	GetUsersWithInvalidEmails(page int, perPage int, restrictedDomains string) ([]*model.User, error)
	InsertUsers(users []*model.User) error
	GetFirstSystemAdminID() (string, error)
}

type BotStore interface {
	Get(userID string, includeDeleted bool) (*model.Bot, error)
	GetAll(options *model.BotGetOptions) ([]*model.Bot, error)
	Save(bot *model.Bot) (*model.Bot, error)
	Update(bot *model.Bot) (*model.Bot, error)
	PermanentDelete(userID string) error
}

type SessionStore interface {
	Get(ctx context.Context, sessionIDOrToken string) (*model.Session, error)
	Save(session *model.Session) (*model.Session, error)
	GetSessions(userID string) ([]*model.Session, error)
	GetSessionsWithActiveDeviceIds(userID string) ([]*model.Session, error)
	GetSessionsExpired(thresholdMillis int64, mobileOnly bool, unnotifiedOnly bool) ([]*model.Session, error)
	UpdateExpiredNotify(sessionid string, notified bool) error
	Remove(sessionIDOrToken string) error
	RemoveAllSessions() error
	PermanentDeleteSessionsByUser(teamID string) error
	GetLastSessionRowCreateAt() (int64, error)
	UpdateExpiresAt(sessionID string, timestamp int64) error
	UpdateLastActivityAt(sessionID string, timestamp int64) error
	UpdateRoles(userID string, roles string) (string, error)
	UpdateDeviceId(id string, deviceID string, expiresAt int64) (string, error)
	UpdateProps(session *model.Session) error
	AnalyticsSessionCount() (int64, error)
	Cleanup(expiryTime int64, batchSize int64) error
}

type AuditStore interface {
	Save(audit *model.Audit) error
	Get(user_id string, offset int, limit int) (model.Audits, error)
	PermanentDeleteByUser(userID string) error
}

type ClusterDiscoveryStore interface {
	Save(discovery *model.ClusterDiscovery) error
	Delete(discovery *model.ClusterDiscovery) (bool, error)
	Exists(discovery *model.ClusterDiscovery) (bool, error)
	GetAll(discoveryType, clusterName string) ([]*model.ClusterDiscovery, error)
	SetLastPingAt(discovery *model.ClusterDiscovery) error
	Cleanup() error
}

type RemoteClusterStore interface {
	Save(rc *model.RemoteCluster) (*model.RemoteCluster, error)
	Update(rc *model.RemoteCluster) (*model.RemoteCluster, error)
	Delete(remoteClusterId string) (bool, error)
	Get(remoteClusterId string) (*model.RemoteCluster, error)
	GetAll(filter model.RemoteClusterQueryFilter) ([]*model.RemoteCluster, error)
	UpdateTopics(remoteClusterId string, topics string) (*model.RemoteCluster, error)
	SetLastPingAt(remoteClusterId string) error
}

type ComplianceStore interface {
	Save(compliance *model.Compliance) (*model.Compliance, error)
	Update(compliance *model.Compliance) (*model.Compliance, error)
	Get(id string) (*model.Compliance, error)
	GetAll(offset, limit int) (model.Compliances, error)
	ComplianceExport(compliance *model.Compliance, cursor model.ComplianceExportCursor, limit int) ([]*model.CompliancePost, model.ComplianceExportCursor, error)
	MessageExport(ctx context.Context, cursor model.MessageExportCursor, limit int) ([]*model.MessageExport, model.MessageExportCursor, error)
}

type OAuthStore interface {
	SaveApp(app *model.OAuthApp) (*model.OAuthApp, error)
	UpdateApp(app *model.OAuthApp) (*model.OAuthApp, error)
	GetApp(id string) (*model.OAuthApp, error)
	GetAppByUser(userID string, offset, limit int) ([]*model.OAuthApp, error)
	GetApps(offset, limit int) ([]*model.OAuthApp, error)
	GetAuthorizedApps(userID string, offset, limit int) ([]*model.OAuthApp, error)
	DeleteApp(id string) error
	SaveAuthData(authData *model.AuthData) (*model.AuthData, error)
	GetAuthData(code string) (*model.AuthData, error)
	RemoveAuthData(code string) error
	RemoveAuthDataByClientId(clientId string, userId string) error
	RemoveAuthDataByUserId(userId string) error
	PermanentDeleteAuthDataByUser(userID string) error
	SaveAccessData(accessData *model.AccessData) (*model.AccessData, error)
	UpdateAccessData(accessData *model.AccessData) (*model.AccessData, error)
	GetAccessData(token string) (*model.AccessData, error)
	GetAccessDataByUserForApp(userID, clientId string) ([]*model.AccessData, error)
	GetAccessDataByRefreshToken(token string) (*model.AccessData, error)
	GetPreviousAccessData(userID, clientId string) (*model.AccessData, error)
	RemoveAccessData(token string) error
	RemoveAllAccessData() error
}

type SystemStore interface {
	Save(system *model.System) error
	SaveOrUpdate(system *model.System) error
	Update(system *model.System) error
	Get() (model.StringMap, error)
	GetByName(name string) (*model.System, error)
	PermanentDeleteByName(name string) (*model.System, error)
	InsertIfExists(system *model.System) (*model.System, error)
	SaveOrUpdateWithWarnMetricHandling(system *model.System) error
}

type WebhookStore interface {
	SaveIncoming(webhook *model.IncomingWebhook) (*model.IncomingWebhook, error)
	GetIncoming(id string, allowFromCache bool) (*model.IncomingWebhook, error)
	GetIncomingList(offset, limit int) ([]*model.IncomingWebhook, error)
	GetIncomingListByUser(userID string, offset, limit int) ([]*model.IncomingWebhook, error)
	GetIncomingByTeam(teamID string, offset, limit int) ([]*model.IncomingWebhook, error)
	GetIncomingByTeamByUser(teamID string, userID string, offset, limit int) ([]*model.IncomingWebhook, error)
	UpdateIncoming(webhook *model.IncomingWebhook) (*model.IncomingWebhook, error)
	GetIncomingByChannel(channelID string) ([]*model.IncomingWebhook, error)
	DeleteIncoming(webhookID string, timestamp int64) error
	PermanentDeleteIncomingByChannel(channelID string) error
	PermanentDeleteIncomingByUser(userID string) error

	SaveOutgoing(webhook *model.OutgoingWebhook) (*model.OutgoingWebhook, error)
	GetOutgoing(id string) (*model.OutgoingWebhook, error)
	GetOutgoingByChannel(channelID string, offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingByChannelByUser(channelID string, userID string, offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingList(offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingListByUser(userID string, offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingByTeam(teamID string, offset, limit int) ([]*model.OutgoingWebhook, error)
	GetOutgoingByTeamByUser(teamID string, userID string, offset, limit int) ([]*model.OutgoingWebhook, error)
	DeleteOutgoing(webhookID string, timestamp int64) error
	PermanentDeleteOutgoingByChannel(channelID string) error
	PermanentDeleteOutgoingByUser(userID string) error
	UpdateOutgoing(hook *model.OutgoingWebhook) (*model.OutgoingWebhook, error)

	AnalyticsIncomingCount(teamID string) (int64, error)
	AnalyticsOutgoingCount(teamID string) (int64, error)
	InvalidateWebhookCache(webhook string)
	ClearCaches()
}

type CommandStore interface {
	Save(webhook *model.Command) (*model.Command, error)
	GetByTrigger(teamID string, trigger string) (*model.Command, error)
	Get(id string) (*model.Command, error)
	GetByTeam(teamID string) ([]*model.Command, error)
	Delete(commandID string, timestamp int64) error
	PermanentDeleteByTeam(teamID string) error
	PermanentDeleteByUser(userID string) error
	Update(hook *model.Command) (*model.Command, error)
	AnalyticsCommandCount(teamID string) (int64, error)
}

type CommandWebhookStore interface {
	Save(webhook *model.CommandWebhook) (*model.CommandWebhook, error)
	Get(id string) (*model.CommandWebhook, error)
	TryUse(id string, limit int) error
	Cleanup()
}

type PreferenceStore interface {
	Save(preferences model.Preferences) error
	GetCategory(userID string, category string) (model.Preferences, error)
	GetCategoryAndName(category string, nane string) (model.Preferences, error)
	Get(userID string, category string, name string) (*model.Preference, error)
	GetAll(userID string) (model.Preferences, error)
	Delete(userID, category, name string) error
	DeleteCategory(userID string, category string) error
	DeleteCategoryAndName(category string, name string) error
	PermanentDeleteByUser(userID string) error
	DeleteOrphanedRows(limit int) (deleted int64, err error)
	CleanupFlagsBatch(limit int64) (int64, error)
}

type LicenseStore interface {
	Save(license *model.LicenseRecord) (*model.LicenseRecord, error)
	Get(id string) (*model.LicenseRecord, error)
	GetAll() ([]*model.LicenseRecord, error)
}

type TokenStore interface {
	Save(recovery *model.Token) error
	Delete(token string) error
	GetByToken(token string) (*model.Token, error)
	Cleanup(expiryTime int64)
	GetAllTokensByType(tokenType string) ([]*model.Token, error)
	RemoveAllTokensByType(tokenType string) error
}

type EmojiStore interface {
	Save(emoji *model.Emoji) (*model.Emoji, error)
	Get(ctx context.Context, id string, allowFromCache bool) (*model.Emoji, error)
	GetByName(ctx context.Context, name string, allowFromCache bool) (*model.Emoji, error)
	GetMultipleByName(names []string) ([]*model.Emoji, error)
	GetList(offset, limit int, sort string) ([]*model.Emoji, error)
	Delete(emoji *model.Emoji, timestamp int64) error
	Search(name string, prefixOnly bool, limit int) ([]*model.Emoji, error)
}

type StatusStore interface {
	SaveOrUpdate(status *model.Status) error
	Get(userID string) (*model.Status, error)
	GetByIds(userIds []string) ([]*model.Status, error)
	ResetAll() error
	GetTotalActiveUsersCount() (int64, error)
	UpdateLastActivityAt(userID string, lastActivityAt int64) error
	UpdateExpiredDNDStatuses() ([]*model.Status, error)
}

type FileInfoStore interface {
	Save(info *model.FileInfo) (*model.FileInfo, error)
	Upsert(info *model.FileInfo) (*model.FileInfo, error)
	Get(id string) (*model.FileInfo, error)
	GetFromMaster(id string) (*model.FileInfo, error)
	GetByIds(ids []string) ([]*model.FileInfo, error)
	GetByPath(path string) (*model.FileInfo, error)
	GetForPost(postID string, readFromMaster, includeDeleted, allowFromCache bool) ([]*model.FileInfo, error)
	GetForUser(userID string) ([]*model.FileInfo, error)
	GetWithOptions(page, perPage int, opt *model.GetFileInfosOptions) ([]*model.FileInfo, error)
	InvalidateFileInfosForPostCache(postID string, deleted bool)
	AttachToPost(fileID string, postID string, creatorID string) error
	DeleteForPost(postID string) (string, error)
	PermanentDelete(fileID string) error
	PermanentDeleteBatch(endTime int64, limit int64) (int64, error)
	PermanentDeleteByUser(userID string) (int64, error)
	SetContent(fileID, content string) error
	Search(paramsList []*model.SearchParams, userID, teamID string, page, perPage int) (*model.FileInfoList, error)
	CountAll() (int64, error)
	GetFilesBatchForIndexing(startTime int64, startFileID string, limit int) ([]*model.FileForIndexing, error)
	ClearCaches()
	GetStorageUsage(allowFromCache, includeDeleted bool) (int64, error)
	// GetUptoNSizeFileTime returns the CreateAt time of the last accessible file with a running-total size upto n bytes.
	GetUptoNSizeFileTime(n int64) (int64, error)
}

type UploadSessionStore interface {
	Save(session *model.UploadSession) (*model.UploadSession, error)
	Update(session *model.UploadSession) error
	Get(ctx context.Context, id string) (*model.UploadSession, error)
	GetForUser(userID string) ([]*model.UploadSession, error)
	Delete(id string) error
}

type ReactionStore interface {
	Save(reaction *model.Reaction) (*model.Reaction, error)
	Delete(reaction *model.Reaction) (*model.Reaction, error)
	GetForPost(postID string, allowFromCache bool) ([]*model.Reaction, error)
	GetForPostSince(postId string, since int64, excludeRemoteId string, inclDeleted bool) ([]*model.Reaction, error)
	DeleteAllWithEmojiName(emojiName string) error
	BulkGetForPosts(postIds []string) ([]*model.Reaction, error)
	DeleteOrphanedRows(limit int) (int64, error)
	PermanentDeleteBatch(endTime int64, limit int64) (int64, error)
	GetTopForTeamSince(teamID string, userID string, since int64, offset int, limit int) (*model.TopReactionList, error)
	GetTopForUserSince(userID string, teamID string, since int64, offset int, limit int) (*model.TopReactionList, error)
}

type JobStore interface {
	Save(job *model.Job) (*model.Job, error)
	UpdateOptimistically(job *model.Job, currentStatus string) (bool, error)
	UpdateStatus(id string, status string) (*model.Job, error)
	UpdateStatusOptimistically(id string, currentStatus string, newStatus string) (bool, error)
	Get(id string) (*model.Job, error)
	GetAllPage(offset int, limit int) ([]*model.Job, error)
	GetAllByType(jobType string) ([]*model.Job, error)
	GetAllByTypeAndStatus(jobType string, status string) ([]*model.Job, error)
	GetAllByTypePage(jobType string, offset int, limit int) ([]*model.Job, error)
	GetAllByTypesPage(jobTypes []string, offset int, limit int) ([]*model.Job, error)
	GetAllByStatus(status string) ([]*model.Job, error)
	GetNewestJobByStatusAndType(status string, jobType string) (*model.Job, error)
	GetNewestJobByStatusesAndType(statuses []string, jobType string) (*model.Job, error)
	GetCountByStatusAndType(status string, jobType string) (int64, error)
	Delete(id string) (string, error)
	Cleanup(expiryTime int64, batchSize int) error
}

type UserAccessTokenStore interface {
	Save(token *model.UserAccessToken) (*model.UserAccessToken, error)
	DeleteAllForUser(userID string) error
	Delete(tokenID string) error
	Get(tokenID string) (*model.UserAccessToken, error)
	GetAll(offset int, limit int) ([]*model.UserAccessToken, error)
	GetByToken(tokenString string) (*model.UserAccessToken, error)
	GetByUser(userID string, page, perPage int) ([]*model.UserAccessToken, error)
	Search(term string) ([]*model.UserAccessToken, error)
	UpdateTokenEnable(tokenID string) error
	UpdateTokenDisable(tokenID string) error
}

type PluginStore interface {
	SaveOrUpdate(keyVal *model.PluginKeyValue) (*model.PluginKeyValue, error)
	CompareAndSet(keyVal *model.PluginKeyValue, oldValue []byte) (bool, error)
	CompareAndDelete(keyVal *model.PluginKeyValue, oldValue []byte) (bool, error)
	SetWithOptions(pluginID string, key string, value []byte, options model.PluginKVSetOptions) (bool, error)
	Get(pluginID, key string) (*model.PluginKeyValue, error)
	Delete(pluginID, key string) error
	DeleteAllForPlugin(PluginID string) error
	DeleteAllExpired() error
	List(pluginID string, page, perPage int) ([]string, error)
}

type RoleStore interface {
	Save(role *model.Role) (*model.Role, error)
	Get(roleID string) (*model.Role, error)
	GetAll() ([]*model.Role, error)
	GetByName(ctx context.Context, name string) (*model.Role, error)
	GetByNames(names []string) ([]*model.Role, error)
	Delete(roleID string) (*model.Role, error)
	PermanentDeleteAll() error

	// HigherScopedPermissions retrieves the higher-scoped permissions of a list of role names. The higher-scope
	// (either team scheme or system scheme) is determined based on whether the team has a scheme or not.
	ChannelHigherScopedPermissions(roleNames []string) (map[string]*model.RolePermissions, error)

	// AllChannelSchemeRoles returns all of the roles associated to channel schemes.
	AllChannelSchemeRoles() ([]*model.Role, error)

	// ChannelRolesUnderTeamRole returns all of the non-deleted roles that are affected by updates to the
	// given role.
	ChannelRolesUnderTeamRole(roleName string) ([]*model.Role, error)
}

type SchemeStore interface {
	Save(scheme *model.Scheme) (*model.Scheme, error)
	Get(schemeID string) (*model.Scheme, error)
	GetByName(schemeName string) (*model.Scheme, error)
	GetAllPage(scope string, offset int, limit int) ([]*model.Scheme, error)
	Delete(schemeID string) (*model.Scheme, error)
	PermanentDeleteAll() error
	CountByScope(scope string) (int64, error)
	CountWithoutPermission(scope, permissionID string, roleScope model.RoleScope, roleType model.RoleType) (int64, error)
}

type TermsOfServiceStore interface {
	Save(termsOfService *model.TermsOfService) (*model.TermsOfService, error)
	GetLatest(allowFromCache bool) (*model.TermsOfService, error)
	Get(id string, allowFromCache bool) (*model.TermsOfService, error)
}

type ProductNoticesStore interface {
	View(userID string, notices []string) error
	Clear(notices []string) error
	ClearOldNotices(currentNotices model.ProductNotices) error
	GetViews(userID string) ([]model.ProductNoticeViewState, error)
}

type UserTermsOfServiceStore interface {
	GetByUser(userID string) (*model.UserTermsOfService, error)
	Save(userTermsOfService *model.UserTermsOfService) (*model.UserTermsOfService, error)
	Delete(userID, termsOfServiceId string) error
}

type GroupStore interface {
	Create(group *model.Group) (*model.Group, error)
	CreateWithUserIds(group *model.GroupWithUserIds) (*model.Group, error)
	Get(groupID string) (*model.Group, error)
	GetByName(name string, opts model.GroupSearchOpts) (*model.Group, error)
	GetByIDs(groupIDs []string) ([]*model.Group, error)
	GetByRemoteID(remoteID string, groupSource model.GroupSource) (*model.Group, error)
	GetAllBySource(groupSource model.GroupSource) ([]*model.Group, error)
	GetByUser(userID string) ([]*model.Group, error)
	Update(group *model.Group) (*model.Group, error)
	Delete(groupID string) (*model.Group, error)
	Restore(groupID string) (*model.Group, error)

	GetMemberUsers(groupID string) ([]*model.User, error)
	GetMemberUsersPage(groupID string, page int, perPage int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetMemberUsersSortedPage(groupID string, page int, perPage int, viewRestrictions *model.ViewUsersRestrictions, teammateNameDisplay string) ([]*model.User, error)
	GetMemberCountWithRestrictions(groupID string, viewRestrictions *model.ViewUsersRestrictions) (int64, error)
	GetMemberCount(groupID string) (int64, error)

	GetNonMemberUsersPage(groupID string, page int, perPage int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)

	GetMemberUsersInTeam(groupID string, teamID string) ([]*model.User, error)
	GetMemberUsersNotInChannel(groupID string, channelID string) ([]*model.User, error)

	UpsertMember(groupID string, userID string) (*model.GroupMember, error)
	DeleteMember(groupID string, userID string) (*model.GroupMember, error)
	PermanentDeleteMembersByUser(userID string) error

	CreateGroupSyncable(groupSyncable *model.GroupSyncable) (*model.GroupSyncable, error)
	GetGroupSyncable(groupID string, syncableID string, syncableType model.GroupSyncableType) (*model.GroupSyncable, error)
	GetAllGroupSyncablesByGroupId(groupID string, syncableType model.GroupSyncableType) ([]*model.GroupSyncable, error)
	UpdateGroupSyncable(groupSyncable *model.GroupSyncable) (*model.GroupSyncable, error)
	DeleteGroupSyncable(groupID string, syncableID string, syncableType model.GroupSyncableType) (*model.GroupSyncable, error)

	// TeamMembersToAdd returns a slice of UserTeamIDPair that need newly created memberships
	// based on the groups configurations. The returned list can be optionally scoped to a single given team.
	//
	// Typically since will be the last successful group sync time.
	// If includeRemovedMembers is true, then team members who left or were removed from the team will
	// be included; otherwise, they will be excluded.
	TeamMembersToAdd(since int64, teamID *string, includeRemovedMembers bool) ([]*model.UserTeamIDPair, error)

	// ChannelMembersToAdd returns a slice of UserChannelIDPair that need newly created memberships
	// based on the groups configurations. The returned list can be optionally scoped to a single given channel.
	//
	// Typically since will be the last successful group sync time.
	// If includeRemovedMembers is true, then channel members who left or were removed from the channel will
	// be included; otherwise, they will be excluded.
	ChannelMembersToAdd(since int64, channelID *string, includeRemovedMembers bool) ([]*model.UserChannelIDPair, error)

	// TeamMembersToRemove returns all team members that should be removed based on group constraints.
	TeamMembersToRemove(teamID *string) ([]*model.TeamMember, error)

	// ChannelMembersToRemove returns all channel members that should be removed based on group constraints.
	ChannelMembersToRemove(channelID *string) ([]*model.ChannelMember, error)

	GetGroupsByChannel(channelID string, opts model.GroupSearchOpts) ([]*model.GroupWithSchemeAdmin, error)
	CountGroupsByChannel(channelID string, opts model.GroupSearchOpts) (int64, error)

	GetGroupsByTeam(teamID string, opts model.GroupSearchOpts) ([]*model.GroupWithSchemeAdmin, error)
	GetGroupsAssociatedToChannelsByTeam(teamID string, opts model.GroupSearchOpts) (map[string][]*model.GroupWithSchemeAdmin, error)
	CountGroupsByTeam(teamID string, opts model.GroupSearchOpts) (int64, error)

	GetGroups(page, perPage int, opts model.GroupSearchOpts, viewRestrictions *model.ViewUsersRestrictions) ([]*model.Group, error)

	TeamMembersMinusGroupMembers(teamID string, groupIDs []string, page, perPage int) ([]*model.UserWithGroups, error)
	CountTeamMembersMinusGroupMembers(teamID string, groupIDs []string) (int64, error)
	ChannelMembersMinusGroupMembers(channelID string, groupIDs []string, page, perPage int) ([]*model.UserWithGroups, error)
	CountChannelMembersMinusGroupMembers(channelID string, groupIDs []string) (int64, error)

	// AdminRoleGroupsForSyncableMember returns the IDs of all of the groups that the user is a member of that are
	// configured as SchemeAdmin: true for the given syncable.
	AdminRoleGroupsForSyncableMember(userID, syncableID string, syncableType model.GroupSyncableType) ([]string, error)

	// PermittedSyncableAdmins returns the IDs of all of the user who are permitted by the group syncable to have
	// the admin role for the given syncable.
	PermittedSyncableAdmins(syncableID string, syncableType model.GroupSyncableType) ([]string, error)

	// GroupCount returns the total count of records in the UserGroups table.
	GroupCount() (int64, error)

	GroupCountBySource(source model.GroupSource) (int64, error)

	// GroupTeamCount returns the total count of records in the GroupTeams table.
	GroupTeamCount() (int64, error)

	// GroupChannelCount returns the total count of records in the GroupChannels table.
	GroupChannelCount() (int64, error)

	// GroupMemberCount returns the total count of records in the GroupMembers table.
	GroupMemberCount() (int64, error)

	// DistinctGroupMemberCount returns the count of records in the GroupMembers table with distinct userID values.
	DistinctGroupMemberCount() (int64, error)

	DistinctGroupMemberCountForSource(source model.GroupSource) (int64, error)

	// GroupCountWithAllowReference returns the count of records in the Groups table with AllowReference set to true.
	GroupCountWithAllowReference() (int64, error)

	UpsertMembers(groupID string, userIDs []string) ([]*model.GroupMember, error)
	DeleteMembers(groupID string, userIDs []string) ([]*model.GroupMember, error)

	GetMember(groupID string, userID string) (*model.GroupMember, error)
}

type LinkMetadataStore interface {
	Save(linkMetadata *model.LinkMetadata) (*model.LinkMetadata, error)
	Get(url string, timestamp int64) (*model.LinkMetadata, error)
}

type NotifyAdminStore interface {
	Save(data *model.NotifyAdminData) (*model.NotifyAdminData, error)
	GetDataByUserIdAndFeature(userId string, feature model.MattermostFeature) ([]*model.NotifyAdminData, error)
	Get(trial bool) ([]*model.NotifyAdminData, error)
	DeleteBefore(trial bool, now int64) error
	Update(userId string, requiredPlan string, requiredFeature model.MattermostFeature, now int64) error
}

type SharedChannelStore interface {
	Save(sc *model.SharedChannel) (*model.SharedChannel, error)
	Get(channelId string) (*model.SharedChannel, error)
	HasChannel(channelID string) (bool, error)
	GetAll(offset, limit int, opts model.SharedChannelFilterOpts) ([]*model.SharedChannel, error)
	GetAllCount(opts model.SharedChannelFilterOpts) (int64, error)
	Update(sc *model.SharedChannel) (*model.SharedChannel, error)
	Delete(channelId string) (bool, error)

	SaveRemote(remote *model.SharedChannelRemote) (*model.SharedChannelRemote, error)
	UpdateRemote(remote *model.SharedChannelRemote) (*model.SharedChannelRemote, error)
	GetRemote(id string) (*model.SharedChannelRemote, error)
	HasRemote(channelID string, remoteId string) (bool, error)
	GetRemoteForUser(remoteId string, userId string) (*model.RemoteCluster, error)
	GetRemoteByIds(channelId string, remoteId string) (*model.SharedChannelRemote, error)
	GetRemotes(opts model.SharedChannelRemoteFilterOpts) ([]*model.SharedChannelRemote, error)
	UpdateRemoteCursor(id string, cursor model.GetPostsSinceForSyncCursor) error
	DeleteRemote(remoteId string) (bool, error)
	GetRemotesStatus(channelId string) ([]*model.SharedChannelRemoteStatus, error)

	SaveUser(remote *model.SharedChannelUser) (*model.SharedChannelUser, error)
	GetSingleUser(userID string, channelID string, remoteID string) (*model.SharedChannelUser, error)
	GetUsersForUser(userID string) ([]*model.SharedChannelUser, error)
	GetUsersForSync(filter model.GetUsersForSyncFilter) ([]*model.User, error)
	UpdateUserLastSyncAt(userID string, channelID string, remoteID string) error

	SaveAttachment(remote *model.SharedChannelAttachment) (*model.SharedChannelAttachment, error)
	UpsertAttachment(remote *model.SharedChannelAttachment) (string, error)
	GetAttachment(fileId string, remoteId string) (*model.SharedChannelAttachment, error)
	UpdateAttachmentLastSyncAt(id string, syncTime int64) error
}

type PostPriorityStore interface {
	GetForPost(postId string) (*model.PostPriority, error)
	GetForPosts(ids []string) ([]*model.PostPriority, error)
}

type DraftStore interface {
	Save(d *model.Draft) (*model.Draft, error)
	Get(userID, channelID, rootID string, includeDeleted bool) (*model.Draft, error)
	Delete(userID, channelID, rootID string) error
	GetDraftsForUser(userID, teamID string) ([]*model.Draft, error)
	Update(d *model.Draft) (*model.Draft, error)
}

type PostAcknowledgementStore interface {
	Get(postID, userID string) (*model.PostAcknowledgement, error)
	GetForPost(postID string) ([]*model.PostAcknowledgement, error)
	GetForPosts(postIds []string) ([]*model.PostAcknowledgement, error)
	Save(postID, userID string, acknowledgedAt int64) (*model.PostAcknowledgement, error)
	Delete(acknowledgement *model.PostAcknowledgement) error
}

type TrueUpReviewStore interface {
	GetTrueUpReviewStatus(dueDate int64) (*model.TrueUpReviewStatus, error)
	CreateTrueUpReviewStatusRecord(reviewStatus *model.TrueUpReviewStatus) (*model.TrueUpReviewStatus, error)
	Update(reviewStatus *model.TrueUpReviewStatus) (*model.TrueUpReviewStatus, error)
}

// ChannelSearchOpts contains options for searching channels.
//
// NotAssociatedToGroup will exclude channels that have associated, active GroupChannels records.
// IncludeDeleted will include channel records where DeleteAt != 0.
// ExcludeChannelNames will exclude channels from the results by name.
// IncludeSearchById will include searching matches against channel IDs in the results
// Paginate whether to paginate the results.
// Page page requested, if results are paginated.
// PerPage number of results per page, if paginated.
type ChannelSearchOpts struct {
	Term                     string
	NotAssociatedToGroup     string
	IncludeDeleted           bool
	Deleted                  bool
	ExcludeChannelNames      []string
	TeamIds                  []string
	GroupConstrained         bool
	ExcludeGroupConstrained  bool
	PolicyID                 string
	ExcludePolicyConstrained bool
	IncludePolicyID          bool
	IncludeTeamInfo          bool
	IncludeSearchById        bool
	CountOnly                bool
	Public                   bool
	Private                  bool
	Page                     *int
	PerPage                  *int
	LastDeleteAt             int
	LastUpdateAt             int
}

func (c *ChannelSearchOpts) IsPaginated() bool {
	return c.Page != nil && c.PerPage != nil
}

type UserGetByIdsOpts struct {
	// IsAdmin tracks whether or not the request is being made by an administrator. Does nothing when provided by a client.
	IsAdmin bool

	// Restrict to search in a list of teams and channels. Does nothing when provided by a client.
	ViewRestrictions *model.ViewUsersRestrictions

	// Since filters the users based on their UpdateAt timestamp.
	Since int64
}

// ThreadMembershipOpts defines some properties to be passed to
// ThreadStore.MaintainMembership()
type ThreadMembershipOpts struct {
	// Following indicates whether or not the user is following the thread.
	Following bool
	// IncrementMentions indicates whether or not the mentions count for
	// the thread should be incremented.
	IncrementMentions bool
	// UpdateFollowing indicates whether or not the following state should be changed.
	UpdateFollowing bool
	// UpdateViewedTimestamp indicates whether or not the LastViewed field of the
	// membership should be updated.
	UpdateViewedTimestamp bool
	// UpdateParticipants indicates whether or not the thread's participants list
	// should be updated.
	UpdateParticipants bool
}

// ChannelMemberGraphQLSearchOpts contains the options for a graphQL query
// to get the channel members.
type ChannelMemberGraphQLSearchOpts struct {
	AfterChannel string
	AfterUser    string
	Limit        int
	LastUpdateAt int
	ExcludeTeam  bool
}

// PostReminderMetadata contains some info needed to send
// the reminder message to the user.
type PostReminderMetadata struct {
	ChannelId  string
	TeamName   string
	UserLocale string
	Username   string
}

// SidebarCategorySearchOpts contains the options for a graphQL query
// to get the sidebar categories.
type SidebarCategorySearchOpts struct {
	TeamID      string
	ExcludeTeam bool
}

// Ensure store service adapter implements `product.StoreService`
var _ product.StoreService = (*StoreServiceAdapter)(nil)

// StoreServiceAdapter provides a simple Store wrapper for use with products.
type StoreServiceAdapter struct {
	store Store
}

func NewStoreServiceAdapter(store Store) *StoreServiceAdapter {
	return &StoreServiceAdapter{
		store: store,
	}
}

func (a *StoreServiceAdapter) GetMasterDB() *sql.DB {
	return a.store.GetInternalMasterDB()
}
