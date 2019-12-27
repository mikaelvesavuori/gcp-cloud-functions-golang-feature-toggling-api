# Go/Golang feature toggling API running on Google Cloud Functions

This is a simple feature toggling API using Google Cloud Functions. When called, it reads a public (could be configured to private with a bit of work) JSON file in a Cloud Storage bucket, returning data for an object that matches the requested **market** query. _The idea is to provide a simple, zero maintenance, easy-to-use, resilient and very fast API to ask for market-based feature flags._

It uses [Serverless Framework](https://serverless.com/) to deploy your function, but you should be able to copy source code into the Cloud Functions inline code view and set your environment variables there to correlate with what's asked for in `serverless.yml`. Or whatever floats your boat!

_Tiny context_: I've been really eager to learn a bit of Go since I work almost completely around serverless tech these days, and Go makes so much sense in that world. Since the company I'm working at currently ([Humblebee](https://www.humblebee.se)) has a UX-first approach and that we've been disappointed with current feature toggling offerings (too much, too little, too expensive, too slow...) I figured that I might as well make my own proof-of-concept for it using Go, since I wanted to learn Go either way. My feeling is it was a good fit and a fun but challenging language to take on.

_Summary_: If you snicker at the code, this is my first attempt at Go! :rofl:

## Data structure

I've provided a sample data file (`data.json`) that you can use as your own basis. Since this is Go, the structure/format itself is typed and so the format is required and any changes or amendments on your end need to be reflected by changes in the source code itself.

The file looks like this:

```
{
  "featureFlags": [
    {
      "market": "SE",
      "newFeatureActive": true,
      "abSplitPercentage": { "new": 50, "current": 50 }
    },
    {
      "market": "US",
      "newFeatureActive": false,
      "abSplitPercentage": { "new": 30, "current": 70 }
    },
    {
      "market": "JP",
      "newFeatureActive": true,
      "abSplitPercentage": { "new": 80, "current": 20 }
    },
    {
      "market": "default",
      "newFeatureActive": false,
      "abSplitPercentage": { "new": 0, "current": 100 }
    }
  ]
}
```

In short, you get all the fields that go with the requested market. I've used the common cases of a boolean feature toggle and a struct carrying A/B splitting percentages. Feel free to add as much as you need!

## Usage

Call your Cloud Functions endpoint with a POST request and a body such as:

```
{
	"market": "JP"
}
```

If you are using my sample data, valid markets would be **JP**, **SE**, **US** and **default** (response code 200). Invalid requests will send back `Bad request` and response code 400.

## Installation and setup

### Setup for Go modules

Since you require a few Go modules for this to work, you need to have `go.mod` and `go.sum` laying around. To create these, run:

- `go mod init {MODULE_NAME}` (example: `github.com/{USER_NAME}/{REPOSITORY_NAME}`)
- `go mod tidy`

My own `go.mod` would look like:

```
module github.com/mikaelvesavuori/gcp-cloud-functions-golang-feature-toggling-api

go 1.11

require (
	cloud.google.com/go/storage v1.4.0
	google.golang.org/api v0.15.0
)
```

Notice that the Go version is actually automatically set to `go 1.13` but I've manually put it to `go 1.11` since that's the version GCP supports.

### Setup for Cloud Storage bucket

#### Create a readable bucket

I've gone with a multi-regional, publicly readable bucket where I've put `data.json`. It should of course be fully possible to lock down security on the bucket to this individual API/function, if you so choose.

#### Update serverless.yml with your settings

In `serverless.yml`, you need to set `project` and `credentials` fields to point to your own stuff. Please refer to [https://serverless.com/framework/docs/providers/google/guide/credentials/](https://serverless.com/framework/docs/providers/google/guide/credentials/) if all of this is new to you.

After creating the bucket, also make sure to update the _environment_ section and its variables to correlate with your own settings.

## Performance

My non-scientific testing gets warm results from my near-by regional endpoint—**europe-north1** (Finland) from western Sweden—in around 135-200 ms. Using a previous version with hardcoded data (instead of fetching from a multi-regional EU storage bucket) response times were typically 70 ms for the same conditions. Thus, adding the external bucket added at least 65 ms of latency, but also made it easier to update the data (not needing to re-deploy the entire function just because of data changes). I've used a couple of SaaS solutions that are a whole lot slower than this, so I feel content with the results.

## Multiple regional endpoints: Proxying requests to nearest endpoint

I have a working example of proxying requests to the nearest endpoint using Cloudflare Workers, which you can find at [https://github.com/mikaelvesavuori/cloudflare-workers-demos/blob/master/edge-proxy-feature-flags.js](https://github.com/mikaelvesavuori/cloudflare-workers-demos/blob/master/edge-proxy-feature-flags.js).

You would call the Cloudflare Worker function (serverless edge function) and let it resolve the user location and then fetch data based on continent. My example uses a US, EU, and Asian location for a few select countries—you'd need to fill out that list of countries unless you use Cloudflare Business that can resolve continents automatically.

This approach requires you to:

- Deploy functions to several independent locations (`serverless.yml` in this repo only does one location at the moment)
- Use Cloudflare (set up an account etc etc) and use their Workers functionality
- Deploy the edge code on Cloudflare (my code example should be ready-to-go, though you must give it valid endpoints)
- Finally, of course tie together your solution that actually starts calling the things in the first place

## Possible future improvements

- Add in-function data caching. Reference: [https://dev.to/yvonnickfrin/how-to-add-cache-to-your-gcp-cloud-functions-3ech](https://dev.to/yvonnickfrin/how-to-add-cache-to-your-gcp-cloud-functions-3ech)

## Thanks

Thanks to [Rick Crawford](https://gist.github.com/rickcrawford) and his gist at [https://gist.github.com/rickcrawford/f1d9ffb9b9d5649e0e6029e4419c08d7](https://gist.github.com/rickcrawford/f1d9ffb9b9d5649e0e6029e4419c08d7) for outlining how to do reading from a JSON file in a GCS bucket.
