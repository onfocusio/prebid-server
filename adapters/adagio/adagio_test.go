package adagio

import (
	"encoding/json"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/prebid/prebid-server/adapters/adapterstest"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func buildFakeBidRequest() openrtb2.BidRequest {
	imp1 := openrtb2.Imp{
		ID:     "some-impression-id",
		Banner: &openrtb2.Banner{},
		Ext:    json.RawMessage(`{"bidder": {"organizationId": "1000", "site": "site-name", "placement": "ban_atf"}}`),
	}

	fakeBidRequest := openrtb2.BidRequest{
		ID:  "some-request-id",
		Imp: []openrtb2.Imp{imp1},
	}

	return fakeBidRequest
}

func TestJsonSamples(t *testing.T) {
	bidder, buildErr := Builder(openrtb_ext.BidderAdagio, config.Adapter{
		Endpoint: "http://localhost/prebid_server"})

	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}

	adapterstest.RunJSONBidderTest(t, "adagiotest", bidder)
}

func TestMakeRequests_NoGzip(t *testing.T) {
	fakeBidRequest := buildFakeBidRequest()
	fakeBidRequest.Test = 1 // Do not use Gzip in Test Mode.

	bidder, buildErr := Builder(openrtb_ext.BidderAdagio, config.Adapter{
		Endpoint: "http://localhost/prebid_server"})

	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}

	requestData, err := bidder.MakeRequests(&fakeBidRequest, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(requestData))

	body := &openrtb2.BidRequest{}
	_ = json.Unmarshal(requestData[0].Body, body)

	assert.Equal(t, 1, len(body.Imp))
}

func TestMakeRequests_Gzip(t *testing.T) {
	fakeBidRequest := buildFakeBidRequest()

	bidder, buildErr := Builder(openrtb_ext.BidderAdagio, config.Adapter{
		Endpoint: "http://localhost/prebid_server"})

	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}

	requestData, _ := bidder.MakeRequests(&fakeBidRequest, nil)
	assert.Equal(t, []string([]string{"gzip"}), requestData[0].Headers["Content-Encoding"])
}