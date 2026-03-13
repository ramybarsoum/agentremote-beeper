package sdk

import "maunium.net/go/mautrix/event"

func convertRoomFeatures(f *RoomFeatures) *event.RoomFeatures {
	if f == nil {
		return defaultSDKRoomFeatures()
	}
	if f.Custom != nil {
		return f.Custom
	}
	maxText := f.MaxTextLength
	if maxText == 0 {
		maxText = 100000
	}
	capID := f.CustomCapabilityID
	if capID == "" {
		capID = "com.beeper.ai.sdk"
	}
	rf := &event.RoomFeatures{
		ID:                  capID,
		MaxTextLength:       maxText,
		Reply:               capLevel(f.SupportsReply),
		Edit:                capLevel(f.SupportsEdit),
		Delete:              capLevel(f.SupportsDelete),
		Reaction:            capLevel(f.SupportsReactions),
		ReadReceipts:        f.SupportsReadReceipts,
		TypingNotifications: f.SupportsTyping,
		DeleteChat:          f.SupportsDeleteChat,
		File:                make(event.FileFeatureMap),
	}
	if f.SupportsImages {
		rf.File[event.MsgImage] = &event.FileFeatures{}
	}
	if f.SupportsAudio {
		rf.File[event.MsgAudio] = &event.FileFeatures{}
	}
	if f.SupportsVideo {
		rf.File[event.MsgVideo] = &event.FileFeatures{}
	}
	if f.SupportsFiles {
		rf.File[event.MsgFile] = &event.FileFeatures{}
	}
	return rf
}

func defaultSDKRoomFeatures() *event.RoomFeatures {
	return &event.RoomFeatures{
		ID:                  "com.beeper.ai.sdk",
		MaxTextLength:       100000,
		Reply:               event.CapLevelFullySupported,
		Reaction:            event.CapLevelFullySupported,
		ReadReceipts:        true,
		TypingNotifications: true,
		DeleteChat:          true,
	}
}

func capLevel(supported bool) event.CapabilitySupportLevel {
	if supported {
		return event.CapLevelFullySupported
	}
	return event.CapLevelRejected
}
