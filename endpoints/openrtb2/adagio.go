package openrtb2

import (
	"errors"
	"net"
	"net/http"

	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/hooks/hookexecution"
	"github.com/prebid/prebid-server/v2/metrics"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/stored_requests"
	"github.com/prebid/prebid-server/v2/stored_requests/backends/empty_fetcher"
	"github.com/prebid/prebid-server/v2/stored_responses"
	"github.com/prebid/prebid-server/v2/util/iputil"
	"github.com/prebid/prebid-server/v2/util/uuidutil"
)

type AdagioParserConfig struct {
	MaxRequestSize int64

	// When true, PBS will assign a randomly generated UUID to req.Source.TID if it is empty
	AutoGenSourceTID bool

	// GenerateRequestID overrides the bidrequest.id in an AMP Request or an App Stored Request with a generated UUID if set to true. The default is false.
	GenerateRequestID bool

	// Map of blacklisted apps that is used to create the hash table BlacklistedAppMap so App.ID's can be instantly accessed.
	BlacklistedAppMap map[string]bool

	// RequestValidation specifies the request validation options.
	RequestValidationIPv4PrivateNetworksParsed []net.IPNet
	RequestValidationIPv6PrivateNetworksParsed []net.IPNet

	CompressionRequest config.CompressionInfo
}

type AdagioParser struct {
	deps *endpointDeps
}

func NewAdagioParser(
	uuidGenerator uuidutil.UUIDGenerator,
	validator openrtb_ext.BidderParamValidator,
	requestsById stored_requests.Fetcher,
	accounts stored_requests.AccountFetcher,
	parserCFG *AdagioParserConfig,
	metricsEngine metrics.MetricsEngine,
	disabledBidders map[string]string,
	defReqJSON []byte,
	bidderMap map[string]openrtb_ext.BidderName,
	storedRespFetcher stored_requests.Fetcher,
) (*AdagioParser, error) {
	if validator == nil || requestsById == nil || accounts == nil || parserCFG == nil || metricsEngine == nil {
		return nil, errors.New("NewEndpoint requires non-nil arguments.")
	}

	cfg := config.Configuration{
		MaxRequestSize:    parserCFG.MaxRequestSize,
		AutoGenSourceTID:  parserCFG.AutoGenSourceTID,
		GenerateRequestID: parserCFG.GenerateRequestID,
		BlacklistedAppMap: parserCFG.BlacklistedAppMap,
		RequestValidation: config.RequestValidation{
			IPv4PrivateNetworksParsed: parserCFG.RequestValidationIPv4PrivateNetworksParsed,
			IPv6PrivateNetworksParsed: parserCFG.RequestValidationIPv6PrivateNetworksParsed,
		},
		Compression: config.Compression{
			Request: parserCFG.CompressionRequest,
		},
	}

	defRequest := len(defReqJSON) > 0

	ipValidator := iputil.PublicNetworkIPValidator{
		IPv6PrivateNetworks: cfg.RequestValidation.IPv6PrivateNetworksParsed,
		IPv4PrivateNetworks: cfg.RequestValidation.IPv4PrivateNetworksParsed,
	}

	return &AdagioParser{
		deps: &endpointDeps{
			uuidGenerator,
			nil,
			validator,
			requestsById,
			empty_fetcher.EmptyFetcher{},
			accounts,
			&cfg,
			metricsEngine,
			nil,
			disabledBidders,
			defRequest,
			defReqJSON,
			bidderMap,
			nil,
			nil,
			ipValidator,
			storedRespFetcher,
			nil,
			nil,
			openrtb_ext.NormalizeBidderName,
		},
	}, nil
}

func (a *AdagioParser) ParseRequest(
	httpRequest *http.Request,
	labels *metrics.Labels,
	hookExecutor hookexecution.HookStageExecutor,
) (
	req *openrtb_ext.RequestWrapper,
	impExtInfoMap map[string]exchange.ImpExtInfo,
	storedAuctionResponses stored_responses.ImpsWithBidResponses,
	storedBidResponses stored_responses.ImpBidderStoredResp,
	bidderImpReplaceImpId stored_responses.BidderImpReplaceImpID,
	account *config.Account,
	errs []error,
) {
	return a.deps.parseRequest(httpRequest, labels, hookExecutor)
}
