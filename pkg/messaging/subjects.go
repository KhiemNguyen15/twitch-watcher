package messaging

// NATS subjects and JetStream stream/consumer name constants.
const (
	// Subjects
	SubjectStreamsRaw = "twitch.streams.raw"
	SubjectStreamsNew = "twitch.streams.new"

	// JetStream stream names
	StreamTwitchStreamsRaw = "TWITCH_STREAMS_RAW"
	StreamTwitchStreamsNew = "TWITCH_STREAMS_NEW"

	// Durable consumer names
	ConsumerStreamFilter           = "stream-filter"
	ConsumerNotificationDispatcher = "notification-dispatcher"
)
