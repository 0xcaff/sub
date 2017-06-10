package sub

import (
	"encoding/xml"
	"io/ioutil"
	"mime"
	"net/url"
	"strings"
)

// Discovers a hub from s.Topic using s.Client.
func (s *Sub) Discover() error {
	// get the topic
	resp, err := s.Client.Get(s.Topic.String())
	if err != nil {
		return err
	}

	// TODO: try to get links from headers
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	rawContentType := resp.Header.Get("Content-Type")
	contentType, _, err := mime.ParseMediaType(rawContentType)
	if err != nil {
		return err
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
				s.Hub, err = url.Parse(link.Href)
				if err != nil {
					return err
				}

				break
			}
		}

	} else {
		return &ResponseError{
			Response: resp,
			Message:  "Unknown response type",
		}
	}

	return nil
}

type atomLink struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom link"`
	Rel     string   `xml:"rel,attr"`
	Href    string   `xml:"href,attr"`
}
