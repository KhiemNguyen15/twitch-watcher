package poller

import (
	"testing"

	"github.com/khiemnguyen15/twitch-watcher/pkg/models"
)

// collectRefs is a method on *Poller but uses no fields, so a zero-value Poller suffices.
var p = &Poller{}

func ref(subID, webhook string) models.SubscriptionRef {
	return models.SubscriptionRef{SubscriptionID: subID, DiscordWebhook: webhook}
}

func stream(gameName, userLogin string) models.TwitchStream {
	return models.TwitchStream{GameName: gameName, UserLogin: userLogin}
}

func TestCollectRefs_GameOnly(t *testing.T) {
	gameMap := map[string][]models.SubscriptionRef{
		"Fortnite": {ref("s1", "https://discord.com/api/webhooks/1/a")},
	}
	refs := p.collectRefs(stream("Fortnite", "streamer1"), gameMap, nil)
	if len(refs) != 1 || refs[0].SubscriptionID != "s1" {
		t.Errorf("expected 1 ref with s1, got %+v", refs)
	}
}

func TestCollectRefs_StreamerOnly(t *testing.T) {
	streamerMap := map[string][]models.SubscriptionRef{
		"streamer1": {ref("s2", "https://discord.com/api/webhooks/2/b")},
	}
	refs := p.collectRefs(stream("Fortnite", "streamer1"), nil, streamerMap)
	if len(refs) != 1 || refs[0].SubscriptionID != "s2" {
		t.Errorf("expected 1 ref with s2, got %+v", refs)
	}
}

func TestCollectRefs_GameAndStreamer_SameWebhook_Deduplicated(t *testing.T) {
	webhook := "https://discord.com/api/webhooks/1/a"
	gameMap := map[string][]models.SubscriptionRef{
		"Fortnite": {ref("s1", webhook)},
	}
	streamerMap := map[string][]models.SubscriptionRef{
		"streamer1": {ref("s2", webhook)}, // same webhook, different sub ID
	}
	refs := p.collectRefs(stream("Fortnite", "streamer1"), gameMap, streamerMap)
	if len(refs) != 1 {
		t.Errorf("expected 1 ref after dedup, got %d: %+v", len(refs), refs)
	}
}

func TestCollectRefs_GameAndStreamer_DifferentWebhooks_BothIncluded(t *testing.T) {
	gameMap := map[string][]models.SubscriptionRef{
		"Fortnite": {ref("s1", "https://discord.com/api/webhooks/1/a")},
	}
	streamerMap := map[string][]models.SubscriptionRef{
		"streamer1": {ref("s2", "https://discord.com/api/webhooks/2/b")},
	}
	refs := p.collectRefs(stream("Fortnite", "streamer1"), gameMap, streamerMap)
	if len(refs) != 2 {
		t.Errorf("expected 2 refs, got %d: %+v", len(refs), refs)
	}
}

func TestCollectRefs_NoMatch(t *testing.T) {
	gameMap := map[string][]models.SubscriptionRef{
		"Minecraft": {ref("s1", "https://discord.com/api/webhooks/1/a")},
	}
	refs := p.collectRefs(stream("Fortnite", "streamer1"), gameMap, nil)
	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(refs))
	}
}

func TestCollectRefs_MultipleSubscribersToSameGame(t *testing.T) {
	gameMap := map[string][]models.SubscriptionRef{
		"Fortnite": {
			ref("s1", "https://discord.com/api/webhooks/1/a"),
			ref("s2", "https://discord.com/api/webhooks/2/b"),
			ref("s3", "https://discord.com/api/webhooks/3/c"),
		},
	}
	refs := p.collectRefs(stream("Fortnite", "streamer1"), gameMap, nil)
	if len(refs) != 3 {
		t.Errorf("expected 3 refs, got %d", len(refs))
	}
}

func TestFormatThumbnail_CurlyBraces(t *testing.T) {
	input := "https://static-cdn.jtvnw.net/previews-ttv/live_user_foo-{width}x{height}.jpg"
	got := formatThumbnail(input, 440, 248)
	want := "https://static-cdn.jtvnw.net/previews-ttv/live_user_foo-440x248.jpg"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatThumbnail_URLEncoded(t *testing.T) {
	input := "https://static-cdn.jtvnw.net/previews-ttv/live_user_foo-%7Bwidth%7Dx%7Bheight%7D.jpg"
	got := formatThumbnail(input, 440, 248)
	want := "https://static-cdn.jtvnw.net/previews-ttv/live_user_foo-440x248.jpg"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatThumbnail_Empty(t *testing.T) {
	if got := formatThumbnail("", 440, 248); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFormatThumbnail_NoPlaceholders(t *testing.T) {
	input := "https://example.com/thumb.jpg"
	if got := formatThumbnail(input, 440, 248); got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}
