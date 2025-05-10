package platforms

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"website-scraper/internal/models"
)

// TildaParser реализация парсера для Tilda
type TildaParser struct{}

// NewTildaParser создает новый экземпляр TildaParser
func NewTildaParser() PlatformParser {
	return &TildaParser{}
}

// DetectPlatform проверяет, соответствует ли страница Tilda
func (p *TildaParser) DetectPlatform(html string) bool {
	tildaPatterns := []string{
		`tilda.ws`,
		`tildacdn.com`,
		`<meta name="generator" content="Tilda`,
		`data-tilda`,
		`t-body`,
		`t-page`,
		`t-records`,
		`t-rec`,
		`class="t`,
		`id="t`,
	}

	for _, pattern := range tildaPatterns {
		if strings.Contains(html, pattern) {
			return true
		}
	}

	return false
}

// ParseHeader парсит шапку сайта Tilda
func (p *TildaParser) ParseHeader(html string) (*models.Block, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	contentMap := make(map[string]interface{})
	block := &models.Block{
		BlockType: models.BlockTypeHeader,
		Platform:  models.PlatformTilda,
		Content:   contentMap,
	}

	headerSelectors := []string{
		"div[id^='t-header']",
		"div[data-record-type='257']",
		"div[data-record-type='258']",
		"div[data-record-type='396']",
		"div.t-site-header-wrapper",
		".t396__elem.header",
	}

	var headerNode *goquery.Selection
	for _, selector := range headerSelectors {
		header := doc.Find(selector).First()
		if header.Length() > 0 {
			headerNode = header
			headerHtml, err := header.Html()
			if err == nil {
				block.HTML = headerHtml
				break
			}
		}
	}

	if headerNode == nil || block.HTML == "" {
		topBlocks := doc.Find("div[id^='rec']:lt(3)")
		if topBlocks.Length() > 0 {
			headerHtml, err := topBlocks.First().Html()
			if err == nil {
				block.HTML = headerHtml
				headerNode = topBlocks.First()
			}
		}
	}

	if block.HTML == "" {
		reHeader := regexp.MustCompile(`(?s)<div\s+(?:class=".*?(?:t-header|tn-header).*?".*?|id=".*?header.*?".*?)>(.*?)</div>`)
		headerMatch := reHeader.FindStringSubmatch(html)
		if len(headerMatch) > 0 {
			block.HTML = headerMatch[0]
		}
	}

	if block.HTML == "" {
		return nil, nil
	}

	if headerNode != nil {
		// Парсим компоненты шапки
		// Логотип
		logoSelectors := []string{
			".t-logo",
			".t-logo__img",
			"img[imgfield='img']",
			"a.t-menu__logo-wrapper img",
		}

		for _, selector := range logoSelectors {
			logo := headerNode.Find(selector).First()
			if logo.Length() > 0 {
				if logoSrc, exists := logo.Attr("src"); exists {
					contentMap["logo"] = logoSrc
					break
				}
			}
		}

		// Меню
		menuSelectors := []string{
			".t-menu__nav",
			".t-menu__nav-item",
			".tn-elem[data-elem-type='text'] a",
		}

		var menuItems []string
		for _, selector := range menuSelectors {
			menu := headerNode.Find(selector)
			if menu.Length() > 0 {
				menu.Find("a").Each(func(i int, item *goquery.Selection) {
					text := strings.TrimSpace(item.Text())
					if text != "" {
						menuItems = append(menuItems, text)
					}
				})

				if len(menuItems) > 0 {
					contentMap["menu"] = menuItems
					break
				}
			}
		}

		// Телефон
		phoneSelectors := []string{
			".t-menu__phone-item",
			".t-info_phone",
			".tn-elem[data-elem-type='text']:contains('+')",
		}

		for _, selector := range phoneSelectors {
			phoneElem := headerNode.Find(selector).First()
			if phoneElem.Length() > 0 {
				phoneText := strings.TrimSpace(phoneElem.Text())
				if phoneText != "" && isLikelyPhone(phoneText) {
					contentMap["phone"] = phoneText
					break
				}
			}
		}

		// Социальные сети
		socialSelectors := []string{
			".t-sociallinks",
			".t-sociallinks__item",
		}

		var socialLinks []string
		for _, selector := range socialSelectors {
			social := headerNode.Find(selector)
			if social.Length() > 0 {
				social.Find("a").Each(func(i int, item *goquery.Selection) {
					if href, exists := item.Attr("href"); exists && href != "" {
						socialLinks = append(socialLinks, href)
					}
				})

				if len(socialLinks) > 0 {
					contentMap["social"] = socialLinks
					break
				}
			}
		}

		// Кнопки
		buttonSelectors := []string{
			".t-btn",
			"a[href]:has(.t-btn)",
			".tn-atom.t-btn",
		}

		for _, selector := range buttonSelectors {
			button := headerNode.Find(selector).First()
			if button.Length() > 0 {
				buttonText := strings.TrimSpace(button.Text())
				if buttonText != "" {
					if href, exists := button.Attr("href"); exists && href != "" {
						contentMap["button"] = map[string]string{
							"text": buttonText,
							"link": href,
						}
					} else {
						contentMap["button"] = map[string]string{
							"text": buttonText,
						}
					}
					break
				}
			}
		}
	}

	return block, nil
}

// ParseFooter парсит подвал сайта Tilda
func (p *TildaParser) ParseFooter(html string) (*models.Block, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	contentMap := make(map[string]interface{})
	block := &models.Block{
		BlockType: models.BlockTypeFooter,
		Platform:  models.PlatformTilda,
		Content:   contentMap,
	}

	footerSelectors := []string{
		"div[id^='t-footer']",
		"div[data-record-type='56']",
		"div[data-record-type='106']",
		"div[data-record-type='331']",
		"div.t-site-footer-wrapper",
		".t396__elem.footer",
	}

	var footerNode *goquery.Selection
	for _, selector := range footerSelectors {
		footer := doc.Find(selector).First()
		if footer.Length() > 0 {
			footerNode = footer
			footerHtml, err := footer.Html()
			if err == nil {
				block.HTML = footerHtml
				break
			}
		}
	}

	if footerNode == nil || block.HTML == "" {
		recordBlocks := doc.Find("div[id^='rec']")
		if recordBlocks.Length() > 0 {
			lastBlock := recordBlocks.Last()
			footerHtml, err := lastBlock.Html()
			if err == nil {
				block.HTML = footerHtml
				footerNode = lastBlock
			}
		}
	}

	if block.HTML == "" {
		reFooter := regexp.MustCompile(`(?s)<div\s+(?:class=".*?(?:t-footer|tn-footer).*?".*?|id=".*?footer.*?".*?)>(.*?)</div>`)
		footerMatch := reFooter.FindStringSubmatch(html)
		if len(footerMatch) > 0 {
			block.HTML = footerMatch[0]
		}
	}

	if block.HTML == "" {
		return nil, nil
	}

	if footerNode != nil {
		// Парсим компоненты подвала
		// Копирайт
		copyrightSelectors := []string{
			".t-copyright",
			".t-footer_inlined-copyright",
			".tn-elem:contains('©')",
			".tn-elem:contains('copyright')",
		}

		for _, selector := range copyrightSelectors {
			copyright := footerNode.Find(selector).First()
			if copyright.Length() > 0 {
				copyrightText := strings.TrimSpace(copyright.Text())
				if copyrightText != "" {
					contentMap["copyright"] = copyrightText
					break
				}
			}
		}

		if _, exists := contentMap["copyright"]; !exists {
			footerNode.Find("div, p").Each(func(i int, elem *goquery.Selection) {
				text := strings.TrimSpace(elem.Text())
				if strings.Contains(text, "©") || strings.Contains(text, "&copy;") {
					contentMap["copyright"] = text
					return
				}
			})
		}

		// Контакты
		contactSelectors := []string{
			".t-footer_address",
			".t-info_address",
			".tn-elem:contains('contact')",
			".tn-elem:contains('адрес')",
		}

		for _, selector := range contactSelectors {
			contactElem := footerNode.Find(selector).First()
			if contactElem.Length() > 0 {
				contactText := strings.TrimSpace(contactElem.Text())
				if contactText != "" {
					contentMap["contact"] = contactText
					break
				}
			}
		}

		// Социальные сети
		socialSelectors := []string{
			".t-sociallinks",
			".t-sociallinks__item",
		}

		var socialLinks []string
		for _, selector := range socialSelectors {
			social := footerNode.Find(selector)
			if social.Length() > 0 {
				social.Find("a").Each(func(i int, item *goquery.Selection) {
					if href, exists := item.Attr("href"); exists && href != "" {
						socialLinks = append(socialLinks, href)
					}
				})

				if len(socialLinks) > 0 {
					contentMap["social"] = socialLinks
					break
				}
			}
		}

		// Меню в подвале
		menuSelectors := []string{
			".t-footer_menu",
			".t-footer__nav-item",
		}

		var menuItems []string
		for _, selector := range menuSelectors {
			menu := footerNode.Find(selector)
			if menu.Length() > 0 {
				menu.Find("a").Each(func(i int, item *goquery.Selection) {
					text := strings.TrimSpace(item.Text())
					if text != "" {
						menuItems = append(menuItems, text)
					}
				})

				if len(menuItems) > 0 {
					contentMap["menu"] = menuItems
					break
				}
			}
		}
	}

	return block, nil
}

// isLikelyPhone проверяет, похожа ли строка на телефонный номер
func isLikelyPhone(text string) bool {
	// Очищаем текст
	text = strings.Map(func(r rune) rune {
		if r == '+' || (r >= '0' && r <= '9') || r == '(' || r == ')' || r == '-' || r == ' ' {
			return r
		}
		return -1
	}, text)

	// Простая проверка: содержит определенное количество цифр
	digitCount := 0
	for _, char := range text {
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}

	return digitCount >= 7 && (strings.Contains(text, "+") || strings.Contains(text, "("))
}
