// Package codec provides Fountain encoding/decoding.
package codec

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strings"

	ftnmodel "screenjson/cli/internal/formats/fountain/model"
)

var (
	sceneHeadingPattern  = regexp.MustCompile(`^(?i)(INT|EXT|EST|INT\./EXT|INT/EXT|I/E)[\.\s]`)
	transitionPattern    = regexp.MustCompile(`^[A-Z\s]+TO:$`)
	characterPattern     = regexp.MustCompile(`^[A-Z][A-Z0-9\s\-'\.]+(\s*\([^)]+\))?$`)
	parentheticalPattern = regexp.MustCompile(`^\([^)]+\)$`)
	forcedCharPattern    = regexp.MustCompile(`^@(.+)$`)
	forcedScenePattern   = regexp.MustCompile(`^\.(.+)$`)
	forcedActionPattern  = regexp.MustCompile(`^!(.+)$`)
	forcedTransPattern   = regexp.MustCompile(`^>(.+)$`)
	centeredPattern      = regexp.MustCompile(`^>(.+)<$`)
	sectionPattern       = regexp.MustCompile(`^(#{1,6})\s*(.+)$`)
	synopsisPattern      = regexp.MustCompile(`^=\s*(.+)$`)
	notePattern          = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	boneyardStart        = regexp.MustCompile(`/\*`)
	boneyardEnd          = regexp.MustCompile(`\*/`)
	pageBreakPattern     = regexp.MustCompile(`^===+$`)
	titleKeyPattern      = regexp.MustCompile(`^([A-Za-z\s]+):\s*(.*)$`)
	dualDialoguePattern  = regexp.MustCompile(`\s*\^$`)
)

// Decoder decodes Fountain files.
type Decoder struct{}

// NewDecoder creates a new Fountain decoder.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode parses Fountain text into the Fountain model.
func (d *Decoder) Decode(ctx context.Context, data []byte) (*ftnmodel.Document, error) {
	doc := &ftnmodel.Document{}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty Fountain input")
	}

	text := string(data)
	lines := strings.Split(text, "\n")

	// Parse title page if present
	titlePage, contentStart := parseTitlePage(lines)
	doc.TitlePage = titlePage

	// Parse content
	doc.Elements = parseContent(lines[contentStart:])

	return doc, nil
}

// parseTitlePage extracts title page metadata.
func parseTitlePage(lines []string) (*ftnmodel.TitlePage, int) {
	tp := &ftnmodel.TitlePage{
		Custom: make(map[string]string),
	}

	// Title page ends at first blank line followed by content
	// or after no title page markers are found

	i := 0
	inTitlePage := false
	var currentKey string
	var currentValue strings.Builder

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check for title page key
		if match := titleKeyPattern.FindStringSubmatch(line); match != nil {
			if currentKey != "" {
				setTitlePageValue(tp, currentKey, strings.TrimSpace(currentValue.String()))
			}
			currentKey = strings.TrimSpace(match[1])
			currentValue.Reset()
			currentValue.WriteString(match[2])
			inTitlePage = true
		} else if inTitlePage && trimmed == "" {
			// End of title page
			if currentKey != "" {
				setTitlePageValue(tp, currentKey, strings.TrimSpace(currentValue.String()))
			}
			i++
			break
		} else if inTitlePage && (strings.HasPrefix(line, "   ") || strings.HasPrefix(line, "\t")) {
			// Continuation of multi-line value
			currentValue.WriteString("\n")
			currentValue.WriteString(strings.TrimSpace(line))
		} else if !inTitlePage && trimmed != "" {
			// No title page, content starts here
			break
		}

		i++
	}

	if currentKey != "" {
		setTitlePageValue(tp, currentKey, strings.TrimSpace(currentValue.String()))
	}

	if !inTitlePage {
		return nil, 0
	}

	return tp, i
}

// setTitlePageValue sets a title page field by key.
func setTitlePageValue(tp *ftnmodel.TitlePage, key, value string) {
	switch strings.ToLower(key) {
	case "title":
		tp.Title = value
	case "credit":
		tp.Credit = value
	case "author":
		tp.Author = value
	case "authors":
		tp.Authors = value
	case "source":
		tp.Source = value
	case "draft date":
		tp.DraftDate = value
	case "contact":
		tp.Contact = value
	case "copyright":
		tp.Copyright = value
	case "notes":
		tp.Notes = value
	default:
		tp.Custom[key] = value
	}
}

// extractSceneNumber extracts and strips scene number markers (#...#) from text.
// Scene numbers appear at the end of lines in Fountain format.
// Returns the scene number (if found) and the cleaned text.
// Example: "ACTION TEXT #42#" → ("42", "ACTION TEXT")
func extractSceneNumber(text string) (string, string) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasSuffix(trimmed, "#") {
		return "", trimmed
	}

	// Find the closing # (last character)
	closingIdx := len(trimmed) - 1

	// Find the matching opening # by searching backwards
	openingIdx := strings.LastIndex(trimmed[:closingIdx], "#")
	if openingIdx == -1 {
		// No opening #
		return "", trimmed
	}

	sceneNo := trimmed[openingIdx+1 : closingIdx]
	if sceneNo == "" {
		// Empty scene number
		return "", trimmed
	}

	// Clean text is everything before the opening #
	cleanedText := strings.TrimSpace(trimmed[:openingIdx])
	return sceneNo, cleanedText
}

// isMultiCharacterLine checks if a character line contains multiple characters (slash-separated).
// Example: "JOHN / JANE" → true, "JOHN" → false
func isMultiCharacterLine(characterText string) bool {
	sceneNo, cleanedText := extractSceneNumber(characterText)
	_ = sceneNo // sceneNo already extracted, use cleaned text
	return strings.Contains(cleanedText, "/")
}

// parseMultiCharacterNames splits a multi-character line and validates/normalizes names.
// Example: "  JOHN  /  JANE  " → ["JOHN", "JANE"]
// Returns error if any name is empty.
func parseMultiCharacterNames(characterText string) ([]string, error) {
	sceneNo, cleanedText := extractSceneNumber(characterText)
	_ = sceneNo // sceneNo already extracted

	parts := strings.Split(cleanedText, "/")
	var names []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, fmt.Errorf("empty character name in multi-character line: %q", characterText)
		}
		names = append(names, trimmed)
	}

	return names, nil
}

// matchDialogueToCharacters matches dialogue lines to character names.
// If fewer dialogue lines than characters, the last dialogue line repeats.
// If more dialogue lines than characters, returns error.
// Returns slice of (characterName, dialogueLine) pairs.
func matchDialogueToCharacters(characterNames []string, dialogueLines []string) ([]struct {
	Name     string
	Dialogue string
}, error) {
	var result []struct {
		Name     string
		Dialogue string
	}

	numChars := len(characterNames)
	numLines := len(dialogueLines)

	if numLines > numChars {
		return nil, fmt.Errorf("more dialogue lines (%d) than characters (%d)", numLines, numChars)
	}

	for i, charName := range characterNames {
		dialogueLine := dialogueLines[i]
		if i >= numLines {
			// Fewer lines than characters: repeat last line
			dialogueLine = dialogueLines[numLines-1]
		}
		result = append(result, struct {
			Name     string
			Dialogue string
		}{Name: charName, Dialogue: dialogueLine})
	}

	return result, nil
}

// parseDialogueLines extracts dialogue lines from the next non-blank, non-parenthetical lines.
// Returns dialogue lines and a count indicating how many lines were consumed.
func parseDialogueLines(scanner *lineScanner) ([]string, int, error) {
	var lines []string
	linesConsumed := 0

	for scanner.hasMore() {
		nextLine := strings.TrimSpace(scanner.peek())

		// Stop at blank line
		if nextLine == "" {
			break
		}

		// Stop at parenthetical (will be handled separately)
		if parentheticalPattern.MatchString(nextLine) {
			break
		}

		// Stop at scene heading
		if sceneHeadingPattern.MatchString(nextLine) {
			break
		}

		// Stop at character line
		if characterPattern.MatchString(nextLine) {
			break
		}

		// Consume dialogue line
		scanner.next()
		lines = append(lines, nextLine)
		linesConsumed++
	}

	if len(lines) == 0 {
		return nil, 0, fmt.Errorf("no dialogue lines found after character")
	}

	return lines, linesConsumed, nil
}

// parseContent parses the main screenplay content.
func parseContent(lines []string) []ftnmodel.Element {
	var elements []ftnmodel.Element

	scanner := &lineScanner{lines: lines, pos: 0}
	var lastCharacter bool
	inBoneyard := false

	for scanner.hasMore() {
		line := scanner.next()
		trimmed := strings.TrimSpace(line)

		// Handle boneyard (comments)
		if boneyardStart.MatchString(line) {
			inBoneyard = true
			continue
		}
		if inBoneyard {
			if boneyardEnd.MatchString(line) {
				inBoneyard = false
			}
			continue
		}

		// Blank line
		if trimmed == "" {
			lastCharacter = false
			elements = append(elements, ftnmodel.Element{Type: ftnmodel.ElementBlank})
			continue
		}

		// Page break
		if pageBreakPattern.MatchString(trimmed) {
			elements = append(elements, ftnmodel.Element{Type: ftnmodel.ElementPageBreak})
			continue
		}

		// Section heading
		if match := sectionPattern.FindStringSubmatch(trimmed); match != nil {
			elements = append(elements, ftnmodel.Element{
				Type:  ftnmodel.ElementSection,
				Text:  match[2],
				Depth: len(match[1]),
			})
			continue
		}

		// Synopsis
		if match := synopsisPattern.FindStringSubmatch(trimmed); match != nil {
			elements = append(elements, ftnmodel.Element{
				Type: ftnmodel.ElementSynopsis,
				Text: match[1],
			})
			continue
		}

		// Centered text
		if match := centeredPattern.FindStringSubmatch(trimmed); match != nil {
			elements = append(elements, ftnmodel.Element{
				Type:   ftnmodel.ElementCentered,
				Text:   strings.TrimSpace(match[1]),
				Forced: true,
			})
			continue
		}

		// Forced scene heading
		if match := forcedScenePattern.FindStringSubmatch(trimmed); match != nil {
			elements = append(elements, ftnmodel.Element{
				Type:   ftnmodel.ElementSceneHeading,
				Text:   match[1],
				Forced: true,
			})
			lastCharacter = false
			continue
		}

		// Forced action
		if match := forcedActionPattern.FindStringSubmatch(trimmed); match != nil {
			sceneNo, cleanedText := extractSceneNumber(match[1])
			elements = append(elements, ftnmodel.Element{
				Type:    ftnmodel.ElementAction,
				Text:    cleanedText,
				SceneNo: sceneNo,
				Forced:  true,
			})
			lastCharacter = false
			continue
		}

		// Scene heading
		if sceneHeadingPattern.MatchString(trimmed) {
			sceneNo, cleanedText := extractSceneNumber(trimmed)
			elements = append(elements, ftnmodel.Element{
				Type:    ftnmodel.ElementSceneHeading,
				Text:    cleanedText,
				SceneNo: sceneNo,
			})
			lastCharacter = false
			continue
		}

		// Forced transition
		if match := forcedTransPattern.FindStringSubmatch(trimmed); match != nil {
			if !centeredPattern.MatchString(trimmed) {
				elements = append(elements, ftnmodel.Element{
					Type:   ftnmodel.ElementTransition,
					Text:   strings.TrimSpace(match[1]),
					Forced: true,
				})
				continue
			}
		}

		// Transition
		if transitionPattern.MatchString(trimmed) {
			elements = append(elements, ftnmodel.Element{
				Type: ftnmodel.ElementTransition,
				Text: trimmed,
			})
			continue
		}

		// Parenthetical (only after character or dialogue)
		if lastCharacter && parentheticalPattern.MatchString(trimmed) {
			elements = append(elements, ftnmodel.Element{
				Type: ftnmodel.ElementParenthetical,
				Text: trimmed,
			})
			continue
		}

		// Forced character
		if match := forcedCharPattern.FindStringSubmatch(trimmed); match != nil {
			sceneNo, cleanedName := extractSceneNumber(match[1])
			dual := dualDialoguePattern.MatchString(cleanedName)
			name := dualDialoguePattern.ReplaceAllString(cleanedName, "")
			elements = append(elements, ftnmodel.Element{
				Type:    ftnmodel.ElementCharacter,
				Text:    name,
				SceneNo: sceneNo,
				Dual:    dual,
				Forced:  true,
			})
			lastCharacter = true
			continue
		}

		// Character (ALL CAPS followed by dialogue)
		if !lastCharacter && characterPattern.MatchString(trimmed) && scanner.hasMore() {
			nextLine := strings.TrimSpace(scanner.peek())
			if nextLine != "" && !sceneHeadingPattern.MatchString(nextLine) {
				// Check if this is multi-character dialogue
				if isMultiCharacterLine(trimmed) {
					// Multi-character dialogue handling
					charNames, err := parseMultiCharacterNames(trimmed)
					if err != nil {
						// Skip multi-character parsing on error, treat as regular action
						lastCharacter = false
						continue
					}

					sceneNo, _ := extractSceneNumber(trimmed)

					// Parse the dialogue lines
					dialogueLines, _, err := parseDialogueLines(scanner)
					if err != nil {
						// No valid dialogue, skip
						lastCharacter = false
						continue
					}

					// Match dialogue to characters
					matched, err := matchDialogueToCharacters(charNames, dialogueLines)
					if err != nil {
						// Dialogue count mismatch error, skip multi-character parsing
						lastCharacter = false
						continue
					}

					// Create character elements for each speaker
					for _, pair := range matched {
						elements = append(elements, ftnmodel.Element{
							Type:    ftnmodel.ElementCharacter,
							Text:    pair.Name,
							SceneNo: sceneNo,
							Multi:   true,
						})
					}

					// Create dialogue element
					dialogueText := ""
					if len(dialogueLines) > 0 {
						dialogueText = strings.Join(dialogueLines, " / ")
					}
					elements = append(elements, ftnmodel.Element{
						Type:    ftnmodel.ElementDialogue,
						Text:    dialogueText,
						SceneNo: sceneNo,
						Multi:   true,
					})

					// Check for parenthetical after dialogue
					if scanner.hasMore() {
						peekLine := strings.TrimSpace(scanner.peek())
						if parentheticalPattern.MatchString(peekLine) {
							scanner.next()
							sceneNo, cleanedText := extractSceneNumber(peekLine)
							elements = append(elements, ftnmodel.Element{
								Type:    ftnmodel.ElementParenthetical,
								Text:    cleanedText,
								SceneNo: sceneNo,
								Multi:   true,
							})
						}
					}

					lastCharacter = false
					continue
				}

				// Single character dialogue (original logic)
				sceneNo, cleanedText := extractSceneNumber(trimmed)
				dual := dualDialoguePattern.MatchString(cleanedText)
				name := dualDialoguePattern.ReplaceAllString(cleanedText, "")
				elements = append(elements, ftnmodel.Element{
					Type:    ftnmodel.ElementCharacter,
					Text:    name,
					SceneNo: sceneNo,
					Dual:    dual,
				})
				lastCharacter = true
				continue
			}
		}

		// Dialogue (after character)
		if lastCharacter {
			sceneNo, cleanedText := extractSceneNumber(trimmed)
			elements = append(elements, ftnmodel.Element{
				Type:    ftnmodel.ElementDialogue,
				Text:    cleanedText,
				SceneNo: sceneNo,
			})
			continue
		}

		// Forced action (must come after other forced patterns)
		if match := forcedActionPattern.FindStringSubmatch(trimmed); match != nil {
			sceneNo, cleanedText := extractSceneNumber(match[1])
			elements = append(elements, ftnmodel.Element{
				Type:    ftnmodel.ElementAction,
				Text:    cleanedText,
				SceneNo: sceneNo,
				Forced:  true,
			})
			lastCharacter = false
			continue
		}

		// Action (default)
		sceneNo, cleanedText := extractSceneNumber(trimmed)
		elements = append(elements, ftnmodel.Element{
			Type:    ftnmodel.ElementAction,
			Text:    cleanedText,
			SceneNo: sceneNo,
		})
		lastCharacter = false

		// Forced action
		if match := forcedActionPattern.FindStringSubmatch(trimmed); match != nil {
			elements = append(elements, ftnmodel.Element{
				Type:   ftnmodel.ElementAction,
				Text:   match[1],
				Forced: true,
			})
			continue
		}

		// Default: action
		elements = append(elements, ftnmodel.Element{
			Type: ftnmodel.ElementAction,
			Text: trimmed,
		})
		lastCharacter = false
	}

	return elements
}

// lineScanner provides a simple scanner for lines.
type lineScanner struct {
	lines []string
	pos   int
}

func (s *lineScanner) hasMore() bool {
	return s.pos < len(s.lines)
}

func (s *lineScanner) next() string {
	if s.pos >= len(s.lines) {
		return ""
	}
	line := s.lines[s.pos]
	s.pos++
	return line
}

func (s *lineScanner) peek() string {
	if s.pos >= len(s.lines) {
		return ""
	}
	return s.lines[s.pos]
}

// ReadFountain is a convenience function to parse Fountain from a bufio.Scanner.
func ReadFountain(scanner *bufio.Scanner) (*ftnmodel.Document, error) {
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	d := NewDecoder()
	return d.Decode(context.Background(), []byte(strings.Join(lines, "\n")))
}
