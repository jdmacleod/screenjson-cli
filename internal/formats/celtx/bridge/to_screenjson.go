// Package bridge provides conversion between Celtx and ScreenJSON models.
package bridge

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	celtxmodel "screenjson/cli/internal/formats/celtx/model"
	"screenjson/cli/internal/model"
)

// ToScreenJSON converts a Celtx project to a ScreenJSON document.
func ToScreenJSON(project *celtxmodel.Project, lang string) *model.Document {
	if lang == "" {
		lang = "en"
	}

	authorID := uuid.New().String()

	// Extract author info
	authorName := "Unknown"
	authorFamily := "Author"
	if project.Author != "" {
		parts := strings.Fields(project.Author)
		if len(parts) >= 2 {
			authorName = parts[0]
			authorFamily = strings.Join(parts[1:], " ")
		} else if len(parts) == 1 {
			authorName = parts[0]
			authorFamily = ""
		}
	}

	title := project.Title
	if title == "" {
		title = "Untitled"
	}

	doc := &model.Document{
		ID:      uuid.New().String(),
		Version: "1.0.0",
		Generator: &model.Generator{
			Name:    "screenjson-cli",
			Version: "1.0.0",
		},
		Title:   model.Text{lang: title},
		Lang:    lang,
		Charset: "utf-8",
		Dir:     "ltr",
		Authors: []model.Author{
			{
				ID:     authorID,
				Given:  authorName,
				Family: authorFamily,
			},
		},
	}

	// Extract characters from scripts
	characters := extractCharacters(project)
	doc.Characters = characters

	// Build character lookup
	charMap := make(map[string]string)
	for _, c := range characters {
		charMap[strings.ToUpper(c.Name)] = c.ID
	}

	// Convert content
	doc.Content = convertContent(project, authorID, charMap, lang)

	return doc
}

// extractCharacters extracts unique character names from scripts.
func extractCharacters(project *celtxmodel.Project) []model.Character {
	seen := make(map[string]bool)
	var chars []model.Character

	for _, script := range project.Scripts {
		for _, elem := range script.Elements {
			if elem.Type == celtxmodel.TypeCharacter {
				name := cleanCharacterName(elem.Text)
				upperName := strings.ToUpper(name)

				if name != "" && !seen[upperName] {
					seen[upperName] = true
					chars = append(chars, model.Character{
						ID:   uuid.New().String(),
						Name: name,
					})
				}
			}
		}
	}

	return chars
}

// cleanCharacterName removes extensions from character names.
func cleanCharacterName(name string) string {
	if idx := strings.Index(name, "("); idx > 0 {
		name = name[:idx]
	}
	return strings.TrimSpace(name)
}

// convertContent converts Celtx scripts to ScreenJSON content.
func convertContent(project *celtxmodel.Project, authorID string, charMap map[string]string, lang string) *model.Content {
	content := &model.Content{
		Cover: &model.Cover{
			Title:   model.Text{lang: project.Title},
			Authors: []string{authorID},
		},
	}

	var scenes []model.Scene
	var currentScene *model.Scene
	var sceneNumber int

	for _, script := range project.Scripts {
		for _, elem := range script.Elements {
			switch elem.Type {
			case celtxmodel.TypeSceneHeading:
				if currentScene != nil {
					scenes = append(scenes, *currentScene)
				}
				sceneNumber++
				currentScene = &model.Scene{
					ID:      uuid.New().String(),
					Authors: []string{authorID},
					Heading: parseSlugline(elem.Text, strconv.Itoa(sceneNumber)),
					Body:    []model.Element{},
				}

			case celtxmodel.TypeAction:
				if currentScene == nil {
					sceneNumber++
					currentScene = createDefaultScene(authorID, strconv.Itoa(sceneNumber))
				}
				currentScene.Body = append(currentScene.Body, model.Element{
					ID:      uuid.New().String(),
					Type:    model.ElementAction,
					Authors: []string{authorID},
					Text:    model.Text{lang: elem.Text},
				})

			case celtxmodel.TypeCharacter:
				if currentScene == nil {
					sceneNumber++
					currentScene = createDefaultScene(authorID, strconv.Itoa(sceneNumber))
				}
				name := cleanCharacterName(elem.Text)
				charID := charMap[strings.ToUpper(name)]
				currentScene.Body = append(currentScene.Body, model.Element{
					ID:        uuid.New().String(),
					Type:      model.ElementCharacter,
					Authors:   []string{authorID},
					Character: charID,
					Display:   elem.Text,
				})
				if charID != "" && !containsString(currentScene.Cast, charID) {
					currentScene.Cast = append(currentScene.Cast, charID)
				}

			case celtxmodel.TypeDialog:
				if currentScene == nil {
					sceneNumber++
					currentScene = createDefaultScene(authorID, strconv.Itoa(sceneNumber))
				}
				currentScene.Body = append(currentScene.Body, model.Element{
					ID:      uuid.New().String(),
					Type:    model.ElementDialogue,
					Authors: []string{authorID},
					Text:    model.Text{lang: elem.Text},
				})

			case celtxmodel.TypeParenthetical:
				if currentScene == nil {
					sceneNumber++
					currentScene = createDefaultScene(authorID, strconv.Itoa(sceneNumber))
				}
				currentScene.Body = append(currentScene.Body, model.Element{
					ID:      uuid.New().String(),
					Type:    model.ElementParenthetical,
					Authors: []string{authorID},
					Text:    model.Text{lang: elem.Text},
				})

			case celtxmodel.TypeTransition:
				if currentScene == nil {
					sceneNumber++
					currentScene = createDefaultScene(authorID, strconv.Itoa(sceneNumber))
				}
				currentScene.Body = append(currentScene.Body, model.Element{
					ID:      uuid.New().String(),
					Type:    model.ElementTransition,
					Authors: []string{authorID},
					Text:    model.Text{lang: elem.Text},
				})
			}
		}
	}

	if currentScene != nil {
		scenes = append(scenes, *currentScene)
	}

	content.Scenes = scenes
	return content
}

func createDefaultScene(authorID string, sceneNumber string) *model.Scene {
	return &model.Scene{
		ID:      uuid.New().String(),
		Authors: []string{authorID},
		Heading: &model.Slugline{
			No:      sceneNumber,
			Context: "INT",
			Setting: "UNKNOWN",
			Time:    "DAY",
		},
		Body: []model.Element{},
	}
}

func parseSlugline(text string, sceneNumber string) *model.Slugline {
	slug := &model.Slugline{
		No:   sceneNumber,
		Time: "DAY",
	}

	text = strings.TrimSpace(text)
	upper := strings.ToUpper(text)

	if strings.HasPrefix(upper, "INT/EXT") || strings.HasPrefix(upper, "I/E") {
		slug.Context = "INT/EXT"
		text = strings.TrimPrefix(upper, "INT/EXT")
		text = strings.TrimPrefix(text, "I/E")
	} else if strings.HasPrefix(upper, "EXT/INT") {
		slug.Context = "EXT/INT"
		text = strings.TrimPrefix(upper, "EXT/INT")
	} else if strings.HasPrefix(upper, "INT.") || strings.HasPrefix(upper, "INT ") {
		slug.Context = "INT"
		text = strings.TrimPrefix(upper, "INT.")
		text = strings.TrimPrefix(text, "INT ")
	} else if strings.HasPrefix(upper, "EXT.") || strings.HasPrefix(upper, "EXT ") {
		slug.Context = "EXT"
		text = strings.TrimPrefix(upper, "EXT.")
		text = strings.TrimPrefix(text, "EXT ")
	} else {
		slug.Context = "INT"
	}

	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, ".")
	text = strings.TrimPrefix(text, "-")
	text = strings.TrimSpace(text)

	parts := strings.Split(text, " - ")
	if len(parts) >= 2 {
		slug.Setting = strings.TrimSpace(parts[0])
		slug.Time = strings.TrimSpace(parts[len(parts)-1])
	} else {
		slug.Setting = text
	}

	return slug
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
