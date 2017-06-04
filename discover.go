package sub

import (
	"encoding/xml"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

type atomLink struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom link"`
	Rel     string   `xml:"rel,attr"`
	Href    string   `xml:"href,attr"`
}

// Discovers details about a subscription. This version allows you to pass in a
// custom client which can be used to modify the discovery request before it is
// sent by overriding client.Get
func DiscoverWithClient(topic string, client *http.Client) (*Sub, error) {
	topicUrl, err := url.Parse(topic)
	if err != nil {
		return nil, err
	}

	// get the topic
	resp, err := client.Get(topic)
	if err != nil {
		return nil, err
	}

	// TODO: try to get links from headers
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var hubUrl *url.URL

	rawContentType := resp.Header.Get("Content-Type")
	contentType, _, err := mime.ParseMediaType(rawContentType)
	if err != nil {
		return nil, err
	}

	suffix := strings.Split(contentType, "/")[1]
	if suffix == "xml" {
		// try parsing as an atom feed

		feed := struct {
			Link []atomLink `xml:"http://www.w3.org/2005/Atom link"`
		}{}

		xml.Unmarshal(body, &feed)

		// find link rel="hub"
		for i := 0; i < len(feed.Link); i++ {
			link := feed.Link[i]
			if link.Rel == "hub" {
				hubString := link.Href
				hub, err := url.Parse(hubString)
				if err != nil {
					return nil, err
				}

				hubUrl = hub
				break
			}
		}

	} else {
		return nil, &ResponseError{
			Response: resp,
			Message:  "Unknown response type",
		}
	}

	return &Sub{
		Client: client,
		Hub:    hubUrl,
		Topic:  topicUrl,
		State:  Requested,
	}, nil
}

// Discover subscription details from a topic url.
func Discover(topic string) (*Sub, error) {
	return DiscoverWithClient(topic, http.DefaultClient)
}
