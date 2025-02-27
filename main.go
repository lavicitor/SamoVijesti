package main

import (
	"embed"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
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

func (a *App) IsPremium(url string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("error fetching RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error parsing HTML: %w", err)
	}

	if strings.Contains(url, "slobodnadalmacija.hr") {
		if doc.Find(".itemFullText.itemFullText--premium").Nodes != nil {
			return true, nil
		}
	} else if strings.Contains(url, "vecernji.hr") {
		// if doc.Find("").Nodes != nil {
		// 	return true, nil
		// } // TODO: Figure out how to filter vecernji.hr articles
	}

	return false, nil
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
	} else if strings.Contains(url, "jutarnji.hr") {
		imgSel := doc.Find(".card__image").First()
		if src, exists := imgSel.Attr("src"); exists {
			ImageURL = src
		}
	} else if strings.Contains(url, "sisak.info") {
		imgSel := doc.Find(".tdb-featured-image-bg").First().Parent()
		re := regexp.MustCompile(`url\('([^']+)'\)`)
		match := re.FindStringSubmatch(imgSel.Text())
		if len(match) > 1 {
			ImageURL = match[1]
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
	} else if strings.Contains(url, "vecernji.hr") {
		ImageURL, _ = doc.Find("meta[property='og:image']").Attr("content")
	}

	return ImageURL, nil
}

func (a *App) GetArticleContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("error fetching URL: %w", err)
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("HTTP error: %d - %s\n", resp.StatusCode, string(body))
		return "", fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading response body: %w", err)
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Index.hr specific, images are commented
	uncommentedBody := string(bodyBytes)
	uncommentedBody = strings.ReplaceAll(uncommentedBody, "<!--<", "<")
	uncommentedBody = strings.ReplaceAll(uncommentedBody, ">-->", ">")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(uncommentedBody)) // Use the uncommented body
	if err != nil {
		log.Println("error parsing HTML: %w", err)
		return "", fmt.Errorf("error parsing HTML: %w", err)
	}

	var content strings.Builder
	var selector string
	var excludedClasses []string

	if strings.Contains(url, "24sata.hr") {
		selector = "article"
		excludedClasses = []string{"article_navigation", "article__info_wrap", "engagement_bar_container", "dfp_banner dfp_banner--billboard_mid", "share_bar", "app_promo_block_container", "article_keywords_container", "article__content_container engagement_bar_wrapper", "article__thread"}
	} else if strings.Contains(url, "index.hr") {
		selector = ".left-part"
		excludedClasses = []string{"home-cube-big-timeline", "items-header flex", "nav main-tab main-content-bg flex sticky", "details-addition vijesti-text", "timeline-content latest", "timeline-content most-read", "tab-content-most-read", "bottom-border", "read-more-slider ", "article-read-more-container", "js-slot-container", "tags-holder", "article-report-container", "article-call-to-action", "main-img-desc", "loading-text", "front-gallery-holder flex", "gallery-thumb-slider gallery-slider swiper", "gallery-desc-slider gallery-slider swiper"}
	} else if strings.Contains(url, "jutarnji.hr") {
		selector = ".itemFullText"
		excludedClasses = []string{"se-embed"}
		doc.Find("*").Each(func(i int, s *goquery.Selection) {
			// Remove spam
			if s.Nodes[0].Data == "h3" {
				s.SetHtml("")
			}
			if s.Nodes[0].Data == "p" && strings.Contains(s.Text(), "PROČITAJTE VIŠE") {
				s.SetHtml("")
			}

		})
	} else if strings.Contains(url, "sisak.info") {
		selector = ".tdc_zone"
		excludedClasses = []string{"vc_column tdi_145  wpb_column vc_column_container tdc-column td-pb-span12", "td_block_wrap tdb_single_date tdi_130 td-pb-border-top td_block_template_1 tdb-post-meta", "vc_row tdi_137  wpb_row td-pb-row", "td_block_inner td-mc1-wrap", "tdc_zone tdi_148  wpb_row td-pb-row", "vc_column tdi_132  wpb_column vc_column_container tdc-column td-pb-span3", "td_block_wrap tdb_single_author tdi_129 td-pb-border-top td_block_template_1 tdb-post-meta", "td_block_wrap tdb_single_tags tdi_127 td-pb-border-top td_block_template_1", "td_block_wrap tdb_single_post_share tdi_126  td-pb-border-top td_block_template_1", "tdc_zone tdi_75  wpb_row td-pb-row tdc-element-style", "tdi_74_rand_style td-element-style", "tdc_zone tdi_2  wpb_row td-pb-row tdc-element-style", "tdc_zone tdi_15  wpb_row td-pb-row tdc-element-style", "tdc_zone tdi_28  wpb_row td-pb-row", "tdc-row tdc-row-is-sticky tdc-rist-bottom stretch_row_1400 td-stretch-content", "td-a-ad id_bottom_ad", "vc_row_inner tdi_112  vc_row vc_inner wpb_row td-pb-row"}
	} else if strings.Contains(url, "slobodnadalmacija.hr") {
		selector = ".row.row--grid.row--content"
		excludedClasses = []string{"se-embed se-embed--photo", "item__image", "item__tags", "item__related", "item__image-info"}
	} else if strings.Contains(url, "telegram.hr") {
		selector = "#article-body"
		excludedClasses = []string{"full flex overtitle-parent relative", "nothfive full flex relative article-meta", "full relative single-article-footer flex column-top-pad", "full flex cxenseignore article-full-width", "full flex cxenseignore", "perex", "fb-post"}
	} else if strings.Contains(url, "vecernji.hr") {
		selector = ".main"
		excludedClasses = []string{"tags", "meta--article-mobile", "uploaded_video", "lazyload", "lazyloaded", "relatedArticle", "card", "tes-widget", "mj-pf-widget", "vecernji_hr_spotlight", "component", "mfp-hide", "top-offer-template js-top-offer-template", "image__icons", "custom_creative__template js_customBillboardTemplate", "banner__label", "block js-articleToolbar block--article-toolbar block--fixed", "text2speech-template js-text2speech-template", "twitter-tweet"}

		doc.Find("*").Each(func(i int, s *goquery.Selection) {
			// Remove spam
			if s.Nodes[0].Data == "p" && (strings.Contains(s.Text(), ">>") || strings.Contains(s.Text(), "VEZANI ČLANCI:")) {
				s.SetHtml("")
			}
			/*
			 *  The site references images relatively, and we are parsing HTML without running
			 *  scripts so this is needed.
			 */
			if s.Nodes[0].Data == "noscript" {
				s.SetHtml(strings.Replace(s.Text(), "src=\"", "src=\"https://vecernji.hr", 1))
			}
			// For images not related to scripts
			if s.Nodes[0].Data == "img" {
				src, _ := s.Attr("src")
				if strings.Compare(src, "") != 0 {
					s.SetAttr("src", "https://vecernji.hr"+src)
				}
				src, _ = s.Attr("srcset")
				if strings.Compare(src, "") != 0 {
					s.SetAttr("srcset", "https://vecernji.hr"+src)
				}
				src, _ = s.Attr("data-src")
				if strings.Compare(src, "") != 0 {
					s.SetAttr("data-src", "https://vecernji.hr"+src)
				}
				src, _ = s.Attr("data-srcset")
				if strings.Compare(src, "") != 0 {
					oldarray := strings.Split(src, ",")
					var newarray []string
					for _, arrayelement := range oldarray {
						arrayelement = "https://vecernji.hr" + arrayelement
						if len(newarray) == 0 {
							newarray = append(newarray, arrayelement)
						}
						newarray = append(newarray, ", "+arrayelement)
					}
					s.SetAttr("data-srcset", strings.Join(newarray, ""))
				}
			}
		})
	}

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		processNode(s, &content, excludedClasses) // Process the selected element and its descendants
	})

	if content.String() == "" {
		log.Printf("element with selector '%s' not found\n", selector)
		return "", fmt.Errorf("element with selector '%s' not found", selector)
	}

	return content.String(), nil
}

func processNode(n *goquery.Selection, content *strings.Builder, excludedClasses []string) {
	if n.Is("script") || n.Is("style") {
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
