package slack

// TopicPurpose represents the topic or purpose metadata of a channel.
type TopicPurpose struct {
	Value   string
	Creator string
	LastSet int64
}

// Channel represents a Slack conversation (channel, group, IM, MPIM).
type Channel struct {
	ID              string
	Name            string
	DisplayName     string
	User            string
	IsChannel       bool
	IsGroup         bool
	IsIM            bool
	IsMPIM          bool
	IsPrivate       bool
	Created         int64
	IsArchived      bool
	IsGeneral       bool
	Unlinked        int
	NameNormalized  string
	IsShared        bool
	IsExtShared     bool
	IsOrgShared     bool
	IsMember        bool
	NumMembers      int
	UnreadCount     int
	UnreadCountDisp int
	LastRead        string
	Topic           *TopicPurpose
	Purpose         *TopicPurpose
}

// ChannelDetail is a subset of channel info returned by
// conversations.info with include_num_members=true.
type ChannelDetail struct {
	ID         string
	Name       string
	IsPrivate  bool
	IsArchived bool
	Created    int64
	NumMembers int
	Topic      *TopicPurpose
	Purpose    *TopicPurpose
}

// Message represents a single Slack message.
type Message struct {
	Type       string
	Text       string
	User       string
	BotID      string
	TS         string
	ThreadTS   string
	ReplyCount int
}

// ScheduledMessage represents a message scheduled for future delivery.
type ScheduledMessage struct {
	ID          string
	ChannelID   string
	PostAt      int64
	DateCreated int64
	Text        string
}

// HistoryResult holds a list of messages together with a user-ID to
// display-name mapping for the users referenced in those messages.
type HistoryResult struct {
	Messages []Message
	Users    map[string]string
}

// ChannelUnreadResult contains unread messages for a single channel.
type ChannelUnreadResult struct {
	Channel               Channel
	Messages              []Message
	Users                 map[string]string
	TotalUnreadCount      int
	DisplayedMessageCount int
}

// SearchMatchChannel identifies the channel a search match belongs to.
type SearchMatchChannel struct {
	ID   string
	Name string
}

// SearchMatch is a single hit returned by search.messages.
type SearchMatch struct {
	Text      string
	User      string
	Username  string
	TS        string
	Permalink string
	Channel   SearchMatchChannel
}

// SearchResult is the response envelope for a search.messages call.
type SearchResult struct {
	Query      string
	Matches    []SearchMatch
	TotalCount int
	Page       int
	PageCount  int
}

// SearchMessagesOptions controls search.messages pagination and sorting.
type SearchMessagesOptions struct {
	Sort    string
	SortDir string
	Count   int
	Page    int
}

// SlackUser represents a workspace member.
type SlackUser struct {
	ID       string
	Name     string
	RealName string
	Profile  *UserProfile
	TZ       string
	TZLabel  string
	IsAdmin  bool
	IsBot    bool
	Deleted  bool
}

// UserProfile contains profile fields for a SlackUser.
type UserProfile struct {
	Email       string
	DisplayName string
	Title       string
	StatusText  string
	StatusEmoji string
}

// UserPresence reports a user's online/away status.
type UserPresence struct {
	Presence string
}

// PinnedMessage is the message body inside a PinnedItem.
type PinnedMessage struct {
	Text string
	User string
	TS   string
}

// PinnedItem represents a pinned message in a channel.
type PinnedItem struct {
	Type      string
	Created   int64
	CreatedBy string
	Message   *PinnedMessage
}

// Reminder represents a Slack reminder.
type Reminder struct {
	ID         string
	Text       string
	Time       int64
	CompleteTS int64
	Recurring  bool
}

// StarMessage is the message body inside a StarredItem.
type StarMessage struct {
	Text string
	TS   string
}

// StarredItem represents a starred (bookmarked) item.
type StarredItem struct {
	Type       string
	Channel    string
	Message    StarMessage
	DateCreate int64
}

// StarListResult wraps a list of starred items.
type StarListResult struct {
	Items []StarredItem
}

// ListChannelsOptions controls conversations.list filtering.
type ListChannelsOptions struct {
	Types           string
	ExcludeArchived bool
	Limit           int
}

// HistoryOptions controls conversations.history pagination.
type HistoryOptions struct {
	Limit  int
	Oldest string
}

// ChannelMembersOptions controls conversations.members pagination.
type ChannelMembersOptions struct {
	Limit  int
	Cursor string
}

// ChannelMembersResult holds a page of channel member IDs.
type ChannelMembersResult struct {
	Members    []string
	NextCursor string
}

// CanvasSectionElement is a node in a canvas document tree.
type CanvasSectionElement struct {
	Type     string
	Text     string
	Elements []CanvasSectionElement
}

// CanvasSection is a top-level section of a Slack canvas.
type CanvasSection struct {
	ID       string
	Elements []CanvasSectionElement
}

// CanvasFile represents a canvas file attached to a channel.
type CanvasFile struct {
	ID       string
	Name     string
	Created  int64
	FileType string
}

// UploadFileOptions configures a file or snippet upload.
type UploadFileOptions struct {
	Channel        string
	FilePath       string
	Content        string
	Filename       string
	Title          string
	InitialComment string
	SnippetType    string
	ThreadTS       string
}
