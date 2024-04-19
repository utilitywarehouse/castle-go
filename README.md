# castle-go

castle-go is a Go library wrapping the https://castle.io API. 

## Install

```
go get github.com/utilitywarehouse/castle-go
```

## Usage

### Providing own http client

```go
castle.NewWithHTTPClient("secret-api-key", &http.Client{Timeout: time.Second * 2})
```

## API

The pkg wraps the two [Risk Assessment endpoints](https://reference.castle.io/#tag/risk_assessment) of the Castle API: Risk and Filter.

The difference between the two are better explained in the [docs](https://docs.castle.io/docs/integration-guide):

> The biggest difference between the Risk API and the Filter API is the former is used for checking a user that has successfully logged in (so you measure risk of their actions in your app), whereas the latter is used to check visitors, before they log in (so you filter out abusive behavior).

The right usage of the Filter API and rest of the technical differences are laid out [here](https://docs.castle.io/docs/anonymous-activity).

All in all, use Filter API for anonymous user events, and Risk API for logged in users.
The model of the endpoints is almost the same, the request is almost identical while the response is 100%. Both return risk assessment scores, which depending on the flow (event and status) might be ignored.

#### Notes

The [Log API](https://reference.castle.io/#tag/logging) is currently not exposed in this pkg, as it is not a risk assessment endpoint, therefore the general risk scoring is not affected by it:

> Scores are computed in real time from the data sent via the Risk and Filter APIs
[1](https://docs.castle.io/docs/risk-scoring)

> Note that these can also be sent to the Log API, but that would degrade risk scoring performance since the risk score isn't evaluated for Log events
[2](https://docs.castle.io/docs/anonymous-activity)

## Repo

Originally forked from [castle/castle-go](https://github.com/castle/castle-go) now it lives on its own. The original repo has not been maintained, and as of today only supports long deprecated Castle APIs.
