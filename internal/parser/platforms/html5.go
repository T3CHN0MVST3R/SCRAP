package platforms

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"website-scraper/internal/models"
)

// HTML5Parser реализация парсера для HTML5
type HTML5Parser struct{}

// NewHTML5Parser создает новый экземпляр HTML5Parser
func NewHTML5Parser() *HTML5Parser {
	return &HTML5Parser{}
}

// DetectPlatform проверяет, соответствует ли страница HTML5
func (p *HTML5Parser) DetectPlatform(html string) bool {
	htmlPatterns := []string{
		`<!DOCTYPE html>`,
		`<html lang`,
		`<meta charset="UTF-8">`,
	}

	for _, pattern := range htmlPatterns {
		if strings.Contains(html, pattern) {
			return true
		}
	}

	return false
}

// ParseHeader парсит шапку HTML5 сайта
func (p *HTML5Parser) ParseHeader(html string) (*models.Block, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var headerHtml string
	var headerNode *goquery.Selection

	// Ищем семантический тег header
	headerNode = doc.Find("header").First()
	if headerNode.Length() > 0 {
		headerHtml, err = headerNode.Html()
		if err != nil {
			return nil, err
		}
	} else {
		// Альтернативные селекторы для шапки
		alternativeSelectors := []string{
			"div.header", "div#header", ".site-header", "#site-header",
			"div[role='banner']", ".main-header", "#main-header",
		}

		for _, selector := range alternativeSelectors {
			headerNode = doc.Find(selector).First()
			if headerNode.Length() > 0 {
				headerHtml, err = headerNode.Html()
				if err != nil {
					return nil, err
				}
				break
			}
		}
	}

	if headerHtml == "" {
		return nil, nil
	}

	// Парсим компоненты шапки
	content := make(map[string]interface{})

	// Логотип
	logo := headerNode.Find("a img, .logo, .site-logo, #logo").First()
	if logo.Length() > 0 {
		logoSrc, exists := logo.Attr("src")
		if exists {
			content["logo"] = logoSrc
		}
	}

	// Навигация
	nav := headerNode.Find("nav, .navigation, .menu, ul.menu").First()
	if nav.Length() > 0 {
		var menuItems []string
		nav.Find("a").Each(func(i int, item *goquery.Selection) {
			text := strings.TrimSpace(item.Text())
			if text != "" {
				menuItems = append(menuItems, text)
			}
		})
		if len(menuItems) > 0 {
			content["menu"] = menuItems
		}
	}

	// Контактная информация
	contact := headerNode.Find(".contact, .phone, .email").First()
	if contact.Length() > 0 {
		content["contact"] = strings.TrimSpace(contact.Text())
	}

	return &models.Block{
		BlockType: models.BlockTypeHeader,
		Platform:  models.PlatformHTML5,
		Content:   content,
		HTML:      headerHtml,
	}, nil
}

// ParseFooter парсит подвал HTML5 сайта
func (p *HTML5Parser) ParseFooter(html string) (*models.Block, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var footerHtml string
	var footerNode *goquery.Selection

	// Ищем семантический тег footer
	footerNode = doc.Find("footer").First()
	if footerNode.Length() > 0 {
		footerHtml, err = footerNode.Html()
		if err != nil {
			return nil, err
		}
	} else {
		// Альтернативные селекторы для подвала
		alternativeSelectors := []string{
			"div.footer", "div#footer", ".site-footer", "#site-footer",
			"div[role='contentinfo']", ".main-footer", "#main-footer",
		}

		for _, selector := range alternativeSelectors {
			footerNode = doc.Find(selector).First()
			if footerNode.Length() > 0 {
				footerHtml, err = footerNode.Html()
				if err != nil {
					return nil, err
				}
				break
			}
		}
	}

	if footerHtml == "" {
		return nil, nil
	}

	// Парсим компоненты подвала
	content := make(map[string]interface{})

	// Копирайт
	copyright := footerNode.Find(".copyright, .site-info").First()
	if copyright.Length() > 0 {
		content["copyright"] = strings.TrimSpace(copyright.Text())
	} else {
		// Ищем копирайт по паттерну
		footerNode.Find("p, div").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if strings.Contains(text, "©") || strings.Contains(text, "&copy;") || strings.Contains(text, "Copyright") {
				content["copyright"] = text
			}
		})
	}

	// Социальные сети
	var socialLinks []string
	footerNode.Find(".social, .social-links, .socials").Find("a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			socialLinks = append(socialLinks, href)
		}
	})
	if len(socialLinks) > 0 {
		content["social"] = socialLinks
	}

	return &models.Block{
		BlockType: models.BlockTypeFooter,
		Platform:  models.PlatformHTML5,
		Content:   content,
		HTML:      footerHtml,
	}, nil
}

// ParseAndClassifyPage парсит всю страницу и классифицирует блоки
func (p *HTML5Parser) ParseAndClassifyPage(html string, templates []models.BlockTemplate) ([]*models.Block, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var blocks []*models.Block

	// Шаг 1: Находим шапку
	header := doc.Find("header").First()
	if header.Length() > 0 {
		headerBlock, err := p.ParseHeader(html)
		if err == nil && headerBlock != nil {
			blocks = append(blocks, headerBlock)
		}
	}

	// Шаг 2: Находим подвал
	footer := doc.Find("footer").First()
	var footerElement *goquery.Selection
	if footer.Length() > 0 {
		footerElement = footer
		footerBlock, err := p.ParseFooter(html)
		if err == nil && footerBlock != nil {
			// Добавим подвал в конце
			defer func() {
				blocks = append(blocks, footerBlock)
			}()
		}
	} else {
		// Альтернативные селекторы для подвала
		alternativeSelectors := []string{
			"div.footer", "div#footer", ".site-footer", "#site-footer",
		}

		for _, selector := range alternativeSelectors {
			elem := doc.Find(selector).First()
			if elem.Length() > 0 {
				footerElement = elem
				break
			}
		}
	}

	// Шаг 3: Находим контентные секции
	potentialBlocks := doc.Find("section, div.section, div[class*='section'], div[class*='block'], div[class*='container'], div.content, main > div")

	// Отслеживаем уже обработанные элементы
	processedElements := make(map[string]bool)

	// Обрабатываем каждый потенциальный блок
	potentialBlocks.Each(func(i int, section *goquery.Selection) {
		// Генерируем уникальный ключ для элемента
		outerHTML, err := goquery.OuterHtml(section)
		if err != nil {
			return
		}

		// Пропускаем уже обработанные
		if processedElements[outerHTML] {
			return
		}

		// Пропускаем элементы внутри шапки или подвала
		if header.Length() > 0 {
			if isDescendantOf(section, header) {
				return
			}
		}
		if footerElement != nil {
			if isDescendantOf(section, footerElement) {
				return
			}
		}

		// Пропускаем маленькие секции
		if len(strings.TrimSpace(section.Text())) < 30 && !containsImage(section) {
			return
		}

		// Отмечаем как обработанный
		processedElements[outerHTML] = true

		// Получаем внутренний HTML для сопоставления
		sectionHTML, err := section.Html()
		if err != nil {
			return
		}

		// Только обрабатываем если секция имеет контент
		if len(sectionHTML) < 50 && !containsImage(section) {
			return
		}

		// Пытаемся сопоставить с шаблонами
		matchedTemplate := p.matchBlockWithTemplates(sectionHTML, templates)

		// Создаем блок
		content := map[string]interface{}{}
		if matchedTemplate != nil {
			content["template_name"] = matchedTemplate.BlockType
			content["matched_pattern"] = true
		} else {
			// Используем эвристику для классификации несопоставленных блоков
			blockType := p.classifyBlockByHeuristics(section)
			if blockType != "" {
				content["template_name"] = blockType
				content["matched_pattern"] = false
			} else {
				content["template_name"] = "Unknown Content Block"
				content["matched_pattern"] = false
			}
		}

		blocks = append(blocks, &models.Block{
			BlockType: models.BlockTypeContent,
			Platform:  models.PlatformHTML5,
			Content:   content,
			HTML:      sectionHTML,
		})
	})

	return blocks, nil
}

// isDescendantOf проверяет, является ли элемент потомком родителя
func isDescendantOf(element, parent *goquery.Selection) bool {
	// Проверяем, является ли элемент тем же, что и родитель
	if element.Is(parent.Nodes[0].Data) {
		return true
	}

	// Проверяем всех родителей элемента
	parents := element.Parents()
	result := false
	parents.Each(func(i int, s *goquery.Selection) {
		// Сравниваем с родителем
		parent.Each(func(j int, p *goquery.Selection) {
			if isSameNode(s, p) {
				result = true
				return
			}
		})
	})

	return result
}

// isSameNode проверяет, относятся ли две выборки к одному и тому же узлу
func isSameNode(a, b *goquery.Selection) bool {
	if a.Length() == 0 || b.Length() == 0 {
		return false
	}

	// Используем OuterHtml как простой способ сравнения
	aHtml, err1 := a.Html()
	bHtml, err2 := b.Html()

	if err1 == nil && err2 == nil {
		return aHtml == bHtml
	}

	return false
}

// matchBlockWithTemplates пытается сопоставить HTML блока с шаблонами
func (p *HTML5Parser) matchBlockWithTemplates(blockHTML string, templates []models.BlockTemplate) *models.BlockTemplate {
	for _, template := range templates {
		if template.HTML5 == nil {
			continue
		}

		var patternData map[string]interface{}
		if jsonBytes, ok := template.HTML5.(json.RawMessage); ok {
			if err := json.Unmarshal(jsonBytes, &patternData); err != nil {
				continue
			}
		} else if jsonStr, ok := template.HTML5.(string); ok {
			if err := json.Unmarshal([]byte(jsonStr), &patternData); err != nil {
				continue
			}
		} else {
			continue
		}

		// Проверяем все шаги в шаблоне
		matched := true
		for i := 1; ; i++ {
			key := fmt.Sprintf("step%d", i)
			patternValue, exists := patternData[key]
			if !exists {
				break // Больше нет шагов
			}

			// Обрабатываем паттерн в зависимости от типа
			switch val := patternValue.(type) {
			case string:
				// Обрабатываем OR паттерны (разделены |)
				if strings.Contains(val, "|") {
					found := false
					for _, option := range strings.Split(val, "|") {
						if strings.Contains(blockHTML, strings.TrimSpace(option)) {
							found = true
							break
						}
					}
					if !found {
						matched = false
						break
					}
				} else {
					// Простой строковый паттерн
					if !strings.Contains(blockHTML, val) {
						matched = false
						break
					}
				}
			case []interface{}:
				// Обрабатываем массив паттернов (достаточно любого совпадения)
				found := false
				for _, item := range val {
					if tag, ok := item.(string); ok && strings.Contains(blockHTML, tag) {
						found = true
						break
					}
				}
				if !found {
					matched = false
					break
				}
			}
		}

		if matched {
			return &template
		}
	}

	return nil
}

// classifyBlockByHeuristics использует эвристику для классификации блока
func (p *HTML5Parser) classifyBlockByHeuristics(section *goquery.Selection) string {
	// Подсчитываем элементы
	imageCount := section.Find("img").Length()
	buttonCount := section.Find("button, a.btn, .button, [class*='btn-']").Length()
	headingCount := section.Find("h1, h2, h3, h4, h5, h6").Length()
	paragraphCount := section.Find("p").Length()
	formCount := section.Find("form").Length()
	tableCount := section.Find("table").Length()

	// Проверяем специфические компоненты
	hasMap := section.Find("[class*='map'], iframe[src*='map']").Length() > 0
	hasContactInfo := section.Find("[class*='contact'], [id*='contact']").Length() > 0 ||
		containsPhoneEmail(section)
	hasProducts := section.Find("[class*='product'], [class*='item'], .card").Length() > 0
	hasSlider := section.Find("[class*='slider'], [class*='carousel'], [class*='swiper']").Length() > 0
	hasFAQ := section.Find("[class*='faq'], [class*='accordion'], .collapse").Length() > 0

	// Логика классификации
	if hasMap {
		return "Карта"
	} else if formCount > 0 {
		return "Форма обратной связи"
	} else if hasContactInfo {
		return "Контакты"
	} else if hasFAQ {
		return "FAQ"
	} else if tableCount > 0 {
		return "Таблица"
	} else if hasSlider {
		if paragraphCount > 0 || headingCount > 0 {
			return "Карусель, слайд шоу с текстом"
		}
		return "Карусель, слайд шоу"
	} else if hasProducts {
		return "Товары"
	} else if imageCount > 0 && buttonCount > 0 {
		return "Картинка + Действие"
	} else if paragraphCount > 0 && imageCount > 0 {
		return "Текст блок + Картинка"
	} else if paragraphCount > 0 && buttonCount > 0 {
		return "Текст + Действие"
	} else if imageCount > 0 {
		// Проверяем колоночную разметку
		if hasColumnLayout(section, 3) {
			return "Блок с картинкой 3 колонки"
		} else {
			return "Блок с картинкой"
		}
	} else if paragraphCount > 0 || headingCount > 0 {
		// Проверяем колоночную разметку
		if hasColumnLayout(section, 2) {
			return "Текстовый блок 2 колонки"
		} else {
			return "Текстовый блок"
		}
	}

	return "Смешанный контент"
}

// containsImage проверяет, содержит ли выборка изображения
func containsImage(s *goquery.Selection) bool {
	return s.Find("img").Length() > 0
}

// containsPhoneEmail проверяет, содержит ли выборка телефон или email
func containsPhoneEmail(s *goquery.Selection) bool {
	text := s.Text()

	// Проверяем телефоны
	phonePatterns := []string{
		`\+\d{1,3}\s*\(\d{3,}\)\s*\d{3,}`, // +7 (999) 123-45-67
		`\+\d{10,}`,                       // +79991234567
		`\d{3,}[\s-]?\d{3,}[\s-]?\d{2,}`,  // 999 123 45 67
	}

	for _, pattern := range phonePatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}

	// Проверяем email
	emailPattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	if matched, _ := regexp.MatchString(emailPattern, text); matched {
		return true
	}

	return false
}

// hasColumnLayout проверяет, имеет ли выборка колоночную разметку
func hasColumnLayout(s *goquery.Selection, columnCount int) bool {
	// Проверяем общие паттерны классов колонок
	columnPatterns := []string{
		fmt.Sprintf("col-%d", columnCount),
		fmt.Sprintf("column-%d", columnCount),
		fmt.Sprintf("grid-%d", columnCount),
		"row",
		"flex",
		"grid",
	}

	for _, pattern := range columnPatterns {
		if s.Find(fmt.Sprintf("[class*='%s']", pattern)).Length() > 0 {
			return true
		}
	}

	// Подсчитываем прямые дочерние div-ы
	directChildDivs := s.Find("> div").Length()
	if directChildDivs == columnCount {
		return true
	}

	return false
}
