package sdk

import "maunium.net/go/mautrix/event"

func defaultSDKFeatureConfig() *RoomFeatures {
	return &RoomFeatures{
		MaxTextLength:        100000,
		SupportsReply:        true,
		SupportsReactions:    true,
		SupportsTyping:       true,
		SupportsReadReceipts: true,
		SupportsDeleteChat:   true,
	}
}

func computeRoomFeaturesForAgents(agents []*Agent) *RoomFeatures {
	if len(agents) == 0 {
		return defaultSDKFeatureConfig()
	}
	maxText := 0
	anyStreaming := false
	anyReasoning := false
	anyTools := false
	anyTextInput := false
	anyImageInput := false
	anyAudioInput := false
	anyVideoInput := false
	anyFileInput := false
	anyPDFInput := false
	anyImageOutput := false
	anyAudioOutput := false
	anyFilesOutput := false
	for _, agent := range agents {
		if agent == nil {
			continue
		}
		caps := agent.Capabilities
		if caps.MaxTextLength > maxText {
			maxText = caps.MaxTextLength
		}
		anyStreaming = anyStreaming || caps.SupportsStreaming
		anyReasoning = anyReasoning || caps.SupportsReasoning
		anyTools = anyTools || caps.SupportsToolCalling
		anyTextInput = anyTextInput || caps.SupportsTextInput
		anyImageInput = anyImageInput || caps.SupportsImageInput
		anyAudioInput = anyAudioInput || caps.SupportsAudioInput
		anyVideoInput = anyVideoInput || caps.SupportsVideoInput
		anyFileInput = anyFileInput || caps.SupportsFileInput
		anyPDFInput = anyPDFInput || caps.SupportsPDFInput
		anyImageOutput = anyImageOutput || caps.SupportsImageOutput
		anyAudioOutput = anyAudioOutput || caps.SupportsAudioOutput
		anyFilesOutput = anyFilesOutput || caps.SupportsFilesOutput
	}

	base := defaultSDKFeatureConfig()
	if maxText > 0 {
		base.MaxTextLength = maxText
	}
	base.SupportsImages = anyImageInput || anyImageOutput
	base.SupportsAudio = anyAudioInput || anyAudioOutput
	base.SupportsVideo = anyVideoInput
	base.SupportsFiles = anyFileInput || anyPDFInput || anyFilesOutput
	base.SupportsReply = anyTextInput
	base.SupportsTyping = anyStreaming
	base.SupportsReactions = anyTools || anyReasoning || anyTextInput
	base.SupportsReadReceipts = true
	base.SupportsDeleteChat = true
	return base
}

func convertRoomFeatures(f *RoomFeatures) *event.RoomFeatures {
	if f == nil {
		f = defaultSDKFeatureConfig()
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
	return convertRoomFeatures(defaultSDKFeatureConfig())
}

func capLevel(supported bool) event.CapabilitySupportLevel {
	if supported {
		return event.CapLevelFullySupported
	}
	return event.CapLevelRejected
}
