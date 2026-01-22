package crypto

import (
	"screenjson/cli/internal/model"
)

// EncryptDocument encrypts content strings in a ScreenJSON document.
func EncryptDocument(doc *model.Document, passphrase string) error {
	enc, err := NewEncryptor(passphrase)
	if err != nil {
		return err
	}

	// Set encryption metadata
	encoding := EncodingHex
	doc.Encrypt = &model.Encrypt{
		Cipher:   "aes-256-ctr",
		Hash:     "sha256",
		Encoding: encoding,
	}

	// Encrypt title
	for lang, text := range doc.Title {
		encrypted, err := enc.Encrypt(text)
		if err != nil {
			return err
		}
		doc.Title[lang] = encrypted
	}

	// Encrypt logline
	for lang, text := range doc.Logline {
		encrypted, err := enc.Encrypt(text)
		if err != nil {
			return err
		}
		doc.Logline[lang] = encrypted
	}

	// Encrypt character descriptions
	for i := range doc.Characters {
		for lang, text := range doc.Characters[i].Desc {
			encrypted, err := enc.Encrypt(text)
			if err != nil {
				return err
			}
			doc.Characters[i].Desc[lang] = encrypted
		}
	}

	// Encrypt content
	if doc.Content != nil {
		// Cover title
		if doc.Content.Cover != nil {
			for lang, text := range doc.Content.Cover.Title {
				encrypted, err := enc.Encrypt(text)
				if err != nil {
					return err
				}
				doc.Content.Cover.Title[lang] = encrypted
			}
			for lang, text := range doc.Content.Cover.Extra {
				encrypted, err := enc.Encrypt(text)
				if err != nil {
					return err
				}
				doc.Content.Cover.Extra[lang] = encrypted
			}
		}

		// Scenes
		for i := range doc.Content.Scenes {
			scene := &doc.Content.Scenes[i]
			
			// Scene heading description
			if scene.Heading != nil {
				for lang, text := range scene.Heading.Desc {
					encrypted, err := enc.Encrypt(text)
					if err != nil {
						return err
					}
					scene.Heading.Desc[lang] = encrypted
				}
			}

			// Scene body elements
			for j := range scene.Body {
				elem := &scene.Body[j]
				for lang, text := range elem.Text {
					encrypted, err := enc.Encrypt(text)
					if err != nil {
						return err
					}
					elem.Text[lang] = encrypted
				}

				// Notes
				for k := range elem.Notes {
					for lang, text := range elem.Notes[k].Text {
						encrypted, err := enc.Encrypt(text)
						if err != nil {
							return err
						}
						elem.Notes[k].Text[lang] = encrypted
					}
				}
			}
		}
	}

	return nil
}

// DecryptDocument decrypts content strings in a ScreenJSON document.
func DecryptDocument(doc *model.Document, passphrase string) error {
	if doc.Encrypt == nil {
		return nil // Not encrypted
	}

	dec, err := NewDecryptor(passphrase)
	if err != nil {
		return err
	}
	dec.WithEncoding(doc.Encrypt.Encoding)

	// Decrypt title
	for lang, text := range doc.Title {
		decrypted, err := dec.Decrypt(text)
		if err != nil {
			return err
		}
		doc.Title[lang] = decrypted
	}

	// Decrypt logline
	for lang, text := range doc.Logline {
		decrypted, err := dec.Decrypt(text)
		if err != nil {
			return err
		}
		doc.Logline[lang] = decrypted
	}

	// Decrypt character descriptions
	for i := range doc.Characters {
		for lang, text := range doc.Characters[i].Desc {
			decrypted, err := dec.Decrypt(text)
			if err != nil {
				return err
			}
			doc.Characters[i].Desc[lang] = decrypted
		}
	}

	// Decrypt content
	if doc.Content != nil {
		// Cover
		if doc.Content.Cover != nil {
			for lang, text := range doc.Content.Cover.Title {
				decrypted, err := dec.Decrypt(text)
				if err != nil {
					return err
				}
				doc.Content.Cover.Title[lang] = decrypted
			}
			for lang, text := range doc.Content.Cover.Extra {
				decrypted, err := dec.Decrypt(text)
				if err != nil {
					return err
				}
				doc.Content.Cover.Extra[lang] = decrypted
			}
		}

		// Scenes
		for i := range doc.Content.Scenes {
			scene := &doc.Content.Scenes[i]

			// Heading description
			if scene.Heading != nil {
				for lang, text := range scene.Heading.Desc {
					decrypted, err := dec.Decrypt(text)
					if err != nil {
						return err
					}
					scene.Heading.Desc[lang] = decrypted
				}
			}

			// Body elements
			for j := range scene.Body {
				elem := &scene.Body[j]
				for lang, text := range elem.Text {
					decrypted, err := dec.Decrypt(text)
					if err != nil {
						return err
					}
					elem.Text[lang] = decrypted
				}

				// Notes
				for k := range elem.Notes {
					for lang, text := range elem.Notes[k].Text {
						decrypted, err := dec.Decrypt(text)
						if err != nil {
							return err
						}
						elem.Notes[k].Text[lang] = decrypted
					}
				}
			}
		}
	}

	// Clear encryption metadata
	doc.Encrypt = nil

	return nil
}
