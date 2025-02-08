package main

import (
	"embed"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
	PubDate     string `xml:"pubDate"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	ImageURL    string `xml:"imageURL"`
}

func (a *App) GetRSS(url string) ([]Item, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var rss RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling XML: %w", err)
	}

	for i := range rss.Channel.Items {
		item := &rss.Channel.Items[i]
		parsedTime, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			parsedTime, err = time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				log.Println("Error parsing date:", err, item.PubDate)
				item.PubDate = "Date parsing error"

			} else {
				item.PubDate = parsedTime.Format("2006-01-02 15:04")
			}

		} else {
			item.PubDate = parsedTime.Format("2006-01-02 15:04")
		}

	}
	return rss.Channel.Items, nil
}

func (a *App) GetArticleImage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %w", err)
	}

	ImageURL := ""
	if strings.Contains(url, "index.hr") {
		imgSel := doc.Find(".img-responsive").First()
		if src, exists := imgSel.Attr("src"); exists {
			ImageURL = src
		}
	} else if strings.Contains(url, "slobodnadalmacija.hr") {
		imgSel := doc.Find(".card__image").First()
		if src, exists := imgSel.Attr("src"); exists {
			ImageURL = src
		}
	} else if strings.Contains(url, "telegram.hr") {
		imgSel := doc.Find(".article-head-image").First().Children().First()
		if src, exists := imgSel.Attr("src"); exists {
			ImageURL = src
		}
	}

	return ImageURL, nil
}

func (a *App) GetArticleContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Index.hr specific, images are commented
	uncommentedBody := string(bodyBytes)
	uncommentedBody = strings.ReplaceAll(uncommentedBody, "<!--<", "<")
	uncommentedBody = strings.ReplaceAll(uncommentedBody, ">-->", ">")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(uncommentedBody)) // Use the uncommented body
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %w", err)
	}

	var content strings.Builder
	selector := ""

	if strings.Contains(url, "24sata.hr") {
		selector = "article"
	} else if strings.Contains(url, "index.hr") {
		selector = ".left-part"
	} else if strings.Contains(url, "slobodnadalmacija.hr") {
		selector = ".itemFullText.itemFullText--freecontent"
	} else if strings.Contains(url, "telegram.hr") {
		selector = "#article-body"
	}

	excluded24sataHr := []string{"article_navigation", "dfp_banner dfp_banner--billboard_mid", "share_bar", "app_promo_block_container", "article_keywords_container", "article__content_container engagement_bar_wrapper", "article__thread"}
	excludedIndexHr := []string{"js-slot-container", "tags-holder", "article-report-container", "article-call-to-action", "main-img-desc", "loading-text", "front-gallery-holder flex", "gallery-thumb-slider gallery-slider swiper", "gallery-desc-slider gallery-slider swiper"}
	excludedSlobodnaDalmacijaHr := []string{"se-embed se-embed--photo"}
	excludedTelegramHr := []string{"full flex overtitle-parent relative", "nothfive full flex relative article-meta", "full relative single-article-footer flex column-top-pad", "full flex cxenseignore article-full-width", "full flex cxenseignore", "perex", "fb-post"}

	excludedClasses := slices.Concat(excluded24sataHr, excludedIndexHr, excludedSlobodnaDalmacijaHr, excludedTelegramHr)

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		processNode(s, &content, excludedClasses) // Process the selected element and its descendants
	})

	if content.String() == "" {
		return "", fmt.Errorf("element with selector '%s' not found", selector)
	}

	return content.String(), nil
}

func processNode(n *goquery.Selection, content *strings.Builder, excludedClasses []string) {
	if n.Is("script") {
		return
	}

	shouldExclude := false
	for _, excludedClass := range excludedClasses {
		if n.HasClass(excludedClass) {
			shouldExclude = true
			break
		}
		classValue, exists := n.Attr("class")
		if exists {
			classes := strings.Fields(classValue)
			for _, class := range classes {
				if strings.Contains(class, "Widget") {
					shouldExclude = true
					break
				}
			}
		}
		if shouldExclude {
			break
		}
	}

	if shouldExclude {
		return // Skip this element and its descendants
	}

	clone := n.Clone()

	clone.Children().RemoveFiltered(":not(em):not(a):not(strong):not(b)")
	n.Children().RemoveFiltered("em, a, strong, b")

	// clone.Children().Remove()

	// div, err := goquery.OuterHtml(clone)
	// if err != nil {
	// 	log.Printf("Error getting HTML: %v", err)
	// }

	for i := range len(n.Nodes) {
		content.WriteString("<" + n.Nodes[i].Data)
		for _, attr := range n.Nodes[i].Attr { // Add attributes
			if strings.Contains(attr.Val, "padding-bottom") { // Index.hr specific, avoid padding under article image
				continue
			}
			content.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Key, attr.Val))
		}
		content.WriteString(">")

		content.WriteString(clone.Text())

		// Recursively process children
		n.Children().Each(func(i int, child *goquery.Selection) {
			processNode(child, content, excludedClasses)
		})

		content.WriteString("</" + n.Nodes[i].Data + ">") // Closing tag
	}
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Samo vijesti",
		Width:     1024,
		Height:    768,
		MinWidth:  950,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
