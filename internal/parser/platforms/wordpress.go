package platforms

import (
	"regexp"
	"strings"

	"website-scraper/internal/models"
)

// WordPressParser реализация парсера для WordPress
type WordPressParser struct{}

// NewWordPressParser создает новый экземпляр WordPressParser
func NewWordPressParser() PlatformParser {
	return &WordPressParser{}
}

// DetectPlatform проверяет, соответствует ли страница WordPress
func (p *WordPressParser) DetectPlatform(html string) bool {
	wpPatterns := []string{
		`wp-content`,
		`wp-includes`,
		`wp-json`,
		`<meta name="generator" content="WordPress`,
		`class="wordpress"`,
		`/wp-admin/`,
		`/wp-login.php`,
	}

	for _, pattern := range wpPatterns {
		if strings.Contains(html, pattern) {
			return true
		}
	}

	return false
}

// ParseHeader парсит шапку сайта WordPress
func (p *WordPressParser) ParseHeader(html string) (*models.Block, error) {
	reHeader := regexp.MustCompile(`(?s)<header.*?>(.*?)</header>`)
	headerMatch := reHeader.FindStringSubmatch(html)

	if len(headerMatch) < 2 {
		reHeaderClass := regexp.MustCompile(`(?s)<div\s+(?:class=".*?header.*?".*?|id=".*?header.*?".*?)>(.*?)</div>`)
		headerMatch = reHeaderClass.FindStringSubmatch(html)
	}

	if len(headerMatch) < 2 {
		reSiteHeader := regexp.MustCompile(`(?s)<div\s+class=".*?site-header.*?".*?>(.*?)</div>`)
		headerMatch = reSiteHeader.FindStringSubmatch(html)
	}

	if len(headerMatch) < 2 {
		return nil, nil
	}

	headerHtml := headerMatch[0]

	// Парсим логотип
	reLogo := regexp.MustCompile(`(?s)<a.*?class=".*?logo.*?".*?>(.*?)</a>`)
	logoMatch := reLogo.FindStringSubmatch(headerHtml)

	if len(logoMatch) < 2 {
		reLogo = regexp.MustCompile(`(?s)<a.*?><img.*?(?:alt|title)=".*?logo.*?".*?></a>`)
		logoMatch = reLogo.FindStringSubmatch(headerHtml)
	}

	if len(logoMatch) < 2 {
		reLogo = regexp.MustCompile(`(?s)<a.*?><img.*?src=".*?".*?></a>`)
		logoMatch = reLogo.FindStringSubmatch(headerHtml)
	}

	var logoHtml string
	if len(logoMatch) >= 1 {
		logoHtml = logoMatch[0]
	}

	// Парсим навигационное меню
	reMenu := regexp.MustCompile(`(?s)<nav.*?>(.*?)</nav>`)
	menuMatch := reMenu.FindStringSubmatch(headerHtml)

	if len(menuMatch) < 2 {
		reMenu = regexp.MustCompile(`(?s)<(?:div|ul)\s+class=".*?(?:menu|navigation).*?".*?>(.*?)</(?:div|ul)>`)
		menuMatch = reMenu.FindStringSubmatch(headerHtml)
	}

	var menuHtml string
	if len(menuMatch) >= 1 {
		menuHtml = menuMatch[0]
	}

	// Парсим контактную информацию
	reContacts := regexp.MustCompile(`(?s)<div\s+class=".*?(?:contact|phone|email|address).*?".*?>(.*?)</div>`)
	contactsMatch := reContacts.FindStringSubmatch(headerHtml)

	var contactsHtml string
	if len(contactsMatch) >= 1 {
		contactsHtml = contactsMatch[0]
	}

	// Парсим поиск
	reSearch := regexp.MustCompile(`(?s)<form.*?(?:class=".*?search.*?".*?|id=".*?search.*?".*?)>(.*?)</form>`)
	searchMatch := reSearch.FindStringSubmatch(headerHtml)

	var searchHtml string
	if len(searchMatch) >= 1 {
		searchHtml = searchMatch[0]
	}

	content := map[string]interface{}{
		"logo":     logoHtml,
		"menu":     menuHtml,
		"contacts": contactsHtml,
		"search":   searchHtml,
	}

	block := &models.Block{
		BlockType: models.BlockTypeHeader,
		Platform:  models.PlatformWordPress,
		Content:   content,
		HTML:      headerHtml,
	}

	return block, nil
}

// ParseFooter парсит подвал сайта WordPress
func (p *WordPressParser) ParseFooter(html string) (*models.Block, error) {
	reFooter := regexp.MustCompile(`(?s)<footer.*?>(.*?)</footer>`)
	footerMatch := reFooter.FindStringSubmatch(html)

	if len(footerMatch) < 2 {
		reFooterClass := regexp.MustCompile(`(?s)<div\s+(?:class=".*?footer.*?".*?|id=".*?footer.*?".*?)>(.*?)</div>`)
		footerMatch = reFooterClass.FindStringSubmatch(html)
	}

	if len(footerMatch) < 2 {
		reSiteFooter := regexp.MustCompile(`(?s)<div\s+class=".*?site-footer.*?".*?>(.*?)</div>`)
		footerMatch = reSiteFooter.FindStringSubmatch(html)
	}

	if len(footerMatch) < 2 {
		return nil, nil
	}

	footerHtml := footerMatch[0]

	// Парсим виджеты
	reWidgets := regexp.MustCompile(`(?s)<div\s+class=".*?widgets.*?".*?>(.*?)</div>`)
	widgetsMatch := reWidgets.FindStringSubmatch(footerHtml)

	if len(widgetsMatch) < 2 {
		reWidgets = regexp.MustCompile(`(?s)<div\s+class=".*?sidebar.*?".*?>(.*?)</div>`)
		widgetsMatch = reWidgets.FindStringSubmatch(footerHtml)
	}

	var widgetsHtml string
	if len(widgetsMatch) >= 1 {
		widgetsHtml = widgetsMatch[0]
	}

	// Парсим копирайт
	reCopyright := regexp.MustCompile(`(?s)<div\s+class=".*?copyright.*?".*?>(.*?)</div>`)
	copyrightMatch := reCopyright.FindStringSubmatch(footerHtml)

	if len(copyrightMatch) < 2 {
		reCopyright = regexp.MustCompile(`(?s)(?:©|&copy;).*?\d{4}`)
		copyrightMatch = reCopyright.FindStringSubmatch(footerHtml)
	}

	var copyrightHtml string
	if len(copyrightMatch) >= 1 {
		copyrightHtml = copyrightMatch[0]
	}

	// Парсим социальные сети
	reSocial := regexp.MustCompile(`(?s)<div\s+class=".*?social.*?".*?>(.*?)</div>`)
	socialMatch := reSocial.FindStringSubmatch(footerHtml)

	if len(socialMatch) < 2 {
		reSocial = regexp.MustCompile(`(?s)<ul\s+class=".*?(?:social|socials).*?".*?>(.*?)</ul>`)
		socialMatch = reSocial.FindStringSubmatch(footerHtml)
	}

	var socialHtml string
	if len(socialMatch) >= 1 {
		socialHtml = socialMatch[0]
	}

	// Парсим ссылки навигации
	reNavLinks := regexp.MustCompile(`(?s)<nav\s+class=".*?footer-navigation.*?".*?>(.*?)</nav>`)
	navLinksMatch := reNavLinks.FindStringSubmatch(footerHtml)

	if len(navLinksMatch) < 2 {
		reNavLinks = regexp.MustCompile(`(?s)<ul\s+class=".*?footer-menu.*?".*?>(.*?)</ul>`)
		navLinksMatch = reNavLinks.FindStringSubmatch(footerHtml)
	}

	var navLinksHtml string
	if len(navLinksMatch) >= 1 {
		navLinksHtml = navLinksMatch[0]
	}

	content := map[string]interface{}{
		"widgets":   widgetsHtml,
		"copyright": copyrightHtml,
		"social":    socialHtml,
		"nav_links": navLinksHtml,
	}

	block := &models.Block{
		BlockType: models.BlockTypeFooter,
		Platform:  models.PlatformWordPress,
		Content:   content,
		HTML:      footerHtml,
	}

	return block, nil
}
