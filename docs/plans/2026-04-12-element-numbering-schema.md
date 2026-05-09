# Plan: Extend Element-Level Numbering to Action, Character, and Dialogue

**Date**: April 12, 2026  
**Status**: Planning  
**Scope**: screenjson-cli  
**Related**: 2026-04-12-element-numbering-schema.md

## Overview

Extend the sceneNo capability (currently implemented for Scene Headings via FDX Paragraph Numbers and Fountain # syntax) to Action, Character, and Dialogue elements. Numbers are optional, stored in element.SceneNo field, preserved in round-trip conversions, and display is configurable for future exporters.

## Objectives

1. Extract and preserve Paragraph Numbers from FDX files for Action/Character/Dialogue elements
2. Parse and preserve Fountain element numbering syntax (#NUMBER#) for Action/Character/Dialogue
3. Support optional display via export configuration (deferred for future phase)
4. Maintain backward compatibility with existing FDX/Fountain files without numbers
5. Ensure round-trip conversion preserves numbers (FDX → ScreenJSON → FDX and Fountain → ScreenJSON → Fountain)

## Scope & Decisions

### In Scope
- **Element Types**: Action, Character (cue), Dialogue ONLY
  - NOT Parenthetical, Transition, Shot, General
- **FDX Format**: Extract `Paragraph.Number` for these elements when importing
- **Fountain Format**: Parse `#NUMBER#` syntax for these elements when importing
- **Storage**: Store numbers in existing `element.SceneNo` field (already present on Element struct)
- **Validation**: Reuse existing `ValidateSceneNumber()` and `NormalizeSceneNumber()` functions
- **Requirement Level**: Numbers are optional (elements without numbers are valid)

### Out of Scope
- Display/rendering in export formats (future enhancement)
- GUI editing of element numbers
- Automatic number generation
- Changes to schema (already supports sceneNo on all element types)

## Current State

### What Already Exists ✅
- **Model**: Element struct has `SceneNo string` field (internal/model/elements.go line ~40)
- **Schema**: ScreenJSON schema allows sceneNo on all element types
- **Validation**: ValidateSceneNumber() and NormalizeSceneNumber() exist in internal/model/types.go
- **FDX Model**: Paragraph struct already has `Number` field
- **Fountain Model**: Element struct has `SceneNo` field

### What's Missing ❌
- **FDX Bridge**: Not extracting Paragraph.Number for Action/Character/Dialogue elements
- **Fountain Codec**: Not parsing #NUMBER# syntax for Action/Character/Dialogue elements
- **Tests**: No unit tests for element numbering (only Scene Heading numbering tested)

## Implementation Plan

### Phase 1: Verification (No Code Changes)
**Objective**: Confirm infrastructure supports element numbering

**Verification Tasks**:
1. ✅ Verify Element struct has SceneNo field - YES
2. ✅ Verify schema allows sceneNo on elements - YES
3. ✅ Verify validation functions exist - YES
4. ✅ Verify FDX Paragraph has Number field - YES
5. ✅ Verify Fountain Element has SceneNo field - YES

---

### Phase 2: FDX Format (Primary Implementation)

#### 2.1 Modify FDX Bridge to_screenjson

**File**: internal/formats/fdx/bridge/to_screenjson.go  
**Function**: `convertParagraphToElement()`

**Changes**: Extract `p.Number` for Action/Character/Dialogue and assign to `element.SceneNo` after normalization

#### 2.2 Modify FDX Bridge from_screenjson

**File**: internal/formats/fdx/bridge/from_screenjson.go  
**Function**: `convertElementToParagraph()`

**Changes**: Preserve `elem.SceneNo` → `p.Number` for round-trip guarantee

#### 2.3 Add FDX Tests (6 test cases)
- Action element with Number
- Character element with Number
- Dialogue element with Number
- Multiple numbered elements
- Element without Number
- Round-trip FDX preservation

---

### Phase 3: Fountain Format

#### 3.1 Modify Fountain Codec decode

**File**: internal/formats/fountain/codec/decode.go

**Changes**: For Action/Character/Dialogue, detect `#NUMBER#` delimiters and extract sceneNo

#### 3.2 Modify Fountain Bridge to_screenjson

**File**: internal/formats/fountain/bridge/to_screenjson.go

**Changes**: Pass `fountainElement.SceneNo` through to ScreenJSON element

#### 3.3 Modify Fountain Bridge from_screenjson

**File**: internal/formats/fountain/bridge/from_screenjson.go

**Changes**: Wrap `elem.SceneNo` in #...# delimiters when present

#### 3.4 Add Fountain Tests (7 test cases)
- Action with #NUMBER#
- Character with #1A#
- Dialogue with #110A/111B#
- Normalization ##42## → 42
- No number (plain text)
- Mixed elements
- Round-trip Fountain preservation

---

### Phase 4: Verification & Testing

**Regression Testing**: Run existing test suites
```bash
go test ./internal/formats/fdx/...
go test ./internal/formats/fountain/...