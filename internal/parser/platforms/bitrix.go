package platforms

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"website-scraper/internal/models"
)

// BitrixParser реализация парсера для Bitrix
type BitrixParser struct{}

// NewBitrixParser создает новый экземпляр BitrixParser
func NewBitrixParser() PlatformParser {
	return &BitrixParser{}
}

// DetectPlatform проверяет, соответствует ли страница Bitrix
func (p *BitrixParser) DetectPlatform(html string) bool {
	bitrixPatterns := []string{
		`bitrix/js`,
		`bitrix/templates`,
		`<meta name="generator" content="Bitrix`,
		`BX\.`,
		`b24-widget`,
		`class="bx-`,
		`id="bx_`,
		`<!-- Bitrix`,
		`1C-Bitrix`,
		`/bitrix/`,
	}

	for _, pattern := range bitrixPatterns {
		if strings.Contains(html, pattern) {
			return true
		}
	}

	return false
}

// ParseHeader парсит шапку сайта Bitrix
func (p *BitrixParser) ParseHeader(html string) (*models.Block, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	contentMap := make(map[string]interface{})
	header := &models.Block{
		BlockType: models.BlockTypeHeader,
		Platform:  models.PlatformBitrix,
		Content:   contentMap,
	}

	// Добавляем версию Bitrix
	contentMap["version"] = detectBitrixVersion(html)

	headerSelectors := []string{
		"header",
		"div.header",
		"div#header",
		".site-header",
		"#site-header",
		"div[role='banner']",
		".main-header",
		"#main-header",
	}

	var headerContainer *goquery.Selection
	for _, selector := range headerSelectors {
		if doc.Find(selector).Length() > 0 {
			headerContainer = doc.Find(selector).First()
			break
		}
	}

	if headerContainer == nil {
		return header, nil
	}

	headerHtml, err := headerContainer.Html()
	if err == nil {
		header.HTML = headerHtml
	}

	// Парсинг компонентов шапки
	// Логотип
	logoSelectors := []string{
		".logo",
		".site-logo",
		"#logo",
		"a img",
	}

	for _, selector := range logoSelectors {
		logo := headerContainer.Find(selector).First()
		if logo.Length() > 0 {
			logoSrc, exists := logo.Attr("src")
			if exists {
				contentMap["logo"] = logoSrc
				break
			} else if href, exists := logo.Attr("href"); exists {
				contentMap["logo_link"] = href
				break
			}
		}
	}

	// Меню
	menuSelectors := []string{
		"nav",
		".navigation",
		".menu",
		"ul.menu",
	}

	var menuItems []string
	for _, selector := range menuSelectors {
		menu := headerContainer.Find(selector)
		if menu.Length() > 0 {
			menu.Find("a").Each(func(i int, item *goquery.Selection) {
				text := strings.TrimSpace(item.Text())
				if text != "" {
					menuItems = append(menuItems, text)
				}
			})
		}
		if len(menuItems) > 0 {
			contentMap["menu"] = menuItems
			break
		}
	}

	// Поиск
	searchSelectors := []string{
		"form[role='search']",
		".search-form",
		"input[type='search']",
	}

	for _, selector := range searchSelectors {
		if headerContainer.Find(selector).Length() > 0 {
			contentMap["search"] = true
			break
		}
	}

	// Телефоны
	phoneSelectors := []string{
		".phone",
		".contact-phone",
		"a[href^='tel:']",
	}

	var phones []string
	for _, selector := range phoneSelectors {
		headerContainer.Find(selector).Each(func(i int, sel *goquery.Selection) {
			phone := strings.TrimSpace(sel.Text())
			if isValidPhone(phone) {
				phones = append(phones, phone)
			}
		})
		if len(phones) > 0 {
			contentMap["phones"] = phones
			break
		}
	}

	// Корзина
	cartSelectors := []string{
		".cart",
		".shopping-cart",
		"a[href*='cart']",
	}

	for _, selector := range cartSelectors {
		if headerContainer.Find(selector).Length() > 0 {
			contentMap["cart"] = true
			break
		}
	}

	// Авторизация
	authSelectors := []string{
		".login",
		".auth",
		"a[href*='login']",
		"a[href*='auth']",
	}

	for _, selector := range authSelectors {
		if headerContainer.Find(selector).Length() > 0 {
			contentMap["auth"] = true
			break
		}
	}

	return header, nil
}

// ParseFooter парсит подвал сайта Bitrix
func (p *BitrixParser) ParseFooter(html string) (*models.Block, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	contentMap := make(map[string]interface{})
	footer := &models.Block{
		BlockType: models.BlockTypeFooter,
		Platform:  models.PlatformBitrix,
		Content:   contentMap,
	}

	// Добавляем версию Bitrix
	contentMap["version"] = detectBitrixVersion(html)

	footerSelectors := []string{
		"footer",
		"div.footer",
		"div#footer",
		".site-footer",
		"#site-footer",
		"div[role='contentinfo']",
		".main-footer",
		"#main-footer",
	}

	var footerContainer *goquery.Selection
	for _, selector := range footerSelectors {
		if doc.Find(selector).Length() > 0 {
			footerContainer = doc.Find(selector).First()
			break
		}
	}

	if footerContainer == nil {
		return footer, nil
	}

	footerHtml, err := footerContainer.Html()
	if err == nil {
		footer.HTML = footerHtml
	}

	// Парсинг компонентов подвала
	// Копирайт
	copyrightSelectors := []string{
		".copyright",
		".site-info",
		"p",
		"div",
	}

	for _, selector := range copyrightSelectors {
		elements := footerContainer.Find(selector)
		elements.Each(func(i int, elem *goquery.Selection) {
			text := strings.TrimSpace(elem.Text())
			if strings.Contains(text, "©") || strings.Contains(text, "&copy;") || strings.Contains(text, "Copyright") {
				contentMap["copyright"] = text
				return
			}
		})
		if _, exists := contentMap["copyright"]; exists {
			break
		}
	}

	// Меню в подвале
	menuSelectors := []string{
		"nav",
		".footer-menu",
		"ul.menu",
	}

	var menuItems []string
	for _, selector := range menuSelectors {
		menu := footerContainer.Find(selector)
		if menu.Length() > 0 {
			menu.Find("a").Each(func(i int, item *goquery.Selection) {
				text := strings.TrimSpace(item.Text())
				if text != "" {
					menuItems = append(menuItems, text)
				}
			})
		}
		if len(menuItems) > 0 {
			contentMap["menu"] = menuItems
			break
		}
	}

	// Контакты
	contactSelectors := []string{
		".contacts",
		".footer-contacts",
		".address",
	}

	var contacts []string
	for _, selector := range contactSelectors {
		footerContainer.Find(selector).Each(func(i int, sel *goquery.Selection) {
			sel.Find("p, div").Each(func(i int, item *goquery.Selection) {
				text := strings.TrimSpace(item.Text())
				if text != "" {
					contacts = append(contacts, text)
				}
			})
		})
		if len(contacts) > 0 {
			contentMap["contacts"] = contacts
			break
		}
	}

	// Социальные сети
	socialSelectors := []string{
		".social",
		".social-links",
		".socials",
	}

	var socialLinks []string
	for _, selector := range socialSelectors {
		footerContainer.Find(selector).Each(func(i int, sel *goquery.Selection) {
			sel.Find("a").Each(func(i int, item *goquery.Selection) {
				if href, exists := item.Attr("href"); exists {
					socialLinks = append(socialLinks, href)
				}
			})
		})
		if len(socialLinks) > 0 {
			contentMap["social"] = socialLinks
			break
		}
	}

	// Разработчик
	developerSelectors := []string{
		".developer",
		".developed-by",
		"a[href*='developer']",
	}

	for _, selector := range developerSelectors {
		if elem := footerContainer.Find(selector).First(); elem.Length() > 0 {
			if href, exists := elem.Attr("href"); exists {
				contentMap["developer"] = href
				break
			}
		}
	}

	return footer, nil
}

// detectBitrixVersion определяет версию Bitrix
func detectBitrixVersion(html string) string {
	versionPatterns := map[string]string{
		"modern": `BX24\.|b24-`,
		"legacy": `bitrix\/js\/main\/core\/core`,
		"old":    `bitrix\/components\/bitrix`,
	}

	for version, pattern := range versionPatterns {
		if matched, _ := regexp.MatchString(pattern, html); matched {
			return version
		}
	}

	return "unknown"
}

// isValidPhone проверяет валидность телефонного номера
func isValidPhone(phone string) bool {
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")
	return len(cleaned) >= 5 && regexp.MustCompile(`[\d]{5,}`).MatchString(cleaned)
}
