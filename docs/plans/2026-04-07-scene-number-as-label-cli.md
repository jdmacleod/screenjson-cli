# Scene Number as Label - Implementation Plan

**Project**: screenjson-cli  
**Branch**: `scene-number-as-label`  
**Objective**: Convert scene numbering from integer-only representation to string-based labels while maintaining backward compatibility and supporting alphanumeric production scene numbers.

## Rationale

Production workflows often use alphanumeric scene numbers (e.g., "1A", "I-1-A", "110A") for flexibility during script revisions. The existing integer-only `Slugline.No` field was insufficient. By converting to string and adding a complementary `Element.SceneNo` field, we enable:

- Preservation of source-format explicit scene numbers across all bridges
- Support for complex numbering schemes without loss or normalization
- Alignment with industry screenplay notation conventions
- Per-element scene number tracking for granular production workflows

## Execution Phases

### Phase 1: Core Model Updates

#### 1.1 Slugline.No: int → string
**File**: `internal/model/scene.go`

- Changed `Slugline.No` from `int` to `string`
- Field now accepts alphanumeric values: "1", "1A", "I-1-A", "110A", etc.
- Marked as optional (omitempty) for cases where no explicit number exists

#### 1.2 Element.SceneNo: New Field
**File**: `internal/model/elements.go`

- Added `SceneNo` field to `Element` struct
- Type: `string` (omitempty)
- Purpose: Preserve source format scene number on each element in a scene body
- Enables per-element tracking for dialogue cues, actions, transitions, etc.

#### 1.3 Validators and Helpers
**File**: `internal/model/types.go`

- `ValidateSceneNumber(s string) error`
  - Validates format: alphanumerics, dashes, periods only
  - Allows empty string (optional field)
  - Returns descriptive error on invalid characters

- `NormalizeSceneNumber(s string) string`
  - Strips leading/trailing `#` markers (Fountain convention)
  - Example: `#1A#` → `1A`

- `isValidSceneNumber(s string) bool`
  - Internal helper for pattern validation

### Phase 2: Schema Updates

#### 2.1 JSON Schema: schema.json
**File**: `internal/schema/schema.json`

Changes to `$defs`:
- **slugline.no** property:
  - Type changed from `integer` (minimum 1) to `string`
  - Regex pattern: `^[A-Za-z0-9][A-Za-z0-9.-]*$`
  - Matches: "1", "1A", "110A", "I-1-A", "1.1", etc.

- **element** definition:
  - Added new `sceneNo` property
  - Type: `string`
  - Regex pattern: same as slugline.no
  - Optional field

#### 2.2 Schema Validation Tests
**File**: `internal/schema/validator_test.go`

- `TestSchemaAllowsStringSceneNumbers`
  - Verifies schema JSON parses correctly
  - Confirms slugline.no type is string
  - Confirms element.sceneNo exists and is string
  - Validates regex patterns accept expected formats

### Phase 3: Format Bridge Updates

All bridges now extract explicit scene numbers from source formats and preserve them through the conversion pipeline.

#### 3.1 Fountain Bridge
**File**: `internal/formats/fountain/bridge/to_screenjson.go`

- Reads `Element.SceneNo` from Fountain model (parsed from `#...#` notation)
- Normalizes with `NormalizeSceneNumber()` to strip hash markers
- Falls back to sequential numbering (1, 2, 3...) if no explicit number
- Assigns `sceneNumberStr` to:
  - `Slugline.No` (one per scene heading)
  - `Element.SceneNo` on all body elements (character cues, dialogue, action, transitions, etc.)

#### 3.2 Celtx Bridge
**File**: `internal/formats/celtx/bridge/to_screenjson.go`

- Converts `sceneNumber` to `strconv.Itoa()` for string scene headings
- Applies consistent pattern to all scene creation calls
- Enables future explicit scene number extraction from Celtx-specific metadata

#### 3.3 FDX (Final Draft) Bridge
**File**: `internal/formats/fdx/bridge/to_screenjson.go`

- Uses `strconv.Itoa()` to convert sequential numbering to strings
- Supports future explicit FDX SceneNumber element extraction
- Normalizes formatting for string scene numbers

#### 3.4 FadeIn Bridge
**File**: `internal/formats/fadein/bridge/to_screenjson.go`

- Converts scene number counter to string via `strconv.Itoa()`
- Maintains sequential numbering strategy
- Prepares for explicit FadeIn scene number extraction

#### 3.5 PDF Bridge
**File**: `internal/formats/pdf/codec/decode.go`

- Updates geometry-based scene classification to work with string numbers
- Converts sequential `sceneNumber` counter to string
- Maintains backward compatibility with numeric-only PDF workflows

### Phase 4: Test Coverage

#### 4.1 Model Tests: Slugline.No and Element.SceneNo
**File**: `internal/model/scene_test.go`

Tests:
- `TestSluglineNoAsString` – validates Slugline.No holds string values
- `TestSluglineJSONMarshaling` – JSON encodes no as string
- `TestSluglineJSONUnmarshaling` – JSON decodes no as string
- `TestElementSceneNo` – Element.SceneNo field assignment
- `TestElementSceneNoJSONMarshaling` – JSON encodes sceneNo
- `TestElementSceneNoJSONUnmarshaling` – JSON decodes sceneNo

#### 4.2 Validation Tests
**File**: `internal/model/scene_validation_test.go`

Tests:
- `TestValidateSceneNumber` – validates allowed formats and rejects invalid characters
- `TestNormalizeSceneNumber` – verifies hash marker stripping
- `TestIsValidSceneNumber` – internal pattern validation

#### 4.3 Fountain Bridge Tests
**File**: `internal/formats/fountain/bridge/to_screenjson_test.go`

Tests:
- `TestToScreenJSONBasic` – basic document conversion
- `TestSceneNumberExtract` – explicit scene number normalization (#1#, #1A#, #I-1-A#)
- `TestSequentialNumbering` – fallback numbering (1, 2, 3...)
- `TestParseSluglineBasic` – slugline parsing for context/setting
- `TestFountainEndToEndSceneNumber` – full text-to-ScreenJSON pipeline
- `TestFountainEndToEndExplicitSceneNo` – explicit label preservation end-to-end

#### 4.4 Celtx Bridge Tests
**File**: `internal/formats/celtx/bridge/to_screenjson_test.go`

Tests:
- `TestToScreenJSONBasic` – basic document structure
- `TestCeltxLanguageDefault` – default language handling
- `TestCeltxGenerator` – generator metadata

#### 4.5 FDX Bridge Tests
**File**: `internal/formats/fdx/bridge/to_screenjson_test.go`

Tests:
- `TestToScreenJSONBasic` – basic document structure
- `TestFdxLanguageDefault` – default language handling
- `TestFdxGenerator` – generator metadata

#### 4.6 FadeIn Bridge Tests
**File**: `internal/formats/fadein/bridge/to_screenjson_test.go`

Tests:
- `TestToScreenJSONBasic` – basic document structure
- `TestFadeInLanguageDefault` – default language handling
- `TestFadeInGenerator` – generator metadata

#### 4.7 JSON Codec Tests
**File**: `internal/formats/json/codec/codec_test.go`

Tests:
- `TestJSONCodecRoundTrip` – encode/decode round trip preserves SceneNo

#### 4.8 Schema Validation Tests
**File**: `internal/schema/validator_test.go`

Tests:
- `TestValidatorValidScreenJSON` – valid ScreenJSON passes
- `TestValidatorMissingRequiredField` – missing fields caught
- `TestValidatorInvalidDocumentType` – type mismatches caught
- `TestSchemaAllowsStringSceneNumbers` – schema structure confirms string types

#### 4.9 Pipeline Integration Tests
**File**: `internal/pipeline/pipeline_test.go`

Tests:
- `TestPipelineJSONRoundTrip` – builder/writer round trip preserves SceneNo
- `TestPipelineDetectFormatJSON` – format detection works

### Verification Phases

#### V1: Build Verification
Command:
```bash
go build -o screenjson ./cmd/screenjson/
```

Result: ✅ Silent success (no build errors)

#### V2: Full Test Suite
Command:
```bash
go test ./...
```

Results:
- `screenjson/cli/internal/formats/celtx/bridge` ✅ ok
- `screenjson/cli/internal/formats/fadein/bridge` ✅ ok
- `screenjson/cli/internal/formats/fdx/bridge` ✅ ok
- `screenjson/cli/internal/formats/fountain/bridge` ✅ ok
- `screenjson/cli/internal/formats/json/codec` ✅ ok (0.307s)
- `screenjson/cli/internal/model` ✅ ok
- `screenjson/cli/internal/pipeline` ✅ ok (0.480s)
- `screenjson/cli/internal/schema` ✅ ok (0.293s)

#### V3-V7: Final Verification Tests

Schema integration tests confirm:
- `slugline.no` JSON schema type is string
- `element.sceneNo` JSON schema property exists and is string
- Both fields use correct regex patterns
- All format bridges assign strings to both fields
- End-to-end Fountain conversion preserves explicit labels
- JSON round trip preserves scene numbers

## Summary of Changes

### Files Modified or Created

**Core Model**:
- `internal/model/scene.go` – Slugline.No int→string
- `internal/model/elements.go` – Added Element.SceneNo
- `internal/model/types.go` – Added validators and normalizers

**Schema**:
- `internal/schema/schema.json` – Updated slugline.no and element.sceneNo definitions
- `internal/schema/validator_test.go` – Added schema structure validation

**Format Bridges**:
- `internal/formats/fountain/bridge/to_screenjson.go` – Scene number preservation
- `internal/formats/celtx/bridge/to_screenjson.go` – String conversion
- `internal/formats/fdx/bridge/to_screenjson.go` – String conversion
- `internal/formats/fadein/bridge/to_screenjson.go` – String conversion
- `internal/formats/pdf/codec/decode.go` – String conversion

**Tests**:
- `internal/model/scene_test.go` – Slugline.No and Element.SceneNo tests
- `internal/model/scene_validation_test.go` – Validator tests
- `internal/formats/fountain/bridge/to_screenjson_test.go` – Enhanced end-to-end tests
- `internal/formats/celtx/bridge/to_screenjson_test.go` – Basic tests
- `internal/formats/fdx/bridge/to_screenjson_test.go` – Basic tests
- `internal/formats/fadein/bridge/to_screenjson_test.go` – Basic tests
- `internal/formats/json/codec/codec_test.go` – Round-trip tests
- `internal/pipeline/pipeline_test.go` – Integration tests

**Documentation**:
- `TESTS.md` – Test guide and run instructions
- `docs/plans/scene-number-as-label.md` – This file

### Design Decisions

1. **String Type Over Integer**
   - Supports industry-standard alphanumeric numbering
   - Optional field (empty string is valid)
   - Pattern validation prevents invalid characters

2. **Dual Fields: Slugline.No + Element.SceneNo**
   - Slugline.No: one per scene (heading level)
   - Element.SceneNo: one per body element (granular tracking)
   - Enables both scene-wide and element-specific numbering queries

3. **Normalization Approach**
   - Strip source-format markers (#...# in Fountain)
   - Preserve alphanumeric content
   - Fall back to sequential for missing explicit numbers

4. **Bridge Consistency**
   - All bridges follow same pattern
   - Sequential numbering fallback uniformly applied
   - String conversion via `strconv.Itoa()` for consistency

## Testing Strategy

- **Unit tests** cover individual functions (validators, helpers, marshaling)
- **Integration tests** verify end-to-end pipelines (text → model → ScreenJSON)
- **Schema tests** confirm JSON Schema structure matches implementation
- **Round-trip tests** ensure encode/decode preserves scene numbers

## Future Considerations

1. **Explicit Scene Number Extraction**
   - PDF: extract from document metadata or heuristics
   - FDX: read from SceneNumber element
   - Celtx: extract from scene properties
   - FadeIn: extract from scene metadata

2. **Advanced Numbering Validation**
   - Custom regex per production workflow
   - Configurable numbering schemes

3. **Scene Number Formatting**
   - Template-based formatting for export
   - Automatic renumbering utilities

## Branch Information

- **Branch Name**: `scene-number-as-label`
- **Base**: `develop`
- **Status**: ✅ Complete (all phases, tests, and verification passed)
