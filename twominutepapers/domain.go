package twominutepapers

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes twominutepapers as a kit Domain so a multi-domain host (ant)
// can enable it with a single blank import:
//
//	import _ "github.com/tamnd/twominutepapers-cli/twominutepapers"
//
// The same Domain builds the standalone twominutepapers binary (see cli/root.go),
// so the binary and any host share one source of truth.
func init() { kit.Register(Domain{}) }

// Host is the YouTube channel URL host, used for URI classification.
const Host = "www.youtube.com"

// Domain is the twominutepapers driver. It carries no state.
type Domain struct{}

// Info describes the scheme, accepted hostnames, and the binary identity.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "twominutepapers",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "twominutepapers",
			Short:  "Browse the Two Minute Papers YouTube channel",
			Long: `twominutepapers reads the Two Minute Papers YouTube channel through its
public Atom RSS feed. No API key is required. It returns records as
table, JSON, JSONL, CSV, TSV, or URLs.

twominutepapers is an independent tool and is not affiliated with Two Minute Papers.`,
			Site: "www.youtube.com/@TwoMinutePapers",
			Repo: "https://github.com/tamnd/twominutepapers-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "videos", Group: "read", List: true,
		Summary: "List the latest Two Minute Papers videos"}, listVideos)
}

// newClient builds the Client from the kit config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

type listVideosIn struct {
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

func listVideos(ctx context.Context, in listVideosIn, emit func(*Video) error) error {
	videos, err := in.Client.Videos(ctx, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for i := range videos {
		if err := emit(&videos[i]); err != nil {
			return err
		}
	}
	return nil
}

// Classify turns a YouTube watch URL into (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	return "", "", errs.Usage("twominutepapers:// URIs are not supported for direct resolution")
}

// Locate is the inverse of Classify.
func (Domain) Locate(uriType, id string) (string, error) {
	return "", errs.Usage("twominutepapers has no resource type %q", uriType)
}

func mapErr(err error) error {
	return err
}
